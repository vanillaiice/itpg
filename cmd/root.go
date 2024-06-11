package cmd

import (
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"github.com/vanillaiice/itpg/server"
)

var rootCmd *cli.Command = &cli.Command{
	Name:    "run",
	Aliases: []string{"r"},
	Usage:   "run itpg server",
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
				Name:    "cache-db",
				Aliases: []string{"C"},
				Usage:   "cache redis database connection `URL`",
				Value:   "",
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Name:    "cache-ttl",
				Aliases: []string{"T"},
				Usage:   "cache time-to-live in seconds",
				Value:   10,
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
				Name:    "smtp-env",
				Aliases: []string{"e"},
				Usage:   "load SMTP configuration from env `FILE`",
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
				Name:    "cert",
				Aliases: []string{"c"},
				Usage:   "load SSL certificate file from `FILE`",
			},
		),
		altsrc.NewPathFlag(
			&cli.PathFlag{
				Name:    "key",
				Aliases: []string{"k"},
				Usage:   "laod SSL secret key from `FILE`",
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Name:    "code-validity",
				Aliases: []string{"I"},
				Usage:   "code validity in minutes",
				Value:   180,
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Name:    "code-length",
				Aliases: []string{"L"},
				Usage:   "length of generated codes",
				Value:   8,
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Name:    "min-password-score",
				Aliases: []string{"S"},
				Usage:   "minimum acceptable password score computed by zxcvbn",
				Value:   3,
			},
		),
		altsrc.NewPathFlag(
			&cli.PathFlag{
				Name:    "handlers",
				Aliases: []string{"H"},
				Usage:   "load JSON handler config from `FILE`",
				Value:   "handlers.json",
			},
		),
		&cli.StringFlag{
			Name:    "load",
			Aliases: []string{"l"},
			Usage:   "load TOML config from `FILE`",
		},
	},
	Action: func(ctx *cli.Context) error {
		return server.Run(
			&server.RunCfg{
				Port:               ctx.String("port"),
				DbUrl:              ctx.String("db"),
				DbBackend:          server.DatabaseBackend(ctx.String("db-backend")),
				CacheDbUrl:         ctx.String("cache-db"),
				CacheTtl:           ctx.Int("cache-ttl"),
				UsersDbPath:        ctx.Path("users-db"),
				AllowedOrigins:     ctx.StringSlice("allowed-origins"),
				AllowedMailDomains: ctx.StringSlice("allowed-mail-domains"),
				PasswordResetUrl:   ctx.String("pass-reset-url"),
				SmtpEnvPath:        ctx.Path("smtp-env"),
				UseSmtp:            ctx.Bool("smtp"),
				UseHttp:            ctx.Bool("http"),
				HandlersFilePath:   ctx.Path("handlers"),
				CertFilePath:       ctx.Path("cert"),
				KeyFilePath:        ctx.Path("key"),
				CookieTimeout:      ctx.Int("cookie-timeout"),
				CodeValidityMinute: ctx.Int("code-validity"),
				CodeLength:         ctx.Int("code-length"),
				MinPasswordScore:   ctx.Int("min-password-score"),
				LogLevel:           server.LogLevel(ctx.String("log-level")),
			},
		)
	},
}
