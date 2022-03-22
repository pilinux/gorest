# GoREST | RESTful API Starter kit

<img align="right" width="350px" src="https://cdn.pilinux.workers.dev/images/GoREST/logo/GoREST-Logo.png">

![CodeQL][02]
![Go][07]
![Linter][08]
[![Go Report Card](https://goreportcard.com/badge/github.com/pilinux/gorest)][01]
[![CodeFactor](https://www.codefactor.io/repository/github/pilinux/gorest/badge)][06]
[![codebeat badge](https://codebeat.co/badges/3e3573cc-2e9d-48bc-a8c5-4f054bfdfcf7)][03]
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)][13]

GoREST is a starter kit, written in [Golang][11] with [Gin framework][12],
for rapid prototyping and developing a RESTful API. The source code is released
under the [MIT license][13] and is free for any personal or commercial project.

## Important!

The default branch has been renamed from `master` to `main`. If you have a local
clone, please rename the branch name:

```bash
git branch -m master main
git fetch origin
git branch -u origin/main main
git remote set-head origin -a
```

## Start building

- convention over configuration

```go
import (
  "github.com/pilinux/gorest/config"
  "github.com/pilinux/gorest/database"
  "github.com/pilinux/gorest/lib/middleware"
)
```

For inspiration, take a look at a small but real-life project built in one night: [repo][09]

## Updates

### v1.4.3 [Mar 22 - 2022]

Link: https://github.com/pilinux/gorest/releases/tag/v1.4.3

&#9889; MongoDB driver switched to `Qmgo`

&#9889; Controller examples for MongoDB updated

&#9889; Critical security issues (CWE-089, CWE-943) fixed in controller examples

&#9889; Code refactored in database config files

### v1.4.2 [Mar 14 - 2022]

Link: https://github.com/pilinux/gorest/releases/tag/v1.4.2

&#9889; MongoDB driver added

&#9889; Bump to gorm 1.23.2

&#9889; Error checks during initialization of redis

&#9889; Option to enable/disable RDBMS

### v1.4.1 [Jan 15 - 2022]

Link: https://github.com/pilinux/gorest/releases/tag/v1.4.1

&#9889; Bump to gorm 1.22.5

&#9889; More error checks during gin engine setup

&#9889; New action workflows added to examine, build, and static analysis of the code

### v1.4.0 [Jan 07 - 2022]

**Breaking changes!!** If your application is built on one of the previous releases, you need to
do some adjustments to your codes before switching to v1.4.

- Features from development branch `v1.4.0-alpha0` are merged into v1.4
- To build a new RESTful application, you do not need to clone this full repository anymore. It is
  recommended to add the required packages as dependencies
- v1.4 is now pretty solid for any future RESTful application development
- In the upcoming days a full tutorial will be published on how to use `GoREST` packages as
  dependency to build any simple or complex applications within the shortest possible time

Development branch: v1.4.0-alpha0 [Jan 02 - 2022]

Safety: Zero-logs policy for the test server (demo live API)

- If the client is a web browser, or when the client requests to
  serve HTML pages, the API will serve HTML page instead of JSON
- Template files are located in `templates` directory
- Template engine: `Pongo2` - similar syntax like Django
- Templates developed for:
  - `GET` - `/api/v1/posts/:id`: [live demo] https://goapi.pilinux.me/api/v1/posts/1

[Jan 07 - 2022]

- `Render` is now an exported function placed in `lib` package
- `Render` moved from `lib` to `renderer` package
- Config modified for `Basic Auth`
- Demo router added - how to implement `Basic Auth`
  - `GET` - `/api/v1/access_resources`: [live demo] https://goapi.pilinux.me/api/v1/access_resources
    with `USERNAME=test_username` and `PASSWORD=secret_password`
- App firewall added
  - to allow all IPs, set `IP=*`
  - to allow one or several IPs, set `LISTTYPE=whitelist` and `IP=[IPv4 addresses]`
  - to block one or several IPs, set `LISTTYPE=blacklist` and `IP=[IPv4 addresses]`

### v1.3.1 [Dec 31 - 2021]

- During the login process, if the provided email is not found,
  API should handle it properly
- A user must not be able to modify resources related to other users
  (controllers have been updated)

### v1.3.0 [Dec 28 - 2021]

- refactored config files to reduce cyclomatic complexity
- organized instance variables

### v1.2.7 [Dec 27 - 2021]

- REDIS database driver and test endpoints added
- removed ineffectual assignments
- check errors during binding of incoming JSON

### v1.2.6 [Dec 26 - 2021]

- fixed security vulnerability [CWE-190][71] and [CWE-681][72]

### v1.2.5 [Dec 25 - 2021]

- new endpoint added for refreshing JWT tokens

### v1.2.4 [Aug 02 - 2021]

- middleware added: `logrus` + `sentry.io`

### v1.2.3 [Jul 31 - 2021]

- Route handlers modified to meet the requirements of doing unit test

### v1.2.2 [Jul 29 - 2021]

- Replaced `github.com/dgrijalva/jwt-go` with `github.com/golang-jwt/jwt`

Package `github.com/dgrijalva/jwt-go <= v3.2.0` allows attackers to bypass
intended access restrictions in situations with []string{} for m["aud"]
(which is allowed by the specification).
More on this: https://github.com/advisories/GHSA-w73w-5m7g-f7qc

### v1.2.1 [Jun 19 - 2021]

- `SHA-256` is replaced by `Argon2id` for password hashing

### v1.2.0 [Jun 17 - 2021]

- `GORM` updated from `v1` to `v2`

Projects developed based on `GORM v1` must checkout at `v1.1.3`

### v1.1 [Jan 03 - 2021]

- **PostgreSQL** and **SQLite3** drivers are included
- `charset` updated from `utf8` to `utf8mb4` in order to fully support UTF-8
  encoding for MySQL database

### v1.0 [Dec 26 - 2020]

- [JWT][14] based authentication is implemented using [dgrijalva/jwt-go][15]
- `One-to-one`, `one-to-many`, and `many-to-many` models are introduced

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

To prevent abuse, only HTTP `GET` requests are accepted by the demo server.

<img width="650px" src="https://cdn.pilinux.workers.dev/images/GoREST/screenshot/GoREST.API.Demo.PNG">

## Setup and start the production-ready app (old procedure)

- Install a relational database (MySQL or PostgreSQL)
- Set up an environment to compile the Go codes (a [quick tutorial][41]
  for any Debian based OS)
- Install `git`
- Clone the project `git clone https://github.com/piLinux/GoREST.git`
- At the root of the cloned repository
  [`cd $GOPATH/src/github.com/pilinux/gorest`], execute `go build` to fetch all
  the dependencies
- Edit `.env.sample` file and save it as `.env` file at the root of the
  project `$GOPATH/src/github.com/pilinux/gorest`
- Edit the `.env.sample` file located at
  `$GOPATH/src/github.com/pilinux/gorest/database/migrate` and save it as `.env`
- Inside `$GOPATH/src/github.com/pilinux/gorest/database/migrate`, run
  `go run autoMigrate.go` to migrate the database
  - Comment the line `setPkFk()` in `autoMigrate.go` file if the driver is not **MySQL**.
    [Check issue: 7][42]
- At `$GOPATH/src/github.com/pilinux/gorest`, run `./gorest` to launch the app

A new guideline will be published in the following days.

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
│    └───middleware
│    │    └---cors.go
│    │    └---firewall.go
│    │    └---ginpongo2.go
│    │    └---jwt.go
│    │    └---sentry.go
│    └───renderer
│         └---render.go
│
└───logs
│    └---README.md
│
└───service
     └---auth.go
     └---common.go
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
     └---common.go
```

To render and serve HTML pages, the template files must be present in the
`templates` directory

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

- `middleware`: All middlewares should belong to this package.

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
