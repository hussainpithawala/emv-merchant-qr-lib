package emvqr

import (
	"fmt"
	"strconv"
	"strings"
)

// EncodeOptions controls optional encoder behaviour.
type EncodeOptions struct {
	// PayloadFormatIndicator overrides the default "01".
	PayloadFormatIndicator string
}

// Encode serialises a Payload into a raw EMV QR Code string, computing and
// appending the CRC automatically.
//
// Required fields: at least one MerchantAccountInfo, MerchantCategoryCode,
// TransactionCurrency, CountryCode, MerchantName, and MerchantCity.
func Encode(p *Payload) (string, error) {
	return EncodeWithOptions(p, EncodeOptions{})
}

// EncodeWithOptions serialises a Payload using the given options.
func EncodeWithOptions(p *Payload, opts EncodeOptions) (string, error) {
	if err := validatePayload(p); err != nil {
		return "", err
	}

	var sb strings.Builder

	// --- Payload Format Indicator (ID "00") --- always first
	pfi := p.PayloadFormatIndicator
	if pfi == "" {
		pfi = "01"
	}
	if opts.PayloadFormatIndicator != "" {
		pfi = opts.PayloadFormatIndicator
	}
	write(&sb, IDPayloadFormatIndicator, pfi)

	// --- Merchant Account Informations (IDs "02"–"51") ---
	for _, mai := range p.MerchantAccountInfos {
		chunk, err := encodeMerchantAccountInfo(mai)
		if err != nil {
			return "", fmt.Errorf("emvqr: encoding MAI %s: %w", mai.ID, err)
		}
		sb.WriteString(chunk)
	}

	// --- Merchant Category Code (ID "52") ---
	write(&sb, IDMerchantCategoryCode, p.MerchantCategoryCode)

	// --- Transaction Currency (ID "53") ---
	write(&sb, IDTransactionCurrency, p.TransactionCurrency)

	// --- Transaction Amount (ID "54") — optional ---
	if p.TransactionAmount != "" {
		write(&sb, IDTransactionAmount, p.TransactionAmount)
	}

	// --- Tip or Convenience Indicator (ID "55") — optional ---
	if p.TipOrConvenienceIndicator != "" {
		write(&sb, IDTipOrConvenienceIndicator, p.TipOrConvenienceIndicator)
		switch p.TipOrConvenienceIndicator {
		case TipIndicatorFixedConvenienceFee:
			if p.ValueConvenienceFeeFixed != "" {
				write(&sb, IDValueConvenienceFeeFixed, p.ValueConvenienceFeeFixed)
			}
		case TipIndicatorPercentageFee:
			if p.ValueConvenienceFeePercent != "" {
				write(&sb, IDValueConvenienceFeePercent, p.ValueConvenienceFeePercent)
			}
		}
	}

	// --- Country Code (ID "58") ---
	write(&sb, IDCountryCode, p.CountryCode)

	// --- Merchant Name (ID "59") ---
	write(&sb, IDMerchantName, p.MerchantName)

	// --- Merchant City (ID "60") ---
	write(&sb, IDMerchantCity, p.MerchantCity)

	// --- Postal Code (ID "61") — optional ---
	if p.PostalCode != "" {
		write(&sb, IDPostalCode, p.PostalCode)
	}

	// --- Additional Data Field Template (ID "62") — optional ---
	if p.AdditionalData != nil {
		chunk, err := encodeAdditionalDataField(p.AdditionalData)
		if err != nil {
			return "", fmt.Errorf("emvqr: encoding additional data field: %w", err)
		}
		sb.WriteString(chunk)
	}

	// --- Merchant Information Language Template (ID "64") — optional ---
	if p.LanguageTemplate != nil {
		chunk, err := encodeLanguageTemplate(p.LanguageTemplate)
		if err != nil {
			return "", fmt.Errorf("emvqr: encoding language template: %w", err)
		}
		sb.WriteString(chunk)
	}

	// --- Unreserved Templates (IDs "80"–"99") — optional ---
	for _, ut := range p.UnreservedTemplates {
		chunk, err := encodeUnreservedTemplate(ut)
		if err != nil {
			return "", fmt.Errorf("emvqr: encoding unreserved template %s: %w", ut.ID, err)
		}
		sb.WriteString(chunk)
	}

	// --- RFU fields ---
	for _, rfu := range p.RFUFields {
		write(&sb, rfu.ID, rfu.Value)
	}

	// --- CRC (ID "63") — computed last, always appended ---
	// The CRC covers everything up to and including the "6304" prefix.
	crcPrefix := sb.String() + "6304"
	crcVal := crc16CCITT([]byte(crcPrefix))
	sb.WriteString("6304")
	sb.WriteString(crcString(crcVal))

	return sb.String(), nil
}

// write appends a TLV-encoded field to the string builder.
// Panics on values > 99 chars (programming error; callers validate first).
func write(sb *strings.Builder, id, value string) {
	sb.WriteString(mustEncodeTLV(id, value))
}

