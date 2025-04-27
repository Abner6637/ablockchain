package cli

import (
	"ablockchain/crypto"
	"log"
	"testing"
)

func TestKeys(t *testing.T) {
	ethPrivateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	// ethAddress := crypto.PubkeyToAddress(ethPrivateKey.PublicKey)

	crypto.InitBLS()
	blsPrivateKey, err := crypto.ConvertEthToBLS(ethPrivateKey)
	if err != nil {
		t.Fatalf("Failed to convert to BLS: %v", err)
	}

	log.Printf("bls private key: %v", blsPrivateKey)
	log.Printf("eth private key: %v", ethPrivateKey)
}
