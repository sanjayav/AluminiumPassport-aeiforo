# Aluminium Passport API Documentation

## Overview
The Aluminium Passport API provides endpoints for managing aluminium supply chain transparency through blockchain technology and digital passports.

## Base URL
```
http://localhost:8080/api
```

## Authentication
All API endpoints (except login) require JWT authentication via Bearer token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## Endpoints

### Authentication

#### POST /api/auth/login
Login to get JWT token.

**Request Body:**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Available Users:**
- `admin` / `admin123` (Role: issuer)
- `auditor` / `audit123` (Role: auditor)  
- `viewer` / `view123` (Role: viewer)

---

### Single Passport Operations

#### POST /api/passports
Create a new passport (Issuer only).

**Request Body:**
```json
{
  "passport_id": "ALU-PASS-001",
  "batch_id": "BATCH-2024-001",
  "bauxite_origin": "Western Australia",
  "mine_operator": "BHP Billiton",
  "date_of_extraction": "2024-01-15",
  "refinery_location": "Queensland Alumina Refinery",
  "carbon_emissions_per_kg": 2.1,
  "alloy_composition": "Al-99.8%, Si-0.1%, Fe-0.1%",
  "esg_score": 87.5,
  "certification_agency": "ASI"
}
```

#### GET /api/passports/{id}
Get passport by ID (All roles).

**Response:**
```json
{
  "passport_id": "ALU-PASS-001",
  "batch_id": "BATCH-2024-001",
  "bauxite_origin": "Western Australia",
  "carbon_emissions_per_kg": 2.1,
  "esg_score": 87.5,
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### Batch Operations (ZIP File Upload)

#### POST /api/batch/upload
Upload ZIP file containing multiple passport data files (Issuer only).

**Content-Type:** `multipart/form-data`

**Form Data:**
- `zip_file`: ZIP file containing JSON/CSV files with passport data

**Response:**
```json
{
  "batch_id": "BATCH_1640995200000",
  "total_processed": 3,
  "successful": 3,
  "failed": 0,
  "errors": [],
  "ipfs_hash": "ipfs://QmXxx..."
}
```

**Supported File Formats in ZIP:**
- `.json` - JSON array of passport objects or single passport object
- `.csv` - CSV file with header row and passport data

#### POST /api/batch/validate
Validate ZIP file without processing (Issuer, Auditor).

**Content-Type:** `multipart/form-data`

**Form Data:**
- `zip_file`: ZIP file to validate

**Response:**
```json
{
  "valid": true,
  "filename": "passports.zip",
  "size": 1024000,
  "max_size": 52428800,
  "message": "ZIP file is valid and ready for processing"
}
```

#### GET /api/batch/status
Get batch processing status (All roles).

**Query Parameters:**
- `batch_id`: Batch ID to check status for

**Response:**
```json
{
  "batch_id": "BATCH_1640995200000",
  "status": "completed",
  "total_count": 3,
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### GET /api/batch/template
Download template files for batch uploads (Issuer, Auditor).

**Query Parameters:**
- `format`: Template format (`csv` or `json`)

**Response:** File download (CSV or JSON template)

---

### Export Operations

#### GET /api/export/csv
Export passport data as CSV (Auditor, Issuer).

**Query Parameters:**
- `batch_id`: Optional batch ID filter

**Response:** CSV file download

#### GET /api/export/json
Export passport data as JSON (Auditor, Issuer).

**Query Parameters:**
- `batch_id`: Optional batch ID filter

**Response:** JSON file download

---

### Advanced Features

#### POST /api/verify/signature
Verify digital signatures (All roles).

**Request Body:**
```json
{
  "data": "passport data",
  "signature": "digital signature",
  "public_key": "public key"
}
```

**Response:**
```json
{
  "valid": true,
  "message": "Signature verification completed"
}
```

#### GET /api/generate/qr/{id}
Generate QR code for passport (All roles).

**Response:** PNG image (QR code)

#### POST /api/zk/generate
Generate zero-knowledge proof (Issuer only).

**Request Body:**
```json
{
  "passport_id": "ALU-PASS-001",
  "claims": {
    "carbon_emissions_below": 3.0,
    "esg_score_above": 80.0
  }
}
```

**Response:**
```json
{
  "proof": "zk_proof_data",
  "message": "ZK proof generated successfully"
}
```

#### POST /api/zk/verify
Verify zero-knowledge proof (All roles).

**Request Body:**
```json
{
  "proof": "zk_proof_data",
  "claims": {
    "carbon_emissions_below": 3.0
  }
}
```

**Response:**
```json
{
  "valid": true,
  "message": "ZK proof verification completed"
}
```

---

### Audit and Reporting

#### GET /api/audit/logs
Get audit logs (Auditor, Issuer).

**Query Parameters:**
- `user`: Filter by user
- `action`: Filter by action type
- `limit`: Number of records (default: 100)

**Response:**
```json
[
  {
    "timestamp": "2024-01-01T00:00:00Z",
    "user": "admin",
    "role": "issuer",
    "action": "BATCH_UPLOAD",
    "target": "BATCH001"
  }
]
```

---

## Error Responses

All endpoints return standard HTTP status codes:

- `200` - Success
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

**Error Response Format:**
```json
{
  "error": "Error message description"
}
```

---

## File Format Specifications

### CSV Format
The CSV file must include a header row with the following columns (case-insensitive):

```csv
passport_id,batch_id,bauxite_origin,mine_operator,date_of_extraction,refinery_location,refiner_id,smelting_energy_source,carbon_emissions_per_kg,alloy_composition,trace_metals,manufacturer_id,process_type,manufactured_product,product_weight,energy_used,manufacturing_emissions,transport_mode,distance_travelled,logistics_partner_id,shipment_date,recycled_content_percent,recycling_date,recycler_id,recycling_method,times_recycled,certification_agency,esg_score,compliance_standards,date_of_certification,verifier_signature
```

### JSON Format
JSON files can contain either:
1. A single passport object
2. An array of passport objects

**Single Object:**
```json
{
  "passport_id": "ALU-PASS-001",
  "batch_id": "BATCH-2024-001",
  "bauxite_origin": "Western Australia",
  ...
}
```

**Array:**
```json
[
  {
    "passport_id": "ALU-PASS-001",
    ...
  },
  {
    "passport_id": "ALU-PASS-002", 
    ...
  }
]
```

---

## Rate Limits
- ZIP file uploads: Maximum 50MB
- Maximum 1000 files per ZIP
- Maximum 100 audit log records per request

---

## Examples

See the `/examples` directory for:
- `sample_passports.json` - JSON format example
- `sample_passports.csv` - CSV format example
- `sample_upload.zip` - Complete ZIP upload example