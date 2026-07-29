package api

import "net/http"

func livez(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNoContent)
}
