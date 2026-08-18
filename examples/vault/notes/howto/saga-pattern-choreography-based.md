---
title: SAGA Pattern (Choreography-based)
slug: saga-pattern-choreography-based
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
# SAGA Pattern (Choreography-based)

> Part of [[food-ordering-system-course-architecture-index]] — Food Ordering System course notes.

---

## The Problem: Distributed Transactions

In a monolith, a single `@Transactional` method spans order creation + payment + approval.
If any step fails, the whole thing rolls back. Simple.

In microservices, there's no shared database. You can't span a transaction across services.

**2-Phase Commit (2PC)** would work but:
- Requires a coordinator that knows all participants
- Locks resources across services during prepare/commit phases
- Kills availability — if coordinator dies, everything stalls

---

## SAGA Solution

Break the distributed transaction into **local transactions**, each with a **compensating action**:

```
Step 1: Order service creates order (local tx) → publishes payment-request
Step 2: Payment service processes payment (local tx) → publishes payment-response
Step 3: Order service receives response → publishes restaurant-approval-request
Step 4: Restaurant service approves (local tx) → publishes approval-response
Step 5: Order service marks APPROVED ✓

If Step 4 FAILS (restaurant rejects):
  Compensate Step 2: publish payment-cancel-request
  Payment service refunds credit → order marked CANCELLED
```

---

## The `SagaStep` Interface

```java
// infrastructure/saga/SagaStep.java
public interface SagaStep<T> {
    void process(T data);    // execute the forward step
    void rollback(T data);   // compensating transaction
}
```

Two implementations in order-service:

| Class | Type param | Forward | Compensate |
|-------|-----------|---------|------------|
| `OrderPaymentSaga` | `PaymentResponse` | mark PAID + write approval outbox | mark CANCELLED |
| `OrderApprovalSaga` | `RestaurantApprovalResponse` | mark APPROVED | initCancel + write payment-cancel outbox |

---

## SagaStatus — Tracks Saga Lifecycle

```java
// infrastructure/saga/SagaStatus.java
public enum SagaStatus {
    STARTED,       // saga created, waiting for first response
    PROCESSING,    // payment done, waiting for restaurant approval
    SUCCEEDED,     // fully approved ✓
    COMPENSATING,  // something failed, compensations in progress
    COMPENSATED,   // fully rolled back ✗
    FAILED         // rollback also failed (needs manual intervention)
}
```

The mapping from `OrderStatus` to `SagaStatus`:

```java
// OrderSagaHelper.java
SagaStatus orderStatusToSagaStatus(OrderStatus orderStatus) {
    return switch (orderStatus) {
        case PAID      -> SagaStatus.PROCESSING;   // paid but not yet approved
        case APPROVED  -> SagaStatus.SUCCEEDED;    // done ✓
        case CANCELLING -> SagaStatus.COMPENSATING; // trying to rollback
        case CANCELLED -> SagaStatus.COMPENSATED;  // rolled back ✗
        default        -> SagaStatus.STARTED;      // PENDING
    };
}
```

---

## OrderPaymentSaga — Full Flow

### `process()` — Called when payment-service says COMPLETED

```java
@Component
public class OrderPaymentSaga implements SagaStep<PaymentResponse> {

    @Override
    @Transactional
    public void process(PaymentResponse paymentResponse) {

        // === IDEMPOTENCY CHECK ===
        // Only process if outbox message is in STARTED state for this sagaId
        Optional<OrderPaymentOutboxMessage> outboxMsgOpt =
            paymentOutboxHelper.getPaymentOutboxMessageBySagaIdAndSagaStatus(
                UUID.fromString(paymentResponse.getSagaId()),
                SagaStatus.STARTED   // ← guard: only process once
            );

        if (outboxMsgOpt.isEmpty()) {
            // Duplicate message — already processed. Silent no-op.
            log.info("Already processed saga id: {}", paymentResponse.getSagaId());
            return;
        }

        // === DOMAIN TRANSITION ===
        OrderPaidEvent domainEvent = completePaymentForOrder(paymentResponse);
        // ^ loads order, calls orderDomainService.payOrder(order) which calls order.pay()
        //   order.pay() validates PENDING → PAID and throws if wrong state

        // === OUTBOX UPDATES (same transaction!) ===
        SagaStatus sagaStatus = orderSagaHelper.orderStatusToSagaStatus(
            domainEvent.getOrder().getOrderStatus()
        );

        // 1. Update payment outbox: STARTED → PROCESSING
        paymentOutboxHelper.save(
            getUpdatedPaymentOutboxMessage(outboxMsgOpt.get(),
                domainEvent.getOrder().getOrderStatus(), sagaStatus)
        );

        // 2. Write new approval outbox message: STARTED (triggers approval request publish)
        approvalOutboxHelper.saveApprovalOutboxMessage(
            orderDataMapper.orderPaidEventToOrderApprovalEventPayload(domainEvent),
            domainEvent.getOrder().getOrderStatus(),
            sagaStatus,
            OutboxStatus.STARTED,
            UUID.fromString(paymentResponse.getSagaId())  // same sagaId threads through all steps
        );
    }

    private OrderPaidEvent completePaymentForOrder(PaymentResponse paymentResponse) {
        Order order = orderSagaHelper.findOrder(paymentResponse.getOrderId());
        OrderPaidEvent domainEvent = orderDomainService.payOrder(order);
        orderRepository.save(order);  // persist state change
        return domainEvent;
    }
```

