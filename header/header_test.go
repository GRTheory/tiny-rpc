package header

import (
	"reflect"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestRequestHeader_Marshal(t *testing.T) {
	header := &RequestHeader{
		CompressType: 0,
		Method:       "Add",
		ID:           12455,
		RequestLen:   266,
		Checksum:     3845236589,
	}

	assert.Equal(t, []byte{0x0, 0x0, 0x3, 0x41, 0x64, 0x64,
		0xa7, 0x61, 0x8a, 0x2, 0x6d, 0xa7, 0x31, 0xe5}, header.Marshal())
}

func TestRequestHeader_Unmarshal(t *testing.T) {
	type expect struct {
		header *RequestHeader
		err    error
	}

	cases := []struct {
		name   string
		data   []byte
		expect expect
	}{
		{
			"test-1",
			[]byte{0x0, 0x0, 0x3, 0x41, 0x64, 0x64,
				0xa7, 0x61, 0x8a, 0x2, 0x6d, 0xa7, 0x31, 0xe5},
			expect{&RequestHeader{
				CompressType: 0,
				Method:       "Add",
				ID:           12455,
				RequestLen:   266,
				Checksum:     3845236589,
			}, nil},
		},
	}

	for _, c := range cases{
		t.Run(c.name, func(t *testing.T) {
			h := &RequestHeader{}
			err := h.Unmarshal(c.data)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.header, h))
			assert.Equal(t, err, c.expect.err)
		})
	}
}

