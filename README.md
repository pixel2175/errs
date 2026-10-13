# errs

A lightweight Go library that provides `try`/`catch`-style error handling using `panic` and `recover`.

## Installation

```sh
go get github.com/pixel2175/errs
```

## Functions

### `TryE(err error, handler ...any)`

Does nothing if `err` is `nil`. Otherwise it panics with an `*errs.Error`.

```go
errs.TryE(os.WriteFile("file.txt", data, 0644))
```

### `TryV[T any](v T, err error, handler ...any) T`

Returns `v` if `err` is `nil`. Otherwise it behaves like `TryE`.

```go
content := errs.TryV(os.ReadFile("file.txt"))
```

### Handlers

An optional handler runs only when `err != nil`. Accepted types: `func() error` and `func(error) error`.

- It returns `nil`: the error is ignored and execution continues (`TryV` returns `v`).
- It returns an error: that error is panicked instead.

```go
errs.TryE(Migrate(), func(err error) error {
    return fmt.Errorf("migrate: %w", err)
})
```

Go does not allow extra arguments after a multi-value call, so split it:

```go
b, e := os.ReadFile("app.toml")
data := errs.TryV(b, e, func(err error) error {
    log.Println("using defaults:", err)
    return nil
})
```

### `Catch(err *error, fn ...func(*error))`

Must be deferred. It recovers panics raised by `TryE`/`TryV`, stores the error in `*err`, and calls the optional callback `fn[0]`.
Any other panic (nil map, index out of range, ...) is re-panicked unchanged.

### `*errs.Error`

The error type shared between `TryE`/`TryV` and `Catch`.

```go
type Error struct {
    Err   error  // the original error
    Stack []byte // stack captured at the first panic
}
```

- `err.Error()` returns the message string of the original error.
- `errors.Is` and `errors.As` work through `Unwrap`.
- `err == someErr` no longer works. Use `errors.Is`.
- If an error that already carries a stack is panicked again, the first stack is reused.

```go
var e *errs.Error
if errors.As(err, &e) {
    fmt.Printf("%s\n%s", e, e.Stack)
}
```

## Goroutines

A panic cannot cross goroutines, so a `Catch` in the parent will not see it. Defer `Catch` inside the goroutine:

```go
go func() {
    var err error
    defer errs.Catch(&err, func(err *error) {
        log.Printf("[worker]: %s", *err)
    })
    errs.TryE(Worker())
}()
```

Without it, the panic crashes the whole program.

## Simple Examples

```go
func Parent() {
    var err error
    defer errs.Catch(&err, func(err *error) {
        CleanUp()
        var e *errs.Error
        if errors.As(*err, &e) {
            fmt.Printf("[Parent]: %s\n%s", e, e.Stack)
        }
        os.Exit(1)
    })

    InitLogic()

    cfg, e := LoadConfig("app.toml")
    config := errs.TryV(cfg, e, func(e error) error {
        log.Printf("config failed: %v", e)
        return nil // ignore the error
    })

    errs.TryE(StartProcessing(config))
}

func InitLogic() {
    errs.TryE(Something1())
    errs.TryE(Something2())
    errs.TryE(Something3())
}

func StartProcessing(config *Config) (err error) {
    defer errs.Catch(&err, func(err *error) {
        *err = fmt.Errorf("processing: %w", *err)
    })

    errs.TryE(DoSomething(config))
    return nil
}
```
