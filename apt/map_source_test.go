package apt_test

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/NatoBoram/ipapm/apt"
	"github.com/NatoBoram/ipapm/config"
	"github.com/NatoBoram/ipapm/wheel"
)

func TestMapSources(t *testing.T) {
	sources := []config.Source{
		{
			Types:         []string{"deb"},
			URIs:          []string{"https://example.org"},
			Suites:        []string{"noble"},
			Components:    []string{"main"},
			Architectures: []string{"amd64"},
			SignedBy:      "/etc/apt/trusted.gpg.d/example.gpg",
		},
		{
			Types:         []string{"deb-src"},
			URIs:          []string{"https://example.org"},
			Suites:        []string{"noble"},
			Components:    []string{"main"},
			Architectures: []string{"amd64"},
			SignedBy:      "/etc/apt/trusted.gpg.d/example.gpg",
		},
		{
			Types:         []string{"deb"},
			URIs:          []string{"https://example.com"},
			Suites:        []string{"noble"},
			Components:    []string{"main"},
			Architectures: []string{"amd64"},
			SignedBy:      "/etc/apt/trusted.gpg.d/example.gpg",
		},
	}

	merged := apt.MapSources(sources)
	expected := apt.MapSource{
		"https://example.org": {
			Types:         wheel.Set[string]{"deb": {}, "deb-src": {}},
			URI:           new(url.URL{Scheme: "https", Host: "example.org"}),
			Suites:        wheel.Set[string]{"noble": {}},
			Components:    wheel.Set[string]{"main": {}},
			Architectures: wheel.Set[string]{"amd64": {}},
			SignedBy:      "/etc/apt/trusted.gpg.d/example.gpg",
		},
		"https://example.com": {
			Types:         wheel.Set[string]{"deb": {}},
			URI:           new(url.URL{Scheme: "https", Host: "example.com"}),
			Suites:        wheel.Set[string]{"noble": {}},
			Components:    wheel.Set[string]{"main": {}},
			Architectures: wheel.Set[string]{"amd64": {}},
			SignedBy:      "/etc/apt/trusted.gpg.d/example.gpg",
		},
	}
	if eq := reflect.DeepEqual(merged, expected); !eq {
		t.Errorf("Expected %v, got %v", expected, merged)
	}
}
