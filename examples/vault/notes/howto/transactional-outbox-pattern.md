---
title: Transactional Outbox Pattern
slug: transactional-outbox-pattern
type: howto
depth: 3
confidence: low
created: 2026-08-09
updated: 2026-08-09
verified: 2026-08-09
freshness_days: 180
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---
# Transactional Outbox Pattern

> Part of [[food-ordering-system-course-architecture-index]] — Food Ordering System course notes.

---

## The Dual-Write Problem

After saving an order to the database, you want to publish a Kafka message.

**Scenario A — DB commit succeeds, Kafka fails:**
```
1. orderRepo.save(order);    ← committed ✓
2. kafka.send(event);         ← network error ✗
Result: Order is in DB, payment-service never knows. Order stuck forever.
```

**Scenario B — Kafka publish succeeds, DB fails:**
```
1. kafka.send(event);         ← published ✓
2. orderRepo.save(order);     ← deadlock, rolled back ✗
Result: Payment-service processes a phantom order. Money deducted, no order in DB.
```

**The fundamental constraint:** You can't atomically commit to a DB and publish to Kafka.
They're different systems.

---

## The Solution: Outbox Table

Instead of publishing to Kafka directly, **write a record to an outbox table in the same DB transaction**.
A separate process polls the table and publishes to Kafka asynchronously.

```
Business Transaction (one @Transactional):
    ├── INSERT into orders table
    └── INSERT into payment_outbox table (status=STARTED)

Separate scheduler process:
    ├── SELECT * FROM payment_outbox WHERE status='STARTED'
    ├── FOR EACH row: kafka.send(payload)
    └── UPDATE payment_outbox SET status='COMPLETED'  (on ack)
```

Now if Kafka is down at commit time: the row sits in outbox, scheduler retries later.
If the scheduler crashes after publish but before updating status: Kafka gets duplicate.
But at-least-once delivery + idempotent consumers (SagaId check) handle that.

---

## Outbox Architecture in This Project

Every service has its own outbox tables. In the order-service there are two:

| Table | Purpose |
|-------|---------|
| `payment_outbox` | Triggers payment requests and payment cancellations |
| `approval_outbox` | Triggers restaurant approval requests |

---

## The Outbox JPA Entity

```java
// order-dataaccess/outbox/payment/entity/PaymentOutboxEntity.java
@Entity
@Table(name = "payment_outbox")
public class PaymentOutboxEntity {

    @Id
    private UUID id;          // unique outbox message ID

    private UUID sagaId;      // ties all steps of one saga run together

    private ZonedDateTime createdAt;
    private ZonedDateTime processedAt;

    private String type;      // always "OrderProcessingSaga" — identifies which saga type

    private String payload;   // JSON-serialized event data (OrderPaymentEventPayload)

    @Enumerated(EnumType.STRING)
    private SagaStatus sagaStatus;  // STARTED, PROCESSING, SUCCEEDED, COMPENSATING, COMPENSATED

    @Enumerated(EnumType.STRING)
    private OrderStatus orderStatus;

    @Enumerated(EnumType.STRING)
    private OutboxStatus outboxStatus;  // STARTED, PROCESSING, COMPLETED — controls scheduler

    @Version
    private int version;     // optimistic locking — prevents two scheduler instances publishing same row
}
```

**The `@Version` field is critical** — if two scheduler instances (in a cluster) race to publish the
same outbox message, one will get an `OptimisticLockException` and back off.

---

## Writing to the Outbox — Same Transaction as Domain Work

