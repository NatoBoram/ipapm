package apt

import (
	"log"
	"net/url"

	"github.com/NatoBoram/ipapm/config"
	"github.com/NatoBoram/ipapm/wheel"
)

type MapSource map[string]Source

type Source struct {
	Types         wheel.Set[string]
	URI           *url.URL
	Suites        wheel.Set[string]
	Components    wheel.Set[string]
	Architectures wheel.Set[string]
	SignedBy      string
}

// MapSources turns [config.Source] into a [MapSource]. Since a source can have
// multiple URIs and they can collide between different sources, we need to know
// which URI has which parameters.
func MapSources(sources []config.Source) MapSource {
	mapped := make(MapSource)

	for _, source := range sources {
		for _, uri := range source.URIs {
			if m, ok := mapped[uri]; ok {
				m.Types.Add(source.Types...)
				m.Suites.Add(source.Suites...)
				m.Components.Add(source.Components...)
				m.Architectures.Add(source.Architectures...)
				mapped[uri] = m
				continue
			}

			url, err := url.ParseRequestURI(uri)
			if err != nil {
				log.Printf("Invalid uri: %s", err)
				continue
			}

			mapped[uri] = Source{
				Types:         wheel.NewSet(source.Types...),
				URI:           url,
				Suites:        wheel.NewSet(source.Suites...),
				Components:    wheel.NewSet(source.Components...),
				Architectures: wheel.NewSet(source.Architectures...),
				SignedBy:      source.SignedBy,
			}
		}
	}

	return mapped
}