### `rollback()` — Called when payment-service says FAILED or CANCELLED

```java
    @Override
    @Transactional
    public void rollback(PaymentResponse paymentResponse) {

        // Different idempotency guard — check for STARTED or PROCESSING depending on why we're rolling back
        Optional<OrderPaymentOutboxMessage> outboxMsgOpt =
            paymentOutboxHelper.getPaymentOutboxMessageBySagaIdAndSagaStatus(
                UUID.fromString(paymentResponse.getSagaId()),
                getCurrentSagaStatus(paymentResponse.getPaymentStatus())
            );

        if (outboxMsgOpt.isEmpty()) {
            log.info("Already rolled back saga id: {}", paymentResponse.getSagaId());
            return;
        }

        Order order = rollbackPaymentForOrder(paymentResponse);
        // ^ calls orderDomainService.cancelOrder(order, failureMessages)
        //   which calls order.cancel() — validates CANCELLING or PENDING → CANCELLED

        SagaStatus sagaStatus = orderSagaHelper.orderStatusToSagaStatus(order.getOrderStatus());

        paymentOutboxHelper.save(
            getUpdatedPaymentOutboxMessage(outboxMsgOpt.get(), order.getOrderStatus(), sagaStatus)
        );

        // Only update approval outbox if this was a mid-saga cancellation (CANCELLED status)
        if (paymentResponse.getPaymentStatus() == PaymentStatus.CANCELLED) {
            approvalOutboxHelper.save(
                getUpdatedApprovalOutboxMessage(paymentResponse.getSagaId(),
                    order.getOrderStatus(), sagaStatus)
            );
        }
    }

    // Which saga statuses are valid for rollback, based on why payment failed?
    private SagaStatus[] getCurrentSagaStatus(PaymentStatus paymentStatus) {
        return switch (paymentStatus) {
            case COMPLETED -> new SagaStatus[] { SagaStatus.STARTED };      // failed at initial payment
            case CANCELLED -> new SagaStatus[] { SagaStatus.PROCESSING };   // payment refunded after restaurant rejected
            case FAILED    -> new SagaStatus[] { SagaStatus.STARTED, SagaStatus.PROCESSING };
        };
    }
}
```

---

## OrderApprovalSaga — Restaurant Approval Step

```java
@Component
public class OrderApprovalSaga implements SagaStep<RestaurantApprovalResponse> {

    @Override
    @Transactional
    public void process(RestaurantApprovalResponse response) {
        // Idempotency: only process if saga is in PROCESSING state
        Optional<OrderApprovalOutboxMessage> outboxMsgOpt =
            approvalOutboxHelper.getApprovalOutboxMessageBySagaIdAndSagaStatus(
                UUID.fromString(response.getSagaId()),
                SagaStatus.PROCESSING   // ← must be PROCESSING (payment was already done)
            );

        if (outboxMsgOpt.isEmpty()) { return; }

        // Approve the order: PAID → APPROVED
        Order order = approveOrder(response);

        SagaStatus sagaStatus = orderSagaHelper.orderStatusToSagaStatus(order.getOrderStatus());

        // Update both outbox messages to terminal SUCCEEDED state
        approvalOutboxHelper.save(getUpdatedApprovalOutboxMessage(outboxMsgOpt.get(),
            order.getOrderStatus(), sagaStatus));
        paymentOutboxHelper.save(getUpdatedPaymentOutboxMessage(response.getSagaId(),
            order.getOrderStatus(), sagaStatus));

        log.info("Order with id: {} is approved ✓", order.getId().getValue());
    }

    @Override
    @Transactional
    public void rollback(RestaurantApprovalResponse response) {
        // Restaurant rejected — we need to refund payment
        // 1. Move order to CANCELLING
        OrderCancelledEvent domainEvent = rollbackOrder(response);
        // ^ calls orderDomainService.cancelOrderPayment(order, failureMessages)
        //   which calls order.initCancel() — PAID → CANCELLING

        SagaStatus sagaStatus = orderSagaHelper.orderStatusToSagaStatus(
            domainEvent.getOrder().getOrderStatus()
        );

        approvalOutboxHelper.save(getUpdatedApprovalOutboxMessage(outboxMsgOpt.get(),
            domainEvent.getOrder().getOrderStatus(), sagaStatus));

        // 2. Write new payment outbox message to trigger payment cancellation
        //    This sends a payment-cancel-request to payment-service
        paymentOutboxHelper.savePaymentOutboxMessage(
            orderDataMapper.orderCancelledEventToOrderPaymentEventPayload(domainEvent),
            domainEvent.getOrder().getOrderStatus(),
            sagaStatus,
            OutboxStatus.STARTED,   // scheduler will publish this to Kafka
            UUID.fromString(response.getSagaId())
        );
    }
}
```

