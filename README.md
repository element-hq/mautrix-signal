## Element fork

The Element fork includes the following changes:
- [User activity tracking & add additional metrics to the bridge](https://github.com/element-hq/mautrix-signal/pull/21)
- [Use Element fork of mautrix-go](https://github.com/element-hq/mautrix-signal/pull/22)
- [Add config to disallow manual double puppeting](https://github.com/mautrix/signal/pull/437)
- [Don't no-op logouts](https://github.com/mautrix/signal/pull/439)
- [Use latest profile info](https://github.com/mautrix/signal/pull/449)

Some changes that appear here may get upstreamed to https://github.com/mautrix/signal, and will be removed from
the list when they appear in both versions.

Tagged versions will appear as `v{UPSTREAM-VERSION}-mod-{VERSION}`

E.g. The third modification release to 1.0 of the upstream bridge would be `v1.0-mod-3`.

# mautrix-signal
![Languages](https://img.shields.io/github/languages/top/mautrix/signal.svg)
[![License](https://img.shields.io/github/license/mautrix/signal.svg)](LICENSE)
[![GitLab CI](https://mau.dev/mautrix/signal/badges/main/pipeline.svg)](https://mau.dev/mautrix/signal/container_registry)

A Matrix-Signal puppeting bridge.

## Documentation
All setup and usage instructions are located on
[docs.mau.fi](https://docs.mau.fi/bridges/go/signal/index.html).
Some quick links:

* [Bridge setup](https://docs.mau.fi/bridges/go/setup.html?bridge=signal)
  (or [with Docker](https://docs.mau.fi/bridges/general/docker-setup.html?bridge=signal))
* Basic usage: [Authentication](https://docs.mau.fi/bridges/go/signal/authentication.html)

### Features & Roadmap
[ROADMAP.md](https://github.com/mautrix/signal/blob/main/ROADMAP.md)
contains a general overview of what is supported by the bridge.

## Discussion
Matrix room: [`#signal:maunium.net`](https://matrix.to/#/#signal:maunium.net)
