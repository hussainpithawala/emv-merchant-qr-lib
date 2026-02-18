package emvqr

import (
	"strings"
	"testing"
)

// -------------------------------------------------------------------------
// CRC Tests
// -------------------------------------------------------------------------

func TestCRC_KnownValue(t *testing.T) {
	// Smoke-test: a well-formed payload encodes and decodes without CRC error.
	p := basePayload()
	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if _, err := Decode(encoded); err != nil {
		t.Fatalf("Decode() CRC error on freshly encoded payload: %v", err)
	}
}

func TestCRC_Algorithm(t *testing.T) {
	// Verify the CRC16-CCITT implementation produces a valid 4-char hex string.
	input := "000201021640001234567890125204525153038405802US5911ABC Hammers6008New York6304"
	crc := crc16CCITT([]byte(input))
	got := crcString(crc)
	if len(got) != 4 {
		t.Errorf("crcString length = %d, want 4", len(got))
	}
	for _, ch := range got {
		if (ch < '0' || ch > '9') && (ch < 'A' || ch > 'F') {
			t.Errorf("crcString contains non-hex character %q", ch)
		}
	}
}

// -------------------------------------------------------------------------
// Round-trip Tests (Encode → Decode)
// -------------------------------------------------------------------------

func TestRoundTrip_BaseExample(t *testing.T) {
	p := NewPayload()
	p.MerchantAccountInfos = []MerchantAccountInfo{
		{ID: "02", Value: "4000123456789012"},
	}
	p.MerchantCategoryCode = "5251"
	p.TransactionCurrency = "840"
	p.CountryCode = "US"
	p.MerchantName = "ABC Hammers"
	p.MerchantCity = "New York"

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	assertEqual(t, "PayloadFormatIndicator", "01", decoded.PayloadFormatIndicator)
	assertEqual(t, "MerchantCategoryCode", "5251", decoded.MerchantCategoryCode)
	assertEqual(t, "TransactionCurrency", "840", decoded.TransactionCurrency)
	assertEqual(t, "CountryCode", "US", decoded.CountryCode)
	assertEqual(t, "MerchantName", "ABC Hammers", decoded.MerchantName)
	assertEqual(t, "MerchantCity", "New York", decoded.MerchantCity)
	if len(decoded.MerchantAccountInfos) != 1 {
		t.Fatalf("expected 1 MAI, got %d", len(decoded.MerchantAccountInfos))
	}
	assertEqual(t, "MAI[0].Value", "4000123456789012", decoded.MerchantAccountInfos[0].Value)
}

func TestRoundTrip_TransactionAmount(t *testing.T) {
	p := basePayload()
	p.TransactionAmount = "10"

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	assertEqual(t, "TransactionAmount", "10", decoded.TransactionAmount)
}

func TestRoundTrip_MultipleNetworks(t *testing.T) {
	p := basePayload()
	p.TransactionAmount = "10"
	if err := p.AddTemplateMerchantAccount("26", "D15600000000",
		DataObject{ID: "01", Value: "A93FO3230QDJ8F93845K"},
	); err != nil {
		t.Fatalf("AddTemplateMerchantAccount: %v", err)
	}

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	if !decoded.HasMultipleNetworks() {
		t.Error("expected HasMultipleNetworks() == true")
	}
	if len(decoded.MerchantAccountInfos) != 2 {
		t.Fatalf("expected 2 MAIs, got %d", len(decoded.MerchantAccountInfos))
	}
	mai26 := decoded.MerchantAccountInfos[1]
	assertEqual(t, "MAI[1].ID", "26", mai26.ID)
	assertEqual(t, "MAI[1].GUID", "D15600000000", mai26.GloballyUniqueID())
	assertEqual(t, "MAI[1].SubField 01", "A93FO3230QDJ8F93845K", mai26.SubField("01"))
}

func TestRoundTrip_FixedConvenienceFee(t *testing.T) {
	p := basePayload()
	p.MerchantCategoryCode = "5812"
	p.TransactionAmount = "50"
	p.MerchantName = "XYZ Restaurant"
	p.MerchantCity = "Miami"
	p.SetFixedConvenienceFee("10.75")

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	assertEqual(t, "TipIndicator", TipIndicatorFixedConvenienceFee, decoded.TipOrConvenienceIndicator)
	assertEqual(t, "FixedFee", "10.75", decoded.ValueConvenienceFeeFixed)

	total, err := decoded.TotalAmount()
	if err != nil {
		t.Fatalf("TotalAmount() error: %v", err)
	}
	if total != 60.75 {
		t.Errorf("TotalAmount() = %v, want 60.75", total)
	}
}

