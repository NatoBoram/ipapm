package kubo

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/ipfs/kubo/client/rpc"
)

func errorf(message string, err *rpc.Error) error {
	if strings.HasSuffix(err.Message, fs.ErrPermission.Error()) {
		return fmt.Errorf("%s (%d): %w", message, err.Code, fs.ErrPermission)
	}
	if strings.HasSuffix(err.Message, fs.ErrExist.Error()) {
		return fmt.Errorf("%s (%d): %w", message, err.Code, fs.ErrExist)
	}
	if strings.HasSuffix(err.Message, fs.ErrNotExist.Error()) {
		return fmt.Errorf("%s (%d): %w", message, err.Code, fs.ErrNotExist)
	}

	if strings.HasSuffix(err.Message, os.ErrDeadlineExceeded.Error()) {
		return fmt.Errorf("%s (%d): %w", message, err.Code, os.ErrDeadlineExceeded)
	}

	return fmt.Errorf("%s (%d): %w", message, err.Code, *err)
}
