package otto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// AES256KeySize is the required key size for AES-256 encryption (32 bytes = 256 bits)
const AES256KeySize = 32

// Generates a cryptographically secure random 32-byte key for AES-256
func GenerateKey() ([]byte, error) {
	key := make([]byte, AES256KeySize)
	if _, err := rand.Reader.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random encryption key: %w", err)
	}
	return key, nil
}

// Ensures the key is exactly 32 bytes for AES-256
func validateKey(key []byte) error {
	if len(key) != AES256KeySize {
		return fmt.Errorf("invalid key size: expected %d bytes, got %d bytes", AES256KeySize, len(key))
	}
	return nil
}

func getBlockFromKey(key []byte) (cipher.AEAD, error) {
	// Validate key size before proceeding
	if err := validateKey(key); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	return gcm, nil
}

func Encrypt(data []byte, key []byte) ([]byte, error) {
	// Validate inputs
	if len(data) == 0 {
		return nil, errors.New("cannot encrypt empty data")
	}

	// Get GCM cipher - MUST check error
	gcm, err := getBlockFromKey(key)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data and prepend nonce
	bytesEncrypted := gcm.Seal(nonce, nonce, data, nil)
	return bytesEncrypted, nil
}

func Decrypt(data []byte, key []byte) ([]byte, error) {
	// Get GCM cipher - MUST check error
	gcm, err := getBlockFromKey(key)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Validate data length
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short: expected at least %d bytes, got %d bytes", nonceSize, len(data))
	}

	// Extract nonce and decrypt
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	bytesDecrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return bytesDecrypted, nil
}

func BytesToStringHex(data []byte) string {
	return hex.EncodeToString(data)
}

func StringHexToBytes(data string) ([]byte, error) {
	return hex.DecodeString(data)
}
