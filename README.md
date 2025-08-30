# (Type‑Safe) Registry 📑

[![Go Reference](https://pkg.go.dev/badge/github.com/mp3cko/registry.svg)](https://pkg.go.dev/github.com/mp3cko/registry) [![Go Report Card](https://goreportcard.com/badge/github.com/mp3cko/registry)](https://goreportcard.com/report/github.com/mp3cko/registry) [![Coverage Status](https://img.shields.io/badge/coverage-81.6%25-brightgreen.svg)](https://github.com/mp3cko/registry)

A tiny, zero‑dependency, type‑safe registry / service locator for Go. Register and retrieve values by their (generic) type and an optional name. Add constraints (unique types, unique names), enforce minimum accessibility & namedness, clone configs/entries, and scope operations to specific registries — all via a consistent, chainable options API.

Why? Sometimes you want: (1) late binding, (2) test overrides, (3) lightweight plugin wiring, or (4) a place to stash cross‑cutting infra without heavy DI frameworks or unsafe casts.

---

## Contents

- Features
- Installation
- 60‑Second Tour
- Core Concepts
- Options & Validity Matrix
- Common Patterns & Recipes
- Advanced Topics (accessibility, namedness, cloning, uniqueness)
- Error Handling
- Best Practices & Anti‑Patterns
- FAQ
- Contributing / License

---

## Features

- Type‑safe: `Get[T]()` returns `T` (no interface / reflection gymnastics in user code)
- Named instances: Multiple instances per type, distinguished by name
- Chain or variadic options: `Set(x, WithName("a"), WithRegistry(r))` or `Set(x, WithName("a").WithRegistry(r))`
- Unique constraints: Per type or per name (per type)
- Accessibility & Namedness enforcement: Prevent registering values you cannot semantically retrieve
- Cloning: Copy config, entries, or both when creating a new registry
- Per‑call scoping: Temporarily target a different registry with `WithRegistry`
- Thread‑safe: Internal locking; per‑call options isolated
- Zero runtime reflection surprises: Reflection is confined & deterministic

---

## Installation

```bash
go get github.com/mp3cko/registry
```

---

## 60‑Second Tour

```go
package main

import (
    "fmt"
    "log"

    reg "github.com/mp3cko/registry"
)

// Define a service contract
type Greeter interface { Greet() string }

// Concrete implementation
type englishGreeter struct{}
func (englishGreeter) Greet() string { return "Hello" }

func main() {
    // (1) Use default registry: register by interface to allow substitution in tests
    if err := reg.Set[Greeter](englishGreeter{}); err != nil { log.Fatal(err) }

    // (2) Retrieve it later, anywhere
    g, err := reg.Get[Greeter]()
    if err != nil { log.Fatal(err) }
    fmt.Println(g.Greet()) // Hello

    // (3) Add another implementation by name
    type pirateGreeter struct{}
    func (pirateGreeter) Greet() string { return "Ahoy" }
    _ = reg.Set[Greeter](pirateGreeter{}, reg.WithName("pirate"))

    // Get named variant
    pg, _ := reg.Get[Greeter](reg.WithName("pirate"))
    fmt.Println(pg.Greet()) // Ahoy
}
```

---

## Core Concepts

### 1. Registry vs Default Registry

`NewRegistry()` creates an isolated registry. The package also maintains a default global registry used when you do not specify `WithRegistry(...)`.

### 2. Key = (Type, Name)

Instances are stored under their generic type `T` and an optional string name (default: empty string). You can therefore have:

```text
T=Cache name=""           -> default Cache
T=Cache name="hot"        -> hot shard
T=Cache name="cold"       -> cold shard
```

### 3. Options Are Contextual

Some options only make sense at construction (e.g. cloning), others only per call (e.g. `WithRegistry`), and some are valid everywhere (e.g. `WithName`). Invalid combinations fail fast with `ErrNotSupported`.

### 4. Per‑Call State Is Ephemeral

Options passed to `Get/Set/Unset/GetAll` are applied for that call only. Internal call state is wiped immediately afterward — you never “leak” options to subsequent calls.

---

## Options & Validity Matrix

Legend: C = Constructor (`NewRegistry`), O = Operation (`Set`, `Get`, `GetAll`, `Unset`), * = limited subset, ✗ = invalid.

| Option              | C   | Set | Get | GetAll | Unset | Notes                                                                             |
| ------------------- | --- | --- | --- | ------ | ----- | --------------------------------------------------------------------------------- |
| `WithName`          | C   | ✓  | ✓  | ✓     | ✓    | At construction sets default name (used when no per‑call name given)              |
| `WithRegistry`      | ✗  | ✓  | ✓  | ✓     | ✓    | Only scopes that single call; cannot be used in constructor (use cloning instead) |
| `WithUniqueType`    | ✓  | ✓  | ✓  | ✓     | ✓    | Constructor: enforce always; per call: assert uniqueness / constrain operation    |
| `WithUniqueName`    | ✓  | ✓  | ✗  | ✗     | ✗    | Name uniqueness per type; retrieval must use name explicitly instead              |
| `WithAccessibility` | ✓  | ✓  | ✗  | ✓     | ✗    | Per call only meaningful for Set/GetAll (Get/Unset already name the type)         |
| `WithNamedness`     | ✓  | ✓  | ✗  | ✓     | ✗    | Prevent anonymous types; retrieval already pins type                              |
| `WithCloneConfig`   | ✓  | ✗  | ✗  | ✗     | ✗    | Applied 3rd to last (before entries + registry)                                   |
| `WithCloneEntries`  | ✓  | ✗  | ✗  | ✗     | ✗    | Applied 2nd to last                                                               |
| `WithCloneRegistry` | ✓  | ✗  | ✗  | ✗     | ✗    | Applied last; conflicts detected & yield `ErrBadOption`                           |

Notes:

1. “Per call” uniqueness (`WithUniqueType()`) on `Get` fails if more than one instance is registered for that type.
2. `WithUniqueName()` makes sense only where a new name might collide (constructor / Set). Using it on retrieval would add no safety, thus invalid.

---

## API Cheat Sheet

```go
reg.Set[T](val, opts...)             // Register
reg.Get[T](opts...) (T, error)       // Retrieve one instance
reg.Unset[T](zeroValOrExample, opts...) // Remove (type + optional name)
reg.GetAll(opts...) (map[reflect.Type]map[string]any, error) // Snapshot of all entries
reg.NewRegistry(opts...) (*registry, error) // Fresh registry
reg.SetDefaultRegistry(r)            // Swap global default atomically
```

You can always refer to tests (`*_test.go`) for executable examples.

---

## Common Patterns & Recipes

### 1. Environment / Mode Isolation

```go
prod, _ := reg.NewRegistry(reg.WithUniqueType())
dev,  _ := reg.NewRegistry()
reg.Set[Greeter](englishGreeter{}, reg.WithRegistry(prod))
reg.Set[Greeter](englishGreeter{}, reg.WithRegistry(dev), reg.WithName("override"))
```

### 2. Test Override

```go
// production registration
reg.Set[Greeter](englishGreeter{})

// in tests
type fakeGreeter struct{ msg string }
func (f fakeGreeter) Greet() string { return f.msg }
_ = reg.Set[Greeter](fakeGreeter{"hi from test"}, reg.WithName("test"))

g, _ := reg.Get[Greeter](reg.WithName("test")) // isolate test instance
```

### 3. Plug‑in / Module Registration

```go
// Each module gets its own registry
moduleReg, _ := reg.NewRegistry(reg.WithUniqueType())
// Modules register their handlers without touching global state
reg.Set[Handler](NewHandler(), reg.WithRegistry(moduleReg))
```

### 4. Enforcing Only One Implementation

```go
singletons, _ := reg.NewRegistry(reg.WithUniqueType())
_ = reg.Set[Config](LoadConfig(), reg.WithRegistry(singletons))
// Another Set[Config] in same registry -> ErrNotUniqueType
```

### 5. Using Interfaces to Wrap Unexported Concrete Types

```go
// external package returns *unexported concrete
extVal := external.NewThing() // *external.unexportedThing

// Register via its exported interface instead
reg.Set[external.Thing](extVal) // passes accessibility checks
```

---

## Advanced Topics

### Accessibility

`WithAccessibility(level)` ensures every registered type is at least that visible to the caller (package vs exported). This avoids trapping an unexportable type you can never refer to again.

Typical: enforce `AccessibleInsidePackage` (default) or tighten to `AccessibleEverywhere` in public plugin ecosystems.

### Namedness

Anonymous types (especially inline interfaces) are legal but awkward:

```go
// BAD – retrieval must use identical anonymous interface definition
reg.Set[interface{ Greet() string }](englishGreeter{})
// Prefer named interface
type Greeter interface { Greet() string }
reg.Set[Greeter](englishGreeter{})
```

Use `WithNamedness(access.NamedType)` to prevent anonymous registrations.

### Uniqueness

Two knobs:

1. `WithUniqueType()` – at registry construction: only one instance per type forever. Per call: assert uniqueness for that operation (helpful during migration to strict mode).
2. `WithUniqueName()` – names may not repeat per type; retrieval still requires specifying a name (so uniqueness adds nothing on Get and is disallowed there).

### Cloning Semantics & Priority

When constructing a registry the option execution order (by priority) matters if you combine cloning with modifiers. High‑level summary:

1. Regular config modifiers run.
2. `WithCloneConfig(src)` copies config (cannot conflict with previous mutations or you get `ErrBadOption`).
3. `WithCloneEntries(src)` copies entries (subject to config already in place).
4. `WithCloneRegistry(src)` copies both (final validation vs earlier options). Use this when you just want “a full duplicate”, otherwise compose the other two.

### `GetAll` Caveats

`GetAll` returns a snapshot map of `reflect.Type -> map[name]any`. It is intentionally not type‑safe; convert carefully. Use it for diagnostics, debugging, or bulk migrations — not as your primary access path.

---

## Error Handling

Use `errors.Is` (errors are wrapped with context):

| Error                    | Meaning                                          |
| ------------------------ | ------------------------------------------------ |
| `ErrNotFound`            | No entry for (type, name)                        |
| `ErrNotUniqueType`       | Multiple instances exist but uniqueness required |
| `ErrNotUniqueName`       | Name already taken (when uniqueness enforced)    |
| `ErrNotSupported`        | Option invalid in this context                   |
| `ErrAccessibilityTooLow` | Value's type visibility below required minimum   |
| `ErrNamednessTooLow`     | Anonymous type rejected by namedness constraint  |
| `ErrBadOption`           | Incompatible or conflicting constructor options  |

Example:

```go
val, err := reg.Get[Greeter]()
if err != nil {
    switch {
    case errors.Is(err, reg.ErrNotFound): /* recover / fallback */
    case errors.Is(err, reg.ErrNotUniqueType): log.Fatal("config error: multiple greeters; enforce WithName or uniqueness")
    default: log.Fatal(err)
    }
}
_ = val
```

---

## Best Practices

1. Register by interface, not struct: encourages substitution & testing
2. Use names when you truly need >1 instance per type (shards, multi‑tenant, stage)
3. Consider `WithUniqueType()` for config objects or true singletons
4. Enforce `WithNamedness(access.NamedType)` early to avoid anonymous retrieval headaches
5. Keep `GetAll` for introspection; prefer typed `Get`
6. Scope temporary lookups with `WithRegistry` rather than swapping the default
7. Replace the default registry only at program bootstrap via `reg.SetDefaultRegistry` if you must
8. Handle errors explicitly — silent failure = hidden misconfiguration

### Anti‑Patterns

- Using the registry everywhere instead of dependency injection for local collaborators
- Storing massive mutable collections — keep entries small, stable references
- Registering anonymous inline interfaces then expecting named retrieval
- Using `GetAll` in hot code paths (avoid reflection map walks)

---


## FAQ

**Q: Is this a Service Locator (an anti‑pattern)?**
A: It can be misused as one. Treat it as a composition helper at boundaries (plugins, bootstrap, tests). Hand regular dependencies explicitly.

**Q: Why does `Get` sometimes return `ErrNotUniqueType`?**
There are multiple instances for that type and you asked for uniqueness (either registry was created with `WithUniqueType` or you passed it per call). Either name them (`WithName`) or remove the uniqueness constraint.

**Q: How do I unregister?**
Call `Unset[T](exampleValue, opts...)`. The value passed supplies the type parameter only; its contents are not used beyond that.

**Q: Can I list everything strongly typed?**
No; enumeration uses reflection. Iterate and cast deliberately.

**Q: Performance?**
Operations are O(1) map lookups with a small reflection cost for the generic type. Unless you are doing these in tight inner loops (unlikely) cost is negligible.

---

## Contributing

PRs welcome. Please:

1. Open an issue or clearly describe the motivation
2. Add/adjust tests (maintain coverage)
3. Keep API surface minimal & coherent
4. Run `go test ./...`

## License

MIT – see [LICENSE](LICENSE)
