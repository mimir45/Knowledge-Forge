---
title: DDD — Aggregates, Value Objects, Domain Events, Domain Services
slug: ddd-aggregates-value-objects-domain-events-domain-services
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
# DDD — Aggregates, Value Objects, Domain Events, Domain Services

> Part of [[food-ordering-system-course-architecture-index]] — Food Ordering System course notes.

---

## Core DDD Building Blocks

| Concept | Identity | Mutability | Purpose |
|---------|----------|------------|---------|
| Value Object | By value (no ID) | Immutable | Money, Address, IDs |
| Entity | By ID | Mutable | OrderItem, Product |
| Aggregate Root | By ID | Mutable, guards invariants | Order, Payment, Restaurant |
| Domain Service | N/A | Stateless | Multi-aggregate operations |
| Domain Event | N/A | Immutable | Signal state transitions |

---

## 1. Value Objects

### BaseId — Type-safe wrappers around UUID

Instead of passing `UUID` everywhere (easy to mix up `customerId` and `restaurantId`), each ID is
its own type:

```java
// common-domain/valueobject/BaseId.java
public abstract class BaseId<T> {
    private final T value;  // always final — immutable

    protected BaseId(T value) { this.value = value; }

    public T getValue() { return value; }

    @Override
    public boolean equals(Object o) {
        if (o == null || getClass() != o.getClass()) return false;
        return value.equals(((BaseId<?>) o).value);  // equality by value, not identity
    }
}

// Concrete ID types:
public class OrderId    extends BaseId<UUID> { public OrderId(UUID v)    { super(v); } }
public class CustomerId extends BaseId<UUID> { public CustomerId(UUID v) { super(v); } }
public class ProductId  extends BaseId<UUID> { public ProductId(UUID v)  { super(v); } }
```

**Why not just use UUID?**  
A method signature `void ship(UUID orderId, UUID customerId)` — easy to swap arguments.  
A method signature `void ship(OrderId orderId, CustomerId customerId)` — compiler catches the swap.

### Money — Behavior-rich Value Object

```java
// common-domain/valueobject/Money.java
public class Money {
    private final BigDecimal amount;  // final — immutable

    public static final Money ZERO = new Money(BigDecimal.ZERO);

    // Returns NEW instance — never mutates this
    public Money add(Money money) {
        return new Money(setScale(this.amount.add(money.getAmount())));
    }

    public Money subtract(Money money) {
        return new Money(setScale(this.amount.subtract(money.getAmount())));
    }

    public Money multiply(int multiplier) {
        return new Money(setScale(this.amount.multiply(new BigDecimal(multiplier))));
    }

    public boolean isGreaterThanZero() {
        return this.amount != null && this.amount.compareTo(BigDecimal.ZERO) > 0;
    }

    public boolean isGreaterThan(Money money) {
        return this.amount != null && this.amount.compareTo(money.getAmount()) > 0;
    }

    private BigDecimal setScale(BigDecimal input) {
        return input.setScale(2, RoundingMode.HALF_EVEN);  // banker's rounding, always 2dp
    }
}
```

**Usage in payment domain service — readable business logic:**

```java
private void validateCreditEntry(Payment payment, CreditEntry creditEntry, List<String> failureMessages) {
    if (payment.getPrice().isGreaterThan(creditEntry.getTotalCreditAmount())) {
        failureMessages.add("Customer doesn't have enough credit for payment!");
    }
}
```

vs. the "usual" way:
```java
// Ugly, error-prone, no encapsulation of rounding:
if (payment.getPrice().compareTo(creditEntry.getTotalCredit()) > 0) { ... }
```

### StreetAddress — Value Object With Own Equality

```java
public class StreetAddress {
    private final UUID id;
    private final String street;
    private final String postalCode;
    private final String city;

    // Equality ignores id — two addresses at same street/postal/city are "equal"
    @Override
    public boolean equals(Object o) {
        StreetAddress that = (StreetAddress) o;
        return street.equals(that.street)
            && postalCode.equals(that.postalCode)
            && city.equals(that.city);
        // NOTE: id is intentionally excluded from equals!
    }
}
```

