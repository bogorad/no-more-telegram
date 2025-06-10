# Telegram Daemon - Deployment Guide

## Overview

This comprehensive deployment guide provides detailed instructions for deploying the Telegram daemon in various environments, from development testing to production VPS deployment. The daemon has been successfully built and tested with the gotd library.

## Build Status

The daemon has been successfully compiled with the following configuration:

- Go version: 1.21+ (automatically upgraded to 1.23 for compatibility)
- gotd library: v0.100.0 (compatible version)
- Dependencies: All required dependencies resolved and tested

## Deployment Options

### 1. Development/Testing Deployment

For local development and testing:

```bash
# Clone the repository
git clone <repository-url>
cd telegram-daemon

# Build the binary
go mod tidy
go build -o telegram-daemon

# Create configuration
cp config.yaml.example config.yaml
nano config.yaml

# Set required values:
# - app_id: Your Telegram app ID
# - app_hash: Your Telegram app hash
# - phone: Your phone number (+1234567890)
# - password: Your 2FA password (if enabled)

# Run for testing
./telegram-daemon
```

### 2. VPS Production Deployment

#### Option A: Direct Binary Deployment

```bash
# On your VPS, install Go 1.21+
wget https://go.dev/dl/go1.21.13.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.13.linux-amd64.tar.gz
export PATH=/usr/local/go/bin:$PATH

# Clone and build
git clone <repository-url>
cd telegram-daemon
go build -o telegram-daemon

# Configure
cp config.yaml.example config.yaml
nano config.yaml

# Test run
./telegram-daemon

# For production, use systemd service (see Option B)
```

#### Option B: Systemd Service Deployment

```bash
# Build the binary first (as above)
go build -o telegram-daemon

# Install as system service
sudo ./install.sh

# Configure the daemon
sudo cp /opt/telegram-daemon/config.yaml.example /opt/telegram-daemon/config.yaml
sudo nano /opt/telegram-daemon/config.yaml

# Start and enable the service
sudo systemctl start telegram-daemon
sudo systemctl enable telegram-daemon

# Monitor the service
sudo systemctl status telegram-daemon
sudo journalctl -u telegram-daemon -f
```

### 3. Docker Deployment

#### Option A: Docker Compose (Recommended)

```bash
# Prepare configuration
cp config.yaml.example config.yaml
nano config.yaml

# Build and run
docker-compose up -d

# Monitor logs
docker-compose logs -f telegram-daemon

# Stop the service
docker-compose down
```

#### Option B: Direct Docker

```bash
# Build the image
docker build -t telegram-daemon .

# Run with configuration file
docker run -d \
  --name telegram-daemon \
  --restart unless-stopped \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -v $(pwd)/session.json:/app/session.json \
  -v $(pwd)/logs:/app/logs \
  telegram-daemon

# Monitor logs
docker logs -f telegram-daemon
```

### 4. Environment Variable Deployment

For containerized or CI/CD environments:

```bash
export APP_ID="your_app_id"
export APP_HASH="your_app_hash"
export PHONE="+1234567890"
export PASSWORD="your_2fa_password"
export RESPONSE_MSG="I'm no longer using Telegram. Contact me via email."
export LOG_LEVEL="info"
export LOG_FILE="/var/log/telegram-daemon.log"

./telegram-daemon
```

## Initial Authentication

**Important**: The daemon requires interactive authentication on first run to obtain the verification code sent to your phone.

### For VPS Deployment:

1. **Interactive Setup**: Run the daemon interactively once to complete authentication:

   ```bash
   ./telegram-daemon
   # Enter the verification code when prompted
   # Press Ctrl+C after successful authentication
   ```

2. **Session Persistence**: The session is saved to `session.json` and will be reused on subsequent runs.

3. **Daemon Mode**: After initial authentication, restart in daemon mode:
   ```bash
   sudo systemctl start telegram-daemon
   ```

### For Docker Deployment:

1. **Initial Run**: Start container interactively:

   ```bash
   docker run -it --rm \
     -v $(pwd)/config.yaml:/app/config.yaml:ro \
     -v $(pwd)/session.json:/app/session.json \
     telegram-daemon
   ```

2. **Production Run**: After authentication, run in detached mode:
   ```bash
   docker-compose up -d
   ```

## Configuration Management

### Configuration File Priority

The daemon uses the following configuration priority (highest to lowest):

1. Environment variables
2. Configuration file (`config.yaml`)
3. Default values

### Required Configuration

Minimum required configuration:

```yaml
app_id: 123456 # From https://my.telegram.org/apps
app_hash: "your_app_hash" # From https://my.telegram.org/apps
phone: "+1234567890" # Your phone number
```

