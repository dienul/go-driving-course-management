# Driving Course Management System

## Project overview

A production-ready REST API for managing a driving school from student
registration through course enrollment, payment, trainer scheduling, practical
sessions, skill assessments, trainer reviews, and certificate issuance.

Production service:

- API: https://go-driving-course-management-production.up.railway.app
- Health: https://go-driving-course-management-production.up.railway.app/health
- Swagger UI:
  https://go-driving-course-management-production.up.railway.app/swagger/index.html

## Business flow

1. An administrator is created by the idempotent database seeder.
2. Administrators manage trainers, students, packages, and the driving
   curriculum.
3. A student registers and signs in with an individual JWT.
4. The student chooses an active course package and creates an enrollment.
5. Enrollment creates a package snapshot and one unpaid payment atomically.
6. Payment issues an invoice and activates the enrollment in one transaction.
7. A trainer proposes weekday availability and an administrator publishes it.
8. The student books an available two-hour training slot.
9. The assigned trainer starts the session, records skill assessments, and
   writes a general evaluation.
10. Completing the final purchased session completes the enrollment and issues
    a certificate atomically.
11. The student can monitor global skill progression and review the trainer.
12. Administrators supervise operations while protected internal endpoints
    expose service health and operational statistics.

## Roles

| Role | Authentication | Responsibilities |
| --- | --- | --- |
| `STUDENT` | JWT Bearer | Register, enroll, pay, inspect invoices, book sessions, monitor skills, review trainers, and access personal certificates. |
| `TRAINER` | JWT Bearer | Maintain availability, conduct assigned training sessions, assess skills, evaluate students, and inspect personal reviews. |
| `ADMIN` | JWT Bearer | Manage users, packages, curriculum, availability publication, enrollment/payment oversight, sessions, reviews, and certificates. |
| Internal monitoring | HTTP Basic Auth | Inspect protected service health and aggregate operational statistics. |

All role-protected requests reload the active account from PostgreSQL. A user
cannot escalate privileges by submitting a different role in an API payload.

## Technology stack

| Component | Technology |
| --- | --- |
| Language | Go 1.24+ with Go Modules |
| HTTP framework | Gin |
| Database | PostgreSQL |
| ORM | GORM |
| Database migration | `golang-migrate` with versioned PostgreSQL SQL files |
| Authentication | JWT Bearer and HTTP Basic Auth |
| Password hashing | bcrypt |
| API documentation | Swagger / OpenAPI 2.0 |
| Automated testing | Go unit tests and PostgreSQL integration tests |
| API regression collection | Postman Collection v2.1 |
| Containerization | Multi-stage Docker build |
| Production hosting | Railway |

## Architecture

```text
HTTP request
  -> Gin router
  -> JWT / Basic Auth / role middleware
  -> handler and DTO validation
  -> service and business rules
  -> repository and database transactions
  -> PostgreSQL
```

Package layout:

```text
cmd/
  main.go             API server entry point
  migrate/main.go     Versioned migration command
  seed/main.go        Admin and master-data seeder
config/               Environment and PostgreSQL configuration
docs/                 Generated Swagger/OpenAPI files
dto/                  API request and response contracts
handlers/             Gin HTTP handlers and Swagger annotations
middleware/           JWT, HTTP Basic Auth, and role authorization
migrations/           15 ordered PostgreSQL UP/DOWN migration pairs
models/               Persistent entities, statuses, and value objects
postman/              Complete Postman regression collection
repositories/         Persistence, locking, and transactional operations
routes/               Public and role-specific endpoint registration
seeds/                Idempotent admin, package, and curriculum seed data
services/             Application use cases and business validations
tests/                API, migration, Postman, and PostgreSQL integration tests
utils/                Shared HTTP response helpers
Dockerfile            Production server, migration, and seeder image
```

## Database

The application uses exactly 15 relational PostgreSQL tables defined by 15
ordered UP migrations and 15 corresponding DOWN migrations:

1. `users`
2. `student_profiles`
3. `trainer_profiles`
4. `course_packages`
5. `materials`
6. `sub_materials`
7. `student_enrollments`
8. `payments`
9. `invoices`
10. `trainer_availabilities`
11. `training_sessions`
12. `session_skill_assessments`
13. `session_evaluations`
14. `trainer_reviews`
15. `certificates`

