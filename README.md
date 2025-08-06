# 🏭 Enhanced Aluminium Passport System

A comprehensive blockchain-based supply chain transparency platform for aluminium products, built with Go, PostgreSQL, IPFS, and smart contracts on Polygon.

## 🌟 Features

### Core Functionality
- **🔐 Role-Based Authentication**: JWT-based auth with 6 user roles (Admin, Miner, Recycler, Certifier, Manufacturer, Auditor, Viewer)
- **📊 ESG Scoring Engine**: AI-powered Environmental, Social, and Governance scoring
- **🌐 IPFS Integration**: Decentralized storage for passport metadata
- **⛓️ Blockchain Integration**: Smart contracts on Polygon for immutable records
- **📱 QR Code Generation**: GS1-compliant QR codes for supply chain tracking
- **🔍 Zero-Knowledge Proofs**: Privacy-preserving verification system
- **📈 Comprehensive Audit Logging**: Complete traceability of all operations

### Advanced Features
- **📦 Batch Operations**: ZIP file upload for bulk passport creation
- **🎯 Supply Chain Tracking**: End-to-end traceability from mining to recycling
- **📋 Certification Management**: Multi-standard compliance tracking
- **📊 Real-time Analytics**: ESG rankings and performance metrics
- **🔄 Recycling Tracking**: Circular economy support with recycled content tracking
- **🛡️ Security Features**: Rate limiting, CORS, comprehensive validation

## 🏗️ Architecture

```
aluminium-passport/
├── abi/                    # Smart contract Go bindings
├── cmd/migrate/           # Database migration tool
├── contracts/             # Solidity smart contracts
├── internal/
│   ├── auth/             # JWT authentication & password management
│   ├── config/           # Configuration management
│   ├── controller/       # HTTP request controllers
│   ├── db/               # Database models & connection
│   ├── ipfs/             # IPFS client integration
│   ├── middleware/       # HTTP middleware (auth, CORS, logging, rate limiting)
│   ├── qr/               # QR code generation
│   └── routes/           # API route definitions
├── migrations/           # SQL migration files
├── examples/             # Sample data files
├── docker-compose.yml    # Multi-service Docker setup
└── main.go              # Application entry point
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15+
- Node.js (for smart contract deployment)

### 1. Clone & Setup
```bash
git clone <repository-url>
cd AluminiumPassport-aeiforo

# Copy environment file
cp env.example .env
# Edit .env with your configuration
```

### 2. Start Services
```bash
# Start all services (PostgreSQL, IPFS, Redis, Backend)
docker-compose up -d

# Or start with monitoring stack
docker-compose --profile monitoring up -d
```

### 3. Run Migrations
```bash
# Migrations run automatically with Docker
# Or manually:
docker-compose run migrate
```

### 4. Access the API
- **API**: http://localhost:8081
- **Health Check**: http://localhost:8081/health
- **IPFS Gateway**: http://localhost:8080
- **Grafana** (if enabled): http://localhost:3000

## 📡 API Endpoints

### Authentication
```http
POST /api/auth/login          # User login
POST /api/auth/register       # User registration  
POST /api/auth/refresh        # Token refresh
POST /api/auth/logout         # User logout
GET  /api/auth/profile        # Get user profile
```

### Passport Management
```http
POST /api/passports           # Create passport (Miner/Manufacturer)
GET  /api/passports/{id}      # Get passport details
PUT  /api/passports/{id}/recycle # Update recycling info (Recycler)
GET  /api/passports/{id}/qr   # Get QR code
GET  /api/passports           # List passports (paginated)
```

### ESG Management
```http
POST /api/esg/assess          # Create ESG assessment (Certifier)
GET  /api/esg/{id}            # Get ESG metrics
POST /api/esg/generate        # Generate AI-based ESG score
GET  /api/esg/ranking         # Get ESG rankings
```

### Batch Operations
```http
POST /api/batch/upload        # Upload ZIP file (Miner/Manufacturer)
POST /api/batch/validate      # Validate ZIP file
GET  /api/batch/status        # Get batch status
```

### Export & Verification
```http
GET  /api/export/csv          # Export CSV (Auditor/Certifier)
GET  /api/export/json         # Export JSON
POST /api/verify/signature    # Verify digital signature
```

## 🗄️ Database Schema

### Core Tables
- **users**: User accounts with role-based access
- **aluminium_passports**: Main passport data with 40+ fields
- **esg_metrics**: Detailed ESG scoring metrics
- **supply_chain_steps**: Supply chain tracking events
- **audit_logs**: Comprehensive audit trail
- **certifications**: Multi-standard certification tracking
- **batch_operations**: Bulk operation tracking
- **zk_proofs**: Zero-knowledge proof storage

## 🔐 Security Features

- **JWT Authentication** with refresh tokens
- **Role-Based Access Control** (RBAC)
- **Rate Limiting** (per-user and per-role)
- **Password Security** (bcrypt + Argon2 options)
- **CORS Protection** with configurable origins
- **Request Logging** and audit trails
- **Input Validation** and sanitization

## 🌐 Smart Contract Integration

### Polygon Network
```solidity
contract AluminiumPassport {
    // Comprehensive passport data structure
    // Role-based access control
    // ESG scoring integration
    // Supply chain event tracking
    // Recycling content updates
}
```

### Deployment
```bash
# Deploy to Polygon Mumbai (testnet)
cd foundry
forge script script/Deploy.s.sol --rpc-url $MUMBAI_RPC_URL --broadcast

