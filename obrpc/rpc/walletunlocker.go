package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/omnilaboratory/obd/obrpc"
	"github.com/omnilaboratory/obd/proxy/rpc"
	"github.com/omnilaboratory/obd/service"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/btcsuite/btcwallet/wallet"
	"github.com/lightningnetwork/lnd/aezeed"
	"github.com/lightningnetwork/lnd/chanbackup"
	"github.com/lightningnetwork/lnd/kvdb"
	"github.com/lightningnetwork/lnd/lnwallet/btcwallet"
)

var (
	// ErrUnlockTimeout signals that we did not get the expected unlock
	// message before the timeout occurred.
	ErrUnlockTimeout = errors.New("got no unlock message before timeout")
)

// WalletUnlockParams holds the variables used to parameterize the unlocking of
// lnd's wallet after it has already been created.
type WalletUnlockParams struct {
	// Password is the public and private wallet passphrase.
	Password []byte

	// Birthday specifies the approximate time that this wallet was created.
	// This is used to bound any rescans on startup.
	Birthday time.Time

	// RecoveryWindow specifies the address lookahead when entering recovery
	// mode. A recovery will be attempted if this value is non-zero.
	RecoveryWindow uint32

	// Wallet is the loaded and unlocked Wallet. This is returned
	// from the unlocker service to avoid it being unlocked twice (once in
	// the unlocker service to check if the password is correct and again
	// later when lnd actually uses it). Because unlocking involves scrypt
	// which is resource intensive, we want to avoid doing it twice.
	Wallet *wallet.Wallet

	// ChansToRestore a set of static channel backups that should be
	// restored before the main server instance starts up.
	ChansToRestore ChannelsToRecover

	// UnloadWallet is a function for unloading the wallet, which should
	// be called on shutdown.
	UnloadWallet func() error

	// StatelessInit signals that the user requested the daemon to be
	// initialized stateless, which means no unencrypted macaroons should be
	// written to disk.
	StatelessInit bool

	// MacResponseChan is the channel for sending back the admin macaroon to
	// the WalletUnlocker service.
	MacResponseChan chan []byte
}

// ChannelsToRecover wraps any set of packed (serialized+encrypted) channel
// back ups together. These can be passed in when unlocking the wallet, or
// creating a new wallet for the first time with an existing seed.
type ChannelsToRecover struct {
	// PackedMultiChanBackup is an encrypted and serialized multi-channel
	// backup.
	PackedMultiChanBackup chanbackup.PackedMulti

	// PackedSingleChanBackups is a series of encrypted and serialized
	// single-channel backup for one or more channels.
	PackedSingleChanBackups chanbackup.PackedSingles
}

// WalletInitMsg is a message sent by the UnlockerService when a user wishes to
// set up the internal wallet for the first time. The user MUST provide a
// passphrase, but is also able to provide their own source of entropy. If
// provided, then this source of entropy will be used to generate the wallet's
// HD seed. Otherwise, the wallet will generate one itself.
type WalletInitMsg struct {
	// Passphrase is the passphrase that will be used to encrypt the wallet
	// itself. This MUST be at least 8 characters.
	Passphrase []byte

	// WalletSeed is the deciphered cipher seed that the wallet should use
	// to initialize itself. The seed might be nil if the wallet should be
	// created from an extended master root key instead.
	WalletSeed *aezeed.CipherSeed

	// WalletExtendedKey is the wallet's extended master root key that
	// should be used instead of the seed, if non-nil. The extended key is
	// mutually exclusive to the wallet seed, but one of both is always set.
	WalletExtendedKey *hdkeychain.ExtendedKey

	// ExtendedKeyBirthday is the birthday of a wallet that's being restored
	// through an extended key instead of an aezeed.
	ExtendedKeyBirthday time.Time

	// WatchOnlyAccounts is a map of scoped account extended public keys
	// that should be imported to create a watch-only wallet.
	WatchOnlyAccounts map[waddrmgr.ScopedIndex]*hdkeychain.ExtendedKey

	// WatchOnlyBirthday is the birthday of the master root key the above
	// watch-only account xpubs were derived from.
	WatchOnlyBirthday time.Time

	// WatchOnlyMasterFingerprint is the fingerprint of the master root key
	// the above watch-only account xpubs were derived from.
	WatchOnlyMasterFingerprint uint32

	// RecoveryWindow is the address look-ahead used when restoring a seed
	// with existing funds. A recovery window zero indicates that no
	// recovery should be attempted, such as after the wallet's initial
	// creation.
	RecoveryWindow uint32

	// ChanBackups a set of static channel backups that should be received
	// after the wallet has been initialized.
	ChanBackups ChannelsToRecover

	// StatelessInit signals that the user requested the daemon to be
	// initialized stateless, which means no unencrypted macaroons should be
	// written to disk.
	StatelessInit bool
}