PostgreSQL constraints enforce valid roles, unique emails, package durations,
positive prices, one active enrollment per student, one payment per
enrollment, one invoice per payment, weekday operating hours, two-hour
sessions, lifecycle timestamps, one assessment per session/sub-material, one
review per session, and one certificate per enrollment.

### Entity relationship diagram

```mermaid
erDiagram
    USERS {
        bigint id PK
        varchar email UK
        varchar role
        varchar status
    }
    STUDENT_PROFILES {
        bigint id PK
        bigint user_id FK
    }
    TRAINER_PROFILES {
        bigint id PK
        bigint user_id FK
    }
    COURSE_PACKAGES {
        bigint id PK
        int total_hours
        bigint price
    }
    MATERIALS {
        bigint id PK
        int sequence
    }
    SUB_MATERIALS {
        bigint id PK
        bigint material_id FK
    }
    STUDENT_ENROLLMENTS {
        bigint id PK
        bigint student_id FK
        bigint package_id FK
        varchar status
    }
    PAYMENTS {
        bigint id PK
        bigint enrollment_id FK
        varchar status
    }
    INVOICES {
        bigint id PK
        bigint payment_id FK
    }
    TRAINER_AVAILABILITIES {
        bigint id PK
        bigint trainer_id FK
        bigint published_by FK
    }
    TRAINING_SESSIONS {
        bigint id PK
        bigint enrollment_id FK
        bigint trainer_id FK
        bigint trainer_availability_id FK
        varchar status
    }
    SESSION_SKILL_ASSESSMENTS {
        bigint id PK
        bigint training_session_id FK
        bigint sub_material_id FK
    }
    SESSION_EVALUATIONS {
        bigint id PK
        bigint training_session_id FK
    }
    TRAINER_REVIEWS {
        bigint id PK
        bigint training_session_id FK
    }
    CERTIFICATES {
        bigint id PK
        bigint enrollment_id FK
    }
    USERS ||--o| STUDENT_PROFILES : has
    USERS ||--o| TRAINER_PROFILES : has
    USERS ||--o{ STUDENT_ENROLLMENTS : purchases
    USERS ||--o{ TRAINER_AVAILABILITIES : offers
    USERS ||--o{ TRAINING_SESSIONS : trains
    COURSE_PACKAGES ||--o{ STUDENT_ENROLLMENTS : selected_for
    MATERIALS ||--o{ SUB_MATERIALS : contains
    STUDENT_ENROLLMENTS ||--o| PAYMENTS : creates
    PAYMENTS ||--o| INVOICES : issues
    STUDENT_ENROLLMENTS ||--o{ TRAINING_SESSIONS : includes
    TRAINER_AVAILABILITIES ||--o{ TRAINING_SESSIONS : schedules
    TRAINING_SESSIONS ||--o{ SESSION_SKILL_ASSESSMENTS : records
    SUB_MATERIALS ||--o{ SESSION_SKILL_ASSESSMENTS : assessed_by
    TRAINING_SESSIONS ||--o| SESSION_EVALUATIONS : receives
    TRAINING_SESSIONS ||--o| TRAINER_REVIEWS : receives
    STUDENT_ENROLLMENTS ||--o| CERTIFICATES : earns
```

## Prerequisites

- Go 1.24 or newer
- PostgreSQL
- GNU Make (optional; commands can also be run directly)
- `migrate` CLI (only needed for `make migrate-create`)

## Local setup

1. Clone the repository and enter its directory.
2. Copy `.env.example` to `.env` and configure PostgreSQL, JWT, Basic Auth,
   and the initial administrator.
3. Create the PostgreSQL database named in `DATABASE_URL`.
4. Download dependencies, apply all migrations, seed the database, and start
   the API:

   ```bash
   go mod download
   go run ./cmd/migrate up
   go run ./cmd/seed
   go run ./cmd
   ```

The application loads `.env` for local development. In deployed environments,
provide the same variables through the platform configuration.

## REST API overview

- Health: `GET http://localhost:8080/health`
- Swagger UI: `GET http://localhost:8080/swagger/index.html`

