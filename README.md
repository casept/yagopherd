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
* Forcing immediate shutdown on second SIGTERM
* A config file. All settings are currently passed as CLI arguments.
* Different log levels
* Log file support
* Gopher + support
* `.gophermap` support
* Support for linking to other servers (Probably via `.gophermap` files)
* Better differentiation between binary and text files (tricky!)
* Search support
* Support for telnet/SSH sessions
* ~~CI~~
* Tests
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

## License
The Project is licensed under the [GNU GPLv3](https://www.gnu.org/licenses/gpl-3.0.html) license. See the LICENSE file for details.
