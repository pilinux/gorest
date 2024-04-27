# gorest | RESTful API Starter kit

<img align="right" width="350px" src="https://cdn.pilinux.workers.dev/images/GoREST/logo/GoREST-Logo.png">

![CodeQL][02]
![Go][07]
![Linter][08]
[![Codecov][04]][05]
[![Go Reference][14]][15]
[![Go Report Card](https://goreportcard.com/badge/github.com/pilinux/gorest)][01]
[![CodeFactor](https://www.codefactor.io/repository/github/pilinux/gorest/badge)][06]
[![codebeat badge](https://codebeat.co/badges/12c01a53-4745-4f90-ad2b-a95c94e4b432)][03]
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)][13]
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)][62]

gorest is a starter kit, written in [Golang][11] with [Gin framework][12],
for rapid prototyping and developing a RESTful API. The source code is released
under the [MIT license][13] and is free for any personal or commercial project.

## Versioning

`1.x.y`

`1`: production-ready

`x`: breaking changes

`y`: new functionality or bug fixes in a backwards compatible manner

## Important

Version `1.6.x` contains breaking changes!

_Note:_ For version `1.4.5` (obsolete): [v1.4.5](https://github.com/pilinux/gorest/tree/v1.4.5)

For all projects, it is recommended to use version `1.6.x` or higher.

## Requirement

`Go 1.19+`

## Supported databases

- [x] MySQL
- [x] PostgreSQL
- [x] SQLite3
- [x] Redis
- [x] MongoDB

_Note:_ gorest uses [GORM][21] as its ORM

## Features

- [x] built on top of [Gin][12]
- [x] option to enable encryption at rest for user private information
- [x] use the supported databases without writing any extra configuration files
- [x] environment variables using [GoDotEnv][51]
- [x] CORS policy
- [x] basic auth
- [x] two-factor authentication
- [x] JWT using [golang-jwt/jwt][16]
- [x] password hashing using `Argon2id` with optional secret (NIST 800-63B
  recommends using a secret value of at least 112 bits)
- [x] JSON protection from hijacking
- [x] simple firewall (whitelist/blacklist IP)
- [x] email validation (pattern + MX lookup)
- [x] email verification (sending verification code)
- [x] forgotten password recovery
- [x] render `HTML` templates
- [x] forward error logs and crash reports to [sentry.io][17]
- [x] handle authentication tokens on client devices' cookies
- [x] logout (individually enable option - delete tokens from cookies, ban active tokens)
- [x] rate limiting (IP-based)
- [x] option to validate origin of the request
- [x] super easy to learn and use - lots of example codes

## Supported JWT signing algorithms

- [x] HS256: HMAC-SHA256
- [x] HS384: HMAC-SHA384
- [x] HS512: HMAC-SHA512
- [x] ES256: ECDSA Signature with SHA-256
- [x] ES384: ECDSA Signature with SHA-384
- [x] ES512: ECDSA Signature with SHA-512
- [x] RS256: RSA Signature with SHA-256
- [x] RS384: RSA Signature with SHA-384
- [x] RS512: RSA Signature with SHA-512

Procedures to generate HS256, HS384, HS512 keys using openssl:

- HS256: `openssl rand -base64 32`
- HS384: `openssl rand -base64 48`
- HS512: `openssl rand -base64 64`

Procedures to generate public-private key pair using openssl:

### ECDSA

#### ES256

- prime256v1: X9.62/SECG curve over a 256 bit prime field, also known as P-256 or NIST P-256
- widely used, recommended for general-purpose cryptographic operations

```bash
openssl ecparam -name prime256v1 -genkey -noout -out private-key.pem
openssl ec -in private-key.pem -pubout -out public-key.pem
```

#### ES384

- secp384r1: NIST/SECG curve over a 384 bit prime field

```bash
openssl ecparam -name secp384r1 -genkey -noout -out private-key.pem
openssl ec -in private-key.pem -pubout -out public-key.pem
```

#### ES512

- secp521r1: NIST/SECG curve over a 521 bit prime field

```bash
openssl ecparam -name secp521r1 -genkey -noout -out private-key.pem
openssl ec -in private-key.pem -pubout -out public-key.pem
```

### RSA

#### RS256

```bash
openssl genpkey -algorithm RSA -out private-key.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -in private-key.pem -pubout -out public-key.pem
```

#### RS384

```bash
openssl genpkey -algorithm RSA -out private-key.pem -pkeyopt rsa_keygen_bits:3072
openssl rsa -in private-key.pem -pubout -out public-key.pem
```

#### RS512

```bash
openssl genpkey -algorithm RSA -out private-key.pem -pkeyopt rsa_keygen_bits:4096
openssl rsa -in private-key.pem -pubout -out public-key.pem
```

## Example docker compose file

```yml
# syntax=docker/dockerfile:1

version: '3.9'
name: go
services:
  goapi:
    image: golang:latest
    container_name: goapi
    working_dir: /app/
    restart: unless-stopped:10s
    command: /app/goapi
    ports:
      - '127.0.0.1:8000:8999'
    volumes:
      - ./app:/app/
```

## Start building

Please study the `.env.sample` file. It is one of the most crucial files required
to properly set up a new project. Please rename the `.env.sample` file to `.env`,
and set the environment variables according to your own instance setup.

_Tutorials:_

For version `1.6.x`, please check the project in [example](example)

For version `1.4.x` and `1.5.x`, [Wiki][10] (obsolete)

- convention over configuration

```go
import (
  "github.com/gin-gonic/gin"

  gconfig "github.com/pilinux/gorest/config"
  gcontroller "github.com/pilinux/gorest/controller"
  gdatabase "github.com/pilinux/gorest/database"
  gmiddleware "github.com/pilinux/gorest/lib/middleware"
)
```

- install a relational (SQLite3, MySQL or PostgreSQL), Redis, or Mongo database
- for 2FA, a relational + a redis database is required
- set up an environment to compile the Go codes (a [quick tutorial][41]
  for any Debian based OS)
- install `git`
- check the [Wiki][10] and [example](example) for tutorials and implementations

_Note:_ For **MySQL** driver, please [check issue: 7][42]

**Note For SQLite3:**

- `DBUSER`, `DBPASS`, `DBHOST` and `DBPORT` environment variables are not required.
- `DBNAME` must contain the full or relative path of the database file name; i.e,

```env
/user/location/database.db
```

or,

```env
./database.db
```

## Debugging with Error Codes

| package | file | error code range |
| ------- | ---- | ---------------- |
| controller | login.go | `1011 - 1012` |
| controller | twoFA.go | `1041 - 1044` |
| database | dbConnect.go | `150 - 155`, `161` |
| handler | auth.go | `1001 - 1003` |
| handler | login.go | `1013 - 1014` |
| handler | logout.go | `1016` |
| handler | passwordReset.go | `1021 - 1030` |
| handler | twoFA.go | `1051 - 1056` |
| handler | verification.go | `1061 - 1065` |
| service | common.go | `401 - 406` |
| service | security.go | `501` |

## Development

For testing:

```bash
export TEST_ENV_URL="https://s3.nl-ams.scw.cloud/ci.config/github.action/gorest.pilinux/.env"
export TEST_INDEX_HTML_URL="https://s3.nl-ams.scw.cloud/ci.config/github.action/gorest.pilinux/index.html"
export TEST_KEY_FILE_LOCATION="https://s3.nl-ams.scw.cloud/ci.config/github.action/gorest.pilinux"
export TEST_SENTRY_DSN="please_set_your_sentry_DSN_here"

go test -v -cover ./...
```

## Contributing

Please see [CONTRIBUTING][61] to join this amazing project.

## Code of conduct

Please see [this][62] document.

## License

© Mahir Hasan 2019 - 2024

Released under the [MIT license][13]

[01]: https://goreportcard.com/report/github.com/pilinux/gorest
[02]: https://github.com/pilinux/gorest/actions/workflows/codeql-analysis.yml/badge.svg
[03]: https://codebeat.co/projects/github-com-pilinux-gorest-main
[04]: https://codecov.io/gh/pilinux/gorest/branch/main/graph/badge.svg?token=xGLBRrCAvB
[05]: https://codecov.io/gh/pilinux/gorest
[06]: https://www.codefactor.io/repository/github/pilinux/gorest
[07]: https://github.com/pilinux/gorest/actions/workflows/go.yml/badge.svg
[08]: https://github.com/pilinux/gorest/actions/workflows/golangci-lint.yml/badge.svg
[10]: https://github.com/pilinux/gorest/wiki
[11]: https://github.com/golang/go
[12]: https://github.com/gin-gonic/gin
[13]: LICENSE
[14]: https://pkg.go.dev/badge/github.com/pilinux/gorest
[15]: https://pkg.go.dev/github.com/pilinux/gorest
[16]: https://github.com/golang-jwt/jwt
[17]: https://sentry.io/
[21]: https://gorm.io
[41]: https://github.com/pilinux/HowtoCode/blob/master/Golang/1.Intro/Installation.md
[42]: https://github.com/pilinux/gorest/issues/7
[51]: https://github.com/joho/godotenv
[61]: CONTRIBUTING.md
[62]: CODE_OF_CONDUCT.md
