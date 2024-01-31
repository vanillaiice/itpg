package main

import (
	"itpg"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "Is This Professor Good - Backend",
		Suggest: true,
		Version: "v0.0.1",
		Authors: []*cli.Author{{Name: "Vanillaiice", Email: "vanillaiice@proton.me"}},
		Usage:   "handle http requests and database transactions",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   "5555",
			},
			&cli.PathFlag{
				Name:     "dbname",
				Aliases:  []string{"db"},
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:    "allowed-origins",
				Aliases: []string{"o"},
				Value:   &cli.StringSlice{},
			},
		},
		Action: func(ctx *cli.Context) error {
			return itpg.Run(ctx.String("port"), ctx.Path("dbname"), ctx.StringSlice("allowed-origins"))
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
