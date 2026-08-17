package kubo

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

type AuthSecret struct {
	Token    string
	Password string
	Username string
}

func DecodeAuthSecret(secret string) (AuthSecret, error) {
	if secret == "" {
		// No secret to decode, no error to throw
		return AuthSecret{}, nil
	}

	if after, found := strings.CutPrefix(secret, "bearer:"); found {
		return AuthSecret{Token: after}, nil
	}

	if after, found := strings.CutPrefix(secret, "basic:"); found {
		parts := strings.SplitN(after, ":", 2)
		if len(parts) == 2 {
			return AuthSecret{Username: parts[0], Password: parts[1]}, nil
		}

		// Decode base64 `after`, split in two `:` then return username and password
		decoded, err := base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			return AuthSecret{}, fmt.Errorf("decoding basic secret: %w", err)
		}

		parts = strings.SplitN(string(decoded), ":", 2)
		if len(parts) == 2 {
			return AuthSecret{Username: parts[0], Password: parts[1]}, nil
		}

		return AuthSecret{}, errors.New("invalid basic secret")

	}

	return AuthSecret{Token: secret}, nil
}
