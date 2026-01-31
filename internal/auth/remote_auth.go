package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
)

type RemoteAuthSession struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  []byte // PEM Encoded
}

func NewRemoteAuthSession() (*RemoteAuthSession, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	pubDer, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}

	pubBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDer,
	}
	pubPem := pem.EncodeToMemory(&pubBlock)

	return &RemoteAuthSession{
		PrivateKey: priv,
		PublicKey:  pubPem,
	}, nil
}

func (s *RemoteAuthSession) Decrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, s.PrivateKey, ciphertext, nil)
}
