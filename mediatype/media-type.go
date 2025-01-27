/* For license and copyright information please see the LEGAL file in the code repository */

package mediatype

import (
	"strings"

	"github.com/GeniusesGroup/libgo/protocol"
	uuid "github.com/GeniusesGroup/libgo/uuid/32byte"
)

// MT is the same as the MediaType.
// Use this type when embed in other struct to solve field & method same name problem(MediaType struct and MediaType() method) to satisfy interfaces.
type MT = MediaType

// MediaType implement protocol.MediaType interface
// type "/" [tree "."] subtype ["+" suffix]* [";" parameter]
// https://datatracker.ietf.org/doc/html/rfc2046
type MediaType struct {
	uuid.Generated

	mediaType  string
	mainType   string
	tree       string
	subType    string
	suffix     string
	parameters []string
}

func (mt *MediaType) Init(mediatype string) {
	mt.mediaType = mediatype
	mt.parse()

	mt.Generated.NewHashString(mediatype)
}

//libgo:impl protocol.MediaType
func (mt *MediaType) MediaType() string                   { return mt.mediaType }
func (mt *MediaType) MainType() string                    { return mt.mainType }
func (mt *MediaType) Tree() string                        { return mt.tree }
func (mt *MediaType) SubType() string                     { return mt.subType }
func (mt *MediaType) Suffix() string                      { return mt.suffix }
func (mt *MediaType) Parameters() []string                { return mt.parameters }
func (mt *MediaType) FileExtension() string               { return "" }
func (mt *MediaType) Status() protocol.SoftwareStatus     { return protocol.Software_Unset }
func (mt *MediaType) ReferenceURI() string                { return "" }
func (mt *MediaType) IssueDate() protocol.Time            { return nil }
func (mt *MediaType) ExpiryDate() protocol.Time           { return nil }
func (mt *MediaType) ExpireInFavorOf() protocol.MediaType { return nil }
func (mt *MediaType) Fields() []protocol.Field            { return nil }
func (mt *MediaType) ToString() string                    { return mt.mediaType }

// TODO::: complete extraction
func (mt *MediaType) parse() {
	var mediatype = mt.mediaType

	var i = strings.IndexByte(mediatype, '/')
	if i < 0 {
		panic("Mediatype isn't in good shape to parse it. Please check it.")
	}
	mt.mainType = mediatype[:i]
	// TODO:::
	mt.subType = mediatype[i+1:]
}
