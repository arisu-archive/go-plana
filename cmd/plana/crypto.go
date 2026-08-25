package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/arisu-archive/go-plana/plana"
)

const aes128KeySize = 16

func newAESKeyBundle(random io.Reader) (plana.AESKeyBundle, error) {
	key := make([]byte, aes128KeySize)
	if _, err := io.ReadFull(random, key); err != nil {
		return plana.AESKeyBundle{}, fmt.Errorf("generate AES key: %w", err)
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(random, iv); err != nil {
		return plana.AESKeyBundle{}, fmt.Errorf("generate AES IV: %w", err)
	}
	return plana.AESKeyBundle{Key: key, IV: iv}, nil
}

func decryptServerValue(key, iv []byte, encodedCiphertext string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("invalid AES IV length: got %d bytes, want %d", len(iv), block.BlockSize())
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, fmt.Errorf("decode encrypted value: %w", err)
	}
	if len(ciphertext) == 0 {
		return nil, errors.New("encrypted value is empty")
	}
	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("invalid encrypted value length: got %d bytes", len(ciphertext))
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)
	plaintext, err = unpadPKCS7(plaintext, block.BlockSize())
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(plaintext))
	if err != nil {
		return nil, fmt.Errorf("decode decrypted value: %w", err)
	}
	return decoded, nil
}

func unpadPKCS7(data []byte, blockSize int) ([]byte, error) {
	paddingLength := int(data[len(data)-1])
	if paddingLength == 0 || paddingLength > blockSize || paddingLength > len(data) {
		return nil, errors.New("decrypted value has invalid PKCS#7 padding")
	}
	for _, value := range data[len(data)-paddingLength:] {
		if int(value) != paddingLength {
			return nil, errors.New("decrypted value has invalid PKCS#7 padding")
		}
	}
	return data[:len(data)-paddingLength], nil
}
