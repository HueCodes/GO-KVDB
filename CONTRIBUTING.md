# Contributing to Fast-Cache

First off, thank you for considering contributing to Fast-Cache! It's people like you that make Fast-Cache such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our commitment to fostering an open and welcoming environment. We pledge to make participation in our project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, sex characteristics, gender identity and expression, level of experience, education, socio-economic status, nationality, personal appearance, race, religion, or sexual identity and orientation.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* **Use a clear and descriptive title** for the issue
* **Describe the exact steps which reproduce the problem**
* **Provide specific examples to demonstrate the steps**
* **Describe the behavior you observed** and explain which behavior you expected to see and why
* **Include code snippets or test cases** that demonstrate the problem
* **Specify which version of Go you're using** (`go version`)
* **Specify which operating system and version you're using**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. Create an issue and provide the following information:

* **Use a clear and descriptive title**
* **Provide a detailed description of the suggested enhancement**
* **Provide specific examples to demonstrate the use case**
* **Explain why this enhancement would be useful** to most Fast-Cache users

### Pull Requests

1. **Fork the repo** and create your branch from `main`
2. **Add tests** for any new functionality
3. **Ensure the test suite passes** (`go test -v -race ./...`)
4. **Run benchmarks** to verify no performance regressions (`go test -bench=. ./kvcache`)
5. **Update documentation** including README if needed
6. **Follow the Go coding style** (run `gofmt` and `go vet`)
7. **Write clear commit messages**

## Development Setup

### Prerequisites

- Go 1.21 or later (1.18+ for generic API)
- Git

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/Fast-Cache.git
cd Fast-Cache

# Install dependencies (if any)
go mod download

# Run tests
go test -v ./kvcache

# Run tests with race detector
go test -race ./kvcache

# Run benchmarks
go test -bench=. -benchmem ./kvcache
```

### Running Tests

```bash
# Run all tests
go test -v ./kvcache

# Run specific test
go test -v -run TestBasicOperations ./kvcache

# Run with race detector (important!)
go test -race ./kvcache

# Check coverage
go test -cover ./kvcache
go test -coverprofile=coverage.out ./kvcache
go tool cover -html=coverage.out
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem -benchtime=5s ./kvcache

# Run specific benchmark
go test -bench=BenchmarkConcurrentReads -benchmem ./kvcache

# Compare before/after performance
go test -bench=. -benchmem ./kvcache > old.txt
# Make your changes
go test -bench=. -benchmem ./kvcache > new.txt
benchstat old.txt new.txt
```

## Coding Guidelines

### Go Style

- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format your code
- Use `go vet` to check for common mistakes
- Consider using `golangci-lint` for additional linting

### Code Organization

```
Fast-Cache/
├── kvcache/
│   ├── kvcache.go           # Legacy interface{} API
│   ├── kvcache_generic.go   # Generic type-safe API
│   ├── kvcache_test.go      # Tests for legacy API
│   └── kvcache_generic_test.go  # Tests for generic API
├── examples/
│   └── example.go
├── README.md
├── CONTRIBUTING.md
└── LICENSE
```

### Documentation

- Add GoDoc comments to all exported types, functions, and methods
- Include usage examples in doc comments when helpful
- Update README.md for user-facing changes
- Add inline comments for complex logic

**Example:**
```go
// Cache is a type-safe generic cache for Go 1.18+
// It provides high-performance concurrent access with TTL support.
//
// Example usage:
//
//	cache := kvcache.New[string, int](5 * time.Minute)
//	cache.Set("counter", 42)
//	val, ok := cache.Get("counter")
type Cache[K comparable, V any] struct {
    // ...
}
```

### Testing

- **Required:** Add tests for all new functionality
- **Required:** Ensure existing tests pass
- **Required:** Run race detector (`go test -race`)
- **Recommended:** Aim for >80% code coverage
- **Recommended:** Add benchmarks for performance-critical paths

**Test naming:**
```go
func TestCacheConcurrentAccess(t *testing.T) { }      // Good
func TestGenericBasicOperations(t *testing.T) { }    // Good
func Test1(t *testing.T) { }                         // Bad
```

**Benchmark naming:**
```go
func BenchmarkConcurrentReads(b *testing.B) { }      // Good
func BenchmarkGenericGet(b *testing.B) { }           // Good
```

### Commit Messages

Use clear and descriptive commit messages following conventional commits:

```
<type>: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `perf`: Performance improvement
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `style`: Code style changes (formatting, etc.)
- `chore`: Maintenance tasks

**Examples:**
```
feat: add context support to Get operation

Add GetWithContext method that respects context cancellation
and deadlines for graceful timeout handling.

Closes #42

---

fix: resolve use-after-free race in Get()

The Get method had a TOCTOU bug where entries could be returned
to the pool while still being accessed. Fixed by avoiding access
to stale entry references after lock upgrade.

Fixes #38

---

perf: replace O(n) eviction with random sampling

Eviction now samples 5 random entries instead of scanning all
entries, reducing eviction from O(n) to O(1).

Benchmark results:
BenchmarkSet-8  481ns → 78ns (-83%)
```

## Performance Guidelines

### Benchmark Requirements

For performance-related changes, include before/after benchmarks:

```bash
# Before your change
go test -bench=BenchmarkSet -benchmem -count=10 ./kvcache > before.txt

# After your change
go test -bench=BenchmarkSet -benchmem -count=10 ./kvcache > after.txt

# Compare
benchstat before.txt after.txt
```

### Performance Priorities

1. **Correctness first** - Never sacrifice correctness for performance
2. **Profile before optimizing** - Use `pprof` to find actual bottlenecks
3. **Measure everything** - Benchmark before and after
4. **Avoid premature optimization** - Simple code is often faster

### Common Performance Pitfalls

- ❌ Allocating in hot paths
- ❌ Holding locks longer than necessary
- ❌ Using `interface{}` when generics are available
- ❌ Copying large structs instead of using pointers
- ❌ Not using `sync.Pool` for frequently allocated objects

## Pull Request Process

1. **Update tests** - Add tests for new functionality
2. **Update documentation** - Keep README and GoDoc in sync
3. **Run full test suite** - `go test -v -race ./...`
4. **Run benchmarks** - Ensure no regressions
5. **Update CHANGELOG** - If we have one
6. **Request review** - Tag maintainers if needed

### PR Checklist

Before submitting your PR, verify:

- [ ] Code follows Go style guidelines
- [ ] All tests pass (`go test -v ./...`)
- [ ] Race detector passes (`go test -race ./...`)
- [ ] Benchmarks show no regressions
- [ ] New code has test coverage
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] No debugging code left in (e.g., `fmt.Println`)

## Areas for Contribution

Looking for ideas? Here are areas where contributions are welcome:

### High Priority
- [ ] W-TinyLFU admission policy (like Ristretto)
- [ ] Prometheus metrics exporter
- [ ] GetOrSet with singleflight support
- [ ] Improved test coverage (aim for 90%+)
- [ ] Comparison benchmarks vs BigCache, Ristretto

### Medium Priority
- [ ] Compression support (zstd, gzip)
- [ ] Range/Iterator API for debugging
- [ ] Batch operations with atomicity guarantees
- [ ] More eviction policies (LFU, FIFO, Random)
- [ ] OpenTelemetry integration

### Low Priority
- [ ] Persistence layer (optional disk backup)
- [ ] Distributed mode with clustering
- [ ] Bloom filters for negative lookups
- [ ] SIMD-optimized hashing

## Questions?

Feel free to open an issue with the `question` label if you have any questions about contributing.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Happy coding! 🚀**
