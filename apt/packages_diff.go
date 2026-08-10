package apt

type PackagesDiff struct {
	Added   Packages
	Removed Packages
	Changed Packages
}

func (p Packages) Diff(n Packages) PackagesDiff {
	pmap := make(map[string]Package, len(p))
	for _, p := range p {
		pmap[p.Filename] = p
	}

	nmap := make(map[string]Package, len(n))
	for _, n := range n {
		nmap[n.Filename] = n
	}

	diff := PackagesDiff{
		Added:   make(Packages, 0, len(n)),
		Removed: make(Packages, 0, len(p)),
		Changed: make(Packages, 0, min(len(p), len(n))),
	}

	// From previous to next
	for pk, pv := range pmap {
		nv, ok := nmap[pk]
		if !ok {
			diff.Removed = append(diff.Removed, pv)
			continue
		}

		if nv.MD5sum != pv.MD5sum ||
			nv.SHA1 != pv.SHA1 || nv.SHA256 != pv.SHA256 || nv.SHA512 != pv.SHA512 ||
			nv.Size != pv.Size {
			diff.Changed = append(diff.Changed, nv)
		}
	}

	// New in next
	for nk, nv := range nmap {
		if _, ok := pmap[nk]; !ok {
			diff.Added = append(diff.Added, nv)
		}
	}

	return diff
}
