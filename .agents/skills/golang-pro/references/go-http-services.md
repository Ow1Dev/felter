# Go HTTP Services (Mat Ryer / Grafana Pattern)

Patterns for building maintainable, testable HTTP services in Go based on the Grafana blog post "How I write HTTP services in Go after 13 years".

## 1. `func main()` only calls `run()`

`main()` should be minimal. All logic lives in a `run()` function that takes OS fundamentals as arguments and returns an error.

```go
func run(
    ctx    context.Context,
    args   []string,
    getenv func(string) string,
    stdin  io.Reader,
    stdout, stderr io.Writer,
) error

func main() {
    ctx := context.Background()
    if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err)
        os.Exit(1)
    }
}
```

Benefits:
- Testable: call `run()` from tests with injected dependencies
- No globals: enables `t.Parallel()`
- Clean shutdown via `signal.NotifyContext` inside `run`

## 2. `NewServer` Constructor

One big constructor that takes ALL dependencies as explicit arguments. Returns `http.Handler`.

```go
func NewServer(
    logger *logging.Logger,
    config *Config,
    store  *Store,
) http.Handler {
    mux := http.NewServeMux()
    addRoutes(mux, logger, config, store)
    
    var handler http.Handler = mux
    handler = loggingMiddleware(logger, handler)
    handler = authMiddleware(handler)
    return handler
}
```

## 3. `addRoutes` Maps the Entire API Surface

One file, one function. All routes listed here.

```go
func addRoutes(
    mux    *http.ServeMux,
    logger *logging.Logger,
    config *Config,
    store  *Store,
) {
    mux.Handle("/api/users", handleListUsers(store))
    mux.Handle("/api/users/create", handleCreateUser(store))
    mux.HandleFunc("/healthz", handleHealthz())
    mux.Handle("/", http.NotFoundHandler())
}
```

## 4. Maker Functions Return `http.Handler`

Handlers don't implement the interface directly - they return it. Dependencies injected as arguments.

```go
func handleListUsers(store *Store) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        users, err := store.ListUsers(r.Context())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if err := encode(w, r, http.StatusOK, users); err != nil {
            // log error
        }
    })
}
```

## 5. Encode / Decode Helpers

Generic JSON encoding/decoding in one place.

```go
func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(v); err != nil {
        return fmt.Errorf("encode json: %w", err)
    }
    return nil
}

func decode[T any](r *http.Request) (T, error) {
    var v T
    if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
        return v, fmt.Errorf("decode json: %w", err)
    }
    return v, nil
}
```

## 6. Validator Interface

Simple single-method interface for validation.

```go
type Validator interface {
    Valid(ctx context.Context) (problems map[string]string)
}

func decodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
    var v T
    if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
        return v, nil, fmt.Errorf("decode json: %w", err)
    }
    if problems := v.Valid(r.Context()); len(problems) > 0 {
        return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
    }
    return v, nil, nil
}
```

## 7. Middleware Adapter Pattern

Middleware takes `http.Handler` and returns `http.Handler`.

```go
func loggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            next.ServeHTTP(w, r)
            logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
        })
    }
}
```

Usage in `addRoutes`:
```go
middleware := newMiddleware(logger, db)
mux.Handle("/api/users", middleware(handleListUsers(store)))
```

## 8. Inline Request/Response Types

Hide types inside the handler function to keep global namespace clean.

```go
func handleCreateUser(store *Store) http.Handler {
    type request struct {
        Email       string `json:"email"`
        Username    string `json:"username"`
        DisplayName string `json:"display_name"`
    }
    type response struct {
        ID int64 `json:"id"`
    }
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        req, err := decode[request](r)
        // ...
    })
}
```

## 9. Graceful Shutdown

Use `sync.WaitGroup` to wait for goroutines to finish during shutdown.

```go
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    <-ctx.Done()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil {
        fmt.Fprintf(stderr, "error shutting down: %s\n", err)
    }
}()
wg.Wait()
```

## 10. Testing Through `run()`

Test the whole program through `run()`, not individual handlers.

```go
func TestUserService(t *testing.T) {
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    t.Cleanup(cancel)
    
    getenv := func(key string) string {
        switch key {
        case "HTTP_ADDRESS":
            return ":0"
        case "DATABASE_DSN":
            return testDSN
        default:
            return ""
        }
    }
    
    go func() {
        if err := run(ctx, []string{"userservice"}, getenv, nil, os.Stdout, os.Stderr); err != nil {
            t.Logf("run error: %v", err)
        }
    }()
    
    // wait for ready, then hit endpoints
}
```

## 11. `sync.Once` for Deferred Setup

Defer expensive initialization until first request.

```go
func handleTemplate(files ...string) http.Handler {
    var (
        init   sync.Once
        tpl    *template.Template
        tplerr error
    )
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        init.Do(func() {
            tpl, tplerr = template.ParseFiles(files...)
        })
        if tplerr != nil {
            http.Error(w, tplerr.Error(), http.StatusInternalServerError)
            return
        }
        // use tpl
    })
}
```

## 12. Wait for Readiness

Use a `/healthz` or `/readyz` endpoint to know when the service is up.

```go
func waitForReady(ctx context.Context, timeout time.Duration, endpoint string) error {
    client := http.Client{}
    startTime := time.Now()
    for {
        req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
        if err != nil {
            return fmt.Errorf("create request: %w", err)
        }
        resp, err := client.Do(req)
        if err != nil {
            time.Sleep(250 * time.Millisecond)
            continue
        }
        if resp.StatusCode == http.StatusOK {
            resp.Body.Close()
            return nil
        }
        resp.Body.Close()
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if time.Since(startTime) >= timeout {
                return fmt.Errorf("timeout waiting for endpoint")
            }
            time.Sleep(250 * time.Millisecond)
        }
    }
}
```

## Key Principles

- **Explicit dependencies**: Pass everything as function arguments
- **No globals**: Enables parallel testing
- **Return errors**: Gophers love returning errors
- **Test through `run()`**: End-to-end testing catches more issues
- **Keep `main()` minimal**: Just delegate to `run()`
- **One place for routes**: `addRoutes` maps the entire API surface
- **Hide types**: Inline request/response types inside handlers
