package ssh

import (
	"encoding/binary"
	"fmt"
	"io"

	sdifxp "github.com/alphasoc/sftp-data-input/packet/fxp"
)

// ReadPacket reads data from the io.Reader and returns the data as a byte slice along
// with an error.
func ReadPacket(r io.Reader) ([]byte, error) {
	// 4 bytes specifying the length of the entire response packet, excluding the length
	// field itself:
	// https://datatracker.ietf.org/doc/html/draft-ietf-secsh-filexfer-13#section-4
	const respLenSize = 4
	// Read the actual response by first reading the response length, then read the
	// remaining response (minus the length).
	resp := make([]byte, respLenSize)
	_, err := io.ReadFull(r, resp[:respLenSize])
	if err != nil {
		return nil, fmt.Errorf("failed reading response: %v", err)
	}
	length := binary.BigEndian.Uint32(resp)
	resp = make([]byte, length)
	if _, err := io.ReadFull(r, resp[:length]); err != nil {
		return nil, fmt.Errorf("failed reading response: %v", err)
	}
	return resp, nil
}

// VersionResp parses an SSH/SFTP response to an init request.  A Version packet and
// an error are returned.
func VersionResp(data []byte) (*sdifxp.Version, error) {
	parser := sdifxp.NewFieldParser(data)
	version := &sdifxp.Version{Version: parser.ReadUint32()}
	if err := parser.GetError(); err != nil {
		return nil, fmt.Errorf("failed parsing version response: %v", err)
	}
	return version, nil
}

// OpenResp parses an SSH/SFTP response to an open file request.  A Handle packet and
// an error are returned.
func OpenResp(data []byte) (*sdifxp.Handle, error) {
	parser := sdifxp.NewFieldParser(data)
	handle := &sdifxp.Handle{
		ID:     parser.ReadUint32(),
		Handle: parser.ReadString(),
	}
	err := parser.GetError()
	if err != nil {
		return nil, fmt.Errorf("failed parsing open response: %v", err)
	}
	return handle, nil
}

// StatusResp parses an SSH/SFTP status response.  A status response may be returned
// for any client request.  In particular, we care about:
//   1. status responses to write requests (success/error cases).
//   2. status responses carrying errors.
// A Status packet and an error are returned.
func StatusResp(data []byte) (*sdifxp.Status, error) {
	status := &sdifxp.Status{}
	err := status.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed parsing status response: %v", err)
	}
	return status, nil
}
