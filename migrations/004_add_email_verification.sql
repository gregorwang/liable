-- Add email fields to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS email VARCHAR(255) UNIQUE,
ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE;

-- Create index on email
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Optional audit table for email verification logs
CREATE TABLE IF NOT EXISTS email_verification_logs (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(10) NOT NULL,
    purpose VARCHAR(20) NOT NULL CHECK (purpose IN ('login', 'register')),
    ip_address VARCHAR(45),
    status VARCHAR(20) NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'verified', 'expired', 'failed')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_verification_logs_email ON email_verification_logs(email);
CREATE INDEX IF NOT EXISTS idx_verification_logs_created_at ON email_verification_logs(created_at);

-- Migration: Add email verification support to users table
-- File: migrations/004_add_email_verification.sql
-- Description: Adds email and email_verified fields to users table for email verification code login

-- Add email column (unique, nullable for backward compatibility)
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS email VARCHAR(255) UNIQUE,
ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE;

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Add comment for documentation
COMMENT ON COLUMN users.email IS 'User email address, used for verification code login';
COMMENT ON COLUMN users.email_verified IS 'Whether the email has been verified via verification code';

-- Optional: Update existing users if you have email data elsewhere
-- UPDATE users SET email_verified = TRUE WHERE email IS NOT NULL;

