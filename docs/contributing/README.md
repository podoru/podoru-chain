# Contributing to Podoru Chain

Thank you for your interest in contributing to Podoru Chain! This guide will help you get started.

## Code of Conduct

Be respectful and constructive in all interactions.

## Ways to Contribute

### 1. Report Bugs

Found a bug? Please open an issue:

**Include**:
- Clear bug description
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, etc.)
- Logs and error messages

**Example**:
```markdown
**Bug**: Node crashes when querying large prefix

**Steps to Reproduce**:
1. Start node with default config
2. Run: curl -X POST http://localhost:8545/api/v1/state/query/prefix -d '{"prefix": "large:"}'
3. Node crashes

**Expected**: Return results or error
**Actual**: Node crashes with panic

**Environment**:
- OS: Ubuntu 22.04
- Go: 1.24.0
- Podoru Chain: v1.0.0

**Logs**:
```
panic: runtime error: out of memory
...
```
```

### 2. Request Features

Have an idea? Open a feature request:

**Include**:
- Clear feature description
- Use case / motivation
- Proposed implementation (optional)
- Alternatives considered

### 3. Improve Documentation

Documentation contributions are always welcome:

- Fix typos and errors
- Improve clarity
- Add examples
- Translate to other languages

### 4. Submit Code

#### Getting Started

1. **Fork the repository**
   ```bash
   # Click "Fork" on GitHub
   git clone https://github.com/YOUR-USERNAME/podoru-chain.git
   cd podoru-chain
   ```

2. **Create a branch**
   ```bash
   git checkout -b feature/my-feature
   # or
   git checkout -b fix/my-bugfix
   ```

3. **Make changes**
   - Write code
   - Add tests
   - Update documentation

4. **Test your changes**
   ```bash
   # Run tests
   make test

   # Check formatting
   make fmt

   # Build
   make build

   # Test manually
   ./bin/podoru-node -config config/test.yaml
   ```

5. **Commit your changes**
   ```bash
   git add .
   git commit -m "Add feature: description"
   ```

   **Commit Message Guidelines**:
   - Use present tense ("Add feature" not "Added feature")
   - Be descriptive but concise
   - Reference issues (#123)

6. **Push to your fork**
   ```bash
   git push origin feature/my-feature
   ```

7. **Open Pull Request**
   - Go to GitHub
   - Click "New Pull Request"
   - Fill in description template
   - Link related issues

#### Code Style

**Go Code Style**:
```bash
# Format code
make fmt
# or
gofmt -s -w .

# Check with golint
golint ./...
```

**Follow**:
- Standard Go conventions
- Effective Go guidelines
- Keep functions small and focused
- Add comments for exported functions
- Handle errors properly

**Example**:
```go
// CalculateStateRoot computes the Merkle root of the current state.
// Returns an error if the state is empty or corrupted.
func CalculateStateRoot(state map[string][]byte) (string, error) {
    if len(state) == 0 {
        return "", errors.New("empty state")
    }

    // Implementation...
    return root, nil
}
```

#### Testing

**Write Tests**:
```go
// blockchain_test.go
func TestBlockValidation(t *testing.T) {
    block := &Block{
        Height: 1,
        // ...
    }

    err := ValidateBlock(block)
    if err != nil {
        t.Errorf("Expected valid block, got error: %v", err)
    }
}
```

**Run Tests**:
```bash
# All tests
make test

# Specific package
go test ./internal/blockchain

# With coverage
make test-coverage

# Verbose
go test -v ./...
```

#### Pull Request Guidelines

**Before Submitting**:
- [ ] Tests pass
- [ ] Code formatted
- [ ] Documentation updated
- [ ] CHANGELOG updated (if applicable)
- [ ] No merge conflicts

**PR Description Should Include**:
- What changes were made
- Why the changes were made
- How to test the changes
- Screenshots (if UI changes)
- Related issues

**Example PR Description**:
```markdown
## Description
Add support for batch state queries to improve API performance.

## Motivation
Currently, clients must make multiple HTTP requests to query multiple keys, which is inefficient.

## Changes
- Added `POST /api/v1/state/batch` endpoint
- Implemented batch query handler
- Added tests for batch queries
- Updated API documentation

## Testing
```bash
# Run new tests
go test ./internal/api/rest

# Manual test
curl -X POST http://localhost:8545/api/v1/state/batch \
  -d '{"keys": ["key1", "key2", "key3"]}'
```

## Related Issues
Closes #42
```

#### Review Process

1. **Automated Checks** (if configured):
   - CI/CD builds
   - Tests pass
   - Linting passes

2. **Code Review**:
   - Maintainer reviews code
   - Requests changes if needed
   - Approves when ready

3. **Merge**:
   - Squash and merge (usually)
   - Delete branch after merge

**Be Patient**: Reviews may take a few days.

## Development Setup

### Prerequisites

```bash
# Install Go 1.24+
# Install Docker
# Install Make
# Install git
```

### Setup

```bash
# Clone repository
git clone https://github.com/podoru/podoru-chain.git
cd podoru-chain

# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Start development environment
make dev
```

### Project Structure

```
podoru-chain/
├── cmd/                    # Command-line applications
│   ├── node/              # Main node executable
│   └── tools/keygen/      # Key generation tool
├── internal/              # Private application code
│   ├── blockchain/        # Core blockchain logic
│   ├── consensus/         # PoA consensus engine
│   ├── crypto/            # Cryptographic operations
│   ├── storage/           # BadgerDB integration
│   ├── network/           # P2P networking
│   ├── api/rest/          # REST API
│   └── node/              # Node orchestration
├── config/                # Configuration examples
├── docker/                # Docker setup
├── docs/                  # Documentation
├── Makefile              # Build automation
└── README.md
```

### Making Changes

**Adding a Feature**:
1. Define interface
2. Implement functionality
3. Write tests
4. Update documentation
5. Submit PR

**Fixing a Bug**:
1. Write test that reproduces bug
2. Fix bug
3. Verify test passes
4. Submit PR

## Areas Needing Help

### High Priority

- [ ] Improve test coverage
- [ ] Performance optimization
- [ ] Security audit
- [ ] Documentation improvements

### Feature Requests

- [ ] WebSocket support for real-time updates
- [ ] State pruning for disk space optimization
- [ ] Additional query methods
- [ ] CLI improvements
- [ ] Monitoring/metrics

### Documentation

- [ ] More example applications
- [ ] Video tutorials
- [ ] Translation to other languages
- [ ] API client libraries

## Communication

### GitHub

- **Issues**: Bug reports, feature requests
- **Discussions**: General questions, ideas
- **Pull Requests**: Code contributions

### Response Time

- Issues: 1-3 days
- Pull Requests: 3-7 days
- Questions: 1-2 days

## Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file
- Release notes
- GitHub contributors page

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

- Check the [Documentation](../README.md)
- Search [GitHub Issues](https://github.com/podoru/podoru-chain/issues)
- Open a new Discussion
- Contact maintainers

## Thank You!

Every contribution, no matter how small, helps make Podoru Chain better. We appreciate your time and effort!
