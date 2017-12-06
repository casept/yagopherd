package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/mitchellh/go-homedir"
)

// Variables to identify the build
var (
	Version string
	Commit  string
)

// Config contains configuration obtained from CLI args and the config file.
var Config struct {
	RawGopherroot string // (Relative) path supplied by the user
	Gopherroot    string // Expanded path
	Port          int    // Port to listen on
	Address       string // Address to listen on
	RootSelector  string // The selector the client has to send to obtain a root listing
	// TODO: maxRequestLength int // How many bytes a client can at most send (DOS protection)
	IsTesting bool // Whether the server was started to run unit/integration tests.
}

// This struct is just used so gopher connection-specific methods can be cleanly declared.
type gopherConn struct {
	net.Conn
}

func init() {
	// Set up config flags and insert them into a global config struct
	// The test setup function sets these values itself, don't interfere with that
	if Config.IsTesting == false {
		// Set the default gopherroot to ~/.gopher
		homedir, err := homedir.Dir()
		if err != nil {
			log.Fatalf("Unable to determine user's home directory: %v\n", err)
		}
		defaultGopherroot := homedir + "/.gopher"

		flag.StringVar(&Config.RawGopherroot, "gopherroot", defaultGopherroot, "path to directory which contains the content that should be served")
		flag.IntVar(&Config.Port, "port", 7077, "port that yagopherd should listen on, the standard port 70 requires root/admin privileges")
		// TODO: flag.IntVar(&Config.maxRequestLength, "max-request-length", 4096, "how many bytes a client can at most send as part of a single request (prevents DOS)")
		flag.StringVar(&Config.Address, "address", "0.0.0.0", "Address that the server should listen on")
		// Parse CLI args
		flag.Parse()
		// TODO
		Config.RootSelector = "\r\n"
	}

	// Make path to gopherroot absolute
	var err error
	Config.Gopherroot, err = filepath.Abs(Config.RawGopherroot)
	if err != nil {
		log.Fatalf("Failed to expand relative path: %v", err)
	}

	// Make sure gopherroot directory exists and is readable
	_, err = os.Stat(Config.Gopherroot)
	if err != nil {
		log.Fatalf("Cannot stat directory %v: %v", Config.Gopherroot, err)
	}
}

// A gopherItem is one item that shows up on the client's menu for selection
type gopherItem struct {
	gophertype    string // Gophertype of the item
	displayString string // Display string that the client will render
	selector      string // Selector that the client sends to retrieve the item
	host          string // Hostname of the server the item resides on
	port          int    // Port of the server the item resides on
	fsLocation    string // Absolute path to the location on the filesystem (blank if item is remote)
	mimetype      string // Mimetype of the item
}

// A gophermap is a flat slice of gopherItems
type gophermap struct {
	items []gopherItem
}

// GopherReq holds all information related to a client request (but not the response, that's gopherItem's job)
type gopherReq struct {
	selector string // Selector the client sent
	path     string // Path to the requested item
}

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", Config.Address, strconv.Itoa(Config.Port)))
	if err != nil {
		log.Panicf("Error while setting up listener: %v\n", err.Error())
	}
	// Could be empty if built by go build instead of make
	if Version != "" {
		log.Printf("Yagopherd version %v starting...\n", Version)
	} else {
		log.Printf("Yagopherd starting (version unknown)...\n")
	}
	if Commit != "" {
		log.Printf("Commit: %v\n", Commit)
	} else {
		log.Printf("Commit: unknown\n")
	}
	log.Printf("Listening on %v:%v\n", Config.Address, strconv.Itoa(Config.Port))

	// Set up some signal handling stuff to enable a clean shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// This WG is used to allow all handler goroutines to cleanly finish.
	wg := new(sync.WaitGroup)
	// Launch a background "signal monitor" goroutine
	// This isn't done in the main loop because that loop blocks on listener.Accept()
	go func() {
		sig := <-sigs
		log.Printf("Received %v signal, waiting for request processing to finish...\n", sig.String())
		wg.Wait()
		// The tests return their own exit code, don't mess with that
		if Config.IsTesting == false {
			log.Println("Done, shutting down...")
			os.Exit(0)
		}
	}()

	// Main listener loop
	for {
		// TODO: Refuse any new incoming connections upon shutdown init
		rawConn, err := listener.Accept()
		conn := gopherConn{
			rawConn,
		}
		defer conn.Close()

		if err != nil {
			log.Printf("An error occurred while trying to accept request: %v\n", err.Error())
		} else {
			log.Printf("Received request from: %v", net.Addr.String(conn.RemoteAddr()))
			wg.Add(1)
			go handleReq(conn, wg)
		}
	}
}