```java
// OrderCreateCommandHandler.java
@Transactional   // ← one transaction for everything below
public CreateOrderResponse createOrder(CreateOrderCommand createOrderCommand) {

    // Step 1: Persist the order (to orders table)
    OrderCreatedEvent orderCreatedEvent = orderCreateHelper.persistOrder(createOrderCommand);

    // Step 2: Write to outbox table (payment_outbox — SAME TRANSACTION)
    paymentOutboxHelper.savePaymentOutboxMessage(
        orderDataMapper.orderCreatedEventToOrderPaymentEventPayload(orderCreatedEvent),
        orderCreatedEvent.getOrder().getOrderStatus(),      // PENDING
        orderSagaHelper.orderStatusToSagaStatus(
            orderCreatedEvent.getOrder().getOrderStatus()), // STARTED
        OutboxStatus.STARTED,    // scheduler will pick this up
        UUID.randomUUID()        // sagaId — new for each saga run
    );

    return orderDataMapper.orderToCreateOrderResponse(
        orderCreatedEvent.getOrder(), "Order created successfully");
}
```

`PaymentOutboxHelper.savePaymentOutboxMessage()` builds and saves the outbox message:

```java
// PaymentOutboxHelper.java
@Transactional  // participates in caller's transaction (no new transaction created)
public void savePaymentOutboxMessage(OrderPaymentEventPayload paymentEventPayload,
                                     OrderStatus orderStatus,
                                     SagaStatus sagaStatus,
                                     OutboxStatus outboxStatus,
                                     UUID sagaId) {
    save(OrderPaymentOutboxMessage.builder()
        .id(UUID.randomUUID())
        .sagaId(sagaId)
        .createdAt(paymentEventPayload.getCreatedAt())
        .type(ORDER_SAGA_NAME)                            // "OrderProcessingSaga"
        .payload(createPayload(paymentEventPayload))      // JSON serialization
        .orderStatus(orderStatus)
        .sagaStatus(sagaStatus)
        .outboxStatus(outboxStatus)                       // STARTED
        .build());
}

private String createPayload(OrderPaymentEventPayload payload) {
    try {
        return objectMapper.writeValueAsString(payload);  // Jackson serialization
    } catch (JsonProcessingException e) {
        throw new OrderDomainException("Could not create OrderPaymentEventPayload json!", e);
    }
}
```

The `payload` stored as JSON looks like:
```json
{
  "orderId": "abc-123",
  "customerId": "def-456",
  "price": 150.00,
  "createdAt": "2026-05-13T10:30:00Z",
  "paymentOrderStatus": "PENDING"
}
```

---

## The Scheduler — Polls and Publishes

```java
// PaymentOutboxScheduler.java
@Component
public class PaymentOutboxScheduler implements OutboxScheduler {

    @Override
    @Transactional
    @Scheduled(
        fixedDelayString = "${order-service.outbox-scheduler-fixed-rate}",    // e.g. 2000ms
        initialDelayString = "${order-service.outbox-scheduler-initial-delay}" // e.g. 1000ms
    )
    public void processOutboxMessage() {
        Optional<List<OrderPaymentOutboxMessage>> outboxMessagesResponse =
            paymentOutboxHelper.getPaymentOutboxMessageByOutboxStatusAndSagaStatus(
                OutboxStatus.STARTED,           // only pick up STARTED messages
                SagaStatus.STARTED,             // forward payment requests
                SagaStatus.COMPENSATING         // payment cancellation requests
            );

        if (outboxMessagesResponse.isPresent() && !outboxMessagesResponse.get().isEmpty()) {
            List<OrderPaymentOutboxMessage> outboxMessages = outboxMessagesResponse.get();

            outboxMessages.forEach(outboxMessage ->
                paymentRequestMessagePublisher.publish(outboxMessage, this::updateOutboxStatus)
                // callback passed as BiConsumer — called by publisher after Kafka ack
            );
        }
    }

    // This callback is called by the Kafka publisher after Kafka sends an ack
    private void updateOutboxStatus(OrderPaymentOutboxMessage msg, OutboxStatus outboxStatus) {
        msg.setOutboxStatus(outboxStatus);    // COMPLETED on success, FAILED on error
        paymentOutboxHelper.save(msg);
        log.info("OrderPaymentOutboxMessage updated to status: {}", outboxStatus.name());
    }
}
```

The publisher calls the callback with `COMPLETED` or `FAILED`:

