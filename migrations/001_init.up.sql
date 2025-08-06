-- Create enum types
CREATE TYPE user_role AS ENUM ('super_admin', 'admin', 'miner', 'recycler', 'certifier', 'manufacturer', 'auditor', 'viewer');
CREATE TYPE passport_status AS ENUM ('active', 'inactive', 'pending', 'verified');
CREATE TYPE action_type AS ENUM ('CREATE', 'UPDATE', 'DELETE', 'VIEW', 'EXPORT', 'VERIFY', 'RECYCLE', 'APPROVE', 'REJECT');
CREATE TYPE approval_status AS ENUM ('pending', 'approved', 'rejected', 'expired');
CREATE TYPE approval_type AS ENUM ('supplier_onboarding', 'user_role_change', 'system_configuration');

-- Users table with role-based access
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'viewer',
    company_name VARCHAR(255),
    contact_info JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE
);

-- Enhanced passport table with comprehensive fields
CREATE TABLE IF NOT EXISTS aluminium_passports (
    id SERIAL PRIMARY KEY,
    passport_id VARCHAR(100) UNIQUE NOT NULL,
    batch_id VARCHAR(100),
    
    -- Basic Information
    manufacturer VARCHAR(255) NOT NULL,
    origin VARCHAR(255) NOT NULL,
    bauxite_source VARCHAR(255),
    alloy_composition TEXT,
    
    -- Mining & Extraction
    mine_operator VARCHAR(255),
    date_of_extraction DATE,
    extraction_method VARCHAR(100),
    mine_location VARCHAR(255),
    
    -- Refining
    refinery_location VARCHAR(255),
    refiner_id VARCHAR(100),
    refining_date DATE,
    refining_method VARCHAR(100),
    
    -- Smelting & Manufacturing
    smelting_location VARCHAR(255),
    smelting_energy_source VARCHAR(255),
    process_type VARCHAR(100),
    manufactured_product VARCHAR(255),
    manufacturing_date DATE,
    
    -- Quantities & Measurements
    product_weight DECIMAL(10,3),
    energy_used DECIMAL(10,2), -- kWh
    water_used DECIMAL(10,2), -- Liters
    waste_generated DECIMAL(10,2), -- kg
    
    -- Environmental Impact
    carbon_emissions_per_kg DECIMAL(8,3),
    co2_footprint DECIMAL(10,2),
    manufacturing_emissions DECIMAL(10,2),
    
    -- Supply Chain & Logistics
    transport_mode VARCHAR(100),
    distance_travelled DECIMAL(10,2),
    logistics_partner_id VARCHAR(100),
    shipment_date DATE,
    
    -- Recycling Information
    recycled_content_percent DECIMAL(5,2) CHECK (recycled_content_percent >= 0 AND recycled_content_percent <= 100),
    recycling_date DATE,
    recycler_id VARCHAR(100),
    recycling_method VARCHAR(255),
    times_recycled INTEGER DEFAULT 0,
    last_recycling_date DATE,
    
    -- Certifications & Compliance
    certification_agency VARCHAR(255),
    certifier VARCHAR(255),
    compliance_standards TEXT,
    date_of_certification DATE,
    certification_expiry DATE,
    verifier_signature VARCHAR(500),
    
    -- ESG Scoring
    esg_score DECIMAL(5,2) CHECK (esg_score >= 0 AND esg_score <= 100),
    environmental_score DECIMAL(5,2) CHECK (environmental_score >= 0 AND environmental_score <= 100),
    social_score DECIMAL(5,2) CHECK (social_score >= 0 AND social_score <= 100),
    governance_score DECIMAL(5,2) CHECK (governance_score >= 0 AND governance_score <= 100),
    esg_last_updated TIMESTAMP WITH TIME ZONE,
    
    -- Digital Assets
    ipfs_hash VARCHAR(100),
    qr_code_data TEXT,
    digital_signature TEXT,
    
    -- Blockchain Integration
    blockchain_tx_hash VARCHAR(66),
    contract_address VARCHAR(42),
    block_number BIGINT,
    
    -- Status & Metadata
    status passport_status DEFAULT 'active',
    is_verified BOOLEAN DEFAULT false,
    verification_date TIMESTAMP WITH TIME ZONE,
    
    -- Additional metadata as JSON
    metadata JSONB,
    supply_chain_steps JSONB, -- Array of supply chain steps with timestamps
    certifications JSONB, -- Array of certifications
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id)
);

