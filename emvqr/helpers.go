package emvqr

import (
	"fmt"
	"strconv"
)

// -------------------------------------------------------------------------
// Convenience constructors
// -------------------------------------------------------------------------

// NewPayload returns a Payload pre-populated with sensible defaults.
// PayloadFormatIndicator is set to "01".
func NewPayload() *Payload {
	return &Payload{
		PayloadFormatIndicator: "01",
	}
}

// AddMerchantIdentifier adds a merchant identifier for a payment network (IDs "02"–"25").
// Multiple networks can be supported (e.g., Visa, Mastercard, RuPay, Bank Account).
// Each tag ID can appear at most once per QR code.
func (p *Payload) AddMerchantIdentifier(tagID, value string) error {
	if tagID == "" {
		return fmt.Errorf("emvqr: merchant identifier tag ID cannot be empty")
	}
	n, err := strconv.Atoi(tagID)
	if err != nil || n < 2 || n > 25 {
		return fmt.Errorf("emvqr: merchant identifier tag ID must be 02–25, got %q", tagID)
	}
	// Check for duplicate tag ID
	for _, mi := range p.MerchantIdentifiers {
		if mi.ID == tagID {
			return fmt.Errorf("emvqr: merchant identifier tag ID %s already exists", tagID)
		}
	}
	p.MerchantIdentifiers = append(p.MerchantIdentifiers, MerchantIdentifier{
		ID:    tagID,
		Value: value,
	})
	return nil
}

// nextTemplateMAIID returns the next available template MAI ID ("26"–"51").
func (p *Payload) nextTemplateMAIID() string {
	used := map[int]bool{}
	for _, m := range p.MerchantIdentifiers {
		if n, err := strconv.Atoi(m.ID); err == nil {
			used[n] = true
		}
	}
	for i := 26; i <= 51; i++ {
		if !used[i] {
			return fmt.Sprintf("%02d", i)
		}
	}
	return ""
}

// SetFixedConvenienceFee configures the payload for a fixed convenience fee.
// amount should be the fee value as a string, e.g. "10.75".
func (p *Payload) SetFixedConvenienceFee(amount string) {
	p.TipOrConvenienceIndicator = TipIndicatorFixedConvenienceFee
	p.ValueConvenienceFeeFixed = amount
	p.ValueConvenienceFeePercent = ""
}

// SetPercentageConvenienceFee configures the payload for a percentage-based
// convenience fee. percent should be a string like "3.00" (meaning 3%).
func (p *Payload) SetPercentageConvenienceFee(percent string) {
	p.TipOrConvenienceIndicator = TipIndicatorPercentageFee
	p.ValueConvenienceFeePercent = percent
	p.ValueConvenienceFeeFixed = ""
}

// SetPromptForTip configures the payload to prompt the consumer for a tip.
func (p *Payload) SetPromptForTip() {
	p.TipOrConvenienceIndicator = TipIndicatorPromptConsumer
	p.ValueConvenienceFeeFixed = ""
	p.ValueConvenienceFeePercent = ""
}

// SetAdditionalData is a convenience method for setting individual Additional
// Data Field values without constructing the struct manually.
func (p *Payload) SetAdditionalData(fn func(*AdditionalDataField)) {
	if p.AdditionalData == nil {
		p.AdditionalData = &AdditionalDataField{}
	}
	fn(p.AdditionalData)
}

// SetLanguageTemplate sets the Merchant Information Language Template.
func (p *Payload) SetLanguageTemplate(lang, name, city string) {
	p.LanguageTemplate = &LanguageTemplate{
		LanguagePreference: lang,
		MerchantName:       name,
		MerchantCity:       city,
	}
}

// -------------------------------------------------------------------------
// Computed helpers on Payload
// -------------------------------------------------------------------------

