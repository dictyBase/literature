# Common fp-go Patterns

This document outlines idiomatic patterns for building applications with `fp-go` v2, derived from the refactoring of the `examples/cli-tool` application.

## Table of Contents

- [Dependency Injection with ReaderIOEither](#dependency-injection-with-readerioeither)
- [Context Layering](#context-layering)
- [Structured Logging in Pipelines](#structured-logging-in-pipelines)
- [Branching and Fallbacks](#branching-and-fallbacks)
- [Domain Actions](#domain-actions)

---

## Dependency Injection with ReaderIOEither

### ❌ Anti-Pattern: Manual Client Creation or Global State

Instantiating clients (HTTP, Database, etc.) directly inside your business logic makes testing difficult and couples your code to specific implementations.

```go
// Bad: Hardcoded dependency creation
func FetchArticle(id string) IOE.IOEither[error, Article] {
    client := literature.NewEuropePMCClient() // Hard dependency
    return client.GetArticle(id)
}
```

### ✅ Idiomatic Pattern: Bind for Injection

Use `IOE.Bind` (or `RIE.Bind`) to progressively build a context containing your dependencies. This happens at the "edge" of your application (e.g., `main` or a composition root).

```go
// main.go
F.Pipe5(
    // 1. Start with basic configuration
    IOE.Do[error](RunContext{
        Identifier: identifier,
        Logger:     logger,
    }),
    // 2. Inject EuropePMC Client
    IOE.Bind(InjectEuropeClient, createEuropeClient),
    // 3. Inject PubMed Client
    IOE.Bind(InjectPubMedClient, createPubMedClient),
    // 4. Execute Flow with fully formed context
    IOE.Chain(ExecuteFlow), 
    // ...
)

// di.go
var InjectEuropeClient = F.Curry2(
    func(client *literature.EuropePMCClient, ctx RunContext) WithEuropeClient {
        return WithEuropeClient{RunContext: ctx, Europe: client}
    },
)
```

**Benefits:**
- **Testability:** You can easily swap `createEuropeClient` for a mock injector in tests.
- **Clarity:** The dependency graph is explicit in the `main` pipeline.

---

## Context Layering

### ❌ Anti-Pattern: Monolithic Context or context.Context abuse

Passing a single huge struct or relying on `context.Context` for application dependencies.

### ✅ Idiomatic Pattern: Evolving Types

Define specific structs that represent the "state" of your dependency graph at each stage. Embedding previous contexts allows access to parent data.

```go
// Base Context
type RunContext struct {
    Identifier string
    OutputFile string
    Logger     *log.Logger
}

// Layer 1: Adds EuropePMC
type WithEuropeClient struct {
    RunContext // Embedded
    Europe     *literature.EuropePMCClient
}

// Layer 2: Adds PubMed
type WithPubMedClient struct {
    WithEuropeClient // Embedded
    PubMed           *literature.Client
}
```

**Benefits:**
- **Type Safety:** Functions declare exactly what they need (`WithEuropeClient` vs `WithPubMedClient`).
- **Access:** Inner layers can access outer layers (e.g., `ctx.Logger` is available everywhere).

---

## Structured Logging in Pipelines

### ❌ Anti-Pattern: Inline Side Effects

Breaking the purity of the pipeline to log intermediate results.

```go
// Bad: Inline side effect
IOE.Map(func(a Article) Article {
    fmt.Println("Got article:", a.Title) // Side effect in Map!
    return a
}),
```

### ✅ Idiomatic Pattern: ChainFirstIOK

Use `ChainFirstIOK` (or `ChainFirst`) to execute a side-effect (logging) without altering the data flowing through the pipe.

```go
// logger.go
var logEuropeArticle = func(ctx WithEuropePMCArticle) IO.IO[WithEuropePMCArticle] {
    return func() WithEuropePMCArticle {
        ctx.Logger.Println("Found:", ctx.Article.Title)
        return ctx
    }
}

// pipeline
IOE.ChainFirstIOK[error](logEuropeArticle),
```

**Benefits:**
- **Purity:** The logging function is wrapped in `IO`.
- **Flow:** The data passes through unchanged.
- **Isoloation:** Logging logic is separated from business logic.

---

## Branching and Fallbacks

### ❌ Anti-Pattern: Imperative If/Else

Using imperative conditional logic inside a `Map` or `Chain`.

```go
// Bad: Imperative branching
IOE.Chain(func(ctx Context) IOE.IOEither[error, any] {
    if ctx.HasPDF {
        return downloadPDF(ctx)
    } else {
        return tryPubMed(ctx)
    }
}),
```

### ✅ Idiomatic Pattern: MonadAlt and Ternary

Use `RIE.MonadAlt` for fallback strategies (try A, if fail, try B) and `F.Ternary` for explicit conditional paths based on data.

**Fallback Strategy (ReaderIOEither MonadAlt):**
```go
// main.go
IOE.Chain(
    RIE.MonadAlt(
        ExecuteEuropeFlow, // Try this first (ReaderIOEither)
        func() RIE.ReaderIOEither[WithPubMedClient, error, any] {
            return ExecutePubMedFlow // Fallback if first fails
        },
    ),
),
```

**Conditional Logic (Ternary):**
```go
// action_europe.go
IOE.Chain(
    F.Ternary(
        hasEuropePDF,      // Predicate: func(Context) bool
        downloadEuropePDF, // True path
        fallbackEurope,    // False path
    ),
),
```

**Benefits:**
- **Declarative:** The flow is visible at the pipeline level.
- **Composition:** Strategies can be composed independently.

---

## Domain Actions

### ❌ Anti-Pattern: massive `run` function

Putting all logic into a single 500-line `run` function.

### ✅ Idiomatic Pattern: Specialized Action Functions

Break logic into functions that accept a specific Context and return `IOEither`.

```go
// action_europe.go
func ExecuteEuropeFlow(ctx WithPubMedClient) IOE.IOEither[error, any] {
    return F.Pipe3(...)
}

// action_pubmed.go
func ExecutePubMedFlow(ctx WithPubMedClient) IOE.IOEither[error, any] {
    return F.Pipe4(...)
}
```

These functions act as "controllers" or "use cases" for specific domains of your application.
