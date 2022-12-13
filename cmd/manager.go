package cmd

import (
	"github.com/krakenh2020/MPCService/config"
	"github.com/krakenh2020/MPCService/manager"
	"github.com/urfave/cli"
)

var ManagerCmd = cli.Command{
	Name:  "manager",
	Usage: "A manager of MPC nodes",
	Subcommands: cli.Commands{
		cli.Command{
			Name:  "start",
			Usage: "Starts manager of MPC nodes",
			Flags: mangerFlags,
			Action: func(ctx *cli.Context) error {
				manager.RunManager(ctx.Int("guiPort"), ctx.Int("managerPort"), ctx.String("assets"), ctx.String("logLevel"),
					ctx.String("logFile"), ctx.String("certLocation"))
				return nil
			},
		},
	},
}

// mpcNodeFlags are the flags used by the server CLI commands.
var mangerFlags = []cli.Flag{
	// guiPort indicates the port where the server will listen for MPC nodes and data providers.
	&cli.IntFlag{
		Name:  "guiPort",
		Value: config.LoadGuiPort(),
		Usage: "`PORT` where the manager will offer GUI.",
	},
	// managerPort indicates the port where the server will listen for MPC nodes and data providers.
	&cli.IntFlag{
		Name:  "managerPort",
		Value: config.LoadManagerPort(),
		Usage: "`PORT` where the manager will listen for MPC nodes and data providers",
	},
	// logFile indicates the location where the private and public key of the node are saved.
	&cli.StringFlag{
		Name:  "logFile",
		Value: config.LoadLogFile(),
		Usage: "destination of the log file",
	},
	// logLevel indicates the level of how much log is written; possibilities are debug, info, error.
	&cli.StringFlag{
		Name:  "logLevel",
		Value: config.LoadLogLevel(),
		Usage: "Level of how much log is written; possibilities are debug, info, error",
	},
	&cli.StringFlag{
		Name:  "assets",
		Value: config.LoadAssets(),
		Usage: "destination of web assets",
	},

	// certLocation indicates the location where the certificates and keys are saved.
	&cli.StringFlag{
		Name:  "certLocation",
		Value: config.LoadCertLocation(),
		Usage: "location of the certificate",
	},
}
