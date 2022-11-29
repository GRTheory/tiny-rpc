package codec

import "errors"

var (
	ErrorInvalidSequence        = errors.New("invalid sequence number in response")
	ErrorUnexpectedChecksum     = errors.New("unexpected checksum")
	ErrorNotFoundCompressor     = errors.New("not found compressor")
	ErrorCompressorTypeMismatch = errors.New("request and response Compressor type mismatch")
)