func TestRoundTrip_PromptForTip(t *testing.T) {
	p := basePayload()
	p.TransactionAmount = "50"
	p.SetPromptForTip()

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	assertEqual(t, "TipIndicator", TipIndicatorPromptConsumer, decoded.TipOrConvenienceIndicator)
}

func TestRoundTrip_PercentageFee(t *testing.T) {
	p := basePayload()
	p.MerchantCategoryCode = "9311"
	p.TransactionAmount = "3000"
	p.MerchantName = "National Tax Service"
	p.MerchantCity = "eCommerce"
	p.SetPercentageConvenienceFee("3.00")

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	assertEqual(t, "TipIndicator", TipIndicatorPercentageFee, decoded.TipOrConvenienceIndicator)
	assertEqual(t, "PctFee", "3.00", decoded.ValueConvenienceFeePercent)

	total, err := decoded.TotalAmount()
	if err != nil {
		t.Fatalf("TotalAmount() error: %v", err)
	}
	if total != 3090.0 {
		t.Errorf("TotalAmount() = %v, want 3090.0", total)
	}
}

func TestRoundTrip_AdditionalData_LoyaltyPrompt(t *testing.T) {
	p := basePayload()
	p.TransactionAmount = "10"
	p.SetAdditionalData(func(adf *AdditionalDataField) {
		adf.LoyaltyNumber = PromptValue
	})

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	if !decoded.LoyaltyNumberRequired() {
		t.Error("expected LoyaltyNumberRequired() == true")
	}
	assertEqual(t, "LoyaltyNumber", PromptValue, decoded.AdditionalData.LoyaltyNumber)
}

func TestRoundTrip_AdditionalData_AllFields(t *testing.T) {
	// The Additional Data Field template (ID "62") uses a 2-digit decimal
	// length field, capping any single TLV value at 99 chars. All 9 sub-fields
	// encoded together must therefore total ≤ 99 chars. Values are kept short
	// intentionally; the round-trip correctness is what matters here, not the
	// business meaning of each value.
	//
	// Inner content accounting (ID(2) + len(2) + value):
	//   01 "INV001"  → 10   02 "***"  →  7   03 "Main"  →  8
	//   04 "LN123"   →  9   05 "ORD99" →  9   06 "C42"   →  7
	//   07 "T07"     →  7   08 "Buy"  →  7   09 "AME"   →  7
	//   Total = 71 chars (well within the 99-char limit)
	p := basePayload()
	p.SetAdditionalData(func(adf *AdditionalDataField) {
		adf.BillNumber = "INV001"
		adf.MobileNumber = "***"
		adf.StoreLabel = "Main"
		adf.LoyaltyNumber = "LN123"
		adf.ReferenceLabel = "ORD99"
		adf.CustomerLabel = "C42"
		adf.TerminalLabel = "T07"
		adf.PurposeOfTransaction = "Buy"
		adf.AdditionalConsumerDataRequest = "AME"
	})

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	adf := decoded.AdditionalData
	if adf == nil {
		t.Fatal("AdditionalData is nil after decode")
	}
	assertEqual(t, "BillNumber", "INV001", adf.BillNumber)
	assertEqual(t, "MobileNumber", "***", adf.MobileNumber)
	assertEqual(t, "StoreLabel", "Main", adf.StoreLabel)
	assertEqual(t, "LoyaltyNumber", "LN123", adf.LoyaltyNumber)
	assertEqual(t, "ReferenceLabel", "ORD99", adf.ReferenceLabel)
	assertEqual(t, "CustomerLabel", "C42", adf.CustomerLabel)
	assertEqual(t, "TerminalLabel", "T07", adf.TerminalLabel)
	assertEqual(t, "PurposeOfTransaction", "Buy", adf.PurposeOfTransaction)
	assertEqual(t, "AdditionalConsumerDataRequest", "AME", adf.AdditionalConsumerDataRequest)
}

