# Quick Start Guide - Sum-100 Game Deployment

This is a condensed version of the deployment guide for quick reference.

## 🚀 5-Minute Overview

### What You Need
- DigitalOcean account (signup: https://digitalocean.com)
- ~200-300 บาท/เดือน budget
- 30-60 minutes time

### What You Get
- Production-ready backend on DigitalOcean Singapore
- Automated backups
- SSL/TLS (if you have a domain)
- Security hardening
- Monitoring tools

---

## 📋 Quick Steps

### 1. Create DigitalOcean Droplet (5 minutes)

1. Login to DigitalOcean
2. Create Droplet:
   - Region: Singapore (sgp1)
   - Image: Ubuntu 22.04 LTS
   - Size: $6/month (1 vCPU, 1GB RAM, 25GB SSD)
   - Authentication: SSH Key (recommended) or Password
3. Wait for droplet to be ready
4. Note the IP address

### 2. SSH & Setup Server (5 minutes)

```bash
# SSH into server
ssh root@your-droplet-ip

# Update system
apt update && apt upgrade -y

# Install tools
apt install -y docker.io docker-compose git curl

# Enable Docker
systemctl enable docker
systemctl start docker

# Set timezone
timedatectl set-timezone Asia/Bangkok

# Configure firewall
apt install -y ufw
ufw allow 22/tcp  # SSH
ufw allow 80/tcp  # HTTP
ufw allow 443/tcp # HTTPS
ufw enable
```

### 3. Deploy Application (10 minutes)

```bash
# Clone repository
mkdir -p /opt/100sumgame
cd /opt/100sumgame
git clone https://github.com/kenshero/100sumgame-backend.git
cd backend

# Setup environment
cp devops/.env.prod.example devops/.env.prod
nano devops/.env.prod

# IMPORTANT: Edit these values:
# - DATABASE_URL (use same password as in docker-compose.prod.yml)
# - GEMINI_API_KEY (your actual key)
# - SESSION_SECRET_KEY (generate with: openssl rand -base64 32)

# Update database password in docker-compose.prod.yml
nano docker-compose.prod.yml
# Change POSTGRES_PASSWORD to a secure password

# Create backup directory
mkdir -p /opt/100sumgame/backups

# Start services
docker-compose -f docker-compose.prod.yml up -d --build

# Wait for containers to start
docker ps

# Run database setup
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/001_create_puzzle_pool.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/002_create_game_sessions.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/003_create_leaderboard.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/004_add_guest_id.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/006_add_guest_puzzle_progress.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/007_add_puzzle_status.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/008_add_solved_positions.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/009_add_puzzle_sets.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/010_create_game_settings.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < internal/database/migrations/011_add_stamina_and_score.sql
docker exec -i sum100-db-prod psql -U postgres -d sum100game < scripts/seed_puzzles.sql

# Test backend
curl http://localhost:8080/health
# Should return: OK
```

### 4. Setup Nginx (5 minutes)

```bash
# Install Nginx
apt install -y nginx

# Copy config
cp devops/deploy/nginx/nginx.conf /etc/nginx/sites-available/sum100game

# Edit config
nano /etc/nginx/sites-available/sum100game
# Replace "your-domain.com" with your domain or droplet IP

# Enable site
ln -s /etc/nginx/sites-available/sum100game /etc/nginx/sites-enabled/
rm /etc/nginx/sites-enabled/default

# Test and reload
nginx -t
systemctl reload nginx

# Test from your local machine
curl http://your-droplet-ip/health
```

### 5. Setup SSL (Optional, 5 minutes)

**If you have a domain:**

```bash
# Install Certbot
apt install -y certbot python3-certbot-nginx

# Get SSL certificate
certbot --nginx -d your-domain.com

# Follow prompts and choose to redirect HTTP to HTTPS
```

**If you don't have a domain:** Skip this step. Your app will work over HTTP.

### 6. Security Hardening (5 minutes)

```bash
# Make scripts executable
chmod +x /opt/100sumgame/backend/scripts/backup.sh
chmod +x /opt/100sumgame/backend/devops/deploy/scripts/security-setup.sh
chmod +x /opt/100sumgame/backend/devops/deploy/scripts/deploy.sh

# Run security setup
bash /opt/100sumgame/backend/devops/deploy/scripts/security-setup.sh

# IMPORTANT: Setup SSH keys before restarting SSH!
# On your local machine:
ssh-keygen -t ed25519 -C "your-email@example.com"
ssh-copy-id root@your-droplet-ip

# Test SSH login with key
ssh root@your-droplet-ip

# If key-based login works, restart SSH on server:
systemctl restart sshd
```

### 7. Setup Automated Backups (2 minutes)

```bash
# Test backup script
bash /opt/100sumgame/backend/scripts/backup.sh

# Setup cron job
crontab -e
# Add this line:
0 2 * * * /opt/100sumgame/backend/scripts/backup.sh >> /opt/100sumgame/backups/backup.log 2>&1
```

---

## ✅ Verification

Test everything works:

```bash
# From server
curl http://localhost/health
curl http://localhost:8080/health

# From your local machine
curl http://your-droplet-ip/health
# or with SSL:
curl https://your-domain.com/health

# Open GraphQL Playground in browser
http://your-droplet-ip/
# or:
https://your-domain.com/
```

---

## 🔄 Updating Your App

When you have code changes:

```bash
# SSH into server
ssh root@your-droplet-ip

# Run deployment script
cd /opt/100sumgame/backend
bash scripts/deploy.sh
```

That's it! The script handles:
- Database backup
- Pulling latest code
- Rebuilding containers
- Health checks
- Nginx reload

---

## 📊 Monitoring Commands

```bash
# Check container status
docker ps

# View logs
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml logs -f backend

# Check system resources
htop

# Check disk space
df -h

# View Nginx logs
tail -f /var/log/nginx/sum100game-error.log

# Check firewall
ufw status

# Check Fail2Ban
fail2ban-client status
```

---

## 💰 Costs

- **Droplet:** $6/month (~200 บาท)
- **Domain (optional):** $10-15/year (~30-45 บาท/เดือน)
- **Total:** ~200-250 บาท/เดือน

---

## 🆘 Common Issues

**Container not starting:**
```bash
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml logs backend
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml restart backend
```

**Database connection error:**
```bash
docker ps | grep sum100-db-prod
docker logs sum100-db-prod
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml restart db
```

**Nginx 502 error:**
```bash
curl http://localhost:8080/health
tail -f /var/log/nginx/error.log
systemctl reload nginx
```

**Out of disk space:**
```bash
df -h
docker system prune -a
find /opt/100sumgame/backups -name "backup_*.sql.gz" -mtime +30 -delete
```

---

## 📚 Full Documentation

For detailed instructions, troubleshooting, and advanced setup:

- **Complete Guide:** `DEPLOYMENT_GUIDE.md`
- **Monitoring Guide:** `MONITORING.md`
- **Overview:** `README.md`

---

## 🎯 Next Steps After Deployment

1. ✅ Test all endpoints
2. ✅ Verify SSL (if configured)
3. ✅ Setup uptime monitoring (UptimeRobot - free)
4. ✅ Test backup and restore
5. ✅ Monitor logs and resources for first week

---

## 🎉 You're Done!

Your Sum-100 Game backend is now live!

**Access URLs:**
- HTTP: `http://your-droplet-ip/`
- HTTPS: `https://your-domain.com/` (if configured)
- GraphQL Playground: Same as above

**Need Help?**
- Check `DEPLOYMENT_GUIDE.md` for detailed steps
- Check `MONITORING.md` for troubleshooting
- Check DigitalOcean community forums

---

**Time to Deploy:** ~30-60 minutes
**Cost:** ~200-250 บาท/เดือน
**Difficulty:** Easy to Medium

---

**Last Updated:** 2026-03-09