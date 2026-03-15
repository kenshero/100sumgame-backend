# Complete Deployment Guide - Sum-100 Game Backend

This guide walks you through deploying your Sum-100 Game backend to DigitalOcean Singapore.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Infrastructure Setup](#infrastructure-setup)
3. [Server Initial Setup](#server-initial-setup)
4. [Application Deployment](#application-deployment)
5. [Nginx & SSL Setup](#nginx--ssl-setup)
6. [Security Hardening](#security-hardening)
7. [Backup Setup](#backup-setup)
8. [Monitoring Setup](#monitoring-setup)
9. [Testing & Verification](#testing--verification)
10. [Maintenance & Updates](#maintenance--updates)

---

## Prerequisites

### Before You Start

- [ ] DigitalOcean account (sign up at https://digitalocean.com)
- [ ] Domain name (optional, but recommended)
- [ ] Git installed locally
- [ ] SSH access configured
- [ ] Gemini API key ready
- [ ] Estimated budget: ~200-300 บาท/เดือน

### Local Requirements

- SSH key pair (recommended) or password
- Git installed
- Text editor (VS Code, Nano, etc.)

---

## Infrastructure Setup

### Step 1: Create DigitalOcean Account

1. Sign up at https://digitalocean.com
2. Add billing method (credit card or PayPal)
3. Get $100 free credit (if using referral)

### Step 2: Create Droplet (VM)

1. Login to DigitalOcean dashboard
2. Click **Create** → **Droplets**
3. Configure droplet:
   - **Region:** Singapore (sgp1) - IMPORTANT for ASEAN users
   - **Image:** Ubuntu 22.04 LTS x64
   - **Size:** Basic - $6/month (1 vCPU, 1GB RAM, 25GB SSD)
   - **Authentication:** 
     - Option A: SSH Key (recommended)
     - Option B: Password
   - **Hostname:** sum100game-backend
   - **Tags:** production, backend

4. Click **Create Droplet**
5. Wait 1-2 minutes for droplet to be ready
6. Note the **IP address** (e.g., `xxx.xxx.xxx.xxx`)

### Step 3: Setup DNS (Optional but Recommended)

If you have a domain:

1. Go to **Networking** in DigitalOcean
2. Add your domain
3. Create **A record**:
   - Type: A
   - Name: `api` (or `@` for root)
   - Value: Your droplet IP
   - TTL: 3600

Your API will be accessible at: `https://api.yourdomain.com` or `https://yourdomain.com`

---

## Server Initial Setup

### Step 1: SSH into Server

```bash
# Using SSH key
ssh root@your-droplet-ip

# Using password
ssh root@your-droplet-ip
```

### Step 2: Update System

```bash
apt update && apt upgrade -y
```

### Step 3: Set Hostname

```bash
hostnamectl set-hostname sum100game-backend
```

### Step 4: Set Timezone

```bash
timedatectl set-timezone Asia/Bangkok
```

### Step 5: Install Essential Tools

```bash
apt install -y docker.io docker-compose git curl wget vim htop
```

### Step 6: Enable Docker

```bash
systemctl enable docker
systemctl start docker

# Verify Docker is running
docker --version
docker-compose --version
```

### Step 7: Configure Firewall (Basic)

```bash
# Install UFW if not installed
apt install -y ufw

# Allow SSH
ufw allow 22/tcp

# Allow HTTP
ufw allow 80/tcp

# Allow HTTPS
ufw allow 443/tcp

# Enable firewall
ufw enable

# Check status
ufw status
```

---

## Application Deployment

### Step 1: Clone Repository

```bash
# Create project directory
mkdir -p /opt/100sumgame
cd /opt/100sumgame

# Clone repository
git clone https://github.com/kenshero/100sumgame-backend.git
cd backend
```

### Step 2: Setup Environment Variables

```bash
# Create environment file from example
cp devops/.env.prod.example devops/.env.prod

# Edit environment file
nano devops/.env.prod
```

**Configure the following variables:**

```env
# Server
PORT=8080
ENVIRONMENT=production

# Database (will be created by docker-compose)
DATABASE_URL=postgres://postgres:CHANGE_ME_PASSWORD@db:5432/sum100game?sslmode=disable

# Gemini AI
GEMINI_API_KEY=your-actual-gemini-api-key

# Session Secret (IMPORTANT!)
SESSION_SECRET_KEY=generate-a-random-32-char-secret-key
```

**Generate a secure secret key:**

```bash
# On your local machine or server
openssl rand -base64 32
```

### Step 3: Setup Database Password

Edit `docker-compose.prod.yml`:

```bash
nano docker-compose.prod.yml
```

Update the database service environment:

```yaml
db:
  image: postgres:16-alpine
  container_name: sum100-db-prod
  restart: always
  environment:
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: your-secure-password-here
    POSTGRES_DB: sum100game
```

Then update `devops/.env.prod`:

```env
DATABASE_URL=postgres://postgres:your-secure-password-here@db:5432/sum100game?sslmode=disable
```

### Step 4: Create Backup Directory

```bash
mkdir -p /opt/100sumgame/backups
```

### Step 5: Build and Start Services

```bash
# Build and start containers
docker-compose -f docker-compose.prod.yml up -d --build

# Check containers are running
docker ps

# Check logs
docker-compose -f docker-compose.prod.yml logs -f
```

Wait for containers to start. You should see:
- `sum100-db-prod` running
- `sum100-backend-prod` running

### Step 6: Verify Backend is Working

```bash
# Check health endpoint
curl http://localhost:8080/health

# Should return: OK
```

### Step 7: Run Initial Database Setup

```bash
# Run migrations
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/001_create_puzzle_pool.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/002_create_game_sessions.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/003_create_leaderboard.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/004_add_guest_id.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/006_add_guest_puzzle_progress.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/007_add_puzzle_status.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/008_add_solved_positions.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/009_add_puzzle_sets.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/010_add_game_settings.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/011_add_stamina_and_score.sql

# Seed initial data
docker exec -i sum100-db-prod psql -U postgres -d sum100game < scripts/seed_puzzles.sql
```

---

## Nginx & SSL Setup

### Step 1: Install Nginx

```bash
apt install -y nginx
```

### Step 2: Copy Nginx Configuration

```bash
# Copy nginx config from repository
cp devops/deploy/nginx/nginx.conf /etc/nginx/sites-available/sum100game
```

### Step 3: Configure Server Name

```bash
# Edit nginx config
nano /etc/nginx/sites-available/sum100game
```

Replace `your-domain.com` with:
- Your actual domain (e.g., `api.yourdomain.com`)
- OR your droplet IP (if no domain)

### Step 4: Enable Site

```bash
# Create symlink
ln -s /etc/nginx/sites-available/sum100game /etc/nginx/sites-enabled/

# Remove default site
rm /etc/nginx/sites-enabled/default

# Test configuration
nginx -t

# Reload Nginx
systemctl reload nginx
```

### Step 5: Test HTTP Access

```bash
# Test from server
curl http://localhost/health

# Test from your local machine
curl http://your-droplet-ip/health
```

### Step 6: Setup SSL (Optional but Recommended)

**If you have a domain:**

```bash
# Install Certbot
apt install -y certbot python3-certbot-nginx

# Obtain SSL certificate
certbot --nginx -d your-domain.com

# Follow prompts:
# - Enter email for renewal notifications
# - Agree to Terms of Service
# - Choose whether to redirect HTTP to HTTPS (recommended)
```

**After SSL is configured:**

```bash
# Test SSL configuration
nginx -t

# Reload Nginx
systemctl reload nginx

# Test HTTPS access
curl https://your-domain.com/health
```

**Auto-renewal is configured automatically by Certbot.**

**If you don't have a domain:**

Skip SSL setup for now. Your application will work over HTTP. You can add SSL later when you have a domain.

---

## Security Hardening

### Step 1: Run Security Setup Script

```bash
# Make script executable
chmod +x /opt/100sumgame/backend/devops/deploy/scripts/security-setup.sh

# Run security setup
bash /opt/100sumgame/backend/devops/deploy/scripts/security-setup.sh
```

This script will:
- Update system packages
- Configure UFW firewall
- Install and configure Fail2Ban
- Harden SSH configuration
- Enable automatic security updates
- Create a deployment user
- Setup log rotation
- Install monitoring tools

### Step 2: SSH Key Setup (Important!)

**Before restarting SSH, ensure you have SSH keys configured:**

**On your local machine:**

```bash
# Generate SSH key pair (if you don't have one)
ssh-keygen -t ed25519 -C "your-email@example.com"

# Copy public key to server
ssh-copy-id root@your-droplet-ip

# Test SSH login
ssh root@your-droplet-ip
```

**Restart SSH only after confirming key-based login works:**

```bash
# On the server
systemctl restart sshd

# Test SSH login from your local machine
ssh root@your-droplet-ip
```

**Note:** If you get locked out, you can always access droplet via DigitalOcean console.

---

## Backup Setup

### Step 1: Setup Backup Script

```bash
# Make backup script executable
chmod +x /opt/100sumgame/backend/scripts/backup.sh

# Test backup script
bash /opt/100sumgame/backend/scripts/backup.sh
```

### Step 2: Setup Automated Backups (Cron)

```bash
# Open crontab
crontab -e

# Add this line to run backup daily at 2 AM
0 2 * * * /opt/100sumgame/backend/scripts/backup.sh >> /opt/100sumgame/backups/backup.log 2>&1
```

### Step 3: Verify Backup Schedule

```bash
# List cron jobs
crontab -l

# Check backup directory
ls -lh /opt/100sumgame/backups/
```

### Step 4: Test Backup Restore (Optional)

```bash
# Create a test backup
bash /opt/100sumgame/backend/scripts/backup.sh

# List backups
ls -lh /opt/100sumgame/backups/

# To restore (in case of emergency):
# gunzip < backup_YYYYMMDD_HHMMSS.sql.gz | docker exec -i sum100-db-prod psql -U postgres -d sum100game
```

---

## Monitoring Setup

### Step 1: Install Monitoring Tools

```bash
apt install -y htop iotop sysstat
```

### Step 2: Setup Uptime Monitoring

Sign up for free uptime monitoring:

1. **UptimeRobot** (https://uptimerobot.com)
   - Monitor: `https://your-domain.com/health` (or `http://your-ip/health`)
   - Interval: 5 minutes
   - Notification: Email, SMS, or push

2. **Alternative:** Pingdom (1 free monitor), StatusCake (free)

### Step 3: Review Monitoring Guide

Read the complete monitoring guide:

```bash
cat /opt/100sumgame/backend/devops/deploy/MONITORING.md
```

---

## Testing & Verification

### Step 1: Test Application Health

```bash
# Health endpoint
curl http://localhost:8080/health

# GraphQL endpoint
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __typename }"}'
```

### Step 2: Test Nginx Reverse Proxy

```bash
# From server
curl http://localhost/health

# From your local machine
curl http://your-droplet-ip/health
# Or with domain:
curl https://your-domain.com/health
```

### Step 3: Test GraphQL Playground

Open in browser:
- `http://your-droplet-ip/` (HTTP)
- `https://your-domain.com/` (HTTPS with SSL)

### Step 4: Verify Database Connection

```bash
# Check database is accessible
docker exec sum100-db-prod pg_isready -U postgres

# Check database size
docker exec -it sum100-db-prod psql -U postgres -d sum100game -c "SELECT pg_size_pretty(pg_database_size('sum100game'));"
```

### Step 5: Check Container Resources

```bash
# View container stats
docker stats

# View container logs
docker logs sum100-backend-prod --tail=50
```

### Step 6: Test Security

```bash
# Check firewall status
ufw status

# Check Fail2Ban status
fail2ban-client status

# Check SSH configuration
cat /etc/ssh/sshd_config
```

---

## Maintenance & Updates

### Daily Deployments

When you need to deploy updates:

```bash
# SSH into server
ssh root@your-droplet-ip

# Run deployment script
cd /opt/100sumgame/backend
bash scripts/deploy.sh
```

The deployment script will:
1. Create database backup
2. Pull latest code from GitHub
3. Stop existing containers
4. Build and start new containers
5. Wait for services to be ready
6. Run health checks
7. Reload Nginx

### Weekly Maintenance

```bash
# Check disk space
df -h

# Check system resources
htop

# Review logs
docker-compose -f docker-compose.prod.yml logs --tail=100

# Update system packages
apt update && apt upgrade -y

# Check backups
ls -lh /opt/100sumgame/backups/
```

### Monthly Maintenance

```bash
# Review security updates
apt list --upgradable

# Clean Docker resources
docker system prune -a

# Review Fail2Ban logs
fail2ban-client status

# Check Nginx logs
tail -100 /var/log/nginx/sum100game-error.log
```

### Useful Commands

```bash
# View logs
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml logs -f backend

# Restart services
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml restart

# Stop services
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml down

# Start services
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml up -d

# Check Nginx status
systemctl status nginx

# Restart Nginx
systemctl restart nginx

# Check firewall
ufw status
```

---

## Cost Summary

### Monthly Costs

- **Droplet:** $6/month (~200 บาท)
- **Domain (optional):** $10-15/year (~30-45 บาท/เดือน)
- **Total:** ~200-250 บาท/เดือน

### Scaling Options

If traffic increases:

1. **Upgrade Droplet:**
   - $6 → $12 (2GB RAM) - Medium traffic
   - $12 → $24 (4GB RAM) - High traffic
   - $24 → $48 (8GB RAM) - Very high traffic

2. **DigitalOcean App Platform:**
   - Starting at $10/month
   - Better for scaling
   - More features

3. **Add CDN:**
   - Cloudflare CDN (free)
   - Improves performance in ASEAN
   - Provides DDoS protection

---

## Troubleshooting

### Common Issues

**1. Container not starting:**
```bash
# Check logs
docker-compose -f docker-compose.prod.yml logs backend

# Restart container
docker-compose -f docker-compose.prod.yml restart backend
```

**2. Database connection error:**
```bash
# Check database is running
docker ps | grep sum100-db-prod

# Check database logs
docker logs sum100-db-prod

# Restart database
docker-compose -f docker-compose.prod.yml restart db
```

**3. Nginx 502 Bad Gateway:**
```bash
# Check if backend is running
curl http://localhost:8080/health

# Check Nginx logs
tail -f /var/log/nginx/error.log

# Reload Nginx
systemctl reload nginx
```

**4. Out of disk space:**
```bash
# Check disk usage
df -h

# Clean Docker resources
docker system prune -a

# Clean old backups
find /opt/100sumgame/backups -name "backup_*.sql.gz" -mtime +30 -delete
```

**5. SSH connection refused:**
```bash
# Access via DigitalOcean console
# Check SSH status
systemctl status sshd

# Restart SSH
systemctl restart sshd
```

---

## Support & Resources

### Documentation

- **Monitoring Guide:** `/opt/100sumgame/backend/devops/deploy/MONITORING.md`
- **Project README:** `/opt/100sumgame/backend/README.md`
- **DigitalOcean Docs:** https://docs.digitalocean.com

### Getting Help

1. Check logs first
2. Review troubleshooting section
3. Check DigitalOcean community forums
4. Create GitHub issue if it's a code problem

---

## Next Steps

### Immediate (After Deployment)

1. ✅ Test all endpoints
2. ✅ Verify SSL is working (if domain configured)
3. ✅ Setup uptime monitoring
4. ✅ Test backup and restore
5. ✅ Document any custom configurations

### Short-term (First Week)

1. Monitor application performance
2. Review logs regularly
3. Setup alerting for critical issues
4. Document any issues encountered

### Long-term (Ongoing)

1. Regular system updates
2. Monitor resource usage
3. Plan scaling strategy
4. Consider CI/CD automation
5. Set up more advanced monitoring

---

## Success! 🎉

Your Sum-100 Game backend is now deployed and running on DigitalOcean Singapore!

**Access your API:**
- HTTP: `http://your-droplet-ip/`
- HTTPS: `https://your-domain.com/` (if configured)

**GraphQL Playground:**
- `http://your-droplet-ip/`
- `https://your-domain.com/`

**Health Check:**
- `http://your-droplet-ip/health`
- `https://your-domain.com/health`

---

**Last Updated:** 2026-03-09
**Version:** 1.0