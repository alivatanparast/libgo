/* For license and copyright information please see LEGAL file in repository */

package json

import (
	"bytes"
	"encoding/base64"
	"strconv"

	"../convert"
	"../protocol"
)

// DecoderMinifed store data to decode data by each method!
type DecoderMinifed struct {
	Buf      []byte
	Token    byte
	LastItem []byte
}

// Offset make d.Buf to start of given offset
func (d *DecoderMinifed) Offset(o int) {
	d.Buf = d.Buf[o:]
}

// FindEndToken find next end json token
func (d *DecoderMinifed) FindEndToken() {
	for i, c := range d.Buf {
		switch c {
		case ',':
			d.Token = ','
			d.LastItem = d.Buf[:i]
			d.Buf = d.Buf[i:]
			return
		case ']':
			d.Token = ']'
			d.LastItem = d.Buf[:i]
			d.Buf = d.Buf[i:]
			return
		case '}':
			d.Token = '}'
			d.LastItem = d.Buf[:i]
			d.Buf = d.Buf[i:]
			return
		}
	}
}

// ResetToken set d.Token to nil
func (d *DecoderMinifed) ResetToken() {
	d.Token = 0
}

// CheckToken set d.Token to nil
func (d *DecoderMinifed) CheckToken(t byte) bool {
	if d.Token == t {
		d.ResetToken()
		return true
	}
	return false
}

// DecodeKey return json key. pass d.Buf start from after {||, and receive from after :
func (d *DecoderMinifed) DecodeKey() string {
	d.Offset(2)
	var loc = bytes.IndexByte(d.Buf, '"')
	var slice []byte = d.Buf[:loc]
	d.Offset(loc + 2) // +2 due to have '":' after key name end!
	return convert.UnsafeByteSliceToString(slice)
}

// NotFoundKey call in default switch of each decode iteration
func (d *DecoderMinifed) NotFoundKey() (err protocol.Error) {
	d.FindEndToken()
	return
}

// NotFoundKeyStrict call in default switch of each decode iteration in strict mode.
func (d *DecoderMinifed) NotFoundKeyStrict() protocol.Error {
	return ErrEncodedIncludeNotDeffiendKey
}

// DecodeBool convert 64bit integer number string to number. pass d.Buf start from after : and receive from after ,
func (d *DecoderMinifed) DecodeBool() (b bool, err protocol.Error) {
	if d.Buf[0] == 't' {
		b = true
		d.Offset(5) // true,
	} else {
		// b = false
		d.Offset(6) // false,
	}
	return
}

// DecodeUInt8 convert 8bit integer number string to number. pass d.Buf start from number and receive from after ,
func (d *DecoderMinifed) DecodeUInt8() (ui uint8, err protocol.Error) {
	d.FindEndToken()
	ui, err = convert.StringToUint8Base10(convert.UnsafeByteSliceToString(d.LastItem))
	if err != nil {
		err = ErrEncodedIntegerCorrupted
		return
	}
	return
}

// DecodeUInt16 convert 16bit integer number string to number. pass d.Buf start from number and receive from after ,
func (d *DecoderMinifed) DecodeUInt16() (ui uint16, err protocol.Error) {
	d.FindEndToken()
	ui, err = convert.StringToUint16Base10(convert.UnsafeByteSliceToString(d.LastItem))
	if err != nil {
		err = ErrEncodedIntegerCorrupted
		return
	}
	return
}

// DecodeUInt32 convert 32bit integer number string to number. pass d.Buf start from number and receive from after ,
func (d *DecoderMinifed) DecodeUInt32() (ui uint32, err protocol.Error) {
	d.FindEndToken()
	ui, err = convert.StringToUint32Base10(convert.UnsafeByteSliceToString(d.LastItem))
	if err != nil {
		err = ErrEncodedIntegerCorrupted
		return
	}
	return
}

