package crypto

import (
	"crypto/hmac"
	cryptorand "crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt" // AX-6-exception: crypto helpers preserve wrapped stdlib errors from hashing/decoding/randomness.

	"dappco.re/go/py/bindings/typemap"
	"dappco.re/go/py/runtime"
)

// Register exposes hashing and encoding helpers.
//
//	crypto.Register(interpreter)
func Register(interpreter runtime.Interpreter) error {
	return interpreter.RegisterModule(runtime.Module{
		Name:          "core.crypto",
		Documentation: "Cryptographic helpers for CorePy",
		Functions: map[string]runtime.Function{
			"sha1":           sha1Digest,
			"sha256":         sha256Digest,
			"hmac_sha256":    hmacSHA256,
			"compare_digest": compareDigest,
			"base64_encode":  base64Encode,
			"base64_decode":  base64Decode,
			"random_bytes":   randomBytes,
		},
	})
}

func sha1Digest(arguments ...any) (any, error) {
	value, err := typemap.ExpectBytes(arguments, 0, "core.crypto.sha1")
	if err != nil {
		return nil, err
	}
	sum := sha1.Sum(value)
	return hex.EncodeToString(sum[:]), nil
}

func sha256Digest(arguments ...any) (any, error) {
	value, err := typemap.ExpectBytes(arguments, 0, "core.crypto.sha256")
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:]), nil
}

func hmacSHA256(arguments ...any) (any, error) {
	key, err := typemap.ExpectBytes(arguments, 0, "core.crypto.hmac_sha256")
	if err != nil {
		return nil, err
	}
	value, err := typemap.ExpectBytes(arguments, 1, "core.crypto.hmac_sha256")
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write(value); err != nil {
		return nil, fmt.Errorf("core.crypto.hmac_sha256 failed to hash input: %w", err)
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func compareDigest(arguments ...any) (any, error) {
	left, err := typemap.ExpectBytes(arguments, 0, "core.crypto.compare_digest")
	if err != nil {
		return nil, err
	}
	right, err := typemap.ExpectBytes(arguments, 1, "core.crypto.compare_digest")
	if err != nil {
		return nil, err
	}
	return hmac.Equal(left, right), nil
}

func base64Encode(arguments ...any) (any, error) {
	value, err := typemap.ExpectBytes(arguments, 0, "core.crypto.base64_encode")
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.EncodeToString(value), nil
}

func base64Decode(arguments ...any) (any, error) {
	value, err := typemap.ExpectString(arguments, 0, "core.crypto.base64_decode")
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("core.crypto.base64_decode failed to decode input: %w", err)
	}
	return decoded, nil
}

func randomBytes(arguments ...any) (any, error) {
	size, err := typemap.ExpectInt(arguments, 0, "core.crypto.random_bytes")
	if err != nil {
		return nil, err
	}
	if size < 0 {
		return nil, fmt.Errorf("core.crypto.random_bytes expected a non-negative size")
	}
	buffer := make([]byte, size)
	if _, err := cryptorand.Read(buffer); err != nil {
		return nil, fmt.Errorf("core.crypto.random_bytes failed to read randomness: %w", err)
	}
	return buffer, nil
}
