# errs

A lightweight Go library that provides `try`/`catch`-style error handling using `panic` and `recover`.

## Installation

```sh
go get github.com/pixel2175/errs
```

## Functions

### `TryE(err error)`

Panics if `err` is not `nil`. Does nothing if `err` is `nil`.

```go
errs.TryE(os.WriteFile("file.txt", data, 0644))
```

### `TryV[T any](v T, err error, custom ...error) T`

Returns `v` if `err` is `nil`. Otherwise, panics with the original error or the first custom error if provided.

```go
content := errs.TryV(os.ReadFile("file.txt"))
```

With a custom error:

```go
content := errs.TryV(
    os.ReadFile("file.txt"),
    errors.New("failed to read file"),
)
```

### `Catch(err *error, fn ...func(*error))`

Recovers a panic containing an `error`, assigns it to `*err`, and optionally calls the first non-nil callback.
If the recovered value is not an `error`, `Catch` re-panics with the original value.
**Important:** `Catch` must be deferred to recover a panic.

# Simple Examples

```go
func Parent() {
    var err error
    defer errs.Catch(&err, func(err *error) {
        CleanUp()
        fmt.Printf("[Parent]: %s\n", (*err).Error())
        os.Exit(1)
    })

    InitLogic()
    config := errs.TryV( LoadFromExternalLibrary() )
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

## Limitations

- `TryE` and `TryV` use panics to propagate errors.
- `Catch` recovers only panics that occur in the same goroutine and while its deferred call is active.
- Recovering a panic does not resume execution at the panic location.
