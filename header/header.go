package header

import (
	"encoding/binary"
	"errors"
	"sync"

	"github.com/GRTheory/tiny-rpc/compressor"
)

const (
	MaxHeaderSize = 36

	Uint32Size = 4
	Uint16Size = 2
)

var ErrorUnmarshal = errors.New("an error occurred in Unmarshal")

type RequestHeader struct {
	sync.RWMutex
	CompressType compressor.CompressType
	Method       string
	ID           uint64
	RequestLen   uint32
	Checksum     uint32
}

func (r *RequestHeader) Marshal() []byte {
	r.RLock()
	defer r.RUnlock()
	idx := 0
	// MaxHeaderSize = 2 + 10 + len(string) + 10 + 10 + 4
	header := make([]byte, MaxHeaderSize+len(r.Method))

	binary.LittleEndian.PutUint16(header[idx:], uint16(r.CompressType))
	idx += Uint16Size

	idx += writeString(header[idx:], r.Method)
	idx += binary.PutUvarint(header[idx:], r.ID)
	idx += binary.PutUvarint(header[idx:], uint64(r.RequestLen))

	binary.LittleEndian.PutUint32(header[idx:], r.Checksum)
	idx += Uint32Size

	return header[:idx]
}

func (r *RequestHeader) Unmarshal(data []byte) (err error) {
	r.Lock()
	defer r.Unlock()
	if len(data) == 0 {
		return ErrorUnmarshal
	}

	defer func() {
		if r := recover(); r != nil {
			err = ErrorUnmarshal
		}
	}()
	idx, size := 0, 0
	r.CompressType = compressor.CompressType(binary.LittleEndian.Uint16(data[idx:]))
	idx += Uint16Size

	r.Method, size = readString(data[idx:])
	idx += size

	r.ID, size = binary.Uvarint(data[idx:])
	idx += size

	length, size := binary.Uvarint(data[idx:])
	r.RequestLen = uint32(length)
	idx += size

	r.Checksum = binary.LittleEndian.Uint32(data[idx:])
	return
}

func (r *RequestHeader) GetCompressType() compressor.CompressType {
	r.RLock()
	defer r.RUnlock()
	return compressor.CompressType(r.CompressType)
}

func (r *RequestHeader) ResetHeaer() {
	r.Lock()
	defer r.Unlock()
	r.ID = 0
	r.Checksum = 0
	r.Method = ""
	r.CompressType = 0
	r.RequestLen = 0
}

type ResponseHeader struct {
	sync.RWMutex
	CompressType compressor.CompressType
	ID           uint64
	Error        string
	ResponseLen  uint32
	Checksum     uint32
}

func (r *ResponseHeader) Marshal() []byte {
	r.RLock()
	defer r.RUnlock()
	idx := 0
	header := make([]byte, MaxHeaderSize+len(r.Error))

	binary.LittleEndian.PutUint16(header[idx:], uint16(r.CompressType))
	idx += Uint16Size

	idx += binary.PutUvarint(header[idx:], r.ID)
	idx += writeString(header[idx:], r.Error)
	idx += binary.PutUvarint(header[idx:], uint64(r.ResponseLen))

	binary.LittleEndian.PutUint32(header[idx:], r.Checksum)
	idx += Uint32Size
	return header[:idx]
}

func (r *ResponseHeader) Unmarshal(data []byte) (err error) {
	r.Lock()
	defer r.Unlock()
	if len(data) == 0 {
		return ErrorUnmarshal
	}

	defer func() {
		if r := recover(); r != nil {
			err = ErrorUnmarshal
		}
	}()

	idx, size := 0, 0
	r.CompressType = compressor.CompressType(r.CompressType)
	idx += Uint16Size

	r.ID, size = binary.Uvarint(data[idx:])
	idx += size

	r.Error, size = readString(data[idx:])
	idx += size

	length, size := binary.Uvarint(data[idx:])
	r.ResponseLen = uint32(length)
	idx += size

	r.Checksum = binary.LittleEndian.Uint32(data[idx:])
	return
}

func (r *ResponseHeader) GetCompressType() compressor.CompressType {
	r.RLock()
	defer r.RUnlock()
	return compressor.CompressType(r.CompressType)
}

func (r *ResponseHeader) ResetHeader() {
	r.Lock()
	defer r.Unlock()
	r.Error = ""
	r.ID = 0
	r.CompressType = 0
	r.Checksum = 0
	r.ResponseLen = 0
}

func readString(data []byte) (string, int) {
	idx := 0
	length, size := binary.Uvarint(data)
	idx += size
	str := string(data[idx : idx+int(length)])
	idx += len(str)
	return str, idx
}

func writeString(data []byte, str string) int {
	idx := 0
	idx += binary.PutUvarint(data, uint64(len(str)))
	copy(data[idx:], str)
	idx += len(str)
	return idx
}
