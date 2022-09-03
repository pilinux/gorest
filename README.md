# gorest | RESTful API Starter kit

<img align="right" width="350px" src="https://cdn.pilinux.workers.dev/images/GoREST/logo/GoREST-Logo.png">

![CodeQL][02]
![Go][07]
![Linter][08]
[![Go Report Card](https://goreportcard.com/badge/github.com/pilinux/gorest)][01]
[![CodeFactor](https://www.codefactor.io/repository/github/pilinux/gorest/badge)][06]
[![codebeat badge](https://codebeat.co/badges/3e3573cc-2e9d-48bc-a8c5-4f054bfdfcf7)][03]
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)][13]
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)][62]

gorest is a starter kit, written in [Golang][11] with [Gin framework][12],
for rapid prototyping and developing a RESTful API. The source code is released
under the [MIT license][13] and is free for any personal or commercial project.

## Versioning

`1.MAJOR.MINOR.PATCH`

`1`: production-ready

`MAJOR`: breaking changes

`MINOR`: new functionality in a backwards compatible manner

`PATCH`: [optional] backwards compatible bug fixes

## Important

Version `1.6.x` contains breaking changes!

_Note:_ For version `1.4.5`: [v1.4.5](https://github.com/pilinux/gorest/tree/v1.4.5)

## Requirement

`Go 1.17+`

## Supported databases

- [x] MySQL
- [x] PostgreSQL
- [x] SQLite3
- [x] Redis
- [x] MongoDB

_Note:_ gorest uses [GORM][21] as its ORM

## Features

- [x] built on top of [Gin][12]
- [x] use the supported databases without writing any extra configuration files
- [x] environment variables using [GoDotEnv][51]
- [x] CORS policy
- [x] basic auth
- [x] two-factor authentication
- [x] JWT using [golang-jwt/jwt][16]
- [x] password hashing with `Argon2id`
- [x] JSON protection from hijacking
- [x] simple firewall (whitelist/blacklist IP)
- [x] email validation (pattern + MX lookup)
- [x] email verification (sending verification code)
- [x] forgotten password recovery
- [x] render `HTML` templates
- [x] forward error logs and crash reports to [sentry.io][17]
- [x] super easy to learn and use - lots of example codes

## Start building

_Tutorials:_ [Wiki][10]

For version `1.6.x`, I will publish new tutorials in my free time.

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
- check the [Wiki][10] for tutorials

_Note:_ Omit the line `setPkFk()` in `autoMigrate.go` file if the driver is not **MySQL**
([check issue: 7][42])

**Note For SQLite3:**

- `DBUSER`, `DBPASS`, `DBHOST` and `DBPORT` environment variables
  should be left unchanged.
- `DBNAME` must contain the full path and the database file name; i.e,

```env
/user/location/database.db
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
[06]: https://www.codefactor.io/repository/github/pilinux/gorest
[07]: https://github.com/pilinux/gorest/actions/workflows/go.yml/badge.svg
[08]: https://github.com/pilinux/gorest/actions/workflows/golangci-lint.yml/badge.svg
[10]: https://github.com/pilinux/gorest/wiki
[11]: https://github.com/golang/go
[12]: https://github.com/gin-gonic/gin
[13]: LICENSE
[16]: https://github.com/golang-jwt/jwt
[17]: https://sentry.io/
[21]: https://gorm.io
[41]: https://github.com/piLinux/HowtoCode/blob/master/Golang/1.Intro/Installation.md
[42]: https://github.com/piLinux/GoREST/issues/7
[51]: https://github.com/joho/godotenv
[61]: CONTRIBUTING.md
[62]: CODE_OF_CONDUCT.md
