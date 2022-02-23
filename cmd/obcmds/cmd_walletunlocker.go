package obcmds

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/lightningnetwork/lnd/walletunlocker"
	"github.com/omnilaboratory/obd/obrpc"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"
)

var (
	//statelessInitFlag = &cli.BoolFlag{
	//	Name: "stateless_init",
	//	Usage: "do not create any macaroon files in the file " +
	//		"system of the daemon",
	//}
	//saveToFlag = &cli.StringFlag{
	//	Name:  "save_to",
	//	Usage: "save returned admin macaroon to this file",
	//}
)

var createCommand = cli.Command{
	Name:     "create",
	Category: "node-key",
	Usage:    "setup obd server local wallet's private key.",
	Description: `
	The create command is used to setup a server local wallet's private key. This is interactive command with one required
	argument (the password), and one optional argument (the mnemonic
	passphrase).

	The first argument (the password) is required and MUST be greater than
	8 characters. This will be used to encrypt the wallet within lnd. This
	MUST be remembered as it will be required to fully start up the daemon.

	The second argument is an optional 24-word mnemonic derived from BIP
	39. If provided, then the internal wallet will use the seed derived
	from this mnemonic to generate all keys.

	This command returns a 24-word seed in the scenario that NO mnemonic
	was provided by the user. This should be written down as it can be used
	to potentially recover all on-chain funds, and most off-chain funds as
	well.

	Finally, it's also possible to use this command and a set of static
	channel backups to trigger a recover attempt for the provided Static
	Channel Backups. Only one of the three parameters will be accepted. See
	the restorechanbackup command for further details w.r.t the format
	accepted.
	`,
	Flags: []cli.Flag{
	},
	Action: actionDecorator(create),
}

// monowidthColumns takes a set of words, and the number of desired columns,
// and returns a new set of words that have had white space appended to the
// word in order to create a mono-width column.
func monowidthColumns(words []string, ncols int) []string {
	// Determine max size of words in each column.
	colWidths := make([]int, ncols)
	for i, word := range words {
		col := i % ncols
		curWidth := colWidths[col]
		if len(word) > curWidth {
			colWidths[col] = len(word)
		}
	}

	// Append whitespace to each word to make columns mono-width.
	finalWords := make([]string, len(words))
	for i, word := range words {
		col := i % ncols
		width := colWidths[col]

		diff := width - len(word)
		finalWords[i] = word + strings.Repeat(" ", diff)
	}

	return finalWords
}

