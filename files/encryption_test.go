package files

import (
	"bytes"
	"crypto/aes"
	"testing"
)

// TestGetKey verifies the key is correct length and consistent
func TestGetKey(t *testing.T) {
	dm := &DataManager{}
	key := dm.getKey()

	// Verify key length is correct for AES-256
	if len(key) != 32 {
		t.Errorf("expected key length 32, got %d", len(key))
	}

	// Verify key is consistent across calls
	key2 := dm.getKey()
	if !bytes.Equal(key, key2) {
		t.Error("getKey() should return consistent results")
	}
}

// TestEncryptDecryptRoundtrip verifies basic encrypt/decrypt cycle
func TestEncryptDecryptRoundtrip(t *testing.T) {
	dm := &DataManager{}
	plaintext := []byte("Hello, World!")

	// Encrypt
	ciphertext, err := dm.encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt() returned error: %v", err)
	}

	// Verify ciphertext contains IV + encrypted data
	if len(ciphertext) < aes.BlockSize {
		t.Errorf("ciphertext too short: expected at least %d bytes, got %d", aes.BlockSize, len(ciphertext))
	}

	// Decrypt
	decrypted, err := dm.decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt() returned error: %v", err)
	}

	// Verify decrypted data matches original
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted data doesn't match plaintext.\nGot: %s\nExpected: %s", string(decrypted), string(plaintext))
	}
}

// TestEncryptProducesDifferentCiphertexts ensures random IV generates different outputs
func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	dm := &DataManager{}
	plaintext := []byte("Test data")

	// Encrypt the same plaintext twice
	ciphertext1, err := dm.encrypt(plaintext)
	if err != nil {
		t.Fatalf("first encrypt() returned error: %v", err)
	}

	ciphertext2, err := dm.encrypt(plaintext)
	if err != nil {
		t.Fatalf("second encrypt() returned error: %v", err)
	}

	// Ciphertexts should be different due to random IV
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("two encryptions of the same plaintext should produce different ciphertexts (random IV)")
	}
}

// TestDecryptWithShortCiphertext verifies error handling for invalid input
func TestDecryptWithShortCiphertext(t *testing.T) {
	dm := &DataManager{}
	shortData := []byte("short")

	_, err := dm.decrypt(shortData)
	if err == nil {
		t.Error("decrypt() should return an error for ciphertext shorter than block size")
	}

	expectedErr := "ciphertext too short"
	if err.Error() != expectedErr {
		t.Errorf("expected error message %q, got %q", expectedErr, err.Error())
	}
}

// TestDecryptWithEmptyCiphertext handles empty input edge case
func TestDecryptWithEmptyCiphertext(t *testing.T) {
	dm := &DataManager{}

	_, err := dm.decrypt([]byte{})
	if err == nil {
		t.Error("decrypt() should return an error for empty ciphertext")
	}
}

// TestEncryptEmptyPlaintext handles empty plaintext
func TestEncryptEmptyPlaintext(t *testing.T) {
	dm := &DataManager{}
	plaintext := []byte{}

	ciphertext, err := dm.encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt() returned error: %v", err)
	}

	// Should still contain IV
	if len(ciphertext) != aes.BlockSize {
		t.Errorf("empty plaintext encryption should produce IV only, got %d bytes", len(ciphertext))
	}

	// Should decrypt back to empty plaintext
	decrypted, err := dm.decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt() returned error: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("decrypted empty plaintext mismatch")
	}
}

// TestEncryptSingleByte handles single byte plaintext
func TestEncryptSingleByte(t *testing.T) {
	dm := &DataManager{}
	plaintext := []byte{0x42}

	ciphertext, err := dm.encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt() returned error: %v", err)
	}

	decrypted, err := dm.decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt() returned error: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("single byte roundtrip failed")
	}
}

// TestEncryptLargePlaintext verifies performance with large data
func TestEncryptLargePlaintext(t *testing.T) {
	dm := &DataManager{}

	// Create large plaintext (1 MB)
	plaintext := make([]byte, 1024*1024)
	for i := 0; i < len(plaintext); i++ {
		plaintext[i] = byte(i % 256)
	}

	ciphertext, err := dm.encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt() returned error: %v", err)
	}

	if len(ciphertext) != len(plaintext)+aes.BlockSize {
		t.Errorf("ciphertext length mismatch: expected %d, got %d", len(plaintext)+aes.BlockSize, len(ciphertext))
	}

	decrypted, err := dm.decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt() returned error: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("large plaintext roundtrip failed")
	}
}

