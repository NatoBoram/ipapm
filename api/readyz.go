package api

import "net/http"

func readyz(config Config) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		_, err := config.Kubo.Version(ctx)
		if err != nil {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		response.WriteHeader(http.StatusNoContent)
	}
}
