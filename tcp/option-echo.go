/* For license and copyright information please see the LEGAL file in the code repository */

package tcp

import "github.com/GeniusesGroup/libgo/binary"

type optionEcho []byte

func (o optionEcho) Length() byte       { return o[0] }
func (o optionEcho) Echo() uint16       { return binary.BigEndian.Uint16(o[1:]) }
func (o optionEcho) NextOption() []byte { return o[5:] }

func (o optionEcho) Process(s *Socket) error {
	return nil
}
