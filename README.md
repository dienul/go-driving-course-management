# Driving Course Management System

REST API final project built with Go, Gin, GORM, PostgreSQL, versioned SQL
migrations, and Swagger/OpenAPI.

The implemented phases provide application bootstrap, versioned PostgreSQL
schema and seed data, JWT authentication, and administrator-only trainer,
student, package, material, and sub-material management.

## Requirements

- Go 1.24 or newer
- PostgreSQL
- GNU Make (optional; commands can also be run directly)
- `migrate` CLI (only needed for `make migrate-create`)

## Local setup

1. Copy `.env.example` to `.env` and update its values.
2. Create the PostgreSQL database named in `DATABASE_URL`.
3. Download dependencies and start the server:

   ```bash
   go mod download
   go run ./cmd
   ```

The application loads `.env` for local development. In deployed environments,
provide the same variables through the platform configuration.

## Endpoints

- Health: `GET http://localhost:8080/health`
- Swagger UI: `GET http://localhost:8080/swagger/index.html`

Expected health response:

```json
{
  "success": true,
  "message": "service is running",
  "data": null
}
```

The server verifies its PostgreSQL connection during startup. If PostgreSQL is
unavailable or `DATABASE_URL` is missing, startup fails with a clear error.

## Commands

```bash
make run
make build
make test
make swagger
make migrate-create name=create_users
make migrate-up
make migrate-down
make migrate-version
```

Equivalent raw migration commands:

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down 1
go run ./cmd/migrate version
```

Migration SQL files belong in `migrations/` and must use matching `.up.sql` and
`.down.sql` files. Schema migration uses `golang-migrate`; GORM `AutoMigrate` is
not used.

## Swagger generation

Regenerate the OpenAPI files after changing endpoint annotations:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/main.go
```

## Phase 1 environment variables

| Variable | Purpose |
| --- | --- |
| `APP_ENV` | Application environment name |
| `APP_PORT` | HTTP server port |
| `DATABASE_URL` | PostgreSQL connection URL |

The example file also reserves the JWT, Basic Auth, and initial admin variables
required by later roadmap phases.

## Phase 2 database setup

Phase 2 defines exactly 15 PostgreSQL tables through 30 ordered SQL files in
`migrations/`. SQL migrations are the schema source of truth; the application
does not call GORM `AutoMigrate`.

Apply every pending migration:

```bash
go run ./cmd/migrate up
```

Roll back one migration or inspect the current version:

```bash
go run ./cmd/migrate down 1
go run ./cmd/migrate version
```

After migration, seed the initial admin, four course packages, five curriculum
materials, and their sub-materials:

```bash
go run ./cmd/seed
```

The seed runs in one database transaction and can be run repeatedly. Configure
`ADMIN_NAME`, `ADMIN_EMAIL`, and `ADMIN_PASSWORD` before running it. The
password is stored only as a bcrypt hash.

Initial package prices are explicit sample values:

| Package | Price |
| --- | ---: |
| Pemula 6 Jam | Rp900,000 |
| Pemula 8 Jam | Rp1,100,000 |
| Dasar 10 Jam | Rp1,400,000 |
| Dasar 12 Jam | Rp1,600,000 |

These seed prices and curriculum descriptions are development defaults and can
be revised before production without changing the migration schema.

## Phase 3 authentication

Set a random JWT secret of at least 32 bytes:

```dotenv
JWT_SECRET=replace-with-a-long-random-secret
JWT_EXPIRES_IN=24h
```

Public student registration:

```bash
curl -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dienul Haq",
    "email": "dienul@example.com",
    "password": "strong-password",
    "phone": "081234567890",
    "address": "Jakarta"
  }'
```

The backend always assigns `role=STUDENT` and `status=ACTIVE`. Public clients
cannot choose an account role.

All roles use the same login endpoint:

```bash
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "dienul@example.com",
    "password": "strong-password"
  }'
```

Use the returned token for protected endpoints:

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <token>"
```

Authentication behavior:

- Passwords are stored with bcrypt.
- JWTs use HS256, a required expiration, and issuer validation.
- The user role is read from PostgreSQL during login.
- Protected requests reload the user from PostgreSQL.
- An `INACTIVE` user cannot login or continue using an existing JWT.
- Role middleware returns `403 Forbidden` when the authenticated role is not
  permitted.

Basic Auth middleware is available for the internal endpoints introduced in
Phase 10. Its credentials come only from:

```dotenv
BASIC_AUTH_USERNAME=internal
BASIC_AUTH_PASSWORD=replace-with-a-strong-password
```

Regenerate Swagger after changing authentication annotations:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/main.go
```

## Phase 4 administrator master data

All 23 administrator endpoints require an ACTIVE ADMIN account and
`Authorization: Bearer <admin-token>`. Other roles receive `403 Forbidden`.

