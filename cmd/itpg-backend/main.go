package main

import (
	"itpg"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "itpg-backend",
		Suggest: true,
		Version: "v0.0.13",
		Authors: []*cli.Author{{Name: "vanillaiice", Email: "vanillaiice1@proton.me"}},
		Usage:   "Backend server for ITPG, handles database transactions and user state management through HTTP(S) requests.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "listen on `PORT`",
				Value:   "443",
			},
			&cli.PathFlag{
				Name:    "db",
				Aliases: []string{"d"},
				Usage:   "professors, courses and scores sqlite database",
				Value:   "itpg.db",
			},
			&cli.BoolFlag{
				Name:    "db-speed",
				Aliases: []string{"b"},
				Usage:   "prioritize database transaction speed at the cost of data integrity",
				Value:   false,
			},
			&cli.PathFlag{
				Name:    "users-db",
				Aliases: []string{"u"},
				Usage:   "user state management bolt database",
				Value:   "users.db",
			},
			&cli.IntFlag{
				Name:    "cookie-timeout",
				Aliases: []string{"i"},
				Usage:   "cookie timeout in minutes",
				Value:   30,
			},
			&cli.PathFlag{
				Name:    "env",
				Aliases: []string{"e"},
				Usage:   "SMTP configuration file",
				Value:   ".env",
			},
			&cli.StringFlag{
				Name:     "pass-reset-url",
				Aliases:  []string{"r"},
				Usage:    "URL of the password reset web page",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "allowed-origins",
				Aliases:  []string{"o"},
				Usage:    "only allow specified origins to access resources",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "allowed-mail-domains",
				Aliases:  []string{"m"},
				Usage:    "only allow specified mail domains to register",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "smtp",
				Usage:   "use SMTP instead of SMTPS",
				Aliases: []string{"s"},
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "http",
				Usage:   "use HTTP instead of HTTPS",
				Aliases: []string{"t"},
				Value:   false,
			},
			&cli.PathFlag{
				Name:    "cert-file",
				Aliases: []string{"c"},
				Usage:   "SSL certificate file",
			},
			&cli.PathFlag{
				Name:    "key-file",
				Aliases: []string{"k"},
				Usage:   "SSL secret key file",
			},
		},
		Action: func(ctx *cli.Context) error {
			return itpg.Run(
				&itpg.RunConfig{
					Port:                    ctx.String("port"),
					DBPath:                  ctx.Path("db"),
					UsersDBPath:             ctx.Path("users-db"),
					CookieTimeout:           ctx.Int("cookie-timeout"),
					SMTPEnvPath:             ctx.Path("env"),
					PasswordResetWebsiteURL: ctx.String("pass-reset-url"),
					Speed:                   ctx.Bool("speed"),
					AllowedOrigins:          ctx.StringSlice("allowed-origins"),
					AllowedMailDomains:      ctx.StringSlice("allowed-mail-domains"),
					UseSMTP:                 ctx.Bool("smtp"),
					UseHTTP:                 ctx.Bool("http"),
					CertFilePath:            ctx.Path("cert-file"),
					KeyFilePath:             ctx.Path("key-file"),
				},
			)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
