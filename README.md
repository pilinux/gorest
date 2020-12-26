# GoREST | RESTful API Starter kit

<img align="right" width="350px" src="https://cdn.pilinux.me/images/GoREST/logo/GoREST-Logo.png">

[![Go Report Card](https://goreportcard.com/badge/github.com/piLinux/GoREST)][01]
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FpiLinux%2FGoREST.svg?type=shield)][02]
[![codebeat badge](https://codebeat.co/badges/c92a8584-d6ba-4606-8d6f-3049630f92c6)][03]

GoREST is a starter kit, written in [Golang][11] with [Gin framework][12],
for rapid prototyping and developing a RESTful API. The source code is released
under the [MIT license][13] and is free for any personal or commercial project.



## Updates [Dec 26 - 2020]
- [JWT][14] based authentication is implemented using [dgrijalva/jwt-go][15]
- `One-to-one`, `one-to-many`, and `many-to-many` models are introduced



## Database Support

GoREST uses [GORM][21] as its ORM. GORM supports **SQLite3**, **MySQL** and
**PostgreSQL**.

In GoREST, **MySQL** driver is included. **SQLite3** and **PostgreSQL** drivers
will be included in future releases after thorough testing.



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

<img width="650px" src="https://cdn.pilinux.me/images/GoREST/screenshot/GoREST.API.Demo.PNG">



## Setup and start the production-ready app

- Install a relational database (at the moment, only MySQL driver is included
in GoREST)
- Set up an environment to compile the Go codes (a [quick tutorial][41]
for any Debian based OS)
- Install `git`
- Clone the project `go get -u github.com/piLinux/GoREST`
- At the root of the cloned repository
[`cd $GOPATH/src/github.com/piLinux/GoREST`], execute `go build` to fetch all
the dependencies
- Edit `.env.sample` file and save it as `.env` file at the root of the
project `$GOPATH/src/github.com/piLinux/GoREST`
- Edit the `.env.sample` file located at
`$GOPATH/src/github.com/piLinux/GoREST/database/migrate` and save it as `.env`
- Inside `$GOPATH/src/github.com/piLinux/GoREST/database/migrate`, run
`go run autoMigrate.go` to migrate the database
- At `$GOPATH/src/github.com/piLinux/GoREST`, run `./GoREST` to launch the app

To the following endpoints `GET`, `POST`, `PUT` and `DELETE` requests can be sent:

- http://localhost:port/api/v1/register
  - `POST` [create new account]
```
{
    "Email":"...@example.com",
    "Password":"..."
}
```
- http://localhost:port/api/v1/login
  - `POST` [generate new JWT]
```
{
    "Email":"...@example.com",
    "Password":"..."
}
```
- http://localhost:port/api/v1/users
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
- http://localhost:port/api/v1/users/:id
  - `GET` [fetch hobbies and posts belonged to a specific user]
- http://localhost:port/api/v1/users/hobbies
  - `PUT` [add a new hobby, requires JWT for verification]
```
{
    "Hobby": "..."
}
```
- http://localhost:port/api/v1/posts
  - `GET` [fetch all published posts]
  - `POST` [create a new post, requires JWT for verification]
```
{
    "Title": "...",
    "Body": "... ..."
}
```
- http://localhost:port/api/v1/posts/:id
  - `GET` [fetch a specific post]
  - `PUT` [edit a specific post, requires JWT for verification]
```
{
    "Title": "...",
    "Body": "... ..."
}
```
  - `DELETE` [delete a specific post, requires JWT for verification]
- http://localhost:port/api/v1/hobbies
  - `GET` [fetch all hobbies created by all users]



## Flow diagram

![Flow.Diagram][05]



## Features

- GoREST uses [Gin][12] as the main framework, [GORM][21] as the ORM and
[GoDotEnv][51] for environment configuration
- [dgrijalva/jwt-go][15] is used for JWT authentication
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
GoREST
│---README.md
│---LICENSE
│---.gitignore
│---.env.sample
│---go.mod
│---go.sum
│---main.go
│
└───config
│    └---configMain.go
│    └---database.go
│    └---server.go
│
│───controller
│    └---auth.go
│    └---login.go
│    └---user.go
│    └---post.go
│    └---hobby.go
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
│         └---user.go
│         └---post.go
│         └---hobby.go
│         └---userHobby.go
│
└───lib
│    └───middleware
│         └---cors.go
│         └---jwt.go
│
└───service
     └---auth.go
```

For API development, one needs to focus only on the following files and directories:

```
GoREST
│---main.go
│
│───controller
│    └---auth.go
│    └---login.go
│    └---user.go
│    └---post.go
│    └---hobby.go
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
└───lib
│    └───middleware
│         └---cors.go
│         └---jwt.go
│
└───service
     └---auth.go
```

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

© Mahir Hasan 2019 - 2020

Released under the [MIT license][13]



[01]: https://goreportcard.com/report/github.com/piLinux/GoREST
[02]: https://app.fossa.io/projects/git%2Bgithub.com%2FpiLinux%2FGoREST?ref=badge_shield
[03]: https://codebeat.co/projects/github-com-pilinux-gorest-master
[04]: https://cdn.pilinux.me/images/GoREST/models/dbModelv1.0.svg
[05]: https://cdn.pilinux.me/images/GoREST/flowchart/flow.diagram.v1.0.svg
[11]: https://github.com/golang/go
[12]: https://github.com/gin-gonic/gin
[13]: LICENSE
[14]: https://jwt.io/introduction
[15]: https://github.com/dgrijalva/jwt-go
[21]: https://github.com/jinzhu/gorm
[31]: https://goapi.pilinux.me/api/v1/users
[32]: https://getpostman.com
[41]: https://github.com/piLinux/HowtoCode/blob/master/Golang/1.Intro/Installation.md
[51]: https://github.com/joho/godotenv
[61]: CONTRIBUTING.md
[62]: CODE_OF_CONDUCT.md


[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FpiLinux%2FGoREST.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FpiLinux%2FGoREST?ref=badge_large)
