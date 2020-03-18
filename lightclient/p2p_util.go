package lightclient

import (
	"LightningOnOmni/config"
	"LightningOnOmni/tool"
	"bufio"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"log"
	mrand "math/rand"
	"strings"

	"github.com/multiformats/go-multiaddr"
)

type P2PChannel struct {
	stream network.Stream
	rw     *bufio.ReadWriter
}

const pid = "/chat/1.0.0"

var localServerDest string
var P2PLocalPeerId string
var privateKey crypto.PrivKey
var p2pChannelMap map[string]*P2PChannel

func generatePrivateKey() (crypto.PrivKey, error) {
	if privateKey == nil {
		r := mrand.New(mrand.NewSource(int64(config.P2P_sourcePort)))
		// Creates a new RSA key pair for this host.
		prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		privateKey = prvKey
	}
	return privateKey, nil
}

func StartP2PServer() {
	prvKey, _ := generatePrivateKey()
	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.P2P_sourcePort))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		log.Println(err)
		return
	}
	p2pChannelMap = make(map[string]*P2PChannel)
	P2PLocalPeerId = host.ID().Pretty()

	host.SetStreamHandler(pid, handleStream)

	localServerDest = fmt.Sprintf("/ip4/127.0.0.1/tcp/%v/p2p/%s", config.P2P_sourcePort, host.ID().Pretty())
	log.Println(localServerDest)
}

func ConnP2PServer(dest string) {
	if tool.CheckIsString(&dest) == false {
		log.Println("wrong dest address")
		return
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.P2P_sourcePort))
	prvKey, _ := generatePrivateKey()

	host, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)

	if err != nil {
		log.Println(err)
		return
	}
	maddr, err := multiaddr.NewMultiaddr(dest)
	if err != nil {
		log.Fatalln(err)
	}
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Fatalln(err)
	}
	if p2pChannelMap[info.ID.Pretty()] != nil {
		log.Println("p2p channel has connect")
		return
	}
	host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	s, err := host.NewStream(context.Background(), info.ID, pid)
	if err != nil {
		log.Println(err)
		return
	}
	rw := addP2PChannel(s)
	go readData(s, rw)
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
func readData(s network.Stream, rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('$')
		if err != nil {
			log.Println(s.Conn().RemotePeer(), err)
			delete(p2pChannelMap, s.Conn().RemotePeer().Pretty())
			return
		}
		if str == "" {
			return
		}
		if str != "\n" {
			log.Println(s.Conn())
			str = strings.TrimSuffix(str, "$")
			GlobalWsClientManager.P2PData <- []byte(str)
		}
	}
}

func addP2PChannel(stream network.Stream) *bufio.ReadWriter {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	node := &P2PChannel{
		stream: stream,
		rw:     rw,
	}
	p2pChannelMap[stream.Conn().RemotePeer().Pretty()] = node
	return rw
}

func SendP2PMsg(remoteP2PPeerId string, msg string) {
	if tool.CheckIsString(&remoteP2PPeerId) == false {
		log.Println("empty remoteP2PPeerId")
		return
	}
	if remoteP2PPeerId == P2PLocalPeerId {
		log.Println("remoteP2PPeerId is self,can not send msg to yourself")
		return
	}

	channel := p2pChannelMap[remoteP2PPeerId]
	if channel != nil {
		_, _ = channel.rw.WriteString(msg)
		_ = channel.rw.Flush()
	}
}
