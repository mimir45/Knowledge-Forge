---
title: CQRS and Event-Driven Messaging
slug: cqrs-and-event-driven-messaging
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
# CQRS and Event-Driven Messaging

> Part of [[food-ordering-system-course-architecture-index]] — Food Ordering System course notes.

---

## CQRS — Command Query Responsibility Segregation

CQRS says: **the model that mutates state should be separate from the model that reads state**.

This project applies it at the application-service level. Not full CQRS (no separate read database),
but the principle is clearly visible.

---

## Commands vs Queries — Separate DTO Classes

```java
// Command — has side effects (creates an order)
@Getter @Builder @AllArgsConstructor
public class CreateOrderCommand {
    @NotNull private final UUID customerId;
    @NotNull private final UUID restaurantId;
    @NotNull private final BigDecimal price;
    @NotNull private final List<OrderItem> items;
    @NotNull private final OrderAddress address;
}

// Query — read-only, no side effects
@Getter @Builder @AllArgsConstructor
public class TrackOrderQuery {
    @NotNull private final UUID orderTrackingId;
}
```

Separate response DTOs — the read and write models are independent:

```java
// Write response — confirms creation, gives tracking ID
@Getter @Builder @AllArgsConstructor
public class CreateOrderResponse {
    @NotNull private final UUID orderTrackingId;
    @NotNull private final OrderStatus orderStatus;
    @NotNull private final String message;
}

// Read response — richer, includes failure messages
@Getter @Builder @AllArgsConstructor
public class TrackOrderResponse {
    @NotNull private final UUID orderTrackingId;
    @NotNull private final OrderStatus orderStatus;
    private final List<String> failureMessages;
}
```

---

## Separate Command Handlers

```java
// OrderApplicationServiceImpl.java — delegates to separate handlers
@Service
class OrderApplicationServiceImpl implements OrderApplicationService {

    private final OrderCreateCommandHandler orderCreateCommandHandler;
    private final OrderTrackCommandHandler orderTrackCommandHandler;

    @Override
    public CreateOrderResponse createOrder(CreateOrderCommand createOrderCommand) {
        return orderCreateCommandHandler.createOrder(createOrderCommand);
        // mutates: creates order + writes to outbox (same transaction)
    }

    @Override
    public TrackOrderResponse trackOrder(TrackOrderQuery trackOrderQuery) {
        return orderTrackCommandHandler.trackOrder(trackOrderQuery);
        // read-only: queries local orders table by tracking ID
    }
}
```

`OrderCreateCommandHandler` — writes:

```java
@Component
public class OrderCreateCommandHandler {

    @Transactional
    public CreateOrderResponse createOrder(CreateOrderCommand createOrderCommand) {
        OrderCreatedEvent orderCreatedEvent = orderCreateHelper.persistOrder(createOrderCommand);
        paymentOutboxHelper.savePaymentOutboxMessage(...);
        return createOrderResponse;
    }
}
```

`OrderTrackCommandHandler` — reads:

```java
@Component
public class OrderTrackCommandHandler {

    public TrackOrderResponse trackOrder(TrackOrderQuery trackOrderQuery) {
        Optional<Order> order = orderRepository.findByTrackingId(
            new TrackingId(trackOrderQuery.getOrderTrackingId())
        );
        if (order.isEmpty()) {
            throw new OrderNotFoundException(
                "Could not find order with tracking id: " + trackOrderQuery.getOrderTrackingId()
            );
        }
        return orderDataMapper.orderToTrackOrderResponse(order.get());
    }
}
```

---

## The Read Model — CQRS in Messaging

The order-service needs to validate that a customer exists before creating an order.
But calling customer-service via REST would create coupling.

**Solution:** order-service maintains its own local `customers` table — a **read model** populated
by consuming Kafka events from customer-service.

