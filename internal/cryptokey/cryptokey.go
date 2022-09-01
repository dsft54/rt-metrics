// Package cryptokey implents rsa encryption/decryption of byte slice by ssh keypair
package cryptokey

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

func ParsePublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parsed, _, _, _, err := ssh.ParseAuthorizedKey(keyData)
	if err != nil {
		return nil, err
	}
	parsedCryptoKey := parsed.(ssh.CryptoPublicKey)
	pubCrypto := parsedCryptoKey.CryptoPublicKey()
	return pubCrypto.(*rsa.PublicKey), nil
}

func ParsePrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	private, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return private.(*rsa.PrivateKey), nil
}

func EncryptMessage(data []byte, pub *rsa.PublicKey) ([]byte, error) {
	if pub == nil {
		return nil, errors.New("nil pub key")
	}
	hash := sha256.New()
	msgLen := len(data)
	step := pub.Size() - 2*hash.Size() - 2
	var encryptedBytes []byte
	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}
		encryptedBlock, err := rsa.EncryptOAEP(hash, rand.Reader, pub, data[start:finish], nil)
		if err != nil {
			return nil, err
		}
		encryptedBytes = append(encryptedBytes, encryptedBlock...)
	}
	return encryptedBytes, nil
}

func DecryptMessage(data []byte, private *rsa.PrivateKey, step int) ([]byte, error) {
	if private == nil {
		return nil, errors.New("nil private key")
	}
	hash := sha256.New()
	msgLen := len(data)
	var decryptedBytes []byte
	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}
		decryptedBlock, err := rsa.DecryptOAEP(hash, rand.Reader, private, data[start:finish], nil)
		if err != nil {
			return nil, err
		}
		decryptedBytes = append(decryptedBytes, decryptedBlock...)
	}
	return decryptedBytes, nil
}
