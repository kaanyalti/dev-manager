# Dev Manager

A command-line tool to manage your development environment, including Git repositories and SSH configurations.

## Features

- **Repository Management**
  - Add repositories to manage
  - List managed repositories
  - Remove repositories
  - Sync repositories (fetch and rebase)
  - Automatic cloning of new repositories
  - Skip sync for repositories with uncommitted changes

- **SSH Key Management**
  - Generate SSH keys
  - Add keys to SSH agent
  - Print public keys
  - Copy public keys to clipboard
  - Remove SSH keys

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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 