---

## How Kafka Connects to the SAGA

The Kafka listener in the messaging module calls the domain input port, which delegates to the saga:

```java
// order-messaging/listener/kafka/PaymentResponseKafkaListener.java
@KafkaListener(topics = "${order-service.payment-response-topic-name}")
public void receive(List<PaymentResponseAvroModel> messages, ...) {
    messages.forEach(avroModel -> {
        if (PaymentStatus.COMPLETED == avroModel.getPaymentStatus()) {
            paymentResponseMessageListener.paymentCompleted(
                messagingMapper.paymentResponseAvroModelToPaymentResponse(avroModel));
        } else if (PaymentStatus.CANCELLED == avroModel.getPaymentStatus()
                || PaymentStatus.FAILED == avroModel.getPaymentStatus()) {
            paymentResponseMessageListener.paymentCancelled(
                messagingMapper.paymentResponseAvroModelToPaymentResponse(avroModel));
        }
    });
}

// PaymentResponseMessageListenerImpl.java (application-service layer)
@Service
public class PaymentResponseMessageListenerImpl implements PaymentResponseMessageListener {

    @Override
    public void paymentCompleted(PaymentResponse paymentResponse) {
        orderPaymentSaga.process(paymentResponse);   // → SagaStep.process()
    }

    @Override
    public void paymentCancelled(PaymentResponse paymentResponse) {
        orderPaymentSaga.rollback(paymentResponse);  // → SagaStep.rollback()
    }
}
```

**The indirection:** Kafka listener → input port interface → SAGA.  
This means the SAGA can be tested with no Kafka — just call `paymentResponseMessageListener.paymentCompleted(...)` directly.

---

## Full SAGA State Chart

```
[HTTP POST /orders]
    │
    ▼
Order: PENDING | PaymentOutbox: STARTED (SagaStatus: STARTED)
    │
    │ [PaymentOutboxScheduler publishes to Kafka]
    ▼
PaymentService processes → PaymentOutbox (in payment-service): STARTED
    │
    │ [Payment's OutboxScheduler publishes payment-response to Kafka]
    ▼
OrderService receives payment-response COMPLETED
    │
    ▼
Order: PAID | PaymentOutbox: PROCESSING | ApprovalOutbox: STARTED (SagaStatus: PROCESSING)
    │
    │ [ApprovalOutboxScheduler publishes to Kafka]
    ▼
RestaurantService processes → RestaurantOutbox: STARTED
    │
    ├─── APPROVED ────────────────────────────────────────────────────────┐
    │                                                                     ▼
    │                                             Order: APPROVED | Both outboxes: SUCCEEDED ✓
    │
    └─── REJECTED ──────────────────────────────────────────────────────┐
                                                                         ▼
                                          Order: CANCELLING | ApprovalOutbox: COMPENSATING
                                          PaymentOutbox: STARTED (new — cancel request)
                                                    │
                                                    │ [PaymentOutboxScheduler publishes cancel]
                                                    ▼
                                          PaymentService refunds credit
                                                    │
                                                    │ [payment-response: CANCELLED]
                                                    ▼
                                          Order: CANCELLED | Both outboxes: COMPENSATED ✗
```

---

## Why Choreography (Not Orchestration)?

This project uses **choreography** — services react to events without a central coordinator.

**Orchestration** would have one service (Order) telling others what to do via commands.
**Choreography** has each service listen to events and decide what to do independently.

Trade-offs:

| | Choreography (this project) | Orchestration |
|-|---------------------------|--------------|
| Coupling | Low — services don't know each other | Higher — orchestrator knows all participants |
| Complexity | Distributed across services | Centralized, easier to follow |
| Scalability | Better | Can become a bottleneck |
| Debugging | Harder — need to trace across topics | Easier — one place to check |

---

## Related Notes

- [[transactional-outbox-pattern]] — how SAGA steps write to the outbox to guarantee message delivery
- [[ddd-aggregates-value-objects-domain-events-domain-services]] — the domain events the SAGA uses
