# GoREST | RESTful API Starter kit

<img align="right" width="350px" src="https://cdn.pilinux.workers.dev/images/GoREST/logo/GoREST-Logo.png">

![CodeQL][02]
![Go][07]
![Linter][08]
[![Go Report Card](https://goreportcard.com/badge/github.com/pilinux/gorest)][01]
[![CodeFactor](https://www.codefactor.io/repository/github/pilinux/gorest/badge)][06]
[![codebeat badge](https://codebeat.co/badges/3e3573cc-2e9d-48bc-a8c5-4f054bfdfcf7)][03]
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)][13]
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)][62]

GoREST is a starter kit, written in [Golang][11] with [Gin framework][12],
for rapid prototyping and developing a RESTful API. The source code is released
under the [MIT license][13] and is free for any personal or commercial project.

## Versioning

`1.MAJOR.MINOR.PATCH`

`1`: can be used in production

`MAJOR`: breaking changes

`MINOR`: new functionality in a backwards compatible manner

`PATCH`: [optional] backwards compatible bug fixes

## Important!

The default branch has been renamed from `master` to `main`. If you have a local
clone, please rename the branch name:

```bash
git branch -m master main
git fetch origin
git branch -u origin/main main
git remote set-head origin -a
```

_Note:_ For version `<= 1.4.5`: https://github.com/pilinux/gorest/tree/v1.4.5

**Upcoming release:** `1.6.x` will contain

- 2FA
- optimized DB configuration files
- better handling of JWT
- json protection from hijacking
- optimzed web application firewall
- password recovery

## Start building

- convention over configuration

```go
import (
  "github.com/pilinux/gorest/config"
  "github.com/pilinux/gorest/database"
  "github.com/pilinux/gorest/lib/middleware"

  "github.com/gin-gonic/gin"
)
```

_Quick tutorial:_ [Wiki][10] + this README.md file

## Database Support

GoREST uses [GORM][21] as its ORM. GORM supports **SQLite3**, **MySQL**,
**PostgreSQL** and **Microsoft SQL Server**.

In GoREST, **MySQL**, **PostgreSQL** and **SQLite3** drivers are included.
Anyone experienced in **Microsoft SQL Server** is welcome to contribute to the
project by including **SQL Server** driver and testing all the features of GoREST.

Newly added drivers: **Redis**, **MongoDB**

## Demo

For demonstration, a test instance can be accessed [here][31] from a web
browser. For API development, it is recommended to use [Postman][32] or any
other similar tool.

Accessible endpoints of the test instance:

- https://goapi.pilinux.me/api/v1/users
- https://goapi.pilinux.me/api/v1/users/:id
- https://goapi.pilinux.me/api/v1/posts
- https://goapi.pilinux.me/api/v1/posts/:id
- https://goapi.pilinux.me/api/v1/hobbies

To prevent abuse, HTTP `GET` requests are accepted by the demo server.

Only the following endpoints accept HTTP `POST` requests to test JWT:

- https://goapi.pilinux.me/api/v1/login

```
{
    "Email": "killua@example.com",
    "Password": "1234.."
}
```

- https://goapi.pilinux.me/api/v1/refresh

```
{
    "RefreshJWT": "",
}
```

<img width="650px" src="https://cdn.pilinux.workers.dev/images/GoREST/screenshot/GoREST.API.Demo.PNG">

## Setup and start a production-ready app

- Install a relational (SQLite3, MySQL or PostgreSQL), Redis, or Mongo database
- Set up an environment to compile the Go codes (a [quick tutorial][41]
  for any Debian based OS)
- Install `git`
- Check the [Wiki][10] + this README.md file to build an application efficiently

_Note:_ Omit the line `setPkFk()` in `autoMigrate.go` file if the driver is not **MySQL**.
[Check issue: 7][42]

**Note For SQLite3:**

- `DBUSER`, `DBPASS`, `DBHOST` and `DBPORT` environment variables
  should be left unchanged.
- `DBNAME` must contain the full path and the database file name; i.e,

```
/user/location/database.db
```

To the following endpoints `GET`, `POST`, `PUT` and `DELETE` requests can be sent:

### Register

http://localhost:port/api/v1/register

- `POST` [create new account]

```
{
    "Email":"...@example.com",
    "Password":"..."
}
```

### Login

http://localhost:port/api/v1/login

- `POST` [generate new JWT]

```
{
    "Email":"...@example.com",
    "Password":"..."
}
```

### Refresh JWT

http://localhost:port/api/v1/refresh

- `POST` [generate new JWT]

```
{
    "RefreshJWT":"use_existing_valid_refresh_token"
}
```

### User profile

http://localhost:port/api/v1/users

- `GET` [get list of all registered users along with their hobbies and posts]
- `POST` [add user info to the database, requires JWT for verification]

```
{
    "FirstName": "...",
    "LastName": "..."
}
```

- `PUT` [edit user info, requires JWT for verification]

```
{
    "FirstName": "...",
    "LastName": "..."
}
```

### Hobbies of a user

http://localhost:port/api/v1/users/:id

- `GET` [fetch hobbies and posts belonged to a specific user]

http://localhost:port/api/v1/users/hobbies

- `PUT` [add a new hobby, requires JWT for verification]

```
{
    "Hobby": "..."
}
```

### Posts

http://localhost:port/api/v1/posts

- `GET` [fetch all published posts]
- `POST` [create a new post, requires JWT for verification]

```
{
    "Title": "...",
    "Body": "... ..."
}
```

##### Any specific post

http://localhost:port/api/v1/posts/:id

- `GET` [fetch a specific post]
- `PUT` [edit a specific post, requires JWT for verification]

```
{
    "Title": "...",
    "Body": "... ..."
}
```

- `DELETE` [delete a specific post, requires JWT for verification]

### List of hobbies available in the database

http://localhost:port/api/v1/hobbies

- `GET` [fetch all hobbies created by all users]

## For REDIS

- Set environment variable `ACTIVATE_REDIS=yes`
- Set `key:value` pair
  - `POST` http://localhost:port/api/v1/playground/redis_create

```
{
    "Key": "test1",
    "Value": "v1"
}
```

- Fetch `key:value` pair
  - `GET` http://localhost:port/api/v1/playground/redis_read

```
{
    "Key": "test1"
}
```

Demo endpoint [`GET`]: https://goapi.pilinux.me/api/v1/playground/redis_read

```
{
    "Key": "goapi-testKey1"
}
```

- Delete `key:value` pair
  - `DELETE` http://localhost:port/api/v1/playground/redis_delete

```
{
    "Key": "test1"
}
```

- Set hashes with key
  - `POST` http://localhost:port/api/v1/playground/redis_create_hash

```
{
    "Key": "test2",
    "Value":
        {
            "Value1": "v1",
            "Value2": "v2",
            "Value3": "v3",
            "Value4": "v4"
        }
}
```

- Fetch hashes by key
  - `GET` http://localhost:port/api/v1/playground/redis_read_hash

```
{
    "Key": "test2"
}
```

Demo endpoint [`GET`]: https://goapi.pilinux.me/api/v1/playground/redis_read_hash

```
{
    "Key": "goapi-testKey2"
}
```

- Delete a key
  - `DELETE` http://localhost:port/api/v1/playground/redis_delete_hash

```
{
    "Key": "test2"
}
```

## For MongoDB

- Set environment variable `ACTIVATE_MONGO=yes`
- Set environment variable `MONGO_URI`

- Controller examples
  - Create a new document `controller.MongoCreateOne`
  - Fetch all documents from a collection `controller.MongoGetAll`
    - demo endpoint [`GET`]: https://goapi.pilinux.me/api/v1/playground-mongo/mongo_get_all
  - Fetch one document by its ID from a collection `controller.MongoGetByID`
    - demo endpoint [`GET`]: https://goapi.pilinux.me/api/v1/playground-mongo/mongo_get_by_id/622f25573fd9e40a1dbf63b0
  - Fetch all documents from a collection using filters `controller.MongoGetByFilter`
    - demo endpoint [`POST`]: https://goapi.pilinux.me/api/v1/playground-mongo/mongo_get_by_filter
    - In `JSON` payload, one or a combination of the following fields can be used for the filter
    ```
    {
      "id": "622f25573fd9e40a1dbf63b0",
      "formatted_address": "221B Baker St, London NW1 6XE, UK",
      "street_name": "Baker Street",
      "house_number": "221B",
      "postal_code": "NW1 6XE",
      "county": "Greater London",
      "state": "England",
      "state_code": "England",
      "country": "United Kingdom",
      "country_code": "GB"
    }
    ```
    - Following filter will fetch all addresses located in the United Kingdom [in the database, only two addresses are saved]
    ```
    {
      "country": "United Kingdom"
    }
    ```
  - Update one specific document `controller.MongoUpdateByID`
  - Delete a field from a document `controller.MongoDeleteFieldByID`
  - Delete a document by its ID from a collection `controller.MongoDeleteByID`

## Flow diagram

![Flow.Diagram][05]

## Features

- GoREST uses [Gin][12] as the main framework, [GORM][21] as the ORM and
  [GoDotEnv][51] for environment configuration
- [golang-jwt/jwt][16] is used for JWT authentication
- [sentry.io][17] error tracker and performance monitor is enabled by default
  as a hook inside `logrus`. They are included as middleware which can be
  disabled by omitting

```
router.Use(middleware.SentryCapture(configure.Logger.SentryDsn))
```

- All codes are written and organized following a straightforward and
  easy-to-understand approach
- For **Logger** and **Recovery**, Gin's in-built middlewares are used

```
router := gin.Default()
```

- Cross-Origin Resource Sharing (CORS) middleware is located at **lib/middleware**

```
router.Use(middleware.CORS())
```

- Included relationship models are:
  - `one to one`
  - `one to many`
  - `many to many`

## Logical Database Model

![DB.Model.Logical][04]

## Architecture

### List of files

```
gorest
│---README.md
│---LICENSE
│---CONTRIBUTING.md
│---CODE_OF_CONDUCT.md
│---SECURITY.md
│---.gitattributes
│---.gitignore
│---.env.sample
│---go.mod
│---go.sum
│---main.go
│
└───config
│    └---config.go
│    └---database.go
│    └---logger.go
│    └---security.go
│    └---server.go
│    └---view.go
│
└───controller
│    └---auth.go
│    └---login.go
│    └---user.go
│    └---post.go
│    └---hobby.go
│    └---playground.go
│    └---playgroundMongo.go
│
└───database
│    │---dbConnect.go
│    │
│    └───migrate
│    │    └---autoMigrate.go
│    │    └---.env.sample
│    │
│    └───model
│         └---auth.go
│         └---errorMsg.go
│         └---user.go
│         └---post.go
│         └---hobby.go
│         └---userHobby.go
│
└───lib
│    └---hashing.go
│    └---validateEmail.go
│    └---removeAllSpace.go
│    │
│    └───middleware
│    │    └---cors.go
│    │    └---firewall.go
│    │    └---ginpongo2.go
│    │    └---jwt.go
│    │    └---sentry.go
│    │
│    └───renderer
│         └---render.go
│
└───logs
│    └---README.md
│
└───service
     └---auth.go
│
└───templates
     └---error.html
     └---read-article.html
```

For API development, one needs to focus mainly on the following files and directories:

```
gorest
│---main.go
│
│───controller
│    └---auth.go
│    └---login.go
│    └---user.go
│    └---post.go
│    └---hobby.go
│    └---playground.go
│    └---playgroundMongo.go
│
└───database
│    │
│    └───migrate
│    │    └---autoMigrate.go
│    │
│    └───model
│         └---auth.go
│         └---user.go
│         └---post.go
│         └---hobby.go
│         └---userHobby.go
│
└───service
     └---auth.go
```

Default path to the HTML template files: `templates/`

### Step 1

- `model`: This package contains all the necessary models. Each file is
  responsible for one specific table in the database. To add new tables and to
  create new relations between those tables, create new models, and place them in
  this directory. All newly created files should have the same package name.

### Step 2

- `controller`: This package contains all functions to process all related
  incoming HTTP requests.

### Step 3

- `autoMigrate.go`: Names of all newly added models should first be included
  in this file to automatically create the complete database. It also contains
  the function to delete the previous data and tables. When only newly created
  tables or columns need to be migrated, first disable `db.DropTableIfExists()`
  function before executing the file.

### Step 4

- `middleware`: All middleware should belong to this package.

### Step 5 (final step)

- Create new routes inside

```
v1 := router.Group()
{
    ...
    ...
}
```

## Contributing

Please see [CONTRIBUTING][61] to join this amazing project.

## Code of conduct

Please see [this][62] document.

## License

© Mahir Hasan 2019 - 2022

Released under the [MIT license][13]

[01]: https://goreportcard.com/report/github.com/pilinux/gorest
[02]: https://github.com/pilinux/gorest/actions/workflows/codeql-analysis.yml/badge.svg
[03]: https://codebeat.co/projects/github-com-pilinux-gorest-main
[04]: https://cdn.pilinux.workers.dev/images/GoREST/models/dbModelv1.0.svg
[05]: https://cdn.pilinux.workers.dev/images/GoREST/flowchart/flow.diagram.v1.0.svg
[06]: https://www.codefactor.io/repository/github/pilinux/gorest
[07]: https://github.com/pilinux/gorest/actions/workflows/go.yml/badge.svg
[08]: https://github.com/pilinux/gorest/actions/workflows/golangci-lint.yml/badge.svg
[09]: https://github.com/pilinux/postmark
[10]: https://github.com/pilinux/gorest/wiki
[11]: https://github.com/golang/go
[12]: https://github.com/gin-gonic/gin
[13]: LICENSE
[14]: https://jwt.io/introduction
[15]: https://github.com/dgrijalva/jwt-go
[16]: https://github.com/golang-jwt/jwt
[17]: https://sentry.io/
[21]: https://gorm.io
[31]: https://goapi.pilinux.me/api/v1/users
[32]: https://getpostman.com
[41]: https://github.com/piLinux/HowtoCode/blob/master/Golang/1.Intro/Installation.md
[42]: https://github.com/piLinux/GoREST/issues/7
[51]: https://github.com/joho/godotenv
[61]: CONTRIBUTING.md
[62]: CODE_OF_CONDUCT.md
[71]: https://cwe.mitre.org/data/definitions/190.html
[72]: https://cwe.mitre.org/data/definitions/681.html
