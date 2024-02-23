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
		Version: "v0.0.7",
		Authors: []*cli.Author{{Name: "Vanillaiice", Email: "vanillaiice1@proton.me"}},
		Usage:   "Backend server for ITPG, database transactions and user state management through HTTP requests.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   "5555",
			},
			&cli.PathFlag{
				Name:    "db",
				Aliases: []string{"d"},
				Value:   "itpg.db",
			},
			&cli.PathFlag{
				Name:    "users-db",
				Aliases: []string{"u"},
				Value:   "users.db",
			},
			&cli.PathFlag{
				Name:    "env",
				Aliases: []string{"e"},
				Value:   ".env",
			},
			&cli.StringSliceFlag{
				Name:     "allowed-origins",
				Aliases:  []string{"o"},
				Value:    &cli.StringSlice{},
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "allowed-mail-domains",
				Aliases:  []string{"m"},
				Value:    &cli.StringSlice{},
				Required: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			return itpg.Run(ctx.String("port"), ctx.Path("db"), ctx.Path("users-db"), ctx.Path("env"), ctx.StringSlice("allowed-origins"), ctx.StringSlice("allowed-mail-domains"))
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
