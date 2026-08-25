CREATE TABLE student_enrollments (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    package_id BIGINT NOT NULL REFERENCES course_packages(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    package_name VARCHAR(150) NOT NULL CHECK (BTRIM(package_name) <> ''),
    package_price BIGINT NOT NULL CHECK (package_price > 0),
    total_hours SMALLINT NOT NULL CHECK (total_hours IN (6, 8, 10, 12)),
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING_PAYMENT'
        CHECK (status IN ('PENDING_PAYMENT', 'ACTIVE', 'COMPLETED', 'CANCELLED')),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_student_enrollment_timestamps CHECK (
        (status <> 'ACTIVE' OR started_at IS NOT NULL)
        AND (status <> 'COMPLETED' OR (started_at IS NOT NULL AND completed_at IS NOT NULL))
    )
);

CREATE UNIQUE INDEX uq_student_active_enrollment
    ON student_enrollments(student_id)
    WHERE status = 'ACTIVE';

CREATE INDEX idx_student_enrollments_student_status
    ON student_enrollments (student_id, status);

CREATE INDEX idx_student_enrollments_package_id
    ON student_enrollments (package_id);