// TestDecryptWithTamperedData demonstrates lack of authentication
func TestDecryptWithTamperedData(t *testing.T) {
	dm := &DataManager{}
	plaintext := []byte("Sensitive data")

	ciphertext, err := dm.encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt() returned error: %v", err)
	}

	// Tamper with a byte in the ciphertext (not the IV)
	if len(ciphertext) > aes.BlockSize {
		ciphertext[aes.BlockSize]++
	}

	decrypted, err := dm.decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt() returned error: %v", err)
	}

	// Decryption succeeds but produces wrong plaintext (CTR mode doesn't authenticate)
	if bytes.Equal(plaintext, decrypted) {
		t.Error("tampering should produce different plaintext")
	}
}

// TestEncryptDecryptBinaryData tests various binary patterns
func TestEncryptDecryptBinaryData(t *testing.T) {
	dm := &DataManager{}

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"zeros", []byte{0x00, 0x00, 0x00, 0x00}},
		{"ones", []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{"mixed", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}},
		{"single zero", []byte{0x00}},
		{"single max", []byte{0xFF}},
		{"ascending", []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}},
		{"descending", []byte{0xFF, 0xFE, 0xFD, 0xFC, 0xFB, 0xFA}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := dm.encrypt(tc.plaintext)
			if err != nil {
				t.Fatalf("encrypt() returned error: %v", err)
			}

			decrypted, err := dm.decrypt(ciphertext)
			if err != nil {
				t.Fatalf("decrypt() returned error: %v", err)
			}

			if !bytes.Equal(tc.plaintext, decrypted) {
				t.Errorf("roundtrip failed: expected %v, got %v", tc.plaintext, decrypted)
			}
		})
	}
}

// TestEncryptMultipleManagers verifies different managers use same key
func TestEncryptMultipleManagers(t *testing.T) {
	dm1 := &DataManager{}
	dm2 := &DataManager{}

	plaintext := []byte("Test data")

	ciphertext, err := dm1.encrypt(plaintext)
	if err != nil {
		t.Fatalf("dm1.encrypt() returned error: %v", err)
	}

	// dm2 should be able to decrypt dm1's ciphertext (same key)
	decrypted, err := dm2.decrypt(ciphertext)
	if err != nil {
		t.Fatalf("dm2.decrypt() returned error: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("cross-manager decryption failed")
	}
}

// TestDecryptExactBlockSize handles data exactly block size
func TestDecryptExactBlockSize(t *testing.T) {
	dm := &DataManager{}

	// Create ciphertext exactly block size (IV only, no data)
	ciphertext := make([]byte, aes.BlockSize)

	decrypted, err := dm.decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt() returned error: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("expected empty decryption, got %d bytes", len(decrypted))
	}
}

// TestEncryptDecryptMultipleRoundtrips verifies consistency across multiple cycles
func TestEncryptDecryptMultipleRoundtrips(t *testing.T) {
	dm := &DataManager{}
	original := []byte("Multi-cycle test data")

	current := original
	for i := 0; i < 5; i++ {
		encrypted, err := dm.encrypt(current)
		if err != nil {
			t.Fatalf("iteration %d: encrypt() returned error: %v", i, err)
		}

		decrypted, err := dm.decrypt(encrypted)
		if err != nil {
			t.Fatalf("iteration %d: decrypt() returned error: %v", i, err)
		}

		if !bytes.Equal(current, decrypted) {
			t.Errorf("iteration %d: roundtrip failed", i)
		}

		// Use decrypted data for next iteration
		current = decrypted
	}

	if !bytes.Equal(original, current) {
		t.Error("final result doesn't match original after multiple roundtrips")
	}
}

// BenchmarkEncrypt measures encryption performance
func BenchmarkEncrypt(b *testing.B) {
	dm := &DataManager{}
	plaintext := []byte("This is test data for benchmarking encryption performance.")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dm.encrypt(plaintext)
		if err != nil {
			b.Fatalf("encrypt() returned error: %v", err)
		}
	}
}

// BenchmarkDecrypt measures decryption performance
func BenchmarkDecrypt(b *testing.B) {
	dm := &DataManager{}
	plaintext := []byte("This is test data for benchmarking decryption performance.")

	ciphertext, err := dm.encrypt(plaintext)
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dm.decrypt(ciphertext)
		if err != nil {
			b.Fatalf("decrypt() returned error: %v", err)
		}
	}
}

// BenchmarkEncryptLarge measures performance with 1MB data
func BenchmarkEncryptLarge(b *testing.B) {
	dm := &DataManager{}
	plaintext := make([]byte, 1024*1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dm.encrypt(plaintext)
		if err != nil {
			b.Fatalf("encrypt() returned error: %v", err)
		}
	}
}

// BenchmarkDecryptLarge measures performance with 1MB data
func BenchmarkDecryptLarge(b *testing.B) {
	dm := &DataManager{}
	plaintext := make([]byte, 1024*1024)

	ciphertext, err := dm.encrypt(plaintext)
	if err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dm.decrypt(ciphertext)
		if err != nil {
			b.Fatalf("decrypt() returned error: %v", err)
		}
	}
}