func create(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getClient(ctxc)
	defer cleanUp()
	resetTerm(ctxc)

	walletPassword, err := capturePassword(
		"Input wallet password: ", false, walletunlocker.ValidatePassword,
	)
	if err != nil {
		return err
	}

	hasMnemonic := false
mnemonicCheck:
	for {
		fmt.Println()
		fmt.Printf("Do you have an existing cipher seed " +
			"mnemonic you want to " +
			"use?\nEnter 'y' to use an existing  seed mnemonic" +
			"\nor 'n' to create a new seed (Enter y/n): ")

		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		fmt.Println()

		answer = strings.TrimSpace(answer)
		answer = strings.ToLower(answer)

		switch answer {
		case "y":
			hasMnemonic = true
			break mnemonicCheck

		case "n":
			break mnemonicCheck
		}
	}

	// If the user *does* have an existing seed or root key they want to
	// use, then we'll read that in directly from the terminal.
	var (
		cipherSeedMnemonic []string
		aezeedPass         []byte
	)
	switch {
	// Use an existing cipher seed mnemonic in the aezeed format.
	case hasMnemonic:
		// We'll now prompt the user to enter in their 24-word
		// mnemonic.
		fmt.Printf("Input your 24-word mnemonic separated by spaces: ")
		reader := bufio.NewReader(os.Stdin)
		mnemonic, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		// We'll trim off extra spaces, and ensure the mnemonic is all
		// lower case, then populate our request.
		mnemonic = strings.TrimSpace(mnemonic)
		mnemonic = strings.ToLower(mnemonic)

		cipherSeedMnemonic = strings.Split(mnemonic, " ")

		fmt.Println()

		if len(cipherSeedMnemonic) != 24 {
			return fmt.Errorf("wrong cipher seed mnemonic "+
				"length: got %v words, expecting %v words",
				len(cipherSeedMnemonic), 24)
		}

		// Additionally, the user may have a passphrase, that will also
		// need to be provided so the daemon can properly decipher the
		// cipher seed.
		aezeedPass, err = readPassword("Input your cipher seed " +
			"passphrase (press enter if your seed doesn't have a " +
			"passphrase): ")
		if err != nil {
			return err
		}

	default:
		// Otherwise, if the user doesn't have a mnemonic that they
		// want to use, we'll generate a fresh one with the GenSeed
		// command.
		fmt.Println("Your cipher seed can optionally be encrypted.")

		instruction := "Input your passphrase if you wish to encrypt it " +
			"(or press enter to proceed without a cipher seed " +
			"passphrase): "
		aezeedPass, err = capturePassword(
			instruction, true, func(_ []byte) error { return nil },
		)
		if err != nil {
			return err
		}

		fmt.Println()
		fmt.Println("Generating fresh cipher seed...")
		fmt.Println()

		genSeedReq := &obrpc.GenSeedRequest{
}
		seedResp, err := client.GenSeed(ctxc, genSeedReq)
		if err != nil {
			return fmt.Errorf("unable to generate seed: %v", err)
		}

		cipherSeedMnemonic = strings.Split(seedResp.Words, " ")
	}

	// Before we initialize the wallet, we'll display the cipher seed to
	// the user so they can write it down.
	if len(cipherSeedMnemonic) > 0 {
		printCipherSeedWords(cipherSeedMnemonic)
	}

	// With either the user's prior cipher seed, or a newly generated one,
	// we'll go ahead and initialize the wallet.
	req := &obrpc.InitWalletRequest{
		WalletPassword:                     string(walletPassword),
		Mnemonic:                 strings.Join( cipherSeedMnemonic ," "),
		SeedPassphrase:                   string(aezeedPass),
	}
	_, err = client.InitWallet(ctxc, req)
	if err != nil {
		return err
	}

	fmt.Println("\nobd successfully initialized!")

	return nil


}

// capturePassword returns a password value that has been entered twice by the
// user, to ensure that the user knows what password they have entered. The user
// will be prompted to retry until the passwords match. If the optional param is
// true, the function may return an empty byte array if the user opts against
// using a password.
func capturePassword(instruction string, optional bool,
	validate func([]byte) error) ([]byte, error) {

	for {
		password, err := readPassword(instruction)
		if err != nil {
			return nil, err
		}

		// Do not require users to repeat password if
		// it is optional and they are not using one.
		if len(password) == 0 && optional {
			return nil, nil
		}

		// If the password provided is not valid, restart
		// password capture process from the beginning.
		if err := validate(password); err != nil {
			fmt.Println(err.Error())
			fmt.Println()
			continue
		}

		passwordConfirmed, err := readPassword("Confirm password: ")
		if err != nil {
			return nil, err
		}

		if bytes.Equal(password, passwordConfirmed) {
			return password, nil
		}

		fmt.Println("Passwords don't match, please try again")
		fmt.Println()
	}
}

var unlockCommand = cli.Command{
	Name:     "unlock",
	Category: "node-key",
	Usage:    "Unlock an encrypted obd server local wallet's private key and init server local wallet private key.",
	Description: `
	The unlock command is used to decrypt obd's wallet private key and init server local wallet private key.`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "stdin",
			Usage: "read password from standard input instead of " +
				"prompting for it. THIS IS CONSIDERED TO " +
				"BE DANGEROUS if the password is located in " +
				"a file that can be read by another user. " +
				"This flag should only be used in " +
				"combination with some sort of password " +
				"manager or secrets vault.",
		},
	},
	Action: actionDecorator(unlock),
}