// DecodeUInt64 convert 64bit integer number string to number. pass d.Buf start from number and receive from after ,
func (d *DecoderMinifed) DecodeUInt64() (ui uint64, err protocol.Error) {
	d.FindEndToken()
	ui, err = convert.StringToUint64Base10(convert.UnsafeByteSliceToString(d.LastItem))
	if err != nil {
		err = ErrEncodedIntegerCorrupted
		return
	}
	return
}

// DecodeInt32 convert 32bit number string to number. pass d.Buf start from number and receive from after end of number
func (d *DecoderUnsafe) DecodeInt32() (i int32, err protocol.Error) {
	d.FindEndToken()
	var goErr error
	var num int64
	num, goErr = strconv.ParseInt(convert.UnsafeByteSliceToString(d.LastItem), 10, 32)
	if goErr != nil {
		return 0, ErrEncodedIntegerCorrupted
	}
	i = int32(num)
	return
}

// DecodeInt64 convert 64bit number string to number. pass d.Buf start from number and receive from after ,
func (d *DecoderMinifed) DecodeInt64() (i int64, err protocol.Error) {
	d.FindEndToken()
	var goErr error
	i, goErr = strconv.ParseInt(convert.UnsafeByteSliceToString(d.LastItem), 10, 64)
	if goErr != nil {
		return 0, ErrEncodedStringCorrupted
	}
	return
}

// DecodeFloat64AsNumber convert float64 number string to float64 number. pass d.Buf start from number and receive from ,
func (d *DecoderMinifed) DecodeFloat64AsNumber() (f float64, err protocol.Error) {
	d.FindEndToken()
	var goErr error
	f, goErr = strconv.ParseFloat(convert.UnsafeByteSliceToString(d.LastItem), 64)
	if goErr != nil {
		return 0, ErrEncodedStringCorrupted
	}
	return
}

// DecodeString return string. pass d.Buf start from after " and receive from from after "
func (d *DecoderMinifed) DecodeString() (s string) {
	var loc int // Coma, Colon, bracket, ... location
	loc = bytes.IndexByte(d.Buf, '"')
	if loc < 0 {
		// Reach last item of d.Buf!
		loc = len(d.Buf) - 1
	}

	var slice []byte = d.Buf[:loc]
	d.Offset(loc + 1)
	return string(slice)
}

/*
	Array part
*/

// DecodeByteArrayAsBase64 convert base64 string to [n]byte
func (d *DecoderMinifed) DecodeByteArrayAsBase64(array []byte) (err protocol.Error) {
	d.Offset(1) // due to have " at start

	var loc = bytes.IndexByte(d.Buf, '"')
	if loc < 0 {
		err = ErrEncodedArrayCorrupted
		return
	}

	var goErr error
	_, goErr = base64.RawStdEncoding.Decode(array, d.Buf[:loc])
	if goErr != nil {
		return ErrEncodedArrayCorrupted
	}

	d.Offset(loc + 1)
	return
}

// DecodeByteArrayAsNumber convert number array to [n]byte
func (d *DecoderMinifed) DecodeByteArrayAsNumber(array []byte) (err protocol.Error) {
	var value uint8
	for i := 0; i < len(array); i++ {
		d.Offset(1) // due to have [ or ,
		value, err = d.DecodeUInt8()
		if err != nil {
			err = ErrEncodedArrayCorrupted
			return
		}
		array[i] = value
	}
	if d.Buf[0] != ']' {
		err = ErrEncodedArrayCorrupted
	}
	d.Offset(1)
	return
}

/*
	Slice as Number
*/

// DecodeByteSliceAsNumber convert number string slice to []byte. pass buf start from after [ and receive from after ]
func (d *DecoderMinifed) DecodeByteSliceAsNumber() (slice []byte, err protocol.Error) {
	d.Offset(1) // due to have [ at start
	slice = make([]byte, 0, 8) // TODO::: Is cap efficient enough?

	var num uint8
	for !d.CheckToken(']') {
		num, err = d.DecodeUInt8()
		if err != nil {
			err = ErrEncodedSliceCorrupted
			return
		}
		slice = append(slice, num)
		d.Offset(1)
	}
	return
}

