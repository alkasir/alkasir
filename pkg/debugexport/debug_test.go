package debugexport

import "testing"

func TestDebug(t *testing.T) {
	pubKey := "9b5a2826ca25d023440bbee884ecadaaedb97648ddbdc9a2b03ba244621bbc42"
	secKey := "b8c29f66fe85c75941b3e140f39dcd69b545259dd1613639f1cdfbf46aae9185"

	type TestConfig struct{ test string }

	resp := NewDebugResposne("0.0.0", TestConfig{test: "test value"})
	resp.Log = append(resp.Log, "test-log-entry")

	if len(resp.Log) != 1 {
		t.Error("resp.Log length != 1")
	}

	if resp.Encrypted != "" {
		t.Error("encrypted field not empty")
	}

	err := resp.Encrypt(pubKey)
	if err != nil {
		t.Error(err)
	}

	if resp.Encrypted == "" {
		t.Error("encrypted field empty")
	}
	if len(resp.Log) != 0 {
		t.Error("resp.Log length != 0")
	}

	err = resp.Decrypt(pubKey, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	if err == nil {
		t.Error("private key is not valid")
	}

	err = resp.Decrypt(pubKey, secKey)
	if err != nil {
		t.Error(err)
	}

	if len(resp.Log) != 1 {
		t.Error("resp.Log length != 1")
	}

	if resp.Encrypted != "" {
		t.Error("encrypted field not empty")
	}

	PublicKey = pubKey
	resp = NewDebugResposne("0.0.0", TestConfig{test: "test value"})

	if resp.Encrypted == "" {
		t.Error("encrypted field empty")
	}

	err = resp.Decrypt(pubKey, secKey)
	if err != nil {
		t.Error(err)
	}

}
