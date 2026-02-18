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

// AddPrimitiveMerchantAccount adds a primitive Merchant Account Information
// entry (IDs "02"–"25"). ID "01" is reserved/RFU and is rejected.
//
// The first free ID starting from "02" is used if id is empty.
func (p *Payload) AddPrimitiveMerchantAccount(id, value string) error {
	if id == "" {
		id = p.nextPrimitiveMAIID()
	}
	n, err := strconv.Atoi(id)
	if err != nil || n < 2 || n > 25 {
		return fmt.Errorf("emvqr: primitive MAI ID must be 02–25, got %q", id)
	}
	p.MerchantAccountInfos = append(p.MerchantAccountInfos, MerchantAccountInfo{
		ID:    id,
		Value: value,
	})
	return nil
}

// AddTemplateMerchantAccount adds a template Merchant Account Information
// entry (IDs "26"–"51") with the given globally unique ID and optional
// additional sub-fields.
func (p *Payload) AddTemplateMerchantAccount(id, globallyUniqueID string, extra ...DataObject) error {
	if id == "" {
		id = p.nextTemplateMAIID()
	}
	n, err := strconv.Atoi(id)
	if err != nil || n < 26 || n > 51 {
		return fmt.Errorf("emvqr: template MAI ID must be 26–51, got %q", id)
	}
	subFields := make([]DataObject, 0, 1+len(extra))
	subFields = append(subFields, DataObject{ID: MAIGloballyUniqueID, Value: globallyUniqueID})
	for _, item := range extra {
		subFields = append(subFields, DataObject{ID: item.ID, Value: item.Value})
	}
	p.MerchantAccountInfos = append(p.MerchantAccountInfos, MerchantAccountInfo{
		ID:        id,
		SubFields: subFields,
	})
	return nil
}

// nextPrimitiveMAIID returns the next available primitive MAI ID ("02"–"25").
func (p *Payload) nextPrimitiveMAIID() string {
	used := map[int]bool{}
	for _, m := range p.MerchantAccountInfos {
		if n, err := strconv.Atoi(m.ID); err == nil {
			used[n] = true
		}
	}
	for i := 2; i <= 25; i++ {
		if !used[i] {
			return fmt.Sprintf("%02d", i)
		}
	}
	return ""
}

// nextTemplateMAIID returns the next available template MAI ID ("26"–"51").
func (p *Payload) nextTemplateMAIID() string {
	used := map[int]bool{}
	for _, m := range p.MerchantAccountInfos {
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

// HasMultipleNetworks reports whether the payload contains more than one
// Merchant Account Information entry, indicating multi-network support.
func (p *Payload) HasMultipleNetworks() bool {
	return len(p.MerchantAccountInfos) > 1
}
