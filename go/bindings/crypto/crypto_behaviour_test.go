package crypto

import (
	core "dappco.re/go"
	"dappco.re/go/py/runtime"
)

// newCryptoInterpreter registers the crypto module against a fresh bootstrap
// interpreter and returns the direct caller used to exercise the bindings.
//
//	caller := newCryptoInterpreter(t)
//	value, err := caller.Call("core.crypto", "sha256", "data")
func newCryptoInterpreter(t *core.T) runtime.DirectCaller {
	t.Helper()

	interpreter, err := runtime.New(runtime.Options{})
	if err != nil {
		t.Fatalf("create interpreter: %v", err)
	}
	t.Cleanup(func() { _ = interpreter.Close() })

	if err := Register(interpreter); err != nil {
		t.Fatalf("register crypto module: %v", err)
	}

	caller, ok := interpreter.(runtime.DirectCaller)
	if !ok {
		t.Fatalf("interpreter does not expose direct calls: %T", interpreter)
	}
	return caller
}

// TestCrypto_Digests_Good produces known SHA-1 and SHA-256 digests.
func TestCrypto_Digests_Good(t *core.T) {
	caller := newCryptoInterpreter(t)

	sha1Digest, callErr := caller.Call("core.crypto", "sha1", "abc")
	if callErr != nil {
		t.Fatalf("sha1: %v", callErr)
	}
	if sha1Digest != "a9993e364706816aba3e25717850c26c9cd0d89d" {
		t.Fatalf("unexpected sha1 %#v", sha1Digest)
	}

	sha256Digest, callErr := caller.Call("core.crypto", "sha256", "abc")
	if callErr != nil {
		t.Fatalf("sha256: %v", callErr)
	}
	if sha256Digest != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("unexpected sha256 %#v", sha256Digest)
	}
}

// TestCrypto_HMACCompare_Good computes an HMAC and confirms a constant-time
// digest comparison.
func TestCrypto_HMACCompare_Good(t *core.T) {
	caller := newCryptoInterpreter(t)

	mac, callErr := caller.Call("core.crypto", "hmac_sha256", "key", "message")
	if callErr != nil {
		t.Fatalf("hmac_sha256: %v", callErr)
	}
	text, ok := mac.(string)
	if !ok || text == "" {
		t.Fatalf("hmac_sha256: unexpected %#v", mac)
	}

	equal, callErr := caller.Call("core.crypto", "compare_digest", text, text)
	if callErr != nil {
		t.Fatalf("compare_digest: %v", callErr)
	}
	if equal != true {
		t.Fatalf("compare_digest: expected equal, got %#v", equal)
	}

	differ, callErr := caller.Call("core.crypto", "compare_digest", text, "tampered")
	if callErr != nil {
		t.Fatalf("compare_digest differ: %v", callErr)
	}
	if differ != false {
		t.Fatalf("compare_digest: expected unequal, got %#v", differ)
	}
}

// TestCrypto_Base64RoundTrip_Good encodes and decodes back to the input.
func TestCrypto_Base64RoundTrip_Good(t *core.T) {
	caller := newCryptoInterpreter(t)

	encoded, callErr := caller.Call("core.crypto", "base64_encode", "corepy")
	if callErr != nil {
		t.Fatalf("base64_encode: %v", callErr)
	}
	if encoded != "Y29yZXB5" {
		t.Fatalf("unexpected encoding %#v", encoded)
	}

	decoded, callErr := caller.Call("core.crypto", "base64_decode", encoded)
	if callErr != nil {
		t.Fatalf("base64_decode: %v", callErr)
	}
	bytes, ok := decoded.([]byte)
	if !ok || string(bytes) != "corepy" {
		t.Fatalf("unexpected decoding %#v", decoded)
	}
}

// TestCrypto_RandomBytes_Good returns a buffer of the requested size.
func TestCrypto_RandomBytes_Good(t *core.T) {
	caller := newCryptoInterpreter(t)

	value, callErr := caller.Call("core.crypto", "random_bytes", 16)
	if callErr != nil {
		t.Fatalf("random_bytes: %v", callErr)
	}
	buffer, ok := value.([]byte)
	if !ok || len(buffer) != 16 {
		t.Fatalf("random_bytes: unexpected %#v", value)
	}
}

// TestCrypto_Sha256_Bad reports a missing argument.
func TestCrypto_Sha256_Bad(t *core.T) {
	caller := newCryptoInterpreter(t)

	if _, callErr := caller.Call("core.crypto", "sha256"); callErr == nil {
		t.Fatal("expected error for missing argument")
	}
}

// TestCrypto_Base64Decode_Ugly rejects malformed base64 and a negative size.
func TestCrypto_Base64Decode_Ugly(t *core.T) {
	caller := newCryptoInterpreter(t)

	if _, callErr := caller.Call("core.crypto", "base64_decode", "*** not base64 ***"); callErr == nil {
		t.Fatal("expected error for malformed base64")
	}
	if _, callErr := caller.Call("core.crypto", "random_bytes", -1); callErr == nil {
		t.Fatal("expected error for negative size")
	}
}