func TestRoundTrip_AlternateLanguage(t *testing.T) {
	p := basePayload()
	p.TransactionAmount = "10"
	p.SetLanguageTemplate("es", "ABC Martillos", "")

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	if decoded.LanguageTemplate == nil {
		t.Fatal("LanguageTemplate is nil after decode")
	}
	assertEqual(t, "LangPref", "es", decoded.LanguageTemplate.LanguagePreference)
	assertEqual(t, "LangName", "ABC Martillos", decoded.LanguageTemplate.MerchantName)
	assertEqual(t, "PreferredName (es)", "ABC Martillos", decoded.PreferredMerchantName("es"))
	assertEqual(t, "PreferredName (en)", "ABC Hammers", decoded.PreferredMerchantName("en"))
	assertEqual(t, "PreferredCity (es)", "New York", decoded.PreferredMerchantCity("es"))
}

func TestRoundTrip_UnreservedTemplate(t *testing.T) {
	p := basePayload()
	p.UnreservedTemplates = []UnreservedTemplate{
		{
			ID:               "80",
			GloballyUniqueID: "EXAMPLE0000000001",
			SubFields:        []DataObject{{ID: "01", Value: "custom-value"}},
		},
	}

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	if len(decoded.UnreservedTemplates) != 1 {
		t.Fatalf("expected 1 unreserved template, got %d", len(decoded.UnreservedTemplates))
	}
	ut := decoded.UnreservedTemplates[0]
	assertEqual(t, "UT.ID", "80", ut.ID)
	assertEqual(t, "UT.GUID", "EXAMPLE0000000001", ut.GloballyUniqueID)
}

func TestRoundTrip_PostalCode(t *testing.T) {
	p := basePayload()
	p.PostalCode = "10001"

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	assertEqual(t, "PostalCode", "10001", decoded.PostalCode)
}

// -------------------------------------------------------------------------
// TLV Unit Tests
// -------------------------------------------------------------------------

func TestParseTLV_ValidInput(t *testing.T) {
	objs, err := parseTLV("5911ABC Hammers")
	if err != nil {
		t.Fatalf("parseTLV error: %v", err)
	}
	if len(objs) != 1 {
		t.Fatalf("expected 1 object, got %d", len(objs))
	}
	assertEqual(t, "ID", "59", objs[0].id)
	assertEqual(t, "Value", "ABC Hammers", objs[0].value)
}

func TestParseTLV_MultipleObjects(t *testing.T) {
	objs, err := parseTLV("5802US5911ABC Hammers")
	if err != nil {
		t.Fatalf("parseTLV error: %v", err)
	}
	if len(objs) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(objs))
	}
	assertEqual(t, "obj[0].id", "58", objs[0].id)
	assertEqual(t, "obj[0].value", "US", objs[0].value)
	assertEqual(t, "obj[1].id", "59", objs[1].id)
	assertEqual(t, "obj[1].value", "ABC Hammers", objs[1].value)
}

func TestParseTLV_TruncatedData(t *testing.T) {
	_, err := parseTLV("5911ABC")
	if err == nil {
		t.Fatal("expected error for truncated TLV, got nil")
	}
}

func TestParseTLV_TooShort(t *testing.T) {
	_, err := parseTLV("59")
	if err == nil {
		t.Fatal("expected error for input shorter than 4 chars, got nil")
	}
}

func TestEncodeTLV_RoundTrip(t *testing.T) {
	s, err := encodeTLV("59", "ABC Hammers")
	if err != nil {
		t.Fatalf("encodeTLV error: %v", err)
	}
	objs, err := parseTLV(s)
	if err != nil {
		t.Fatalf("parseTLV error: %v", err)
	}
	if len(objs) != 1 {
		t.Fatalf("expected 1 object, got %d", len(objs))
	}
	assertEqual(t, "value", "ABC Hammers", objs[0].value)
}

func TestEncodeTLV_ValueTooLong(t *testing.T) {
	_, err := encodeTLV("59", strings.Repeat("X", 100))
	if err == nil {
		t.Fatal("expected error for value > 99 chars, got nil")
	}
}

// -------------------------------------------------------------------------
// Validation Tests
// -------------------------------------------------------------------------

