# Deployment Guide - Dokploy

This guide covers deploying Oreo Tennis Scoring (OTS) to a VPS using Dokploy.

## Prerequisites

- VPS with Docker installed
- Dokploy installed ([docs.dokploy.com](https://docs.dokploy.com))
- Domain name (optional but recommended)
- SSH access to your VPS

## Step 1: Prepare Secrets

### Generate Admin Password Hash

On your local machine:

```bash
cd backend
go run ../scripts/hash_password.go "YOUR_SECURE_PASSWORD"
```

Copy the hash output - you'll need it for environment variables.

### Generate JWT Secret

```bash
openssl rand -base64 32
```

Save this for the `JWT_SECRET` environment variable.

## Step 2: Push Code to Repository

```bash
git init
git add .
git commit -m "Initial commit"
git remote add origin YOUR_REPO_URL
git push -u origin main
```

## Step 3: Dokploy Setup

1. **Login to Dokploy Dashboard**
   - Access your Dokploy installation (usually `http://your-server-ip:3000`)

2. **Create New Application**
   - Click "New Application"
   - Select "Docker Compose"
   - Connect your Git repository

3. **Configure Build Settings**
   - Repository: `your-repo-url`
   - Branch: `main`
   - Compose File: `docker-compose.prod.yml`

4. **Set Environment Variables**
   
   Navigate to the "Environment" tab and add:

   ```env
   # Database
   POSTGRES_DB=ots
   POSTGRES_USER=ots
   POSTGRES_PASSWORD=<generate-strong-password>
   DATABASE_URL=postgres://ots:<same-password>@postgres:5432/ots?sslmode=disable

   # Admin Credentials
   ADMIN_USERNAME=admin
   ADMIN_PASSWORD_HASH=<your-bcrypt-hash-from-step-1>

   # JWT
   JWT_SECRET=<your-jwt-secret-from-step-1>

   # Frontend/CORS
   FRONTEND_URL=https://your-domain.com
   CORS_ORIGIN=https://your-domain.com
   VITE_API_URL=https://your-domain.com
   ```

## Step 4: Domain Configuration

### Option A: Using Dokploy's Built-in Proxy

1. In Dokploy, go to your application settings
2. Under "Domains", add your domain
3. Enable SSL/TLS (Let's Encrypt)
4. Dokploy will automatically configure nginx reverse proxy

### Option B: Custom Nginx Setup

If using external nginx:

```nginx
# /etc/nginx/sites-available/ots

# API Backend
server {
    listen 443 ssl http2;
    server_name api.your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Frontend
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Step 5: Deploy

1. Click "Deploy" in Dokploy
2. Monitor build logs
3. Wait for all services to be healthy

## Step 6: Verify Deployment

### Check Backend Health

```bash
curl https://your-domain.com/health
# Should return: {"success":true,"data":{"status":"healthy"}}
```

### Test Admin Login

1. Navigate to `https://your-domain.com`
2. Click "Admin"
3. Login with your credentials

### Create Test Data

As admin:
1. Add a player (e.g., "John Doe")
2. Add a venue (e.g., "Central Court", surface: "hard")
3. Logout

### Test Match Flow

1. Return to home
2. Click "Start New Match"
3. Select venue and match type
4. Select players
5. Record some points
6. End match and view summary

## Step 7: PWA Installation

### iOS (Safari)

1. Open site in Safari
2. Tap Share button
3. Select "Add to Home Screen"
4. Tap "Add"

### Android (Chrome)

1. Open site in Chrome
2. Tap menu (⋮)
3. Select "Add to Home screen"
4. Tap "Add"

## Maintenance

### View Logs

```bash
# In Dokploy dashboard
Applications → Your App → Logs

# Or via Docker
docker-compose -f docker-compose.prod.yml logs -f
```

### Database Backup

```bash
# Create backup
docker exec ots-postgres-1 pg_dump -U ots ots > backup_$(date +%Y%m%d).sql

# Restore from backup
docker exec -i ots-postgres-1 psql -U ots ots < backup_20231215.sql
```

### Update Deployment

```bash
# Push changes to git
git add .
git commit -m "Update..."
git push

# In Dokploy, click "Redeploy"
```

## Security Checklist

- [ ] Strong PostgreSQL password
- [ ] Strong admin password with bcrypt hash
- [ ] Secure JWT secret (min 32 chars)
- [ ] HTTPS enabled
- [ ] CORS limited to your domain
- [ ] Firewall configured (only 80, 443, 22 open)
- [ ] Regular backups enabled

## Troubleshooting

### Frontend can't reach backend

- Check `VITE_API_URL` matches backend domain
- Verify CORS settings include frontend domain
- Check nginx/proxy configuration

### Database connection fails

- Verify `DATABASE_URL` format
- Check PostgreSQL is running: `docker ps`
- Review backend logs

### Admin login fails

- Verify bcrypt hash is correct
- Check `ADMIN_USERNAME` and `ADMIN_PASSWORD_HASH` in env
- Review rate limiting (max 5 attempts per 10 seconds)

## Support

For issues, refer to:
- Backend logs: Check HTTP status codes and error messages
- Frontend: Browser console for JavaScript errors
- Database: PostgreSQL logs in Docker

---

**Note**: This is a production application. Never commit `.env` files or expose secrets in logs.
