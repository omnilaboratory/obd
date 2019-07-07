package LightningOnOmni


import (
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	
)
// rpcServer is a gRPC, RPC front end to the lnd daemon.
// TODO(roasbeef): pagination support for the list-style calls
type rpcServer struct {
	started  int32 // To be used atomically.
	shutdown int32 // To be used atomically.

	server *server

	wg sync.WaitGroup

	// subServers are a set of sub-RPC servers that use the same gRPC and
	// listening sockets as the main RPC server, but which maintain their
	// own independent service. This allows us to expose a set of
	// micro-service like abstractions to the outside world for users to
	// consume.
	subServers []lnrpc.SubServer

	// grpcServer is the main gRPC server that this RPC server, and all the
	// sub-servers will use to register themselves and accept client
	// requests from.
	grpcServer *grpc.Server

	// listenerCleanUp are a set of closures functions that will allow this
	// main RPC server to clean up all the listening socket created for the
	// server.
	listenerCleanUp []func()

	// restDialOpts are a set of gRPC dial options that the REST server
	// proxy will use to connect to the main gRPC server.
	restDialOpts []grpc.DialOption

	// restProxyDest is the address to forward REST requests to.
	restProxyDest string

	// tlsCfg is the TLS config that allows the REST server proxy to
	// connect to the main gRPC server to proxy all incoming requests.
	tlsCfg *tls.Config

	// routerBackend contains the backend implementation of the router
	// rpc sub server.
	routerBackend *routerrpc.RouterBackend

	quit chan struct{}
}

// OpenChannel attempts to open a singly funded channel specified in the
// request to a remote peer.
func (r *rpcServer) OpenChannel(in *lnrpc.OpenChannelRequest,
	updateStream lnrpc.Lightning_OpenChannelServer) error {

	rpcsLog.Tracef("[openchannel] request to NodeKey(%v) "+
		"allocation(us=%v, them=%v)", in.NodePubkeyString,
		in.LocalFundingAmount, in.PushSat)

	if !r.server.Started() {
		return fmt.Errorf("chain backend is still syncing, server " +
			"not active yet")
	}

	localFundingAmt := btcutil.Amount(in.LocalFundingAmount)
	remoteInitialBalance := btcutil.Amount(in.PushSat)
	minHtlc := lnwire.MilliSatoshi(in.MinHtlcMsat)
	remoteCsvDelay := uint16(in.RemoteCsvDelay)

	// Ensure that the initial balance of the remote party (if pushing
	// satoshis) does not exceed the amount the local party has requested
	// for funding.
	//
	// TODO(roasbeef): incorporate base fee?
	if remoteInitialBalance >= localFundingAmt {
		return fmt.Errorf("amount pushed to remote peer for initial " +
			"state must be below the local funding amount")
	}

	// Ensure that the user doesn't exceed the current soft-limit for
	// channel size. If the funding amount is above the soft-limit, then
	// we'll reject the request.
	if localFundingAmt > MaxFundingAmount {
		return fmt.Errorf("funding amount is too large, the max "+
			"channel size is: %v", MaxFundingAmount)
	}

	// Restrict the size of the channel we'll actually open. At a later
	// level, we'll ensure that the output we create after accounting for
	// fees that a dust output isn't created.
	if localFundingAmt < minChanFundingSize {
		return fmt.Errorf("channel is too small, the minimum channel "+
			"size is: %v SAT", int64(minChanFundingSize))
	}

	// Then, we'll extract the minimum number of confirmations that each
	// output we use to fund the channel's funding transaction should
	// satisfy.
	minConfs, err := extractOpenChannelMinConfs(in)
	if err != nil {
		return err
	}

	var (
		nodePubKey      *btcec.PublicKey
		nodePubKeyBytes []byte
	)

	// TODO(roasbeef): also return channel ID?

	// Ensure that the NodePubKey is set before attempting to use it
	if len(in.NodePubkey) == 0 {
		return fmt.Errorf("NodePubKey is not set")
	}

	// Parse the raw bytes of the node key into a pubkey object so we
	// can easily manipulate it.
	nodePubKey, err = btcec.ParsePubKey(in.NodePubkey, btcec.S256())
	if err != nil {
		return err
	}

	// Making a channel to ourselves wouldn't be of any use, so we
	// explicitly disallow them.
	if nodePubKey.IsEqual(r.server.identityPriv.PubKey()) {
		return fmt.Errorf("cannot open channel to self")
	}

	nodePubKeyBytes = nodePubKey.SerializeCompressed()

	// Based on the passed fee related parameters, we'll determine an
	// appropriate fee rate for the funding transaction.
	satPerKw := lnwallet.SatPerKVByte(in.SatPerByte * 1000).FeePerKWeight()
	feeRate, err := sweep.DetermineFeePerKw(
		r.server.cc.feeEstimator, sweep.FeePreference{
			ConfTarget: uint32(in.TargetConf),
			FeeRate:    satPerKw,
		},
	)
	if err != nil {
		return err
	}

	rpcsLog.Debugf("[openchannel]: using fee of %v sat/kw for funding tx",
		int64(feeRate))

	// Instruct the server to trigger the necessary events to attempt to
	// open a new channel. A stream is returned in place, this stream will
	// be used to consume updates of the state of the pending channel.
	req := &openChanReq{
		targetPubkey:    nodePubKey,
		chainHash:       *activeNetParams.GenesisHash,
		localFundingAmt: localFundingAmt,
		pushAmt:         lnwire.NewMSatFromSatoshis(remoteInitialBalance),
		minHtlc:         minHtlc,
		fundingFeePerKw: feeRate,
		private:         in.Private,
		remoteCsvDelay:  remoteCsvDelay,
		minConfs:        minConfs,
	}

	updateChan, errChan := r.server.OpenChannel(req)

	var outpoint wire.OutPoint
out:
	for {
		select {
		case err := <-errChan:
			rpcsLog.Errorf("unable to open channel to NodeKey(%x): %v",
				nodePubKeyBytes, err)
			return err
		case fundingUpdate := <-updateChan:
			rpcsLog.Tracef("[openchannel] sending update: %v",
				fundingUpdate)
			if err := updateStream.Send(fundingUpdate); err != nil {
				return err
			}

			// If a final channel open update is being sent, then
			// we can break out of our recv loop as we no longer
			// need to process any further updates.
			switch update := fundingUpdate.Update.(type) {
			case *lnrpc.OpenStatusUpdate_ChanOpen:
				chanPoint := update.ChanOpen.ChannelPoint
				txid, err := GetChanPointFundingTxid(chanPoint)
				if err != nil {
					return err
				}
				outpoint = wire.OutPoint{
					Hash:  *txid,
					Index: chanPoint.OutputIndex,
				}

				break out
			}
		case <-r.quit:
			return nil
		}
	}

	rpcsLog.Tracef("[openchannel] success NodeKey(%x), ChannelPoint(%v)",
		nodePubKeyBytes, outpoint)
	return nil
}