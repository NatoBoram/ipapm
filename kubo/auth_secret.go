package kubo

import (
	"encoding/base64"
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

		if len(parts) != 1 {
			return AuthSecret{}, fmt.Errorf("error splitting basic secret")
		}

		// Decode base64 after, split in two `:` then return username and password
		decoded, err := base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			return AuthSecret{}, fmt.Errorf("couldn't decode basic secret")
		}

		parts = strings.SplitN(string(decoded), ":", 2)
		if len(parts) == 2 {
			return AuthSecret{Username: parts[0], Password: parts[1]}, nil
		}

		return AuthSecret{}, fmt.Errorf("invalid basic secret")

	}

	return AuthSecret{Token: secret}, nil
}
