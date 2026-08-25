CREATE TABLE session_evaluations (
    id BIGSERIAL PRIMARY KEY,
    training_session_id BIGINT NOT NULL UNIQUE
        REFERENCES training_sessions(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    predicate VARCHAR(30) NOT NULL
        CHECK (predicate IN ('KURANG', 'CUKUP', 'BAIK', 'SANGAT_BAIK')),
    notes TEXT NOT NULL CHECK (BTRIM(notes) <> ''),
    recommendation TEXT NOT NULL CHECK (BTRIM(recommendation) <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