// encodeMerchantAccountInfo encodes a single MAI entry.
func encodeMerchantAccountInfo(mai MerchantAccountInfo) (string, error) {
	if mai.IsTemplate() {
		// Build inner TLV from sub-fields
		var inner strings.Builder
		for _, sf := range mai.SubFields {
			chunk, err := encodeTLV(sf.ID, sf.Value)
			if err != nil {
				return "", err
			}
			inner.WriteString(chunk)
		}
		return encodeTLV(mai.ID, inner.String())
	}
	return encodeTLV(mai.ID, mai.Value)
}

// encodeAdditionalDataField encodes the Additional Data Field Template.
func encodeAdditionalDataField(adf *AdditionalDataField) (string, error) {
	var inner strings.Builder
	appendIf := func(id, val string) error {
		if val == "" {
			return nil
		}
		chunk, err := encodeTLV(id, val)
		if err != nil {
			return fmt.Errorf("field %s: %w", id, err)
		}
		inner.WriteString(chunk)
		return nil
	}
	for _, pair := range []struct{ id, val string }{
		{ADFBillNumber, adf.BillNumber},
		{ADFMobileNumber, adf.MobileNumber},
		{ADFStoreLabel, adf.StoreLabel},
		{ADFLoyaltyNumber, adf.LoyaltyNumber},
		{ADFReferenceLabel, adf.ReferenceLabel},
		{ADFCustomerLabel, adf.CustomerLabel},
		{ADFTerminalLabel, adf.TerminalLabel},
		{ADFPurposeOfTransaction, adf.PurposeOfTransaction},
		{ADFAdditionalConsumerDataRequest, adf.AdditionalConsumerDataRequest},
	} {
		if err := appendIf(pair.id, pair.val); err != nil {
			return "", err
		}
	}
	for _, rfu := range adf.RFUFields {
		if err := appendIf(rfu.ID, rfu.Value); err != nil {
			return "", err
		}
	}
	return encodeTLV(IDAdditionalDataFieldTemplate, inner.String())
}

// encodeLanguageTemplate encodes the Merchant Information Language Template.
func encodeLanguageTemplate(lt *LanguageTemplate) (string, error) {
	var inner strings.Builder
	if lt.LanguagePreference != "" {
		chunk, err := encodeTLV(LangPreference, lt.LanguagePreference)
		if err != nil {
			return "", err
		}
		inner.WriteString(chunk)
	}
	if lt.MerchantName != "" {
		chunk, err := encodeTLV(LangMerchantName, lt.MerchantName)
		if err != nil {
			return "", err
		}
		inner.WriteString(chunk)
	}
	if lt.MerchantCity != "" {
		chunk, err := encodeTLV(LangMerchantCity, lt.MerchantCity)
		if err != nil {
			return "", err
		}
		inner.WriteString(chunk)
	}
	for _, rfu := range lt.RFUFields {
		chunk, err := encodeTLV(rfu.ID, rfu.Value)
		if err != nil {
			return "", err
		}
		inner.WriteString(chunk)
	}
	return encodeTLV(IDMerchantInfoLanguageTemplate, inner.String())
}

// encodeUnreservedTemplate encodes an Unreserved Template.
func encodeUnreservedTemplate(ut UnreservedTemplate) (string, error) {
	n, err := strconv.Atoi(ut.ID)
	if err != nil || n < 80 || n > 99 {
		return "", fmt.Errorf("emvqr: unreserved template ID %q must be 80–99", ut.ID)
	}
	var inner strings.Builder
	if ut.GloballyUniqueID != "" {
		chunk, err := encodeTLV(MAIGloballyUniqueID, ut.GloballyUniqueID)
		if err != nil {
			return "", err
		}
		inner.WriteString(chunk)
	}
	for _, sf := range ut.SubFields {
		chunk, err := encodeTLV(sf.ID, sf.Value)
		if err != nil {
			return "", err
		}
		inner.WriteString(chunk)
	}
	return encodeTLV(ut.ID, inner.String())
}

// validatePayload ensures required fields are present.
func validatePayload(p *Payload) error {
	if p == nil {
		return fmt.Errorf("%w: nil payload", ErrMissingRequired)
	}
	required := []struct{ name, val string }{
		{"MerchantCategoryCode", p.MerchantCategoryCode},
		{"TransactionCurrency", p.TransactionCurrency},
		{"CountryCode", p.CountryCode},
		{"MerchantName", p.MerchantName},
		{"MerchantCity", p.MerchantCity},
	}
	for _, r := range required {
		if r.val == "" {
			return fmt.Errorf("%w: %s", ErrMissingRequired, r.name)
		}
	}
	if len(p.MerchantAccountInfos) == 0 {
		return fmt.Errorf("%w: at least one MerchantAccountInfo is required", ErrMissingRequired)
	}
	// Validate tip/fee consistency
	switch p.TipOrConvenienceIndicator {
	case "", TipIndicatorPromptConsumer, TipIndicatorFixedConvenienceFee, TipIndicatorPercentageFee:
		// valid
	default:
		return fmt.Errorf("emvqr: invalid TipOrConvenienceIndicator %q (must be 01, 02, or 03)", p.TipOrConvenienceIndicator)
	}
	return nil
}