// TotalAmount returns the total amount including any fixed or percentage-based
// convenience fee, given the base transaction amount as a float64.
// Returns 0 and an error if the TransactionAmount field cannot be parsed.
func (p *Payload) TotalAmount() (float64, error) {
	if p.TransactionAmount == "" {
		return 0, fmt.Errorf("emvqr: TransactionAmount not present in payload")
	}
	base, err := strconv.ParseFloat(p.TransactionAmount, 64)
	if err != nil {
		return 0, fmt.Errorf("emvqr: invalid TransactionAmount %q: %w", p.TransactionAmount, err)
	}
	switch p.TipOrConvenienceIndicator {
	case TipIndicatorFixedConvenienceFee:
		if p.ValueConvenienceFeeFixed == "" {
			return base, nil
		}
		fee, err := strconv.ParseFloat(p.ValueConvenienceFeeFixed, 64)
		if err != nil {
			return 0, fmt.Errorf("emvqr: invalid ValueConvenienceFeeFixed %q: %w", p.ValueConvenienceFeeFixed, err)
		}
		return base + fee, nil

	case TipIndicatorPercentageFee:
		if p.ValueConvenienceFeePercent == "" {
			return base, nil
		}
		pct, err := strconv.ParseFloat(p.ValueConvenienceFeePercent, 64)
		if err != nil {
			return 0, fmt.Errorf("emvqr: invalid ValueConvenienceFeePercent %q: %w", p.ValueConvenienceFeePercent, err)
		}
		return base + base*(pct/100), nil
	}
	return base, nil
}

// LoyaltyNumberRequired reports whether the consumer QR application should
// prompt the consumer to enter a loyalty number.
func (p *Payload) LoyaltyNumberRequired() bool {
	return p.AdditionalData != nil && p.AdditionalData.LoyaltyNumber == PromptValue
}

// MobileNumberRequired reports whether the consumer QR application should
// prompt the consumer to enter a mobile number.
func (p *Payload) MobileNumberRequired() bool {
	return p.AdditionalData != nil && p.AdditionalData.MobileNumber == PromptValue
}

// PreferredMerchantName returns the merchant name in the given BCP-47 language
// tag (e.g. "es"). Falls back to the primary MerchantName field if no
// alternate language template is present or if the language does not match.
func (p *Payload) PreferredMerchantName(lang string) string {
	if p.LanguageTemplate != nil && p.LanguageTemplate.MerchantName != "" {
		if p.LanguageTemplate.LanguagePreference == lang {
			return p.LanguageTemplate.MerchantName
		}
	}
	return p.MerchantName
}

// PreferredMerchantCity returns the merchant city in the given language,
// falling back to the primary MerchantCity field.
func (p *Payload) PreferredMerchantCity(lang string) string {
	if p.LanguageTemplate != nil && p.LanguageTemplate.MerchantCity != "" {
		if p.LanguageTemplate.LanguagePreference == lang {
			return p.LanguageTemplate.MerchantCity
		}
	}
	return p.MerchantCity
}

// HasMultipleNetworks reports whether the payload contains multiple payment networks.
// Per EMV QRCPS spec, merchant identifiers include:
//   - Primitive (IDs 02-25): Visa, Mastercard, RuPay, Bank Account, AmEx, etc. (multiple allowed)
//   - Template (IDs 26-51): UPI VPA (Tag 26), and future network templates
//
// Multiple networks = more than one merchant identifier in the slice.
func (p *Payload) HasMultipleNetworks() bool {
	return len(p.MerchantIdentifiers) > 1
}

// -------------------------------------------------------------------------
// Bharat QR specific helpers (Tags 01, 27, 28)
// -------------------------------------------------------------------------

// SetPointOfInitiationMethod sets the Point of Initiation Method (Tag 01, Bharat QR).
// method: "1"=QR, "2"=BLE, "3"=NFC
// dataType: "1"=static, "2"=dynamic
// Combined as "XY", e.g. "11" for static QR, "12" for dynamic QR.
func (p *Payload) SetPointOfInitiationMethod(method, dataType string) error {
	if method == "" || dataType == "" {
		return fmt.Errorf("emvqr: method and dataType must not be empty")
	}
	if method != "1" && method != "2" && method != "3" {
		return fmt.Errorf("emvqr: method must be 1 (QR), 2 (BLE), or 3 (NFC), got %q", method)
	}
	if dataType != "1" && dataType != "2" {
		return fmt.Errorf("emvqr: dataType must be 1 (static) or 2 (dynamic), got %q", dataType)
	}
	p.PointOfInitiationMethod = method + dataType
	return nil
}

