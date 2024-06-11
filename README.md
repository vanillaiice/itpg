# Is This Professor Good ? - Backend (itpg-backend) [![Go Reference](https://pkg.go.dev/badge/golang.org/x/example.svg)](https://pkg.go.dev/github.com/vanillaiice/itpg) [![Go Report Card](https://goreportcard.com/badge/github.com/vanillaiice/itpg)](https://goreportcard.com/report/github.com/vanillaiice/itpg)

Backend server for itpg, which is a platform where students can grade their professors after taking courses.
This allows students to make more informed decisions when choosing their courses.
This package handles http requests, database transactions, and user state management.

# Installation

## Go install

```sh
$ go install github.com/vanillaiice/itpg@latest
```

## Docker

```sh
$ docker pull vanillaiice/itpg:latest
```

## git

```sh
$ git clone https://github.com/vanillaiice/itpg
$ cd itpg
$ go install .
```

# Setup

## Installation

First, install the itpg package following the instructions above.

## Super admin user

When running itpg, if the the users database does not exists, we will be prompted to create a super admin user.

We can set the super admin's credentials using environment variables:

```sh
$ export ADMIN_USERNAME=admin ADMIN_PASSWORD=password ADMIN_EMAIL=admin@admin.com
# or
$ ADMIN_USERNAME=admin ADMIN_PASSWORD=password ADMIN_EMAIL=admin@admin.com itpg ---load config.toml
# or put them in a file named .env
$ itpg --load config.toml
```

Or enter them interactively when running itpg:

```sh
$ itpg
{"level":"info","time":"2024-06-10T13:32:02+00:00","message":"Initializing users database users.db"}
enter admin username:
admin
enter admin password:
password
enter admin email:
admin@admin.com
{"level":"info","time":"2024-06-10T13:32:21+00:00","message":"Initialized users database users.db with super admin admin"}
```

## Mail client

To send confirmation and reset code to users, we need to use a SMTP mail server.

