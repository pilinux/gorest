# gorest | RESTful API Starter kit

<img align="right" width="350px" src="https://cdn.pilinux.workers.dev/images/GoREST/logo/GoREST-Logo.png">

![CodeQL][02]
![Go][07]
![Linter][08]
[![Codecov][04]][05]
[![Go Reference][14]][15]
[![Go Report Card](https://goreportcard.com/badge/github.com/pilinux/gorest)][01]
[![CodeFactor](https://www.codefactor.io/repository/github/pilinux/gorest/badge)][06]
[![codebeat badge](https://codebeat.co/badges/3e3573cc-2e9d-48bc-a8c5-4f054bfdfcf7)][03]
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

_Note:_ For version `1.4.5`: [v1.4.5](https://github.com/pilinux/gorest/tree/v1.4.5)

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

Please study the `.env.sample` file. It is one of the most crucial files required
to properly set up a new project. Please rename the `.env.sample` file to `.env`,
and set the environment variables according to your own instance setup.

_Tutorials:_

For version `1.6.x`, please check the project in [example](example)

For version `1.4.x` and `1.5.x`, [Wiki][10]

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

© Mahir Hasan 2019 - 2023

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
[41]: https://github.com/piLinux/HowtoCode/blob/master/Golang/1.Intro/Installation.md
[42]: https://github.com/piLinux/GoREST/issues/7
[51]: https://github.com/joho/godotenv
[61]: CONTRIBUTING.md
[62]: CODE_OF_CONDUCT.md
