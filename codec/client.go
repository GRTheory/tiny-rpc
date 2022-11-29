package codec

import (
	"bufio"
	"hash/crc32"
	"io"
	"net/rpc"
	"sync"

	"github.com/GRTheory/tiny-rpc/compressor"
	"github.com/GRTheory/tiny-rpc/header"
	"github.com/GRTheory/tiny-rpc/serializer"
)

type clientCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	compressor compressor.CompressType
	serializer serializer.Serializer
	response   header.ResponseHeader
	mutex      sync.Mutex
	pending    map[uint64]string
}

func NewClientCodec(conn io.ReadWriteCloser,
	compressType compressor.CompressType, serializer serializer.Serializer) rpc.ClientCodec {

	return &clientCodec{
		r : bufio.NewReader(conn),
		w : bufio.NewWriter(conn),
		c: conn,
		compressor: compressType,
		serializer: serializer,
		pending: make(map[uint64]string),
	}
}

func (c *clientCodec) WriteRequest(r *rpc.Request, param interface{}) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()

	if _, ok := compressor.Compressors[c.compressor]; !ok {
		return ErrorNotFoundCompressor
	}
	reqBody, err := c.serializer.Marshal(param)
	if err != nil {
		return err
	}
	compressedReqBody, err := compressor.Compressors[c.compressor].Zip(reqBody)
	if err != nil {
		return err
	}
	h := header.ReqeustPool.Get().(*header.RequestHeader)
	defer func() {
		h.ResetHeaer()
		header.ReqeustPool.Put(h)
	}()
	h.ID = r.Seq
	h.Method = r.ServiceMethod
	h.RequestLen = uint32(len(compressedReqBody))
	h.CompressType = compressor.CompressType(c.compressor)
	h.Checksum = crc32.ChecksumIEEE(compressedReqBody)

	if err := sendFrame(c.w, h.Marshal()); err != nil {
		return err
	}
	if err := write(c.w, compressedReqBody); err != nil {
		return err
	}

	c.w.(*bufio.Writer).Flush()
	return nil
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	c.response.ResetHeader()
	data, err := recvFrame(c.r)
	if err != nil {
		return err
	}
	err = c.response.Unmarshal(data)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	r.Seq = c.response.ID
	r.Error = c.response.Error
	r.ServiceMethod = c.pending[r.Seq]
	delete(c.pending, r.Seq)
	c.mutex.Unlock()
	return nil
}

func (c *clientCodec) ReadResponseBody(param interface{}) error {
	if param == nil {
		if c.response.ResponseLen != 0 {
			if err := read(c.r, make([]byte, c.response.ResponseLen)); err != nil {
				return err
			}
		}
		return nil
	}

	respBody := make([]byte, c.response.ResponseLen)
	err := read(c.r, respBody)
	if err != nil {
		return err
	}

	if c.response.Checksum != 0 {
		if crc32.ChecksumIEEE(respBody) != c.response.Checksum {
			return ErrorUnexpectedChecksum
		}
	}

	if c.response.GetCompressType() != c.compressor {
		return ErrorCompressorTypeMismatch
	}

	resp, err := compressor.Compressors[c.response.GetCompressType()].Unzip(respBody)
	if err != nil {
		return err
	}

	return c.serializer.Unmarshal(resp, param)
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}