---

## 2. Entities

`OrderItem` is an entity (has an ID), but not an aggregate root — it can only be accessed through `Order`:

```java
// order-domain-core/entity/OrderItem.java
public class OrderItem extends BaseEntity<OrderItemId> {
    private OrderId orderId;          // parent reference
    private final Product product;
    private final int quantity;
    private final Money price;
    private final Money subTotal;

    // Package-private — only Order can call this!
    void initializeOrderItem(OrderId orderId, OrderItemId orderItemId) {
        this.orderId = orderId;
        super.setId(orderItemId);
    }

    // Business validation — price * quantity must equal subTotal
    boolean isPriceValid() {
        return price.isGreaterThanZero()
            && price.equals(product.getPrice())          // matches restaurant's price
            && price.multiply(quantity).equals(subTotal); // subtotal is correct
    }
}
```

`Product` is a mutable entity. Its name/price get updated when confirmed by the restaurant:

```java
// entity/Product.java
public class Product extends BaseEntity<ProductId> {
    private String name;
    private Money price;

    public void updateWithConfirmedNameAndPrice(String name, Money price) {
        this.name = name;
        this.price = price;
    }
}
```

---

## 3. Aggregate Root — The Order

The aggregate root **owns** all invariants. External code cannot reach inside to change state.

```java
// entity/Order.java — extends AggregateRoot<OrderId>
public class Order extends AggregateRoot<OrderId> {
    private final CustomerId customerId;
    private final RestaurantId restaurantId;
    private final StreetAddress deliveryAddress;
    private final Money price;
    private final List<OrderItem> items;

    // Mutable state — controlled by methods below:
    private TrackingId trackingId;
    private OrderStatus orderStatus;
    private List<String> failureMessages;

    // === STATE MACHINE METHODS ===

    public void initializeOrder() {
        setId(new OrderId(UUID.randomUUID()));
        trackingId = new TrackingId(UUID.randomUUID());
        orderStatus = OrderStatus.PENDING;
        initializeOrderItems();           // assigns IDs 1, 2, 3... to items
    }

    public void validateOrder() {
        validateInitialOrder();   // must not already have ID
        validateTotalPrice();     // price > 0
        validateItemsPrice();     // sum of items == total price
    }

    public void pay() {
        if (orderStatus != OrderStatus.PENDING)
            throw new OrderDomainException("Order is not in correct state for pay!");
        orderStatus = OrderStatus.PAID;
    }

    public void approve() {
        if (orderStatus != OrderStatus.PAID)
            throw new OrderDomainException("Order is not in correct state for approve!");
        orderStatus = OrderStatus.APPROVED;
    }

    public void initCancel(List<String> failureMessages) {
        if (orderStatus != OrderStatus.PAID)
            throw new OrderDomainException("Order is not in correct state for initCancel!");
        orderStatus = OrderStatus.CANCELLING;
        updateFailureMessages(failureMessages);
    }

    public void cancel(List<String> failureMessages) {
        if (!(orderStatus == OrderStatus.CANCELLING || orderStatus == OrderStatus.PENDING))
            throw new OrderDomainException("Order is not in correct state for cancel!");
        orderStatus = OrderStatus.CANCELLED;
        updateFailureMessages(failureMessages);
    }

    // === INTERNAL VALIDATION ===

    private void validateItemsPrice() {
        // Uses Money::add — functional fold over items
        Money orderItemsTotal = items.stream()
            .map(orderItem -> {
                validateItemPrice(orderItem);  // each item's price must be valid
                return orderItem.getSubTotal();
            })
            .reduce(Money.ZERO, Money::add);

        if (!price.equals(orderItemsTotal)) {
            throw new OrderDomainException("Total price: " + price.getAmount()
                + " is not equal to Order items total: " + orderItemsTotal.getAmount() + "!");
        }
    }

    private void initializeOrderItems() {
        long itemId = 1;
        for (OrderItem orderItem : items) {
            orderItem.initializeOrderItem(super.getId(), new OrderItemId(itemId++));
        }
    }
}
```

