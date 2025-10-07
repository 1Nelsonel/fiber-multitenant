# Contributing to Fiber Multitenant

Thank you for your interest in contributing! We welcome contributions from the community.

## How to Contribute

### Reporting Issues

If you find a bug or have a feature request:

1. Check if the issue already exists in [GitHub Issues](https://github.com/1Nelsonel/fiber-multitenant/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior
   - Go version and environment details

### Submitting Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit with clear messages (`git commit -m 'Add amazing feature'`)
7. Push to your fork (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Style

- Follow standard Go conventions and `gofmt` formatting
- Write clear comments for exported functions and types
- Keep functions focused and reasonably sized
- Add examples for new features

### Testing

- Write unit tests for new functionality
- Ensure existing tests pass
- Run `go test -v ./...` before submitting

### Documentation

- Update README.md for new features
- Add examples to the `examples/` directory
- Update inline code comments

## Development Setup

```bash
# Clone your fork
git clone https://github.com/1Nelsonel/fiber-multitenant.git
cd fiber-multitenant

# Install dependencies
go mod download

# Run tests
go test -v ./...

# Run example
cd examples/basic
go run main.go
```

## Questions?

Feel free to open an issue for questions or discussions!

## Code of Conduct

Be respectful, inclusive, and professional. We're all here to learn and build together.
