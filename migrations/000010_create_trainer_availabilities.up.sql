CREATE TABLE trainer_availabilities (
    id BIGSERIAL PRIMARY KEY,
    trainer_id BIGINT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    available_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'PUBLISHED', 'CANCELLED')),
    published_by BIGINT REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_trainer_availability_weekday
        CHECK (EXTRACT(ISODOW FROM available_date) BETWEEN 1 AND 5),
    CONSTRAINT chk_trainer_availability_hours CHECK (
        start_time >= TIME '08:00:00'
        AND end_time <= TIME '17:00:00'
        AND start_time < end_time
    ),
    CONSTRAINT chk_trainer_availability_full_hour CHECK (
        EXTRACT(MINUTE FROM start_time) = 0
        AND EXTRACT(SECOND FROM start_time) = 0
        AND EXTRACT(MINUTE FROM end_time) = 0
        AND EXTRACT(SECOND FROM end_time) = 0
    ),
    CONSTRAINT chk_trainer_availability_publication CHECK (
        status <> 'PUBLISHED' OR (published_by IS NOT NULL AND published_at IS NOT NULL)
    )
);

CREATE INDEX idx_trainer_availabilities_trainer_date
    ON trainer_availabilities (trainer_id, available_date, start_time, end_time);

CREATE INDEX idx_trainer_availabilities_public_schedule
    ON trainer_availabilities (available_date, status, start_time)
    WHERE status = 'PUBLISHED';
