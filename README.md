## Project Overview

A furniture moving company booking system built with Go and Supabase.

Core features: customer booking management, admin dashboard, one-time booking link generation, PDF invoice generation, and JWT-based authentication.

## Tech Stack

- **Backend**: Go 1.21.1
- **Database/Auth**: Supabase (PostgreSQL)
- **PDF Generation**: `github.com/go-pdf/fpdf`
- **JWT**: `github.com/golang-jwt/jwt/v5`
- **Frontend**: HTML + HTMX (v1.9.10) + Tailwind CSS (CDN)

## Build & Run

```bash
# Run the application
go run main.go

# Build binary
go build -o moving-company
```

Default port: `8080` (configurable via `PORT` env var)

## Environment Variables

Required in `.env`:
```
SUPABASE_URL=...
SUPABASE_KEY=...
```

## Architecture

Layered architecture with clear separation of concerns:

```
api/          → HTTP handlers and router (HTMX-driven responses)
config/       → Supabase client init
config/service/ → Business logic (token service, invoice generator)
models/       → Data models (Booking, FurnitureItem)
repository/   → Data access layer (Supabase CRUD)
static/       → Frontend HTML (index.html, admin.html, login.html)
main.go       → Entry point
```

## Key Files

| File | Purpose |
|------|---------|
| `api/handlers.go` | Booking CRUD handlers |
| `api/router.go` | Route setup and middleware |
| `api/auth_handler.go` | Login/auth handling |
| `api/token_handler.go` | One-time booking link generation/validation |
| `config/service/token_service.go` | JWT generation and validation |
| `config/service/invoice_generator.go` | PDF invoice creation |
| `models/booking.go` | Booking and FurnitureItem structs |
| `repository/supabase_booking.go` | Booking CRUD against Supabase |
| `repository/supabase_token.go` | Token storage in Supabase |

## API Routes

**Public:**
- `POST /api/submit-booking` — Submit booking
- `GET /api/generate-link` — Generate one-time booking link
- `GET /booking-form?t=<token>` — Render booking form (validates token)
- `POST /api/auth/login` — Authenticate user

**Protected (JWT cookie required):**
- `GET /admin` — Admin dashboard
- `GET /admin/generate-invoice` — Generate PDF invoice
- `GET /api/bookings/` — List all bookings
- `GET /api/bookings/?id=<id>` — Get single booking
- `PUT /api/bookings/` — Update booking
- `DELETE /api/bookings/?id=<id>` — Delete booking

## Database Tables (Supabase)

- `booking_user` — Customer bookings (name, email, phone, addresses, date, status)
- `booking_furniture_items` — Furniture items linked to bookings
- `booking_tokens` — One-time tokens for booking form links

## Booking Status Flow

`PENDING` → `CONFIRMED` → `IN_PROGRESS` → `COMPLETED` / `CANCELLED`

## Coding Conventions

- Follow standard Go conventions: `camelCase` for unexported, `PascalCase` for exported
- Use the **repository pattern** — all DB access goes through `repository/`
- Use the **service layer** for business logic — not directly in handlers
- HTTP handlers return HTMX-compatible HTML snippets where applicable
- Auth middleware: `RequireAuth` wraps protected routes
- Error handling: always explicit `if err != nil` checks — no panic in handlers
- One-time tokens use Base64-encoded gzipped JSON for shareable links

## No Tests

There are currently no automated tests in the project. When adding features, consider at minimum testing the service layer manually.