// SetUPIVPATemplate sets the UPI VPA Template (Tag 26, Bharat QR).
// ruPayRID: typically "A000000524" (RuPay RID)
// vpa: merchant's UPI VPA address (e.g., "merchant@bank")
// minimumAmount: optional minimum amount for dynamic QRs
func (p *Payload) SetUPIVPATemplate(ruPayRID, vpa, minimumAmount string) error {
	if vpa == "" {
		return fmt.Errorf("emvqr: VPA must not be empty")
	}
	p.UPIVPAInfo = &UPIVPATemplate{
		RuPayRID:      ruPayRID,
		VPA:           vpa,
		MinimumAmount: minimumAmount,
	}
	return nil
}

// SetUPIVPAReference sets the UPI VPA Reference template (Tag 27, Bharat QR dynamic).
// transactionRef: 4-35 character transaction reference (order number, booking ID, bill ID, etc.)
// url: optional reference URL (max 26 chars)
func (p *Payload) SetUPIVPAReference(transactionRef, url string) error {
	if transactionRef == "" {
		return fmt.Errorf("emvqr: transaction reference must not be empty")
	}
	if len(transactionRef) < 4 || len(transactionRef) > 35 {
		return fmt.Errorf("emvqr: transaction reference must be 4-35 characters, got %d", len(transactionRef))
	}
	if url != "" && len(url) > 26 {
		return fmt.Errorf("emvqr: reference URL must be max 26 characters, got %d", len(url))
	}
	p.UPITransactionRef = &UPIVPAReference{
		RuPayRID:       RuPayRIDValue,
		TransactionRef: transactionRef,
		ReferenceURL:   url,
	}
	return nil
}

// SetAadhaarNumber sets the Aadhaar template (Tag 28, Bharat QR).
// aadhaarNum: 12-digit Aadhaar number.
func (p *Payload) SetAadhaarNumber(aadhaarNum string) error {
	if aadhaarNum == "" {
		return fmt.Errorf("emvqr: Aadhaar number must not be empty")
	}
	if len(aadhaarNum) != 12 {
		return fmt.Errorf("emvqr: Aadhaar number must be exactly 12 digits, got %d", len(aadhaarNum))
	}
	// Validate that it contains only digits
	for _, ch := range aadhaarNum {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("emvqr: Aadhaar number must contain only digits, got %q", aadhaarNum)
		}
	}
	p.MerchantAadhaar = &AadhaarInfo{
		RuPayRID:      RuPayRIDValue,
		AadhaarNumber: aadhaarNum,
	}
	return nil
}

// GetMerchantVPA returns the merchant VPA from Tag 26 (UPI VPA Template),
// or an empty string if not present.
func (p *Payload) GetMerchantVPA() string {
	if p.UPIVPAInfo != nil {
		return p.UPIVPAInfo.VPA
	}
	return ""
}

// GetMinimumAmount returns the minimum amount from Tag 26 (UPI VPA Template),
// or an empty string if not present.
func (p *Payload) GetMinimumAmount() string {
	if p.UPIVPAInfo != nil {
		return p.UPIVPAInfo.MinimumAmount
	}
	return ""
}

// GetTransactionReference returns the transaction reference from Tag 27 (UPI VPA Reference),
// or an empty string if not present.
func (p *Payload) GetTransactionReference() string {
	if p.UPITransactionRef != nil {
		return p.UPITransactionRef.TransactionRef
	}
	return ""
}

// GetAadhaarNumber returns the Aadhaar number from Tag 28,
// or an empty string if not present.
func (p *Payload) GetAadhaarNumber() string {
	if p.MerchantAadhaar != nil {
		return p.MerchantAadhaar.AadhaarNumber
	}
	return ""
}
