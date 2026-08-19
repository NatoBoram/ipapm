package http

import (
	"fmt"
	"net/http"
)

var ErrNotFound = fmt.Errorf("%d %s", http.StatusNotFound, http.StatusText(http.StatusNotFound))
