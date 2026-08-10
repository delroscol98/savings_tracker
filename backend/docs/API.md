# Savings Tracker API Reference

Base URL: `http://localhost:8080`

The API is JSON. All timestamps are RFC 3339 (`"2026-06-01T00:00:00Z"`). Money
fields (`target`, `amount`) are integer amounts in your local currency units.

## Authentication

Protected routes require a JWT in the `Authorization` header:

```
Authorization: Bearer <token>
```

Tokens are issued by `POST /api/login` and signed with HS256. The `expires_in`
field (seconds) controls lifetime and defaults to `3600` (1 hour).

## Common response envelopes

Plain error:

```json
{ "error": "message" }
```

Validation error (field keys map to arrays of messages):

```json
{
  "error": "Invalid parameters to create new user",
  "fields": {
    "email": ["Invalid email"],
    "password": ["Password must be at least 8 characters"]
  }
}
```

## Quick start

```sh
# Register
curl -X POST localhost:8080/api/users \
  -d '{"email":"ada@example.com","password":"correct-horse-battery","full_name":"Ada Lovelace"}'

# Login
curl -X POST localhost:8080/api/login \
  -d '{"email":"ada@example.com","password":"correct-horse-battery"}'

# Use the returned token for protected routes
curl localhost:8080/api/goals -H "Authorization: Bearer <token>"
```

---

## Health

### `GET /health`

No authentication. Pings the database.

Responses:
- `200` — `1`
- `503` — `{"error": "Error pinging database"}`

---

## Users

### `POST /api/users`

Register a new user.

Request body:

```json
{
  "email": "ada@example.com",
  "password": "correct-horse-battery",
  "full_name": "Ada Lovelace"
}
```

Validation: `email` must parse as an address; `password` must be 8–128
characters and not in the common-password list; `full_name` must not be empty.

Responses:
- `201` — created user (no password returned):

  ```json
  {
    "id": "uuid",
    "created_at": "2026-08-10T12:00:00Z",
    "updated_at": "2026-08-10T12:00:00Z",
    "email": "ada@example.com"
  }
  ```

- `400` — invalid body or validation error
- `409` — `{"error": "Email already exists"}`
- `500` — internal error

### `POST /api/login`

Authenticate and receive a JWT.

Request body:

```json
{
  "email": "ada@example.com",
  "password": "correct-horse-battery",
  "expires_in": 3600
}
```

`expires_in` is optional (seconds); defaults to `3600` when omitted/`0`, and
cannot be negative.

Responses:
- `200` — user plus JWT token:

  ```json
  {
    "id": "uuid",
    "created_at": "2026-08-10T12:00:00Z",
    "updated_at": "2026-08-10T12:00:00Z",
    "email": "ada@example.com",
    "token": "<jwt>"
  }
  ```

- `400` — invalid body or validation error
- `403` — `{"error": "Incorrect email or password"}` (failures are also
  recorded for the lockout counter)
- `429` — `{"error": "Too many login attempts"}` (IP rate limit: 10 per 15
  minutes) or `{"error": "Too many failed login attempts"}` (account lockout:
  5 failures per 15 minutes)
- `500` — internal error

### `POST /api/forgot-password`

Request a password reset email.

Request body:

```json
{ "email": "ada@example.com" }
```

Always returns the same message whether or not the email exists, to avoid
revealing registered accounts. The reset link is
`{BASE_URL}/reset-password?token={token}` and expires after 30 minutes.

Responses:
- `200` — `{"message": "If the email exists, a reset link has been sent"}`
- `400` — invalid body or email
- `429` — `{"error": "Exceeded password reset limit"}` (5 per 15 minutes per
  IP)
- `500` — internal error

### `POST /api/reset-password`

Set a new password using the token from the reset email. Tokens are single-use
and expire after 30 minutes.

Request body:

```json
{
  "token": "<token-from-email>",
  "password": "new-password"
}
```

Responses:
- `200` — `{"message": "Password successfully reset"}`
- `400` — `{"error": "Invalid or expired token"}` or validation error
- `500` — internal error

---

## Goals

Protected routes. All require `Authorization: Bearer <token>`, and the token's
user must own the goal.

### `GET /api/goals`

List the authenticated user's goals, each with `progress` (the sum of its
deposits, `0` when none exist).

Responses:
- `200`:

  ```json
  [
    {
      "id": "uuid",
      "target": 2499,
      "deadline": "2026-06-01T00:00:00Z",
      "user_id": "uuid",
      "progress": 1700
    }
  ]
  ```

- `401` — missing or invalid token
- `500` — internal error

### `POST /api/goals`

Create a goal for the authenticated user.

Request body:

```json
{
  "target": 2499,
  "deadline": "2026-06-01T00:00:00Z",
  "user_id": "uuid"
}
```

`user_id` must match the token's user. `target` cannot be negative; `deadline`
cannot be in the past.

Responses:
- `201`:

  ```json
  {
    "id": "uuid",
    "target": 2499,
    "deadline": "2026-06-01T00:00:00Z",
    "user_id": "uuid"
  }
  ```

- `400` — invalid body, validation error, or unknown `user_id`
- `401` — missing/invalid token, or `user_id` mismatch
- `500` — internal error

### `PUT /api/goals/{goalId}`

Update a goal's `target` and `deadline`.

Request body:

```json
{
  "target": 3000,
  "deadline": "2026-07-01T00:00:00Z"
}
```

Responses:
- `200` — updated goal `{id, target, deadline, user_id}`
- `400` — malformed `goalId` or validation error
- `401` — missing/invalid token
- `403` — `{"error": "mismatch user id"}` (goal belongs to another user)
- `404` — `{"error": "error finding goal"}`
- `500` — internal error

### `DELETE /api/goals/{goalId}`

Delete a goal (its deposits cascade).

Responses:
- `200` — `{"message": "Goal successfully deleted"}`
- `400` — malformed `goalId`
- `401` — missing/invalid token
- `403` — `{"error": "mismatch user id"}`
- `404` — `{"error": "error finding goal"}`
- `500` — internal error

---

## Deposits

Nested under goals. Protected routes; the token's user must own the goal.

### `GET /api/goals/{goalId}/deposits`

List deposits for a goal.

Responses:
- `200`:

  ```json
  [
    {
      "id": "uuid",
      "amount": 500,
      "note": "Starting fund from freelance project",
      "created_at": "2026-08-10T12:00:00Z"
    }
  ]
  ```

- `400` — malformed `goalId`
- `401` — missing/invalid token
- `403` — `{"error": "mismatch user id"}`
- `404` — `{"error": "error finding goal"}`
- `500` — internal error

### `POST /api/goals/{goalId}/deposits`

Add a deposit to a goal. `amount` must be greater than `0`.

Request body:

```json
{
  "amount": 500,
  "note": "Starting fund from freelance project"
}
```

`note` is optional.

Responses:
- `201`:

  ```json
  {
    "id": "uuid",
    "amount": 500,
    "note": "Starting fund from freelance project",
    "created_at": "2026-08-10T12:00:00Z"
  }
  ```

- `400` — malformed `goalId`, invalid body, or validation error
- `401` — missing/invalid token
- `403` — `{"error": "mismatch user id"}`
- `404` — `{"error": "error finding goal"}`
- `500` — internal error

---

## Static files

### `GET /app`

Serves static files (the `index.html` in the working directory) via
`http.FileServer` (see `main.go`). No authentication; each request increments
an in-memory hit counter logged to stdout.