func unlock(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getClient(ctxc)
	defer cleanUp()

	var (
		pw  []byte
		err error
	)
	switch {
	// Read the password from standard in as if it were a file. This should
	// only be used if the password is piped into lncli from some sort of
	// password manager. If the user types the password instead, it will be
	// echoed in the console.
	case ctx.IsSet("stdin"):
		reader := bufio.NewReader(os.Stdin)
		pw, err = reader.ReadBytes('\n')

		// Remove carriage return and newline characters.
		pw = bytes.Trim(pw, "\r\n")

	// Read the password from a terminal by default. This requires the
	// terminal to be a real tty and will fail if a string is piped into
	// lncli.
	default:
		resetTerm(ctxc)
		pw, err = readPassword("Input wallet password: ")
	}
	if err != nil {
		return err
	}

	req := &obrpc.UnlockWalletRequest{
		WalletPassword: pw,
		//RecoveryWindow: recoveryWindow,
		//StatelessInit:  ctx.Bool(statelessInitFlag.Name),
	}
	_, err = client.UnlockWallet(ctxc, req)
	if err != nil {
		return err
	}

	fmt.Println("\nobd obd-node key successfully unlocked!")

	return nil
}

var changePasswordCommand = cli.Command{
	Name:     "changepassword",
	Category: "node-key",
	Usage:    "Change an encrypted obd wallet's password at startup.",
	Description: `
	The changepassword command is used to Change obd's encrypted wallet-key's
	password..
	`,
	Flags: []cli.Flag{
		//statelessInitFlag,
		//saveToFlag,
		//&cli.BoolFlag{
		//	Name: "new_mac_root_key",
		//	Usage: "rotate the macaroon root key resulting in " +
		//		"all previously created macaroons to be " +
		//		"invalidated",
		//},
	},
	Action: actionDecorator(changePassword),
}

func resetTerm(ctxc context.Context){
	oldstat,err:=term.GetState(syscall.Stdin)
	if err != nil {
		panic(err)
	}
	defer term.Restore(syscall.Stdin, oldstat)
	go func() {
		<-ctxc.Done()
		term.Restore(syscall.Stdin, oldstat)
		os.Exit(1)
	}()
}
func changePassword(ctx *cli.Context) error {
	ctxc := getContext()
	client, cleanUp := getClient(ctxc)
	defer cleanUp()
	resetTerm(ctxc)

	currentPw, err := readPassword("Input current wallet password: ")
	if err != nil {
		return err
	}

	newPw, err := readPassword("Input new wallet password: ")
	if err != nil {
		return err
	}

	confirmPw, err := readPassword("Confirm new wallet password: ")
	if err != nil {
		return err
	}

	if !bytes.Equal(newPw, confirmPw) {
		return fmt.Errorf("passwords don't match")
	}

	req := &obrpc.ChangePasswordRequest{
		CurrentPassword:    string(currentPw),
		NewPassword:        string(newPw),
	}
	_,err = client.ChangePassword(ctxc, req)
	return err
}

func printCipherSeedWords(mnemonicWords []string) {
	fmt.Println("!!!YOU MUST WRITE DOWN THIS SEED TO BE ABLE TO " +
		"RESTORE THE WALLET!!!")
	fmt.Println()

	fmt.Println("---------------BEGIN OBD CIPHER SEED---------------")

	numCols := 4
	colWords := monowidthColumns(mnemonicWords, numCols)
	for i := 0; i < len(colWords); i += numCols {
		fmt.Printf("%2d. %3s  %2d. %3s  %2d. %3s  %2d. %3s\n",
			i+1, colWords[i], i+2, colWords[i+1], i+3,
			colWords[i+2], i+4, colWords[i+3])
	}

	fmt.Println("---------------END OBD CIPHER SEED-----------------")

	fmt.Println("\n!!!YOU MUST WRITE DOWN THIS SEED TO BE ABLE TO " +
		"RESTORE THE WALLET!!!")
}
