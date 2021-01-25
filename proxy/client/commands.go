package client

import (
	context "context"
	fmt "fmt"
	"log"

	proxy "github.com/omnilaboratory/obd/proxy/pb"
	"github.com/urfave/cli"
	grpc "google.golang.org/grpc"
)


var HelloCommand = cli.Command{
	Name:      "hello",
	Category:  "testing",
	Usage:     "Say Hello to Proxy Mode of OBD",
	ArgsUsage: "your_name",
	Description: "Say Hello to Proxy Mode of OBD",
	Action: hello,
}

func hello(ctx *cli.Context) error {
	
	client, cleanUp := getClient(ctx)
	defer cleanUp()
	
	inputParam := ctx.Args().First()

	var outputInfo string
	switch inputParam {
	case "":
		outputInfo = "You can try to input anything."
		return fmt.Errorf(outputInfo)
	}

	ctxb := context.Background()
	resp, err := client.Hello(ctxb, &proxy.HelloRequest{
		Sayhi: inputParam,
	})

	if err != nil {
		return err
	}

	fmt.Println(resp)
	return nil
}

func getClient(ctx *cli.Context) (proxy.ProxyClient, func()) {

	fmt.Println("HelloProxy client ...")

	opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatal(err)
		// fmt.Printf("unable to connect to RPC server: %v", err)
	}
	// defer cc.Close()

	cleanUp := func() {
		cc.Close()
	}

	return proxy.NewProxyClient(cc), cleanUp
}
