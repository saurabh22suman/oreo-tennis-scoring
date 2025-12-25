# Dokploy Deployment Guide for OTS

## Prerequisites

- Dokploy instance running
- Domain name configured
- SSL certificate (Dokploy handles this automatically)

## Deployment Steps

### 1. Create a New Project in Dokploy

1. Log into your Dokploy dashboard
2. Create a new project named "oreo-tennis-scoring"
3. Select "Docker Compose" as the deployment type

### 2. Configure Environment Variables

In the Dokploy environment variables section, add the following:

#### Database
```
POSTGRES_DB=ots
POSTGRES_USER=ots
POSTGRES_PASSWORD=<generate-secure-password>
DATABASE_URL=postgres://ots:<your-password>@postgres:5432/ots?sslmode=disable
```

#### Admin Credentials
```
ADMIN_USERNAME=admin
ADMIN_PASSWORD_HASH=<bcrypt-hash-of-your-password>
JWT_SECRET=<generate-random-string-32-chars>
```

To generate bcrypt hash:
```bash
# Using Go
echo -n "your_password" | docker run -i --rm alpine/openssl passwd -6 -stdin

# Or using htpasswd
htpasswd -bnBC 10 "" your_password | tr -d ':\n'
```

To generate JWT secret:
```bash
openssl rand -hex 32
```

#### Frontend URLs
```
FRONTEND_URL=https://your-domain.com
CORS_ORIGIN=https://your-domain.com
VITE_API_URL=https://your-domain.com/api
```

Replace `your-domain.com` with your actual domain.

### 3. Upload Docker Compose File

1. In Dokploy, navigate to your project
2. Upload or paste the contents of `docker-compose.prod.yml`
3. Save the configuration

### 4. Configure Domain & SSL

1. Go to the "Domains" tab in Dokploy
2. Add your domain name
3. Enable SSL (Dokploy will use Let's Encrypt automatically)
4. Configure routing:
   - `/` → frontend service (port 80)
   - `/api/*` → backend service (port 8080)

### 5. Deploy

1. Click "Deploy" in Dokploy
2. Wait for the build to complete
3. Monitor logs for any errors

### 6. Verify Deployment

Check the following endpoints:

- **Frontend**: `https://your-domain.com`
- **Backend Health**: `https://your-domain.com/api/health`
- **API Test**: `https://your-domain.com/api/players`

## Post-Deployment

### Database Migrations

The application automatically runs migrations on startup. Check the backend logs to verify:

```bash
# In Dokploy logs
2024/12/24 Database migrations completed
2024/12/24 Server starting on port 8080
```

### Create Admin User

The admin user is created automatically using the `ADMIN_USERNAME` and `ADMIN_PASSWORD_HASH` environment variables.

Login at: `https://your-domain.com` → Click "Admin"

### PWA Installation

Users can install the app on their devices:
- **Mobile**: Tap "Add to Home Screen" in browser menu
- **Desktop**: Click install icon in address bar

## Monitoring

### Health Checks

Dokploy automatically monitors:
- **Postgres**: Every 10 seconds
- **Backend**: Every 30 seconds
- **Frontend**: Depends on backend

### Logs

View logs in Dokploy:
1. Select your project
2. Choose service (postgres/backend/frontend)
3. View real-time logs

## Updating

To update the application:

1. Push changes to your Git repository
2. In Dokploy, click "Rebuild"
3. Wait for deployment to complete
4. Verify the update

## Troubleshooting

### Backend can't connect to database

- Check `DATABASE_URL` format
- Verify postgres service is healthy
- Check network connectivity

### Frontend can't reach API

- Verify `VITE_API_URL` is correct
- Check CORS settings in `CORS_ORIGIN`
- Ensure domain routing is configured

### Admin login fails

- Verify `ADMIN_PASSWORD_HASH` is correct bcrypt hash
- Check `JWT_SECRET` is set
- Ensure password matches the hash

## Backup Strategy

### Database Backup

Dokploy provides automatic volume backups. Additionally:

```bash
# Manual backup
docker exec <postgres-container> pg_dump -U ots ots > backup.sql

# Restore
docker exec -i <postgres-container> psql -U ots ots < backup.sql
```

### Volume Locations

- **Database**: `postgres_data` volume
- Managed by Dokploy

## Security Notes

1. **Always use HTTPS** in production (Dokploy handles this)
2. **Strong passwords** for database and admin
3. **Rotate JWT_SECRET** periodically
4. **Backup database** regularly
5. **Monitor logs** for suspicious activity

## Support

- GitHub: https://github.com/saurabh22suman/oreo-tennis-scoring
- Documentation: See `README.md` and `PROJECT_SUMMARY.md`

---

**Note**: Replace all `<placeholders>` and `your-domain.com` with actual values before deploying.