**Order State Machine:**

```
null (not created)
    │  initializeOrder() + validateOrder()
    ▼
PENDING ──[pay()]──▶ PAID ──[approve()]──▶ APPROVED ✓
    │                  │
    │                  │ [initCancel()] → CANCELLING ──[cancel()]──▶ CANCELLED ✗
    │
    └──[cancel()]──▶ CANCELLED ✗  (direct cancel from PENDING)
```

---

## 4. Domain Services

When a business operation touches **multiple aggregates**, it goes in a domain service.
Domain services are plain Java — no Spring, no DB.

```java
// OrderDomainServiceImpl.java — no @Service annotation, no Spring imports
@Slf4j
public class OrderDomainServiceImpl implements OrderDomainService {

    @Override
    public OrderCreatedEvent validateAndInitiateOrder(Order order, Restaurant restaurant) {
        validateRestaurant(restaurant);            // touches Restaurant aggregate
        setOrderProductInformation(order, restaurant); // enriches Order from Restaurant
        order.validateOrder();                     // delegates to Order aggregate
        order.initializeOrder();                   // delegates to Order aggregate
        log.info("Order with id: {} is initiated", order.getId().getValue());
        return new OrderCreatedEvent(order, ZonedDateTime.now(ZoneId.of(UTC)));
    }

    @Override
    public OrderPaidEvent payOrder(Order order) {
        order.pay();
        log.info("Order with id: {} is paid", order.getId().getValue());
        return new OrderPaidEvent(order, ZonedDateTime.now(ZoneId.of(UTC)));
    }

    @Override
    public OrderCancelledEvent cancelOrderPayment(Order order, List<String> failureMessages) {
        order.initCancel(failureMessages);
        log.info("Order payment is cancelling for order id: {}", order.getId().getValue());
        return new OrderCancelledEvent(order, ZonedDateTime.now(ZoneId.of(UTC)));
    }

    private void validateRestaurant(Restaurant restaurant) {
        if (!restaurant.isActive())
            throw new OrderDomainException(
                "Restaurant with id " + restaurant.getId().getValue() + " is currently not active!");
    }

    // Enrich order items with confirmed product name and price from restaurant's product catalogue
    private void setOrderProductInformation(Order order, Restaurant restaurant) {
        order.getItems().forEach(orderItem ->
            restaurant.getProducts().forEach(restaurantProduct -> {
                Product currentProduct = orderItem.getProduct();
                if (currentProduct.equals(restaurantProduct)) { // equals() by ProductId
                    currentProduct.updateWithConfirmedNameAndPrice(
                        restaurantProduct.getName(),
                        restaurantProduct.getPrice()
                    );
                }
            })
        );
    }
}
```

**Payment Domain Service — validates credit balance with full audit trail:**

```java
// PaymentDomainServiceImpl.java
@Override
public PaymentEvent validateAndInitiatePayment(Payment payment,
                                               CreditEntry creditEntry,
                                               List<CreditHistory> creditHistories,
                                               List<String> failureMessages) {
    payment.validatePayment(failureMessages);           // price > 0
    payment.initializePayment();                        // assign ID + timestamp
    validateCreditEntry(payment, creditEntry, failureMessages);  // enough credit?
    subtractCreditEntry(payment, creditEntry);          // deduct from balance
    updateCreditHistory(payment, creditHistories, TransactionType.DEBIT); // record transaction
    validateCreditHistory(creditEntry, creditHistories, failureMessages); // double-check audit

    if (failureMessages.isEmpty()) {
        payment.updateStatus(PaymentStatus.COMPLETED);
        return new PaymentCompletedEvent(payment, ZonedDateTime.now(ZoneId.of(UTC)));
    } else {
        payment.updateStatus(PaymentStatus.FAILED);
        return new PaymentFailedEvent(payment, ZonedDateTime.now(ZoneId.of(UTC)), failureMessages);
    }
}

private void validateCreditHistory(CreditEntry creditEntry,
                                   List<CreditHistory> creditHistories,
                                   List<String> failureMessages) {
    Money totalCredit = getTotalHistoryAmount(creditHistories, TransactionType.CREDIT);
    Money totalDebit  = getTotalHistoryAmount(creditHistories, TransactionType.DEBIT);

    // Current balance must equal total credits minus total debits across all history
    if (!creditEntry.getTotalCreditAmount().equals(totalCredit.subtract(totalDebit))) {
        failureMessages.add("Credit history total is not equal to current credit!");
    }
}

private Money getTotalHistoryAmount(List<CreditHistory> histories, TransactionType type) {
    return histories.stream()
        .filter(h -> type == h.getTransactionType())
        .map(CreditHistory::getAmount)
        .reduce(Money.ZERO, Money::add);  // Money::add returns new Money each time
}
```

