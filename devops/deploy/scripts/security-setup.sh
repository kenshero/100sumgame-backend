#!/bin/bash

# Security Hardening Script for Sum-100 Game Server
# Place this at /opt/100sumgame/backend/scripts/security-setup.sh on the server
# Make executable: chmod +x security-setup.sh
# Run as root or with sudo

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo "========================================="
echo "🔒 Security Hardening Setup"
echo "========================================="
echo "Time: $(date)"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root${NC}"
    exit 1
fi

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Step 1: Update system
print_info "Step 1: Updating system packages..."
apt update && apt upgrade -y
print_success "System updated"

# Step 2: Install security tools
print_info "Step 2: Installing security tools..."
apt install -y fail2ban ufw htop nmap
print_success "Security tools installed"

# Step 3: Configure firewall
print_info "Step 3: Configuring UFW firewall..."

# Reset firewall
ufw --force reset

# Default policies
ufw default deny incoming
ufw default allow outgoing

# Allow SSH (port 22)
ufw allow 22/tcp comment 'SSH'

# Allow HTTP
ufw allow 80/tcp comment 'HTTP'

# Allow HTTPS
ufw allow 443/tcp comment 'HTTPS'

# Enable firewall
ufw --force enable

print_success "Firewall configured"
print_info "Current rules:"
ufw status numbered

# Step 4: Configure Fail2Ban
print_info "Step 4: Configuring Fail2Ban..."

# Create local jail config
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5
destemail = root@localhost

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600

[nginx-http-auth]
enabled = true
port = http,https
filter = nginx-http-auth
logpath = /var/log/nginx/*error.log
maxretry = 5

[nginx-noscript]
enabled = true
port = http,https
filter = nginx-noscript
logpath = /var/log/nginx/*error.log
maxretry = 6
bantime = 86400
EOF

# Start Fail2Ban
systemctl enable fail2ban
systemctl start fail2ban

print_success "Fail2Ban configured and started"

# Step 5: Secure SSH configuration
print_info "Step 5: Securing SSH configuration..."

# Backup original SSH config
cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup.$(date +%Y%m%d)

# Configure SSH
sed -i 's/#PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config

# Additional security settings
cat >> /etc/ssh/sshd_config << 'EOF'

# Security settings
Protocol 2
ClientAliveInterval 300
ClientAliveCountMax 2
MaxAuthTries 3
EOF

print_warning "SSH configured to:"
echo "  - Disable root login"
echo "  - Disable password authentication"
echo "  - Require SSH key authentication"
echo ""
print_warning "IMPORTANT: Make sure you have SSH keys set up before restarting SSH!"
print_info "You'll need to restart SSH manually: systemctl restart sshd"

# Step 6: Configure automatic security updates
print_info "Step 6: Configuring automatic security updates..."

apt install -y unattended-upgrades

cat > /etc/apt/apt.conf.d/50unattended-upgrades << 'EOF'
Unattended-Upgrade::Allowed-Origins {
    "${distro_id}:${distro_codename}";
    "${distro_id}:${distro_codename}-security";
};
Unattended-Upgrade::AutoFixInterruptedDpkg "true";
Unattended-Upgrade::MinimalSteps "true";
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "false";
Unattended-Upgrade::Automatic-Reboot-Time "02:00";
EOF

cat > /etc/apt/apt.conf.d/20auto-upgrades << 'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Download-Upgradeable-Packages "1";
APT::Periodic::AutocleanInterval "7";
APT::Periodic::Unattended-Upgrade "1";
EOF

systemctl enable unattended-upgrades
systemctl start unattended-upgrades

print_success "Automatic security updates configured"

# Step 7: Create secure user for deployment
print_info "Step 7: Creating deployment user (if not exists)..."

if ! id -u deploy > /dev/null 2>&1; then
    useradd -m -s /bin/bash deploy
    usermod -aG docker deploy
    print_success "User 'deploy' created"
    print_info "Set password for deploy user: passwd deploy"
else
    print_info "User 'deploy' already exists"
fi

# Step 8: Configure log rotation
print_info "Step 8: Configuring log rotation..."

cat > /etc/logrotate.d/docker-containers << 'EOF'
/var/lib/docker/containers/*/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    copytruncate
}
EOF

print_success "Log rotation configured"

# Step 9: Install monitoring tools
print_info "Step 9: Installing monitoring tools..."

apt install -y sysstat iotop
systemctl enable sysstat
systemctl start sysstat

print_success "Monitoring tools installed"

# Step 10: Security summary
echo ""
echo "========================================="
echo "✅ Security Hardening Completed!"
echo "========================================="
echo ""
print_info "Summary of changes:"
echo "  ✓ System packages updated"
echo "  ✓ Firewall configured (UFW)"
echo "  ✓ Fail2Ban installed and configured"
echo "  ✓ SSH hardened (requires manual restart)"
echo "  ✓ Automatic security updates enabled"
echo "  ✓ Log rotation configured"
echo "  ✓ Monitoring tools installed"
echo ""
print_warning "IMPORTANT NEXT STEPS:"
echo "  1. Ensure SSH key authentication is working"
echo "  2. Restart SSH: systemctl restart sshd"
echo "  3. Test SSH login with new configuration"
echo "  4. Set up deploy user SSH key: ssh-copy-id deploy@localhost"
echo ""
print_info "Useful commands:"
echo "  - Check firewall: ufw status"
echo "  - Check Fail2Ban: fail2ban-client status"
echo "  - View Fail2Ban logs: tail -f /var/log/fail2ban.log"
echo "  - Check SSH config: cat /etc/ssh/sshd_config"
echo ""