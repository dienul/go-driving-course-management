CREATE TABLE session_skill_assessments (
    id BIGSERIAL PRIMARY KEY,
    training_session_id BIGINT NOT NULL
        REFERENCES training_sessions(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    sub_material_id BIGINT NOT NULL
        REFERENCES sub_materials(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    skill_status VARCHAR(30) NOT NULL
        CHECK (skill_status IN ('NOT_STARTED', 'PRACTICED', 'MASTERED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_session_skill_assessments_session_sub_material
        UNIQUE (training_session_id, sub_material_id)
);

CREATE INDEX idx_session_skill_assessments_sub_material
    ON session_skill_assessments (sub_material_id, training_session_id);