func TestEncode_MissingRequiredFields(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*Payload)
	}{
		{"NoMAI", func(p *Payload) { p.MerchantAccountInfos = nil }},
		{"NoMCC", func(p *Payload) { p.MerchantCategoryCode = "" }},
		{"NoCurrency", func(p *Payload) { p.TransactionCurrency = "" }},
		{"NoCountry", func(p *Payload) { p.CountryCode = "" }},
		{"NoMerchantName", func(p *Payload) { p.MerchantName = "" }},
		{"NoMerchantCity", func(p *Payload) { p.MerchantCity = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := basePayload()
			tc.fn(p)
			if _, err := Encode(p); err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestEncode_InvalidTipIndicator(t *testing.T) {
	p := basePayload()
	p.TipOrConvenienceIndicator = "99"
	if _, err := Encode(p); err == nil {
		t.Fatal("expected error for invalid tip indicator, got nil")
	}
}

func TestEncode_NilPayload(t *testing.T) {
	if _, err := Encode(nil); err == nil {
		t.Fatal("expected error encoding nil payload, got nil")
	}
}

func TestDecode_InvalidCRC(t *testing.T) {
	encoded, err := Encode(basePayload())
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	corrupted := encoded[:len(encoded)-4] + "0000"
	_, err = Decode(corrupted)
	if err == nil {
		t.Fatal("expected CRC error, got nil")
	}
	if !strings.Contains(err.Error(), "CRC") {
		t.Errorf("expected CRC error message, got: %v", err)
	}
}

func TestDecode_SkipCRCValidation(t *testing.T) {
	encoded, err := Encode(basePayload())
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	corrupted := encoded[:len(encoded)-4] + "0000"
	if _, err = DecodeWithOptions(corrupted, DecodeOptions{SkipCRCValidation: true}); err != nil {
		t.Fatalf("expected no error with SkipCRCValidation, got: %v", err)
	}
}

func TestDecode_MalformedTLV(t *testing.T) {
	if _, err := DecodeWithOptions("0002", DecodeOptions{SkipCRCValidation: true}); err == nil {
		t.Fatal("expected error for truncated TLV, got nil")
	}
}

// -------------------------------------------------------------------------
// Helper Method Tests
// -------------------------------------------------------------------------

func TestAddPrimitiveMerchantAccount_AutoID(t *testing.T) {
	p := NewPayload()
	if err := p.AddPrimitiveMerchantAccount("", "value-a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := p.AddPrimitiveMerchantAccount("", "value-b"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "first MAI ID", "02", p.MerchantAccountInfos[0].ID)
	assertEqual(t, "second MAI ID", "03", p.MerchantAccountInfos[1].ID)
}

func TestAddPrimitiveMerchantAccount_InvalidID(t *testing.T) {
	p := NewPayload()
	if err := p.AddPrimitiveMerchantAccount("01", "value"); err == nil {
		t.Fatal("expected error for reserved ID 01, got nil")
	}
	if err := p.AddPrimitiveMerchantAccount("26", "value"); err == nil {
		t.Fatal("expected error for template-range ID 26, got nil")
	}
}

func TestAddTemplateMerchantAccount_AutoID(t *testing.T) {
	p := NewPayload()
	if err := p.AddTemplateMerchantAccount("", "GUID-001"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "first template MAI ID", "26", p.MerchantAccountInfos[0].ID)
}

func TestTotalAmount_NoAmount(t *testing.T) {
	p := basePayload()
	if _, err := p.TotalAmount(); err == nil {
		t.Fatal("expected error when TransactionAmount is absent, got nil")
	}
}

func TestTotalAmount_BaseOnly(t *testing.T) {
	p := basePayload()
	p.TransactionAmount = "100"
	total, err := p.TotalAmount()
	if err != nil {
		t.Fatalf("TotalAmount() error: %v", err)
	}
	if total != 100.0 {
		t.Errorf("TotalAmount() = %v, want 100.0", total)
	}
}

func TestPreferredMerchantName_NoTemplate(t *testing.T) {
	p := basePayload()
	assertEqual(t, "name without template", "ABC Hammers", p.PreferredMerchantName("es"))
}

func TestHasMultipleNetworks_Single(t *testing.T) {
	if basePayload().HasMultipleNetworks() {
		t.Error("single MAI should not report multiple networks")
	}
}

// -------------------------------------------------------------------------
// Shared test helpers
// -------------------------------------------------------------------------

func basePayload() *Payload {
	p := NewPayload()
	p.MerchantAccountInfos = []MerchantAccountInfo{
		{ID: "02", Value: "4000123456789012"},
	}
	p.MerchantCategoryCode = "5251"
	p.TransactionCurrency = "840"
	p.CountryCode = "US"
	p.MerchantName = "ABC Hammers"
	p.MerchantCity = "New York"
	return p
}

func assertEqual(t *testing.T, field, want, got string) {
	t.Helper()
	if want != got {
		t.Errorf("%s: want %q, got %q", field, want, got)
	}
}
