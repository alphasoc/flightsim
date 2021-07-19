package ssh

import (
	"fmt"
	"io"

	"github.com/alphasoc/flightsim/simulator/ssh/fxp"
	"github.com/alphasoc/flightsim/utils"
)

// Packet interface represents specific sftp packet type (data section to be exact)
// ex. sshFxpInitPacket, sshFxpVersionPacket
type Packet interface {
	Marshal() []byte
	Unmarshal([]byte) error
	GetCode() byte
}

// RawPacket represents a partialy parsed sftp packet (without the parsed data section).
type RawPacket struct {
	length   uint32
	typecode byte
	data     []byte
}

// MakeRawPacket wraps a specific packet with code and length data and returns a *RawPacket.
func MakeRawPacket(p Packet) *RawPacket {
	data := p.Marshal()
	return &RawPacket{
		length:   uint32(len(data) + 1),
		typecode: p.GetCode(),
		data:     data,
	}
}

// Marshal converts the length field of a RawPacket to network byte order.
func (p *RawPacket) Marshal() []byte {
	if p.length == 0 {
		return []byte{0, 0, 0, 0}
	}
	result := make([]byte, 0)
	result = append(result, utils.MarshalUint32(p.length)...)
	result = append(result, p.typecode)
	result = append(result, p.data...)
	return result
}

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
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return resp, err
		}
		return nil, fmt.Errorf("failed reading response: %w", err)
	}
	length := utils.UnmarshalUint32(resp)
	resp = make([]byte, length)
	if _, err := io.ReadFull(r, resp[:length]); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return resp, err
		}
		return nil, fmt.Errorf("failed reading response: %w", err)
	}
	return resp, nil
}

// VersionResp parses an SSH/SFTP response to an init request.  A Version packet and
// an error are returned.
func VersionResp(data []byte) (*fxp.Version, error) {
	parser := fxp.NewFieldParser(data)
	version := &fxp.Version{Version: parser.ReadUint32()}
	if err := parser.GetError(); err != nil {
		return nil, fmt.Errorf("failed parsing version response: %w", err)
	}
	return version, nil
}

// OpenResp parses an SSH/SFTP response to an open file request.  A Handle packet and
// an error are returned.
func OpenResp(data []byte) (*fxp.Handle, error) {
	parser := fxp.NewFieldParser(data)
	handle := &fxp.Handle{
		ID:     parser.ReadUint32(),
		Handle: parser.ReadString(),
	}
	err := parser.GetError()
	if err != nil {
		return nil, fmt.Errorf("failed parsing open response: %w", err)
	}
	return handle, nil
}

// StatusResp parses an SSH/SFTP status response.  A status response may be returned
// for any client request.  In particular, we care about:
//   1. status responses to write requests (success/error cases).
//   2. status responses carrying errors.
// A Status packet and an error are returned.
func StatusResp(data []byte) (*fxp.Status, error) {
	status := &fxp.Status{}
	err := status.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed parsing status response: %w", err)
	}
	return status, nil
}
