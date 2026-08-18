---
title: Hexagonal Architecture (Ports & Adapters)
slug: hexagonal-architecture-ports-and-adapters
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
# Hexagonal Architecture (Ports & Adapters)

> Part of [[food-ordering-system-course-architecture-index]] — Food Ordering System course notes.

---

## The Problem With Layered Architecture

In a typical Spring Boot app:

```java
// OrderService depends directly on JPA repo
@Service
public class OrderService {
    @Autowired
    private OrderJpaRepository orderRepo;  // ← coupled to Spring Data JPA

    @Autowired
    private KafkaTemplate<String, ?> kafka; // ← coupled to Kafka
}
```

Swap the DB? Rewrite `OrderService`. Add a test that uses an in-memory store? Impossible without Spring context. The business logic is **glued** to infrastructure.

---

## Hexagonal Architecture Solution

The **domain defines interfaces (ports)** for everything it needs.  
Infrastructure **implements those interfaces (adapters)**.  
The domain has **zero infrastructure imports**.

```
┌─────────────────────────────────────────────────────────┐
│                        DOMAIN                           │
│                                                         │
│   [Input Port interface] ←── Application Service        │
│                               calls →                   │
│   [Domain Service / Aggregates]                         │
│                               needs →                   │
│   [Output Port interface] ──→ (implemented by adapter)  │
└─────────────────────────────────────────────────────────┘
          ↑ implemented by          ↑ implemented by
   REST Controller adapter     JPA/Kafka adapter
```

---

## Layer-by-Layer With Real Code

### 1. Input Port — What the domain exposes outward

Defined in `order-domain/application-service`. Called by REST controller.

```java
// ports/input/service/OrderApplicationService.java
public interface OrderApplicationService {
    CreateOrderResponse createOrder(@Valid CreateOrderCommand createOrderCommand);
    TrackOrderResponse trackOrder(@Valid TrackOrderQuery trackOrderQuery);
}
```

The REST controller only knows about this interface — never the implementation:

```java
// order-application/rest/OrderController.java
@RestController
@RequestMapping(value = "/orders", produces = MediaType.APPLICATION_JSON_VALUE)
public class OrderController {

    private final OrderApplicationService orderApplicationService; // ← interface, not impl

    public OrderController(OrderApplicationService orderApplicationService) {
        this.orderApplicationService = orderApplicationService;
    }

    @PostMapping
    public ResponseEntity<CreateOrderResponse> createOrder(
            @RequestBody CreateOrderCommand createOrderCommand) {
        log.info("Creating order for customer: {} at restaurant: {}",
                createOrderCommand.getCustomerId(),
                createOrderCommand.getRestaurantId());
        return ResponseEntity.ok(orderApplicationService.createOrder(createOrderCommand));
    }
}
```

### 2. Input Port Implementation — Lives in domain, package-private

```java
// OrderApplicationServiceImpl.java — note: package-private class!
@Slf4j
@Validated
@Service
class OrderApplicationServiceImpl implements OrderApplicationService {

    private final OrderCreateCommandHandler orderCreateCommandHandler;
    private final OrderTrackCommandHandler orderTrackCommandHandler;

    @Override
    public CreateOrderResponse createOrder(CreateOrderCommand createOrderCommand) {
        return orderCreateCommandHandler.createOrder(createOrderCommand);
    }

    @Override
    public TrackOrderResponse trackOrder(TrackOrderQuery trackOrderQuery) {
        return orderTrackCommandHandler.trackOrder(trackOrderQuery);
    }
}
```

> **Why package-private?** Prevents other modules from accessing the implementation directly.  
> They must go through the `OrderApplicationService` interface.

### 3. Output Port — What the domain needs from infrastructure

```java
// ports/output/repository/OrderRepository.java
public interface OrderRepository {
    Order save(Order order);
    Optional<Order> findById(OrderId orderId);
    Optional<Order> findByTrackingId(TrackingId trackingId);
}
```

Notice: uses **domain types** (`Order`, `OrderId`) — no JPA `@Entity` anywhere here.

```java
// ports/output/message/publisher/payment/PaymentRequestMessagePublisher.java
public interface PaymentRequestMessagePublisher {
    void publish(OrderPaymentOutboxMessage orderPaymentOutboxMessage,
                 BiConsumer<OrderPaymentOutboxMessage, OutboxStatus> outboxCallback);
}
```

### 4. Output Port Adapter — Infrastructure implements the port

