package cmd

import (
	"ablockchain/crypto"
	"crypto/ecdsa"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"
)

var keyGenNum int

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate key pairs",
	Run: func(cmd *cobra.Command, args []string) {
		if keyGenNum <= 0 {
			fmt.Println("Please set --num to a positive integer.")
			return
		}
		for i := 0; i < keyGenNum; i++ {
			ethPriv, err := crypto.GenerateKey()
			if err != nil {
				log.Fatalf("failed to create keys: %v", err)
			}
			keyfile := fmt.Sprintf("./key_store/nodekey%d", i)
			storePrivkey(keyfile, ethPriv)

			addressfile := fmt.Sprintf("./key_store/address%d", i)
			storeAddress(addressfile, ethPriv)

			p2pfile := fmt.Sprintf("./key_store/p2pID%d", i)
			storeP2PID(p2pfile, ethPriv)
		}

		bootnodeKey, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("failed to create keys: %v", err)
		}
		keyfile := "./key_store/nodekeyboot"
		storePrivkey(keyfile, bootnodeKey)

		addressfile := "./key_store/addressboot"
		storeAddress(addressfile, bootnodeKey)

		p2pfile := "./key_store/p2pIDboot"
		storeP2PID(p2pfile, bootnodeKey)
	},
}

func init() {
	rootCmd.AddCommand(keygenCmd)
	keygenCmd.Flags().IntVarP(&keyGenNum, "num", "n", 1, "Number of key pairs to generate")
}

func storePrivkey(keyfile string, ethPriv *ecdsa.PrivateKey) {
	dir := filepath.Dir(keyfile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create key directory %s: %v", dir, err)
	}
	if err := crypto.SaveECDSA(keyfile, ethPriv); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to persist node key: %v", err))
	}
}

func storeAddress(addressfile string, ethPriv *ecdsa.PrivateKey) {
	dir := filepath.Dir(addressfile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create key directory %s: %v", dir, err)
	}
	address := crypto.PubkeyToAddress(ethPriv.PublicKey).Bytes()
	hexAddress := fmt.Sprintf("0x%x", address)
	if err := os.WriteFile(addressfile, []byte(hexAddress), 0644); err != nil {
		log.Fatalf("Failed to persist node address: %v\n", err)
	}
}

func storeP2PID(p2pfile string, ethPriv *ecdsa.PrivateKey) {
	dir := filepath.Dir(p2pfile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create key directory %s: %v", dir, err)
	}
	libPriv, _ := crypto.ConvertEthToLibPrivateKey(ethPriv)
	p2pID, _ := peer.IDFromPrivateKey(libPriv)
	if err := os.WriteFile(p2pfile, []byte(p2pID.String()), 0644); err != nil {
		log.Fatalf("Failed to persist node P2PID: %v\n", err)
	}
}