```java
// OrderPaymentEventKafkaPublisher.java
@Override
public void publish(OrderPaymentOutboxMessage orderPaymentOutboxMessage,
                    BiConsumer<OrderPaymentOutboxMessage, OutboxStatus> outboxCallback) {
    // Deserialize JSON payload back to typed object
    OrderPaymentEventPayload eventPayload =
        kafkaMessageHelper.getOrderEventPayload(
            orderPaymentOutboxMessage.getPayload(), OrderPaymentEventPayload.class
        );

    String sagaId = orderPaymentOutboxMessage.getSagaId().toString();

    try {
        PaymentRequestAvroModel avroModel =
            messagingMapper.orderPaymentEventPayloadToPaymentRequestAvroModel(sagaId, eventPayload);

        kafkaProducer.send(
            paymentRequestTopicName,
            sagaId,          // ← using sagaId as Kafka message key guarantees partition ordering
            avroModel,
            kafkaMessageHelper.getKafkaCallback(
                topicName, avroModel, orderPaymentOutboxMessage,
                outboxCallback,         // ← success: outboxCallback(msg, COMPLETED)
                eventPayload.getOrderId(),
                "PaymentRequestAvroModel"
            )
        );
    } catch (Exception e) {
        log.error("Error sending to Kafka for order id: {}, saga id: {}",
            eventPayload.getOrderId(), sagaId);
        // Note: outboxCallback NOT called with FAILED here — message stays STARTED for retry
    }
}
```

---

## The Cleaner Scheduler

Completed outbox rows accumulate. A midnight cron job removes finished ones:

```java
// PaymentOutboxCleanerScheduler.java
@Scheduled(cron = "@midnight")
public void processOutboxMessage() {
    Optional<List<OrderPaymentOutboxMessage>> completed =
        paymentOutboxHelper.getPaymentOutboxMessageByOutboxStatusAndSagaStatus(
            OutboxStatus.COMPLETED,
            SagaStatus.SUCCEEDED,    // happy path complete
            SagaStatus.FAILED,       // failed beyond recovery
            SagaStatus.COMPENSATED   // rollback complete
        );

    if (completed.isPresent()) {
        paymentOutboxHelper.deletePaymentOutboxMessageByOutboxStatusAndSagaStatus(
            OutboxStatus.COMPLETED,
            SagaStatus.SUCCEEDED, SagaStatus.FAILED, SagaStatus.COMPENSATED
        );
    }
}
```

---

## Outbox Repository — Full Stack

The domain defines the port (interface), data-access implements it:

```java
// Domain output port (no JPA imports):
public interface PaymentOutboxRepository {
    OrderPaymentOutboxMessage save(OrderPaymentOutboxMessage msg);

    Optional<List<OrderPaymentOutboxMessage>> findByTypeAndOutboxStatusAndSagaStatus(
        String type, OutboxStatus outboxStatus, SagaStatus... sagaStatus);

    Optional<OrderPaymentOutboxMessage> findByTypeAndSagaIdAndSagaStatus(
        String type, UUID sagaId, SagaStatus... sagaStatus);

    void deleteByTypeAndOutboxStatusAndSagaStatus(
        String type, OutboxStatus outboxStatus, SagaStatus... sagaStatus);
}

// Spring Data JPA repository:
@Repository
public interface PaymentOutboxJpaRepository extends JpaRepository<PaymentOutboxEntity, UUID> {
    Optional<List<PaymentOutboxEntity>> findByTypeAndOutboxStatusAndSagaStatusIn(
        String type, OutboxStatus outboxStatus, List<SagaStatus> sagaStatus);

    Optional<PaymentOutboxEntity> findByTypeAndSagaIdAndSagaStatusIn(
        String type, UUID sagaId, List<SagaStatus> sagaStatus);

    void deleteByTypeAndOutboxStatusAndSagaStatusIn(
        String type, OutboxStatus outboxStatus, List<SagaStatus> sagaStatus);
}

// Adapter (implements domain port using JPA):
@Component
public class PaymentOutboxRepositoryImpl implements PaymentOutboxRepository {

    @Override
    public Optional<OrderPaymentOutboxMessage> findByTypeAndSagaIdAndSagaStatus(
            String type, UUID sagaId, SagaStatus... sagaStatus) {
        return paymentOutboxJpaRepository
            .findByTypeAndSagaIdAndSagaStatusIn(type, sagaId, Arrays.asList(sagaStatus))
            .map(mapper::paymentOutboxEntityToOrderPaymentOutboxMessage);
    }
}
```

