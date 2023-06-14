package main

import (
	"fmt"
	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lightningnetwork/lnd/build"
	"github.com/lightningnetwork/lnd/lnrpc/lnclipb"
	"github.com/lightningnetwork/lnd/lnrpc/verrpc"
	"github.com/urfave/cli"
)

var versionCommand = cli.Command{
	Name:  "version",
	Usage: "Display lncli and lnd version info.",
	Description: `
	Returns version information about both lncli and lnd. If lncli is unable
	to connect to lnd, the command fails but still prints the lncli version.
	`,
	Action: actionDecorator(version),
}

func version(ctx *cli.Context) error {
	ctxc := getContext()
	conn := getClientConn(ctx, false)
	defer conn.Close()

	versions := &lnclipb.VersionResponse{
		Lncli: &verrpc.Version{
			Commit:        build.Commit,
			CommitHash:    build.CommitHash,
			Version:       build.Version(),
			AppMajor:      uint32(build.AppMajor),
			AppMinor:      uint32(build.AppMinor),
			AppPatch:      uint32(build.AppPatch),
			AppPreRelease: build.AppPreRelease,
			BuildTags:     build.Tags(),
			GoVersion:     build.GoVersion,
		},
	}

	client := verrpc.NewVersionerClient(conn)

	lndVersion, err := client.GetVersion(ctxc, &verrpc.VersionRequest{})
	if err != nil {
		printRespJSON(versions)
		return fmt.Errorf("unable fetch version from lnd: %v", err)
	}
	versions.Lnd = lndVersion

	printRespJSON(versions)

	return nil
}

var listAssetCommand = cli.Command{
	Name:        "listasset",
	Usage:       "list all asset on omnicore-server",
	Description: ``,
	Action:      actionDecorator(listasset),
}

func listasset(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	items, err := client.OB_ListAsset(ctxc, &lnrpc.ListAssetRequest{})
	if err != nil {
		return fmt.Errorf("unable fetch version from lnd: %v", err)
	}
	printRespJSON(items)
	return nil
}

var dumpPrivkeyCommand = cli.Command{
	Name:        "dumpkey",
	Usage:       "dumpkey",
	Description: `Reveals the private key corresponding to ‘address’.`,
	Action:      actionDecorator(dumpPrivkey),
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Usage: "",
		},
	},
}

func dumpPrivkey(ctx *cli.Context) error {
	address := ctx.String("address")
	ctxc := getContext()
	client, cleanUp := getClient(ctx)
	defer cleanUp()
	res, err := client.OB_DumpPrivkey(ctxc, &lnrpc.DumpPrivkeyRequest{Address: address})
	if err != nil {
		return fmt.Errorf("unable DumpPrivkey from address: %v %v", address, err)
	}
	printRespJSON(res)
	return nil
}

var listAddressCommand = cli.Command{
	Name:        "listaddress",
	Usage:       "list all pubkey address on omni-wallet",
	Description: ``,
	Action:      actionDecorator(listaddress),
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "type",
			Usage: "p2wkh p2tr",
			Value: "p2kh",
		},
	},
}

func listaddress(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	typeStr := ctx.String("type")
	atype := lnrpc.AddressType_PUBKEY
	switch typeStr {
	case "p2wkh":
		atype = lnrpc.AddressType_NFT_WITNESS_PUBKEY_HASH
	case "p2tr":
		atype = lnrpc.AddressType_TAPROOT_PUBKEY
	}
	items, err := client.OB_ListAddresses(ctxc, &lnrpc.ListAddressesRequest{AddressType: atype})
	if err != nil {
		return fmt.Errorf("unable ListAddresses from lnd: %v", err)
	}
	printRespJSON(items)
	return nil
}

var listRecAddressCommand = cli.Command{
	Name:        "listrecaddress",
	Usage:       "list all pubkey address which create by user-self ",
	Description: ``,
	Action:      actionDecorator(listRecAddress),
}

func listRecAddress(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	items, err := client.OB_ListRecAddress(ctxc, &lnrpc.ListRecAddressRequest{})
	if err != nil {
		return fmt.Errorf("unable listRecAddress from lnd: %v", err)
	}
	printRespJSON(items)
	return nil
}

var setDefaultAddressCommand = cli.Command{
	Name:        "setdefaultaddress",
	Usage:       "set the pubkey address as default",
	Description: ``,
	Action:      actionDecorator(setDefaultAddress),
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Usage: "the address  will be set",
		},
	},
}

func setDefaultAddress(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getClient(ctx)
	defer cleanUp()
	address := ctx.String("address")
	_, err := client.OB_SetDefaultAddress(ctxc, &lnrpc.SetDefaultAddressRequest{Address: address})
	if err != nil {
		return fmt.Errorf("unable setDefaultAddress : %v", err)
	}
	fmt.Println("setDefaultAddress ok")
	//printRespJSON(items)
	return nil
}
