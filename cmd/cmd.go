package cmd

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

// version is the current version.
const version = "v0.6.3"

// Exec starts the cli app.
func Exec() {
	app := &cli.App{
		Name:    "itpg-backend",
		Suggest: true,
		Version: version,
		Authors: []*cli.Author{{Name: "vanillaiice", Email: "vanillaiice1@proton.me"}},
		Usage:   "Backend server for ITPG, handles database transactions and user state management through HTTP(S) requests.",
		Commands: []*cli.Command{
			rootCmd,
			adminCmd,
		},
		Flags:  rootCmd.Flags,
		Action: rootCmd.Action,
	}

	app.Before = altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewTomlSourceFromFlagFunc("load"))

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
