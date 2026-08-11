package apt_test

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/NatoBoram/ipapm/apt"
	"github.com/NatoBoram/ipapm/config"
	"github.com/NatoBoram/ipapm/wheel"
)

func TestMapConfigs(t *testing.T) {
	sources := []config.Source{
		{
			URIs:     []string{"https://example.org"},
			Suites:   []string{"noble"},
			SignedBy: "/etc/apt/trusted.gpg.d/example.gpg",
		},
		{
			URIs:     []string{"https://example.org"},
			Suites:   []string{"noble"},
			SignedBy: "/etc/apt/trusted.gpg.d/example.gpg",
		},
		{
			URIs:     []string{"https://example.com"},
			Suites:   []string{"noble"},
			SignedBy: "/etc/apt/trusted.gpg.d/example.gpg",
		},
	}

	merged, err := apt.MapConfigs(sources)
	if err != nil {
		t.Fatalf("MapSources returned an error: %v", err)
	}

	expected := apt.Configs{
		"https://example.org": {
			URI:      new(url.URL{Scheme: "https", Host: "example.org"}),
			Suites:   wheel.Set[string]{"noble": {}},
			SignedBy: "/etc/apt/trusted.gpg.d/example.gpg",
		},
		"https://example.com": {
			URI:      new(url.URL{Scheme: "https", Host: "example.com"}),
			Suites:   wheel.Set[string]{"noble": {}},
			SignedBy: "/etc/apt/trusted.gpg.d/example.gpg",
		},
	}
	if eq := reflect.DeepEqual(merged, expected); !eq {
		t.Errorf("Expected %v, got %v", expected, merged)
	}
}
