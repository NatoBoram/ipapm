package apt

import "fmt"

type FileHashes map[string]FileHash

type FileHash struct {
	MD5Sum string
	SHA1   string
	SHA256 string
	SHA512 string

	Size uint

	// Filename may begin with `${component}/` from [InRelease] or
	// `pool/${suite}/` form [Packages].
	Filename string
}

func (i InRelease) Files() (FileHashes, error) {
	size := max(len(i.MD5Sum), len(i.SHA1), len(i.SHA256), len(i.SHA512))
	filemap := make(FileHashes, size)

	targets := []filesTarget{
		{i.MD5Sum, func(file *FileHash, hash string) { file.MD5Sum = hash }},
		{i.SHA1, func(file *FileHash, hash string) { file.SHA1 = hash }},
		{i.SHA256, func(file *FileHash, hash string) { file.SHA256 = hash }},
		{i.SHA512, func(file *FileHash, hash string) { file.SHA512 = hash }},
	}

	for _, target := range targets {
		for _, sum := range target.sums {
			file, ok := filemap[sum.Path]
			if !ok {
				file = FileHash{Size: sum.Size, Filename: sum.Path}
			} else if file.Size != sum.Size {
				return nil, fmt.Errorf("inconsistent file size for %s: %d vs %d", sum.Path, file.Size, sum.Size)
			}

			target.setHash(&file, sum.Hash)
			filemap[sum.Path] = file
		}
	}

	return filemap, nil
}

type filesTarget struct {
	sums    []InReleaseSum
	setHash func(file *FileHash, hash string)
}
