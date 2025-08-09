# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

logv2fs is a Go-based VPN/proxy server built on sing-box with a React frontend. It features hybrid MongoDB/PostgreSQL architecture, real-time WebSocket updates, and comprehensive user/node management.

## Common Development Commands

### Backend Development
```bash
# Start main application services
./main httpserver          # Web API server with frontend
./main singbox             # sing-box proxy server with traffic logging
make backend               # Development API server
make singbox               # Production sing-box with full features

# Build with required tags for full functionality
export GOOS=linux; export GOARCH=amd64
go build -tags="with_gvisor,with_quic,with_wireguard,with_utls,with_reality_server,with_clash_api,with_v2ray_api,with_grpc" -o ./logv2fs ./
```

### Frontend Development
```bash
cd frontend
npm i && npm run build     # Production build
make web                   # Development server with API proxy
```

### Database Operations
```bash
# Migration commands (MongoDB → PostgreSQL)
./main migrate --type=full --batch-size=1000   # Full migration with batching
./main migrate --type=schema                    # Structure only
./main migrate --type=data                      # Data only

# Specialized migrations
./main migrate_user_traffic    # User traffic logs
./main migrate_node_traffic    # Node statistics
./main migrate_payment         # Payment records

# Data synchronization
./main sync --direction=mongo-to-postgres
```

## Architecture Overview

### Core Components
- **Main Application**: Cobra CLI with multiple subcommands (httpserver, singbox, migrate, sync)
- **Database Layer**: Hybrid MongoDB + PostgreSQL with automatic fallback
- **Proxy Engine**: sing-box with custom traffic logging integration
- **Real-time Layer**: WebSocket + PostgreSQL LISTEN/NOTIFY for live updates
- **Frontend**: React SPA with Redux state management

### Database Architecture
The application uses a sophisticated dual-database pattern:

**MongoDB Collections:**
- `USER_TRAFFIC_LOGS` - User data and traffic logs
- `NODE_TRAFFIC_LOGS` - Server statistics
- `subscription_nodes` - Proxy configurations
- `payment_records` - Billing and payment history
- `daily_payment_allocations` - Daily payment summaries
- `expiry_check_domains` - Domain expiry tracking

**PostgreSQL Tables (Migration Target):**
- `user_traffic_logs` - Relational data + JSONB time-series
- `node_traffic_logs` - Aggregated node statistics
- `payment_records` + `daily_payment_allocations` - Billing system
- `subscription_nodes` - Proxy configurations with geo-blocking
- `expiry_check_domains` - Domain expiry tracking

### Key Controllers and Models
- `controllers/controller_pg.go` - PostgreSQL operations
- `controllers/controller.go` - MongoDB operations  
- `model/postgres_models.go` - PostgreSQL schema definitions
- `websocket/` - Real-time notification system

## Development Patterns

### Database Pattern Selection
The codebase automatically detects and switches between databases:
```go
if isUsingPostgreSQL() {
    // Use PostgreSQL controller methods
} else {
    // Fallback to MongoDB controller methods
}
```

### Traffic Logging Integration
Traffic data is collected from sing-box API every 15 minutes via cron and stored with proper user association. The system handles both real-time logging and batch aggregation.

### Configuration Generation
The system generates multiple proxy client configurations:
- **sing-box**: JSON format with geo-blocking rules (`config/template_singbox.json`)
- **Clash Verge rev**: YAML format (`config/template_clash.yaml`)  
- **Shadowrocket**: Base64 encoded subscription URLs

### Real-time Features
WebSocket implementation provides live updates for:
- User traffic statistics
- Node health monitoring  
- Payment record changes
- User status modifications

## API Structure

### Authentication
- JWT token-based auth with bcrypt password hashing
- Role-based access (admin vs normal users)
- Protected routes with middleware validation

### Key Endpoints
```
POST /v1/signup              # User registration
PUT /v1/edit/:name          # User management
GET /v1/user/:name          # User details
GET /v1/getconfig/:name     # Generate proxy config
POST /admin/manage          # Node management (obfuscated endpoint)
```

## Environment Configuration

### Required Environment Variables
```bash
# Database selection
USE_POSTGRES=true
postgresURI=postgres://guestuser:@localhost:5432/logv2fs?sslmode=disable&TimeZone=Asia/Shanghai
mongoURI = "mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+1.5.4"

# Application settings  
CURRENT_DOMAIN="localhost"
SERVER_PORT=8079
GIN_MODE=release

# sing-box integration
SING_BOX_TEMPLATE_CONFIG=./config/template_singbox.json
```

### Build Tags
Always include full feature tags for production builds:
```
with_gvisor,with_quic,with_wireguard,with_utls,with_reality_server,with_clash_api,with_v2ray_api,with_grpc
```

## Migration Strategy

The migration system implements gradual MongoDB→PostgreSQL transition:

1. **Schema Phase**: Create PostgreSQL tables with proper indexes
2. **Data Phase**: Batch transfer with resume capability  
3. **Verification Phase**: Data integrity checks
4. **Switch Phase**: Update application to use PostgreSQL

Migration is resumable and supports different batch sizes for performance tuning.

## Testing and Deployment

### Local Development
```bash
make backend && make web   # Concurrent development
```

### Production Deployment
```bash
# Docker deployment
docker-compose up -d
```

## IPv6 Support

The application fully supports IPv6:
- Node IP configurations accept both IPv4 and IPv6
- Subscription URL generation handles IPv6 address formatting
- Database schemas accommodate longer IP address fields
- Helper functions ensure proper IPv6 URL construction

## Security Considerations

- Input sanitization for all user inputs using `github.com/mrz1836/go-sanitize`
- SQL injection prevention with prepared statements
- Rate limiting and CORS protection
- Automatic SSL certificate management
- Password hashing with bcrypt
- JWT token expiration and refresh handling