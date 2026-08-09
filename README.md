# Inter-Planetary Advanced Package Mirror

[![Go CI](https://github.com/NatoBoram/ipapm/actions/workflows/go.yaml/badge.svg)](https://github.com/NatoBoram/ipapm/actions/workflows/go.yaml) [![Docker CI](https://github.com/NatoBoram/ipapm/actions/workflows/docker.yaml/badge.svg)](https://github.com/NatoBoram/ipapm/actions/workflows/docker.yaml) [![CodeQL](https://github.com/NatoBoram/ipapm/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/NatoBoram/ipapm/actions/workflows/github-code-scanning/codeql)

Mirrors APT repositories to IPFS.

## Roadmap

> [!CAUTION]
> This is a work in progress.

- [x] Get Kubo config
  - [x] Connect to Kubo (`github.com/ipfs/kubo/client/rpc`)
- [x] Get APT sources and config
  - [x] Merge sources per URIs
  - [x] Download InRelease files
  - [x] Verify GPG keys (`github.com/ProtonMail/go-crypto/openpgp`)
  - [x] Get `InRelease` from MFS (<https://docs.ipfs.tech/reference/kubo/rpc/#api-v0-files-read>)
    - [x] Search differences
    - [x] Get `Packages` files
      - [x] Sync all if no previous
      - [ ] Search differences
      - [x] Stream added/changed files to MFS while verifying hashes
      - [ ] Delete outdated files from MFS
      - [x] Commit `Packages` files to MFS
    - [ ] Get `Sources` files
      - [ ] Sync all if no previous
      - [ ] Search differences
      - [ ] Stream added/changed files to MFS while verifying hashes
      - [ ] Delete outdated files from MFS
      - [ ] Commit `Sources` files to MFS
    - [x] Commit `InRelease` file to MFS
  - [ ] Publish to IPNS
- [ ] Parallelize downloads by URI
- [ ] Progress bar for downloading files (`github.com/cheggaaa/pb/v3`)

## Docker

The default config file is at `/home/nonroot/.config/ipapm/config.yaml`.

```yaml
Kubo:
	MFS: /ipapm
Port: 9090
Sources: []
```

Environment variables are loaded in the following order:

- `.env.${GO_ENV}.local`
- `.env.${GO_ENV}`
- `.env.local`
- `.env`

These are the default values:

```env
CONFIG_DIR=~/.config/ipapm
KUBO_API_AUTH=
KUBO_API_URL=http://localhost:5001
```

## License

Copyright © 2026 Nato Boram

This program is free software: you can redistribute it and/or modify it under the terms of the **GNU Affero General Public License** as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but _without any warranty_; without even the implied warranty of _merchantability_ or _fitness for a particular purpose_. See the GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License along with this program. If not, see <https://www.gnu.org/licenses>.
