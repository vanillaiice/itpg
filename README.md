# Is This Professor Good ? - Backend (itpg-backend)

[![Go Reference](https://pkg.go.dev/badge/golang.org/x/example.svg)](https://pkg.go.dev/github.com/vanillaiice/itpg)

Backend for itpg, which is a platform where students can grade their professors after taking courses.
This permits future students to make more informed decisions when choosing their courses.
This repository handles http requests, database transactions, and user state management.

# Installation

## Go install

```sh
$ go install github.com/vanillaiice/itpg/cmd/itpg@latest
```

## Docker

```sh
$ docker pull vanillaiice/itpg:latest
```

# Usage

```sh
NAME:
   itpg-backend - Backend server for ITPG, handles database transactions and user state management through HTTP(S) requests.

USAGE:
   itpg-backend [global options] command [command options]

VERSION:
   v0.4.3

AUTHOR:
   vanillaiice <vanillaiice1@proton.me>

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port PORT, -p PORT                                                               listen on PORT (default: "443")
   --db-backend value, -b value                                                       database backend, either sqlite or postgres (default: "sqlite")
   --db URL, -d URL                                                                   database connection URL (default: "itpg.db")
   --users-db value, -u value                                                         user state management bolt database (default: "users.db")
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
   --load FILE, -l FILE                                                               load TOML config from FILE
   --help, -h                                                                         show help
   --version, -v                                                                      print the version
```

# Examples

## Using Go

If itpg was installed using `go install`, you can simply run it from the command line.

However, there should be an .env file containing the SMTP credentials needed to send confirmation emails.

```sh
# run the server with HTTP and pass an env file
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

# Author

Vanillaiice

# Licence

GPLv3
