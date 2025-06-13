# Dev Manager

A command-line tool for managing development repositories and SSH keys.

## Features

- Repository management (add, remove, sync)
- SSH key management (generate, add to agent, print public key)
- Configuration management
- Tool management

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/kaanyalti/dev-manager.git
   cd dev-manager
   ```

2. Build the project:
   ```bash
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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.