# <img src="logo/bine-logo.png" width="180px">

[![GoDoc](https://godoc.org/github.com/cretz/bine?status.svg)](https://godoc.org/github.com/cretz/bine)

Bine is a Go API for using and controlling Tor. It is similar to [Stem](https://stem.torproject.org/).

Features:

* Full support for the Tor controller API
* Support for `net.Conn` and `net.Listen` style APIs
* Supports statically compiled Tor to embed Tor into the binary
* Supports both v2 and v3 onion services
* Support for embedded control socket in Tor >= 0.3.5 (non-Windows)

See info below, the [API docs](http://godoc.org/github.com/cretz/bine), and the [examples](examples). The project is
MIT licensed. The Tor docs/specs and https://github.com/yawning/bulb were great helps when building this.

## Example

It is really easy to create an onion service. For example, assuming `tor` is on the `PATH`, this bit of code will show
a directory server of the current directory:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cretz/bine/tor"
)

func main() {
	// Start tor with default config (can set start conf's DebugWriter to os.Stdout for debug logs)
	fmt.Println("Starting and registering onion service, please wait a couple of minutes...")
	t, err := tor.Start(nil, nil)
	if err != nil {
		log.Panicf("Unable to start Tor: %v", err)
	}
	defer t.Close()
	// Wait at most a few minutes to publish the service
	listenCtx, listenCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer listenCancel()
	// Create a v3 onion service to listen on any port but show as 80
	onion, err := t.Listen(listenCtx, &tor.ListenConf{Version3: true, RemotePorts: []int{80}})
	if err != nil {
		log.Panicf("Unable to create onion service: %v", err)
	}
	defer onion.Close()
	fmt.Printf("Open Tor browser and navigate to http://%v.onion\n", onion.ID)
	fmt.Println("Press enter to exit")
	// Serve the current folder from HTTP
	errCh := make(chan error, 1)
	go func() { errCh <- http.Serve(onion, http.FileServer(http.Dir("."))) }()
	// End when enter is pressed
	go func() {
		fmt.Scanln()
		errCh <- nil
	}()
	if err = <-errCh; err != nil {
		log.Panicf("Failed serving: %v", err)
	}
}
```

If in `main.go` it can simply be run with `go run main.go`. Of course this uses a separate `tor` process. To embed Tor
statically in the binary, follow the [embedded package docs](https://godoc.org/github.com/cretz/bine/process/embedded)
which will require [building Tor statically](https://github.com/cretz/tor-static). Then with
`github.com/cretz/bine/process/embedded` imported, change the start line above to:

```go
t, err := tor.Start(nil, &tor.StartConf{ProcessCreator: embedded.NewCreator()})
```

This defaults to Tor 0.3.5.x versions but others can be used from different packages. In non-Windows environments, the
`UseEmbeddedControlConn` field in `StartConf` can be set to `true` to use an embedded socket that does not open a
control port.

Tested on Windows, the original exe file is ~7MB. With Tor statically linked it comes to ~24MB, but Tor does not have to
be distributed separately. Of course take notice of all licenses in accompanying projects.

## Testing

To test, a simple `go test ./...` from the base of the repository will work (add in a `-v` in there to see the tests).
The integration tests in `tests` however will be skipped. To execute those tests, `-tor` must be passed to the test.
Also, `tor` must be on the `PATH` or `-tor.path` must be set to the path of the `tor` executable. Even with those flags,
only the integration tests that do not connect to the Tor network are run. To also include the tests that use the Tor
network, add the `-tor.network` flag. For details Tor logs during any of the integration tests, use the `-tor.verbose`
flag.