```java
// CustomerMessageListenerImpl.java — listens to "customer" Kafka topic
@Service
public class CustomerMessageListenerImpl implements CustomerMessageListener {

    @Override
    public void customerCreated(CustomerModel customerModel) {
        // Saves customer to ORDER-SERVICE's own database (not customer-service's)
        Customer customer = customerRepository.save(
            orderDataMapper.customerModelToCustomer(customerModel)
        );
        log.info("Customer saved to order DB with id: {}", customer.getId());
    }
}
```

When an order is created, it checks this local table — no HTTP call needed:

```java
// OrderCreateHelper.java
private void checkCustomer(UUID customerId) {
    Optional<Customer> customer = customerRepository.findCustomer(customerId);
    if (customer.isEmpty()) {
        throw new OrderDomainException("Could not find customer with id: " + customerId);
    }
}
```

**This is CQRS at the service level:** customer-service owns the write model for customers.
Order-service maintains a read-only copy for its own queries. They sync via Kafka events.

Same pattern for `restaurants` — order-service has a local `restaurants` table populated from a
shared `restaurant` DB view.

---

## Avro Models — The Message Contract

All Kafka messages use Avro (schema-based serialization, not JSON). Avro schemas are version-managed,
giving strong contracts between producers and consumers.

Generated classes live in `kafka-model`:

```java
// kafka-model/PaymentRequestAvroModel.java (auto-generated from .avsc schema)
public class PaymentRequestAvroModel {
    private CharSequence id;
    private CharSequence sagaId;      // ← threads through all saga steps
    private CharSequence customerId;
    private CharSequence orderId;
    private BigDecimal price;
    private Instant createdAt;
    private PaymentOrderStatus paymentOrderStatus;  // PENDING or CANCELLED
}

public class PaymentResponseAvroModel {
    private CharSequence id;
    private CharSequence sagaId;
    private CharSequence orderId;
    private CharSequence paymentId;
    private CharSequence customerId;
    private BigDecimal price;
    private Instant createdAt;
    private PaymentStatus paymentStatus;     // COMPLETED, CANCELLED, FAILED
    private List<CharSequence> failureMessages;
}
```

---

## Messaging Data Mapper — Avro ↔ Domain

A dedicated mapper converts between Avro models and domain DTOs, isolating the messaging layer:

```java
// payment-messaging/mapper/PaymentMessagingDataMapper.java
@Component
public class PaymentMessagingDataMapper {

    // Incoming: Avro model (from Kafka) → domain DTO
    public PaymentRequest paymentRequestAvroModelToPaymentRequest(
            PaymentRequestAvroModel model) {
        return PaymentRequest.builder()
            .id(model.getId().toString())
            .sagaId(model.getSagaId().toString())    // keep sagaId for idempotency
            .customerId(model.getCustomerId().toString())
            .orderId(model.getOrderId().toString())
            .price(model.getPrice())
            .createdAt(model.getCreatedAt())
            .paymentOrderStatus(PaymentOrderStatus.valueOf(model.getPaymentOrderStatus().name()))
            .build();
    }

    // Outgoing: domain payload → Avro model (for Kafka)
    public PaymentResponseAvroModel orderEventPayloadToPaymentResponseAvroModel(
            String sagaId, OrderEventPayload payload) {
        return PaymentResponseAvroModel.newBuilder()
            .setId(UUID.randomUUID().toString())
            .setSagaId(sagaId)
            .setOrderId(payload.getOrderId())
            .setPrice(payload.getAmount())
            .setCreatedAt(payload.getCreatedAt().toInstant())
            .setPaymentStatus(PaymentStatus.valueOf(payload.getPaymentStatus()))
            .setFailureMessages(payload.getFailureMessages())
            .build();
    }
}
```

---

## Kafka Consumer — Full Implementation

