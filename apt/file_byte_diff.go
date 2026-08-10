package apt

type FileBytesDiff struct {
	Added   FileBytes
	Removed FileBytes
	Changed FileBytes
}

func (p FileBytes) Diff(n FileBytes) FileBytesDiff {
	pmap := make(map[string]FileByte, len(p))
	for _, p := range p {
		pmap[p.Hashes.Filename] = p
	}

	nmap := make(map[string]FileByte, len(n))
	for _, n := range n {
		nmap[n.Hashes.Filename] = n
	}

	diff := FileBytesDiff{
		Added:   make(FileBytes, 0, len(n)),
		Removed: make(FileBytes, 0, len(p)),
		Changed: make(FileBytes, 0, min(len(p), len(n))),
	}

	// From previous to next
	for pk, pv := range pmap {
		nv, ok := nmap[pk]
		if !ok {
			diff.Removed = append(diff.Removed, pv)
			continue
		}

		if nv.Hashes.MD5Sum != pv.Hashes.MD5Sum ||
			nv.Hashes.SHA1 != pv.Hashes.SHA1 || nv.Hashes.SHA256 != pv.Hashes.SHA256 || nv.Hashes.SHA512 != pv.Hashes.SHA512 ||
			nv.Hashes.Size != pv.Hashes.Size {
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
