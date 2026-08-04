package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/clearsign"
)

func loadPgp(signedBy string) ([]byte, error) {
	if strings.Contains(signedBy, "-----BEGIN PGP PUBLIC KEY BLOCK-----") {
		return []byte(signedBy), nil
	}

	key, err := os.ReadFile(signedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file %s: %w", signedBy, err)
	}

	return key, nil
}

func readPgp(key []byte) (openpgp.EntityList, error) {
	keyring1, err1 := openpgp.ReadArmoredKeyRing(bytes.NewReader(key))
	if err1 == nil {
		return keyring1, nil
	}

	keyring2, err2 := openpgp.ReadKeyRing(bytes.NewReader(key))
	if err2 == nil {
		return keyring2, nil
	}

	return nil, fmt.Errorf("failed to read keyring: %v", errors.Join(err1, err2))
}

func checkPgp(message string, keyring openpgp.EntityList) error {
	block, _ := clearsign.Decode([]byte(message))
	if block == nil {
		return fmt.Errorf("invalid OpenPGP clear-signed data")
	}

	signature := block.ArmoredSignature.Body
	signed := bytes.NewReader(block.Bytes)

	_, err := openpgp.CheckDetachedSignature(keyring, signed, signature, nil)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

func verifyPgp(signedBy, message string) error {
	key, err := loadPgp(signedBy)
	if err != nil {
		return fmt.Errorf("failed to load PGP key: %w", err)
	}

	keyring, err := readPgp(key)
	if err != nil {
		return fmt.Errorf("failed to read PGP key: %w", err)
	}

	err = checkPgp(message, keyring)
	if err != nil {
		return fmt.Errorf("failed to check PGP signature: %w", err)
	}

	return nil
}
