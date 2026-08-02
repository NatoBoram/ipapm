package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NatoBoram/ipapm/api"
)

func TestNew_readyz(t *testing.T) {
	handler := api.New(api.Config{
		Kubo: new(Kubo{}),
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Errorf("Expected %d, got %d", http.StatusOK, response.Code)
	}
}
