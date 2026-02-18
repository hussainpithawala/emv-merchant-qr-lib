// Package emvqr implements encoding and decoding of EMV QR Code payloads
// as defined in the EMV QR Code Specification for Payment Systems (EMV QRCPS)
// Merchant-Presented Mode v1.0.
//
// The payload format uses a TLV (Tag-Length-Value) structure where each data
// object is encoded as:
//
//	ID (2 digits) + Length (2 digits) + Value (variable alphanumeric string)
//
// Example usage:
//
//	// Decode
//	payload, err := emvqr.Decode("000201...")
//
//	// Encode
//	p := &emvqr.Payload{
//	    MerchantAccountInfo: []emvqr.MerchantAccountInfo{{ID: "02", Value: "4000123456789012"}},
//	    MerchantCategoryCode: "5251",
//	    TransactionCurrency:  "840",
//	    CountryCode:          "US",
//	    MerchantName:         "ABC Hammers",
//	    MerchantCity:         "New York",
//	}
//	s, err := emvqr.Encode(p)
package emvqr

import (
	"errors"
	"fmt"
	"strconv"
)

// -------------------------------------------------------------------------
// Field IDs (as defined in EMV QRCPS Merchant-Presented Mode v1.0)
// -------------------------------------------------------------------------

const (

	// IDPayloadFormatIndicator represents the identifier for the payload format indicator, commonly set as "00".
	IDPayloadFormatIndicator = "00"

	// IDMerchantAccountInfoMin IDMerchantAccountInfoRangeStart–End covers primitive MAI (IDs 02–25)
	// and template MAI (IDs 26–51). ID "01" is RFU.
	// IDMerchantAccountInfoMin          = "02"
	// IDMerchantAccountInfoPrimitiveMax = "25"
	// IDMerchantAccountInfoTemplateMax  = "51"

	// IDMerchantCategoryCode represents the identifier for the merchant category code, defined as "52".
	IDMerchantCategoryCode = "52"
	// IDTransactionCurrency represents the identifier for the transaction currency, defined as "53".
	IDTransactionCurrency = "53"
	// IDTransactionAmount represents the identifier for the transaction amount, defined as "54".
	IDTransactionAmount = "54"

	// IDTipOrConvenienceIndicator Tip or Convenience Indicator values
	IDTipOrConvenienceIndicator = "55"

	// IDValueConvenienceFeeFixed and IDValueConvenienceFeePercent represent the values for fixed and percentage-based convenience fees, respectively.
	IDValueConvenienceFeeFixed = "56"

	// IDValueConvenienceFeePercent represents the identifier for the percentage-based convenience fee, defined as "57".
	IDValueConvenienceFeePercent = "57"

	// IDCountryCode represents the identifier for the country code, defined as "58".
	IDCountryCode = "58"

	// IDMerchantName specifies the identifier for the merchant's name in the EMV QR code data structure.
	IDMerchantName = "59"

	// IDMerchantCity specifies the identifier for the merchant's city in the EMV QR code data structure.'
	IDMerchantCity = "60"

	// IDPostalCode specifies the identifier for the merchant's postal code in the EMV QR code data structure.'
	IDPostalCode = "61"

	// IDAdditionalDataFieldTemplate specifies the identifier for the additional data field template in the EMV QR code data structure.
	IDAdditionalDataFieldTemplate = "62"

	// IDCRC represents the constant identifier for the CRC (Cyclic Redundancy Check) field in the payload.
	IDCRC = "63"

	// IDMerchantInfoLanguageTemplate specifies the identifier for the merchant information language template in the EMV QR code data structure.
	IDMerchantInfoLanguageTemplate = "64"

	// Unreserved Template range
	// IDUnreservedTemplateMin = "80"
	// IDUnreservedTemplateMax = "99"
)

// Tip or Convenience Indicator values
const (
	TipIndicatorPromptConsumer      = "01" // consumer prompted to enter tip
	TipIndicatorFixedConvenienceFee = "02" // fixed convenience fee
	TipIndicatorPercentageFee       = "03" // percentage-based convenience fee
)

// Sub-field IDs for Additional Data Field Template (ID "62")
const (
	ADFBillNumber                    = "01"
	ADFMobileNumber                  = "02"
	ADFStoreLabel                    = "03"
	ADFLoyaltyNumber                 = "04"
	ADFReferenceLabel                = "05"
	ADFCustomerLabel                 = "06"
	ADFTerminalLabel                 = "07"
	ADFPurposeOfTransaction          = "08"
	ADFAdditionalConsumerDataRequest = "09"
)

// Subfield IDs for Merchant Information – Language Template (ID "64")
const (
	LangPreference   = "00"
	LangMerchantName = "01"
	LangMerchantCity = "02"
)

// MAIGloballyUniqueID Sub-field IDs for Merchant Account Information templates (IDs "26"–"51")
const (
	MAIGloballyUniqueID = "00"
)

