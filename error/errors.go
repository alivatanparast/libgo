/* For license and copyright information please see the LEGAL file in the code repository */

package error

import (
	"github.com/GeniusesGroup/libgo/protocol"
)

// Errors store
type Errors struct {
	poolByID        map[protocol.MediaTypeID]protocol.Error
	poolByMediaType map[string]protocol.Error
}

func (e *Errors) Init() {
	e.poolByID = make(map[protocol.MediaTypeID]protocol.Error, 512)
	e.poolByMediaType = make(map[string]protocol.Error, 512)
}

func (e *Errors) RegisterError(err protocol.Error) {
	var errID = err.ID()

	if errID == 0 {
		// This condition will just be true in the dev phase.
		panic("Error must have valid ID to save it in platform errors pools. Initialize inner e.MediaType.Init() first if use libgo/service package.")
	}

	if protocol.AppMode_Dev && e.poolByID[errID] != nil {
		// This condition will just be true in the dev phase.
		panic("Error id exist and used for other Error. Check it now for bad media-type set or collision occurred" +
			"\nExiting error >> " + e.poolByID[errID].ToString() +
			"\nNew error >> " + err.ToString())
	}

	e.poolByID[errID] = err
	e.poolByMediaType[err.ToString()] = err
}

func (e *Errors) UnRegisterError(err protocol.Error) {
	delete(e.poolByID, err.ID())
	delete(e.poolByMediaType, err.ToString())
}

// GetErrorByID returns desire error if exist or ErrNotFound!
func (e *Errors) GetErrorByID(id protocol.MediaTypeID) (err protocol.Error) {
	if id == 0 {
		return
	}
	var ok bool
	err, ok = e.poolByID[id]
	if !ok {
		err = &ErrNotFound
	}
	return
}

// GetErrorByMediaType returns desire error if exist or ErrNotFound!
func (e *Errors) GetErrorByMediaType(mt string) (err protocol.Error) {
	var ok bool
	err, ok = e.poolByMediaType[mt]
	if !ok {
		err = &ErrNotFound
	}
	return
}