// DecodeUInt16SliceAsNumber convert uint16 number string slice to []byte. pass buf start from after [ and receive from after ]
func (d *DecoderMinifed) DecodeUInt16SliceAsNumber() (slice []uint16, err protocol.Error) {
	d.Offset(1) // due to have [ at start
	slice = make([]uint16, 0, 8) // TODO::: Is cap efficient enough?

	var num uint16
	for !d.CheckToken(']') {
		num, err = d.DecodeUInt16()
		if err != nil {
			err = ErrEncodedSliceCorrupted
			return
		}
		slice = append(slice, num)
		d.Offset(1)
	}
	return
}

// DecodeUInt32SliceAsNumber convert uint32 number string slice to []byte. pass buf start from after [ and receive from after ]
func (d *DecoderMinifed) DecodeUInt32SliceAsNumber() (slice []uint32, err protocol.Error) {
	d.Offset(1) // due to have [ at start
	slice = make([]uint32, 0, 8) // TODO::: Is cap efficient enough?

	var num uint32
	for !d.CheckToken(']') {
		num, err = d.DecodeUInt32()
		if err != nil {
			err = ErrEncodedSliceCorrupted
			return
		}
		slice = append(slice, num)
		d.Offset(1)
	}
	return
}

// DecodeUInt64SliceAsNumber convert uint64 number string slice to []byte. pass buf start from after [ and receive from after ]
func (d *DecoderMinifed) DecodeUInt64SliceAsNumber() (slice []uint64, err protocol.Error) {
	d.Offset(1) // due to have [ at start
	slice = make([]uint64, 0, 8) // TODO::: Is cap efficient enough?

	var num uint64
	for !d.CheckToken(']') {
		num, err = d.DecodeUInt64()
		if err != nil {
			err = ErrEncodedSliceCorrupted
			return
		}
		slice = append(slice, num)
		d.Offset(1)
	}
	return
}

/*
	Slice as Base64
*/

// DecodeByteSliceAsBase64 convert base64 string to []byte
func (d *DecoderMinifed) DecodeByteSliceAsBase64() (slice []byte, err protocol.Error) {
	d.Offset(1) // due to have " at start

	// Coma, Colon, bracket, ... location
	var loc int = bytes.IndexByte(d.Buf, '"')
	slice = make([]byte, base64.RawStdEncoding.DecodedLen(len(d.Buf[:loc])))
	var n int
	var goErr error
	n, goErr = base64.RawStdEncoding.Decode(slice, d.Buf[:loc])
	if goErr != nil {
		err = ErrEncodedSliceCorrupted
		return
	}
	slice = slice[:n]

	d.Offset(loc + 1)
	return
}

// Decode32ByteArraySliceAsBase64 decode [32]byte base64 string slice. pass buf start from after [ and receive from after ]
func (d *DecoderMinifed) Decode32ByteArraySliceAsBase64() (slice [][32]byte, err protocol.Error) {
	d.Offset(1) // due to have [ at start

	const base64Len = 43 // base64.RawStdEncoding.EncodedLen(len(32))	>>	(32*8 + 5) / 6
	slice = make([][32]byte, 0, 8)

	var goErr error
	var array [32]byte
	for d.Buf[1] != ']' {
		d.Offset(2) // due to have `["` || `",`
		_, goErr = base64.RawStdEncoding.Decode(array[:], d.Buf[:base64Len])
		if goErr != nil {
			err = ErrEncodedSliceCorrupted
			return
		}
		slice = append(slice, array)
		d.Buf = d.Buf[base64Len:]
	}

	d.Offset(2) // due to have	`"]`
	return
}
