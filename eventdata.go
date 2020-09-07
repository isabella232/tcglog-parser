package tcglog

import (
	"fmt"
	"io"
)

// EventData is an interface that represents all event data types that appear in a log. Most implementations of
// this are private to this module.
type EventData interface {
	String() string // Textual representation of the event data
	Bytes() []byte  // The raw event data bytes
}

type invalidEventData struct {
	data []byte
	err  error
}

func (e *invalidEventData) String() string {
	if e.err == io.ErrUnexpectedEOF {
		return "Invalid event data: event data smaller than expected"
	}
	return fmt.Sprintf("Invalid event data: %v", e.err)
}

func (e *invalidEventData) Bytes() []byte {
	return e.data
}

func (e *invalidEventData) Error() string {
	return e.err.Error()
}

func (e *invalidEventData) Unwrap() error {
	return e.err
}

type opaqueEventData struct {
	data []byte
}

func (e *opaqueEventData) String() string {
	return ""
}

func (e *opaqueEventData) Bytes() []byte {
	return e.data
}

func decodeEventDataImpl(pcrIndex PCRIndex, eventType EventType, data []byte, options *LogOptions,
	hasDigestOfSeparatorError bool) (EventData, int, error) {
	switch {
	case options.EnableGrub && (pcrIndex == 8 || pcrIndex == 9):
		if d, n := decodeEventDataGRUB(pcrIndex, eventType, data); d != nil {
			return d, n, nil
		}
		fallthrough
	case options.EnableSystemdEFIStub && pcrIndex == options.SystemdEFIStubPCR && eventType == EventTypeIPL:
		if d, n, e := decodeEventDataSystemdEFIStub(data); d != nil {
			return d, n, nil
		} else if e != nil {
			return nil, 0, e
		}
		fallthrough
	default:
		return decodeEventDataTCG(eventType, data, hasDigestOfSeparatorError)
	}
}

func decodeEventData(pcrIndex PCRIndex, eventType EventType, data []byte, options *LogOptions,
	hasDigestOfSeparatorError bool) (EventData, int) {
	event, trailingBytes, err :=
		decodeEventDataImpl(pcrIndex, eventType, data, options, hasDigestOfSeparatorError)

	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return &invalidEventData{data: data, err: err}, 0
	}

	if event != nil {
		return event, trailingBytes
	}

	return &opaqueEventData{data: data}, 0
}
