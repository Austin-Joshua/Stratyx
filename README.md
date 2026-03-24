# STRATYX Full-Stack App

Production-ready full-stack application with real persistence and secure auth.

- Frontend: Next.js + TypeScript + Tailwind + Axios
- Backend: Go + Gin (clean layers)
- Database: MongoDB (official Go driver)

## Project Structure

### Backend

- `backend/cmd/server/main.go`
- `backend/internal/config`
- `backend/internal/domain/models`
- `backend/internal/repository`
- `backend/internal/service`
- `backend/internal/platform/http`
- `backend/internal/platform/security`

### Frontend

- `frontend/src/components`
- `frontend/src/services/api.ts`
- `frontend/src/context/authContext.tsx`
- `frontend/src/hooks`
- `frontend/src/layouts`
- `frontend/src/app`

## API Endpoints

### Auth

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/logout`
- `POST /api/auth/refresh`
- `GET /api/auth/me`
- `POST /api/auth/forgot-password`
- `POST /api/auth/reset-password`
- `POST /api/auth/verify-email`
- `GET /api/auth/sessions`
- `POST /api/auth/logout-all`
- `GET /api/auth/oauth/google` + callback exchange
- `GET /api/auth/oauth/github` + callback exchange
- `POST /api/auth/2fa/setup`
- `POST /api/auth/2fa/verify-setup`
- `POST /api/auth/2fa/disable`
- `POST /api/auth/2fa/complete-login`
- `PUT /api/auth/profile`
- `POST /api/auth/change-password`
- `GET /api/auth/connected-accounts`
- `DELETE /api/auth/connected-accounts/:provider`

### Items

- `POST /api/items`
- `GET /api/items`
- `GET /api/items/:id`
- `PUT /api/items/:id`
- `DELETE /api/items/:id`

`GET /api/items` supports:
- `page` (default `1`)
- `pageSize` (default `10`, max `100`)
- `search` (title/description)
- `sortBy` (`updatedAt`, `createdAt`, `title`)
- `sortOrder` (`asc`, `desc`)

### System

- `GET /health`

### Collaboration & Notifications

- `GET /api/items/:id/comments`
- `POST /api/items/:id/comments`
- `POST /api/comments/:id/report`
- `GET /api/activity`
- `GET /api/notifications`
- `PUT /api/notifications/:id/read`

### Files

- `POST /api/items/:id/attachments` (`multipart/form-data`, field: `file`)
- `POST /api/files/:id/access` (mint one-time signed access URL)
- `GET /api/files/download/:id?token=...` (one-time private file download)

### Admin

- `GET /api/admin/metrics`
- `GET /api/admin/users`
- `PUT /api/admin/users/:id/role`
- `PUT /api/admin/users/:id/active`
- `GET /api/admin/moderation/reports`
- `PUT /api/admin/moderation/reports/:id/review`
- `GET /api/admin/audit-logs`
- `GET /api/admin/email-queue-health`

## Database Collections

- `users` (unique index on `email`)
- `sessions`
- `items` (indexes on `owner_id`, `updated_at`)
- `comments`
- `notifications`
- `activity_logs`
- `password_resets`
- `email_verifications`
- `oauth_accounts`
- `auth_challenges`
- `files`
- `moderation_reports`
- `oauth_states`
- `email_jobs`
- `file_access_tokens`

## Security Features

- bcrypt password hashing
- JWT access + refresh token flow
- session persistence with logout revocation
- protected routes (frontend + backend)
- CORS allowlist from env
- auth rate limiting middleware
- input validation and owner-based data isolation
- structured JSON request logging with `X-Request-ID`
- account lockout after repeated failed logins
- multi-device session tracking + revoke all sessions
- TOTP 2FA setup/verify/disable + login challenge flow
- SMTP-based verification/reset email delivery (when configured)
- OAuth callback exchange for Google/GitHub
- local or S3-compatible attachment storage
- OAuth state persistence and callback CSRF verification
- one-time signed/private file download URLs
- background SMTP email queue with retry + dead-letter state
- TTL cleanup indexes for OAuth states, file access tokens, and email tokens/jobs

## TTL Cleanup Indexes

MongoDB TTL indexes automatically clean up expired or stale token/job documents:

- `oauth_states`
  - `expires_at` -> immediate expiry cleanup
  - `used_at` (partial) -> remove consumed states after 24h
- `file_access_tokens`
  - `expires_at` -> immediate expiry cleanup
  - `used_at` (partial) -> remove consumed tokens after 24h
- `password_resets`
  - `expires_at` -> immediate expiry cleanup
  - `used_at` (partial) -> remove used reset tokens after 24h
- `email_verifications`
  - `expires_at` -> immediate expiry cleanup
  - `used_at` (partial) -> remove used verification tokens after 24h
- `email_jobs`
  - `updated_at` (partial `status=SENT`) -> keep sent jobs 30 days
  - `updated_at` (partial `status=DEAD`) -> keep dead jobs 7 days

## New Operational Endpoints

- `GET /api/admin/email-queue-health` (admin/moderator)
  - returns:
    - `pending` count
    - `failed` count
    - `dead` count
    - `recentFailed` (latest 10 failed jobs with error metadata)

## Run Locally

### 1) Start MongoDB

```bash
docker compose up -d mongodb
```

### 2) Start backend

```bash
cd backend
cp .env.example .env
go mod tidy
go run ./cmd/server
```

### 3) Start frontend

```bash
cd frontend
cp .env.local.example .env.local
npm install
npm run dev
```

Open:
- Frontend: [http://localhost:3000](http://localhost:3000)
- Backend health: [http://localhost:8080/health](http://localhost:8080/health)

## Environment Variables

### Backend `.env`

See `backend/.env.example`:
- `PORT`
- `MONGO_URI`
- `MONGO_DATABASE`
- `JWT_SECRET`
- `ALLOWED_ORIGIN`
- `FRONTEND_URL`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GOOGLE_REDIRECT_URL`
- `GITHUB_CLIENT_ID`
- `GITHUB_CLIENT_SECRET`
- `GITHUB_REDIRECT_URL`
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`
- `SMTP_FROM_EMAIL`
- `UPLOAD_STORAGE`
- `UPLOAD_LOCAL_PATH`
- `S3_ENDPOINT`
- `S3_REGION`
- `S3_BUCKET`
- `S3_ACCESS_KEY_ID`
- `S3_SECRET_ACCESS_KEY`
- `S3_USE_PATH_STYLE`

### Frontend `.env.local`

See `frontend/.env.local.example`:
- `NEXT_PUBLIC_API_BASE_URL`

## Sample Test Flow

1. Register new user at `/register`
2. Login at `/login`
3. Open `/items`
4. Create, edit, and delete records
5. Add comments and see activity updates
6. Open `/notifications`, `/profile`, and `/settings`
7. Reload page and verify persistence from MongoDB

Sample seed payloads are available at `backend/sample-test-data.json`.

## Automated API Tests

Basic integration tests are in `backend/tests/api_integration_test.go`.

Run them against a live local API:

```bash
cd backend
RUN_API_TESTS=1 TEST_API_BASE_URL=http://localhost:8080 go test ./tests -run TestAPIAuthAndItemsCRUD
```

## Sample Test Accounts

- User: `demo@stratyx.test` / `Password123!`
- Admin: `admin@stratyx.test` / `AdminPass123!`