- Trainers: `POST /trainers`, `GET /trainers`, `GET /trainers/:id`,
  `PUT /trainers/:id`, and `PATCH /trainers/:id/status`.
- Students: `GET /students`, `GET /students/:id`, and
  `PATCH /students/:id/status`.
- Packages: `POST /packages`, `GET /packages`, `GET /packages/:id`,
  `PUT /packages/:id`, and `PATCH /packages/:id/status`.
- Materials: `POST /materials`, `GET /materials`, `GET /materials/:id`,
  `PUT /materials/:id`, and `PATCH /materials/:id/status`.
- Sub-materials: `POST /materials/:material_id/sub-materials`,
  `GET /materials/:material_id/sub-materials`, `GET /sub-materials/:id`,
  `PUT /sub-materials/:id`, and `PATCH /sub-materials/:id/status`.

All paths above are relative to `/api/v1/admin`. Example trainer creation:

```bash
curl -X POST http://localhost:8080/api/v1/admin/trainers \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Pandu Pratama","email":"pandu@example.com","password":"strong-password"}'
```

Trainer accounts and profiles are created transactionally with bcrypt-hashed
passwords. Package levels are `PEMULA` or `DASAR`; durations are 6, 8, 10,
or 12 hours; prices and curriculum sequences must be positive. Curriculum
lists are sequence-ordered. Resource status is `ACTIVE` or `INACTIVE`;
there are no hard-delete endpoints.

Set `TEST_DATABASE_URL` to a migrated PostgreSQL test database before running
`go test ./...` to exercise the administrator integration suite.

## Phase 5 enrollment, payment, and invoice flow

Students can browse only ACTIVE packages and create an enrollment using only a
package ID:

```bash
curl -X POST http://localhost:8080/api/v1/student/enrollments \
  -H "Authorization: Bearer <student-token>" \
  -H "Content-Type: application/json" \
  -d '{"package_id":1}'
```

The backend transaction copies the package name, price, and total hours into the
enrollment, creates it as `PENDING_PAYMENT`, and creates one `UNPAID` payment
whose amount comes from the snapshot. Client-supplied student IDs, prices, and
payment amounts are ignored.

Simulate payment using one supported method:

```bash
curl -X POST http://localhost:8080/api/v1/student/payments/1/pay \
  -H "Authorization: Bearer <student-token>" \
  -H "Content-Type: application/json" \
  -d '{"payment_method":"BANK_TRANSFER"}'
```

In one transaction this changes the payment to `PAID`, records `paid_at`,
generates an `INV-YYYYMMDD-XXXX` invoice, and changes the enrollment to
`ACTIVE` with `started_at`. Any failure rolls back all changes. Double
payment is rejected, and PostgreSQL enforces at most one ACTIVE enrollment per
student.

Student Phase 5 endpoints:

- `GET /api/v1/student/packages` and `GET /api/v1/student/packages/:id`
- `POST /api/v1/student/enrollments`, `GET /api/v1/student/enrollments`,
  and `GET /api/v1/student/enrollments/:id`
- `GET /api/v1/student/enrollments/:id/payment` and
  `POST /api/v1/student/payments/:id/pay`
- `GET /api/v1/student/invoices` and `GET /api/v1/student/invoices/:id`

Administrators can monitor all records through read-only
`/api/v1/admin/enrollments`, `/api/v1/admin/payments`, and
`/api/v1/admin/invoices` collection/detail endpoints. Students can access
only their own enrollment, payment, and invoice records.

## Phase 6 trainer availability and student schedules

Trainers create weekday availability using full-hour ranges inside 08:00-17:00:

```bash
curl -X POST http://localhost:8080/api/v1/trainer/availabilities \
  -H "Authorization: Bearer <trainer-token>" \
  -H "Content-Type: application/json" \
  -d '{"available_date":"2026-08-31","start_time":"08:00","end_time":"12:00"}'
```

Trainer-owned endpoints:

- `POST /api/v1/trainer/availabilities`
- `GET /api/v1/trainer/availabilities`
- `GET /api/v1/trainer/availabilities/:id`
- `PUT /api/v1/trainer/availabilities/:id`
- `PATCH /api/v1/trainer/availabilities/:id/cancel`

Availability starts as `PENDING`. Weekends, partial-hour values, ranges outside
08:00-17:00, invalid ordering, and overlapping non-cancelled ranges for the
same trainer are rejected. Adjacent ranges and overlapping schedules for
different trainers are allowed.

Administrator review endpoints:

- `GET /api/v1/admin/trainer-availabilities`
- `GET /api/v1/admin/trainer-availabilities/:id`
- `POST /api/v1/admin/trainer-availabilities/:id/publish`
- `PATCH /api/v1/admin/trainer-availabilities/:id/cancel`

