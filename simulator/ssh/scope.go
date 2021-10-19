package ssh

import (
	"fmt"
	"strings"

	bytesize "github.com/inhies/go-bytesize"
)

// ParseScope parses the common commandline portion after the module name (if supplied)
// for SSH-based simulators.  Defaults for target and transfer size are passed as arguments.
//   ie. flightsim run ssh-transfer:this-part-is-scope:and-can-contain-more
// For the moment, only send size is parsed, but ultimately we also want to parse
// destination host and port.  Returns a string representation of the destination
// host, the send size as a ByteSize, and an error.
func ParseScope(
	scope string,
	defaultTargets []string,
	defaultSendSize bytesize.ByteSize) ([]string, bytesize.ByteSize, error) {
	// scope can be "", in which case, apply defaults.
	if scope == "" {
		return defaultTargets, defaultSendSize, nil
	}
	// scope may contain just the send size (ie. a lack of futher ":" separators
	// present in the string).
	var sendSize bytesize.ByteSize
	var err error
	if !strings.Contains(scope, ":") {
		sendSize, err = bytesize.Parse(scope)
		if err != nil {
			return []string{""},
				bytesize.ByteSize(0),
				fmt.Errorf("invalid command line: '%v': %w", scope, err)
		}
		return defaultTargets, sendSize, nil
	}
	// TODO scope may contain more information, separated by ":", perhaps as key-value
	// pairs.  For now, not supported.
	return []string{""}, bytesize.ByteSize(0), fmt.Errorf("invalid command line: '%v'", scope)

}
