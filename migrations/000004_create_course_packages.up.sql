CREATE TABLE course_packages (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL UNIQUE CHECK (BTRIM(name) <> ''),
    level VARCHAR(20) NOT NULL CHECK (level IN ('PEMULA', 'DASAR')),
    total_hours SMALLINT NOT NULL CHECK (total_hours IN (6, 8, 10, 12)),
    price BIGINT NOT NULL CHECK (price > 0),
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_course_packages_status ON course_packages (status);
