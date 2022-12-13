package cmd

import (
	"github.com/krakenh2020/MPCService/config"
	"github.com/krakenh2020/MPCService/mpc_node"
	"github.com/urfave/cli"
)

var MpcNodeCmd = cli.Command{
	Name:  "mpc_node",
	Usage: "An MPC node listening for requests",
	Subcommands: cli.Commands{
		cli.Command{
			Name:  "start",
			Usage: "Starts server",
			Flags: mpcNodeFlags,
			Action: func(ctx *cli.Context) error {
				mpc_node.RunNode(ctx.String("name"),
					ctx.String("nodeAddr"),
					ctx.Int("scalePort"),
					ctx.String("certLocation"),
					ctx.String("sm"),
					ctx.String("logLevel"),
					ctx.String("logFile"),
					ctx.String("manAddr"),
					ctx.String("description"))
				return nil
			},
		},
	},
}

// mpcNodeFlags are the flags used by the server CLI commands.
var mpcNodeFlags = []cli.Flag{
	// portFlag indicates the port where the server will listen.
	&cli.StringFlag{
		Name:  "name",
		Value: config.LoadServerName(),
		Usage: "name of the server",
	},
	&cli.StringFlag{
		Name:  "nodeAddr",
		Value: config.LoadNodeAddr(),
		Usage: "address of the server",
	},
	&cli.IntFlag{
		Name:  "scalePort",
		Value: config.LoadScalePort(),
		Usage: "port on which MPC protocol will be communication with other nodes",
	},
	// certLocation indicates the location where the certificates and keys are saved.
	&cli.StringFlag{
		Name:  "certLocation",
		Value: config.LoadCertLocation(),
		Usage: "location of the certificate",
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
		Name:  "manAddr",
		Value: config.LoadManAddr(),
		Usage: "Address on which node manager is running.",
	},
	&cli.StringFlag{
		Name:  "sm",
		Value: config.LoadScaleMamba(),
		Usage: "Location of SCALE-MAMBA",
	},
	&cli.StringFlag{
		Name:  "description",
		Value: config.LoadDescription(),
		Usage: "Description of the MPC node",
	},
}