```java
// order-dataaccess/order/adapter/OrderRepositoryImpl.java
@Component
public class OrderRepositoryImpl implements OrderRepository {

    private final OrderJpaRepository orderJpaRepository;       // Spring Data JPA
    private final OrderDataAccessMapper orderDataAccessMapper; // domain ↔ JPA mapper

    public OrderRepositoryImpl(OrderJpaRepository orderJpaRepository,
                               OrderDataAccessMapper orderDataAccessMapper) {
        this.orderJpaRepository = orderJpaRepository;
        this.orderDataAccessMapper = orderDataAccessMapper;
    }

    @Override
    public Order save(Order order) {
        // 1. Convert domain object to JPA entity
        // 2. Save via Spring Data
        // 3. Convert JPA entity back to domain object
        return orderDataAccessMapper.orderEntityToOrder(
            orderJpaRepository.save(
                orderDataAccessMapper.orderToOrderEntity(order)
            )
        );
    }

    @Override
    public Optional<Order> findById(OrderId orderId) {
        return orderJpaRepository
            .findById(orderId.getValue())           // unwrap value object to UUID
            .map(orderDataAccessMapper::orderEntityToOrder); // re-wrap on return
    }
}
```

The Kafka adapter implements the messaging output port:

```java
// order-messaging/.../publisher/kafka/OrderPaymentEventKafkaPublisher.java
@Component
public class OrderPaymentEventKafkaPublisher implements PaymentRequestMessagePublisher {

    @Override
    public void publish(OrderPaymentOutboxMessage orderPaymentOutboxMessage,
                        BiConsumer<OrderPaymentOutboxMessage, OutboxStatus> outboxCallback) {
        // ... deserialize payload, build Avro model, send to Kafka
        // on success: outboxCallback.accept(message, COMPLETED)
        // on failure: outboxCallback.accept(message, FAILED)
    }
}
```

### 5. Spring Wiring — Only in the container module

`domain-core` is plain Java — the `OrderDomainServiceImpl` has **no Spring annotations**.  
The container module creates the bean manually:

```java
// order-container/BeanConfiguration.java
@Configuration
public class BeanConfiguration {

    @Bean
    public OrderDomainService orderDomainService() {
        return new OrderDomainServiceImpl(); // manually instantiated
    }
}
```

**Why?** `OrderDomainServiceImpl` lives in `order-domain-core`. That module has no Spring dependency
in its `pom.xml` at all. It cannot have `@Service` even if you wanted to.

---

## The Mapper Layer (Anti-Corruption)

Every boundary has a dedicated mapper. Nothing leaks between layers:

```java
// OrderDataAccessMapper.java  — domain ↔ JPA
@Component
public class OrderDataAccessMapper {

    public OrderEntity orderToOrderEntity(Order order) {
        OrderEntity entity = OrderEntity.builder()
            .id(order.getId().getValue())            // UUID from value object
            .customerId(order.getCustomerId().getValue())
            .restaurantId(order.getRestaurantId().getValue())
            .trackingId(order.getTrackingId().getValue())
            .price(order.getPrice().getAmount())     // BigDecimal from Money
            .orderStatus(order.getOrderStatus())
            .failureMessages(order.getFailureMessages() != null
                ? String.join(",", order.getFailureMessages()) : "")  // List → String
            .build();
        return entity;
    }

    public Order orderEntityToOrder(OrderEntity entity) {
        return Order.builder()
            .orderId(new OrderId(entity.getId()))           // re-wrap UUID
            .customerId(new CustomerId(entity.getCustomerId()))
            .price(new Money(entity.getPrice()))            // re-wrap to Money
            .failureMessages(entity.getFailureMessages().isEmpty()
                ? new ArrayList<>()
                : new ArrayList<>(Arrays.asList(entity.getFailureMessages().split(","))))
            .build();
    }
}
```

---

## Dependency Direction Rule

```
REST ──▶ domain (input port interface)
domain ──▶ infra (output port interface defined in domain, adapter in infra)

NEVER: domain ──▶ @Entity / KafkaTemplate / @Repository
```

If you ever see an import like `import org.springframework.data.jpa.*` inside `order-domain-core`,
that is an architecture violation.

---

## Why It Matters

- **Testability:** Test the entire domain with no Spring, no Kafka, no DB. Inject mock adapters.
- **Swappability:** Replace PostgreSQL with Cassandra by writing a new adapter. Domain untouched.
- **Clear boundaries:** If your domain-core `pom.xml` has no Spring dependency, you cannot
  accidentally couple business logic to infrastructure.

---

## Related Notes

- [[ddd-aggregates-value-objects-domain-events-domain-services]] — what those domain types the ports use actually look like
- [[transactional-outbox-pattern]] — how the messaging output ports are used with outbox guarantee
