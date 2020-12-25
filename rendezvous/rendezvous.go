package main

import (
	"context"
	"flag"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ma "github.com/multiformats/go-multiaddr"
	"log"
	"math/rand"
	"strconv"
)

// gox -os "windows linux darwin" -arch amd64

func main() {
	var port int
	flag.IntVar(&port, "port", 6000, "node listen port")
	flag.Parse()

	ctx := context.Background()

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	sourceMultiAddr, _ := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/" + strconv.Itoa(port))

	r := rand.New(rand.NewSource(int64(port)))
	prvKey, _, err := crypto.GenerateECDSAKeyPair(r)
	if err != nil {
		panic(err)
	}

	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		panic(err)
	}

	_, _ = dht.New(ctx, host, dht.Mode(dht.ModeAutoServer))
	log.Println("This node: ", host.ID().Pretty(), " ", host.Addrs())
	if err != nil {
		panic(err)
	}

	select {}
}
