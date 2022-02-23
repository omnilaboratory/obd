package main

import (
	"github.com/omnilaboratory/obd/cmd/obcmds"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "rpcserver",
			Value: defaultRPCHostPort,
			Usage: "The host:port of LN daemon.",
		},
		&cli.StringFlag{
			Name:  "lnddir",
			Value: defaultLndDir,
			Usage: "The path to lnd's base directory.",
		},
		&cli.StringFlag{
			Name:  "tlscertpath",
			Value: defaultTLSCertPath,
			Usage: "The path to lnd's TLS certificate.",
		},
		&cli.StringFlag{
			Name:  "chain, c",
			Usage: "The chain lnd is running on, e.g. bitcoin.",
			Value: "bitcoin",
		},
		&cli.StringFlag{
			Name: "network, n",
			Usage: "The network lnd is running on, e.g. mainnet, " +
				"testnet, etc.",
			Value: "mainnet",
		},
		&cli.BoolFlag{
			Name:  "no-macaroons",
			Usage: "Disable macaroon authentication.",
		},
		&cli.StringFlag{
			Name:  "macaroonpath",
			Usage: "The path to macaroon file.",
		},
		&cli.Int64Flag{
			Name:  "macaroontimeout",
			Value: 60,
			Usage: "Anti-replay macaroon validity time in seconds.",
		},
		&cli.StringFlag{
			Name:  "macaroonip",
			Usage: "If set, lock macaroon to specific IP address.",
		},
		&cli.StringFlag{
			Name: "profile, p",
			Usage: "Instead of reading settings from command " +
				"line parameters or using the default " +
				"profile, use a specific profile. If " +
				"a default profile is set, this flag can be " +
				"set to an empty string to disable reading " +
				"values from the profiles file.",
		},
		&cli.StringFlag{
			Name: "macfromjar",
			Usage: "Use this macaroon from the profile's " +
				"macaroon jar instead of the default one. " +
				"Can only be used if profiles are defined.",
		},
	}

	btcSubcmd := cli.Command{
		Name:    "btc",
		Aliases: []string{"b"},
		Usage:   "btc cli wallet",
		//Action: func(c *cli.Context) error {
		//	log.Println("btc command execute:", c.Args())
		//	return nil
		//},
	}
	btcSubcmd.Subcommands = []*cli.Command{
		&createCommand,
		&unlockCommand,
		&createWatchOnlyCommand,
		&changePasswordCommand,
		&newAddressCommand,
		&estimateFeeCommand,
		&sendManyCommand,
		&sendCoinsCommand,
		&listUnspentCommand,
		&connectCommand,
		&disconnectCommand,
		&openChannelCommand,
		&batchOpenChannelCommand,
		&closeChannelCommand,
		&closeAllChannelsCommand,
		&abandonChannelCommand,
		&listPeersCommand,
		&walletBalanceCommand,
		&channelBalanceCommand,
		&getInfoCommand,
		&getRecoveryInfoCommand,
		&pendingChannelsCommand,
		&sendPaymentCommand,
		&payInvoiceCommand,
		&sendToRouteCommand,
		&addInvoiceCommand,
		&lookupInvoiceCommand,
		&listInvoicesCommand,
		&listChannelsCommand,
		&closedChannelsCommand,
		&listPaymentsCommand,
		&describeGraphCommand,
		&getNodeMetricsCommand,
		&getChanInfoCommand,
		&getNodeInfoCommand,
		&queryRoutesCommand,
		&getNetworkInfoCommand,
		&debugLevelCommand,
		&decodePayReqCommand,
		&listChainTxnsCommand,
		&stopCommand,
		&signMessageCommand,
		&verifyMessageCommand,
		&feeReportCommand,
		&updateChannelPolicyCommand,
		&forwardingHistoryCommand,
		&exportChanBackupCommand,
		&verifyChanBackupCommand,
		&restoreChanBackupCommand,
		&bakeMacaroonCommand,
		&listMacaroonIDsCommand,
		&deleteMacaroonIDCommand,
		&listPermissionsCommand,
		&printMacaroonCommand,
		&trackPaymentCommand,
		&versionCommand,
		&profileSubCommand,
		&getStateCommand,
		&deletePaymentsCommand,
		&sendCustomCommand,
		&subscribeCustomCommand,
	}

	// Add any extra commands determined by build flags.
	btcSubcmd.Subcommands = append(btcSubcmd.Subcommands, autopilotCommands()...)
	btcSubcmd.Subcommands = append(btcSubcmd.Subcommands, invoicesCommands()...)
	btcSubcmd.Subcommands = append(btcSubcmd.Subcommands, routerCommands()...)
	btcSubcmd.Subcommands = append(btcSubcmd.Subcommands, walletCommands()...)
	btcSubcmd.Subcommands = append(btcSubcmd.Subcommands, watchtowerCommands()...)
	btcSubcmd.Subcommands = append(btcSubcmd.Subcommands, wtclientCommands()...)

	usdtSubcmd := cli.Command{
		Name:    "usdt",
		Aliases: []string{"u"},
		Usage:   "usdt cli wallet",
		Action: func(c *cli.Context) error {
			log.Println("usdt command execute")
			return nil
		},
	}
	usdtSubcmd.Subcommands=obcmds.Commands()

	app.Commands = []*cli.Command{
		&btcSubcmd,
		&usdtSubcmd,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}