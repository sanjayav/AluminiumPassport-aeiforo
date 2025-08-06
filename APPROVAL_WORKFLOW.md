# 🔐 Approval Workflow System

## Overview

The Enhanced Aluminium Passport System now includes a **hierarchical approval workflow** where Super Admins must approve supplier onboarding requests made by Admins.

## 👥 Role Hierarchy

### 1. **Super Admin** (`super_admin`)
- **Highest privilege level** (100)
- Can approve all admin actions
- Can approve supplier onboarding
- Full system access
- Default user: `superadmin` / `superadmin@aluminiumpassport.com`

### 2. **Admin** (`admin`)
- **System administration** (90)
- Can request supplier onboarding (requires Super Admin approval)
- Manages users and system operations
- Cannot directly onboard suppliers
- Default user: `admin` / `admin@aluminiumpassport.com`

### 3. **Supplier Roles** (require Super Admin approval)
- **Certifier** (`certifier`) - Level 80
- **Miner** (`miner`) - Level 60
- **Manufacturer** (`manufacturer`) - Level 60
- **Recycler** (`recycler`) - Level 50

### 4. **Other Roles**
- **Auditor** (`auditor`) - Level 70
- **Viewer** (`viewer`) - Level 10

## 🔄 Supplier Onboarding Workflow

### Step 1: Admin Request
```http
POST /api/approvals/supplier-onboarding
Authorization: Bearer <admin-token>

{
  "wallet_address": "0x1234567890123456789012345678901234567890",
  "username": "mining_corp",
  "email": "contact@miningcorp.com",
  "role": "miner",
  "company_name": "Global Mining Corporation",
  "company_type": "Mining Operations",
  "business_license": "LICENSE123456",
  "contact_info": {
    "phone": "+1234567890",
    "address": "123 Mining Street, City, Country"
  },
  "justification": "Large scale bauxite mining operation with 20+ years experience"
}
```

**Response:**
```json
{
  "message": "Supplier onboarding request created successfully",
  "approval_request_id": 1,
  "status": "pending_super_admin_approval",
  "expires_at": "2024-01-08T00:00:00Z",
  "approver_role": "super_admin"
}
```

### Step 2: Super Admin Reviews
```http
GET /api/approvals?for_approval=true
Authorization: Bearer <super-admin-token>
```

**Response:**
```json
{
  "requests": [
    {
      "id": 1,
      "request_type": "supplier_onboarding",
      "title": "Supplier Onboarding: Global Mining Corporation (miner)",
      "description": "Admin admin requests onboarding of Global Mining Corporation as miner...",
      "status": "pending",
      "requested_by_user": {
        "username": "admin",
        "role": "admin"
      },
      "expires_at": "2024-01-08T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Step 3: Super Admin Approves/Rejects
```http
POST /api/approvals/1/approve
Authorization: Bearer <super-admin-token>

{
  "action": "approved",
  "reason": "Valid mining operation with proper documentation"
}
```

**Response:**
```json
{
  "message": "Request approved successfully",
  "status": "approved",
  "reason": "Valid mining operation with proper documentation",
  "actioned_by": "superadmin",
  "actioned_at": "2024-01-01T12:00:00Z"
}
```

## 📊 Database Schema

### Approval Requests Table
```sql
CREATE TABLE approval_requests (
    id SERIAL PRIMARY KEY,
    request_type approval_type NOT NULL,
    requested_by INTEGER REFERENCES users(id),
    approver_role user_role NOT NULL,
    approved_by INTEGER REFERENCES users(id),
    status approval_status DEFAULT 'pending',
    title VARCHAR(255) NOT NULL,
    description TEXT,
    request_data JSONB,
    approval_reason TEXT,
    rejection_reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP WITH TIME ZONE
);
```

### Pending Users Table
```sql
CREATE TABLE pending_users (
    id SERIAL PRIMARY KEY,
    approval_request_id INTEGER REFERENCES approval_requests(id),
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    requested_role user_role NOT NULL,
    company_name VARCHAR(255),
    company_type VARCHAR(100),
    business_license VARCHAR(255),
    contact_info JSONB,
    justification TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 🔐 API Endpoints

### Approval Management
- `POST /api/approvals` - Create approval request (Admin+)
- `POST /api/approvals/supplier-onboarding` - Request supplier onboarding (Admin only)
- `GET /api/approvals` - List approval requests (filtered by role)
- `GET /api/approvals/{id}` - Get specific approval request
- `POST /api/approvals/{id}/approve` - Approve request (Super Admin)
- `POST /api/approvals/{id}/reject` - Reject request (Super Admin)

### Role-Based Access
- `GET /api/admin/*` - Admin and Super Admin access
- `GET /api/super-admin/*` - Super Admin only access

## 🔄 Workflow States

### Approval Request Status
- `pending` - Awaiting approval
- `approved` - Approved by authorized user
- `rejected` - Rejected with reason
- `expired` - Request expired (default 7 days)

### Process Flow
1. **Admin creates** supplier onboarding request
2. **System creates** pending user record
3. **Super Admin reviews** request details
4. **Super Admin approves/rejects** with reason
5. **If approved**: User account created, pending record removed
6. **If rejected**: Pending record removed, admin notified

## 🚨 Security Features

### Role Validation
- Hierarchical role checking
- Approval authority validation
- Request expiration handling

### Audit Trail
- All approval actions logged
- User actions tracked
- Request history maintained

### Notifications (Planned)
- Email notifications to approvers
- Status updates to requesters
- Expiration warnings

## 📋 Usage Examples

### Check Role Hierarchy
```javascript
// In frontend or API calls
const canApprove = models.CanApprove("super_admin", "admin"); // true
const hasAccess = models.HasHigherOrEqualRole("admin", "miner"); // true
const needsApproval = models.RequiresSuperAdminApproval("miner"); // true
```

### Default Users for Testing
- **Super Admin**: `superadmin` / password (to be set)
- **Admin**: `admin` / password (to be set)  
- **Auditor**: `auditor` / password (to be set)
- **Viewer**: `viewer` / password (to be set)

## 🎯 Benefits

1. **Enhanced Security**: Two-level approval for sensitive operations
2. **Compliance**: Audit trail for all supplier onboarding
3. **Scalability**: Hierarchical role system supports growth
4. **Transparency**: Clear approval workflow and status tracking
5. **Flexibility**: Configurable approval rules and expiration

## 🔄 Future Enhancements

- **Multi-step Approvals**: Chain of approval for complex requests
- **Conditional Approval**: Business rules-based approval logic
- **Bulk Operations**: Mass approval/rejection capabilities
- **Integration**: External approval systems integration
- **Analytics**: Approval workflow performance metrics

---

This approval workflow ensures that supplier onboarding maintains proper oversight while providing clear processes for both admins and super admins.