CREATE TABLE training_sessions (
    id BIGSERIAL PRIMARY KEY,
    enrollment_id BIGINT NOT NULL
        REFERENCES student_enrollments(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    trainer_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    trainer_availability_id BIGINT NOT NULL
        REFERENCES trainer_availabilities(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    session_number INTEGER NOT NULL CHECK (session_number > 0),
    scheduled_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'SCHEDULED'
        CHECK (status IN ('SCHEDULED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED', 'RESCHEDULED')),
    actual_started_at TIMESTAMPTZ,
    actual_completed_at TIMESTAMPTZ,
    rescheduled_from_id BIGINT UNIQUE
        REFERENCES training_sessions(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    cancelled_by BIGINT REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    cancellation_reason TEXT,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_training_session_weekday
        CHECK (EXTRACT(ISODOW FROM scheduled_date) BETWEEN 1 AND 5),
    CONSTRAINT chk_training_session_duration CHECK (
        start_time >= TIME '08:00:00'
        AND end_time <= TIME '17:00:00'
        AND end_time = start_time + INTERVAL '2 hours'
    ),
    CONSTRAINT chk_training_session_full_hour CHECK (
        EXTRACT(MINUTE FROM start_time) = 0
        AND EXTRACT(SECOND FROM start_time) = 0
        AND EXTRACT(MINUTE FROM end_time) = 0
        AND EXTRACT(SECOND FROM end_time) = 0
    ),
    CONSTRAINT chk_training_session_started CHECK (
        status NOT IN ('IN_PROGRESS', 'COMPLETED') OR actual_started_at IS NOT NULL
    ),
    CONSTRAINT chk_training_session_completed CHECK (
        status <> 'COMPLETED' OR actual_completed_at IS NOT NULL
    ),
    CONSTRAINT chk_training_session_cancelled CHECK (
        status <> 'CANCELLED'
        OR (
            cancelled_by IS NOT NULL
            AND NULLIF(BTRIM(cancellation_reason), '') IS NOT NULL
            AND cancelled_at IS NOT NULL
        )
    )
);

CREATE INDEX idx_training_sessions_enrollment_status
    ON training_sessions (enrollment_id, status, session_number);

CREATE INDEX idx_training_sessions_trainer_schedule
    ON training_sessions (trainer_id, scheduled_date, start_time, end_time, status);

CREATE INDEX idx_training_sessions_availability
    ON training_sessions (trainer_availability_id);
