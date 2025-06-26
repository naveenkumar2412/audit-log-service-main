-- Add status column to audit_logs table
-- This migration adds a configurable status column to track audit log processing status

-- Add status column with default value
ALTER TABLE audit_logs ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'pending';

-- Create index for status column for better query performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status);

-- Create composite index for tenant and status
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_status ON audit_logs(tenant_id, status);
