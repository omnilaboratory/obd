package lightclient

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/multiformats/go-multiaddr"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"log"
	"math/rand"
	"strings"
	"sync"
)

type P2PChannel struct {
	IsLocalChannel bool
	Address        string
	stream         network.Stream
	rw             *bufio.ReadWriter
}

const obdRendezvousString = "obd meet at tracker"
const protocolIdForBetweenObd = "obd/betweenObd/1.0.1"
const protocolIdForScanObd = "obd/forScanObd/1.0.1"

var hostNode host.Host

var localServerDest string
var p2PLocalPeerId string
var privateKey crypto.PrivKey
var p2pChannelMap map[string]*P2PChannel

func generatePrivateKey() (crypto.PrivKey, error) {
	if privateKey == nil {
		nodeId := int64(binary.BigEndian.Uint64([]byte(tool.GetObdNodeId())))
		r := rand.New(rand.NewSource(nodeId))
		prvKey, _, err := crypto.GenerateECDSAKeyPair(r)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		privateKey = prvKey
	}
	return privateKey, nil
}

func StartP2PNode() (err error) {

	log.Println("start to p2p node")
	prvKey, err := generatePrivateKey()
	if err != nil {
		return err
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.P2P_sourcePort))

	ctx := context.Background()
	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	hostNode, err = libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.EnableRelay(),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		log.Println(err)
		return err
	}
	p2pChannelMap = make(map[string]*P2PChannel)
	p2PLocalPeerId = hostNode.ID().Pretty()
	service.P2PLocalPeerId = p2PLocalPeerId

	localServerDest = fmt.Sprintf("/ip4/%s/tcp/%v/p2p/%s", config.P2P_hostIp, config.P2P_sourcePort, hostNode.ID().Pretty())
	bean.CurrObdNodeInfo.P2pAddress = localServerDest
	log.Println("local p2p address", localServerDest)

	//把自己也作为终点放进去，阻止自己连接自己
	p2pChannelMap[p2PLocalPeerId] = &P2PChannel{
		IsLocalChannel: true,
		Address:        localServerDest,
	}
	hostNode.SetStreamHandler(protocolIdForScanObd, handleScanStream)
	hostNode.SetStreamHandler(protocolIdForBetweenObd, handleStream)

	kademliaDHT, _ := dht.New(ctx, hostNode, dht.Mode(dht.ModeAutoServer))
	if err != nil {
		log.Println(err)
		return err
	}

	err = kademliaDHT.Bootstrap(ctx)
	if err != nil {
		log.Println(err)
	}

	var wg sync.WaitGroup
	for _, peerAddr := range config.BootstrapPeers {
		peerInfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = hostNode.Connect(ctx, *peerInfo)
			if err != nil {
				log.Println(err, peerInfo)
			} else {
				log.Println("connected to bootstrap node ", *peerInfo)
			}
		}()
	}
	wg.Wait()

	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(ctx, routingDiscovery, obdRendezvousString)

	return nil
}

func connP2PNode(dest string) (string, error) {
	if tool.CheckIsString(&dest) == false {
		log.Println("wrong dest address")
		return "", errors.New("wrong dest address")
	}
	ctx := context.Background()

	destMaddr, err := multiaddr.NewMultiaddr(dest)
	if err != nil {
		log.Println(err)
		return "", err
	}

	destHostPeerInfo, err := peer.AddrInfoFromP2pAddr(destMaddr)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if destHostPeerInfo.ID == hostNode.ID() {
		return "", errors.New("wrong dest address")
	}

	if p2pChannelMap[destHostPeerInfo.ID.Pretty()] != nil {
		log.Println("Remote peer has been connected")
		return " Remote peer has been connected", nil
	}

	relayNode := "QmWp5amgedYDc1Feay1sn2Q1dWvePtKAzwCnzCattH2xLR"
	relayaddr, err := multiaddr.NewMultiaddr("/p2p/" + relayNode + "/p2p-circuit/p2p/" + destHostPeerInfo.ID.Pretty())
	if err != nil {
		log.Println(err)
		return "", err
	}
	hostNode.Network().(*swarm.Swarm).Backoff().Clear(destHostPeerInfo.ID)
	peerRelayInfo := peer.AddrInfo{
		ID:    destHostPeerInfo.ID,
		Addrs: []multiaddr.Multiaddr{relayaddr},
	}

	if err := hostNode.Connect(ctx, peerRelayInfo); err != nil {
		log.Println(err)
		return "", err
	} else {
		log.Println("Connection established with RELAY node:", relayaddr)
	}

	hostNode.Peerstore().AddAddrs(destHostPeerInfo.ID, destHostPeerInfo.Addrs, peerstore.PermanentAddrTTL)

	stream, err := hostNode.NewStream(context.Background(), destHostPeerInfo.ID, protocolIdForBetweenObd)
	if err != nil {
		log.Println(err)
		return "", err
	}

	rw := addP2PChannel(stream)
	go readData(stream, rw)
	return localServerDest, nil
}

func handleStream(s network.Stream) {
	if s != nil {
		if p2pChannelMap[s.Conn().RemotePeer().Pretty()] == nil {
			log.Println("Got a new stream!")
			rw := addP2PChannel(s)
			go readData(s, rw)
		}
	}
}

func handleScanStream(s network.Stream) {
	if s != nil {
		log.Println("scan channel data request from tracker", s.Conn().RemotePeer().Pretty())
	}
}
func readData(s network.Stream, rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('~')
		if err != nil {
			delete(p2pChannelMap, s.Conn().RemotePeer().Pretty())
			return
		}
		if str == "" {
			return
		}
		if str != "" {
			log.Println(s.Conn())
			str = strings.TrimSuffix(str, "~")
			reqData := &bean.RequestMessage{}
			err := json.Unmarshal([]byte(str), reqData)
			if err == nil {
				err = getDataFromP2PSomeone(*reqData)
				if err != nil {
					//msg := err.Error() + "~"
					//_, _ = rw.WriteString(msg)
					//_ = rw.Flush()
				}
			}
		}
	}
}

func addP2PChannel(stream network.Stream) *bufio.ReadWriter {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	node := &P2PChannel{
		Address: stream.Conn().RemoteMultiaddr().String() + "/" + stream.Conn().RemotePeer().String(),
		stream:  stream,
		rw:      rw,
	}
	p2pChannelMap[stream.Conn().RemotePeer().Pretty()] = node
	return rw
}

func sendP2PMsg(remoteP2PPeerId string, msg string) error {
	if tool.CheckIsString(&remoteP2PPeerId) == false {
		return errors.New("empty remoteP2PPeerId")
	}
	if remoteP2PPeerId == p2PLocalPeerId {
		return errors.New("remoteP2PPeerId is yourself,can not send msg to yourself")
	}

	channel := p2pChannelMap[remoteP2PPeerId]
	if channel != nil {
		msg = msg + "~"
		_, _ = channel.rw.WriteString(msg)
		_ = channel.rw.Flush()
	}
	return nil
}