// WalletUnlockMsg is a message sent by the UnlockerService when a user wishes
// to unlock the internal wallet after initial setup. The user can optionally
// specify a recovery window, which will resume an interrupted rescan for used
// addresses.
type WalletUnlockMsg struct {
	// Passphrase is the passphrase that will be used to encrypt the wallet
	// itself. This MUST be at least 8 characters.
	Passphrase []byte

	// RecoveryWindow is the address look-ahead used when restoring a seed
	// with existing funds. A recovery window zero indicates that no
	// recovery should be attempted, such as after the wallet's initial
	// creation, but before any addresses have been created.
	RecoveryWindow uint32

	// Wallet is the loaded and unlocked Wallet. This is returned through
	// the channel to avoid it being unlocked twice (once to check if the
	// password is correct, here in the WalletUnlocker and again later when
	// lnd actually uses it). Because unlocking involves scrypt which is
	// resource intensive, we want to avoid doing it twice.
	Wallet *wallet.Wallet

	// ChanBackups a set of static channel backups that should be received
	// after the wallet has been unlocked.
	ChanBackups ChannelsToRecover

	// UnloadWallet is a function for unloading the wallet, which should
	// be called on shutdown.
	UnloadWallet func() error

	// StatelessInit signals that the user requested the daemon to be
	// initialized stateless, which means no unencrypted macaroons should be
	// written to disk.
	StatelessInit bool
}

// UnlockerService implements the WalletUnlocker service used to provide lnd
// with a password for wallet encryption at startup. Additionally, during
// initial setup, users can provide their own source of entropy which will be
// used to generate the seed that's ultimately used within the wallet.
type UnlockerService struct {
	// Required by the grpc-gateway/v2 library for forward compatibility.
	obrpc.UnimplementedWalletUnlockerServer

	// InitMsgs is a channel that carries all wallet init messages.
	InitMsgs chan *WalletInitMsg

	// UnlockMsgs is a channel where unlock parameters provided by the rpc
	// client to be used to unlock and decrypt an existing wallet will be
	// sent.
	UnlockMsgs chan *WalletUnlockMsg

	// MacResponseChan is the channel for sending back the admin macaroon to
	// the WalletUnlocker service.
	MacResponseChan chan []byte

	netParams *chaincfg.Params

	// macaroonFiles is the path to the three generated macaroons with
	// different access permissions. These might not exist in a stateless
	// initialization of lnd.
	macaroonFiles []string

	// resetWalletTransactions indicates that the wallet state should be
	// reset on unlock to force a full chain rescan.
	resetWalletTransactions bool

	// LoaderOpts holds the functional options for the wallet loader.
	loaderOpts []btcwallet.LoaderOption

	// macaroonDB is an instance of a database backend that stores all
	// macaroon root keys. This will be nil on initialization and must be
	// set using the SetMacaroonDB method as soon as it's available.
	macaroonDB kvdb.Backend
}