| Group | Prefix | Authentication | Operations |
| --- | --- | --- | ---: |
| Public health and account registration/login | `/health`, `/api/users` | None | 3 |
| Current authenticated account | `/api/v1/auth` | JWT Bearer | 1 |
| Student | `/api/v1/student` | Student JWT | 22 |
| Trainer | `/api/v1/trainer` | Trainer JWT | 16 |
| Administrator | `/api/v1/admin` | Administrator JWT | 41 |
| Internal monitoring | `/api/v1/internal` | HTTP Basic Auth | 2 |

The complete API therefore exposes 85 documented operations; Swagger UI itself
is served separately at `/swagger/index.html`.

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

## Swagger and OpenAPI

Interactive documentation is available at:

- Local: http://localhost:8080/swagger/index.html
- Production:
  https://go-driving-course-management-production.up.railway.app/swagger/index.html
- Generated OpenAPI JSON: `/swagger/doc.json`.

In Swagger UI, select **Authorize** and enter `Bearer <JWT token>` for
`BearerAuth`, or enter the configured `BASIC_AUTH_USERNAME` and
`BASIC_AUTH_PASSWORD` for `BasicAuth`.

Regenerate the OpenAPI files after changing endpoint annotations:

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/main.go
```

## Environment variables

| Variable | Required | Example | Purpose |
| --- | --- | --- | --- |
| `APP_ENV` | No | `production` | Runtime environment label; defaults to `development`. |
| `APP_PORT` | No | `8080` | Local HTTP port; Docker derives it from Railway `PORT`. |
| `CORS_ALLOWED_ORIGINS` | Yes for a deployed frontend | `https://drive-academy.up.railway.app,http://localhost:5173` | Comma-separated browser origins allowed to call the API; defaults to `http://localhost:5173`. |
| `PORT` | Railway | `8080` | Platform-assigned port consumed automatically at startup. |
| `DATABASE_URL` | Yes | `postgres://user:password@host:5432/database?sslmode=require` | PostgreSQL connection URL used by migration, seeding, and the API. |
| `JWT_SECRET` | Yes | `replace-with-at-least-32-random-characters` | JWT HS256 signing secret containing at least 32 bytes. |
| `JWT_EXPIRES_IN` | No | `24h` | JWT token lifetime; defaults to 24 hours. |
| `BASIC_AUTH_USERNAME` | Internal endpoints | `internal` | Protected monitoring username. |
| `BASIC_AUTH_PASSWORD` | Internal endpoints | `a-strong-monitoring-password` | Protected monitoring password. |
| `ADMIN_NAME` | Seeder | `System Administrator` | Initial administrator name. |
| `ADMIN_EMAIL` | Seeder | `admin@example.com` | Initial administrator login email. |
| `ADMIN_PASSWORD` | Seeder | `a-strong-admin-password` | Initial administrator password, stored as a bcrypt hash. |
| `TEST_DATABASE_URL` | Integration tests | `postgres://user:password@localhost:5432/driving_course_test?sslmode=disable` | Separate migrated PostgreSQL test database. |

Never commit `.env` files, production database credentials, administrator
passwords, JWT secrets, or Basic Auth credentials. The application intentionally
fails to start when PostgreSQL cannot be reached or the JWT secret is invalid.
Docker deployment also requires all three administrator variables because the
seeder runs before the API.

## Migration and seed

### PostgreSQL schema and initial data

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

## Authentication

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

## Postman collection

Import the version 2.1 collection:

```text
postman/Driving-Course-Management.postman_collection.json
```

The collection contains 124 requests in eight folders:

1. Health and Swagger.
2. Public authentication and role-specific login.
3. Student operations.
4. Trainer operations.
5. Administrator operations.
6. Internal HTTP Basic Auth operations.
7. Role and security rejection scenarios.
8. End-to-end registration, enrollment, payment, three practical sessions,
   skill assessment, trainer review, certificate issuance, and internal
   statistics.

Configure these collection variables before running requests:

| Variable | Local value | Production value |
| --- | --- | --- |
| `base_url` | `http://localhost:8080` | `https://go-driving-course-management-production.up.railway.app` |
| `admin_email` | Your configured `ADMIN_EMAIL` | Your production `ADMIN_EMAIL` |
| `admin_password` | Your configured `ADMIN_PASSWORD` | Your production `ADMIN_PASSWORD` |
| `basic_username` | Your configured `BASIC_AUTH_USERNAME` | Your production `BASIC_AUTH_USERNAME` |
| `basic_password` | Your configured `BASIC_AUTH_PASSWORD` | Your production `BASIC_AUTH_PASSWORD` |

