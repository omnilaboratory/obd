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

const pid = "/chat/1.0.0"

var serverReaderWriters map[string]*bufio.ReadWriter

func handleStream(s network.Stream) {
	if s != nil {
		if serverReaderWriters[s.Conn().RemotePeer().Pretty()] == nil {
			log.Println("Got a new stream!")
			rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
			serverReaderWriters[s.Conn().RemotePeer().Pretty()] = rw
			go readData(s, rw)
		}
	}
	// stream 's' will stay open until you close it (or the other side closes it).
}
func readData(s network.Stream, rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('$')
		if err != nil {
			log.Println(s.Conn().RemotePeer(), err)
			delete(serverReaderWriters, s.Conn().RemotePeer().Pretty())
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

func StartP2PServer() {
	sourcePort := config.P2P_sourcePort
	r := mrand.New(mrand.NewSource(int64(sourcePort)))
	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", sourcePort))

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
	serverReaderWriters = make(map[string]*bufio.ReadWriter)

	host.SetStreamHandler(pid, handleStream)

	log.Println(host.Peerstore().Addrs(pid))

	localServerDest = fmt.Sprintf("/ip4/127.0.0.1/tcp/%v/p2p/%s", sourcePort, host.ID().Pretty())
	log.Println("localServerDest", localServerDest)

	//本地client to 本地的server
	ConnP2PServer(localServerDest)
}

var localServerDest string

var clientReaderWriter *bufio.ReadWriter

func ConnP2PServer(dest string) {
	if tool.CheckIsString(&dest) == false {
		dest = localServerDest
	}

	sourcePort := 0
	r := mrand.New(mrand.NewSource(int64(sourcePort)))
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}
	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", sourcePort))

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

	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(dest)
	if err != nil {
		log.Fatalln(err)
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Fatalln(err)
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	// Start a stream with the destination.
	// Multiaddress of the destination peer is fetched from the peerstore using 'peerId'.
	s, err := host.NewStream(context.Background(), info.ID, pid)
	if err != nil {
		log.Println(err)
		return
	}

	localClientId = s.Conn().LocalPeer().Pretty()
	log.Println("localClientId", localClientId)

	clientReaderWriter = bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go readData(s, clientReaderWriter)
}

var localClientId string

func ClientSendMsg(msg string) {
	if clientReaderWriter != nil {
		msg = msg + "$"
		_, _ = clientReaderWriter.WriteString(msg)
		_ = clientReaderWriter.Flush()
	}
}
func ServerSendMsg(msg string) {
	if len(serverReaderWriters) > 0 {
		msg = msg + "$"
		_, _ = serverReaderWriters[localClientId].WriteString(msg)
		_ = serverReaderWriters[localClientId].Flush()
	}
}
