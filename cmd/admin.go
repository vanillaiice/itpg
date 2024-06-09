package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/xyproto/permissionbolt/v2"
	"github.com/xyproto/pinterface"
)

var userState pinterface.IUserState

var addAdminCmd = &cli.Command{
	Name:    "add",
	Aliases: []string{"a"},
	Usage:   "add admin",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "username",
			Aliases: []string{"U"},
			Usage:   "new admin username",
		},
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"P"},
			Usage:   "new admin password",
		},
		&cli.StringFlag{
			Name:    "email",
			Aliases: []string{"e"},
			Usage:   "new admin email",
		},
		&cli.BoolFlag{
			Name:    "super",
			Aliases: []string{"s"},
			Usage:   "make admin super",
		},
	},
	Action: func(ctx *cli.Context) error {
		if !userState.CorrectPassword(ctx.String("login-username"), ctx.String("login-password")) {
			return fmt.Errorf("wrong login credentials")
		}

		if !userState.IsAdmin(ctx.String("login-username")) {
			return fmt.Errorf("not an admin")
		}

		if !userState.BooleanField(ctx.String("login-username"), "super") {
			return fmt.Errorf("not a super admin")
		}

		if userState.HasUser(ctx.String("username")) {
			return fmt.Errorf("user %s already exists", ctx.String("username"))
		}

		if err := permissionbolt.ValidUsernamePassword(ctx.String("username"), ctx.String("password")); err != nil {
			return err
		}

		userState.AddUser(ctx.String("username"), ctx.String("password"), ctx.String("email"))

		userState.SetAdminStatus(ctx.String("username"))

		if ctx.Bool("super") {
			userState.SetBooleanField(ctx.String("username"), "super", true)
		}

		if ctx.Bool("verbose") {
			if ctx.Bool("super") {
				fmt.Printf("added super admin %s\n", ctx.String("username"))
			} else {
				fmt.Printf("added admin %s\n", ctx.String("username"))
			}
		}

		return nil
	},
}

var rmAdminCmd = &cli.Command{
	Name:    "remove",
	Aliases: []string{"r"},
	Usage:   "remove admin",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "username",
			Aliases: []string{"U"},
			Usage:   "admin username to remove",
		},
	},
	Action: func(ctx *cli.Context) error {
		if !userState.CorrectPassword(ctx.String("login-username"), ctx.String("login-password")) {
			return fmt.Errorf("wrong login credentials")
		}

		if !userState.IsAdmin(ctx.String("login-username")) {
			return fmt.Errorf("not an admin")
		}

		if !userState.BooleanField(ctx.String("login-username"), "super") {
			return fmt.Errorf("not a super admin")
		}

		if !userState.HasUser(ctx.String("username")) {
			return fmt.Errorf("user %s does not exist", ctx.String("username"))
		}

		userState.RemoveUser(ctx.String("username"))

		if ctx.Bool("verbose") {
			fmt.Printf("removed admin %s\n", ctx.String("username"))
		}

		return nil
	},
}

var rmSuperAdminStatusCmd = &cli.Command{
	Name:    "rm-super-status",
	Aliases: []string{"s"},
	Usage:   "remove super admin status",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "username",
			Aliases: []string{"U"},
			Usage:   "admin username to remove",
		},
	},
	Action: func(ctx *cli.Context) error {
		if !userState.CorrectPassword(ctx.String("login-username"), ctx.String("login-password")) {
			return fmt.Errorf("wrong login credentials")
		}

		if !userState.IsAdmin(ctx.String("login-username")) {
			return fmt.Errorf("not an admin")
		}

		if !userState.BooleanField(ctx.String("login-username"), "super") {
			return fmt.Errorf("not a super admin")
		}

		if !userState.HasUser(ctx.String("username")) {
			return fmt.Errorf("user %s does not exist", ctx.String("username"))
		}

		if !userState.IsAdmin(ctx.String("username")) {
			return fmt.Errorf("user %s is not an admin", ctx.String("username"))
		}

		userState.SetBooleanField(ctx.String("username"), "super", false)

		if ctx.Bool("verbose") {
			fmt.Printf("removed super admin status for %s\n", ctx.String("username"))
		}

		return nil
	},
}

var changeAdminPasswordCmd = &cli.Command{
	Name:    "change-password",
	Aliases: []string{"c"},
	Usage:   "change admin password",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"P"},
			Usage:   "admin password",
		},
	},
	Action: func(ctx *cli.Context) error {
		if !userState.CorrectPassword(ctx.String("login-username"), ctx.String("login-password")) {
			return fmt.Errorf("wrong login credentials")
		}

		if !userState.IsAdmin(ctx.String("login-username")) {
			return fmt.Errorf("not an admin")
		}

		userState.SetPassword(ctx.String("login-username"), ctx.String("password"))

		if ctx.Bool("verbose") {
			fmt.Printf("changed admin password for %s\n", ctx.String("login-username"))
		}

		return nil
	},
}

var adminCmd = &cli.Command{
	Name:    "admin",
	Aliases: []string{"a"},
	Usage:   "admin management",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "db",
			Aliases: []string{"d"},
			Usage:   "users database connection `URL`",
			Value:   "users.db",
		},
		&cli.StringFlag{
			Name:    "login-username",
			Aliases: []string{"u"},
			Usage:   "admin login username",
		},
		&cli.StringFlag{
			Name:    "login-password",
			Aliases: []string{"p"},
			Usage:   "admin login password",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "verbose output",
		},
	},
	Subcommands: []*cli.Command{
		addAdminCmd,
		rmSuperAdminStatusCmd,
		rmAdminCmd,
		changeAdminPasswordCmd,
	},
	Before: func(ctx *cli.Context) error {
		perm, err := permissionbolt.NewWithConf(ctx.String("db"))
		if err != nil {
			return err
		}

		userState = perm.UserState()

		return nil
	},
}
