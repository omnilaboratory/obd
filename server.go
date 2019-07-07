package LightningOnOmni

import(
	"github.com/lightningnetwork/lnd/channeldb"
	"github.com/lightningnetwork/lnd/htlcswitch"
	"github.com/lightningnetwork/lnd/netann"
)
// server is the main server of the Lightning Network Daemon. The server houses
// global state pertaining to the wallet, database, and the rpcserver.
// Additionally, the server is also used as a central messaging bus to interact
// with any of its companion objects.
type server struct {
	active   int32 // atomic
	stopping int32 // atomic

	start sync.Once
	stop  sync.Once

	// identityPriv is the private key used to authenticate any incoming
	// connections.
	identityPriv *btcec.PrivateKey

	// nodeSigner is an implementation of the MessageSigner implementation
	// that's backed by the identity private key of the running lnd node.
	nodeSigner *netann.NodeSigner

	chanStatusMgr *netann.ChanStatusManager

	// listenAddrs is the list of addresses the server is currently
	// listening on.
	listenAddrs []net.Addr

	// torController is a client that will communicate with a locally
	// running Tor server. This client will handle initiating and
	// authenticating the connection to the Tor server, automatically
	// creating and setting up onion services, etc.
	torController *tor.Controller

	// natTraversal is the specific NAT traversal technique used to
	// automatically set up port forwarding rules in order to advertise to
	// the network that the node is accepting inbound connections.
	natTraversal nat.Traversal

	// lastDetectedIP is the last IP detected by the NAT traversal technique
	// above. This IP will be watched periodically in a goroutine in order
	// to handle dynamic IP changes.
	lastDetectedIP net.IP

	mu         sync.RWMutex
	peersByPub map[string]*peer

	inboundPeers  map[string]*peer
	outboundPeers map[string]*peer

	peerConnectedListeners    map[string][]chan<- lnpeer.Peer
	peerDisconnectedListeners map[string][]chan<- struct{}

	persistentPeers        map[string]struct{}
	persistentPeersBackoff map[string]time.Duration
	persistentConnReqs     map[string][]*connmgr.ConnReq
	persistentRetryCancels map[string]chan struct{}

	// ignorePeerTermination tracks peers for which the server has initiated
	// a disconnect. Adding a peer to this map causes the peer termination
	// watcher to short circuit in the event that peers are purposefully
	// disconnected.
	ignorePeerTermination map[*peer]struct{}

	// scheduledPeerConnection maps a pubkey string to a callback that
	// should be executed in the peerTerminationWatcher the prior peer with
	// the same pubkey exits.  This allows the server to wait until the
	// prior peer has cleaned up successfully, before adding the new peer
	// intended to replace it.
	scheduledPeerConnection map[string]func()

	cc *chainControl

	fundingMgr *fundingManager

	chanDB *channeldb.DB

	htlcSwitch *htlcswitch.Switch

	invoices *invoices.InvoiceRegistry

	channelNotifier *channelnotifier.ChannelNotifier

	witnessBeacon contractcourt.WitnessBeacon

	breachArbiter *breachArbiter

	missionControl *routing.MissionControl

	chanRouter *routing.ChannelRouter

	controlTower routing.ControlTower

	authGossiper *discovery.AuthenticatedGossiper

	utxoNursery *utxoNursery

	sweeper *sweep.UtxoSweeper

	chainArb *contractcourt.ChainArbitrator

	sphinx *htlcswitch.OnionProcessor

	towerClient wtclient.Client

	connMgr *connmgr.ConnManager

	sigPool *lnwallet.SigPool

	writePool *pool.Write

	readPool *pool.Read

	// globalFeatures feature vector which affects HTLCs and thus are also
	// advertised to other nodes.
	globalFeatures *lnwire.FeatureVector

	// currentNodeAnn is the node announcement that has been broadcast to
	// the network upon startup, if the attributes of the node (us) has
	// changed since last start.
	currentNodeAnn *lnwire.NodeAnnouncement

	// chansToRestore is the set of channels that upon starting, the server
	// should attempt to restore/recover.
	chansToRestore walletunlocker.ChannelsToRecover

	// chanSubSwapper is a sub-system that will ensure our on-disk channel
	// backups are consistent at all times. It interacts with the
	// channelNotifier to be notified of newly opened and closed channels.
	chanSubSwapper *chanbackup.SubSwapper

	quit chan struct{}

	wg sync.WaitGroup
}