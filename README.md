# Inter-Planetary Advanced Package Mirror

[![Go CI](https://github.com/NatoBoram/ipapm/actions/workflows/go.yaml/badge.svg)](https://github.com/NatoBoram/ipapm/actions/workflows/go.yaml) [![Docker CI](https://github.com/NatoBoram/ipapm/actions/workflows/docker.yaml/badge.svg)](https://github.com/NatoBoram/ipapm/actions/workflows/docker.yaml) [![CodeQL](https://github.com/NatoBoram/ipapm/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/NatoBoram/ipapm/actions/workflows/github-code-scanning/codeql)

Mirrors APT repositories to IPFS.

![Screenshot](https://github.com/user-attachments/assets/dd4d94bc-a2e4-4996-acc1-d1c4df6ec854)

## Usage

This program downloads packages from APT repositories, mirrors them to IPFS using Kubo's Mutable FileSystem and then publishes them to IPNS. It will check for changes between previous and next versions and only download added and updated files.

To crawl APT repositories, it uses `InRelease` files and verifies their PGP signature. It also verifies the hash of any file downloaded from that `InRelease` file and from `Packages` & `Sources` files. An unsigned or incorrect repository cannot be mirrored.

## Installation

Binaries are available in [Releases](https://github.com/NatoBoram/ipapm/releases) or it can be installed from source.

```sh
go install github.com/NatoBoram/ipapm@latest
```

Environment variables set private values while the config file sets public values.

### Environment variables

These are the default values:

```env
CONFIG_DIR=~/.config/ipapm
GO_ENV=development
KUBO_API_AUTH=
KUBO_API_URL=http://localhost:5001
```

`GO_ENV` changes the logging style into JSON Lines when set to `production`.

`KUBO_API_AUTH` corresponds to `API.Authorizations.api.AuthSecret` in Kubo's config file. See [API.Authorizations: AuthSecret](https://github.com/ipfs/kubo/blob/master/docs/config.md#apiauthorizations-authsecret).

#### Examples

```env
KUBO_API_AUTH='bearer:RGypK3ftgjie4aglyFh034j1e1dDnlJLpp2PiQAxuZ2JuITKKfGM7F6/2428WFtp+8AqMQArl3Sirpus26gauN1G'
KUBO_API_AUTH='basic:user:RGypK3ftgjie4aglyFh034j1e1dDnlJLpp2PiQAxuZ2JuITKKfGM7F6/2428WFtp+8AqMQArl3Sirpus26gauN1G'
KUBO_API_AUTH='basic:dXNlcjpSR3lwSzNmdGdqaWU0YWdseUZoMDM0ajFlMWREbmxKTHBwMlBpUUF4dVoySnVJVEtLZkdNN0Y2LzI0MjhXRnRwKzhBcU1RQXJsM1NpcnB1czI2Z2F1TjFH'
```

### Config

The config is, by default, at `~/.config/ipapm/config.yaml`. Its parent folder can be set with `CONFIG_DIR`.

Here's the default values:

```yaml
Kubo:
  MFS: /ipapm
Port: 9090
Sources: []
```

Here's a full example:

```yaml
Kubo:
  MFS: /ipapm
Sources:
  - URIs:
      - https://packages.termux.dev/apt/termux-main
    Suites:
      - stable
      - staging
    Signed-By: /usr/share/keyrings/termux-autobuilds.gpg
  - URIs:
      - https://packages.microsoft.com/repos/code
    Suites:
      - stable
    Signed-By: /usr/share/keyrings/microsoft.gpg
Port: 9090
```

### Docker

The default config file is at `/home/nonroot/.config/ipapm/config.yaml`. Don't forget to mount `.gpg` signatures.

```yaml
services:
  ipapm:
    container_name: ipapm
    env_file:
      - path: .env
      - path: .env.local
    environment:
      GO_ENV: production
    healthcheck:
      test:
        - CMD
        - /usr/local/bin/readyz
    image: natoboram/ipapm
    volumes:
      - ~/.config/ipapm/:/home/nonroot/.config/ipapm/
      - /usr/share/keyrings/:/usr/share/keyrings/:ro
```

## API

Two endpoints are exposed on port `9090`:

- `/livez`
- `/readyz`

They respond with `204 No Content` when everything is right. If Kubo cannot be contacted, then `/readyz` responds with `503 Service Unavailable`.

## Development

Environment variables are loaded in the following order:

- `.env.${GO_ENV}.local`
- `.env.${GO_ENV}`
- `.env.local`
- `.env`

`GO_ENV` cannot be set from an `.env` file and must be set in the environment. It defaults to `development`.

## License

Copyright © 2026 Nato Boram

This program is free software: you can redistribute it and/or modify it under the terms of the **GNU Affero General Public License** as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but _without any warranty_; without even the implied warranty of _merchantability_ or _fitness for a particular purpose_. See the GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License along with this program. If not, see <https://www.gnu.org/licenses>.