// PromptValue is the sentinel value used in Additional Data Fields to signal
// that the consumer QR application should prompt the consumer for input.
const PromptValue = "***"

// -------------------------------------------------------------------------
// Data structures
// -------------------------------------------------------------------------

// MerchantAccountInfo represents a Merchant Account Information field.
// For primitive entries (IDs "02"–"25") SubFields is nil and Value holds
// the raw account string. For template entries (IDs "26"–"51") SubFields
// holds the nested TLV data objects and Value is empty.
type MerchantAccountInfo struct {
	// ID is the two-digit field identifier, e.g. "02" or "26".
	ID string

	// Value is the raw value for primitive entries.
	Value string

	// SubFields holds decoded sub-data-objects for template entries.
	SubFields []DataObject
}

// GloballyUniqueID returns the Globally Unique ID sub-field value for
// template Merchant Account Information objects, or an empty string if absent.
func (m *MerchantAccountInfo) GloballyUniqueID() string {
	for _, f := range m.SubFields {
		if f.ID == MAIGloballyUniqueID {
			return f.Value
		}
	}
	return ""
}

// SubField returns the value of a sub-field by ID, or empty string if absent.
func (m *MerchantAccountInfo) SubField(id string) string {
	for _, f := range m.SubFields {
		if f.ID == id {
			return f.Value
		}
	}
	return ""
}

// IsTemplate reports whether this MAI uses template encoding (IDs "26"–"51").
func (m *MerchantAccountInfo) IsTemplate() bool {
	n, err := strconv.Atoi(m.ID)
	if err != nil {
		return false
	}
	return n >= 26 && n <= 51
}

// AdditionalDataField holds the parsed contents of the Additional Data Field
// Template (ID "62").
type AdditionalDataField struct {
	BillNumber                    string
	MobileNumber                  string
	StoreLabel                    string
	LoyaltyNumber                 string
	ReferenceLabel                string
	CustomerLabel                 string
	TerminalLabel                 string
	PurposeOfTransaction          string
	AdditionalConsumerDataRequest string

	// RFUFields holds any unrecognised sub-fields for forward compatibility.
	RFUFields []DataObject
}

// LanguageTemplate holds the parsed contents of the Merchant Information –
// Language Template (ID "64").
type LanguageTemplate struct {
	LanguagePreference string // e.g. "es", "zh"
	MerchantName       string
	MerchantCity       string

	// RFUFields holds any unrecognised sub-fields.
	RFUFields []DataObject
}

// UnreservedTemplate holds the parsed contents of an Unreserved Template
// (IDs "80"–"99").
type UnreservedTemplate struct {
	ID               string
	GloballyUniqueID string
	SubFields        []DataObject
}

// DataObject is a generic TLV data object used for unknown or nested fields.
type DataObject struct {
	ID    string
	Value string
}

// Payload is the top-level decoded EMV QR Code payload.
type Payload struct {
	// PayloadFormatIndicator is always "01" for the current version.
	PayloadFormatIndicator string

	// MerchantAccountInfos contains all MAI entries (IDs "02"–"51").
	MerchantAccountInfos []MerchantAccountInfo

	MerchantCategoryCode string
	TransactionCurrency  string // ISO 4217 numeric code, e.g. "840"
	TransactionAmount    string // omitted if consumer enters amount at POS

	// TipOrConvenienceIndicator: "", "01", "02", or "03"
	TipOrConvenienceIndicator  string
	ValueConvenienceFeeFixed   string
	ValueConvenienceFeePercent string

	CountryCode  string
	MerchantName string
	MerchantCity string
	PostalCode   string

	AdditionalData   *AdditionalDataField
	LanguageTemplate *LanguageTemplate

	UnreservedTemplates []UnreservedTemplate

	// CRC is the four-character CRC16-CCITT hex value (upper-case).
	CRC string

	// RFUFields holds any unrecognised top-level fields.
	RFUFields []DataObject
}

// -------------------------------------------------------------------------
// Errors
// -------------------------------------------------------------------------

var (

	// ErrInvalidLength is returned when the provided data length is insufficient for processing.
	ErrInvalidLength = errors.New("emvqr: data length is too short")
	// ErrInvalidTLV is returned when the provided TLV structure is invalid.
	ErrInvalidTLV = errors.New("emvqr: malformed TLV structure")
	// ErrCRCMismatch is returned when the provided CRC does not match the calculated value.
	ErrCRCMismatch = errors.New("emvqr: CRC mismatch")
	// ErrMissingRequired is returned when a required field is missing.
	ErrMissingRequired = errors.New("emvqr: missing required field")
)

// ParseError is returned when a specific field cannot be parsed.
type ParseError struct {
	ID  string
	Err error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("emvqr: error parsing field %s: %v", e.ID, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }
