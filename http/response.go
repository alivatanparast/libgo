/* For license and copyright information please see the LEGAL file in the code repository */

package http

import (
	"io"
	"strconv"
	"strings"

	"github.com/GeniusesGroup/libgo/codec"
	"github.com/GeniusesGroup/libgo/convert"
	"github.com/GeniusesGroup/libgo/protocol"
)

// Response is represent response protocol structure!
// https://tools.ietf.org/html/rfc2616#section-6
type Response struct {
	version      string
	statusCode   string
	reasonPhrase string

	H header // Exported field to let consumers use other methods that protocol.HTTPHeader
	body
}

func (r *Response) Init() { 
	r.H.Init()
	r.body.Init()
}
func (r *Response) Reinit() {
	r.version = ""
	r.statusCode = ""
	r.reasonPhrase = ""
	r.H.Reinit()
	r.body.Reinit()
}
func (r *Response) Deinit() {
	r.H.Deinit()
	r.body.Deinit()
}

func (r *Response) Version() string               { return r.version }
func (r *Response) StatusCode() string            { return r.statusCode }
func (r *Response) ReasonPhrase() string          { return r.reasonPhrase }
func (r *Response) SetVersion(version string)     { r.version = version }
func (r *Response) SetStatus(code, phrase string) { r.statusCode = code; r.reasonPhrase = phrase }
func (r *Response) Header() protocol.HTTPHeader   { return &r.H }

// GetStatusCode get status code as uit16
func (r *Response) GetStatusCode() (code uint16, err protocol.Error) {
	// TODO::: don't use strconv for such simple task
	var c, goErr = strconv.ParseUint(r.statusCode, 10, 16)
	if goErr != nil {
		return 0, &ErrParseStatusCode
	}
	return uint16(c), nil
}

// GetError return related protocol.Error in header of the Response
func (r *Response) GetError() (err protocol.Error) {
	var errIDString = r.H.Get(HeaderKeyErrorID)
	var errID, _ = strconv.ParseUint(errIDString, 10, 64)
	if errID == 0 {
		return
	}
	err = protocol.App.GetErrorByID(protocol.ID(errID))
	return
}

// SetError set given protocol.Error to header of the response
func (r *Response) SetError(err protocol.Error) {
	r.H.Set(HeaderKeyErrorID, err.IDasString())
}

// Redirect set given status and target location to the response
// httpRes.Redirect(http.StatusMovedPermanentlyCode, http.StatusMovedPermanentlyPhrase, "http://www.google.com/")
func (r *Response) Redirect(code, phrase string, target string) {
	r.SetStatus(code, phrase)
	r.H.Set(HeaderKeyLocation, target)
}

//libgo:impl protocol.Codec
func (r *Response) MediaType() protocol.MediaType       { return &MediaTypeResponse }
func (r *Response) CompressType() protocol.CompressType { return nil }
func (r *Response) Len() (ln int) {
	ln = r.LenWithoutBody()
	ln += r.body.Len()
	return
}
func (r *Response) Decode(source protocol.Codec) (n int, err protocol.Error) {
	if source.Len() > MaxHTTPHeaderSize {
		// err =
		return
	}

	// Make a buffer to hold incoming data.
	// TODO::: change to get from buffer pool??
	var buf = make([]byte, 0, MaxHTTPHeaderSize)
	// Read the incoming connection into the buffer.
	buf, err = source.MarshalTo(buf)
	if err != nil {
		// err = connection.ErrNoConnection
		return
	}

	buf, err = r.UnmarshalFrom(buf)
	if err != nil {
		return
	}
	err = r.body.checkAndSetCodecAsIncomeBody(buf, source, &r.H)
	return
}
func (r *Response) Encode(destination protocol.Codec) (n int, err protocol.Error) {
	var lenWithoutBody = r.LenWithoutBody()
	var bodyLen = r.body.Len()
	var wholeLen = lenWithoutBody + bodyLen
	// Check if whole request has fewer length than MaxHTTPHeaderSize and Decide to send header and body separately
	if wholeLen > MaxHTTPHeaderSize {
		var withoutBody = make([]byte, 0, lenWithoutBody)
		withoutBody = r.MarshalToWithoutBody(withoutBody)

		n, err = destination.Unmarshal(withoutBody)
		if err == nil && r.body.Codec != nil {
			var bodyWrote int
			bodyWrote, err = destination.Encode(&r.body)
			n += bodyWrote
		}
	} else {
		var httpPacket = make([]byte, 0, wholeLen)
		httpPacket, err = r.MarshalTo(httpPacket)
		n, err = destination.Unmarshal(httpPacket)
	}
	return
}

// Marshal encodes whole r *Response data and return httpPacket!
func (r *Response) Marshal() (httpPacket []byte, err protocol.Error) {
	httpPacket = make([]byte, 0, r.Len())
	httpPacket, err = r.MarshalTo(httpPacket)
	return
}

// MarshalTo encodes whole r *Response data to given httpPacket and return it by new len!
func (r *Response) MarshalTo(httpPacket []byte) (added []byte, err protocol.Error) {
	httpPacket = append(httpPacket, r.version...)
	httpPacket = append(httpPacket, SP)
	httpPacket = append(httpPacket, r.statusCode...)
	httpPacket = append(httpPacket, SP)
	httpPacket = append(httpPacket, r.reasonPhrase...)
	httpPacket = append(httpPacket, CRLF...)

	httpPacket = r.H.MarshalTo(httpPacket)
	httpPacket = append(httpPacket, CRLF...)

	httpPacket, err = r.body.MarshalTo(httpPacket)
	added = httpPacket
	return
}