```java
// PaymentRequestKafkaListener.java
@Component
public class PaymentRequestKafkaListener implements KafkaConsumer<PaymentRequestAvroModel> {

    @Override
    @KafkaListener(
        id = "${kafka-consumer-config.payment-consumer-group-id}",
        topics = "${payment-service.payment-request-topic-name}"
    )
    public void receive(@Payload List<PaymentRequestAvroModel> messages,
                        @Header(KafkaHeaders.RECEIVED_MESSAGE_KEY) List<String> keys,
                        @Header(KafkaHeaders.RECEIVED_PARTITION_ID) List<Integer> partitions,
                        @Header(KafkaHeaders.OFFSET) List<Long> offsets) {

        log.info("{} payment requests received with keys: {}, partitions: {}, offsets: {}",
            messages.size(), keys, partitions, offsets);

        messages.forEach(avroModel -> {
            try {
                if (PaymentOrderStatus.PENDING == avroModel.getPaymentOrderStatus()) {
                    paymentRequestMessageListener.completePayment(
                        mapper.paymentRequestAvroModelToPaymentRequest(avroModel));

                } else if (PaymentOrderStatus.CANCELLED == avroModel.getPaymentOrderStatus()) {
                    paymentRequestMessageListener.cancelPayment(
                        mapper.paymentRequestAvroModelToPaymentRequest(avroModel));
                }

            } catch (DataAccessException e) {
                // Handle unique constraint violation — duplicate message, safe to ignore
                SQLException sqlEx = (SQLException) e.getRootCause();
                if (sqlEx != null
                        && PSQLState.UNIQUE_VIOLATION.getState().equals(sqlEx.getSQLState())) {
                    log.error("Unique constraint violation for order: {}",
                        avroModel.getOrderId());
                    // NO-OP — idempotent
                } else {
                    throw new PaymentApplicationServiceException(e.getMessage(), e);
                }

            } catch (PaymentNotFoundException e) {
                // Payment doesn't exist — also safe to ignore (already processed/cleaned up)
                log.error("No payment found for order id: {}", avroModel.getOrderId());
            }
        });
    }
}
```

Key things to notice:
1. **Batch consumption** — `List<PaymentRequestAvroModel>` — processes multiple messages per poll
2. **Headers captured** — keys, partitions, offsets — used for logging and debugging
3. **Two exception paths that are silent no-ops** — unique constraint and PaymentNotFound both swallowed
4. **Delegates immediately to input port** — listener itself has zero business logic

---

## Kafka Producer — Async With Callback

```java
// KafkaProducerImpl.java
@Component
public class KafkaProducerImpl<K extends Serializable, V extends SpecificRecordBase>
        implements KafkaProducer<K, V> {

    @Override
    public void send(String topicName, K key, V message,
                     ListenableFutureCallback<SendResult<K, V>> callback) {
        log.info("Sending message={} to topic={}", message, topicName);
        try {
            ListenableFuture<SendResult<K, V>> kafkaResultFuture =
                kafkaTemplate.send(topicName, key, message);

            kafkaResultFuture.addCallback(callback); // callback updates outbox status
        } catch (KafkaException e) {
            log.error("Error sending message to topic={}, message={}", topicName, message);
            throw new KafkaProducerException("Error sending kafka message!");
        }
    }
}
```

The callback built by `KafkaMessageHelper`:

```java
// KafkaMessageHelper.java
public <T extends SpecificRecordBase> ListenableFutureCallback<SendResult<String, T>>
        getKafkaCallback(String topicName, T avroModel,
                         OrderPaymentOutboxMessage outboxMessage,
                         BiConsumer<OrderPaymentOutboxMessage, OutboxStatus> outboxCallback,
                         String orderId, String avroModelName) {
    return new ListenableFutureCallback<>() {
        @Override
        public void onSuccess(SendResult<String, T> result) {
            RecordMetadata metadata = result.getRecordMetadata();
            log.info("Sent {} to topic: {} partition: {} offset: {} at {}",
                avroModelName, metadata.topic(), metadata.partition(),
                metadata.offset(), metadata.timestamp());
            outboxCallback.accept(outboxMessage, OutboxStatus.COMPLETED); // ← mark as done
        }

        @Override
        public void onFailure(Throwable ex) {
            log.error("Error sending {} to topic: {}, error: {}",
                avroModelName, topicName, ex.getMessage());
            outboxCallback.accept(outboxMessage, OutboxStatus.FAILED); // ← mark for retry
        }
    };
}
```

