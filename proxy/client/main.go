package client

import (
	context "context"
	fmt "fmt"

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
	
	client := getClient(ctx)
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



// func main() {
// 	app := cli.NewApp()
// 	app.Name = "obdcli"
// 	app.Version = "0.0.1-beta"
// 	app.Usage = "Control plane for your Omni Bolt Daemon (obd)"
// 	app.Commands = []cli.Command{
// 		helloCommand,
// 		startCommand,
// 	}

// 	if err := app.Run(os.Args); err != nil {
// 		log.Fatal(err)
// 	}
// }

func getClient(ctx *cli.Context) (proxy.ProxyClient) {

	fmt.Println("HelloProxy client ...")

	opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		// log.Fatal(err)
		fmt.Printf("unable to connect to RPC server: %v", err)
	}
	defer cc.Close()

	return proxy.NewProxyClient(cc)
}
