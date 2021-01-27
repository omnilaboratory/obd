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
var kademliaDHT *dht.IpfsDHT
var relayNode string

var localServerDest string
var p2PLocalNodeId string
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
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.P2P_port))

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
	p2PLocalNodeId = hostNode.ID().Pretty()
	service.P2PLocalNodeId = p2PLocalNodeId

	localServerDest = fmt.Sprintf("/ip4/%s/tcp/%v/p2p/%s", config.P2P_hostIp, config.P2P_port, hostNode.ID().Pretty())
	bean.CurrObdNodeInfo.P2pAddress = localServerDest
	log.Println("local p2p address", localServerDest)

	//把自己也作为终点放进去，阻止自己连接自己
	p2pChannelMap[p2PLocalNodeId] = &P2PChannel{
		IsLocalChannel: true,
		Address:        localServerDest,
	}
	hostNode.SetStreamHandler(protocolIdForScanObd, handleTrackerScanStream)
	hostNode.SetStreamHandler(protocolIdForBetweenObd, handleStream)
	hostNode.SetStreamHandler(bean.ProtocolIdForLockChannel, handleLockChannelStream)
	hostNode.SetStreamHandler(bean.ProtocolIdForUnlockChannel, handleUnlockChannelStream)

	kademliaDHT, err = dht.New(ctx, hostNode)
	if err != nil {
		log.Println(err)
		return err
	}

	err = kademliaDHT.Bootstrap(ctx)
	if err != nil {
		log.Println(err)
		return err
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
				if len(relayNode) == 0 {
					relayNode = peerInfo.ID.Pretty()
				}
			}
		}()
	}
	wg.Wait()

	routingDiscovery = discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(ctx, routingDiscovery, obdRendezvousString)

	return nil
}

var routingDiscovery *discovery.RoutingDiscovery
var ctx = context.Background()

func scanAndConnNode(nodeId string) error {
	peerChan, err := routingDiscovery.FindPeers(context.Background(), obdRendezvousString)
	if err != nil {
		return err
	}
	for node := range peerChan {
		if node.ID == hostNode.ID() {
			continue
		}
		if node.ID.Pretty() == nodeId {
			return connSomeNode(node)
		}
	}
	return errors.New("find no node")
}

func connSomeNode(node peer.AddrInfo) error {
	if p2pChannelMap[node.ID.Pretty()] != nil {
		log.Println("Remote peer has been connected")
		return nil
	}

	relayAddr, err := multiaddr.NewMultiaddr("/p2p/" + relayNode + "/p2p-circuit/p2p/" + node.ID.Pretty())
	if err != nil {
		log.Println(err)
		return err
	}
	hostNode.Network().(*swarm.Swarm).Backoff().Clear(node.ID)
	peerRelayInfo := peer.AddrInfo{
		ID:    node.ID,
		Addrs: []multiaddr.Multiaddr{relayAddr},
	}

	if err := hostNode.Connect(ctx, peerRelayInfo); err != nil {
		log.Println(err)
		return err
	} else {
		log.Println("Connection established with RELAY node:", relayAddr)
	}

	hostNode.Peerstore().AddAddrs(node.ID, node.Addrs, peerstore.PermanentAddrTTL)

	stream, err := hostNode.NewStream(ctx, node.ID, protocolIdForBetweenObd)
	if err != nil {
		log.Println(err)
		return err
	}

	rw := addP2PChannel(stream)
	go readData(stream, rw)
	return nil
}

func connP2PNode(dest string) (string, error) {
	if tool.CheckIsString(&dest) == false {
		log.Println("wrong dest address")
		return "", errors.New("wrong dest address")
	}

	if strings.Contains(dest, p2PLocalNodeId) {
		return "", errors.New("do not need connect self")
	}

	_, err := multiaddr.NewMultiaddr(dest)
	if err != nil {
		log.Println(err)
		return "", err
	}

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

	//split := strings.Split(dest, "/")
	//p2PNodeId := split[len(split)-1]
	//findID, err := peer.Decode(p2PNodeId)
	//destHostPeerInfo, err := kademliaDHT.FindPeer(ctx, findID)
	//if err != nil {
	//	return "", err
	//}

	if destHostPeerInfo.ID == hostNode.ID() {
		return "", errors.New("wrong dest address")
	}
	err = connSomeNode(*destHostPeerInfo)
	return localServerDest, err
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

var trackerNodeIdMap = make(map[string]bool)

func handleTrackerScanStream(stream network.Stream) {
	if stream != nil {
		log.Println("request to scan channel and online user info from tracker", stream.Conn().RemotePeer().Pretty())

		users := make(map[string]string)
		for _, item := range globalWsClientManager.OnlineClientMap {
			if item.User != nil {
				users[item.User.PeerId] = item.User.P2PLocalAddress
			}
		}
		flag := false
		info := make(map[string]string)
		info["obdP2pNodeId"] = p2PLocalNodeId
		if len(users) > 0 {
			marshal, _ := json.Marshal(users)
			info["userInfo"] = string(marshal)
			flag = true
		}

		nodes := getChannelInfos()
		if len(nodes) > 0 {
			marshal, _ := json.Marshal(nodes)
			info["channelInfo"] = string(marshal)
			flag = true
		}

		if flag {
			marshal, _ := json.Marshal(info)
			msg := string(marshal) + "~"
			rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
			_, err := rw.WriteString(msg)
			if err == nil {
				_ = rw.Flush()
			}
		}
		trackerNodeIdMap[stream.Conn().RemotePeer().Pretty()] = true
		_ = stream.Close()
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
	if remoteP2PPeerId == p2PLocalNodeId {
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

func sendInfoOnUserStateChange(userId string) {
	for key := range trackerNodeIdMap {
		findID, err := peer.Decode(key)
		if err == nil {
			findPeer, err := kademliaDHT.FindPeer(ctx, findID)
			if err == nil {
				stream, err := hostNode.NewStream(ctx, findPeer.ID, bean.ProtocolIdForUserState)
				if err == nil {
					rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
					_, _ = rw.WriteString(hostNode.ID().Pretty() + "_" + userId + "_" + localServerDest + "~")
					err = rw.Flush()
					log.Println(err)
				}
			}
		}
	}
}

func sendChannelInfoToIndirectTracker(msg string) {
	log.Println("sendChannelInfoToIndirectTracker", msg)
	for key := range trackerNodeIdMap {
		findID, err := peer.Decode(key)
		if err == nil {
			findPeer, err := kademliaDHT.FindPeer(ctx, findID)
			if err == nil {
				stream, err := hostNode.NewStream(ctx, findPeer.ID, bean.ProtocolIdForChannelInfoChange)
				if err == nil {
					rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
					_, _ = rw.WriteString(msg + "~")
					err = rw.Flush()
					log.Println(err)
				}
			}
		}
	}
}

func handleLockChannelStream(stream network.Stream) {
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
		request := &bean.TrackerLockChannelRequest{}
		_ = json.Unmarshal([]byte(str), request)
		err = lockChannel(request.UserId, request.ChannelId)
		result := "1"
		if err != nil {
			result = "0"
		}
		//request.UserId
		_, _ = rw.WriteString(result + "~")
		_ = rw.Flush()
	}
}

func handleUnlockChannelStream(stream network.Stream) {
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
		request := &bean.TrackerLockChannelRequest{}
		_ = json.Unmarshal([]byte(str), request)
		err = unlockChannel(request.UserId, request.ChannelId)
		result := "1"
		if err != nil {
			result = "0"
		}
		//request.UserId
		_, _ = rw.WriteString(result + "~")
		_ = rw.Flush()
	}
}
