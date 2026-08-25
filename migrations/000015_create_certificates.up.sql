CREATE TABLE certificates (
    id BIGSERIAL PRIMARY KEY,
    enrollment_id BIGINT NOT NULL UNIQUE
        REFERENCES student_enrollments(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    certificate_number VARCHAR(100) NOT NULL UNIQUE CHECK (BTRIM(certificate_number) <> ''),
    skill_score SMALLINT NOT NULL CHECK (skill_score BETWEEN 0 AND 100),
    skill_level VARCHAR(30) NOT NULL
        CHECK (skill_level IN ('BEGINNER', 'DEVELOPING', 'CAPABLE', 'PROFICIENT')),
    issued_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
