// Создаёт пару private-public ключей.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	path := flag.String("path", "/tmp", "Save t o path")
	flag.Parse()

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	pathPrivate := filepath.Join(*path, "private.pem")
	if err2 := savePrivate(privateKey, pathPrivate); err2 != nil {
		log.Fatal(err2)
	}

	pathPublic := filepath.Join(*path, "public.pem")
	if err2 := savePublic(&privateKey.PublicKey, pathPublic); err2 != nil {
		log.Fatal(err2)
	}
}

func savePrivate(key *rsa.PrivateKey, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	b := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	if _, err2 := f.Write(b); err2 != nil {
		return err2
	}
	if err2 := f.Close(); err2 != nil {
		return err2
	}

	log.Printf("saved private key to %s", path)

	return nil
}

func savePublic(key *rsa.PublicKey, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	bPublic, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return err
	}
	b := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: bPublic,
		},
	)

	if _, err2 := f.Write(b); err2 != nil {
		return err2
	}
	if err2 := f.Close(); err2 != nil {
		return err2
	}

	log.Printf("saved public key to %s", path)

	return nil
}
