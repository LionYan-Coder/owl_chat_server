package chat

import (
	"bytes"
	_ "encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestValidateSignature(t *testing.T) {
	var testmsg = hexutil.MustDecode("0xe629272d1a982630dd780e2ec8b2175d437fca253a8dce72425397938001a55e")
	var testpubkey = hexutil.MustDecode("0x0464a074f9c86af7113a94557e55d01b6ebd321c94b242607bd8878055208020cf3fd8261901d2bd7ecdbd5ef75614b8e664a4fa255a3b72bfae9a90e63d8705d0")
	var testsig = hexutil.MustDecode("0xaf7306df92592c7d727e90a518e6008340c4c8083286db982d1a9c276762238841c005b00cfbcde3dcf35be1408e21c34a49dfb9daeda3bd7ffc78538e75639e")
	//sig := testsig[:len(testsig)-1]
	//
	if !crypto.VerifySignature(testpubkey, testmsg, testsig) {
		t.Errorf("can't verify signature with uncompressed key")
	}
	//
	//pubkey, err := crypto.Ecrecover(testmsg, testsig)
	//if err != nil {
	//	t.Fatalf("recover error: %s", err)
	//}
	//if !bytes.Equal(pubkey, testpubkey) {
	//	t.Errorf("pubkey mismatch: want: %x have: %x", testpubkey, pubkey)
	//}
}

var (
	testmsg     = hexutil.MustDecode("0xce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008")
	testsig     = hexutil.MustDecode("0x90f27b8b488db00b00606796d2987f6a5f59ae62ea05effe84fef5b8b0e549984a691139ad57a3f0b906637673aa2f63d1f55cb1a69199d4009eea23ceaddc9301")
	testpubkey  = hexutil.MustDecode("0x04e32df42865e97135acfb65f3bae71bdc86f4d49150ad6a440b6f15878109880a0a2b2667f7e725ceea70c673093bf67663e0312623c8e091b13cf2c0f11ef652")
	testpubkeyc = hexutil.MustDecode("0x02e32df42865e97135acfb65f3bae71bdc86f4d49150ad6a440b6f15878109880a")
)

func TestEcrecover(t *testing.T) {
	pubkey, err := crypto.Ecrecover(testmsg, testsig)
	if err != nil {
		t.Fatalf("recover error: %s", err)
	}

	if !bytes.Equal(pubkey, testpubkey) {
		t.Errorf("pubkey mismatch: want: %x have: %x", testpubkey, pubkey)
	}
}

func TestVerifySignature(t *testing.T) {
	sig := testsig[:len(testsig)-1] // remove recovery id
	if !crypto.VerifySignature(testpubkey, testmsg, sig) {
		t.Errorf("can't verify signature with uncompressed key")
	}
	if !crypto.VerifySignature(testpubkeyc, testmsg, sig) {
		t.Errorf("can't verify signature with compressed key")
	}

	if crypto.VerifySignature(nil, testmsg, sig) {
		t.Errorf("signature valid with no key")
	}
	if crypto.VerifySignature(testpubkey, nil, sig) {
		t.Errorf("signature valid with no message")
	}
	if crypto.VerifySignature(testpubkey, testmsg, nil) {
		t.Errorf("nil signature valid")
	}
	if crypto.VerifySignature(testpubkey, testmsg, append(common.CopyBytes(sig), 1, 2, 3)) {
		t.Errorf("signature valid with extra bytes at the end")
	}
	if crypto.VerifySignature(testpubkey, testmsg, sig[:len(sig)-2]) {
		t.Errorf("signature valid even though it's incomplete")
	}
	wrongkey := common.CopyBytes(testpubkey)
	wrongkey[10]++
	if crypto.VerifySignature(wrongkey, testmsg, sig) {
		t.Errorf("signature valid with wrong public key")
	}
}

// This test checks that VerifySignature rejects malleable signatures with s > N/2.
func TestVerifySignatureMalleable(t *testing.T) {
	sig := hexutil.MustDecode("0x638a54215d80a6713c8d523a6adc4e6e73652d859103a36b700851cb0e61b66b8ebfc1a610c57d732ec6e0a8f06a9a7a28df5051ece514702ff9cdff0b11f454")
	key := hexutil.MustDecode("0x03ca634cae0d49acb401d8a4c6b6fe8c55b70d115bf400769cc1400f3258cd3138")
	msg := hexutil.MustDecode("0xd301ce462d3e639518f482c7f03821fec1e602018630ce621e1e7851c12343a6")
	if crypto.VerifySignature(key, msg, sig) {
		t.Error("VerifySignature returned true for malleable signature")
	}
}
