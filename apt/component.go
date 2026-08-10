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

type ComponentsDiff struct {
	Added   Components
	Removed Components
	Changed Components
}

func (p Components) Diff(n Components) ComponentsDiff {
	pmap := make(map[string]Component, len(p))
	for _, p := range p {
		pmap[fmt.Sprintf("%s/%s", p.Name, p.Architecture)] = p
	}

	nmap := make(map[string]Component, len(n))
	for _, n := range n {
		nmap[fmt.Sprintf("%s/%s", n.Name, n.Architecture)] = n
	}

	diff := ComponentsDiff{
		Added:   make(Components, 0, len(n)),
		Removed: make(Components, 0, len(p)),
		Changed: make(Components, 0, min(len(p), len(n))),
	}

	for pk, pv := range pmap {
		nv, ok := nmap[pk]
		if !ok {
			diff.Removed = append(diff.Removed, pv)
			continue
		}

		dfiles := pv.Files.Diff(nv.Files)
		if len(dfiles.Added) > 0 || len(dfiles.Removed) > 0 || len(dfiles.Changed) > 0 {
			diff.Changed = append(diff.Changed, nv)
		}
	}

	for nk, nv := range nmap {
		if _, ok := pmap[nk]; !ok {
			diff.Added = append(diff.Added, nv)
		}
	}

	return diff
}
