package sodiumbox

import (
	"encoding/hex"
	"testing"
)

const recipientPrivKey = "121586667a3f88aa5ab6b1bfca31fc58b089f6c11b1a2db8d758e0be65e43219"
const recipientPubKey = "c0a938fd04f1268d33d8960d3014ce07cfc119c64d7b94d95b7f605fc36b0260"

type keyPair struct {
	secretKey [32]byte
	publicKey [32]byte
}

func checkErr(t *testing.T, err error) bool {
	if err != nil {
		t.Logf("Error: %+v", err)
		t.Fail()
		return true
	}
	return false
}

func getTestKeyPair() *keyPair {
	publicKeySlice, err := hex.DecodeString(recipientPubKey)
	if err != nil {
		panic(err)
	}
	publicKey := *new([32]byte)
	copy(publicKey[:], publicKeySlice[0:32])
	secretKeySlice, err := hex.DecodeString(recipientPrivKey)
	if err != nil {
		panic(err)
	}
	secretKey := *new([32]byte)
	copy(secretKey[:], secretKeySlice[0:32])
	return &keyPair{
		secretKey: secretKey,
		publicKey: publicKey,
	}
}

func TestSeal(t *testing.T) {
	testKeyPair := getTestKeyPair()
	sealedMsg, err := Seal([]byte("secretmessage"), &testKeyPair.publicKey)
	t.Logf("Encrypted Message for PubKey (%s)\n%s\n", recipientPubKey, hex.EncodeToString(sealedMsg.Box))
	checkErr(t, err)
}

func TestSealOpen(t *testing.T) {
	testKeyPair := getTestKeyPair()
	sealedMsg, err := Seal([]byte("message"), &testKeyPair.publicKey)
	t.Logf("Encrypted Message for PubKey (%s)\n%s\n", recipientPubKey, hex.EncodeToString(sealedMsg.Box))
	if checkErr(t, err) {
		return
	}
	msg, err := SealOpen(sealedMsg.Box, &testKeyPair.publicKey, &testKeyPair.secretKey)
	if checkErr(t, err) {
		return
	}
	if string(msg.Content) != "message" {
		t.Logf("%+v", msg)
		t.Fail()
	}
}