---

## 5. Domain Events

Events are **return values** from domain service methods — not side effects.
The domain itself never publishes events; the application service decides when/how.

```java
// common-domain/event/DomainEvent.java — just a marker interface
public interface DomainEvent<T> { }

// order-domain-core/event/OrderCreatedEvent.java
public class OrderCreatedEvent implements DomainEvent<Order> {
    private final Order order;
    private final ZonedDateTime createdAt;

    public OrderCreatedEvent(Order order, ZonedDateTime createdAt) {
        this.order = order;
        this.createdAt = createdAt;
    }

    public Order getOrder() { return order; }
    public ZonedDateTime getCreatedAt() { return createdAt; }
}
```

How the event is used (application service layer):

```java
// OrderCreateCommandHandler.java
@Transactional
public CreateOrderResponse createOrder(CreateOrderCommand createOrderCommand) {
    // Domain service RETURNS an event object — doesn't publish anything
    OrderCreatedEvent orderCreatedEvent = orderCreateHelper.persistOrder(createOrderCommand);

    // Application service decides to write it to outbox (same transaction)
    paymentOutboxHelper.savePaymentOutboxMessage(
        orderDataMapper.orderCreatedEventToOrderPaymentEventPayload(orderCreatedEvent),
        orderCreatedEvent.getOrder().getOrderStatus(),
        orderSagaHelper.orderStatusToSagaStatus(orderCreatedEvent.getOrder().getOrderStatus()),
        OutboxStatus.STARTED,
        UUID.randomUUID()  // new sagaId for this saga run
    );

    return createOrderResponse;
}
```

---

## 6. Builder Pattern on All Aggregates

All aggregates use inner `Builder` classes. Constructor is private — only `Builder.build()` creates instances.

```java
// Usage when creating from a command:
Order order = Order.builder()
    .customerId(new CustomerId(command.getCustomerId()))
    .restaurantId(new RestaurantId(command.getRestaurantId()))
    .deliveryAddress(new StreetAddress(UUID.randomUUID(),
        command.getAddress().getStreet(),
        command.getAddress().getPostalCode(),
        command.getAddress().getCity()))
    .price(new Money(command.getPrice()))
    .items(orderItemsFromCommand(command.getItems()))
    .build();
// At this point: id=null, trackingId=null, orderStatus=null
// These are set by order.initializeOrder() inside the domain service
```

**Pattern:** Build with known data → domain service validates and initializes → domain returns event.

---

## Common Gotchas

1. **`OrderItem.initializeOrderItem()` is package-private** — only `Order.initializeOrderItems()` can call it.
   This enforces that items are always initialized through their aggregate root.

2. **`OrderApplicationServiceImpl` is package-private** — the REST layer can only access it via the
   `OrderApplicationService` interface. Prevents bypassing the port abstraction.

3. **`OrderDomainServiceImpl` has NO `@Service`** — it's wired manually in `BeanConfiguration`.
   This is deliberate: the domain-core module has no Spring dependency.

---

## Related Notes

- [[hexagonal-architecture-ports-and-adapters]] — how these domain objects flow through ports and adapters
- [[saga-pattern-choreography-based]] — how domain events trigger SAGA steps
