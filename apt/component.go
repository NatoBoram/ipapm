package apt

import (
	"fmt"
	"strings"
)

// Components is a [Component] slice.
type Components []Component

// Component is a component/architecture pair in an [InRelease] file.
type Component struct {
	Name         string
	Architecture string
	Files        FileHashes
}

// ByComponents slices an [InRelease] file into component/architecture pairs.
func (i InRelease) ByComponents(files FileHashes) Components {
	result := make(Components, 0, len(i.Components)*len(i.Architectures))

	for _, c := range i.Components {
		for _, a := range i.Architectures {
			prefix := archPrefix(c, a)

			match := make(FileHashes)
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
