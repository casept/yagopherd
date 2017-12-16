package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/viper"
)

// Variables to identify the build
var (
	Version string
	Commit  string
)

// This struct is just used so gopher-connection specific methods can be cleanly declared.
type gopherConn struct {
	net.Conn
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
	// Load viper config
	setupConfig()
	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", viper.Get("address"), viper.Get("port")))
	if err != nil {
		log.Panicf("Error while setting up listener on address %v: %v\n", viper.GetString("addr"), err.Error())
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
	if viper.ConfigFileUsed() != "" {
		log.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}
	log.Printf("Listening on %v:%v\n", viper.GetString("address"), strconv.Itoa(viper.GetInt("port")))

	// Set up some signal handling stuff to enable a clean shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// This WG is used to allow all handler goroutines to cleanly finish.
	wg := new(sync.WaitGroup)
	// Launch a background "signal monitor" goroutine
	// This isn't done in the main loop because that loop blocks on listener.Accept()
	go func() {
		sig := <-sigs
		log.Printf("Received %v signal, waiting for request processing to finish and other cleanup...\n", sig.String())
		log.Print("Hit CTRL-C again if you wish to terminate without cleanup.")
		// Launch another one if a second SIGTERM is received.
		// This one terminates immediately, without doing any cleanup, useful if the program gets stuck.
		go func() {
			sig := <-sigs
			log.Printf("Received second %v signal, forcing shutdown without cleanup!", sig.String())
			os.Exit(0)
		}()
		wg.Wait()
		// The tests return their own exit code, don't mess with that
		if viper.GetBool("testmode") == false {
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
	if req.selector == "\r\n" {
		req.path = viper.GetString("gopherroot")
	} else {
		req.path, err = appendDir(viper.GetString("gopherroot"), req.selector)
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