// handleReq handles an incoming request by parsing the selector and sending the selected content to the client.
func handleReq(conn gopherConn, wg *sync.WaitGroup) {
	defer wg.Done()
	var req gopherReq
	// Extract the selector from response
	rd := bufio.NewReader(conn)
	var err error
	// TODO: Limit number of bytes read to avoid DOS, make that configurable
	rawStr, err := rd.ReadString('\n')
	// Remove the \r\n sent at the end of a request (this one was fun to debug... not.)
	// Kinda messy, fuck it
	req.selector = strings.TrimRight(rawStr, "\n\r")

	if err != nil {
		// Abort the goroutine if selector extraction fails
		log.Printf("Error while trying to process selector: %v", err)
		return
	}

	var gophermap gophermap

	// The CRLF selector indicates we should send a root listing
	if req.selector == Config.RootSelector {
		req.path = Config.Gopherroot
	} else {
		req.path, err = appendDir(Config.Gopherroot, req.selector)
		if err != nil {
			conn.sendErr(err)
			return
		}
	}
	fInfo, err := os.Stat(req.path)
	if err != nil {
		conn.sendErr(err)
		return
	}

	if fInfo.IsDir() == true {
		gophermap, err = dirToGophermap(req.path)
		if err != nil {
			conn.sendErr(err)
			return
		}
		err = gophermap.send(conn)
		if err != nil {
			conn.sendErr(err)
		}
	} else {
		err = conn.sendFile(req.path)
		if err != nil {
			conn.sendErr(err)
			return
		}
	}

	conn.Close()
	return
}

// serialize serializes a gophermap.
func (gophermap *gophermap) serialize() (serializedGophermap []byte, err error) {
	// Check whether all required fields are filled out
	for i := 0; i < len(gophermap.items); i++ {
		if len(gophermap.items[i].gophertype) == 0 || len(gophermap.items[i].displayString) == 0 || len(gophermap.items[i].selector) == 0 || len(gophermap.items[i].host) == 0 || gophermap.items[i].port == 0 {
			return []byte{}, errors.New("gophermap is invalid")
		}
	}

	var serializedString string
	for i := 0; i < len(gophermap.items); i++ {
		// TODO:
		fmt.Printf(gophermap.items[i].gophertype)
		serializedString = serializedString + gophermap.items[i].gophertype + gophermap.items[i].displayString + "\t" + gophermap.items[i].selector + "\t" + gophermap.items[i].host + "\t" + strconv.Itoa(gophermap.items[i].port) + "\r\n"
	}
	// Response must end with a "."
	serializedString = serializedString + "."
	return []byte(serializedString), nil
}

// send sends the gohpermap over conn.
func (gophermap gophermap) send(conn gopherConn) (err error) {
	sGophermap, err := gophermap.serialize()
	if err != nil {
		return err
	}
	_, err = conn.Write(sGophermap)
	if err != nil {
		return err
	}
	return nil
}

// sendErr sends an error to the client and closes the connection.
func (conn gopherConn) sendErr(errorMsg error) (err error) {
	log.Printf("Sending error to client: %v\n", errorMsg)
	resp := fmt.Sprintf("3" + errorMsg.Error() + "\r\n" + ".")
	_, err = conn.Write([]byte(resp))
	if err != nil {
		return fmt.Errorf("error while sending error message to %v: %v", conn.RemoteAddr(), err.Error())
	}
	return nil
}
