package apt

import (
	"fmt"
	"net/url"

	"github.com/NatoBoram/ipapm/config"
	"github.com/NatoBoram/ipapm/wheel"
)

type Configs map[string]Config

type Config struct {
	URI      *url.URL
	Suites   wheel.Set[string]
	SignedBy string
}

// MapConfigs turns [config.Source] into a [Configs]. Since a source can have
// multiple URIs and those can collide between different sources, we need to
// know which URI has which suites.
func MapConfigs(configs []config.Source) (Configs, error) {
	mapped := make(Configs)

	for _, source := range configs {
		for _, uri := range source.URIs {
			if m, ok := mapped[uri]; ok {
				if m.SignedBy != source.SignedBy {
					return nil, fmt.Errorf("conflicting SignedBy for %s: %s vs %s", uri, m.SignedBy, source.SignedBy)
				}

				m.Suites.Add(source.Suites...)
				mapped[uri] = m
				continue
			}

			url, err := url.ParseRequestURI(uri)
			if err != nil {
				return nil, fmt.Errorf("invalid uri: %w", err)
			}

			mapped[uri] = Config{
				URI:      url,
				Suites:   wheel.NewSet(source.Suites...),
				SignedBy: source.SignedBy,
			}
		}
	}

	return mapped, nil
}
