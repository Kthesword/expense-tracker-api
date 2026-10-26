# Expense Tracker API

A production-ready RESTful Expense Tracker API written in Go using Gin, GORM, PostgreSQL, and JWT authentication.

## Features

- **Authentication:** User registration (`/signup`) and session authentication (`/login`) issuing HS256 JWT tokens.
- **Password Security:** Passwords hashed with `bcrypt`.
- **Expense CRUD:** Full management of personal expenses (Create, Read, Update, Delete).
- **Date & Category Filtering:**
  - Past week (`?filter=week`)
  - Past month (`?filter=month`)
  - Last 3 months (`?filter=3months`)
  - Custom date range (`?filter=custom&start=YYYY-MM-DD&end=YYYY-MM-DD`)
  - By category (`?category=Food`)
- **Multi-Tenant Isolation:** Expenses are strictly scoped by authenticated `user_id`. Attempting to access another user's expense yields `404 Not Found`.
- **Supported Categories:** `Food`, `Leisure`, `Electronics`, `Utilities`, `Clothing`, `Health`.

---

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL database (or hosted instance e.g. Neon, Supabase)

### 1. Environment Configuration

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Set your values in `.env`:

```env
PORT=8080
DATABASE_URL=postgres://USER:PASSWORD@HOST/DBNAME?sslmode=require
JWT_SECRET=your-32-character-secret-key-goes-here
```

### 2. Run the Server

```bash
go run ./cmd/api
```

GORM will automatically run migrations for `users` and `expenses` tables on startup.

---

## API Endpoints

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/health` | Server health check | No |
| `POST` | `/signup` | Register new user | No |
| `POST` | `/login` | Log in and receive JWT | No |
| `POST` | `/expenses` | Create a new expense | Yes (Bearer Token) |
| `GET` | `/expenses` | List expenses (with filters) | Yes (Bearer Token) |
| `GET` | `/expenses/:id` | Get single expense | Yes (Bearer Token) |
| `PATCH` | `/expenses/:id` | Partially update expense | Yes (Bearer Token) |
| `DELETE` | `/expenses/:id` | Delete expense | Yes (Bearer Token) |

---

## Testing with Postman

Import the included `postman_collection.json` into Postman. Running `/signup` or `/login` automatically stores the JWT token into the collection variable `{{token}}`.
