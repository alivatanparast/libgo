/* For license and copyright information please see LEGAL file in repository */

package uuid

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	"golang.org/x/crypto/sha3"

	"github.com/GeniusesGroup/libgo/binary"
	"github.com/GeniusesGroup/libgo/protocol"
	"github.com/GeniusesGroup/libgo/time/unix"
)

type UUID [8]byte

func (id UUID) UUID() [8]byte      { return id }
func (id UUID) ID() [3]byte        { return id.id() }
func (id UUID) IDasString() string { return base64.RawURLEncoding.EncodeToString(id[:8]) }
func (id UUID) ToString() string   { return "TODO:::" }
func (id UUID) ExistenceTime() protocol.Time {
	var time unix.Time
	time.ChangeTo(unix.SecElapsed(id.secondElapsed()), 0)
	return &time
}

// New will generate 8 byte time based UUID.
// **CAUTION**: Use for ObjectID in a clustered software without any hash cause all writes always go to one node.
// 99.999999% collision free on distribution generation.
func (id *UUID) New() {
	var err error
	_, err = io.ReadFull(rand.Reader, id[5:])
	if err != nil {
		// TODO::: make random by other ways
	}

	// Set time to UUID
	var now = unix.Now()
	id.setSecondElapsed(now.SecondElapsed())
}

// NewHash generate 8 byte incremental by time + hash of data UUID
// CAUTION::: Use for ObjectID in a clustered software cause all writes always go to one node!
// 99.999% collision free on distribution generation.
func (id *UUID) NewHash(data []byte) {
	var uuid32 = sha3.Sum256(data)
	copy(id[5:], uuid32[:])

	// Set time to UUID
	var now = unix.Now()
	id.setSecondElapsed(now.SecondElapsed())
}

// NewRandom generate 8 byte random UUID.
func (id *UUID) NewRandom() {
	var err error
	_, err = io.ReadFull(rand.Reader, id[:])
	if err != nil {
		// TODO::: make random by other ways
	}
}

func (id UUID) id() (rid [3]byte)    { copy(rid[:], id[5:]); return }
func (id UUID) secondElapsed() int64 { 
	var sec [8]byte
	copy(sec[:], id[:])
	return int64(binary.LittleEndian.Uint64(id[0:])) >> (64 - 40)
}
func (id *UUID) setSecondElapsed(sec int64) {
	binary.LittleEndian.PutUint64(id[0:], (uint64(sec) << (64 - 40)))
}
