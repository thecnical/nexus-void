# Contributing to Nexus Void

First off, thank you for considering contributing to Nexus Void! It's people like
you that make this tool a powerful asset for the cybersecurity community.

## How Can I Contribute?

### Reporting Bugs

- Check if the bug has already been reported in [Issues](https://github.com/thecnical/nexus-void/issues)
- If not, open a new issue with:
  - Clear title and description
  - Steps to reproduce
  - Expected vs actual behavior
  - System information (OS, Go version)
  - Relevant logs or screenshots

### Suggesting Enhancements

- Open an issue with the `enhancement` label
- Describe the feature and its use case
- Explain why it would be useful to most users

### Pull Requests

1. Fork the repository
2. Create a branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `go test ./...`
5. Ensure code is formatted: `gofmt -w .`
6. Commit with clear messages
7. Push to your fork and open a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/nexus-void.git
cd nexus-void

# Build CLI
go build -o nexus-void ./cmd/nexus-void

# Build backend
cd backend && go build -o server ./cmd/server

# Run tests
go test ./...
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Add comments for exported functions
- Keep functions focused and small
- Write tests for new features

### Commit Message Format

```
type: short description

Longer explanation if needed.

Fixes #123
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

## Community

- GitHub Discussions: [thecnical/nexus-void/discussions](https://github.com/thecnical/nexus-void/discussions)
- Website: [cybermindcli.com](https://cybermindcli.com)

## Recognition

Contributors will be listed in the README and release notes.

---

*Created by Chandan Pandey | cybermindcli.com*
