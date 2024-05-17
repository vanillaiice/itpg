# Is This Professor Good ? - Backend (itpg-backend)

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
   v0.4.0

AUTHOR:
   vanillaiice <vanillaiice1@proton.me>

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port PORT, -p PORT                                                               listen on PORT (default: "443")
   --db value, -d value                                                               professors, courses and scores sqlite database (default: "itpg.db")
   --db-backend value, -n value                                                       database backend: sqlite or postgres (default: "sqlite")
   --users-db value, -u value                                                         user state management bolt database (default: "users.db")
   --cookie-timeout value, -i value                                                   cookie timeout in minutes (default: 30)
   --env-path value, -e value                                                         SMTP configuration file (default: ".env")
   --pass-reset-url value, -r value                                                   URL of the password reset web page
   --allowed-origins value, -o value [ --allowed-origins value, -o value ]            only allow specified origins to access resources
   --allowed-mail-domains value, -m value [ --allowed-mail-domains value, -m value ]  only allow specified mail domains to register
   --smtp, -s                                                                         use SMTP instead of SMTPS (default: false)
   --http, -t                                                                         use HTTP instead of HTTPS (default: false)
   --cert-file value, -c value                                                        SSL certificate file
   --key-file value, -k value                                                         SSL secret key file
   --load value, -l value                                                             load TOML config from file
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
