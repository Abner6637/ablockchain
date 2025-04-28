package crypto

import (
	"testing"
)

func TestBLSSign(t *testing.T) {

	InitBLS()

	var aggSig *BLSSig
	var blsPubKeys []BLSPubKey
	var blsSigs []BLSSig

	for i := 0; i < 4; i++ {

		// Generate a BLS private key
		ethPrivateKey, err := GenerateKey()
		if err != nil {
			t.Fatalf("Failed to generate private key: %v", err)
		}

		blsPrivateKey, err := ConvertEthToBLS(ethPrivateKey)
		if err != nil {
			t.Fatalf("Failed to convert to BLS private key: %v", err)
		}

		blsPubKey := blsPrivateKey.GetPublicKey()
		blsPubKeys = append(blsPubKeys, *blsPubKey)

		blsPubKeyHex := blsPubKey.SerializeToHexStr()
		t.Logf("BLS Public Key %d: %s", i, blsPubKeyHex)

		var blsPubKeyFromHex BLSPubKey
		blsPubKeyFromHex.DeserializeHexStr(blsPubKeyHex)
		if blsPubKeyFromHex.IsEqual(blsPubKey) {
			t.Logf("BLS Public Key %d deserialized successfully", i)
		} else {
			t.Fatalf("Failed to deserialize BLS Public Key %d", i)
		}

		// Create a message to sign
		msg := []byte("Hello, BLS!")

		// Sign the message
		blsSig, err := BLSSign(msg, blsPrivateKey)
		if err != nil {
			t.Fatalf("Failed to sign message: %v", err)
		}
		blsSigs = append(blsSigs, *blsSig)

		// Verify the signature
		ok := BLSVerify(msg, blsPubKey, blsSig)
		if !ok {
			t.Fatalf("Signature verification failed")
		} else {
			t.Logf("Signature verification succeeded")
		}
	}

	aggSig, err := AggregateSignatures(blsSigs)
	if err != nil {
		t.Fatalf("Failed to aggregate signatures: %v", err)
	}

	ok := VerifyAggregateSignature(blsPubKeys, aggSig, []byte("Hello, BLS!"))
	if !ok {
		t.Fatalf("Aggregate signature verification failed")
	} else {
		t.Logf("Aggregate signature verification succeeded")
	}
}