### Optional Configuration

```yaml
password: "" # 2FA password (if enabled)
session_file: "session.json" # Session storage location
response_message: "Custom message" # Auto-response text
response_timeout_hours: 24 # Rate limiting
log_level: "info" # Logging verbosity
log_file: "" # Log file path (empty = stdout)
```

## Security Considerations

### Credential Security

1. **API Credentials**: Keep `APP_ID` and `APP_HASH` secure
2. **Session File**: Protect `session.json` with appropriate file permissions
3. **Configuration**: Ensure config files are not world-readable

```bash
# Set secure permissions
chmod 600 config.yaml
chmod 600 session.json
```

### Network Security

1. **Firewall**: No inbound ports required (client-only application)
2. **Outbound**: Requires HTTPS access to Telegram servers
3. **VPS**: Consider running in a restricted user account

### Operational Security

1. **Dedicated Account**: Use a separate Telegram account if possible
2. **Monitoring**: Monitor logs for unexpected activity
3. **Updates**: Keep the daemon and dependencies updated

## Monitoring and Maintenance

### Log Monitoring

```bash
# Systemd logs
sudo journalctl -u telegram-daemon -f

# Docker logs
docker-compose logs -f

# File logs (if configured)
tail -f /var/log/telegram-daemon.log
```

### Health Checks

```bash
# Check service status
sudo systemctl status telegram-daemon

# Check process
ps aux | grep telegram-daemon

# Check Docker container
docker ps | grep telegram-daemon
```

### Troubleshooting

#### Common Issues

1. **Authentication Failed**

   - Verify `APP_ID`, `APP_HASH`, and phone number
   - Check if 2FA password is required
   - Ensure session file permissions are correct

2. **Connection Issues**

   - Verify internet connectivity
   - Check firewall settings for outbound HTTPS
   - Confirm Telegram servers are accessible

3. **Permission Denied**
   - Check file permissions for session and log files
   - Verify user has write access to working directory
   - Ensure proper ownership of files

#### Debug Mode

Enable detailed logging:

```bash
export LOG_LEVEL=debug
./telegram-daemon
```

#### Service Recovery

```bash
# Restart systemd service
sudo systemctl restart telegram-daemon

# Restart Docker container
docker-compose restart telegram-daemon

# Check service logs
sudo journalctl -u telegram-daemon --since "1 hour ago"
```

## Performance Optimization

### Resource Usage

- **Memory**: ~50-100MB typical usage
- **CPU**: Minimal when idle, brief spikes during message processing
- **Network**: Low bandwidth usage (text messages only)
- **Storage**: Session file (~1KB), logs (variable)

### Scaling Considerations

- Single instance per Telegram account
- Multiple accounts require separate daemon instances
- Consider rate limiting for high-volume scenarios

## Backup and Recovery

### Critical Files

1. **Session File**: `session.json` - Contains authentication session
2. **Configuration**: `config.yaml` - Contains settings
3. **Logs**: For troubleshooting and audit trails

### Backup Strategy

```bash
# Create backup
tar -czf telegram-daemon-backup-$(date +%Y%m%d).tar.gz \
  config.yaml session.json logs/

# Restore from backup
tar -xzf telegram-daemon-backup-YYYYMMDD.tar.gz
```

### Disaster Recovery

1. **Session Loss**: Re-authenticate interactively
2. **Configuration Loss**: Recreate from template
3. **Complete Loss**: Redeploy and re-authenticate

## Production Checklist

### Pre-Deployment

- [ ] Telegram API credentials obtained
- [ ] Configuration file created and validated
- [ ] Security permissions set correctly
- [ ] Backup strategy implemented
- [ ] Monitoring configured

### Deployment

- [ ] Binary built successfully
- [ ] Initial authentication completed
- [ ] Service starts automatically
- [ ] Logs are being written correctly
- [ ] Auto-response functionality tested

### Post-Deployment

- [ ] Service running stably
- [ ] Logs monitored for errors
- [ ] Response rate limiting working
- [ ] Contact filtering functioning
- [ ] Backup schedule active

## Support and Maintenance

### Regular Maintenance

1. **Weekly**: Check service status and logs
2. **Monthly**: Review and rotate logs
3. **Quarterly**: Update dependencies and rebuild
4. **Annually**: Review and update configuration

### Updates

```bash
# Update dependencies
go mod tidy
go mod download

# Rebuild binary
go build -o telegram-daemon

# Restart service
sudo systemctl restart telegram-daemon
```

This deployment guide provides comprehensive instructions for deploying the Telegram daemon in various environments. The daemon has been successfully tested and is ready for production deployment with proper configuration and security measures.
