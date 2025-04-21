package otto

import (
	"errors"
	"fmt"
)

// Please note: This is not encryption or safe at all
// This is designed to provide basic protections on analysis tools like "strings"
// If you want true protections, look at encryption.go (requires providing a secret key)

func EncodeStrings(input []string, key string) ([]string, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("input slice is empty")
	}
	if key == "" {
		return nil, errors.New("blank key provided")
	}
	keyBytes := []byte(key)

	result := make([]string, len(input))
	for i, str := range input {
		if str == "" {
			return nil, fmt.Errorf("empty string at index %d", i)
		}
		// simple XOR of each byte
		encoded := make([]byte, len(str))
		for j, char := range []byte(str) {
			encoded[j] = char ^ keyBytes[j%len(keyBytes)]
		}
		// Convert to hex
		result[i] = fmt.Sprintf("%x", encoded)
	}
	return result, nil
}

func DecodeStrings(input []string, key string) ([]string, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("input slice is empty")
	}
	if key == "" {
		return nil, errors.New("blank key provided")
	}
	keyBytes := []byte(key)

	result := make([]string, len(input))
	for i, str := range input {
		if str == "" {
			return nil, fmt.Errorf("empty string at index %d", i)
		}
		// Decode hex string
		hexBytes := make([]byte, len(str)/2)
		_, err := fmt.Sscanf(str, "%x", &hexBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid hex string at index %d: %v", i, err)
		}
		// Undo XOR
		decoded := make([]byte, len(hexBytes))
		for j, char := range hexBytes {
			decoded[j] = char ^ keyBytes[j%len(keyBytes)]
		}
		result[i] = string(decoded)
	}
	return result, nil
}
