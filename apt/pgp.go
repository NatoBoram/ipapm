package apt

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
		return nil, fmt.Errorf("reading key file %q: %w", signedBy, err)
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

	return nil, fmt.Errorf("reading keyring: %w", errors.Join(err1, err2))
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
		return fmt.Errorf("verifying signature: %w", err)
	}

	return nil
}

func VerifyPgp(signedBy, message string) error {
	key, err := loadPgp(signedBy)
	if err != nil {
		return fmt.Errorf("loading PGP key: %w", err)
	}

	keyring, err := readPgp(key)
	if err != nil {
		return fmt.Errorf("reading PGP key: %w", err)
	}

	err = checkPgp(message, keyring)
	if err != nil {
		return fmt.Errorf("checking PGP signature: %w", err)
	}

	return nil
}
