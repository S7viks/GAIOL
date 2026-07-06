package keys

import (
	"os"
	"testing"
)

func TestEncryptionKeyConfigured(t *testing.T) {
	if EncryptionKeyConfigured() {
		t.Skip("GAIOL_ENCRYPTION_KEY already set in environment")
	}
	if EncryptionKeyConfigured() {
		t.Fatal("expected false when unset")
	}

	t.Setenv("GAIOL_ENCRYPTION_KEY", "not-hex")
	if EncryptionKeyConfigured() {
		t.Fatal("expected false for invalid hex")
	}

	// 32 zero bytes as hex
	t.Setenv("GAIOL_ENCRYPTION_KEY", "0000000000000000000000000000000000000000000000000000000000000000")
	if !EncryptionKeyConfigured() {
		t.Fatal("expected true for valid 32-byte hex key")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := os.Getenv("GAIOL_ENCRYPTION_KEY")
	if key == "" {
		t.Setenv("GAIOL_ENCRYPTION_KEY", "0000000000000000000000000000000000000000000000000000000000000000")
	}
	enc, err := Encrypt([]byte("secret-api-key"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	plain, err := Decrypt(enc)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(plain) != "secret-api-key" {
		t.Fatalf("round-trip = %q", plain)
	}
}
