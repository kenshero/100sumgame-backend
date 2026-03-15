# Deployment Files for Sum-100 Game Backend

This directory contains all the necessary files and scripts for deploying your Sum-100 Game backend to production.

## 📁 Directory Structure

```
devops/deploy/
├── README.md                  # This file
├── DEPLOYMENT_GUIDE.md       # Complete step-by-step deployment guide
├── MONITORING.md             # Monitoring and maintenance guide
├── nginx/
│   └── nginx.conf            # Nginx configuration for reverse proxy
└── scripts/
    ├── backup.sh             # Automated database backup script
    ├── deploy.sh             # Application deployment script
    └── security-setup.sh     # Security hardening script
```

## 🚀 Quick Start

### 1. Read the Deployment Guide

Start with `DEPLOYMENT_GUIDE.md` for a complete, step-by-step guide to deploying your application.

### 2. Follow These Steps

1. **Infrastructure Setup** - Create DigitalOcean droplet
2. **Server Setup** - Install Docker, configure firewall
3. **Application Deployment** - Clone repo, configure environment, start services
4. **Nginx & SSL** - Setup reverse proxy and SSL certificates
5. **Security Hardening** - Run security setup script
6. **Backup & Monitoring** - Setup automated backups and monitoring

### 3. Use the Scripts

All scripts are designed to be run on the server:

```bash
# Security setup (run once)
bash /opt/100sumgame/backend/devops/deploy/scripts/security-setup.sh

# Database backup (manual or automated via cron)
bash /opt/100sumgame/backend/scripts/backup.sh

# Deploy updates (run when you have code changes)
bash /opt/100sumgame/backend/scripts/deploy.sh
```

## 📋 File Descriptions

### DEPLOYMENT_GUIDE.md
Complete deployment guide covering:
- Infrastructure setup on DigitalOcean
- Server initial configuration
- Application deployment
- Nginx and SSL setup
- Security hardening
- Backup setup
- Monitoring setup
- Testing and verification
- Maintenance procedures
- Troubleshooting common issues

**Read this first!**

### MONITORING.md
Comprehensive monitoring guide covering:
- Basic monitoring (system resources, containers)
- Log monitoring (application, Nginx, system)
- Health checks
- Performance monitoring
- Alerting setup
- Troubleshooting common issues
- Maintenance tasks (daily, weekly, monthly)
- Useful command reference

**Read this after deployment is complete.**

### nginx/nginx.conf
Nginx configuration for:
- Reverse proxy to Go backend
- HTTP/HTTPS setup
- Security headers
- GraphQL endpoint
- Health check endpoint
- SSL certificate support (Let's Encrypt)

**Place this at `/etc/nginx/sites-available/sum100game` on the server.**

### scripts/backup.sh
Automated database backup script that:
- Creates timestamped database backups
- Compresses backups with gzip
- Cleans up old backups (7-day retention by default)
- Logs backup operations

**Setup:** Run daily via cron job at 2 AM.

### scripts/deploy.sh
Automated deployment script that:
- Creates database backup before deployment
- Pulls latest code from GitHub
- Stops existing containers
- Builds and starts new containers
- Waits for services to be ready
- Runs health checks
- Reloads Nginx

**Use this when deploying updates.**

### scripts/security-setup.sh
Security hardening script that:
- Updates system packages
- Installs security tools (fail2ban, ufw)
- Configures UFW firewall
- Configures Fail2Ban for SSH and Nginx
- Hardens SSH configuration
- Enables automatic security updates
- Creates deployment user
- Sets up log rotation
- Installs monitoring tools

**Run this once after initial deployment.**

## 💰 Cost Summary

- **DigitalOcean Droplet:** $6/month (~200 บาท)
- **Domain (optional):** $10-15/year (~30-45 บาท/เดือน)
- **Total:** ~200-250 บาท/เดือน

## 🔒 Security Features

All deployment files include security best practices:

- ✅ Firewall configuration (UFW)
- ✅ SSH hardening (key-based authentication)
- ✅ Fail2Ban for brute force protection
- ✅ SSL/TLS encryption (Let's Encrypt)
- ✅ Security headers in Nginx
- ✅ Automated security updates
- ✅ Regular database backups
- ✅ Log rotation
- ✅ Non-root container execution

## 📊 Monitoring & Maintenance

### Daily (Automated)
- Database backup at 2 AM
- Health checks (if configured)

### Weekly (Manual)
- Check disk space
- Review logs for errors
- Update system packages
- Check Fail2Ban status

### Monthly (Manual)
- Review security updates
- Clean Docker resources
- Review backup retention
- Check system logs

## 🛠️ Useful Commands

```bash
# View application logs
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml logs -f backend

# View database logs
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml logs -f db

# Check container status
docker ps

# Restart services
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml restart

# Check system resources
htop

# Check disk space
df -h

# View Nginx logs
tail -f /var/log/nginx/sum100game-error.log

# Check firewall status
ufw status

# Check Fail2Ban status
fail2ban-client status
```

## 📞 Support & Resources

### Documentation
- **Deployment Guide:** `DEPLOYMENT_GUIDE.md`
- **Monitoring Guide:** `MONITORING.md`
- **Project README:** `/README.md`

### External Resources
- **DigitalOcean Docs:** https://docs.digitalocean.com
- **Nginx Docs:** https://nginx.org/en/docs/
- **Docker Docs:** https://docs.docker.com/
- **PostgreSQL Docs:** https://www.postgresql.org/docs/

## 🎯 Next Steps

1. ✅ Read `DEPLOYMENT_GUIDE.md`
2. ✅ Create DigitalOcean account and droplet
3. ✅ Follow the deployment guide step by step
4. ✅ Setup SSL certificate (if you have a domain)
5. ✅ Run security setup script
6. ✅ Setup automated backups
7. ✅ Read `MONITORING.md` for ongoing maintenance
8. ✅ Test all endpoints and features

## 📝 Notes

- All scripts are designed to be run on the server
- Make sure to make scripts executable: `chmod +x script.sh`
- Always backup your database before running deployments
- Test SSL configuration in a staging environment if possible
- Monitor resource usage, especially in the first week after deployment
- Keep your system packages updated regularly

## 🎉 Success!

Your Sum-100 Game backend is production-ready with:
- ✅ Automated deployment
- ✅ Automated backups
- ✅ Security hardening
- ✅ SSL/TLS support
- ✅ Monitoring tools
- ✅ Comprehensive documentation

---

**Last Updated:** 2026-03-09
**Version:** 1.0