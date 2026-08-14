package apt

import (
	"fmt"
	"path"
)

// FileHashes contains all the file entries of an [InRelease] file.
type FileHashes map[string]FileHash

// FileHash is a single file entry in an [InRelease] file. It's also used as a
// lowest common denominator for [Packages] file entries.
type FileHash struct {
	MD5Sum string
	SHA1   string
	SHA256 string
	SHA512 string

	Size uint

	// Filename may begin with `${component}/` from [InRelease], `pool/${suite}/`
	// from [Packages] or `pool/${suite}/` from [Source].
	Filename string
}

// FileHashes obtains the file entries from an [InRelease] file as a
// [FileHashes].
func (i InRelease) FileHashes() (FileHashes, error) {
	size := max(len(i.MD5Sum), len(i.SHA1), len(i.SHA256), len(i.SHA512))
	filemap := make(FileHashes, size)

	targets := []irTarget{
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

type irTarget struct {
	sums    []InReleaseSum
	setHash func(file *FileHash, hash string)
}

// FileHash turns a [Package] info a [FileHash].
func (p Package) FileHash() FileHash {
	return FileHash{
		MD5Sum:   p.MD5sum,
		SHA1:     p.SHA1,
		SHA256:   p.SHA256,
		SHA512:   p.SHA512,
		Size:     p.Size,
		Filename: p.Filename,
	}
}

// FileHash turns a [Source] info a [FileHashes].
func (s Source) FileHashes() (FileHashes, error) {
	size := max(
		len(s.Files),
		len(s.ChecksumsSha1), len(s.ChecksumsSha256), len(s.ChecksumsSha512),
	)
	filemap := make(FileHashes, size)

	targets := []sTarget{
		{s.Files, func(file *FileHash, hash string) { file.MD5Sum = hash }},
		{s.ChecksumsSha1, func(file *FileHash, hash string) { file.SHA1 = hash }},
		{s.ChecksumsSha256, func(file *FileHash, hash string) { file.SHA256 = hash }},
		{s.ChecksumsSha512, func(file *FileHash, hash string) { file.SHA512 = hash }},
	}

	for _, target := range targets {
		for _, sum := range target.sums {
			file, ok := filemap[sum.Name]
			if !ok {
				file = FileHash{Size: sum.Size, Filename: path.Join(s.Directory, sum.Name)}
			} else if file.Size != sum.Size {
				return nil, fmt.Errorf("inconsistent file size for %s: %d vs %d", sum.Name, file.Size, sum.Size)
			}

			target.setHash(&file, sum.Hash)
			filemap[sum.Name] = file
		}
	}

	return filemap, nil
}

type sTarget struct {
	sums    []SourceSum
	setHash func(file *FileHash, hash string)
}