//
//func NewInstance(netName,dataPath string ) *UnlockerService {
//	var netParams *chaincfg.Params
//	if strings.HasPrefix(netName , "main")  {
//		netParams = &chaincfg.MainNetParams
//	} else if strings.HasPrefix(netName , "test") {
//		netParams = &chaincfg.TestNet3Params
//	} else if strings.HasPrefix(netName , "reg") {
//		netParams = &chaincfg.RegressionNetParams
//	} else {
//		log.Println("supported netName should be: main or test or reg")
//		log.Fatalln("error netName",netName)
//	}
//	dbDir := btcwallet.NetworkDir(dataPath, netParams)
//	loadOpts := []btcwallet.LoaderOption{
//		btcwallet.LoaderWithLocalWalletDB(dbDir, true, time.Minute),
//	}
//	return New(
//		netParams, nil, false, loadOpts,
//	)
//}
//
//// New creates and returns a new UnlockerService.
//func New(params *chaincfg.Params, macaroonFiles []string,
//	resetWalletTransactions bool,
//	loaderOpts []btcwallet.LoaderOption) *UnlockerService {
//
//	return &UnlockerService{
//		InitMsgs:   make(chan *WalletInitMsg, 1),
//		UnlockMsgs: make(chan *WalletUnlockMsg, 1),
//
//		// Make sure we buffer the channel is buffered so the main lnd
//		// goroutine isn't blocking on writing to it.
//		MacResponseChan:         make(chan []byte, 1),
//		netParams:               params,
//		macaroonFiles:           macaroonFiles,
//		resetWalletTransactions: resetWalletTransactions,
//		loaderOpts:              loaderOpts,
//	}
//}
//
//// SetLoaderOpts can be used to inject wallet loader options after the unlocker
//// service has been hooked to the main RPC server.
//func (u *UnlockerService) SetLoaderOpts(loaderOpts []btcwallet.LoaderOption) {
//	u.loaderOpts = loaderOpts
//}
//
//// SetMacaroonDB can be used to inject the macaroon database after the unlocker
//// service has been hooked to the main RPC server.
//func (u *UnlockerService) SetMacaroonDB(macaroonDB kvdb.Backend) {
//	u.macaroonDB = macaroonDB
//}
//
//func (u *UnlockerService) newLoader(recoveryWindow uint32) (*wallet.Loader,
//	error) {
//
//	return btcwallet.NewWalletLoader(
//		u.netParams, recoveryWindow, u.loaderOpts...,
//	)
//}
//
//// WalletExists returns whether a wallet exists on the file path the
//// UnlockerService is using.
//func (u *UnlockerService) WalletExists() (bool, error) {
//	loader, err := u.newLoader(0)
//	if err != nil {
//		return false, err
//	}
//	return loader.WalletExists()
//}

func (u *UnlockerService) GenSeed(_ context.Context,
	in *obrpc.GenSeedRequest) (*obrpc.GenSeedResponse, error) {
	mnemonic, err:=service.HDWalletService.GenSeed()
	return &obrpc.GenSeedResponse{
		Words: mnemonic,
	}, err
}

func (u *UnlockerService) InitWallet(ctx context.Context,
	in *obrpc.InitWalletRequest) (*obrpc.InitWalletResponse, error) {

	// Make sure the password meets our constraints.
	password := in.WalletPassword
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	if len(in.Mnemonic)==0{
		return nil, fmt.Errorf("must specify " +
			"mnemonic")
	}
	_ ,err:=service.HDWalletService.SaveNodeMasterKeyFormMnemonic(in.Mnemonic ,password,in.SeedPassphrase)
	if err==nil {
		key, err1 := service.HDWalletService.LoadNodeMasterKey(password)
		err = err1
		if err == nil {
			err = rpc.LoginByKey(key)
		}
	}
	return &obrpc.InitWalletResponse{},err
}


// UnlockWallet sends the password provided by the incoming UnlockWalletRequest
// over the UnlockMsgs channel in case it successfully decrypts an existing
// wallet found in the chain's wallet database directory.
func (u *UnlockerService) UnlockWallet(ctx context.Context,
	in *obrpc.UnlockWalletRequest) (*obrpc.UnlockWalletResponse, error) {
	password := in.WalletPassword
	key,err:=service.HDWalletService.LoadNodeMasterKey(string(password))
	if err == nil {
		err=rpc.LoginByKey(key)
		if err == nil {
			fmt.Println("\nobd node key successfully unlocked!")
		}
	}
	return &obrpc.UnlockWalletResponse{},err
}

// ChangePassword changes the password of the wallet and sends the new password
// across the UnlockPasswords channel to automatically unlock the wallet if
// successful.
func (u *UnlockerService) ChangePassword(ctx context.Context,
	in *obrpc.ChangePasswordRequest) (*obrpc.ChangePasswordResponse, error) {

	err:=ValidatePassword(in.CurrentPassword)
	if err != nil {
		return nil,fmt.Errorf("CurrentPassword err: %s",err.Error())
	}

	// Make sure the new password meets our constraints.
	if err := ValidatePassword(in.NewPassword); err != nil {
		return nil,fmt.Errorf("NewPassword err: %s",err.Error())
	}

	err=service.HDWalletService.ChangePwd(in.CurrentPassword,in.NewPassword)
	return &obrpc.ChangePasswordResponse{},err
}

// ValidatePassword assures the password meets all of our constraints.
func ValidatePassword(password string) error {
	// Passwords should have a length of at least 8 characters.
	if len(password) < 4 {
		return errors.New("password must have at least 8 characters")
	}

	return nil
}
