package cmd

import (
	"strings"

	"github.com/krakenh2020/MPCService/config"
	"github.com/krakenh2020/MPCService/data_provider"
	"github.com/urfave/cli"
)

var DataProviderCMD = cli.Command{
	Name:  "data_provider",
	Usage: "A provider of data for MPC",
	Subcommands: cli.Commands{
		cli.Command{
			Name:  "start",
			Usage: "Starts data a data provider",
			Flags: dataProviderFlags,
			Action: func(ctx *cli.Context) error {
				data_provider.RunDatasetProvider(ctx.String("name"),
					ctx.String("dataLoc"),
					ctx.String("logLevel"),
					ctx.String("logFile"),
					ctx.String("manAddr"),
					ctx.String("certLocation"),
					strings.Split(ctx.String("shareWith"), ","),
				)
				return nil
			},
		},
	},
}

// dataProviderFlags are the flags used by the server CLI commands.
var dataProviderFlags = []cli.Flag{
	// portFlag indicates the port where the server will listen.
	&cli.StringFlag{
		Name:  "name",
		Value: config.LoadServerName(),
		Usage: "name of the server",
	},
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
		Name:  "dataLoc",
		Value: config.LoadDataLoc(),
		Usage: "Location of datasets",
	},
	// certLocation indicates the location where the certificates and keys are saved.
	&cli.StringFlag{
		Name:  "certLocation",
		Value: config.LoadCertLocation(),
		Usage: "location of the certificate",
	},
	&cli.StringFlag{
		Name:  "shareWith",
		Value: config.LoadShareWith(),
		Usage: "location of the certificate",
	},
}
