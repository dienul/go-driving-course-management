CREATE TABLE sub_materials (
    id BIGSERIAL PRIMARY KEY,
    material_id BIGINT NOT NULL REFERENCES materials(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    name VARCHAR(200) NOT NULL CHECK (BTRIM(name) <> ''),
    description TEXT,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_sub_materials_material_sequence UNIQUE (material_id, sequence),
    CONSTRAINT uq_sub_materials_material_name UNIQUE (material_id, name)
);

CREATE INDEX idx_sub_materials_material_status
    ON sub_materials (material_id, status, sequence);
