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
//	    MerchantIdentifiers: []emvqr.MerchantIdentifier{{ID: "02", Value: "4000123456789012"}},
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
)

// -------------------------------------------------------------------------
// Field IDs (as defined in EMV QRCPS Merchant-Presented Mode v1.0)
// -------------------------------------------------------------------------

const (
	IDPayloadFormatIndicator  = "00" // Payload Format Indicator
	IDPointOfInitiationMethod = "01" // Point of Initiation Method (Bharat QR)

	IDMerchantCategoryCode       = "52" // Merchant Category Code (ISO 18245)
	IDTransactionCurrency        = "53" // Transaction Currency (ISO 4217)
	IDTransactionAmount          = "54" // Transaction Amount (optional)
	IDTipOrConvenienceIndicator  = "55" // Tip or Convenience Fee Indicator
	IDValueConvenienceFeeFixed   = "56" // Convenience Fee Fixed Amount
	IDValueConvenienceFeePercent = "57" // Convenience Fee Percentage

	IDCountryCode                  = "58" // Country Code (ISO 3166-1)
	IDMerchantName                 = "59" // Merchant Name
	IDMerchantCity                 = "60" // Merchant City
	IDPostalCode                   = "61" // Postal Code
	IDAdditionalDataFieldTemplate  = "62" // Additional Data Field Template
	IDUPIVPATemplate               = "26" // UPI VPA Template (Bharat QR)
	IDUPIVPAReference              = "27" // UPI VPA Reference (Bharat QR dynamic)
	IDAadhaarTemplate              = "28" // Aadhaar Number Template (Bharat QR)
	IDCRC                          = "63" // CRC16-CCITT Checksum
	IDMerchantInfoLanguageTemplate = "64" // Merchant Information Language Template
)

// Tip or Convenience Indicator values
const (
	TipIndicatorPromptConsumer      = "01" // consumer prompted to enter tip
	TipIndicatorFixedConvenienceFee = "02" // fixed convenience fee
	TipIndicatorPercentageFee       = "03" // percentage-based convenience fee
)

// Payload Format Indicator value (EMV QRCPS version)
const (
	PayloadFormatIndicatorValue = "01" // Current EMV QRCPS version
)

// Point of Initiation Method (Tag 01) values - Format: <Method><DataType>
const (
	// POI Method component values
	POIMethodQR  = "1" // QR code based initiation
	POIMethodBLE = "2" // Bluetooth Low Energy initiation
	POIMethodNFC = "3" // Near Field Communication initiation

	// POI Data type component values
	POIDataTypeStatic  = "1" // Static QR code (reusable)
	POIDataTypeDynamic = "2" // Dynamic QR code (per-transaction)

	// Common POI combined values
	POIStaticQR   = "11" // Static QR code
	POIDynamicQR  = "12" // Dynamic QR code
	POIStaticBLE  = "21" // Static BLE
	POIDynamicBLE = "22" // Dynamic BLE
	POIStaticNFC  = "31" // Static NFC
	POIDynamicNFC = "32" // Dynamic NFC
)

// RuPay Application Provider Identifier (AID) constants
const (
	RuPayRIDValue       = "A000000524"       // RuPay Registered Application Provider Identifier
	NPCIUPIAIDIndicator = "A000000677010111" // NPCI/RuPay UPI Indicator
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

// Sub-field IDs for UPI VPA Reference (ID "27")
const (
	UPIVPARefRuPayRID       = "00"
	UPIVPARefTransactionRef = "01"
	UPIVPARefURL            = "02"
)

// Sub-field IDs for Aadhaar Template (ID "28")
const (
	AadhaarRuPayRID   = "00"
	AadhaarAadhaarNum = "01"
)

// PromptValue is the sentinel value used in Additional Data Fields to signal
// that the consumer QR application should prompt the consumer for input.
const PromptValue = "***"

// -------------------------------------------------------------------------
// Data structures
// -------------------------------------------------------------------------

// MerchantIdentifier represents a merchant account information identifier (IDs "02"–"51").
// Per EMV QRCPS v1.0, at least one merchant identifier is mandatory, and each tag ID can appear at most once.
// Primitive identifiers (IDs 02-25) hold payment network account values; template identifiers (IDs 26-51)
// contain nested TLV structures (SubFields) for advanced networks like UPI VPA (26), UPI VPA Reference (27), and Aadhaar (28).
type MerchantIdentifier struct {
	ID        string       // Two-digit field identifier (02-51)
	Value     string       // Account value for primitives (IDs 02-25); empty for templates (IDs 26-51)
	SubFields []DataObject // Nested TLV data for template entries (IDs 26-51); nil for primitives
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

// UPIVPATemplate holds the parsed contents of the UPI VPA template
// (ID "26"), used for merchant VPA in Bharat QRs (both static and dynamic).
type UPIVPATemplate struct {
	RuPayRID      string // Sub-tag 00: "A000000524" (RuPay RID, mandatory)
	VPA           string // Sub-tag 01: Merchant's UPI VPA (e.g., "merchant@bank")
	MinimumAmount string // Sub-tag 02: Minimum amount for dynamic QRs (optional)
}

// UPIVPAReference holds the parsed contents of the UPI VPA Reference template
// (ID "27"), used for dynamic Bharat QRs with transaction-specific references.
type UPIVPAReference struct {
	RuPayRID       string // Sub-tag 00: "A000000524"
	TransactionRef string // Sub-tag 01: min 4, max 35 digits/alphanumeric (order number, booking ID, bill ID, etc.)
	ReferenceURL   string // Sub-tag 02: optional, max 26 chars
}

// AadhaarInfo holds the parsed contents of the Aadhaar Number Template
// (ID "28"), used for Aadhaar-linked Bharat QRs.
type AadhaarInfo struct {
	RuPayRID      string // Sub-tag 00: "A000000524"
	AadhaarNumber string // Sub-tag 01: 12 digits
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

	// PointOfInitiationMethod (ID "01", Bharat QR) indicates how the QR was initiated.
	// Format: "XY" where X is method (1=QR, 2=BLE, 3=NFC), Y is data type (1=static, 2=dynamic).
	PointOfInitiationMethod string

	// MerchantIdentifiers contains merchant identifiers for supported payment networks (IDs "02"–"25").
	// Per EMV QRCPS spec, at least one merchant identifier is mandatory.
	// Multiple identifiers allowed (e.g., both Visa and Mastercard, or RuPay and Bank Account).
	// Each tag ID (02-25) can appear at most once.
	MerchantIdentifiers []MerchantIdentifier

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

	// UPIVPAInfo (ID "26", Bharat QR) holds merchant UPI Virtual Payment Address information.
	// Extracted from MerchantIdentifiers[tag="26"] for convenient typed access.
	UPIVPAInfo *UPIVPATemplate

	// UPITransactionRef (ID "27", Bharat QR) holds transaction-specific reference for dynamic UPI QRs.
	// Used to link QR to order number, booking ID, bill ID, etc.
	// Extracted from MerchantIdentifiers[tag="27"] for convenient typed access.
	UPITransactionRef *UPIVPAReference

	// MerchantAadhaar (ID "28", Bharat QR) holds Aadhaar-linked merchant authentication information.
	// Extracted from MerchantIdentifiers[tag="28"] for convenient typed access.
	MerchantAadhaar *AadhaarInfo

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
