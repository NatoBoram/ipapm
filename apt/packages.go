package apt

import (
	"net/url"
)

type Packages struct {
	Package       string
	Version       string
	Architecture  string
	Section       string
	Priority      string
	InstalledSize uint
	Maintainer    string
	Description   string
	Homepage      *url.URL
	Conflicts     string
	Depends       string
	Recommends    string
	Provides      string
	Replaces      string
	MD5sum        string
	SHA1          string
	SHA256        string
	SHA512        string
	Size          uint
	Filename      string

	Raw string
}

type ComponentBytes struct {
	Component Component
	Packages  []Packages
}

// func (c *Client) Packages(
// 	ctx context.Context, uri *url.URL, suite string, component Component,
// ) ([]Packages, error) {
// }
