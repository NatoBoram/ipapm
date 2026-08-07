package apt

import (
	"fmt"

	"github.com/NatoBoram/ipapm/wheel"
)

type InReleaseFiles map[string]InReleaseFile

type InReleaseDiff struct {
	Added   InReleaseFiles
	Removed wheel.Set[string]
	Changed InReleaseFiles
}

func (p InReleaseFiles) Diff(n InReleaseFiles) InReleaseDiff {
	o := InReleaseDiff{
		Added:   make(InReleaseFiles, len(n)),
		Removed: wheel.MakeSet[string](len(p)),
		Changed: make(InReleaseFiles, min(len(p), len(n))),
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

func (p InRelease) Diff(n InRelease) (InReleaseDiff, error) {
	pfiles, err := p.Files()
	if err != nil {
		return InReleaseDiff{}, fmt.Errorf("failed to get files from previous InRelease: %w", err)
	}

	nfiles, err := n.Files()
	if err != nil {
		return InReleaseDiff{}, fmt.Errorf("failed to get files from next InRelease: %w", err)
	}

	return pfiles.Diff(nfiles), nil
}

type InReleaseFile struct {
	MD5Sum string
	SHA1   string
	SHA256 string
	SHA512 string

	Size uint
	Path string
}

type filesTarget struct {
	sums    []InReleaseSum
	setHash func(file *InReleaseFile, hash string)
}

func (i InRelease) Files() (InReleaseFiles, error) {
	size := max(len(i.MD5Sum), len(i.SHA1), len(i.SHA256), len(i.SHA512))
	filemap := make(InReleaseFiles, size)

	targets := []filesTarget{
		{i.MD5Sum, func(file *InReleaseFile, hash string) { file.MD5Sum = hash }},
		{i.SHA1, func(file *InReleaseFile, hash string) { file.SHA1 = hash }},
		{i.SHA256, func(file *InReleaseFile, hash string) { file.SHA256 = hash }},
		{i.SHA512, func(file *InReleaseFile, hash string) { file.SHA512 = hash }},
	}

	for _, target := range targets {
		for _, sum := range target.sums {
			file, ok := filemap[sum.Path]
			if !ok {
				file = InReleaseFile{Size: sum.Size, Path: sum.Path}
			} else if file.Size != sum.Size {
				return nil, fmt.Errorf("inconsistent file size for %s: %d vs %d", sum.Path, file.Size, sum.Size)
			}

			target.setHash(&file, sum.Hash)
			filemap[sum.Path] = file
		}
	}

	return filemap, nil
}