// Unmarshal parses and decodes data of given httpPacket to r *Response.
// In some bad packet may occur panic, handle panic by recover otherwise app will crash and exit!
func (r *Response) Unmarshal(httpPacket []byte) (n int, err protocol.Error) {
	var maybeBody []byte
	maybeBody, err = r.UnmarshalFrom(httpPacket)
	if err != nil {
		return
	}
	err = r.body.checkAndSetIncomeBody(maybeBody, &r.H)
	n = len(httpPacket)
	return
}

// UnmarshalFrom parses and decodes data of given httpPacket to r *Response.
// In some bad packet may occur panic, handle panic by recover otherwise app will crash and exit!
func (r *Response) UnmarshalFrom(httpPacket []byte) (maybeBody []byte, err protocol.Error) {
	// By use unsafe pointer here all strings assign in Response will just point to httpPacket slice
	// and no need to alloc lot of new memory locations and copy response line and headers keys & values!
	var s = convert.UnsafeByteSliceToString(httpPacket)

	// si hold s index and i hold s index in new sliced state.
	var si, i int

	// First line: HTTP/1.0 200 OK
	i = strings.IndexByte(s[:versionMaxLength], SP)
	if i == -1 {
		maybeBody = httpPacket[si:]
		err = &ErrParseVersion
		return
	}
	r.version = s[:i]
	i++ // +1 due to have ' '
	si = i
	s = s[i:]

	i = strings.IndexByte(s[:statusCodeMaxLength], SP)
	if i == -1 {
		maybeBody = httpPacket[si:]
		err = &ErrParseStatusCode
		return
	}
	r.statusCode = s[:i]
	i++ // +1 due to have ' '
	si += i
	s = s[i:]

	i = strings.IndexByte(s, '\r')
	if i == -1 {
		maybeBody = httpPacket[si:]
		err = &ErrParseReasonPhrase
		return
	}
	r.reasonPhrase = s[:i]
	i += 2 // +2 due to have "\r\n"
	si += i
	s = s[i:]

	i, err = r.H.unmarshal(s)
	if err != nil {
		maybeBody = httpPacket[i:]
		return
	}
	si += i
	// By https://tools.ietf.org/html/rfc2616#section-4 very simple http packet must end with CRLF even packet without header or body!
	// So it can be occur panic if very simple request end without any CRLF
	si += 2 // +2 due to have "\r\n" after header end

	return httpPacket[si:], nil
}

/*
********** protocol.Buffer interface **********
 */

// ReadFrom decodes r *Response data by read from given io.Reader!
// Declare to respect io.ReaderFrom interface!
func (r *Response) ReadFrom(reader io.Reader) (n int64, goErr error) {
	// Make a buffer to hold incoming data.
	var buf = make([]byte, MaxHTTPHeaderSize)
	var headerReadLength int
	var err protocol.Error

	// Read the incoming connection into the buffer.
	headerReadLength, goErr = reader.Read(buf)
	if goErr != nil || headerReadLength == 0 {
		return
	}

	buf = buf[:headerReadLength]
	buf, err = r.UnmarshalFrom(buf)
	if err != nil {
		return int64(headerReadLength), err
	}
	err = r.body.checkAndSetReaderAsIncomeBody(buf, codec.ReaderAdaptor{reader}, &r.H)
	n = int64(headerReadLength)
	return
}

// WriteTo encodes r *Response data and write it to given io.Writer!
// Declare to respect io.WriterTo interface!
func (r *Response) WriteTo(writer io.Writer) (n int64, err error) {
	var lenWithoutBody = r.LenWithoutBody()
	var bodyLen = r.body.Len()
	var wholeLen = lenWithoutBody + bodyLen
	// Check if whole request has fewer length than MaxHTTPHeaderSize and Decide to send header and body separately
	if wholeLen > MaxHTTPHeaderSize {
		var httpPacket = make([]byte, 0, lenWithoutBody)
		httpPacket = r.MarshalToWithoutBody(httpPacket)

		var headerWriteLength int
		headerWriteLength, err = writer.Write(httpPacket)
		if err == nil && r.body.Codec != nil {
			n, err = r.body.WriteTo(writer)
		}
		n += int64(headerWriteLength)
	} else {
		var httpPacket = make([]byte, 0, wholeLen)
		httpPacket, _ = r.MarshalTo(httpPacket)
		var packetWriteLength int
		packetWriteLength, err = writer.Write(httpPacket)
		n = int64(packetWriteLength)
	}
	return
}

/*
********** local methods **********
 */

// MarshalWithoutBody encodes r *Response data and return httpPacket without body part!
func (r *Response) MarshalWithoutBody() (httpPacket []byte) {
	httpPacket = make([]byte, 0, r.LenWithoutBody())
	httpPacket = r.MarshalToWithoutBody(httpPacket)
	return
}

// MarshalToWithoutBody encodes r *Response data and return httpPacket without body part!
func (r *Response) MarshalToWithoutBody(httpPacket []byte) []byte {
	httpPacket = append(httpPacket, r.version...)
	httpPacket = append(httpPacket, SP)
	httpPacket = append(httpPacket, r.statusCode...)
	httpPacket = append(httpPacket, SP)
	httpPacket = append(httpPacket, r.reasonPhrase...)
	httpPacket = append(httpPacket, CRLF...)

	httpPacket = r.H.MarshalTo(httpPacket)
	httpPacket = append(httpPacket, CRLF...)
	return httpPacket
}

// LenWithoutBody return length of response without body length!
func (r *Response) LenWithoutBody() (ln int) {
	ln = 6 // 6==1+1+2+2==len(SP)+len(SP)+len(CRLF)+len(CRLF)
	ln += len(r.version)
	ln += len(r.statusCode)
	ln += len(r.reasonPhrase)
	ln += r.H.Len()
	return
}