Publishing stores `published_by` and `published_at`. Published ranges cannot
be edited by trainers, and ranges with scheduled or in-progress sessions
cannot be cancelled.

Students inspect published schedules with optional date filtering:

```bash
curl "http://localhost:8080/api/v1/student/schedules?date=2026-08-31" \
  -H "Authorization: Bearer <student-token>"
```

An 08:00-12:00 availability produces the two-hour slots 08:00-10:00,
09:00-11:00, and 10:00-12:00. Occupied slots, inactive trainers, and
unpublished or cancelled ranges are excluded.

## Phase 7 training session booking

Students book a published two-hour training slot using only its availability
identifier and full-hour start time:

```bash
curl -X POST http://localhost:8080/api/v1/student/training-sessions \
  -H "Authorization: Bearer <student-token>" \
  -H "Content-Type: application/json" \
  -d '{"trainer_availability_id":10,"start_time":"08:00"}'
```

Student-owned session endpoints:

- `POST /api/v1/student/training-sessions`
- `GET /api/v1/student/training-sessions`
- `GET /api/v1/student/training-sessions/:id`
- `PATCH /api/v1/student/training-sessions/:id/cancel`
- `POST /api/v1/student/training-sessions/:id/reschedule`

The backend derives the active enrollment, trainer, scheduled date, two-hour
end time, and session number. Only published weekday availability from active
trainers can be booked. The entire slot must fall within the published range
and operating hours of 08:00-17:00. Each enrollment can have only one scheduled
or in-progress session, and completed sessions cannot exceed the purchased
session allowance. Transactions lock the trainer to prevent concurrent
overlapping bookings.

Cancel a scheduled session with a reason:

```bash
curl -X PATCH http://localhost:8080/api/v1/student/training-sessions/10/cancel \
  -H "Authorization: Bearer <student-token>" \
  -H "Content-Type: application/json" \
  -d '{"cancellation_reason":"Schedule changed"}'
```

Reschedule a scheduled session by choosing another published slot:

```bash
curl -X POST http://localhost:8080/api/v1/student/training-sessions/10/reschedule \
  -H "Authorization: Bearer <student-token>" \
  -H "Content-Type: application/json" \
  -d '{"trainer_availability_id":12,"start_time":"10:00"}'
```

Cancellation records the actor, reason, and timestamp. Rescheduling marks the
original session `RESCHEDULED` and creates a linked `SCHEDULED` replacement
with the same session number in a single transaction. Cancelled or rescheduled
records do not consume purchased training sessions.

Administrator oversight endpoints:

- `GET /api/v1/admin/training-sessions`
- `GET /api/v1/admin/training-sessions/:id`
- `PATCH /api/v1/admin/training-sessions/:id/cancel`
- `POST /api/v1/admin/training-sessions/:id/reschedule`

## Phase 8 trainer-led training process

Trainer-owned training process endpoints:

- `GET /api/v1/trainer/training-sessions`
- `GET /api/v1/trainer/training-sessions/:id`
- `POST /api/v1/trainer/training-sessions/:id/start`
- `PUT /api/v1/trainer/training-sessions/:id/assessments`
- `GET /api/v1/trainer/training-sessions/:id/assessments`
- `PUT /api/v1/trainer/training-sessions/:id/evaluation`
- `GET /api/v1/trainer/training-sessions/:id/evaluation`
- `GET /api/v1/trainer/training-sessions/:id/student-progress`
- `POST /api/v1/trainer/training-sessions/:id/complete`

Only the assigned active trainer can view or modify a training session. Starting
a `SCHEDULED` session transitions it to `IN_PROGRESS` and records
`actual_started_at`:

```bash
curl -X POST http://localhost:8080/api/v1/trainer/training-sessions/10/start \
  -H "Authorization: Bearer <trainer-token>"
```

Record only the active sub-materials practiced during this session:

```bash
curl -X PUT http://localhost:8080/api/v1/trainer/training-sessions/10/assessments \
  -H "Authorization: Bearer <trainer-token>" \
  -H "Content-Type: application/json" \
  -d '{"assessments":[{"sub_material_id":1,"skill_status":"MASTERED"},{"sub_material_id":2,"skill_status":"PRACTICED"}]}'
```

Allowed assessment statuses are `NOT_STARTED`, `PRACTICED`, and `MASTERED`.
Each session/sub-material combination has one record; submitting it again
updates that same record while preserving assessments from other sessions.
Assessment batches are transactional, duplicate sub-material identifiers are
rejected, and a previously mastered skill can later decline.

Provide the required general evaluation:

