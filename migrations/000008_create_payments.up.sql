CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    enrollment_id BIGINT NOT NULL UNIQUE
        REFERENCES student_enrollments(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    payment_code VARCHAR(100) NOT NULL UNIQUE CHECK (BTRIM(payment_code) <> ''),
    amount BIGINT NOT NULL CHECK (amount > 0),
    payment_method VARCHAR(30)
        CHECK (payment_method IS NULL OR payment_method IN ('BANK_TRANSFER', 'CASH')),
    status VARCHAR(20) NOT NULL DEFAULT 'UNPAID' CHECK (status IN ('UNPAID', 'PAID')),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_payment_state CHECK (
        (status = 'UNPAID' AND paid_at IS NULL)
        OR
        (status = 'PAID' AND payment_method IS NOT NULL AND paid_at IS NOT NULL)
    )
);

CREATE INDEX idx_payments_status ON payments (status);
