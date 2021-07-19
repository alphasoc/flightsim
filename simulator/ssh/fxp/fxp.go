// Package fxp handles ssh fxp packet abstractions.
package fxp

import (
	"bytes"
	"errors"

	"github.com/alphasoc/flightsim/utils"
)

// Only SSH_FX_OK error code is needed for our purposes.
const SSH_FX_OK = 0

const (
	TypeCodeInit           = 1
	TypeCodeVersion        = 2
	TypeCodeOpen           = 3
	TypeCodeClose          = 4
	TypeCodeRead           = 5
	TypeCodeWrite          = 6
	TypeCodeLStat          = 7
	TypeCodeFStat          = 8
	TypeCodeSetStat        = 9
	TypeCodeFSetStat       = 10
	TypeCodeOpenDir        = 11
	TypeCodeReadDir        = 12
	TypeCodeRemove         = 13
	TypeCodeMkDir          = 14
	TypeCodeRmDir          = 15
	TypeCodeRealPath       = 16
	TypeCodeStat           = 17
	TypeCodeRename         = 18
	TypeCodeReadLink       = 19
	TypeCodeSymLink        = 20
	TypeCodeStatus         = 101
	TypeCodeHandle         = 102
	TypeCodeData           = 103
	TypeCodeName           = 104
	TypeCodeAttrs          = 105
	TypeCodeExtended       = 200
	TypeCodeExtendedReplay = 201
	TypeCodeUnknown        = 255
)

// ErrSize signals invalid size of the packet.
var ErrSize = errors.New("Packet too short")

// ErrUnimplemented signals that feature is not implemented.
var ErrUnimplemented = errors.New("Feature is not implemented")

// FieldParser represents read only parser for sftp packet data fields.
type FieldParser struct {
	buffer bytes.Buffer
	err    error
}

// NewFieldParser returns new field parser, it has to be initialized with data, which later
// will be parsed.
func NewFieldParser(in []byte) *FieldParser {
	var buffer bytes.Buffer
	buffer.Write(in)
	return &FieldParser{buffer, nil}
}

// GetUint32 returns uint32 parsed from data supplied in NewFieldParser() and removes used
// bytes from internal buffer.
func (f *FieldParser) ReadUint32() uint32 {
	if f.buffer.Len() < 4 {
		f.err = ErrSize
		return 0
	}
	return utils.UnmarshalUint32(f.buffer.Next(4))
}

// GetUint64 returns uint64 parsed from data supplied in NewFieldParser() and removes used
// bytes from internal buffer
func (f *FieldParser) ReadUint64() uint64 {
	if f.buffer.Len() < 8 {
		f.err = ErrSize
		return 0
	}
	return utils.UnmarshalUint64(f.buffer.Next(8))
}

// GetString returns DataString parsed from data supplied in NewFieldParser() and removes
// used bytes from internal buffer
func (f *FieldParser) ReadString() string {
	if f.buffer.Len() < 4 {
		f.err = ErrSize
		return ""
	}
	length := utils.UnmarshalUint32(f.buffer.Next(4))
	if f.buffer.Len() < int(length) {
		f.err = ErrSize
		return ""
	}
	return string(f.buffer.Next(int(length)))
}

// GetError returns last encountered error.
func (f *FieldParser) GetError() error {
	return f.err
}

// fxp Version packet representation.
type Version struct {
	Version uint32
}

func (p *Version) Marshal() []byte {
	return utils.MarshalUint32(p.Version)
}

func (p *Version) Unmarshal([]byte) error {
	return ErrUnimplemented
}

func (p *Version) GetCode() byte {
	return TypeCodeVersion
}

// fxp Handle packet representation.
type Handle struct {
	ID     uint32
	Handle string
}

func (p *Handle) Marshal() []byte {
	result := make([]byte, 0)
	result = append(result, utils.MarshalUint32(p.ID)...)
	result = append(result, utils.MarshalString(p.Handle)...)
	return result
}

func (p *Handle) Unmarshal([]byte) error {
	return ErrUnimplemented
}

func (p *Handle) GetCode() byte {
	return TypeCodeHandle
}

// fxp Status packet representation.
type Status struct {
	ID      uint32
	ErrCode uint32
	ErrMsg  string
	Lang    string
}

func (p *Status) Marshal() []byte {
	result := make([]byte, 0)
	result = append(result, utils.MarshalUint32(p.ID)...)
	result = append(result, utils.MarshalUint32(p.ErrCode)...)
	result = append(result, utils.MarshalString(p.ErrMsg)...)
	result = append(result, utils.MarshalString(p.Lang)...)
	return result
}

func (p *Status) Unmarshal(data []byte) error {
	parser := NewFieldParser(data)
	p.ID = parser.ReadUint32()
	p.ErrCode = parser.ReadUint32()
	p.ErrMsg = parser.ReadString()
	p.Lang = parser.ReadString()
	return parser.GetError()
}

func (p *Status) GetCode() byte {
	return TypeCodeStatus
}

// fxp Init packet representation.
type Init struct {
	Version uint32
}

func (p *Init) Marshal() []byte {
	return utils.MarshalUint32(p.Version)
}

func (p *Init) Unmarshal(data []byte) error {
	parser := NewFieldParser(data)
	p.Version = parser.ReadUint32()
	return parser.GetError()
}

func (p *Init) GetCode() byte {
	return TypeCodeInit
}

// fxp Open (file) packet representation
type Open struct {
	ID       uint32
	Filename string
	Flags    uint32
}

func (p *Open) Marshal() []byte {
	result := make([]byte, 0)
	result = append(result, utils.MarshalUint32(p.ID)...)
	result = append(result, utils.MarshalString(p.Filename)...)
	result = append(result, utils.MarshalUint32(p.Flags)...)
	result = append(result, []byte{0, 0, 0, 0}...)
	return result
}

func (p *Open) Unmarshal(data []byte) error {
	parser := NewFieldParser(data)
	p.ID = parser.ReadUint32()
	p.Filename = parser.ReadString()
	p.Flags = parser.ReadUint32()
	return parser.GetError()
}

func (p *Open) GetCode() byte {
	return TypeCodeOpen
}

// fxp Write packet representation.
type Write struct {
	ID     uint32
	Handle string
	Offset uint64
	Data   string
}

func (p *Write) Marshal() []byte {
	result := make([]byte, 0)
	result = append(result, utils.MarshalUint32(p.ID)...)
	result = append(result, utils.MarshalString(p.Handle)...)
	result = append(result, utils.MarshalUint64(p.Offset)...)
	result = append(result, utils.MarshalString(p.Data)...)
	return result
}

func (p *Write) Unmarshal(data []byte) error {
	parser := NewFieldParser(data)
	p.ID = parser.ReadUint32()
	p.Handle = parser.ReadString()
	p.Offset = parser.ReadUint64()
	p.Data = parser.ReadString()
	return parser.GetError()
}

func (p *Write) GetCode() byte {
	return TypeCodeWrite
}

// fxp Close packet representation.
type Close struct {
	ID     uint32
	Handle string
}

func (p *Close) Marshal() []byte {
	result := make([]byte, 0)
	result = append(result, utils.MarshalUint32(p.ID)...)
	result = append(result, utils.MarshalString(p.Handle)...)
	return result
}

func (p *Close) Unmarshal(data []byte) error {
	parser := NewFieldParser(data)
	p.ID = parser.ReadUint32()
	p.Handle = parser.ReadString()
	return parser.GetError()
}

func (p *Close) GetCode() byte {
	return TypeCodeClose
}
