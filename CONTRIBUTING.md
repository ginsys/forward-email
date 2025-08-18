# Contributing to Forward Email CLI

We love your input! We want to make contributing to Forward Email CLI as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Setting Up Development Environment

### Prerequisites

- Go 1.21 or higher
- Make (for build automation)
- Git

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/forward-email.git
cd forward-email

# Install dependencies
make deps

# Set up development tools
make dev-setup

# Run tests
make test

# Build the project
make build
```

## Code Style

We use standard Go formatting and linting tools:

- `go fmt` for formatting
- `golangci-lint` for linting
- `go vet` for static analysis

Run `make lint` to check your code before submitting.

## Testing

We maintain high test coverage standards:

- Unit tests for all public functions
- Integration tests for API interactions
- End-to-end tests for CLI commands

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific test
go test ./pkg/api -v
```

## Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

Examples:
```
feat(api): add domain verification endpoint
fix(cli): resolve authentication timeout issue
docs: update installation instructions
```

## Pull Request Process

1. **Create a feature branch** from `main`
2. **Make your changes** following our coding standards
3. **Add or update tests** as necessary
4. **Update documentation** if you're changing APIs
5. **Run the full test suite** and ensure it passes
6. **Submit a pull request** with a clear title and description

### Pull Request Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] New tests added for new functionality
- [ ] Integration tests updated if necessary

## Checklist
- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
```

## Reporting Bugs

We use GitHub issues to track public bugs. Report a bug by [opening a new issue](https://github.com/ginsys/forward-email/issues/new?template=bug_report.md).

**Great Bug Reports** tend to have:

- A quick summary and/or background
- Steps to reproduce
  - Be specific!
  - Give sample code if you can
- What you expected would happen
- What actually happens
- Notes (possibly including why you think this might be happening, or stuff you tried that didn't work)

## Feature Requests

We welcome feature requests! Please [open an issue](https://github.com/ginsys/forward-email/issues/new?template=feature_request.md) with:

- **Use case**: Describe the problem you're trying to solve
- **Proposed solution**: How you envision the feature working
- **Alternatives considered**: Other approaches you've thought about
- **Additional context**: Screenshots, mockups, or examples

## Code of Conduct

### Our Pledge

We pledge to make participation in our project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

Examples of behavior that contributes to creating a positive environment include:

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

### Enforcement

Project maintainers are responsible for clarifying the standards of acceptable behavior and are expected to take appropriate and fair corrective action in response to any instances of unacceptable behavior.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