-- ESG Metrics detailed table
CREATE TABLE IF NOT EXISTS esg_metrics (
    id SERIAL PRIMARY KEY,
    passport_id VARCHAR(100) REFERENCES aluminium_passports(passport_id) ON DELETE CASCADE,
    
    -- Environmental metrics
    carbon_footprint DECIMAL(10,3),
    energy_efficiency_score DECIMAL(5,2),
    water_usage_score DECIMAL(5,2),
    waste_management_score DECIMAL(5,2),
    renewable_energy_percent DECIMAL(5,2),
    
    -- Social metrics
    labor_practices_score DECIMAL(5,2),
    community_impact_score DECIMAL(5,2),
    health_safety_score DECIMAL(5,2),
    human_rights_score DECIMAL(5,2),
    
    -- Governance metrics
    transparency_score DECIMAL(5,2),
    ethics_score DECIMAL(5,2),
    compliance_score DECIMAL(5,2),
    stakeholder_engagement_score DECIMAL(5,2),
    
    -- Composite scores
    overall_esg_score DECIMAL(5,2),
    
    -- Metadata
    assessment_date DATE,
    assessor_id INTEGER REFERENCES users(id),
    assessment_methodology VARCHAR(255),
    notes TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Supply chain tracking table
CREATE TABLE IF NOT EXISTS supply_chain_steps (
    id SERIAL PRIMARY KEY,
    passport_id VARCHAR(100) REFERENCES aluminium_passports(passport_id) ON DELETE CASCADE,
    step_name VARCHAR(255) NOT NULL,
    step_description TEXT,
    location VARCHAR(255),
    responsible_party VARCHAR(255),
    step_date DATE,
    verification_status BOOLEAN DEFAULT false,
    verifier_id INTEGER REFERENCES users(id),
    step_data JSONB, -- Additional step-specific data
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    user_role user_role,
    action action_type NOT NULL,
    resource_type VARCHAR(50), -- 'passport', 'user', 'esg_metrics', etc.
    resource_id VARCHAR(100),
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    session_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Certifications table
CREATE TABLE IF NOT EXISTS certifications (
    id SERIAL PRIMARY KEY,
    passport_id VARCHAR(100) REFERENCES aluminium_passports(passport_id) ON DELETE CASCADE,
    certification_name VARCHAR(255) NOT NULL,
    certification_body VARCHAR(255),
    certificate_number VARCHAR(100),
    issue_date DATE,
    expiry_date DATE,
    status VARCHAR(50) DEFAULT 'active',
    certificate_url VARCHAR(500),
    verification_hash VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Batch operations tracking
CREATE TABLE IF NOT EXISTS batch_operations (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) UNIQUE NOT NULL,
    operation_type VARCHAR(50), -- 'upload', 'update', 'export'
    total_records INTEGER,
    successful_records INTEGER,
    failed_records INTEGER,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    error_log JSONB,
    ipfs_hash VARCHAR(100),
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- ZK Proofs table (for zero-knowledge proof storage)
CREATE TABLE IF NOT EXISTS zk_proofs (
    id SERIAL PRIMARY KEY,
    passport_id VARCHAR(100) REFERENCES aluminium_passports(passport_id) ON DELETE CASCADE,
    proof_type VARCHAR(100),
    proof_data TEXT, -- The actual proof
    public_inputs JSONB,
    verification_key TEXT,
    is_verified BOOLEAN DEFAULT false,
    verifier_address VARCHAR(42),
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    verified_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better performance
CREATE INDEX idx_passports_passport_id ON aluminium_passports(passport_id);
CREATE INDEX idx_passports_batch_id ON aluminium_passports(batch_id);
CREATE INDEX idx_passports_manufacturer ON aluminium_passports(manufacturer);
CREATE INDEX idx_passports_status ON aluminium_passports(status);
CREATE INDEX idx_passports_created_at ON aluminium_passports(created_at);
CREATE INDEX idx_passports_esg_score ON aluminium_passports(esg_score);

CREATE INDEX idx_users_wallet_address ON users(wallet_address);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_resource_id ON audit_logs(resource_id);

CREATE INDEX idx_supply_chain_passport_id ON supply_chain_steps(passport_id);
CREATE INDEX idx_supply_chain_step_date ON supply_chain_steps(step_date);

CREATE INDEX idx_esg_metrics_passport_id ON esg_metrics(passport_id);
CREATE INDEX idx_esg_metrics_overall_score ON esg_metrics(overall_esg_score);

CREATE INDEX idx_certifications_passport_id ON certifications(passport_id);
CREATE INDEX idx_certifications_expiry_date ON certifications(expiry_date);

CREATE INDEX idx_batch_operations_batch_id ON batch_operations(batch_id);
CREATE INDEX idx_batch_operations_status ON batch_operations(status);

CREATE INDEX idx_zk_proofs_passport_id ON zk_proofs(passport_id);
CREATE INDEX idx_zk_proofs_is_verified ON zk_proofs(is_verified);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_passports_updated_at BEFORE UPDATE ON aluminium_passports FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_esg_metrics_updated_at BEFORE UPDATE ON esg_metrics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Approval requests table
CREATE TABLE IF NOT EXISTS approval_requests (
    id SERIAL PRIMARY KEY,
    request_type approval_type NOT NULL,
    requested_by INTEGER REFERENCES users(id) ON DELETE CASCADE,
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

-- Pending users table (for users awaiting approval)
CREATE TABLE IF NOT EXISTS pending_users (
    id SERIAL PRIMARY KEY,
    approval_request_id INTEGER REFERENCES approval_requests(id) ON DELETE CASCADE,
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

-- Create indexes for approval system
CREATE INDEX idx_approval_requests_status ON approval_requests(status);
CREATE INDEX idx_approval_requests_type ON approval_requests(request_type);
CREATE INDEX idx_approval_requests_approver_role ON approval_requests(approver_role);
CREATE INDEX idx_approval_requests_requested_by ON approval_requests(requested_by);
CREATE INDEX idx_approval_requests_expires_at ON approval_requests(expires_at);
CREATE INDEX idx_pending_users_approval_request_id ON pending_users(approval_request_id);
CREATE INDEX idx_pending_users_wallet_address ON pending_users(wallet_address);

-- Apply updated_at trigger to approval_requests
CREATE TRIGGER update_approval_requests_updated_at BEFORE UPDATE ON approval_requests FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default super admin and admin users
INSERT INTO users (wallet_address, username, email, password_hash, role, company_name, is_active) 
VALUES 
('0x0000000000000000000000000000000000000000', 'superadmin', 'superadmin@aluminiumpassport.com', '$2a$10$example_hash', 'super_admin', 'Aluminium Passport System', true),
('0x1111111111111111111111111111111111111111', 'admin', 'admin@aluminiumpassport.com', '$2a$10$example_hash', 'admin', 'Aluminium Passport System', true),
('0x2222222222222222222222222222222222222222', 'auditor', 'auditor@example.com', '$2a$10$example_hash', 'auditor', 'Audit Services Ltd', true),
('0x3333333333333333333333333333333333333333', 'viewer', 'viewer@example.com', '$2a$10$example_hash', 'viewer', 'Public Viewer', true)
ON CONFLICT (wallet_address) DO NOTHING;