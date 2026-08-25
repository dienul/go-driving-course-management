CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL CHECK (BTRIM(name) <> ''),
    email VARCHAR(255) NOT NULL UNIQUE CHECK (BTRIM(email) <> ''),
    password_hash VARCHAR(255) NOT NULL CHECK (BTRIM(password_hash) <> ''),
    role VARCHAR(20) NOT NULL CHECK (role IN ('ADMIN', 'TRAINER', 'STUDENT')),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_role_status ON users (role, status);
