package obcmds

import (
	"context"
	"github.com/omnilaboratory/obd/obrpc"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"log"
)
func Commands() []*cli.Command {
	return []*cli.Command{
		&createCommand,
		&unlockCommand,
		&changePasswordCommand,
	}
}
func getClient(ctx context.Context) (obrpc.WalletUnlockerClient,func()) {
	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalln(err)
	}
	cleanUp := func() {
		conn.Close()
	}
	return obrpc.NewWalletUnlockerClient(conn),cleanUp
}




