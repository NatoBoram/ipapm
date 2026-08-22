package kubo

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/ipfs/kubo/client/rpc"
)

func errorf(message string, err *rpc.Error) error {
	if before, ok := strings.CutSuffix(err.Message, fs.ErrPermission.Error()); ok {
		return fmt.Errorf("%s (%d): %s%w", message, err.Code, before, fs.ErrPermission)
	}
	if before, ok := strings.CutSuffix(err.Message, fs.ErrExist.Error()); ok {
		return fmt.Errorf("%s (%d): %s%w", message, err.Code, before, fs.ErrExist)
	}
	if before, ok := strings.CutSuffix(err.Message, fs.ErrNotExist.Error()); ok {
		return fmt.Errorf("%s (%d): %s%w", message, err.Code, before, fs.ErrNotExist)
	}

	if before, ok := strings.CutSuffix(err.Message, os.ErrDeadlineExceeded.Error()); ok {
		return fmt.Errorf("%s (%d): %s%w", message, err.Code, before, os.ErrDeadlineExceeded)
	}

	return fmt.Errorf("%s (%d): %w", message, err.Code, *err)
}
