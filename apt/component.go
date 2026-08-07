package apt

import (
	"fmt"
	"strings"
)

type Component struct {
	Name         string
	Architecture string
	Files        InReleaseFiles
}

func (i InRelease) ByComponents(files InReleaseFiles) []Component {
	result := make([]Component, 0, len(i.Components)*len(i.Architectures))

	for _, c := range i.Components {
		for _, a := range i.Architectures {
			prefix := archPrefix(c, a)

			match := make(InReleaseFiles)
			for p, f := range files {
				if strings.HasPrefix(p, prefix) {
					match[p] = f
				}
			}

			if len(match) == 0 {
				continue
			}

			result = append(result, Component{
				Name:         c,
				Architecture: a,
				Files:        match,
			})
		}
	}

	return result
}

func archPrefix(component, architecture string) string {
	if architecture == "source" {
		return fmt.Sprintf("%s/source/", component)
	}
	return fmt.Sprintf("%s/binary-%s/", component, architecture)
}
