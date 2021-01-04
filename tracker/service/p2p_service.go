package service

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	ma "github.com/multiformats/go-multiaddr"
	cfg "github.com/omnilaboratory/obd/tracker/config"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const protocolIdForScanObd = "obd/forScanObd/1.0.1"
const obdRendezvousString = "obd meet at tracker"
const trackerRendezvousString = "tracker meet here"

var ctx = context.Background()
var routingDiscovery *discovery.RoutingDiscovery
var hostNode host.Host

func StartP2PNode() {

	nodeId := int64(binary.BigEndian.Uint64([]byte(GetTrackerNodeId())))
	r := rand.New(rand.NewSource(nodeId))
	prvKey, _, err := crypto.GenerateECDSAKeyPair(r)
	if err != nil {
		panic(err)
	}

	sourceMultiAddr, _ := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/" + strconv.Itoa(cfg.P2P_sourcePort))
	hostNode, err = libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
		libp2p.EnableRelay(circuit.OptHop),
	)
	if err != nil {
		panic(err)
	}
	cfg.P2pLocalAddress = fmt.Sprintf("/ip4/%s/tcp/%v/p2p/%s", cfg.P2P_hostIp, cfg.P2P_sourcePort, hostNode.ID().Pretty())
	log.Println("local p2p node address: ", cfg.P2pLocalAddress)

	kademliaDHT, _ := dht.New(ctx, hostNode, dht.Mode(dht.ModeServer))

	if err != nil {
		panic(err)
	}

	err = kademliaDHT.Bootstrap(ctx)
	if err != nil {
		log.Println(err)
	}

	needAnnounceSelf := false
	if len(cfg.BootstrapPeers) > 0 {
		var wg sync.WaitGroup
		for _, peerAddr := range cfg.BootstrapPeers {
			peerInfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
			if peerInfo.ID == hostNode.ID() {
				continue
			}
			log.Println("connecting to bootstrap node: ", peerAddr)
			wg.Add(1)
			needAnnounceSelf = true
			go func() {
				defer wg.Done()
				err = hostNode.Connect(ctx, *peerInfo)
				if err != nil {
					log.Println(err, peerInfo)
				} else {
					log.Println("connect to bootstrap node ", *peerInfo)
				}
			}()
		}
		wg.Wait()
	}
	routingDiscovery = discovery.NewRoutingDiscovery(kademliaDHT)
	if needAnnounceSelf {
		log.Println("announce self")
		discovery.Advertise(ctx, routingDiscovery, trackerRendezvousString)
	}
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
	peerChan, err := routingDiscovery.FindPeers(ctx, obdRendezvousString)
	if err != nil {
		panic(err)
	}
	for peer := range peerChan {
		if peer.ID == hostNode.ID() {
			continue
		}

		//和tracker直接连接的obd，不需要同步数据
		if obdOnlineNodesMap[peer.ID.Pretty()] != nil {
			continue
		}

		log.Println("begin to connect ", peer.ID, peer.Addrs)
		err = hostNode.Connect(ctx, peer)
		if err == nil {
			stream, err := hostNode.NewStream(ctx, peer.ID, protocolIdForScanObd)
			if err == nil {
				go handleStream(stream)
			}
		} else {
			delete(userOnlineOfOtherObdMap, peer.ID.Pretty())
		}
	}
}

var userOnlineOfOtherObdMap = make(map[string]string)

func handleStream(stream network.Stream) {
	log.Println("begin scan obd channel")
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	str, err := rw.ReadString('~')
	if err != nil {
		return
	}

	if str == "" {
		return
	}
	if str != "" {
		str = strings.TrimSuffix(str, "~")
		data := make(map[string]string)
		err = json.Unmarshal([]byte(str), &data)
		if err == nil {
			if _, ok := data["userInfo"]; ok == true {
				log.Println(data["userInfo"])
				userOnlineOfOtherObdMap[stream.Conn().RemotePeer().Pretty()] = data["userInfo"]
			}
			//channel
			if _, ok := data["channelInfo"]; ok == true {
				log.Println(data["channelInfo"])
				_ = ChannelService.updateChannelInfo(data["obdP2pNodeId"], data["channelInfo"])
			}
		}
	}
	_ = stream.Close()
}
