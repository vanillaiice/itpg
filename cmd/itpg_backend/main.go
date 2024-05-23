package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"github.com/vanillaiice/itpg"
)

const Version = "v0.4.3"

func main() {
	app := &cli.App{
		Name:    "itpg-backend",
		Suggest: true,
		Version: Version,
		Authors: []*cli.Author{{Name: "vanillaiice", Email: "vanillaiice1@proton.me"}},
		Usage:   "Backend server for ITPG, handles database transactions and user state management through HTTP(S) requests.",
		Flags: []cli.Flag{
			altsrc.NewStringFlag(
				&cli.StringFlag{
					Name:    "port",
					Aliases: []string{"p"},
					Usage:   "listen on `PORT`",
					Value:   "443",
				},
			),
			altsrc.NewStringFlag(
				&cli.StringFlag{
					Name:    "db-backend",
					Aliases: []string{"b"},
					Usage:   "database backend, either sqlite or postgres",
					Value:   "sqlite",
				},
			),
			altsrc.NewStringFlag(
				&cli.StringFlag{
					Name:    "db",
					Aliases: []string{"d"},
					Usage:   "database connection `URL`",
					Value:   "itpg.db",
				},
			),
			altsrc.NewPathFlag(
				&cli.PathFlag{
					Name:    "users-db",
					Aliases: []string{"u"},
					Usage:   "user state management bolt database",
					Value:   "users.db",
				},
			),
			altsrc.NewStringFlag(
				&cli.StringFlag{
					Name:    "log-level",
					Aliases: []string{"g"},
					Usage:   "log level",
					Value:   "info",
				},
			),
			altsrc.NewIntFlag(
				&cli.IntFlag{
					Name:    "cookie-timeout",
					Aliases: []string{"i"},
					Usage:   "cookie timeout in minutes",
					Value:   30,
				},
			),

			altsrc.NewPathFlag(
				&cli.PathFlag{
					Name:    "env",
					Aliases: []string{"e"},
					Usage:   "load SMTP configuration from `FILE`",
					Value:   ".env",
				},
			),
			altsrc.NewStringFlag(
				&cli.StringFlag{
					Name:    "pass-reset-url",
					Aliases: []string{"r"},
					Usage:   "password reset web page `URL`",
				},
			),
			altsrc.NewStringSliceFlag(
				&cli.StringSliceFlag{
					Name:    "allowed-origins",
					Aliases: []string{"o"},
					Usage:   "only allow specified origins to access resources",
					Value:   cli.NewStringSlice("*"),
				},
			),
			altsrc.NewStringSliceFlag(
				&cli.StringSliceFlag{
					Name:    "allowed-mail-domains",
					Aliases: []string{"m"},
					Usage:   "only allow specified mail domains to register",
					Value:   cli.NewStringSlice("*"),
				},
			),
			altsrc.NewBoolFlag(
				&cli.BoolFlag{
					Name:    "smtp",
					Usage:   "use SMTP instead of SMTPS",
					Aliases: []string{"s"},
					Value:   false,
				},
			),
			altsrc.NewBoolFlag(
				&cli.BoolFlag{
					Name:    "http",
					Usage:   "use HTTP instead of HTTPS",
					Aliases: []string{"t"},
					Value:   false,
				},
			),
			altsrc.NewPathFlag(
				&cli.PathFlag{
					Name:    "cert-file",
					Aliases: []string{"c"},
					Usage:   "load SSL certificate file from `FILE`",
				},
			),
			altsrc.NewPathFlag(
				&cli.PathFlag{
					Name:    "key-file",
					Aliases: []string{"k"},
					Usage:   "laod SSL secret key from `FILE`",
				},
			),
			&cli.StringFlag{
				Name:    "load",
				Aliases: []string{"l"},
				Usage:   "load TOML config from `FILE`",
			},
		},
		Action: func(ctx *cli.Context) error {
			return itpg.Run(
				&itpg.RunConfig{
					Port:                    ctx.String("port"),
					DbURL:                   ctx.String("db"),
					DbBackend:               itpg.DatabaseBackend(ctx.String("db-backend")),
					UsersDBPath:             ctx.Path("users-db"),
					LogLevel:                itpg.LogLevel(ctx.String("log-level")),
					CookieTimeout:           ctx.Int("cookie-timeout"),
					SMTPEnvPath:             ctx.Path("env"),
					PasswordResetWebsiteURL: ctx.String("pass-reset-url"),
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

	app.Before = altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewTomlSourceFromFlagFunc("load"))

	if err := app.Run(os.Args); err != nil {
		itpg.Logger.Fatal().Msg(err.Error())
	}
}
