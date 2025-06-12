# Example2: Building a RESTful API with gorest

This example shows how to build a secure, production-ready RESTful API in Go using the gorest framework.
You’ll learn project structure, configuration, database migrations, routing, middleware, and how to add your own resources step by step.

## Table of Contents

1. Features
2. Prerequisites
3. Installation & Setup
4. Environment Configuration
5. Project Structure
6. Running the Application
7. Database Migrations
8. API Reference & Examples
9. Authentication & Security
10. Best Practices
11. Contributing
12. License

## Features

- Multi-database support: MySQL/Postgres/SQLite, Redis, MongoDB
- Flexible authentication: Basic Auth, JWT, 2FA
- Security: CORS, IP firewall, rate limiting, secure headers
- Clean layers: Models → Repositories → Services → Handlers
- Auto migrations, graceful shutdown, performance tracing
- Email verification & password recovery (Postmark)

## Prerequisites

- Go 1.23+ installed
- One or more databases:
  - RDBMS (MySQL, PostgreSQL, or SQLite)
  - Redis (optional)
  - MongoDB (optional)
- Git & a terminal

## Installation & Setup

Note: For a new project, you do not need to clone this project. You only need to import the packages you need in your own project.
But for this example, clone this repo and enter the `example2` directory.

```bash
# 1. Clone repo and enter example2
git clone https://github.com/pilinux/gorest.git
cd gorest/example2

# 2. Copy sample env and edit values
cd cmd/app
cp .env.sample .env
# open .env in your editor and configure

# 3. Fetch dependencies
cd ../..
go mod tidy
```

## Environment Configuration

All settings live in cmd/app/.env. Key sections:

- APP\_\*: Name, host, port, env
- DB: Activate RDBMS & credentials
- REDIS / MONGO: optional cache & NoSQL
- AUTH: Basic Auth, JWT, 2FA flags & keys
- SECURITY: CORS, firewall, rate limiter
- EMAIL: Postmark settings for verification & recovery (optional)

## Project Structure

```
example2/
├── cmd/
│   └── app/
│       ├── main.go          # Application entry
│       └── .env.sample      # Env template
└── internal/
    ├── database/
    │   ├── migrate/         # Auto migrations
    │   └── model/           # GORM, Mongo and Redis models
    ├── handler/             # Gin HTTP handlers
    ├── repo/                # Data access layer
    ├── router/              # Route definitions & middleware
    └── service/             # Business logic
```

- main.go: boots config, DB/Redis/Mongo clients, migrations, router, graceful shutdown
- model: data structures + validation
- repo: DB/Redis/Mongo operations using ORMs
- service: higher-level business flows
- handler: HTTP parsing, response formatting
- router: groups, middleware, versioning

## Running the Application

```bash
# From project root:
cd cmd/app
go build -o app
./app
```

Visit `http://localhost:8999` and explore `/api/v1/...` endpoints.

## Database Migrations

Migrations run automatically at startup if `ACTIVATE_RDBMS=yes`.

To customize tables:

```
// internal/database/migrate/migrate.go
func StartMigration(configure gconfig.Configuration) error {
	db := gdb.GetDB()
	configureDB := configure.Database.RDBMS
	driver := configureDB.Env.Driver

	if err := db.AutoMigrate(
		&auth{},
		&twoFA{},
		&twoFABackup{},
		&tempEmail{},
		&user{},
		&post{},
		&hobby{},
        // add &model.YourNewModel{} here
    )
}
```

## API Reference & Examples

All routes are mounted under the base path `/api/v1` and defined in [`internal/router/router.go`](internal/router/router.go).

**Base URL**: `http://localhost:8999/api/v1`

---

### 🔐 Authentication

#### POST /register

Register a new user.

Body:

```json
{
  "email": "jon@example.com",
  "password": "your_password"
}
```

```bash
curl -X POST http://localhost:8999/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"jon@example.com","password":"your_password"}'
```

#### POST /login

Login with email and password to receive JWT.

Body:

```json
{
  "email": "jon@example.com",
  "password": "your_password"
}
```

```bash
curl -X POST http://localhost:8999/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"jon@example.com","password":"your_password"}'
```

---

### 🧑‍💻 Users

#### GET /users

Retrieve all users, their posts and hobbies.

```bash
curl http://localhost:8999/api/v1/users
```

#### GET /users/:id

Retrieve a single user by `userID`.

```bash
curl http://localhost:8999/api/v1/users/123
```