```bash
curl -X PUT http://localhost:8080/api/v1/trainer/training-sessions/10/evaluation \
  -H "Authorization: Bearer <trainer-token>" \
  -H "Content-Type: application/json" \
  -d '{"predicate":"BAIK","notes":"Kontrol kendaraan sudah cukup baik.","recommendation":"Fokus latihan parkir."}'
```

Evaluation predicates are `KURANG`, `CUKUP`, `BAIK`, and `SANGAT_BAIK`. Notes
and recommendations must contain non-whitespace text. Completing a session
requires `IN_PROGRESS` status, at least one assessment, and a complete
evaluation:

```bash
curl -X POST http://localhost:8080/api/v1/trainer/training-sessions/10/complete \
  -H "Authorization: Bearer <trainer-token>"
```

Completion records `actual_completed_at` and permanently locks assessment and
evaluation edits while keeping both resources readable. The `student-progress`
endpoint returns previous completed sessions, assessments, and evaluations
across every enrollment for the assigned student.

## Phase 9 skills, trainer reviews, and certificates

Students can inspect global driving skills and completed-session history:

- `GET /api/v1/student/skills`
- `GET /api/v1/student/skills/history`

Current skill includes every active sub-material belonging to an active
material. The latest assessment from a completed session across every student
enrollment determines its current status; unassessed skills remain
`NOT_STARTED`. In-progress assessments do not affect the current score.

Statuses are scored `NOT_STARTED = 0`, `PRACTICED = 1`, and `MASTERED = 2`.
The rounded percentage is `obtained / (active sub-materials * 2) * 100`:

- `0-39`: `BEGINNER`
- `40-59`: `DEVELOPING`
- `60-79`: `CAPABLE`
- `80-100`: `PROFICIENT`

A skill may improve or decline in later sessions or enrollments; historical
assessments remain unchanged.

Student-owned trainer review endpoints:

- `POST /api/v1/student/training-sessions/:id/review`
- `PUT /api/v1/student/training-sessions/:id/review`
- `GET /api/v1/student/training-sessions/:id/review`

```bash
curl -X POST http://localhost:8080/api/v1/student/training-sessions/10/review \
  -H "Authorization: Bearer <student-token>" \
  -H "Content-Type: application/json" \
  -d '{"rating":5,"feedback":"Patient and clear instruction."}'
```

Only the student who owns a `COMPLETED` session may review its assigned
trainer. Ratings must be between 1 and 5; each session has at most one review,
which its student can subsequently update. Trainers and administrators have
read-only oversight:

- `GET /api/v1/trainer/reviews`
- `GET /api/v1/trainer/reviews/summary`
- `GET /api/v1/admin/trainer-reviews`
- `GET /api/v1/admin/trainers/:id/reviews`

Trainer review totals and average ratings are calculated from existing review
records rather than stored as manually maintained counters.

When a trainer completes the final purchased session, the session transition,
enrollment completion, and certificate issuance occur within one transaction.
Only completed two-hour sessions count toward the purchased allowance.
Certificate numbers follow `CERT-YYYYMMDD-XXXX` and are unique; each enrollment
receives exactly one certificate.

Certificate inspection endpoints:

- `GET /api/v1/student/certificates`
- `GET /api/v1/student/certificates/:id`
- `GET /api/v1/admin/certificates`
- `GET /api/v1/admin/certificates/:id`

Certificates preserve the global skill score and proficiency level at issuance.
Later training may change current skills, but existing certificate snapshots
remain unchanged.

## Phase 10 internal health and operational statistics

Internal monitoring endpoints require HTTP Basic Auth:

- `GET /api/v1/internal/health`
- `GET /api/v1/internal/stats`

Configure both credentials through environment variables:

```dotenv
BASIC_AUTH_USERNAME=internal-monitor
BASIC_AUTH_PASSWORD=replace-with-a-strong-secret
```

The protected health endpoint verifies the API and PostgreSQL connection:

```bash
curl --user "$BASIC_AUTH_USERNAME:$BASIC_AUTH_PASSWORD" \
  http://localhost:8080/api/v1/internal/health
```

Operational statistics are calculated directly from the current database:

```bash
curl --user "$BASIC_AUTH_USERNAME:$BASIC_AUTH_PASSWORD" \
  http://localhost:8080/api/v1/internal/stats
```

The response includes student, trainer, and administrator totals; total and
active enrollments; total, scheduled, in-progress, and completed training
sessions; paid payments; issued certificates; and submitted trainer reviews.

Missing, malformed, or incorrect Basic Auth credentials return `401` and an
HTTP Basic challenge. Bearer JWTs cannot access internal endpoints, and Basic
credentials cannot replace JWTs on student, trainer, or administrator routes.
If either environment credential is absent, internal endpoints fail closed with
`500`. Database failures return `503` without exposing connection details.
The existing public `GET /health` endpoint remains unauthenticated.
