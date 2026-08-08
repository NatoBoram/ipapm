package apt

import (
	"fmt"

	"github.com/NatoBoram/ipapm/wheel"
)

type FileHashesDiff struct {
	Added   FileHashes
	Removed wheel.Set[string]
	Changed FileHashes
}

func (p FileHashes) Diff(n FileHashes) FileHashesDiff {
	o := FileHashesDiff{
		Added:   make(FileHashes, len(n)),
		Removed: wheel.MakeSet[string](len(p)),
		Changed: make(FileHashes, min(len(p), len(n))),
	}

	// From previous to next
	for pk, pv := range p {
		nv, ok := n[pk]
		if !ok {
			o.Removed.Add(pk)
			continue
		}

		if nv.MD5Sum != pv.MD5Sum ||
			nv.SHA1 != pv.SHA1 || nv.SHA256 != pv.SHA256 || nv.SHA512 != pv.SHA512 ||
			nv.Size != pv.Size {
			o.Changed[pk] = nv
		}
	}

	// New in next
	for nk, nv := range n {
		if _, ok := p[nk]; !ok {
			o.Added[nk] = nv
		}
	}

	return o
}

func (p InRelease) Diff(n InRelease) (FileHashesDiff, error) {
	pfiles, err := p.Files()
	if err != nil {
		return FileHashesDiff{}, fmt.Errorf("failed to get files from previous InRelease: %w", err)
	}

	nfiles, err := n.Files()
	if err != nil {
		return FileHashesDiff{}, fmt.Errorf("failed to get files from next InRelease: %w", err)
	}

	return pfiles.Diff(nfiles), nil
}

type filesTarget struct {
	sums    []InReleaseSum
	setHash func(file *FileHash, hash string)
}
