package emvqr

import (
	"fmt"
	"strconv"
	"strings"
)

// DecodeOptions controls optional decoder behaviour.
type DecodeOptions struct {
	// SkipCRCValidation disables CRC checking. Useful when the CRC field is
	// absent (e.g., during unit tests with partial payloads).
	SkipCRCValidation bool
}

// Decode parses a raw EMV QR Code string and returns the structured Payload.
// The CRC is validated by default. Use DecodeWithOptions to customise this.
func Decode(raw string) (*Payload, error) {
	return DecodeWithOptions(raw, DecodeOptions{})
}

// DecodeWithOptions parses the raw string using the given options.
func DecodeWithOptions(raw string, opts DecodeOptions) (*Payload, error) {
	if len(raw) < 4 {
		return nil, ErrInvalidLength
	}

	// Validate and strip CRC before parsing
	if !opts.SkipCRCValidation {
		if err := validateCRC(raw); err != nil {
			return nil, err
		}
	}

	objects, err := parseTLV(raw)
	if err != nil {
		return nil, err
	}

	p := &Payload{}
	for _, obj := range objects {
		if err := p.applyObject(obj); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// validateCRC checks the CRC16-CCITT checksum embedded in the raw string.
// Per the spec, the CRC covers the entire payload including the "6304" prefix
// of the CRC field but not the 4-char CRC value itself.
func validateCRC(raw string) error {
	// Locate the CRC field: ID "63" + length "04" + 4-char value = 8 chars at end
	if len(raw) < 8 {
		return fmt.Errorf("%w: payload too short to contain CRC", ErrInvalidTLV)
	}
	crcFieldStart := strings.LastIndex(raw, "6304")
	if crcFieldStart == -1 {
		return fmt.Errorf("%w: CRC field (ID 63) not found", ErrInvalidTLV)
	}
	dataPart := raw[:crcFieldStart+4] // up to and including "6304"
	crcValue := raw[crcFieldStart+4 : crcFieldStart+8]

	computed := crc16CCITT([]byte(dataPart))
	expected := crcString(computed)
	if !strings.EqualFold(crcValue, expected) {
		return fmt.Errorf("%w: got %s, want %s", ErrCRCMismatch, strings.ToUpper(crcValue), expected)
	}
	return nil
}

// applyObject maps a single top-level TLV object onto the Payload.
func (p *Payload) applyObject(obj tlvObject) error {
	id := obj.id
	val := obj.value

	switch {
	case id == IDPayloadFormatIndicator:
		p.PayloadFormatIndicator = val

	case isMerchantAccountInfo(id):
		mai, err := decodeMerchantAccountInfo(id, val)
		if err != nil {
			return err
		}
		p.MerchantAccountInfos = append(p.MerchantAccountInfos, *mai)

	case id == IDMerchantCategoryCode:
		p.MerchantCategoryCode = val

	case id == IDTransactionCurrency:
		p.TransactionCurrency = val

	case id == IDTransactionAmount:
		p.TransactionAmount = val

	case id == IDTipOrConvenienceIndicator:
		p.TipOrConvenienceIndicator = val

	case id == IDValueConvenienceFeeFixed:
		p.ValueConvenienceFeeFixed = val

	case id == IDValueConvenienceFeePercent:
		p.ValueConvenienceFeePercent = val

	case id == IDCountryCode:
		p.CountryCode = val

	case id == IDMerchantName:
		p.MerchantName = val

	case id == IDMerchantCity:
		p.MerchantCity = val

	case id == IDPostalCode:
		p.PostalCode = val

	case id == IDAdditionalDataFieldTemplate:
		adf, err := decodeAdditionalDataField(val)
		if err != nil {
			return &ParseError{ID: id, Err: err}
		}
		p.AdditionalData = adf

	case id == IDCRC:
		p.CRC = strings.ToUpper(val)

	case id == IDMerchantInfoLanguageTemplate:
		lt, err := decodeLanguageTemplate(val)
		if err != nil {
			return &ParseError{ID: id, Err: err}
		}
		p.LanguageTemplate = lt

	case isUnreservedTemplate(id):
		ut, err := decodeUnreservedTemplate(id, val)
		if err != nil {
			return err
		}
		p.UnreservedTemplates = append(p.UnreservedTemplates, *ut)

	default:
		p.RFUFields = append(p.RFUFields, DataObject{ID: id, Value: val})
	}
	return nil
}

// isMerchantAccountInfo reports whether id falls in "02"–"51".
func isMerchantAccountInfo(id string) bool {
	n, err := strconv.Atoi(id)
	if err != nil {
		return false
	}
	return n >= 2 && n <= 51
}

// isUnreservedTemplate reports whether id falls in "80"–"99".
func isUnreservedTemplate(id string) bool {
	n, err := strconv.Atoi(id)
	if err != nil {
		return false
	}
	return n >= 80 && n <= 99
}

// decodeMerchantAccountInfo decodes either a primitive or template MAI entry.
func decodeMerchantAccountInfo(id, val string) (*MerchantAccountInfo, error) {
	n, _ := strconv.Atoi(id)
	mai := &MerchantAccountInfo{ID: id}
	if n >= 26 && n <= 51 {
		// Template: parse sub-fields
		subs, err := parseTLV(val)
		if err != nil {
			return nil, &ParseError{ID: id, Err: err}
		}
		for _, s := range subs {
			mai.SubFields = append(mai.SubFields, DataObject{ID: s.id, Value: s.value})
		}
	} else {
		mai.Value = val
	}
	return mai, nil
}

// decodeAdditionalDataField parses the contents of ID "62".
func decodeAdditionalDataField(val string) (*AdditionalDataField, error) {
	subs, err := parseTLV(val)
	if err != nil {
		return nil, err
	}
	adf := &AdditionalDataField{}
	for _, s := range subs {
		switch s.id {
		case ADFBillNumber:
			adf.BillNumber = s.value
		case ADFMobileNumber:
			adf.MobileNumber = s.value
		case ADFStoreLabel:
			adf.StoreLabel = s.value
		case ADFLoyaltyNumber:
			adf.LoyaltyNumber = s.value
		case ADFReferenceLabel:
			adf.ReferenceLabel = s.value
		case ADFCustomerLabel:
			adf.CustomerLabel = s.value
		case ADFTerminalLabel:
			adf.TerminalLabel = s.value
		case ADFPurposeOfTransaction:
			adf.PurposeOfTransaction = s.value
		case ADFAdditionalConsumerDataRequest:
			adf.AdditionalConsumerDataRequest = s.value
		default:
			adf.RFUFields = append(adf.RFUFields, DataObject{ID: s.id, Value: s.value})
		}
	}
	return adf, nil
}

// decodeLanguageTemplate parses the contents of ID "64".
func decodeLanguageTemplate(val string) (*LanguageTemplate, error) {
	subs, err := parseTLV(val)
	if err != nil {
		return nil, err
	}
	lt := &LanguageTemplate{}
	for _, s := range subs {
		switch s.id {
		case LangPreference:
			lt.LanguagePreference = s.value
		case LangMerchantName:
			lt.MerchantName = s.value
		case LangMerchantCity:
			lt.MerchantCity = s.value
		default:
			lt.RFUFields = append(lt.RFUFields, DataObject{ID: s.id, Value: s.value})
		}
	}
	return lt, nil
}

// decodeUnreservedTemplate parses an Unreserved Template (IDs "80"–"99").
func decodeUnreservedTemplate(id, val string) (*UnreservedTemplate, error) {
	subs, err := parseTLV(val)
	if err != nil {
		return nil, &ParseError{ID: id, Err: err}
	}
	ut := &UnreservedTemplate{ID: id}
	for _, s := range subs {
		if s.id == MAIGloballyUniqueID {
			ut.GloballyUniqueID = s.value
		} else {
			ut.SubFields = append(ut.SubFields, DataObject{ID: s.id, Value: s.value})
		}
	}
	return ut, nil
}