Student/trainer credentials, the next available weekday, JWTs, and resource
identifiers are managed through collection variables and response test scripts.
Run the **07 End-to-End Business Flow** folder for the complete three-session
workflow. The environment must already have migrated tables, seeded
administrator data, the four course packages, and active curriculum records.

## Testing

Run all available unit tests and database-independent contract checks:

```bash
go test ./...
go vet ./...
go build ./...
```

PostgreSQL integration tests are enabled by setting `TEST_DATABASE_URL` to a
dedicated migrated test database:

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/driving_course_test?sslmode=disable'
export TEST_DATABASE_URL="$DATABASE_URL"
go run ./cmd/migrate up
go test -count=1 ./...
go run ./cmd/migrate down 15
```

PowerShell equivalent:

```powershell
$env:DATABASE_URL = 'postgres://postgres:postgres@localhost:5432/driving_course_test?sslmode=disable'
$env:TEST_DATABASE_URL = $env:DATABASE_URL
go run ./cmd/migrate up
go test -count=1 ./...
go run ./cmd/migrate down 15
```

The suite covers API registration/login, student/trainer/administrator role
isolation, Basic Auth, 15 ordered migrations, PostgreSQL business constraints,
transaction rollback, concurrent trainer booking conflicts, trainer
availability, complete training session lifecycle, skill score thresholds,
trainer reviews, certificate issuance, Swagger UI/OpenAPI coverage, all 85
documented operations, Postman request contracts, the executable three-session
Postman business flow, and idempotent bcrypt-protected admin seeding.

Never run integration tests or migration rollbacks against a production
database.

## Railway deployment

The root `Dockerfile` uses a multi-stage build. Its builder compiles three
independent binaries:

```text
/app/server     API server from ./cmd
/app/migrate    PostgreSQL migration command from ./cmd/migrate
/app/seed       Administrator and master-data seeder from ./cmd/seed
```

The Alpine runtime performs startup in this strict order:

```sh
export APP_PORT=${PORT:-8080}
/app/migrate up
/app/seed
exec /app/server
```

This ensures the production schema is migrated and the administrator, course
packages, and curriculum are available before the health check succeeds. The
seed is transactional and idempotent.

To deploy:

1. Push the repository, including `Dockerfile` and `migrations/`, to GitHub.
2. Create a Railway project and attach a PostgreSQL service.
3. Create an application service from the GitHub repository. Railway detects
   the root `Dockerfile` automatically.
4. Configure `DATABASE_URL` using the Railway PostgreSQL connection reference.
5. Configure `JWT_SECRET`, `JWT_EXPIRES_IN`, `BASIC_AUTH_USERNAME`,
   `BASIC_AUTH_PASSWORD`, `ADMIN_NAME`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`, and
   `CORS_ALLOWED_ORIGINS=https://drive-academy.up.railway.app,http://localhost:5173`.
6. Set the healthcheck path to `/health` and configure public networking.
7. Leave **Custom Start Command** and **Pre-deploy Command** empty; the
   Dockerfile already handles migration, seed, the platform port, and startup.
8. Deploy or redeploy the application.
9. Confirm the deployment logs contain `migration completed` or
   `no migration changes`, followed by `seed completed`.
10. Verify the production service:

```bash
curl https://go-driving-course-management-production.up.railway.app/health
curl https://go-driving-course-management-production.up.railway.app/swagger/doc.json
```

Then open:

```text
https://go-driving-course-management-production.up.railway.app/swagger/index.html
```

The public `/health` endpoint pings PostgreSQL, so a successful response
verifies both the deployed API and its production database connection. Access
protected internal operational statistics with:

```bash
curl --user "$BASIC_AUTH_USERNAME:$BASIC_AUTH_PASSWORD" \
  https://go-driving-course-management-production.up.railway.app/api/v1/internal/stats
```

If deployment exits immediately, inspect Railway **Deploy Logs**. Common causes
are a missing `DATABASE_URL`, an unreachable PostgreSQL service, a `JWT_SECRET`
shorter than 32 bytes, or missing `ADMIN_NAME`/`ADMIN_EMAIL`/`ADMIN_PASSWORD`.
Avoid custom startup commands containing `go run` because the final Alpine
image intentionally contains compiled binaries without the Go toolchain.
