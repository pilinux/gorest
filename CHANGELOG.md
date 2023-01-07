# Changelog

## v1.6.x and above

Please check the releases

## v1.6.0-rc.1 [Sep 03 - 2022]

Link: https://github.com/pilinux/gorest/releases/tag/v1.6.0-rc.1

&#9889; optimized database configuration files

&#9889; optimized web application firewall

&#9889; JSON protection from hijacking

&#9889; better handling of JWT

&#9889; two-factor authentication

&#9889; email verification

&#9889; password recovery

&#9889; password update

## v1.5.1 [Jul 23 - 2022]

Link: https://github.com/pilinux/gorest/releases/tag/v1.5.1

&#9889; middleware, renderer and commonly used functions merged here

After hours of testing, it felt more intuitive
to have all middleware inside `gorest`.
There is no need to import anything from `gorestlib` anymore.

## v1.5.0 [Jul 23 - 2022] [_Do not use this version_]

- Release and tag removed from github to avoid import

&#9889; middleware, renderer and commonly used functions moved to a separate repo `github.com/pilinux/gorestlib`

&#9889; `logrus` updated to 1.9.0

&#9889; `postgres` updated to 1.3.8

## v1.4.5 [Jul 18 - 2022]

Link: https://github.com/pilinux/gorest/releases/tag/v1.4.5

&#9889; `gin` updated to 1.8.1

&#9889; `gorm` updated to 1.23.8

&#9889; gorm `mysql` driver updated to 1.3.5

&#9889; gorm `sqlite` driver updated to 1.3.6

&#9889; mongodb `mongo` driver updated to 1.10.0

&#9889; `Qmgo` updated to 1.1.1

&#9889; `radix` driver updated to 4.1.0

## v1.4.4 [Jun 05 - 2022]

Link: https://github.com/pilinux/gorest/releases/tag/v1.4.4

&#9889; gorm `postgres` driver updated to 1.3.7

&#9889; gorm `mysql` driver updated to 1.3.4

&#9889; gorm `mongo` driver updated to 1.9.1

&#9889; `gorm` updated to 1.23.5

&#9889; `Qmgo` updated to 1.1.0

## v1.4.3 [Mar 22 - 2022]

Link: https://github.com/pilinux/gorest/releases/tag/v1.4.3

&#9889; MongoDB driver switched to `Qmgo`

&#9889; Controller examples for MongoDB updated

&#9889; Critical security issues (CWE-089, CWE-943) fixed in controller examples

&#9889; Code refactored in database config files

## v1.4.2 [Mar 14 - 2022]

Link: https://github.com/pilinux/gorest/releases/tag/v1.4.2

&#9889; MongoDB driver added

&#9889; Bump to gorm 1.23.2

&#9889; Error checks during initialization of redis

&#9889; Option to enable/disable RDBMS

## v1.4.1 [Jan 15 - 2022]

Link: https://github.com/pilinux/gorest/releases/tag/v1.4.1

&#9889; Bump to gorm 1.22.5

&#9889; More error checks during gin engine setup

&#9889; New action workflows added to examine, build, and static analysis of the code

## v1.4.0 [Jan 07 - 2022]

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

## v1.3.1 [Dec 31 - 2021]

- During the login process, if the provided email is not found,
  API should handle it properly
- A user must not be able to modify resources related to other users
  (controllers have been updated)

## v1.3.0 [Dec 28 - 2021]

- refactored config files to reduce cyclomatic complexity
- organized instance variables

## v1.2.7 [Dec 27 - 2021]

- REDIS database driver and test endpoints added
- removed ineffectual assignments
- check errors during binding of incoming JSON

## v1.2.6 [Dec 26 - 2021]

- fixed security vulnerability [CWE-190][71] and [CWE-681][72]

## v1.2.5 [Dec 25 - 2021]

- new endpoint added for refreshing JWT tokens

## v1.2.4 [Aug 02 - 2021]

- middleware added: `logrus` + `sentry.io`

## v1.2.3 [Jul 31 - 2021]

- Route handlers modified to meet the requirements of doing unit test

## v1.2.2 [Jul 29 - 2021]

- Replaced `github.com/dgrijalva/jwt-go` with `github.com/golang-jwt/jwt`

Package `github.com/dgrijalva/jwt-go <= v3.2.0` allows attackers to bypass
intended access restrictions in situations with []string{} for m["aud"]
(which is allowed by the specification).
More on this: https://github.com/advisories/GHSA-w73w-5m7g-f7qc

## v1.2.1 [Jun 19 - 2021]

- `SHA-256` is replaced by `Argon2id` for password hashing

## v1.2.0 [Jun 17 - 2021]

- `GORM` updated from `v1` to `v2`

Projects developed based on `GORM v1` must checkout at `v1.1.3`

## v1.1 [Jan 03 - 2021]

- **PostgreSQL** and **SQLite3** drivers are included
- `charset` updated from `utf8` to `utf8mb4` in order to fully support UTF-8
  encoding for MySQL database

## v1.0 [Dec 26 - 2020]

- [JWT][14] based authentication is implemented using [dgrijalva/jwt-go][15]
- `One-to-one`, `one-to-many`, and `many-to-many` models are introduced

[14]: https://jwt.io/introduction
[15]: https://github.com/dgrijalva/jwt-go
[71]: https://cwe.mitre.org/data/definitions/190.html
[72]: https://cwe.mitre.org/data/definitions/681.html