#### POST /users

Add the user profile. _(Requires JWT)_

Body:

```json
{
  "firstName": "Jon",
  "lastName": "Doe"
}
```

```bash
curl -X POST http://localhost:8999/api/v1/users \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"firstName":"Jon","lastName":"Doe"}'
```

#### PUT /users

Update authenticated user’s profile. _(Requires JWT)_

Body:

```json
{
  "firstName": "Jon",
  "lastName": "Smith"
}
```

#### DELETE /users

Delete authenticated user account and all related posts/hobbies. _(Requires JWT)_

---

### 📝 Posts

#### GET /posts

Get all posts.

```bash
curl http://localhost:8999/api/v1/posts
```

#### GET /posts/:id

Get a post by `postID`.

```bash
curl http://localhost:8999/api/v1/posts/456
```

#### POST /posts

Create a new post. _(Requires JWT)_

Body:

```json
{
  "title": "My First Post",
  "body": "Hello, gorest!"
}
```

```bash
curl -X POST http://localhost:8999/api/v1/posts \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"My First Post","body":"Hello, gorest!"}'
```

#### PUT /posts/:id

Update a post you own. _(Requires JWT)_

Body:

```json
{
  "title": "Updated Title",
  "body": "Updated content"
}
```

#### DELETE /posts/:id

Delete a single post. _(Requires JWT)_

#### DELETE /posts/all

Delete **all** your posts. _(Requires JWT)_

---

### 🎯 Hobbies

#### GET /hobbies

List all hobbies.

```bash
curl http://localhost:8999/api/v1/hobbies
```

#### GET /hobbies/:id

Get hobby by `hobbyID`.

```bash
curl http://localhost:8999/api/v1/hobbies/7
```

#### GET /hobbies/me

List your hobbies. _(Requires JWT)_

#### POST /hobbies

Add a hobby to your profile. _(Requires JWT)_

Body:

```json
{
  "hobby": "painting"
}
```

#### DELETE /hobbies/:id

Remove a hobby from your profile. _(Requires JWT)_

---

### 🗝️ Key-Value (Redis)

#### POST /kv

Set a key/value.

Body:

```json
{
  "key": "theme",
  "value": "dark"
}
```

#### GET /kv/:key

Get value by key.

```bash
curl http://localhost:8999/api/v1/kv/theme
```

#### DELETE /kv/:key

Delete a key/value pair.

```bash
curl -X DELETE http://localhost:8999/api/v1/kv/theme
```

---

### 📍 Addresses (MongoDB)

#### POST /addresses

Add a new address.

Body (Geocoding model):

```json
{
  "formattedAddress": "1600 Amphitheatre Pkwy, Mountain View, CA",
  "city": "Mountain View",
  "state": "California",
  "country": "United States of America",
  "countryCode": "USA",
  "latitude": 37.422,
  "longitude": -122.0841
}
```

#### GET /addresses

List all saved addresses.

```bash
curl http://localhost:8999/api/v1/addresses
```

#### GET /addresses/:id

Get one address by Mongo `_id`.

```bash
curl http://localhost:8999/api/v1/addresses/60d5f4832f8fb814c8a1e7d4
```

#### POST /addresses/filter?exclude-address-id=<true|false>

- Search by fields. _(exclude-address-id: omit `_id` in filter)_
- Body same as above, just send the fields you want to match.

#### PUT /addresses

- Update an existing address.
- Body: full Geocoding object with `_id` and fields to update.

#### DELETE /addresses/:id

Delete an address by Mongo `_id`.

---

Feel free to experiment with these endpoints via **curl**, **Postman**, or your preferred HTTP client.

## Authentication & Security

- Basic Auth: set `ACTIVATE_BASIC_AUTH=yes`
- JWT: set `ACTIVATE_JWT=yes`, configure keys & TTLs
- 2FA: set `ACTIVATE_2FA=yes`
- CORS/firewall/rate limiter: toggle `ACTIVATE_CORS`, `ACTIVATE_FIREWALL`, `RATE_LIMIT`
- Middleware is wired in internal/router.

## Best Practices

- Validate inputs early (handler & model)
- Use context.Context for cancellation & timeouts
- Structure code in layers for testability
- Log errors and HTTP events
- Keep secrets out of code (.env)
- Write unit & integration tests

## Contributing

- Fork & clone
- Create feature branch
- Write tests & update docs
- Submit a PR

## License

MIT © pilinux
