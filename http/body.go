/* For license and copyright information please see the LEGAL file in the code repository */

package http

import (
	"io"

	"github.com/GeniusesGroup/libgo/compress/raw"
	"github.com/GeniusesGroup/libgo/protocol"
)

// body is represent HTTP body.
// Due to many performance impact, MediaType() method of body not return any true data. use header ContentType() method instead. This can be change if ...
// https://datatracker.ietf.org/doc/html/rfc2616#section-4.3
type body struct {
	protocol.Codec
}

func (b *body) Init()   {}
func (b *body) Reinit() { b.Codec = nil }
func (b *body) Deinit() {}

func (b *body) Body() protocol.Codec         { return b }
func (b *body) SetBody(codec protocol.Codec) { b.Codec = codec }

//libgo:impl protocol.Codec
func (b *body) Len() int {
	if b.Codec != nil {
		return b.Codec.Len()
	}
	return 0
}
func (b *body) MediaType() protocol.MediaType {
	if b.Codec != nil {
		return b.Codec.MediaType()
	}
	return nil
}
func (b *body) CompressType() protocol.CompressType {
	if b.Codec != nil {
		return b.Codec.CompressType()
	}
	return nil
}
func (b *body) Decode(source protocol.Codec) (n int, err protocol.Error) {
	if b.Codec != nil {
		n, err = b.Codec.Decode(source)
	}
	return
}
func (b *body) Encode(destination protocol.Codec) (n int, err protocol.Error) {
	if b.Codec != nil {
		n, err = b.Codec.Encode(destination)
	}
	return
}
func (b *body) Marshal() (data []byte, err protocol.Error) {
	if b.Codec != nil {
		data, err = b.Codec.Marshal()
	}
	return
}
func (b *body) MarshalTo(data []byte) (added []byte, err protocol.Error) {
	if b.Codec != nil {
		return b.Codec.MarshalTo(data)
	}
	return data, nil
}
func (b *body) Unmarshal(data []byte) (n int, err protocol.Error) {
	if b.Codec != nil {
		n, err = b.Codec.Unmarshal(data)
	}
	return
}
func (b *body) UnmarshalFrom(data []byte) (remaining []byte, err protocol.Error) {
	if b.Codec != nil {
		return b.Codec.UnmarshalFrom(data)
	}
	return
}

// ReadFrom decodes r *Request data by read from given io.Reader
//
//libgo:impl io.ReaderFrom
func (b *body) ReadFrom(reader io.Reader) (n int64, goErr error) {

	return
}

// WriteTo encodes r(*Request) data and write it to given io.Writer
//
//libgo:impl io.WriterTo
func (b *body) WriteTo(writer io.Writer) (n int64, err error) {

	return
}

/*
********** local methods **********
 */

func (b *body) checkAndSetCodecAsIncomeBody(maybeBody []byte, c protocol.Codec, h *header) (err protocol.Error) {
	var transferEncoding, _ = h.TransferEncoding()
	switch transferEncoding {
	case "":
		var contentLength = h.ContentLength()
		var maybeBodyLength = len(maybeBody)
		if maybeBodyLength == int(contentLength) {
			b.setReadedIncomeBody(maybeBody, h)
		} else {
			// TODO::: allow this situation that peer send some part of body with header??
			// Header length maybe other than stream income data length e.g. send body in multiple TCP.PSH flag set.
			if maybeBodyLength > 0 {
				var bodySlice = make([]byte, maybeBodyLength, contentLength)
				copy(bodySlice, maybeBody)
				for {
					bodySlice, err = c.MarshalTo(bodySlice)
					if err != nil || len(bodySlice) == int(contentLength) {
						break
					}
				}
				b.setReadedIncomeBody(bodySlice, h)
			} else {
				b.setCodecAsIncomeBody(c, h)
			}
		}
	case HeaderValueChunked:
		// TODO:::
	default:
		// Like nginx, due to security, we only support a single Transfer-Encoding header field, and
		// only if set to "chunked".
		// err =
	}
	return
}

func (b *body) checkAndSetReaderAsIncomeBody(maybeBody []byte, reader protocol.Reader, h *header) (err protocol.Error) {
	var transferEncoding, _ = h.TransferEncoding()
	switch transferEncoding {
	case "":
		var contentLength = h.ContentLength()
		var maybeBodyLength = len(maybeBody)
		if maybeBodyLength == int(contentLength) {
			b.setReadedIncomeBody(maybeBody, h)
		} else {
			// Header length maybe other than stream income data length e.g. send body in multiple TCP.PSH flag set.
			if maybeBodyLength > 0 {
				var bodyReadLength int
				var goErr error
				var bodySlice = make([]byte, contentLength)
				bodyReadLength, goErr = reader.Read(bodySlice[maybeBodyLength:])
				if goErr != nil {
					// err =
					return
				}
				if bodyReadLength+maybeBodyLength != int(contentLength) {
					// err =
					return
				}
				copy(bodySlice, maybeBody)
				b.setReadedIncomeBody(bodySlice, h)
			} else {
				b.setReaderAsIncomeBody(reader, h, contentLength)
			}
		}
	case HeaderValueChunked:
		// TODO:::
	default:
		// Like nginx, due to security, we only support a single Transfer-Encoding header field, and
		// only if set to "chunked".
		// err =
	}
	return
}

// Call this method just if body marshaled with first line and headers.
func (b *body) checkAndSetIncomeBody(maybeBody []byte, h *header) (err protocol.Error) {
	var maybeBodyLength = len(maybeBody)
	if maybeBodyLength > 0 {
		var contentLength = h.ContentLength()
		if maybeBodyLength == int(contentLength) {
			b.setReadedIncomeBody(maybeBody, h)
		} else {
			// err =
		}
	}
	return
}

func (b *body) setCodecAsIncomeBody(c protocol.Codec, h *header) (err protocol.Error) {
	var contentEncoding, _ = h.ContentEncoding()
	if contentEncoding == "" {
		b.Codec = c
		return
	}

	var compressType protocol.CompressType
	compressType, err = protocol.OS.GetCompressTypeByContentEncoding(contentEncoding)
	if err != nil {
		return
	}
	b.Codec, err = compressType.Decompress(c)
	return
}

func (b *body) setReaderAsIncomeBody(reader protocol.Reader, h *header, contentLength uint64) (err protocol.Error) {
	var contentEncoding, _ = h.ContentEncoding()
	if contentEncoding == "" {
		b.Codec, err = raw.RAW.DecompressFromReader(reader, int(contentLength))
		return
	}

	var compressType protocol.CompressType
	compressType, err = protocol.OS.GetCompressTypeByContentEncoding(contentEncoding)
	if err != nil {
		return
	}
	b.Codec, err = compressType.DecompressFromReader(reader, int(contentLength))
	return
}

func (b *body) setReadedIncomeBody(body []byte, h *header) (err protocol.Error) {
	var contentEncoding, _ = h.ContentEncoding()
	if contentEncoding == "" {
		b.Codec, err = raw.RAW.DecompressFromSlice(body)
		return
	}

	var compressType protocol.CompressType
	compressType, err = protocol.OS.GetCompressTypeByContentEncoding(contentEncoding)
	if err != nil {
		return
	}
	b.Codec, err = compressType.DecompressFromSlice(body)
	return
}