We can use a third party mail sending service like [Mailtrap](https://mailtrap.io/), [SendGrid](https://sendgrid.com/) or [Mailgun](https://mailgun.com/).

Or manually set up a self-hosted mail server on using the following guides:
- [Landchad](https://landchad.net/mail/smtp/)
- [linuxbabe](https://www.linuxbabe.com/mail-server/postfix-send-only-multiple-domains-ubuntu)

We can also use Docker:
- [ixtodai/smtp](https://gitlab.com/ix.ai/smtp/)
- [docker-mailserver](https://github.com/docker-mailserver/docker-mailserver)

After setting up the mail server, we store the SMTP credentials in an .env file:

```sh
SMTP_HOST = "example.com"
SMTP_PORT = "587"
MAIL_FROM = "mailer@example.com"
USERNAME = "mailer"
PASSWORD = "mailerr"
```

If the mail server is running on the same machine as itpg, the .env file should look like this:

```sh
SMTP_HOST = "localhost"
SMTP_PORT = "25"
MAIL_FROM = "mailer@example.com"
```

## Handlers

The handlers.json file contains the configuration for the server's HTTP endpoints.

### Configuring handlers

- `path` is the path of the HTTP endpoint (e.g. `/course/grade`).

- `pathType` is the type of the path.

Path types include `super`, `admin`, `user`, and `public`.

- `handler` is the name of the Go function that handles the HTTP request.

- `limiter` is the type of the limiter function that handles the HTTP request.

Limiter types include `lenient` (1000 req/s/ip), `moderate` (1000 req/min/ip), `strict` (500 req/hr/ip), and `veryStrict` (100 req/hr/ip).

- `method` is the HTTP method of the HTTP request.

Methods include `GET`, `POST`, `PUT`, and `DELETE`.

> method names should be in uppercase.

### handlers.json snippet:

```json
{
	"handlers": [
		{
			"path": "/course/grade",
			"pathType": "user",
			"handler": "gradeCourseProfessor",
			"limiter": "moderate",
			"method": "POST"
		},
		{
			"path": "/refresh",
			"pathType": "user",
			"handler": "refreshCookie",
			"limiter": "lenient",
			"method": "POST" 
		},
```

> There should be a sample handlers.json file in the root of the project, that can be used as a reference.

## HTTPS

It is <strike>`mandatory`</strike> recommended to use HTTPS when running the itpg server.

### Manual setup

#### Certbot

We can use `certbot` to generate the needed server.key and server.crt files:

```sh
$ sudo certbot certonly --standalone -d <YOUR_DOMAIN>
```

The needed files should be available at `/etc/letsencrypt/live/<YOUR_DOMAIN>/`.
The `privkey.pem`and `fullchain.pem` represent the `server.key` and `server.crt` files, respectively.

> note: make sure to renew the certificate before they expire as they are only valid for 90 days.

#### Self signed (not recommended)

We can generate the needed server.key and server.crt files using `openssl`:

```sh
$ openssl genrsa -out server.key 2048
$ openssl req -new -x509 -key server.key -out server.crt -days 3650
```

> source: https://github.com/denji/golang-tls

After getting the files, we can pass them to the server like so:

```sh
$ itpg --key server.key --cert server.crt
```

### Using Caddy

We can set up automatic HTTPS with `Caddy`.

First, run the itpg server locally on the machine with HTTP.

Let's run it on port 5555:

```sh
$ itpg --port 5555 -http
```

We then create a Caddy reverse proxy to the itpg server with the following Caddyfile:

```
https://<YOUR_DOMAIN> {
    reverse_proxy localhost:5555
}
```

Finally, we can run Caddy:

```sh
$ caddy run --config Caddyfile
```

> note: ports 80 and 443 need to be open for Caddy to generate the certificates.

> source: https://caddyserver.com/docs/quick-starts/reverse-proxy

> source: https://caddyserver.com/docs/automatic-https

## Seeding the database

For the itpg server to be functional, we need to seed the database with data.

It can easily be done in most programming language that has support for sqlite or postgres.

There is an example in [Go](https://github.com/vanillaiice/itpg-seeder), that uses the [jaswdr/faker](https://github.com/jaswdr/faker) package to seed the database with fake data.

> note: Ideally, we will need to gather data on the institutions where we want to implement itpg.

> The needed data include course codes, names, professor names, and the courses given by professors.

> For more information about the structure of the database, please read the table schemas in the db package.

## Config

Please read the sample-config.toml file in the root of the project.

# Usage

```sh
NAME:
   itpg-backend - Backend server for ITPG, handles database transactions and user state management through HTTP(S) requests.

USAGE:
   itpg-backend [global options] command [command options]

VERSION:
   v0.7.0

AUTHOR:
   vanillaiice <vanillaiice1@proton.me>

COMMANDS:
   run, r    run itpg server
   admin, a  admin management
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port PORT, -p PORT                                                               listen on PORT (default: "443")
   --db-backend value, -b value                                                       database backend, either sqlite or postgres (default: "sqlite")
   --db URL, -d URL                                                                   database connection URL (default: "itpg.db")
   --users-db value, -u value                                                         user state management bolt database (default: "users.db")
   --cache-db URL, -C URL                                                             cache redis database connection URL
   --cache-ttl value, -T value                                                        cache time-to-live in seconds (default: 10)
   --log-level value, -g value                                                        log level (default: "info")
   --cookie-timeout value, -i value                                                   cookie timeout in minutes (default: 30)
   --env FILE, -e FILE                                                                load SMTP configuration from FILE (default: ".env")
   --pass-reset-url URL, -r URL                                                       password reset web page URL
   --allowed-origins value, -o value [ --allowed-origins value, -o value ]            only allow specified origins to access resources (default: "*")
   --allowed-mail-domains value, -m value [ --allowed-mail-domains value, -m value ]  only allow specified mail domains to register (default: "*")
   --smtp, -s                                                                         use SMTP instead of SMTPS (default: false)
   --http, -t                                                                         use HTTP instead of HTTPS (default: false)
   --cert-file FILE, -c FILE                                                          load SSL certificate file from FILE
   --key-file FILE, -k FILE                                                           laod SSL secret key from FILE
   --code-validity-min value, -I value                                                code validity in minutes (default: 180)
   --code-length value, -L value                                                      length of generated codes (default: 8)
   --min-password-score value, -S value                                               minimum acceptable password score computed by zxcvbn (default: 3)
   --handler-config FILE, -n FILE                                                     load JSON handler config from FILE (default: "handlers.json")
   --load FILE, -l FILE                                                               load TOML config from FILE
   --help, -h                                                                         show help
   --version, -v                                                                      print the version
```

# Examples

## Using Go

If itpg was installed using `go install`, you can simply run it from the command line:

```sh
# run the server with HTTP and pass an env file containing SMTP credentials
$ itpg -t -e .smtp-env

# run the server with a TOML config file
$ itpg -l config.toml
```

## Using Docker

```sh
# run the server with HTTPS and pass a TOML config file
$ ls itpg-data
# output: server.crt cert.key config.toml
$ docker run --rm -v ${PWD}/itpg-data:/itpg-data vanillaiice/itpg --load itpg-data/config.toml
```

# Related Projects

- [itpg-docker-compose](https://github.com/vanillaiice/itpg-docker-compose)
- [itpg-frontend](https://github.com/vanillaiice/itpg-frontend)

# Author

vanillaiice

# Licence

GPLv3
