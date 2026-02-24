package emvqr

import (
	"fmt"
	"strconv"
)

// tlvObject is the raw result of parsing one TLV unit.
type tlvObject struct {
	id    string
	value string
}

// parseTLV splits a raw string into a sequence of TLV data objects.
// Each object is: 2-char ID + 2-char decimal length + <length> chars value.
func parseTLV(s string) ([]tlvObject, error) {
	var objects []tlvObject
	for len(s) > 0 {
		if len(s) < 4 {
			return nil, fmt.Errorf("%w: expected at least 4 chars, got %d", ErrInvalidTLV, len(s))
		}
		id := s[0:2]
		lenStr := s[2:4]
		length, err := strconv.Atoi(lenStr)
		if err != nil {
			return nil, fmt.Errorf("%w: non-numeric length %q for ID %s", ErrInvalidTLV, lenStr, id)
		}
		if len(s) < 4+length {
			return nil, fmt.Errorf("%w: declared length %d for ID %s exceeds remaining data (%d chars)", ErrInvalidTLV, length, id, len(s)-4)
		}
		value := s[4 : 4+length]
		objects = append(objects, tlvObject{id: id, value: value})
		s = s[4+length:]
	}
	return objects, nil
}

// encodeTLV encodes a single ID+value pair into TLV format.
// Returns an error if the value length exceeds 99 (the maximum representable
// in a 2-digit decimal length field).
func encodeTLV(id, value string) (string, error) {
	if len(value) > 99 {
		return "", fmt.Errorf("emvqr: value for ID %s is %d chars, exceeds maximum of 99", id, len(value))
	}
	return fmt.Sprintf("%s%02d%s", id, len(value), value), nil
}

// mustEncodeTLV is a helper that panics on encoding errors (for use with
// values that are already validated).
func mustEncodeTLV(id, value string) string {
	s, err := encodeTLV(id, value)
	if err != nil {
		panic(err)
	}
	return s
}

// convertTLVToDataObjects converts a slice of parsed tlvObject to DataObject.
// Used to populate SubFields in MerchantIdentifier and other template structures.
// Empty values are skipped as they are invalid per EMV QRCPS specification.
func convertTLVToDataObjects(objects []tlvObject) []DataObject {
	var result []DataObject
	for _, obj := range objects {
		// Skip empty values as they are not valid per spec
		if obj.value == "" {
			continue
		}
		result = append(result, DataObject{ID: obj.id, Value: obj.value})
	}
	return result
}