Note: JPA uses `List<SagaStatus>` for `IN` queries. Domain uses `SagaStatus...` varargs (cleaner API).
The adapter bridges the difference with `Arrays.asList(sagaStatus)`.

---

## Idempotency via Outbox + SagaId

When a duplicate Kafka message arrives (Kafka guarantees at-least-once, not exactly-once),
the SAGA checks if the outbox record has already been processed:

```java
// OrderPaymentSaga.process()
Optional<OrderPaymentOutboxMessage> outboxMsgOpt =
    paymentOutboxHelper.getPaymentOutboxMessageBySagaIdAndSagaStatus(
        UUID.fromString(paymentResponse.getSagaId()),
        SagaStatus.STARTED    // only found if NOT yet processed
    );

if (outboxMsgOpt.isEmpty()) {
    // Message already processed — outbox is now in PROCESSING or SUCCEEDED
    log.info("Duplicate message, already processed saga id: {}",
        paymentResponse.getSagaId());
    return; // safe no-op
}
```

**The unique constraint:** The `payment_outbox` table has a unique index on `(type, saga_id, saga_status)`.
If two concurrent threads try to insert the same saga twice, one gets a `UNIQUE_VIOLATION` SQL exception.
The Kafka listener catches this explicitly:

```java
// PaymentRequestKafkaListener.java
} catch (DataAccessException e) {
    SQLException sqlEx = (SQLException) e.getRootCause();
    if (sqlEx != null && PSQLState.UNIQUE_VIOLATION.getState().equals(sqlEx.getSQLState())) {
        log.error("Unique constraint exception — duplicate message for order: {}",
            avroModel.getOrderId());
        // NO-OP — safe to swallow, this was a duplicate
    } else {
        throw new PaymentApplicationServiceException("DataAccessException: " + e.getMessage(), e);
    }
}
```

---

## Outbox Status Lifecycle Diagram

```
                    ┌──────────────────────────────────┐
                    │          payment_outbox           │
                    └──────────────────────────────────┘
                                    │
                     INSERT with outboxStatus=STARTED
                                    │
                    ┌───────────────▼───────────────┐
                    │           STARTED             │
                    │  sagaStatus: STARTED or       │
                    │             COMPENSATING      │
                    └───────────────┬───────────────┘
                                    │
                    @Scheduled picks up STARTED rows
                                    │
                    ┌───────────────▼───────────────┐
                    │        Kafka.send()           │
                    └───────┬───────────────┬───────┘
                   success  │               │  failure
                    ┌───────▼──────┐   ┌───▼──────────┐
                    │  COMPLETED   │   │  stays STARTED│
                    │ (Kafka ack'd)│   │ (retry next   │
                    └───────┬──────┘   │  scheduler    │
                            │          │  run)         │
                            │          └──────────────-┘
                    @midnight cleaner deletes
                    COMPLETED rows where
                    sagaStatus IN (SUCCEEDED, COMPENSATED, FAILED)
```

---

## Why Not Just Use Kafka Transactions?

Kafka does have transactions (`producer.initTransactions()`), but:
1. They don't span Kafka + PostgreSQL — you still have the dual-write problem.
2. They require all consumers to use `isolation.level=read_committed` which reduces throughput.
3. The outbox pattern is simpler, works with any message broker, and gives you a queryable audit log.

---

## Related Notes

- [[saga-pattern-choreography-based]] — what triggers outbox messages and how they drive SAGA state
- [[cqrs-and-event-driven-messaging]] — how Avro models and Kafka topics are structured
