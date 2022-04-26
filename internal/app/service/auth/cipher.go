package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"math/rand"
	"sync"
	"time"
)

type cipherData struct {
	key    []byte
	nonce  []byte
	aesGCM cipher.AEAD
}

var cipherInstance *cipherData
var once sync.Once

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateRandom(size int) []byte {
	var Letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
	b := make([]byte, size)
	for i := range b {
		b[i] = Letters[rand.Intn(len(Letters))]
	}
	return b
}

func cipherInit() error {
	var e error
	once.Do(func() {
		key := generateRandom(2 * aes.BlockSize)

		aesblock, err := aes.NewCipher(key)
		if err != nil {
			e = err
		}

		aesgcm, err := cipher.NewGCM(aesblock)
		if err != nil {
			e = err
		}

		nonce := generateRandom(aesgcm.NonceSize())
		cipherInstance = &cipherData{key: key, aesGCM: aesgcm, nonce: nonce}
	})
	return e
}

func EncryptPassword(password string) (string, error) {
	if err := cipherInit(); err != nil {
		return "", err
	}
	encrypted := cipherInstance.aesGCM.Seal(nil, cipherInstance.nonce, []byte(password), nil)
	return hex.EncodeToString(encrypted), nil
}

func DecryptPassword(password string) (string, error) {
	if err := cipherInit(); err != nil {
		return "", err
	}

	b, err := hex.DecodeString(password)
	if err != nil {
		return "", err
	}

	decrypted, err := cipherInstance.aesGCM.Open(nil, cipherInstance.nonce, b, nil)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
