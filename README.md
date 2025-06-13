# Dev Manager

A command-line tool for managing development repositories and SSH keys.

## Features

- Repository management (add, remove, sync)
- SSH key management (generate, add to agent, print public key)
- Configuration management
- Tool management
- Dependency management

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/kaanyalti/dev-manager.git
   cd dev-manager
   ```

2. Build the project:
   ```bash
   # Using mage (recommended)
   go install github.com/magefile/mage@latest
   mage build

   # Or using go directly
   go build -o bin/dev-manager ./cmd/dev-manager
   ```

3. Add the binary to your PATH (optional):
   ```bash
   sudo mv bin/dev-manager /usr/local/bin/
   ```

## Usage

### Repository Management

```bash
# Add a repository
dev-manager repos add --name my-project --url https://github.com/username/my-project.git

# List managed repositories
dev-manager repos list

# Remove a repository
dev-manager repos remove --name my-project

# Sync all repositories
dev-manager repos sync
```

### SSH Key Management

```bash
# Generate a new SSH key
dev-manager ssh generate --algo ed25519 --name my-key

# Add a key to SSH agent
dev-manager ssh add-agent --key ~/.ssh/my-key

# Print public key
dev-manager ssh print-public --key ~/.ssh/my-key

# Copy public key to clipboard
dev-manager ssh copy-public --key ~/.ssh/my-key

# Remove a key
dev-manager ssh remove --key ~/.ssh/my-key
```

### Configuration Management

- `dev-manager config show [--raw]`: Display current configuration
  - `--raw`: Show raw YAML content
- `dev-manager config validate [-f|--file <path>]`: Validate configuration file
  - Validates required fields and structure
  - Shows detailed report of any validation errors
  - Example: `dev-manager config validate -f config.yaml`
- `dev-manager init`: Initialize configuration
  - Creates default config file
  - Sets up workspace directory
  - Configures update frequency

### Dependency Management

```bash
# Add a new dependency
dev-manager deps add go@1.21.0

# Install dependencies
dev-manager deps install

# List installed dependencies
dev-manager deps list

# Remove a dependency
dev-manager deps remove go
```

## Planned Features

### Repository Management
- [ ] Interactive repository selection with fuzzy search
- [ ] Support for multiple branches per repository
- [ ] Automatic branch switching based on ticket/issue
- [ ] Repository templates for quick project setup
- [ ] Git hooks management
- [ ] Repository backup and restore
- [ ] Repository health checks (dependencies, security, etc.)

### SSH Management
- [ ] SSH config file management
- [ ] SSH key rotation
- [ ] SSH key backup and restore
- [ ] SSH key usage statistics
- [ ] SSH key expiration management

### Tool Configuration
- [ ] Neovim configuration management
- [ ] Tmux configuration management
- [ ] Zsh configuration management
- [ ] Dotfiles synchronization
- [ ] Configuration templates
- [ ] Configuration versioning

### Dependency Server
- [ ] Centralized registry for development dependencies
- [ ] Smart source resolution for dependencies
- [ ] Version management and platform-specific configurations
- [ ] Community contributions to expand the registry
- [ ] Automatic dependency updates and notifications
- [ ] Dependency health checks and security scanning

### General Improvements
- [ ] Plugin system for extensibility
- [ ] Configuration validation
- [ ] Backup and restore functionality
- [ ] Command aliases for common operations
- [ ] Progress bars for long-running operations
- [ ] Colored output for better readability
- [ ] Shell completion scripts
- [ ] Windows support
- [ ] Docker support for isolated environments

### Git Automation
- [ ] LLM-powered commit message generation
- [ ] Smart branch naming based on ticket/issue
- [ ] Automated PR creation and management
- [ ] Git workflow templates (e.g., GitFlow, GitHub Flow)
- [ ] Automated code review suggestions
- [ ] Git hooks management and templates
- [ ] Commit history analysis and cleanup
- [ ] Branch management and cleanup
- [ ] Automated semantic versioning
- [ ] Changelog generation

### VM Orchestration
- [ ] Cross-platform VM management (QEMU/KVM, VirtualBox, VMware)
- [ ] VM templates for different OS and architectures
- [ ] Automated VM provisioning for cross-compilation
- [ ] VM snapshots and state management
- [ ] VM networking configuration
- [ ] Resource allocation and monitoring
- [ ] Automated testing across different OS environments
- [ ] VM lifecycle management (create, start, stop, delete)
- [ ] VM image management and versioning
- [ ] Integration with cloud providers for remote VM access
- [ ] VM performance optimization settings
- [ ] Automated VM updates and maintenance
- [ ] VM backup and restore functionality
- [ ] VM monitoring and health checks
- [ ] VM resource usage analytics

### Declarative Integration Test Runner
- [ ] Framework for defining and running integration tests in a declarative format
- [ ] Support for test definitions in a structured format
- [ ] Validation of command outputs and system state

A planned feature is a declarative test framework that will generate test code. The framework will:

1. Allow users to define tests in a declarative format (e.g., YAML or JSON)
2. Generate corresponding test code in Go
3. Support user modifications to the generated code
4. Preserve user changes when regenerating tests
5. Update only the non-modified parts of the test code

This will make it easier to:
- Write and maintain tests
- Keep test code consistent
- Allow customization while maintaining automation
- Reduce boilerplate code

## Configuration

The tool uses a YAML configuration file located at `~/.dev-manager/config.yaml`. You can specify a different location using the `--config` flag.

Example configuration:
```yaml
workspace_path: ~/workspace
repositories:
  - name: my-project
    url: https://github.com/username/my-project.git
    path: ~/workspace/my-project
    branch: main
    last_sync: "2024-03-20T10:00:00Z"
dependencies:
  - name: go
    version: 1.21.0
    source: https://go.dev/dl/go1.21.0.darwin-amd64.tar.gz
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.