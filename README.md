# Inter-Planetary Advanced Package Mirror

[![Go CI](https://github.com/NatoBoram/ipapm/actions/workflows/go.yaml/badge.svg)](https://github.com/NatoBoram/ipapm/actions/workflows/go.yaml) [![Docker CI](https://github.com/NatoBoram/ipapm/actions/workflows/docker.yaml/badge.svg)](https://github.com/NatoBoram/ipapm/actions/workflows/docker.yaml) [![CodeQL](https://github.com/NatoBoram/ipapm/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/NatoBoram/ipapm/actions/workflows/github-code-scanning/codeql)

Mirrors APT repositories to IPFS.

## Roadmap

> [!CAUTION]
> This is a work in progress.

- [x] Get Kubo config
- [x] Connect to Kubo (`github.com/ipfs/kubo/client/rpc`)
- [x] Get APT sources and config
- [ ] Merge sources per URIs
- [ ] Download InRelease files (`github.com/aptly-dev/aptly`)
- [ ] Verify GPG keys (`github.com/ProtonMail/gopenpgp/v3`)
- [ ] Get `InRelease` from MFS
- [ ] Search differences
- [ ] Stream different files to MFS while verifying hashes
- [ ] Delete outdated files from MFS
- [ ] Commit `InRelease` file to MFS
- [ ] Publish to IPNS

## Docker

The default config file is at `/home/nonroot/.config/ipapm/config.yaml`.

```yaml
Sources: []
Port: 9090
```

Environment variables are loaded in the following order:

- `.env.${GO_ENV}.local`
- `.env.${GO_ENV}`
- `.env.local`
- `.env`

```env
CONFIG_DIR=
KUBO_API_AUTH=
KUBO_API_URL=
```

## License

Copyright © 2026 Nato Boram

This program is free software: you can redistribute it and/or modify it under the terms of the **GNU Affero General Public License** as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but _without any warranty_; without even the implied warranty of _merchantability_ or _fitness for a particular purpose_. See the GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License along with this program. If not, see <https://www.gnu.org/licenses>.
