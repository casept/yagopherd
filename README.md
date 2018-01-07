# yagopherd, yet another gopher server written in golang

This is a simple implementation of a server for the `gopher` protocol as defined in [RFC 1436](https://tools.ietf.org/html/rfc1436) (with some parts omitted, see `unplanned features`).
It isn't very well optimized right now and doesn't support some of the protocol's features, it's mostly a way for me to learn golang and writing servers.

[![Windows build status](https://ci.appveyor.com/api/projects/status/ik3q9xkr6cc1eufw/branch/master?svg=true)](https://ci.appveyor.com/project/casept/yagopherd/branch/master)
[![Linux/OSX build status](https://travis-ci.org/casept/yagopherd.svg?branch=master)](https://travis-ci.org/casept/yagopherd)
[![Go Report Card](https://goreportcard.com/badge/github.com/casept/yagopherd)](https://goreportcard.com/report/github.com/casept/yagopherd)
[![Coverage Status](https://coveralls.io/repos/github/casept/yagopherd/badge.svg?branch=master)](https://coveralls.io/github/casept/yagopherd?branch=master)

The project's far from finished:

## TODO

* ~~Basic functionality (retrieve files/directory listings)~~
* ~~Forcing immediate shutdown on second SIGTERM~~
* ~~A config file. All settings are currently passed as CLI arguments.~~
* Different log levels
* Log file support
* Gopher + support
	* ~~Handle legacy clients~~
	* ~~Signal g+ support~~
	* ~~Send size of files w/ response when possible~~
	* ~~Support gopher+ errors~~
	* Support gopher+ attributes
		* ADMIN
		* INFO
		* VIEWS
		* user-defined attributes (requires support for attribute files)
* `.gophermap` support
* Support for linking to other servers (Probably via `.gophermap` files)
* Better differentiation between binary and text files (tricky!)
* Search support
* Support for telnet/SSH sessions
* ~~CI~~
* Tests
* Stress tests
* Benchmarks
* HTTP gateway (maybe, should be simple, perhaps embed caddy?)
* `chroot` support (maybe, requires server to start as root)
* Redundant server support (maybe)
* Deciding gophertype based on mimetype instead of file extension (requires a good, cgo-free `libmagick`-like library)
* Ability to use manually created `.gophermap` files instead of indexing automatically
* General code cleanup
* Caching
* DOS protection:
	* Against "slowloris-like" attacks
	* Rate limiting
	* Bandwith limit per client

## Unplanned features

Several features of the gopher protocol are no longer used in the wild. These won't be implemented to avoid needless complexity.

* CSO server support (gophertype `2`)
* Special treatment for binHexed files (gophertype `4`)
* tn3270 session support (gophertype `T`)

## Installation

The project is `go get`-compatible:
```
go get -v -u github.com/casept/yagopherd/
```
Alternatively, if you wish to build a `release` version with the version and commit hash embedded you'll need to build using `make`:
```
go get -v -u -d github.com/casept/yagopherd/
cd $GOPATH/src/github.com/casept/yagopherd/
# To simply build
make
# To install the binary into $GOPATH/bin/
make install
# To run tests and benchmarks
make test
```
If you don't have `make` available for some reason you can inspect the Makefile (it's fairly straightforward) and run the commands manually.

### Usage

Assuming `$GOPATH/bin` is in your `$PATH` you can simply run yagopherd:
```
yagopherd

```
To view CLI options:
```
yagopherd --help
```
These values can also be set by config file.
The following locations are searched for config files (Each item takes precedence over the item below it):
* On \*nix :
	* The present working directory
	* `$XDG_CONFIG_HOME/yagopherd/yagohperd/` (`$HOME/.config/yagopherd/yagopherd/` by default)
	* `$XDG_CONFIG_DIRS/yagopherd/yagopherd` (`/etc/xdg/yagopherd/yagopherd/` by default)

* On windows:
	* The present working directory
	* `%APPDATA%/yagopherd/yagopherd`
	* `%PROGRAMDATA%/yagopherd/yagopherd`

* On MacOS/OSX:
	* The present working directory
	* `${HOME}/Library/Application Support/yagopherd/yagopherd`
	* `/Library/Application Support/yagopherd/yagopherd`

The config file must be within one of these directories, named `yagopherd.replacewithformat` and in `TOML`, `YAML`, `JSON`, `HCL` or `Java properties` format.
`TOML` is the format used in the documentation.

Environment variables can be used as well. Simply prefix the CLI flag's name with `YAGOPHERD_`.
For example, `export YAGOPHERD_PORT=1337` would set the port the server listens on to `1337`.

## License

The Project is licensed under the [GNU GPLv3](https://www.gnu.org/licenses/gpl-3.0.html) license. See the LICENSE file for details.