# Update CONTRACT_ADDRESS in .env
```

## 📊 ESG Scoring System

### Scoring Categories
- **Environmental**: Energy efficiency, waste management, carbon emissions, recycled content
- **Social**: Labor practices, community impact, health & safety, human rights
- **Governance**: Transparency, ethics, compliance, stakeholder engagement

### AI-Powered Scoring
- Machine learning algorithms analyze manufacturing data
- Real-time score updates based on supply chain events
- Certification level assignments (Bronze, Silver, Gold, Platinum)
- Automated recommendations for improvement

## 🔄 Supply Chain Tracking

### Lifecycle Stages
1. **Mining/Extraction**: Bauxite sourcing and extraction
2. **Refining**: Alumina production from bauxite
3. **Smelting**: Primary aluminium production
4. **Manufacturing**: Product fabrication
5. **Transportation**: Logistics and shipping
6. **Usage**: Product lifecycle
7. **Recycling**: End-of-life processing

### Traceability Features
- GPS location tracking
- Timestamp verification
- Digital signatures
- IPFS document storage
- QR code integration

## 📱 QR Code System

### GS1 Compliance
- Standard GS1 format support
- Batch and lot tracking
- Expiry date management
- Weight and measurement data

### QR Code Types
- **Passport QR**: Basic passport information
- **ESG QR**: ESG score verification
- **Batch QR**: Batch operation summary
- **GS1 QR**: Supply chain compliant format

## 🐳 Docker Configuration

### Services
- **Backend**: Go API server
- **PostgreSQL**: Primary database
- **IPFS**: Decentralized storage
- **Redis**: Caching and sessions
- **Nginx**: Reverse proxy (production)
- **Prometheus**: Metrics collection
- **Grafana**: Monitoring dashboard

### Environment Profiles
- **Development**: Basic services
- **Production**: + Nginx reverse proxy
- **Monitoring**: + Prometheus & Grafana

## 📈 Monitoring & Analytics

### Metrics Tracked
- API request rates and latencies
- Database connection pools
- IPFS upload/download stats
- User authentication events
- ESG score distributions
- Supply chain event frequencies

### Dashboards
- System performance metrics
- Business intelligence dashboards
- ESG compliance reporting
- Supply chain analytics

## 🧪 Testing

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Load testing
go test -tags=load ./...

# API testing with provided Postman collection
# Import: postman_collection.json
```

## 🚀 Production Deployment

### Infrastructure Requirements
- **Compute**: 2+ CPU cores, 4GB+ RAM
- **Storage**: 100GB+ SSD
- **Network**: HTTPS with valid SSL certificate
- **Database**: PostgreSQL 15+ with backup strategy

### Security Checklist
- [ ] Change default JWT secret
- [ ] Configure proper CORS origins
- [ ] Set up SSL/TLS certificates
- [ ] Enable rate limiting
- [ ] Configure firewall rules
- [ ] Set up monitoring alerts
- [ ] Regular security updates

### Scaling Considerations
- Horizontal API server scaling
- Database read replicas
- IPFS cluster setup
- CDN for static assets
- Load balancer configuration

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [API Docs](API_DOCUMENTATION.md)
- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Email**: support@aluminiumpassport.com

## 🙏 Acknowledgments

- **Empower.eco** - Inspiration for sustainable supply chain tracking
- **Aluminium Stewardship Initiative (ASI)** - Standards and best practices
- **Polygon Network** - Blockchain infrastructure
- **IPFS** - Decentralized storage solution

---

**Built with ❤️ for sustainable supply chains and circular economy**