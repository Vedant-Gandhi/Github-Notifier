# GithubNotifier

GithubNotifier is a lightweight, cross-platform desktop notification service that monitors GitHub repositories for new issues. When a new issue is created, it sends a desktop notification with the issue details and a clickable link to view it.

## Features

- Real-time desktop notifications for new GitHub issues
- Cross-platform support (Windows, macOS, Linux)
- Configurable polling interval
- Clickable notifications that open issues directly in your browser
- Rate limiting and retry mechanisms to handle API failures gracefully
- Support for GitHub's fine-grained personal access tokens
- Minimal resource usage

## Prerequisites

- Go 1.19 or higher
- GitHub Personal Access Token (fine-grained with Issues: Read permission)
- OS-specific notification dependencies:
  - **macOS**: terminal-notifier
  - **Linux**: libnotify
  - **Windows**: No additional dependencies

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/GithubNotifier.git
cd GithubNotifier
```

2. Install OS-specific dependencies:

**macOS**:
```bash
brew install terminal-notifier
```

**Ubuntu/Debian**:
```bash
sudo apt-get install libnotify-bin
```

**Fedora**:
```bash
sudo dnf install libnotify
```

3. Build the application:
```bash
go build
```

## Configuration

1. Create your configuration file:
```bash
cp .env.example .env
```

2. Edit the `.env` file with your settings:
```env
# Required: GitHub repository to monitor
GITHUB_REPO_URL=owner/repo

# Recommended: Your GitHub Personal Access Token
GITHUB_TOKEN=github_pat_your_token_here

# Optional: How often to check for new issues (default: 5m)
POLL_INTERVAL=5m
```

### GitHub Token Setup

1. Go to GitHub → Settings → Developer settings → Personal access tokens → Fine-grained tokens
2. Click "Generate new token"
3. Configure the token:
   - Set a token name (e.g., "GithubNotifier")
   - Choose repository access (Only select repositories)
   - Select the repository you want to monitor
   - Under Permissions → Repository permissions:
     - Issues: Read-only

## Usage

Run the application:
```bash
./GithubNotifier
```

The service will:
- Start monitoring the configured repository for new issues
- Send desktop notifications when new issues are created
- Display the issue number, title, and a clickable link to view the issue
- Continue running until stopped with Ctrl+C

## Troubleshooting

### No Notifications on macOS
- Ensure terminal-notifier is installed: `brew install terminal-notifier`
- Check macOS notification settings for terminal-notifier
- Make sure Do Not Disturb is disabled

### No Notifications on Linux
- Verify libnotify is installed
- Check if your desktop environment supports notifications
- Try running `notify-send "Test" "Message"` to verify notifications work

### API Rate Limiting
- Ensure you've configured a GitHub token
- Check if the token has the correct permissions
- Verify the token isn't expired

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the GNU General Public License v2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [beeep](https://github.com/gen2brain/beeep) for cross-platform notifications
- Uses [godotenv](https://github.com/joho/godotenv) for configuration management