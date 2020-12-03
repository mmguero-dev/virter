package sshkeys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
)

// A KeyStore stores a single private key and the matching public key. It provides various methods to access
// both public and private key.
type KeyStore interface {
	// Authentication methods configured to use this KeyStore.
	Auth() []ssh.AuthMethod
	// The raw private key bytes, as stored on disk.
	KeyBytes() []byte
	// Path to the private key
	KeyPath() string
	// The public key bytes, as stored on disk
	PublicKey() []byte
}

type keyStore struct {
	privateKey      ssh.Signer
	privateKeyBytes []byte
	publicKeyBytes  []byte
	privateKeyPath  string
}

// Creates a new keystore by reading the private and public key from the given path
// If the paths do not exist, the keys will be created at these locations
func NewKeyStore(privateKeyPath string, publicKeyPath string) (KeyStore, error) {
	privateKey, privateKeyBuf, err := loadPrivateKeyAt(privateKeyPath)
	if err != nil {
		return nil, err
	}

	publicKey, err := loadPublicKeyAt(privateKey, publicKeyPath)
	if err != nil {
		return nil, err
	}

	return &keyStore{
		privateKey:      privateKey,
		privateKeyBytes: privateKeyBuf,
		publicKeyBytes:  publicKey,
		privateKeyPath:  privateKeyPath,
	}, nil
}

func (store *keyStore) Auth() []ssh.AuthMethod {
	return []ssh.AuthMethod{
		ssh.PublicKeys(store.privateKey),
	}
}

func (store *keyStore) KeyBytes() []byte {
	return store.privateKeyBytes
}

func (store *keyStore) KeyPath() string {
	return store.privateKeyPath
}

func (store *keyStore) PublicKey() []byte {
	return store.publicKeyBytes
}

func loadPrivateKeyAt(path string) (ssh.Signer, []byte, error) {
	exists, err := pathExists(path)
	if err != nil {
		return nil, nil, fmt.Errorf("error checking for existence of private key: %w", err)
	}

	if !exists {
		err := generatePrivateKeyAt(path)
		if err != nil {
			return nil, nil, fmt.Errorf("error generating private key: %w", err)
		}
	}

	keyBuf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading private key: %w", err)
	}

	// Ensure it can be parsed
	signer, err := ssh.ParsePrivateKey(keyBuf)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing private key: %w", err)
	}

	return signer, keyBuf, nil
}

func loadPublicKeyAt(key ssh.Signer, path string) ([]byte, error) {
	exists, err := pathExists(path)
	if err != nil {
		return nil, fmt.Errorf("error checking for existence of public key: %w", err)
	}

	if !exists {
		err = ioutil.WriteFile(path, ssh.MarshalAuthorizedKey(key.PublicKey()), 0644)
		if err != nil {
			return nil, fmt.Errorf("error writing public key: %w", err)
		}
	}

	pubBuf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading public key: %w", err)
	}

	return pubBuf, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if !os.IsNotExist(err) {
		return false, err
	}

	return false, nil
}

func generatePrivateKeyAt(path string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	pemBuf := pem.EncodeToMemory(privateKeyPEM)

	err = ioutil.WriteFile(path, pemBuf, 0600)
	if err != nil {
		return err
	}

	return nil
}
