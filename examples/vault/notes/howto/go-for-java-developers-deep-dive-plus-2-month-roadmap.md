---
title: Go for Java Developers — Deep Dive + 2-Month Roadmap
slug: go-for-java-developers-deep-dive-plus-2-month-roadmap
type: howto
stack: [go, java]
tags: [backend, learning]
depth: 3
confidence: low
created: 2026-05-19
updated: 2026-08-09
verified: 2026-08-09
freshness_days: 180
sources:
  - url: go-for-java-devs.md
    accessed: 2026-05-19
    kind: session
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---

# Go for Java Developers — Deep Dive + 2-Month Roadmap

> **Target audience:** Java/Spring Boot backend developer wanting to understand Go deeply and reach mid-level competency.

---

## Table of Contents

1. [Why Go Exists](#1-why-go-exists)
2. [Type System & Syntax](#2-type-system--syntax)
3. [OOP vs Go's Composition Model](#3-oop-vs-gos-composition-model)
4. [Error Handling](#4-error-handling)
5. [Concurrency — Threads vs Goroutines](#5-concurrency)
6. [Memory & Runtime — JVM vs Go Runtime](#6-memory--runtime)
7. [Tooling & Ecosystem](#7-tooling--ecosystem)
8. [Frameworks — Spring Boot vs Go Alternatives](#8-frameworks)
9. [Generics (Go 1.18+)](#9-generics)
10. [Context Deep Dive](#10-context-deep-dive)
11. [Profiling & pprof](#11-profiling--pprof)
12. [Common Go Gotchas for Java Devs](#12-common-go-gotchas-for-java-devs)
13. [When Go Wins, When Java Wins](#13-when-go-wins-when-java-wins)
14. [2-Month Learning Roadmap](#14-2-month-learning-roadmap)

---

## 1. Why Go Exists

Go was created at Google in 2009 by Rob Pike, Ken Thompson, and Robert Griesemer. They were frustrated working with C++, Java, and Python on large-scale distributed systems. The core complaints were: **C++ compiled slowly**, **Java was too verbose with too much ceremony**, and **Python was too slow**. They wanted something that had C's performance profile but felt as productive as Python.

The philosophy is radical simplicity. Go deliberately has fewer features than Java. No generics for a long time (added in 1.18). No exceptions. No inheritance. No operator overloading. No implicit type conversions. Every design choice is about reducing cognitive load so a large team of engineers can read each other's code without context.

**Go is NOT a better Java.** It is a different tool optimized for different constraints — mostly networking, concurrency, and fast startup in containerized/cloud environments.

---

## 2. Type System & Syntax

### Static Typing Without the Boilerplate

Both Java and Go are statically typed, but Go's type inference is far more aggressive. You almost never write types explicitly in local scope.

```java
// Java — explicit everywhere
Map<String, List<Integer>> map = new HashMap<>();
String result = someMethod();
```

```go
// Go — := infers the type at compile time
map := make(map[string][]int)
result := someMethod()
```

The `:=` operator declares AND assigns. Once declared, you use `=`. No `var` keyword needed in most cases.

### No Null, But Zero Values

Java has `null`, which causes NullPointerExceptions. Go has **zero values** — every type has a sensible default when declared without assignment.

| Type | Zero Value |
|------|-----------|
| `int`, `float64` | `0` |
| `string` | `""` (empty string) |
| `bool` | `false` |
| `pointer`, `slice`, `map`, `channel` | `nil` |
| `struct` | all fields set to their zero values |

```go
// This is valid and safe — name is "" not null
var name string
fmt.Println(len(name)) // 0, no NPE
```

### Pointers — Yes, But Simpler Than C

Go has pointers, Java doesn't expose them. This matters because Go lets you explicitly control pass-by-value vs pass-by-reference.

```go
// Pass by value — original unchanged
func addTax(price float64) float64 {
    return price * 1.2
}

// Pass by pointer — original is mutated
func applyDiscount(price *float64) {
    *price *= 0.9
}

p := 100.0
applyDiscount(&p)
fmt.Println(p) // 90.0
```

In Java, primitives are always by value, objects are always by reference (technically by value of the reference). In Go, you choose explicitly — cleaner mental model.

### Arrays vs Slices

Java has arrays and `ArrayList`. Go has arrays (fixed size, rarely used) and **slices** (dynamic, backed by an array internally).

```go
// Slice — the bread and butter
prices := []float64{10.5, 20.0, 30.75}
prices = append(prices, 45.0)   // grows automatically

// Slice of slice — no copy, same backing array
sub := prices[1:3] // [20.0, 30.75]
```

> ⚠️ **Gotcha:** Slicing doesn't copy data. `sub` and `prices` share the same backing array. Modifying `sub[0]` modifies `prices[1]`. This bites everyone at least once.

### Maps

```go
// Declare and initialize
account := map[string]float64{
    "USD": 1000.0,
    "EUR": 850.0,
}

// Safe read — two-return idiom
balance, ok := account["GBP"]
if !ok {
    fmt.Println("currency not found")
}
```

In Java you'd call `map.containsKey()` or check for null. Go's two-return idiom (`value, ok`) is idiomatic and eliminates the null check entirely.

---

## 3. OOP vs Go's Composition Model

This is the biggest mental shift coming from Java.

**Go has no classes. No inheritance. No abstract classes.**

Instead: **structs + interfaces + composition**.

### Structs Replace Classes

```go
// Go struct — data container only
type BankAccount struct {
    ID      string
    Balance float64
    Owner   string
}

// Methods are defined outside the struct, attached via receiver
func (a *BankAccount) Deposit(amount float64) error {
    if amount <= 0 {
        return errors.New("amount must be positive")
    }
    a.Balance += amount
    return nil
}
```

The receiver `(a *BankAccount)` is equivalent to Java's `this`. Using a pointer receiver `*BankAccount` means the method can mutate the struct. Value receiver `BankAccount` gets a copy.

### Interfaces Are Implicit — No `implements`

In Java you declare `class Dog implements Animal`. In Go, **if your type has the methods, it satisfies the interface automatically**.

```go
// Define the interface
type Account interface {
    Deposit(amount float64) error
    Balance() float64
}

// BankAccount satisfies Account — no declaration needed
// as long as it has Deposit() and Balance() methods
func ProcessDeposit(acc Account, amount float64) {
    if err := acc.Deposit(amount); err != nil {
        log.Println("deposit failed:", err)
    }
}
```

This is called **structural typing** (vs Java's nominal typing). It makes Go code extremely easy to mock in tests — you just implement the interface in your test file without touching production code.

### Composition Over Inheritance

Go doesn't have `extends`. Instead you **embed** one struct into another.

```java
// Java inheritance
class SavingsAccount extends BankAccount {
    private double interestRate;
}
```

```go
// Go embedding (composition)
type SavingsAccount struct {
    BankAccount        // embedded — promotes all fields and methods
    InterestRate float64
}

s := SavingsAccount{
    BankAccount:  BankAccount{ID: "s001", Balance: 5000},
    InterestRate: 0.03,
}
s.Deposit(1000) // calls BankAccount.Deposit directly
```

Embedding is NOT inheritance. `SavingsAccount` is not a `BankAccount`. You can't assign a `SavingsAccount` to a `*BankAccount` variable. But all promoted methods are accessible directly. This forces you to think in terms of behavior composition, which leads to less coupled code.

---

## 4. Error Handling

Java uses checked and unchecked exceptions. Go has **no exceptions** — errors are values returned from functions.

```java
// Java — exception thrown, caught somewhere up the call stack
public Account findById(String id) throws AccountNotFoundException {
    ...
}
```

```go
// Go — error is a return value, caller MUST handle it
func FindById(id string) (*Account, error) {
    acc, ok := db[id]
    if !ok {
        return nil, fmt.Errorf("account %s not found", id)
    }
    return acc, nil
}

// Calling site — you cannot ignore errors silently
acc, err := FindById("acc-123")
if err != nil {
    return fmt.Errorf("FindById: %w", err) // wrap with context
}
```

The `%w` verb wraps an error so callers can unwrap it: `errors.Is(err, ErrNotFound)` works through the chain. This is equivalent to Java's exception cause chain.

### Custom Error Types

```go
type InsufficientFundsError struct {
    Available float64
    Requested float64
}

func (e *InsufficientFundsError) Error() string {
    return fmt.Sprintf("need %.2f but only %.2f available", e.Requested, e.Available)
}

// Check type with errors.As
var fundErr *InsufficientFundsError
if errors.As(err, &fundErr) {
    fmt.Printf("short by %.2f\n", fundErr.Requested-fundErr.Available)
}
```

### Sentinel Errors

```go
// Define package-level sentinel errors (like Java's checked exceptions)
var ErrNotFound = errors.New("account not found")
var ErrInsufficientFunds = errors.New("insufficient funds")

// Check with errors.Is (works through wrapped chains)
if errors.Is(err, ErrNotFound) {
    return http.StatusNotFound
}
```

### Panic / Recover — Not for Business Logic

Go does have `panic` (like an unchecked exception) and `recover` (like catch in a deferred function). But the Go convention is: **only panic on truly unrecoverable programmer errors** (nil dereference, index out of bounds). Use errors for expected failures. Never use panic as a control flow mechanism.

---

## 5. Concurrency

This is where Go genuinely beats Java for most use cases. It's the primary reason teams adopt Go.

### Java Concurrency Model

Java threads map 1:1 to OS threads. Each thread costs ~1MB of stack memory. Spawning 10,000 threads causes memory pressure and context-switching overhead. Java's solution: thread pools (`ExecutorService`), `CompletableFuture`, reactive streams (Project Reactor/WebFlux). These work but add significant complexity.

Java 21 added **Virtual Threads** (Project Loom) which gets much closer to Go, but Go had this from day one.

### Goroutines

A **goroutine** is a lightweight concurrent function. Initial stack is ~2KB (vs 1MB for OS thread). The Go runtime multiplexes thousands of goroutines onto a small pool of OS threads (M:N threading model).

```go
// Spawning 10,000 goroutines is totally normal
for i := 0; i < 10_000; i++ {
    go processPayment(paymentID[i])
}
```

This would OOM or thrash on Java without a thread pool. In Go, the runtime scheduler handles it efficiently.

### Channels — Communication Between Goroutines

Go's motto: **"Don't communicate by sharing memory; share memory by communicating."**

In Java, you share data between threads using synchronized blocks, locks, `volatile`, `AtomicInteger`, etc. Go encourages passing data through **channels**.

```go
// Channel is a typed pipe between goroutines
results := make(chan float64, 100) // buffered channel, capacity 100

go func() {
    total := calculateTotal(transactions)
    results <- total // send
}()

total := <-results // receive (blocks until data arrives)
fmt.Printf("Total: %.2f\n", total)
```

### Select — Multiplexing Channels

`select` is like a `switch` for channels — waits for whichever channel has data first.

```go
func processWithTimeout(job chan Job, done chan struct{}) {
    select {
    case j := <-job:
        handle(j)
    case <-time.After(5 * time.Second):
        fmt.Println("timeout, giving up")
    case <-done:
        fmt.Println("cancelled")
    }
}
```

In Java this would require `Future.get(timeout, TimeUnit.SECONDS)` + `CancellationToken` logic. Go's `select` is far more readable.

### sync Package — When You Do Need Shared State

Go also has mutexes, WaitGroups, and Once for cases where channels aren't the right fit.

```go
var mu sync.Mutex
var balance float64

func credit(amount float64) {
    mu.Lock()
    defer mu.Unlock() // defer = runs when function returns (like finally)
    balance += amount
}
```

`defer` is Go's `finally`. It always runs when the enclosing function returns, even on panic. Use it for cleanup, mutex unlocks, closing files.

### Worker Pool Pattern

```go
func workerPool(jobs <-chan Job, results chan<- Result, workerCount int) {
    var wg sync.WaitGroup
    for i := 0; i < workerCount; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- process(job)
            }
        }()
    }
    wg.Wait()
    close(results)
}
```

This replaces Java's `ExecutorService.submit()` loop. The range over a channel blocks until the channel is closed.

---

## 6. Memory & Runtime

### JVM vs Go Runtime

| Aspect | Java (JVM) | Go |
|--------|-----------|-----|
| Startup time | 200ms–2s (JVM warmup) | 5–50ms (no VM) |
| Memory baseline | 50–200MB (JVM overhead) | 5–30MB |
| Compilation | JIT at runtime | AOT to native binary |
| Deployment | JAR + JRE/JDK required | Single static binary |
| GC pauses | Can be long (tunable) | Sub-millisecond (concurrent GC) |
| Reflection | Full, powerful | Limited, slower |

The single binary deployment is a huge operational advantage for Docker/Kubernetes. Your Go Docker image can be `FROM scratch` — literally just the binary, ~10MB image vs a Spring Boot image that's typically 200–400MB.

### Go GC

Go uses a **concurrent, tri-color mark-and-sweep** GC with very low stop-the-world pauses (typically <1ms). Java's GC (G1, ZGC, Shenandoah) has gotten excellent but requires tuning (`-Xms`, `-Xmx`, GC flags). Go's GC is simpler and works well by default, though it trades throughput for low latency.

### Escape Analysis

Go's compiler performs **escape analysis** — if a variable doesn't outlive the function, it stays on the stack (fast). If it escapes to a goroutine or gets returned, it moves to the heap (GC managed). You don't control this directly, but you can inspect it:

```bash
go build -gcflags="-m" ./...
# prints: "x escapes to heap" or "x does not escape"
```

---

## 7. Tooling & Ecosystem

This is where Go is refreshingly minimal compared to the Java/Maven/Spring ecosystem.

### Built-in Tools (No Maven/Gradle Needed)

```bash
go mod init github.com/user/myapp   # init module (like pom.xml)
go get github.com/gin-gonic/gin     # add dependency
go build ./...                       # compile
go test ./...                        # run all tests
go fmt ./...                         # format code (no style debates)
go vet ./...                         # static analysis
go run main.go                       # run directly
```

`go fmt` is enforced universally — there are no tabs vs spaces debates in Go because the tool decides. This eliminates an entire category of team friction.

### go.mod — The pom.xml Equivalent

```
module github.com/myorg/myapp

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/jmoiron/sqlx v1.3.5
)
```

`go.sum` is the lock file (like `package-lock.json`). Both must be committed. No SNAPSHOT versions, no dependency hell — the module system is strict and reproducible by design.

### Testing — Built In

```go
// account_test.go
func TestDeposit(t *testing.T) {
    acc := &BankAccount{Balance: 100}
    err := acc.Deposit(50)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if acc.Balance != 150 {
        t.Errorf("expected 150, got %.2f", acc.Balance)
    }
}
```

No JUnit dependency. No `@Test` annotation. No test runner to configure. File ends in `_test.go`, function starts with `Test` — that's it. Run with `go test ./...`.

### Table-Driven Tests (Idiomatic)

```go
func TestDeposit_TableDriven(t *testing.T) {
    tests := []struct {
        name    string
        initial float64
        deposit float64
        wantErr bool
    }{
        {"valid deposit", 100, 50, false},
        {"zero deposit", 100, 0, true},
        {"negative deposit", 100, -10, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            acc := &BankAccount{Balance: tt.initial}
            err := acc.Deposit(tt.deposit)
            if (err != nil) != tt.wantErr {
                t.Errorf("Deposit() error = %v, wantErr = %v", err, tt.wantErr)
            }
        })
    }
}
```

This is the Go equivalent of JUnit's `@ParameterizedTest`. Cleaner, no annotations.

---

## 8. Frameworks

### Spring Boot vs Go Equivalents

Spring Boot is an opinionated, full-stack framework with DI container, autoconfiguration, AOP, security, data access, everything. Go frameworks are deliberately thin — they handle HTTP routing and middleware, nothing more.

| Spring Boot | Go Equivalent | Notes |
|------------|---------------|-------|
| Spring Web MVC | **Gin**, **Echo**, **Chi**, **Fiber** | Gin is most popular |
| Spring Data JPA | **GORM**, **sqlx**, **pgx** | GORM ≈ Hibernate, sqlx is lighter |
| Spring Security | Roll your own JWT middleware | No dominant security framework |
| Spring Boot Actuator | **Prometheus + pprof** | Built-in profiling via `net/http/pprof` |
| Kafka (Spring Kafka) | **confluent-kafka-go**, **sarama** | No annotation magic, more explicit |
| Spring DI / IoC | **Wire** (Google), **fx** (Uber) | DI is optional, most apps don't need it |
| Spring Cloud Config | **Viper** | Config from files/env/Consul |
| Spring Validation | **go-playground/validator** | Struct tags: `validate:"required,min=1"` |

### Gin — The Spring MVC Equivalent

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type DepositRequest struct {
    Amount float64 `json:"amount" binding:"required,gt=0"`
}

func main() {
    r := gin.Default() // includes Logger + Recovery middleware

    r.POST("/accounts/:id/deposit", func(c *gin.Context) {
        id := c.Param("id")

        var req DepositRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // call service...
        c.JSON(http.StatusOK, gin.H{"account_id": id, "deposited": req.Amount})
    })

    r.Run(":8080")
}
```

Compare this to Spring Boot: no `@RestController`, no `@PostMapping`, no `@RequestBody`, no `@Valid`. Less magic — you see every step.

### GORM Zero-Value Gotcha

```go
// ❌ This will NOT update Balance to 0 — GORM skips zero values with .Save()
db.Save(&Account{ID: 1, Balance: 0})

// ✅ Use .Updates() with a map for partial updates involving zero values
db.Model(&Account{ID: 1}).Updates(map[string]interface{}{
    "balance": 0,
})
```

### sqlx — Better for Complex Queries

```go
type AccountSummary struct {
    ID      string  `db:"id"`
    Balance float64 `db:"balance"`
    TxCount int     `db:"tx_count"`
}

var summary AccountSummary
err := db.Get(&summary, `
    SELECT a.id, a.balance, COUNT(t.id) as tx_count
    FROM accounts a
    LEFT JOIN transactions t ON t.account_id = a.id
    WHERE a.id = $1
    GROUP BY a.id, a.balance
`, accountID)
```

---

## 9. Generics

Go 1.18 (2022) added generics. If you're coming from Java, the mental model is similar but the syntax is different.

### Why it matters for Java devs

Java has had generics since 1.5 (with type erasure at runtime). Go generics use **type parameters** with **constraints** (Go interfaces). No type erasure — the compiler generates specialized code per type.

### Basic Generic Function

```java
// Java
public <T extends Comparable<T>> T max(T a, T b) {
    return a.compareTo(b) >= 0 ? a : b;
}
```

```go
// Go — comparable is a built-in constraint
func Max[T constraints.Ordered](a, b T) T {
    if a >= b {
        return a
    }
    return b
}

fmt.Println(Max(3, 5))       // int → 5
fmt.Println(Max(3.14, 2.71)) // float64 → 3.14
```

### Constraints — Go's Bounded Type Parameters

```go
import "golang.org/x/exp/constraints"

// constraints.Ordered = integer | float | string (supports < > operators)
// comparable = supports == != (built-in)

// Custom constraint using interface union types
type Number interface {
    int | int64 | float32 | float64
}

func Sum[T Number](values []T) T {
    var total T
    for _, v := range values {
        total += v
    }
    return total
}

fmt.Println(Sum([]int{1, 2, 3}))         // 6
fmt.Println(Sum([]float64{1.1, 2.2}))   // 3.3
```

### Generic Data Structures

```go
// Generic stack — replaces the interface{} anti-pattern
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    var zero T
    if len(s.items) == 0 {
        return zero, false
    }
    last := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return last, true
}

// Usage
stack := &Stack[int]{}
stack.Push(1)
stack.Push(2)
val, ok := stack.Pop() // 2, true
```

### When to Use Generics vs Interfaces

| Situation | Use |
|-----------|-----|
| Type-safe containers (stack, queue, set) | Generics |
| Functions that work on multiple numeric types | Generics |
| Behavior abstraction (logging, storage, auth) | Interfaces |
| You need runtime polymorphism | Interfaces |
| You need to call methods on the type | Interface constraint |

> **Rule of thumb:** If you're writing `interface{}` or `any` and casting with type assertions, generics are likely the better solution. If you're abstracting behavior, use plain interfaces.

---

## 10. Context Deep Dive

`context.Context` is the most important concept for any Go backend developer. Every HTTP handler, DB query, gRPC call, and external API call should accept and propagate a context.

### Why Context Exists

In Java, cancellation is solved via `Thread.interrupt()`, `Future.cancel()`, or reactive `Disposable`. These are implicit and require framework knowledge. Go's approach: pass a **Context** explicitly as the first parameter to every function that does I/O or long work.

### The Four Context Functions

```go
// 1. Background — root context, never cancelled (use in main, tests)
ctx := context.Background()

// 2. WithCancel — manual cancellation
ctx, cancel := context.WithCancel(parent)
defer cancel() // always call cancel to release resources

// 3. WithTimeout — automatically cancels after duration
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()

// 4. WithDeadline — cancels at absolute time
deadline := time.Now().Add(10 * time.Second)
ctx, cancel := context.WithDeadline(parent, deadline)
defer cancel()
```

### Propagating Context Through HTTP Handler → Service → DB

```go
// Handler — extract context from request
func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // already has deadline from server config
    id := chi.URLParam(r, "id")

    if err := h.service.Deposit(ctx, id, amount); err != nil {
        // ...
    }
}

// Service — forward context to repository
func (s *AccountService) Deposit(ctx context.Context, id string, amount float64) error {
    acc, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return fmt.Errorf("Deposit: %w", err)
    }
    // ...
    return s.repo.Save(ctx, acc)
}

// Repository — pass context to DB query
func (r *PostgresRepo) FindByID(ctx context.Context, id string) (*Account, error) {
    var acc Account
    err := r.db.GetContext(ctx, &acc, "SELECT * FROM accounts WHERE id = $1", id)
    return &acc, err
}
```

If the client disconnects mid-request, `ctx.Done()` closes, `db.GetContext` returns immediately with `context.Canceled`. No goroutine leak.

### Storing Values in Context

```go
// Define a typed key — prevents collision with other packages
type contextKey string
const userIDKey contextKey = "userID"

// Middleware sets value
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := extractUserIDFromJWT(r)
        ctx := context.WithValue(r.Context(), userIDKey, userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Handler reads value
func getAccountHandler(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value(userIDKey).(string)
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    // use userID...
}
```

> ⚠️ Only store request-scoped values in context (user ID, trace ID, request ID). Never store optional function parameters in context — that's an anti-pattern.

### errgroup — Parallel Calls with Context

```go
import "golang.org/x/sync/errgroup"

func (s *Service) GetDashboard(ctx context.Context, userID string) (*Dashboard, error) {
    g, ctx := errgroup.WithContext(ctx)

    var balance float64
    var txHistory []Transaction

    g.Go(func() error {
        var err error
        balance, err = s.repo.GetBalance(ctx, userID)
        return err
    })

    g.Go(func() error {
        var err error
        txHistory, err = s.repo.GetHistory(ctx, userID)
        return err
    })

    if err := g.Wait(); err != nil {
        return nil, err // first error wins; other goroutines get context cancelled
    }

    return &Dashboard{Balance: balance, History: txHistory}, nil
}
```

This is the Go equivalent of `CompletableFuture.allOf()` in Java, but with automatic context cancellation on first error.

---

## 11. Profiling & pprof

Go has **built-in profiling** via the `net/http/pprof` package. No JVM flags, no VisualVM, no YourKit license needed.

### Enable pprof in Your Server

```go
import _ "net/http/pprof" // side-effect import registers pprof routes

func main() {
    // Your app server
    go http.ListenAndServe(":8080", myRouter)

    // pprof server on separate port (never expose to internet)
    http.ListenAndServe(":6060", nil)
}
```

### Available Profiles

```bash
# CPU profile — where is time being spent?
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory/heap profile — what's allocating?
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine count and stacks — goroutine leak detector
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# Block profile — where are goroutines blocking?
go tool pprof http://localhost:6060/debug/pprof/block

# Mutex contention
go tool pprof http://localhost:6060/debug/pprof/mutex
```

### Reading the Flame Graph

```bash
# In pprof interactive mode
(pprof) top10          # top 10 functions by CPU time
(pprof) web            # opens flame graph in browser (requires graphviz)
(pprof) list MyFunc    # annotated source for a specific function
```

### Benchmark Tests

```go
// bench_test.go
func BenchmarkDeposit(b *testing.B) {
    acc := &BankAccount{Balance: 1000}
    for i := 0; i < b.N; i++ {
        acc.Deposit(10)
    }
}
```

```bash
go test -bench=. -benchmem ./...
# BenchmarkDeposit-8   10000000   120 ns/op   0 B/op   0 allocs/op
```

`-benchmem` shows allocations per operation — critical for identifying GC pressure. As a Java dev, you'd previously rely on JProfiler or async-profiler for this. In Go it's built in.

### Race Detector

```bash
go test -race ./...
go run -race main.go
```

The race detector finds concurrent memory access bugs at runtime. Always run it in CI. It has ~5–10x overhead so don't run it in production. Unlike Java's Helgrind/ThreadSanitizer, it's trivially integrated.

---

## 12. Common Go Gotchas for Java Devs

### 1. Loop Variable Capture (Pre Go 1.22)

```go
// ❌ All goroutines print the same value (last i)
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // captures reference to i
    }()
}

// ✅ Capture by value
for i := 0; i < 5; i++ {
    i := i // shadow variable — new binding each iteration
    go func() {
        fmt.Println(i)
    }()
}
```

> Go 1.22 fixed this: loop variables are now per-iteration, not per-loop. If you're on 1.22+, the `i := i` trick is unnecessary.

### 2. Nil Map Write Panic

```go
// ❌ Panics — nil map
var m map[string]int
m["key"] = 1 // panic: assignment to entry in nil map

// ✅ Always initialize maps
m := make(map[string]int)
m["key"] = 1
```

### 3. Goroutine Leak

```go
// ❌ Goroutine blocks forever if nobody reads from ch
func leak() {
    ch := make(chan int)
    go func() {
        ch <- compute() // blocks if caller returns early
    }()
    // caller returns — goroutine is leaked
}

// ✅ Use buffered channel or context cancellation
func safe(ctx context.Context) (int, error) {
    ch := make(chan int, 1) // buffer of 1 prevents goroutine leak
    go func() {
        ch <- compute()
    }()
    select {
    case result := <-ch:
        return result, nil
    case <-ctx.Done():
        return 0, ctx.Err()
    }
}
```

### 4. Interface With Nil Pointer — The Nil Trap

```go
// ❌ This is NOT nil — it's an interface holding a typed nil
var p *MyError = nil
var err error = p
fmt.Println(err == nil) // false! (interface has type info even with nil value)

// ✅ Return untyped nil
func findAccount(id string) (*Account, error) {
    if notFound {
        return nil, nil // untyped nil — caller's err == nil check works
    }
}
```

This is one of the most infamous Go gotchas. If you return a typed nil pointer as an interface, the interface is NOT nil.

### 5. Defer in Loops

```go
// ❌ Files only closed when function returns, not each iteration
for _, file := range files {
    f, _ := os.Open(file)
    defer f.Close() // deferred until function exit, not loop iteration
}

// ✅ Wrap in anonymous function
for _, file := range files {
    func() {
        f, _ := os.Open(file)
        defer f.Close() // now closes when anonymous function returns
        process(f)
    }()
}
```

### 6. Shadowing Errors with :=

```go
// ❌ err in inner scope shadows outer err — outer err never set
err := doFirst()
if true {
    result, err := doSecond() // new err, shadows outer
    _ = result
}
// outer err is still from doFirst()

// ✅ Declare result separately
var result string
result, err = doSecond() // assigns to outer err
```

### 7. Append Doesn't Always Allocate

```go
a := make([]int, 3, 10) // len=3, cap=10
b := a[0:3]

b = append(b, 99) // writes into a's backing array — a[3] is now 99!
fmt.Println(a[:4]) // [0 0 0 99] — surprise!
```

When slices share a backing array, appending within capacity mutates both. Use `a[0:3:3]` (three-index slice) to limit capacity and force a copy on append.

---

## 13. When Go Wins, When Java Wins

### Go Wins When:

- **High concurrency with low memory** — thousands of simultaneous connections (WebSocket servers, API gateways, proxies)
- **Fast startup / Lambda / containers** — Go binaries start in milliseconds; Spring Boot takes 2–10s
- **CLI tools and DevOps tooling** — Docker, Kubernetes, Terraform, and most modern infra tools are written in Go
- **Network-heavy microservices** — HTTP proxies, gRPC services, event consumers
- **Team wants readable simplicity** — fewer opinions, less magic, everyone can understand the code

### Java Wins When:

- **Complex enterprise business logic** — Spring's ecosystem (Security, Batch, State Machine, Camunda) has no equivalent in Go
- **Rich ORM / complex persistence** — Hibernate/Spring Data JPA handles entity graphs, lazy loading, caching that GORM doesn't
- **Legacy integration** — SOAP, JMS, existing Java libs
- **Big data / analytics** — Spark, Flink, Kafka Streams are JVM-native
- **Deep reflection-based frameworks** — Spring's AOP, proxies, annotation processing are JVM superpowers
- **Large teams on complex domains** — Java's type system and Spring's structure scale better for 50-developer monorepos

```
Go is better for: infrastructure, networking, concurrency, cloud-native services
Java is better for: complex business domains, enterprise ecosystems, big data
```

They're not competitors — Google uses both.

---

## 14. 2-Month Learning Roadmap

**Goal:** After 2 months of consistent work (1–1.5 hrs/day), you'll be able to build production-quality REST APIs in Go, handle concurrency correctly, and be genuinely competitive for a mid/strong-junior Go role.

### Week 1 — Language Fundamentals

| Day | Topic | Resource / Task |
|-----|-------|-----------------|
| 1 | Setup + syntax basics | Install Go, run `go tour` (tour.golang.org), read first 3 chapters of *Go By Example* |
| 2 | Types, variables, functions | Write a simple bank account struct with Deposit/Withdraw methods |
| 3 | Slices, maps, ranges | Implement a transaction history using slices + map aggregation |
| 4 | Pointers + zero values | Rewrite bank account to use pointer receivers, understand why |
| 5 | Structs + methods + interfaces | Define `Account` interface, implement 2 types (Savings, Current) |
| 6 | Error handling patterns | Add error returns to all methods, practice wrapping with `%w` |
| 7 | Review + mini project | Build a CLI bank account simulator: create, deposit, withdraw, print balance |

### Week 2 — Go Idioms & Standard Library

| Day | Topic | Task |
|-----|-------|------|
| 1 | Interfaces in depth | Implement the `io.Writer` interface; understand why Go interfaces are small |
| 2 | Error types + `errors.Is/As` | Create custom error types for your bank app |
| 3 | `defer`, `panic`, `recover` | Add deferred logging and recovery to a server handler |
| 4 | Goroutines basics | Spawn 100 goroutines, use `sync.WaitGroup` to wait for completion |
| 5 | Channels | Build a pipeline: producer → channel → consumer |
| 6 | `select` + timeouts | Add timeout logic to your pipeline using `time.After` |
| 7 | `context.Context` | Pass context through all function calls; cancel on timeout |

### Week 3 — HTTP, Testing, Tooling

| Day | Topic | Task |
|-----|-------|------|
| 1 | `net/http` standard library | Build a basic HTTP server WITHOUT Gin — understand the fundamentals |
| 2 | Gin introduction | Rebuild the server with Gin; compare the two |
| 3 | Request binding + validation | Add `ShouldBindJSON` + `go-playground/validator` struct tags |
| 4 | Middleware | Write auth middleware that checks a JWT header |
| 5 | Testing basics | Write unit tests for your service layer using table-driven tests |
| 6 | HTTP handler tests | Use `httptest.NewRecorder()` to test handlers without a real server |
| 7 | Project: Account API | Full CRUD for accounts: POST /accounts, GET /accounts/:id, POST /accounts/:id/deposit |

### Week 4 — Database + Real Patterns

| Day | Topic | Task |
|-----|-------|------|
| 1 | `database/sql` basics | Connect to Postgres, run a raw query, scan results into structs |
| 2 | `sqlx` | Migrate to sqlx; use named queries and struct tags |
| 3 | DB transactions | Implement a transfer endpoint: debit one account, credit another, atomic |
| 4 | GORM intro | Rebuild models with GORM; understand `.Save()` vs `.Updates()` gotcha |
| 5 | Repository pattern | Extract DB logic into a repository interface — makes mocking easy |
| 6 | Mock testing | Use `testify/mock` to mock the repository in service tests |
| 7 | Config + environment | Use `viper` to load config from `.env` and env variables |

### Week 5 — Concurrency Patterns in Production

| Day | Topic | Task |
|-----|-------|------|
| 1 | Worker pool pattern | Process 1000 transactions concurrently using a fixed pool of goroutines |
| 2 | Race detector | Run `go test -race ./...`; fix any races found |
| 3 | `sync.Mutex` vs channels | Compare approaches; know when to use each |
| 4 | `errgroup` | Use `golang.org/x/sync/errgroup` to run parallel DB queries safely |
| 5 | Rate limiting | Implement a simple rate limiter using a ticker channel |
| 6 | Graceful shutdown | Handle `SIGTERM` with `context.WithCancel` + HTTP server `Shutdown()` |
| 7 | Mini project | Batch payment processor: read from a channel, process concurrently, aggregate results |

### Week 6 — gRPC + Kafka + Cloud Patterns

| Day | Topic | Task |
|-----|-------|------|
| 1 | gRPC basics | Define a `.proto` file for Account service, generate Go code |
| 2 | gRPC server + client | Implement `GetAccount` and `Transfer` RPC methods |
| 3 | gRPC interceptors | Add auth + logging interceptors (equivalent to Spring AOP advice) |
| 4 | Kafka with sarama | Publish a `TransactionCreated` event after each deposit |
| 5 | Kafka consumer | Consume events, update a read-model table |
| 6 | Docker + multi-stage build | Dockerfile: build stage (golang) → runtime stage (scratch/distroless) |
| 7 | Health checks + metrics | Add `/health` endpoint + Prometheus metrics with `prometheus/client_golang` |

### Week 7 — Architecture & Clean Code

| Day | Topic | Task |
|-----|-------|------|
| 1 | Project layout | Read `golang-standards/project-layout`, restructure: `cmd/`, `internal/`, `pkg/` |
| 2 | Dependency injection with Wire | Auto-generate wiring for your service dependencies |
| 3 | Hexagonal architecture in Go | Separate domain, ports, adapters — no framework magic needed |
| 4 | Middleware chains | Build a full middleware stack: recovery → tracing → auth → rate limit |
| 5 | Structured logging | Use `zap` or `slog` (stdlib in Go 1.21+); log with fields not formatted strings |
| 6 | Distributed tracing | Add OpenTelemetry tracing spans across HTTP → service → DB |
| 7 | Code review | Review your week 3 project — rewrite it applying everything learned since |

### Week 8 — Final Project + Interview Prep

**Final Project: Mini Banking API**

Build a production-ready microservice with:
- `POST /accounts` — create account
- `POST /transfers` — atomic transfer (Postgres transaction)
- `GET /accounts/:id/balance` — current balance
- Kafka event publishing on each transfer
- JWT middleware on all routes
- Full test coverage (unit + integration)
- Dockerized with docker-compose (app + postgres + kafka)
- Prometheus metrics endpoint
- Graceful shutdown

**Interview Topics to Review:**

- Explain goroutines vs threads (M:N model, scheduler, stack size)
- When would you use a channel vs mutex?
- What is `context.Context` and why is it the first parameter of every function?
- How does Go handle errors? Why not exceptions?
- What does `defer` guarantee?
- Explain interface satisfaction and why there's no `implements`
- What's the difference between `new()` and `make()`?
- How would you structure a Go project? (cmd/internal/pkg)
- What's the zero value and why does it matter?
- How do you avoid data races?
- What are generics constraints and when do you use generics vs interfaces?
- What is escape analysis and how does it affect performance?

---

### Tools to Install Now

```bash
# Go itself
brew install go                    # or https://go.dev/dl/

# Language server (IntelliJ / VSCode)
go install golang.org/x/tools/gopls@latest

# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Air — hot reload (like Spring DevTools)
go install github.com/air-verse/air@latest

# Wire — dependency injection
go install github.com/google/wire/cmd/wire@latest

# staticcheck — deeper static analysis
go install honnef.co/go/tools/cmd/staticcheck@latest
```

---

### Key Books & Resources

| Resource | Why |
|----------|-----|
| [Go By Example](https://gobyexample.com) | Best quick reference for syntax |
| [The Go Programming Language (Donovan & Kernighan)](https://www.gopl.io) | The definitive book — read ch 1–8 |
| [100 Go Mistakes (Harsanyi)](https://100go.co) | Must read before going to production |
| [Effective Go](https://go.dev/doc/effective_go) | Official idioms guide |
| [Go Patterns](https://github.com/tmrts/go-patterns) | Concurrency + design patterns |
| [Ardan Labs courses](https://www.ardanlabs.com/training/) | Best paid Go training |

---

### Realistic Milestone Check

| End of Week | What You Can Do |
|-------------|----------------|
| 2 | Read and write Go fluently; no syntax surprises |
| 4 | Build tested REST APIs with Postgres; solid error handling |
| 6 | Concurrent services; Kafka integration; Docker deployment |
| 8 | Production-quality service; can discuss Go internals confidently |

> **After 2 months with this plan:** You're genuinely hireable as a mid-junior Go developer. Especially strong given your Java/Spring Boot background — you already understand distributed systems, transactions, and HTTP APIs. Go will feel like removing 40% of the ceremony you deal with daily in Spring.

---

## Why It Matters

Go is the dominant language for cloud-native infrastructure and microservices (Docker, Kubernetes, Terraform, Prometheus — all Go). For a Java/Spring Boot developer, Go represents the highest-ROI adjacent skill: the concepts (REST, gRPC, DB transactions, Kafka) transfer directly; only the idioms differ. A team that ships Java monoliths can extract Go microservices for latency-critical or high-concurrency paths without rebuilding domain knowledge from scratch.
