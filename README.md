# yagopherd, yet another gopher server written in golang

This is a simple implementation of a server for the `gopher` protocol as defined in [RFC 1436](https://tools.ietf.org/html/rfc1436) (with some parts omitted, see `unplanned features`).
It isn't very well optimized right now and doesn't support some of the protocol's features, it's mostly a way for me to learn golang and writing servers. It's far from finished:

## TODO

* ~~Basic functionality (retrieve files/directory listings)~~
* Gopher + support
* `.gophermap` support
* Support for linking to other servers (Probably via `.gophermap` files)
* Better differentiation between binary and text files (tricky!)
* Search support
* Support for telnet/SSH sessions
* A config file. All settings are currently passed as CLI arguments.
* CI
* Tests
* HTTP gateway (maybe, should be simple, perhaps embed caddy?)
* `chroot` support (maybe, requires server to start as root)
* Redundant server support (maybe)
* Deciding gophertype based on mimetype instead of file extension (requires a good, cgo-free `libmagick`-like library)
* Ability to use manually created `.gophermap` files instead of indexing automatically
* General code cleanup

## Unplanned features

Several features of the gopher protocol are no longer used in the wild. These won't be implemented to avoid needless complexity.

* CSO server support (gophertype `2`)
* Special treatment for binHexed files (gophertype `4`)
* tn3270 session support (gophertype `T`)

## License
The Project is licensed under the [GNU GPLv3](https://www.gnu.org/licenses/gpl-3.0.html) license. See the LICENSE file for details.
