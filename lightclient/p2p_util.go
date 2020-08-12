package lightclient

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"github.com/omnilaboratory/obd/bean"
	"github.com/omnilaboratory/obd/config"
	"github.com/omnilaboratory/obd/service"
	"github.com/omnilaboratory/obd/tool"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

type P2PChannel struct {
	IsLocalChannel bool
	Address        string
	stream         network.Stream
	rw             *bufio.ReadWriter
}

const pid = "/chat/1.0.0"

var localServerDest string
var P2PLocalPeerId string
var privateKey crypto.PrivKey
var p2pChannelMap map[string]*P2PChannel

func generatePrivateKey() (crypto.PrivKey, error) {
	if privateKey == nil {
		//r := rand.New(rand.NewSource(time.Now().UnixNano()))

		nodeId := httpGetNodeIdFromTracker()
		if nodeId == 0 {
			return nil, errors.New("fail to get nodeId from tracker")
		}
		r := rand.New(rand.NewSource(int64(8080 + nodeId)))
		//prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
		prvKey, _, err := crypto.GenerateECDSAKeyPair(r)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		privateKey = prvKey
	}
	return privateKey, nil
}

func StartP2PServer() (err error) {
	prvKey, err := generatePrivateKey()
	if err != nil {
		return err
	}
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
		return err
	}
	p2pChannelMap = make(map[string]*P2PChannel)
	P2PLocalPeerId = host.ID().Pretty()
	service.P2PLocalPeerId = P2PLocalPeerId

	localServerDest = fmt.Sprintf("/ip4/%s/tcp/%v/p2p/%s", config.P2P_hostIp, config.P2P_sourcePort, host.ID().Pretty())
	bean.MyObdNodeInfo.P2pAddress = localServerDest

	//把自己也作为终点放进去，阻止自己连接自己
	p2pChannelMap[P2PLocalPeerId] = &P2PChannel{
		IsLocalChannel: true,
		Address:        localServerDest,
	}
	host.SetStreamHandler(pid, handleStream)
	return nil
}

func ConnP2PServer(dest string) (string, error) {
	if tool.CheckIsString(&dest) == false {
		log.Println("wrong dest address")
		return "", errors.New("wrong dest address")
	}

	//id := httpGetNodeInfoByP2pAddressFromTracker(dest)
	//if id == 0 {
	//	return "", errors.New("target dest address not exist or online")
	//}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.P2P_sourcePort))
	prvKey, _ := generatePrivateKey()

	host, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)

	if err != nil {
		log.Println(err)
		return "", err
	}
	destMaddr, err := multiaddr.NewMultiaddr(dest)
	if err != nil {
		log.Println(err)
		return "", err
	}

	destHostInfo, err := peer.AddrInfoFromP2pAddr(destMaddr)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if p2pChannelMap[destHostInfo.ID.Pretty()] != nil {
		log.Println("p2p channel has connect")
		return localServerDest, nil
	}
	host.Peerstore().AddAddrs(destHostInfo.ID, destHostInfo.Addrs, peerstore.PermanentAddrTTL)
	s, err := host.NewStream(context.Background(), destHostInfo.ID, pid)
	if err != nil {
		log.Println(err)
		return "", err
	}
	rw := addP2PChannel(s)
	go readData(s, rw)
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

func SendP2PMsg(remoteP2PPeerId string, msg string) error {
	if tool.CheckIsString(&remoteP2PPeerId) == false {
		return errors.New("empty remoteP2PPeerId")
	}
	if remoteP2PPeerId == P2PLocalPeerId {
		return errors.New("remoteP2PPeerId is self,can not send msg to yourself")
	}

	channel := p2pChannelMap[remoteP2PPeerId]
	if channel != nil {
		msg = msg + "~"
		_, _ = channel.rw.WriteString(msg)
		_ = channel.rw.Flush()
	}
	return nil
}

func httpGetNodeInfoByP2pAddressFromTracker(p2pAddress string) (id int) {
	url := "http://" + config.TrackerHost + "/api/v1/getNodeInfoByP2pAddress?p2pAddress=" + p2pAddress
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return int(gjson.Get(string(body), "data").Get("id").Int())
	}
	return 0
}

func httpGetNodeIdFromTracker() (nodeId int) {
	url := "http://" + config.TrackerHost + "/api/v1/getNodeDbId?nodeId=" + tool.GetObdNodeId()
	resp, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return int(gjson.Get(string(body), "data").Get("id").Int())
	}
	return 0
}