---

## Kafka Config — Data-Driven Setup

No hard-coded values. Everything is in `application.yml`:

```java
// KafkaConfigData.java
@Configuration
@ConfigurationProperties(prefix = "kafka-config")
public class KafkaConfigData {
    private String bootstrapServers;
    private String schemaRegistryUrlKey;
    private String schemaRegistryUrl;
    private Integer numOfPartitions;
    private Short replicationFactor;
}

// KafkaConsumerConfigData.java
@Configuration
@ConfigurationProperties(prefix = "kafka-consumer-config")
public class KafkaConsumerConfigData {
    private String keyDeserializer;
    private String valueDeserializer;
    private String autoOffsetReset;
    private String specificAvroReaderKey;
    private String specificAvroReader;
    private Boolean batchListener;
    private Boolean autoStartup;
    private Integer concurrencyLevel;
    private Integer sessionTimeoutMs;
    private Integer heartbeatIntervalMs;
    private Integer maxPollIntervalMs;
    private Long pollTimeoutMs;
    private Integer maxPollRecords;
    private Integer maxPartitionFetchBytesDefault;
    private Integer maxPartitionFetchBytesBoostFactor;
    private String paymentConsumerGroupId;
    private String restaurantApprovalConsumerGroupId;
}
```

---

## Differences From "Usual" Spring Boot Kafka

| Aspect | Usual | This Project |
|--------|-------|-------------|
| Message format | JSON | Avro (schema-validated) |
| Publishing | `kafkaTemplate.send()` directly in service | Write to outbox first, scheduler publishes |
| Duplicate handling | Usually ignored | SagaId idempotency check + unique constraint |
| Listener | Single message handler | Batch handler with partition/offset headers |
| Callback | Usually ignored | Callback updates outbox status (COMPLETED/FAILED) |
| Domain coupling | Listener has business logic | Listener delegates to input port (interface), zero logic |
| Consumer group | Single | Separate consumer groups per topic |

---

## Message Flow Summary

```
Order service writes to outbox (DB)
    │
    │ [PaymentOutboxScheduler - every 2s]
    ▼
OrderPaymentEventKafkaPublisher.publish()
    ├── deserialize JSON payload from outbox row
    ├── map to PaymentRequestAvroModel
    ├── kafkaProducer.send(topicName, sagaId, avroModel, callback)
    └── on Kafka ack: callback sets outboxStatus=COMPLETED
    │
    │ [Kafka: payment-request topic]
    ▼
PaymentRequestKafkaListener.receive()
    ├── map PaymentRequestAvroModel → PaymentRequest DTO
    └── paymentRequestMessageListener.completePayment(paymentRequest)
            │
            ▼
    PaymentRequestMessageListenerImpl.completePayment()
            │
            ▼
    PaymentRequestHelper.persistPayment()
            ├── Load CreditEntry + CreditHistory from DB
            ├── paymentDomainService.validateAndInitiatePayment()
            ├── Save payment + updated credit to DB
            └── Save to OrderOutboxMessage (payment-service's outbox)
                    │
                    │ [payment-service OrderOutboxScheduler]
                    ▼
            PaymentEventKafkaPublisher → payment-response topic
                    │
                    ▼
    OrderService PaymentResponseKafkaListener
            │
            ▼
    OrderPaymentSaga.process() or .rollback()
```

---

## Related Notes

- [[saga-pattern-choreography-based]] — how CQRS read models feed into SAGA validation
- [[transactional-outbox-pattern]] — the outbox that decouples writing from publishing
