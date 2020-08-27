package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/tmax-cloud/l2c-operator/internal"
)

// Encryption Util (AES + Base64)

const EncKeySize = 32

func EncryptPassword(str string) (string, error) {
	c, err := aes.NewCipher([]byte(encryptKey()))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(str), nil)), nil
}

func DecryptPassword(s string) (string, error) {
	str, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	c, err := aes.NewCipher([]byte(encryptKey()))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(str) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, text := str[:nonceSize], str[nonceSize:]
	result, err := gcm.Open(nil, nonce, text, nil)
	return string(result), err
}

func IsEncrypted(s string) bool {
	_, err := DecryptPassword(s)
	return err == nil
}

func encryptKey() string {
	s := internal.EncryptKey
	if len(s) < EncKeySize {
		return s + strings.Repeat("l", EncKeySize-len(s))
	} else if len(s) > EncKeySize {
		return s[:EncKeySize]
	}
	return s
}
