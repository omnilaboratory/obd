package service

import (
	"context"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ma "github.com/multiformats/go-multiaddr"
	cfg "github.com/omnilaboratory/obd/tracker/config"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var protocolID = "tracker/1.0.1"
var rendezvousString = "tracker/1.0.1"
var ctx = context.Background()
var routingDiscovery *discovery.RoutingDiscovery
var hostNode host.Host

func handleStream(stream network.Stream) {
	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	//rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
}
func StartP2PNode() {

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	sourceMultiAddr, _ := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/" + strconv.Itoa(cfg.P2P_sourcePort))

	r := rand.New(rand.NewSource(int64(cfg.P2P_sourcePort)))
	prvKey, _, err := crypto.GenerateECDSAKeyPair(r)
	if err != nil {
		panic(err)
	}

	hostNode, err = libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
		libp2p.EnableRelay(),
	)
	if err != nil {
		panic(err)
	}
	log.Println("This node: ", hostNode.ID().Pretty(), " ", hostNode.Addrs())

	kademliaDHT, _ := dht.New(ctx, hostNode, dht.Mode(dht.ModeAutoServer))

	if err != nil {
		panic(err)
	}

	err = kademliaDHT.Bootstrap(ctx)
	if err != nil {
		log.Println(err)
	}

	if len(cfg.BootstrapPeers) > 0 {
		var wg sync.WaitGroup
		for _, peerAddr := range cfg.BootstrapPeers {
			log.Println("peerAddr is: ", peerAddr)
			peerInfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
			wg.Add(1)
			go func() {
				defer wg.Done()
				log.Println("链接bootstrap节点", *peerInfo)
				err = hostNode.Connect(ctx, *peerInfo)
				if err != nil {
					log.Println(err, peerInfo)
				} else {
					log.Println("链接到了bootstrap节点:", *peerInfo)
				}
			}()
		}
		wg.Wait()
	}

	routingDiscovery = discovery.NewRoutingDiscovery(kademliaDHT)
	startSchedule()
}

func startSchedule() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				log.Println("timer 1m", t)
				scanNodes()
			}
		}
	}()
}

func scanNodes() {
	peerChan, err := routingDiscovery.FindPeers(ctx, rendezvousString)
	if err != nil {
		panic(err)
	}
	for peer := range peerChan {
		if peer.ID == hostNode.ID() {
			continue
		}

		// 找到peer，然后将开始建立通信通道
		log.Println("开始建立连接", peer.ID, peer.Addrs)
		stream, err := hostNode.NewStream(ctx, peer.ID, protocol.ID(protocolID))

		// 如果建立成功
		if err == nil {
			stream.Write([]byte("hello peer,ping"))
			handleStream(stream)
			log.Println("连接成功 ", peer.ID)
		} else {
			log.Println(err)
		}
	}
}
