# Monitoring Guide for Sum-100 Game Backend

This guide covers monitoring and maintenance for your production deployment on DigitalOcean.

## Table of Contents

1. [Basic Monitoring](#basic-monitoring)
2. [Log Monitoring](#log-monitoring)
3. [Health Checks](#health-checks)
4. [Performance Monitoring](#performance-monitoring)
5. [Alerting](#alerting)
6. [Troubleshooting](#troubleshooting)

---

## Basic Monitoring

### System Resources

Check CPU, Memory, and Disk usage:

```bash
# Quick overview
htop

# Check disk space
df -h

# Check memory usage
free -h

# Check Docker container stats
docker stats
```

### Container Status

```bash
# List running containers
docker ps

# Check container logs
docker logs sum100-backend-prod

# Follow logs in real-time
docker logs -f sum100-backend-prod

# Check container resource usage
docker stats --no-stream
```

---

## Log Monitoring

### Application Logs

```bash
# View backend logs
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml logs -f backend

# View database logs
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml logs -f db

# View recent logs (last 100 lines)
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml logs --tail=100 backend
```

### Nginx Logs

```bash
# Access logs
tail -f /var/log/nginx/sum100game-access.log

# Error logs
tail -f /var/log/nginx/sum100game-error.log

# Search for errors in last hour
tail -n 1000 /var/log/nginx/sum100game-error.log | grep "error"
```

### System Logs

```bash
# System messages
tail -f /var/log/syslog

# Authentication logs
tail -f /var/log/auth.log

# Kernel messages
dmesg | tail
```

### Fail2Ban Logs

```bash
# View Fail2Ban status
fail2ban-client status

# Check Fail2Ban logs
tail -f /var/log/fail2ban.log

# Check banned IPs
iptables -L -n
```

---

## Health Checks

### Application Health

```bash
# Check if backend is responding
curl http://localhost:8080/health

# Should return: OK
```

### Database Health

```bash
# Check if database is accessible
docker exec sum100-db-prod pg_isready -U postgres

# Should return: accepting connections

# Check database size
docker exec -it sum100-db-prod psql -U postgres -d sum100game -c "SELECT pg_size_pretty(pg_database_size('sum100game'));"
```

### Docker Services Status

```bash
# Check all containers
docker ps -a

# Check specific container health
docker inspect sum100-backend-prod | grep -A 10 Health
```

---

## Performance Monitoring

### Database Performance

```bash
# Connect to database
docker exec -it sum100-db-prod psql -U postgres -d sum100game

# Check active connections
SELECT count(*) FROM pg_stat_activity;

# Check long-running queries
SELECT pid, now() - pg_stat_activity.query_start AS duration, query 
FROM pg_stat_activity 
WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes';

# Check table sizes
SELECT relname AS table_name, 
       pg_size_pretty(pg_total_relation_size(relid)) AS total_size
FROM pg_catalog.pg_statio_user_tables
ORDER BY pg_total_relation_size(relid) DESC;
```

### Application Performance

```bash
# Check response time
time curl http://localhost:8080/health

# Check GraphQL endpoint
time curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __typename }"}'
```

---

## Alerting

### Setup Uptime Monitoring (Free Options)

1. **UptimeRobot** (Free)
   - Register at https://uptimerobot.com
   - Monitor: `https://your-domain.com/health`
   - Get alerts via email, SMS, or push notifications

2. **Pingdom** (Free for 1 monitor)
   - Register at https://www.pingdom.com
   - Monitor: `https://your-domain.com/health`

3. **StatusCake** (Free)
   - Register at https://www.statuscake.com
   - Monitor: `https://your-domain.com/health`

### Custom Alerts

Create a simple monitoring script at `/opt/100sumgame/backend/scripts/monitor.sh`:

```bash
#!/bin/bash

# Simple monitoring script
HEALTH_URL="http://localhost:8080/health"
ALERT_EMAIL="your-email@example.com"

if ! curl -f -s "$HEALTH_URL" > /dev/null; then
    echo "Application is DOWN!" | mail -s "Alert: Sum-100 Game Down" "$ALERT_EMAIL"
    logger "Sum-100 Game Health Check: FAILED"
else
    logger "Sum-100 Game Health Check: OK"
fi
```

Add to crontab to run every 5 minutes:
```bash
*/5 * * * * /opt/100sumgame/backend/scripts/monitor.sh
```

---

## Troubleshooting

### Common Issues

#### 1. Application Not Starting

```bash
# Check logs
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml logs backend

# Check if port is already in use
netstat -tulpn | grep :8080

# Restart container
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml restart backend
```

#### 2. Database Connection Issues

```bash
# Check database status
docker ps | grep sum100-db-prod

# Check database logs
docker logs sum100-db-prod

# Restart database
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml restart db
```

#### 3. High CPU/Memory Usage

```bash
# Check which container is using resources
docker stats

# Check system processes
top

# Restart resource-hungry container
docker-compose -f /opt/100sumgame/backend/docker-compose.prod.yml restart backend
```

#### 4. Disk Space Full

```bash
# Check disk usage
df -h

# Clean Docker resources
docker system prune -a

# Clean old Docker images
docker image prune -a

# Clean old backups
find /opt/100sumgame/backups -name "backup_*.sql.gz" -mtime +30 -delete
```

#### 5. Nginx Issues

```bash
# Check Nginx configuration
nginx -t

# Check Nginx logs
tail -f /var/log/nginx/error.log

# Restart Nginx
systemctl restart nginx
```

---

## Maintenance Tasks

### Daily (Automated via cron)

- Database backup (runs at 2 AM)
- Health check (runs every 5 minutes if configured)

### Weekly (Manual)

```bash
# Check disk space
df -h

# Review logs for errors
grep -i error /var/log/nginx/sum100game-error.log | tail -50

# Check Fail2Ban status
fail2ban-client status

# Update system packages
apt update && apt upgrade -y
```

### Monthly (Manual)

```bash
# Review backup retention
ls -lh /opt/100sumgame/backups/

# Check security updates
apt list --upgradable

# Review system logs
journalctl -p err -n 100

# Check Docker resources
docker system df
```

---

## Useful Commands Reference

### Docker Commands

```bash
# Container management
docker ps                          # List running containers
docker ps -a                       # List all containers
docker stop <container>            # Stop container
docker start <container>           # Start container
docker restart <container>         # Restart container
docker logs -f <container>         # Follow logs
docker exec -it <container> bash   # Access container shell

# Resource management
docker stats                       # Live stats
docker system df                   # Disk usage
docker system prune -a             # Clean unused resources
```

### System Commands

```bash
# Service management
systemctl status nginx             # Check Nginx status
systemctl restart nginx            # Restart Nginx
systemctl reload nginx             # Reload Nginx config
systemctl status fail2ban          # Check Fail2Ban

# Logs
journalctl -u nginx                # Nginx system logs
journalctl -u docker              # Docker system logs
tail -f /var/log/syslog            # System logs

# Firewall
ufw status                         # Check firewall rules
ufw allow <port>                   # Allow port
ufw deny <port>                    # Deny port
```

---

## Monitoring Dashboard (Optional)

### Install Grafana + Prometheus (Advanced)

If you want more advanced monitoring:

```bash
# Install docker-compose monitoring stack
cd /opt/100sumgame
git clone https://github.com/prometheus/prometheus.git

# Or use lightweight alternatives like:
# - Netdata (https://www.netdata.cloud/)
# - Glances (https://nicolargo.github.io/glances/)
```

---

## Contact & Support

If you encounter issues not covered in this guide:

1. Check Docker logs first
2. Review Nginx error logs
3. Check system logs
4. Review application logs
5. Consult the main deployment guide

---

**Last Updated:** 2026-03-09