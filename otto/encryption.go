package otto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// key is 32 bytes (1byte = 8bits * 32 = 256bits for AES-256)
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Reader.Read(key); err != nil {
		fmt.Println("error generating random encryption key", err)
		return []byte{}, err
	}
	return key, nil
}

func getBlockFromKey(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("error creating aes block cipher", err)
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("error setting gcm mode", err)
		return nil, err
	}
	return gcm, nil
}

func Encrypt(data []byte, key []byte) ([]byte, error) {
	gcm, _ := getBlockFromKey(key)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("error generating nonce ", err)
		return []byte{}, err
	}
	bytesEncrypted := gcm.Seal(nonce, nonce, data, nil)
	return bytesEncrypted, nil
}

func Decrypt(data []byte, key []byte) ([]byte, error) {
	gcm, _ := getBlockFromKey(key)
	bytesDecrypted, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
	if err != nil {
		return []byte{}, err
	}
	return bytesDecrypted, nil
}

func BytesToStringHex(data []byte) string {
	return hex.EncodeToString(data)
}

func StringHexToBytes(data string) ([]byte, error) {
	return hex.DecodeString(data)
}
