package apt

import "fmt"

type SourcesDiff struct {
	Added   Sources
	Removed Sources
	Changed Sources
}

func (p Sources) Diff(n Sources) (SourcesDiff, error) {
	pmap := make(map[string]Source, len(p))
	for _, p := range p {
		pmap[p.Directory] = p
	}

	nmap := make(map[string]Source, len(n))
	for _, n := range n {
		nmap[n.Directory] = n
	}

	diff := SourcesDiff{
		Added:   make(Sources, 0, len(n)),
		Removed: make(Sources, 0, len(p)),
		Changed: make(Sources, 0, min(len(p), len(n))),
	}

	// From previous to next
	for pk, pv := range pmap {
		nv, ok := nmap[pk]
		if !ok {
			diff.Removed = append(diff.Removed, pv)
			continue
		}

		ph, err := pv.FileHashes()
		if err != nil {
			return SourcesDiff{}, fmt.Errorf("failed to get files from previous source %s: %w", pk, err)
		}

		nh, err := nv.FileHashes()
		if err != nil {
			return SourcesDiff{}, fmt.Errorf("failed to get files from next source %s: %w", pk, err)
		}

		fhd := ph.Diff(nh)
		if len(fhd.Added) > 0 || len(fhd.Removed) > 0 || len(fhd.Changed) > 0 {
			diff.Changed = append(diff.Changed, nv)
		}
	}

	// New in next
	for nk, nv := range nmap {
		if _, ok := pmap[nk]; !ok {
			diff.Added = append(diff.Added, nv)
		}
	}

	return diff, nil
}
