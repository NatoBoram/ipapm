package apt

import "fmt"

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
