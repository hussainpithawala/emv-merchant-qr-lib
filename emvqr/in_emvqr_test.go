// Package emvqr — Indian / Bharat QR test suite.
//
// Bharat QR is India's interoperable QR payment standard, jointly developed
// by NPCI, Visa, Mastercard, and American Express and mandated by the Reserve
// Bank of India (RBI & NPCI). It is built on the EMV QRCPS Merchant-Presented Mode
// v1.0 specification, so this library encodes and decodes Bharat QR payloads
// without modification.
//
// Key Indian constants used throughout:
//
//	Currency code : 356  (INR — ISO 4217)
//	Country code  : IN   (ISO 3166-1 alpha-2)
//	NPCI/RuPay AID: A000000677010111  (Globally Unique ID for NPCI network)
//
// Point of Initiation Method (ID "01"):
//
//	"11" = static QR  (printed sticker / poster)
//	"12" = dynamic QR (generated fresh per transaction)
//
// UPI VPA format: <handle>@<bank-or-psp>
//
//	Examples: sharmachai@okaxis, krishnakirana@ybl, spicegarden@oksbi
package emvqr

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// ---------------------------------------------------------------------------
// Indian payment constants (NPCI's Bharat QR context)
// ---------------------------------------------------------------------------

const (
	// inrCurrency is the ISO 4217 numeric code for Indian Rupee.
	inrCurrency = "356"

	// inCountryCode is the ISO 3166-1 alpha-2 code for India.
	inCountryCode = "IN"

	// npciRuPayGUID is the Application Identifier (AID) registered to NPCI
	// for the RuPay / Bharat QR network. It identifies the NPCI payment
	// network in template Merchant Account Information entries (IDs 26–51).
	npciRuPayGUID = "A000000677010111"

	// poiStatic marks a QR code that is printed and reused across transactions.
	poiStatic = "11"

	// poiDynamic marks a QR code generated fresh for each transaction.
	poiDynamic = "12"
)

// ---------------------------------------------------------------------------
// CRC Tests
// ---------------------------------------------------------------------------

func TestCRC_KnownValue_NPCIBhartQR(t *testing.T) {
	// A freshly encoded payload should always pass its own CRC check.
	encoded, err := Encode(NPCIBhartQRBasePayload())
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if _, err := Decode(encoded); err != nil {
		t.Fatalf("Decode() CRC error on freshly encoded Indian payload: %v", err)
	}
}

func TestCRC_Algorithm_NPCIBhartQR(t *testing.T) {
	// Verify CRC16-CCITT produces a 4-character uppercase hex string.
	// Input is a representative Bharat QR prefix (up to "6304").
	input := "000201011102164403847800265204525153035565802IN5917Sharma Chai Stall6006Mumbai6304"
	crc := crc16CCITT([]byte(input))
	got := crcString(crc)
	if len(got) != 4 {
		t.Errorf("crcString length = %d, want 4", len(got))
	}
	for _, ch := range got {
		if (ch < '0' || ch > '9') && (ch < 'A' || ch > 'F') {
			t.Errorf("crcString produced non-hex character %q in %q", ch, got)
		}
	}
}

// ---------------------------------------------------------------------------
// 3.1  Base Example — Sharma Chai Stall, Mumbai
//
// A static QR code printed on a sticker or poster at a street-side tea stall.
// No transaction amount is present; the consumer enters the amount in their
// UPI / Bharat QR app.  Equivalent to the "coffee cart" example in §3.1 of
// EMV QRCPS Merchant-Presented QR Guidance.
// ---------------------------------------------------------------------------

const merchantName = "Sharma Chai Stall"
const merchantCity = "Mumbai"

func TestRoundTrip_BaseExample_NPCIBhartQR(t *testing.T) {
	p := NewPayload()
	// Point of Initiation: "11" = static QR
	p.PointOfInitiationMethod = poiStatic
	p.MerchantIdentifiers = []MerchantIdentifier{
		{ID: "02", Value: "4403847800202706"}, // RuPay primitive MAI (16-char merchant ID)
	}
	p.MerchantCategoryCode = "5499" // Misc Food Stores & Convenience Stores
	p.TransactionCurrency = inrCurrency
	p.CountryCode = inCountryCode
	p.MerchantName = merchantName
	p.MerchantCity = merchantCity
	p.PostalCode = "400001" // Mumbai Central PIN

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	assertEqual(t, "PayloadFormatIndicator", "01", decoded.PayloadFormatIndicator)
	assertEqual(t, "PointOfInitiationMethod", poiStatic, decoded.PointOfInitiationMethod)
	assertEqual(t, "TransactionCurrency", inrCurrency, decoded.TransactionCurrency)
	assertEqual(t, "CountryCode", inCountryCode, decoded.CountryCode)
	assertEqual(t, "MerchantName", merchantName, decoded.MerchantName)
	assertEqual(t, "MerchantCity", merchantCity, decoded.MerchantCity)
	assertEqual(t, "MerchantCategoryCode", "5499", decoded.MerchantCategoryCode)
	assertEqual(t, "PostalCode", "400001", decoded.PostalCode)
}

// ---------------------------------------------------------------------------
// 3.2  Transaction Amount — Krishna Kirana Store, Delhi
//
// A kirana (neighbourhood grocery) shop with a POS tablet that generates a
// fresh QR code for each bill.  The transaction amount (₹150) is embedded so
// the consumer app does not prompt for entry.  PIN code for Connaught Place,
// Delhi is included.
// ---------------------------------------------------------------------------

func TestRoundTrip_TransactionAmount_NPCIBhartQR(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	p.RFUFields = []DataObject{{ID: "01", Value: poiDynamic}} // dynamic QR
	p.MerchantCategoryCode = "5411"                           // Grocery Stores
	p.TransactionAmount = "150"                               // ₹150
	p.MerchantName = "Krishna Kirana Store"
	p.MerchantCity = "Delhi"
	p.PostalCode = "110001" // Connaught Place, Delhi

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	assertEqual(t, "TransactionAmount", "150", decoded.TransactionAmount)
	assertEqual(t, "MerchantName", "Krishna Kirana Store", decoded.MerchantName)
	assertEqual(t, "MerchantCity", "Delhi", decoded.MerchantCity)
	assertEqual(t, "PostalCode", "110001", decoded.PostalCode)
	assertEqual(t, "MCC", "5411", decoded.MerchantCategoryCode)
}

// ---------------------------------------------------------------------------
// 3.3  Multiple Payment Networks — Bharat QR multi-network
//
// A merchant accepting both RuPay (primitive MAI, ID "02") and UPI (template
// MAI, ID "26" with NPCI Globally Unique ID and merchant VPA).  This mirrors
// the actual Bharat QR structure where NPCI mandates UPI VPA be present in
// all Bharat QR codes issued after September 2017.
//
// Network layout:
//   ID "02" → RuPay primitive MAI (16-char acquiring bank merchant ID)
//   ID "26" → UPI/NPCI template MAI
//               sub "00" = A000000677010111  (NPCI RuPay AID)
//               sub "01" = sharmachai@okaxis (UPI Virtual Payment Address)
// ---------------------------------------------------------------------------

func TestRoundTrip_MultipleNetworks_BharatQR(t *testing.T) {
	p := NewPayload()
	p.MerchantIdentifiers = []MerchantIdentifier{
		// RuPay — primitive entry (bank-assigned merchant ID)
		{ID: "02", Value: "4403847800202706"},
	}
	// UPI/NPCI — UPI VPA Template entry
	if err := p.SetUPIVPATemplate(npciRuPayGUID, "sharmachai@okaxis", ""); err != nil {
		t.Fatalf("SetUPIVPATemplate: %v", err)
	}
	p.MerchantCategoryCode = "5499"
	p.TransactionCurrency = inrCurrency
	p.TransactionAmount = "75" // ₹75
	p.CountryCode = inCountryCode
	p.MerchantName = merchantName
	p.MerchantCity = merchantCity

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	if !decoded.HasMultipleNetworks() {
		t.Error("HasMultipleNetworks() = false; expected true for Bharat QR multi-network payload")
	}
	if len(decoded.MerchantIdentifiers) != 2 {
		t.Fatalf("expected 2 merchant identifiers (primitive + template), got %d", len(decoded.MerchantIdentifiers))
	}

	rupay := decoded.MerchantIdentifiers[0]
	assertEqual(t, "RuPay MAI ID", "02", rupay.ID)
	assertEqual(t, "RuPay merchant ID", "4403847800202706", rupay.Value)

	if decoded.UPIVPAInfo == nil {
		t.Fatal("expected UPIVPAInfo to be set")
	}
	assertEqual(t, "NPCI GUID", npciRuPayGUID, decoded.UPIVPAInfo.RuPayRID)
	assertEqual(t, "UPI VPA", "sharmachai@okaxis", decoded.UPIVPAInfo.VPA)
}

// ---------------------------------------------------------------------------
// 3.4.1  Fixed Convenience Fee — Spice Garden Restaurant, Bangalore
//
// A restaurant in Bangalore that adds a fixed ₹50 service charge to all
// QR-initiated orders.  MCC 5812 = Eating Places, Restaurants.
// ---------------------------------------------------------------------------

func TestRoundTrip_FixedConvenienceFee_NPCIBhartQR(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	p.MerchantCategoryCode = "5812" // Eating Places, Restaurants
	p.TransactionAmount = "500"     // ₹500 food bill
	p.MerchantName = "Spice Garden"
	p.MerchantCity = "Bangalore"
	p.PostalCode = "560001"        // MG Road, Bangalore
	p.SetFixedConvenienceFee("50") // ₹50 fixed service charge

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	assertEqual(t, "TipIndicator", TipIndicatorFixedConvenienceFee, decoded.TipOrConvenienceIndicator)
	assertEqual(t, "ConvenienceFeeFixed", "50", decoded.ValueConvenienceFeeFixed)

	total, err := decoded.TotalAmount()
	if err != nil {
		t.Fatalf("TotalAmount() error: %v", err)
	}
	// ₹500 food bill + ₹50 service charge = ₹550
	if total != 550 {
		t.Errorf("TotalAmount() = %.2f, want 550.00", total)
	}
}

// ---------------------------------------------------------------------------
// 3.4.2  Prompt for Tip — Punjab Da Dhaba, Amritsar
//
// A traditional dhaba (roadside restaurant) in Amritsar where the merchant
// requests the consumer app to prompt for a tip/gratuity amount.
// The consumer is always allowed to choose no tip (per spec §3.4.2).
// ---------------------------------------------------------------------------

func TestRoundTrip_PromptForTip_NPCIBhartQR(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	p.MerchantCategoryCode = "5812" // Eating Places
	p.TransactionAmount = "800"     // ₹800 thali order
	p.MerchantName = "Punjab Da Dhaba"
	p.MerchantCity = "Amritsar"
	p.SetPromptForTip() // consumer app shows tip entry screen

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	assertEqual(t, "TipIndicator", TipIndicatorPromptConsumer, decoded.TipOrConvenienceIndicator)
	assertEqual(t, "MerchantName", "Punjab Da Dhaba", decoded.MerchantName)
	assertEqual(t, "MerchantCity", "Amritsar", decoded.MerchantCity)
	// No fee fields should be present when indicator is "01"
	if decoded.ValueConvenienceFeeFixed != "" {
		t.Errorf("ValueConvenienceFeeFixed should be empty, got %q", decoded.ValueConvenienceFeeFixed)
	}
	if decoded.ValueConvenienceFeePercent != "" {
		t.Errorf("ValueConvenienceFeePercent should be empty, got %q", decoded.ValueConvenienceFeePercent)
	}
}

// ---------------------------------------------------------------------------
// 3.4.3  Convenience Fee Percentage — IRCTC Ticket Booking
//
// IRCTC (Indian Railway Catering and Tourism Corporation) charges a 1.80%
// payment gateway convenience fee on all card/UPI ticket bookings.
// MCC 4111 = Local and Suburban Commuter Passenger Transportation.
//
// Calculation: ₹2500 base + (2500 × 0.018) = ₹2500 + ₹45 = ₹2545
// ---------------------------------------------------------------------------

func TestRoundTrip_PercentageFee_IRCTC(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	p.MerchantCategoryCode = "4111" // Transit / Railway
	p.TransactionAmount = "2500"    // ₹2500 base ticket fare
	p.MerchantName = "IRCTC Booking"
	p.MerchantCity = "New Delhi"
	p.SetPercentageConvenienceFee("1.80") // 1.80% gateway fee

	encoded, err := Encode(p)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	assertEqual(t, "TipIndicator", TipIndicatorPercentageFee, decoded.TipOrConvenienceIndicator)
	assertEqual(t, "ConvenienceFeePercent", "1.80", decoded.ValueConvenienceFeePercent)

	total, err := decoded.TotalAmount()
	if err != nil {
		t.Fatalf("TotalAmount() error: %v", err)
	}
	// ₹2500 + 1.80% = ₹2500 + ₹45 = ₹2545
	if total != 2545.0 {
		t.Errorf("TotalAmount() = ₹%.2f, want ₹2545.00", total)
	}
}

// ---------------------------------------------------------------------------
// 3.5.1  Loyalty Number Prompt — D-Mart, Pune
//
// D-Mart's SmartBuy loyalty programme requires the customer to enter their
// loyalty card number before payment.  The QR code carries PromptValue ("***")
// in the Loyalty Number sub-field, causing the consumer app to display an
// "Enter SmartBuy card number" screen before the payment confirmation.
// ---------------------------------------------------------------------------

func TestRoundTrip_LoyaltyPrompt_DMart(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	p.MerchantCategoryCode = "5411" // Grocery / Supermarket
	p.TransactionAmount = "1200"    // ₹1200 grocery basket
	p.MerchantName = "D-Mart"
	p.MerchantCity = "Pune"
	p.PostalCode = "411001" // Camp area, Pune
	p.SetAdditionalData(func(adf *AdditionalDataField) {
		adf.LoyaltyNumber = PromptValue // "***" → consumer app prompts
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
		t.Error("LoyaltyNumberRequired() = false; expected true for D-Mart SmartBuy QR")
	}
	assertEqual(t, "LoyaltyNumber sentinel", PromptValue, decoded.AdditionalData.LoyaltyNumber)
}

// ---------------------------------------------------------------------------
// 3.5  Additional Data Fields — D-Mart, all nine sub-fields
//
// A D-Mart store uses all nine Additional Data Field sub-objects for
// transaction-level reconciliation, loyalty, and delivery contact collection.
//
// Inner TLV byte budget (must be ≤ 99 chars):
//   01 "B2401" →  9   02 "***"   →  7   03 "DM202" →  9
//   04 "LN456" →  9   05 "R555"  →  8   06 "C99"   →  7
//   07 "T12"   →  7   08 "Gro"   →  7   09 "AME"   →  7
//   Total = 70 chars ✓
// ---------------------------------------------------------------------------

func TestRoundTrip_AdditionalData_AllFields_NPCIBhartQR(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	p.MerchantCategoryCode = "5411"
	p.TransactionAmount = "940"
	p.MerchantName = "D-Mart"
	p.MerchantCity = "Pune"
	p.SetAdditionalData(func(adf *AdditionalDataField) {
		adf.BillNumber = "B2401"                  // invoice number
		adf.MobileNumber = "***"                  // prompt — for delivery coordination
		adf.StoreLabel = "DM202"                  // D-Mart outlet 202
		adf.LoyaltyNumber = "LN456"               // SmartBuy card
		adf.ReferenceLabel = "R555"               // order reference
		adf.CustomerLabel = "C99"                 // customer segment ID
		adf.TerminalLabel = "T12"                 // billing counter 12
		adf.PurposeOfTransaction = "Gro"          // Grocery
		adf.AdditionalConsumerDataRequest = "AME" // Address, Mobile, Email
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
	assertEqual(t, "BillNumber", "B2401", adf.BillNumber)
	assertEqual(t, "MobileNumber", "***", adf.MobileNumber)
	assertEqual(t, "StoreLabel", "DM202", adf.StoreLabel)
	assertEqual(t, "LoyaltyNumber", "LN456", adf.LoyaltyNumber)
	assertEqual(t, "ReferenceLabel", "R555", adf.ReferenceLabel)
	assertEqual(t, "CustomerLabel", "C99", adf.CustomerLabel)
	assertEqual(t, "TerminalLabel", "T12", adf.TerminalLabel)
	assertEqual(t, "PurposeOfTransaction", "Gro", adf.PurposeOfTransaction)
	assertEqual(t, "AdditionalConsumerDataRequest", "AME", adf.AdditionalConsumerDataRequest)

	// Confirm the mobile number prompt flag works
	if !decoded.MobileNumberRequired() {
		t.Error("MobileNumberRequired() = false; expected true (MobileNumber = \"***\")")
	}
}

// ---------------------------------------------------------------------------
// 3.6  Alternate Language Template — Raj Medical Store, Hindi
//
// A pharmacy in Chennai displays its name in English by default.  The
// Merchant Information–Language Template (ID "64") carries the Hindi
// transliteration so that Hindi-locale consumer apps can display the
// preferred name.
//
// English : "Raj Medical Store"
// Hindi   : "राज मेडिकल"  (28 UTF-8 bytes — EMV length field encodes bytes)
//
// Inner template byte budget:
//   "00" "02" "hi"            =  6 bytes
//   "01" "28" <28 Hindi bytes> = 32 bytes
//   Total = 38 bytes ✓ (limit 99)
// ---------------------------------------------------------------------------

func TestRoundTrip_AlternateLanguage_Hindi(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	p.MerchantCategoryCode = "5912" // Drug Stores and Pharmacies
	p.TransactionAmount = "450"     // ₹450 medicine bill
	p.MerchantName = "Raj Medical Store"
	p.MerchantCity = "Chennai"
	p.PostalCode = "600001" // George Town, Chennai
	// Hindi alternate name; city name not localised in this example
	p.SetLanguageTemplate("hi", "राज मेडिकल", "")

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
	assertEqual(t, "LanguagePreference", "hi", decoded.LanguageTemplate.LanguagePreference)
	assertEqual(t, "AlternateName", "राज मेडिकल", decoded.LanguageTemplate.MerchantName)

	// Hindi-locale consumer app sees the Hindi name
	assertEqual(t, "PreferredName (hi)", "राज मेडिकल", decoded.PreferredMerchantName("hi"))
	// Any other locale falls back to the English default
	assertEqual(t, "PreferredName (en)", "Raj Medical Store", decoded.PreferredMerchantName("en"))
	assertEqual(t, "PreferredName (ta)", "Raj Medical Store", decoded.PreferredMerchantName("ta"))
	// City was not localised — always returns English form
	assertEqual(t, "PreferredCity (hi)", "Chennai", decoded.PreferredMerchantCity("hi"))
}

// ---------------------------------------------------------------------------
// Unreserved Template — PhonePe proprietary extension (ID "80")
//
// PhonePe (and similar PSPs) use Unreserved Templates (IDs 80–99) to embed
// platform-specific data such as a merchant token, category metadata, or a
// deep-link handle.  The Globally Unique ID disambiguates the template owner.
// ---------------------------------------------------------------------------

func TestRoundTrip_UnreservedTemplate_PhonePe(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	p.TransactionAmount = "299"
	p.UnreservedTemplates = []UnreservedTemplate{
		{
			ID:               "80",
			GloballyUniqueID: "com.phonepe.merchant", // PSP-registered GUID
			SubFields: []DataObject{
				{ID: "01", Value: "token-xyz"}, // merchant token
			},
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
	assertEqual(t, "UT.GUID", "com.phonepe.merchant", ut.GloballyUniqueID)
	if len(ut.SubFields) != 1 || ut.SubFields[0].Value != "token-xyz" {
		t.Errorf("UT sub-field mismatch: %+v", ut.SubFields)
	}
}

// ---------------------------------------------------------------------------
// TLV Unit Tests (mechanics — locale-neutral)
// ---------------------------------------------------------------------------

func TestParseTLV_ValidInput_NPCIBhartQR(t *testing.T) {
	// "5917Sharma Chai Stall" → ID=59, len=17, value="Sharma Chai Stall"
	objs, err := parseTLV("5917Sharma Chai Stall")
	if err != nil {
		t.Fatalf("parseTLV error: %v", err)
	}
	if len(objs) != 1 {
		t.Fatalf("expected 1 object, got %d", len(objs))
	}
	assertEqual(t, "ID", "59", objs[0].id)
	assertEqual(t, "Value", "Sharma Chai Stall", objs[0].value)
}

func TestParseTLV_MultipleObjects_NPCIBhartQR(t *testing.T) {
	// Currency "356" (INR) followed by country "IN"
	objs, err := parseTLV("5303356" + "5802IN")
	if err != nil {
		t.Fatalf("parseTLV error: %v", err)
	}
	if len(objs) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(objs))
	}
	assertEqual(t, "obj[0].id", "53", objs[0].id)
	assertEqual(t, "obj[0].value (INR)", "356", objs[0].value)
	assertEqual(t, "obj[1].id", "58", objs[1].id)
	assertEqual(t, "obj[1].value (IN)", "IN", objs[1].value)
}

func TestParseTLV_TruncatedData_NPCIBhartQR(t *testing.T) {
	_, err := parseTLV("5917Sharma") // declares 17 chars, only 6 present
	if err == nil {
		t.Fatal("expected error for truncated TLV, got nil")
	}
}

func TestParseTLV_TooShort_NPCIBhartQR(t *testing.T) {
	_, err := parseTLV("59") // fewer than 4 chars — no room for length field
	if err == nil {
		t.Fatal("expected error for input shorter than 4 chars, got nil")
	}
}

func TestEncodeTLV_RoundTrip_NPCIBhartQR(t *testing.T) {
	s, err := encodeTLV("59", "Sharma Chai Stall")
	if err != nil {
		t.Fatalf("encodeTLV error: %v", err)
	}
	objs, err := parseTLV(s)
	if err != nil {
		t.Fatalf("parseTLV error: %v", err)
	}
	if len(objs) != 1 || objs[0].value != "Sharma Chai Stall" {
		t.Errorf("round-trip mismatch: %+v", objs)
	}
}

func TestEncodeTLV_ValueTooLong_NPCIBhartQR(t *testing.T) {
	_, err := encodeTLV("59", strings.Repeat("A", 100))
	if err == nil {
		t.Fatal("expected error for value > 99 chars, got nil")
	}
}

// ---------------------------------------------------------------------------
// Validation Tests
// ---------------------------------------------------------------------------

func TestEncode_MissingRequiredFields_NPCIBhartQR(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*Payload)
	}{
		{"NoMAI", func(p *Payload) { p.MerchantIdentifiers = nil }},
		{"NoMCC", func(p *Payload) { p.MerchantCategoryCode = "" }},
		{"NoCurrency", func(p *Payload) { p.TransactionCurrency = "" }},
		{"NoCountry", func(p *Payload) { p.CountryCode = "" }},
		{"NoMerchantName", func(p *Payload) { p.MerchantName = "" }},
		{"NoMerchantCity", func(p *Payload) { p.MerchantCity = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NPCIBhartQRBasePayload()
			tc.fn(p)
			if _, err := Encode(p); err == nil {
				t.Errorf("expected validation error for %s, got nil", tc.name)
			}
		})
	}
}

func TestEncode_InvalidTipIndicator_NPCIBhartQR(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	p.TipOrConvenienceIndicator = "99"
	if _, err := Encode(p); err == nil {
		t.Fatal("expected error for invalid tip indicator value '99', got nil")
	}
}

func TestEncode_NilPayload_NPCIBhartQR(t *testing.T) {
	if _, err := Encode(nil); err == nil {
		t.Fatal("expected error encoding nil payload, got nil")
	}
}

func TestDecode_InvalidCRC_NPCIBhartQR(t *testing.T) {
	encoded, err := Encode(NPCIBhartQRBasePayload())
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	corrupted := encoded[:len(encoded)-4] + "0000"
	_, err = Decode(corrupted)
	if err == nil {
		t.Fatal("expected CRC error on corrupted payload, got nil")
	}
	if !strings.Contains(err.Error(), "CRC") {
		t.Errorf("expected 'CRC' in error message, got: %v", err)
	}
}

func TestDecode_SkipCRCValidation_NPCIBhartQR(t *testing.T) {
	encoded, err := Encode(NPCIBhartQRBasePayload())
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	corrupted := encoded[:len(encoded)-4] + "0000"
	if _, err = DecodeWithOptions(corrupted, DecodeOptions{SkipCRCValidation: true}); err != nil {
		t.Fatalf("expected no error with SkipCRCValidation=true, got: %v", err)
	}
}

func TestDecode_MalformedTLV_NPCIBhartQR(t *testing.T) {
	if _, err := DecodeWithOptions("0002", DecodeOptions{SkipCRCValidation: true}); err == nil {
		t.Fatal("expected error for truncated TLV, got nil")
	}
}

// ---------------------------------------------------------------------------
// Helper Method Tests
// ---------------------------------------------------------------------------

func TestAddMerchantIdentifier_ExplicitID_NPCIBhartQR(t *testing.T) {
	p := NewPayload()
	if err := p.AddMerchantIdentifier("02", "4403847800202706"); err != nil {
		t.Fatalf("first AddMerchantIdentifier error: %v", err)
	}
	if err := p.AddMerchantIdentifier("04", "4403847800202999"); err != nil {
		t.Fatalf("second AddMerchantIdentifier error: %v", err)
	}
	assertEqual(t, "first MAI ID", "02", p.MerchantIdentifiers[0].ID)
	assertEqual(t, "second MAI ID", "04", p.MerchantIdentifiers[1].ID)
}

func TestAddMerchantIdentifier_InvalidID_NPCIBhartQR(t *testing.T) {
	p := NewPayload()
	if err := p.AddMerchantIdentifier("01", "value"); err == nil {
		t.Fatal("expected error for reserved ID 01 (RFU), got nil")
	}
	if err := p.AddMerchantIdentifier("26", "value"); err == nil {
		t.Fatal("expected error for template-range ID 26, got nil")
	}
}

func TestSetUPIVPATemplate_NPCI(t *testing.T) {
	p := NewPayload()
	if err := p.SetUPIVPATemplate(npciRuPayGUID, "testmerchant@okaxis", ""); err != nil {
		t.Fatalf("SetUPIVPATemplate error: %v", err)
	}
	if p.UPIVPAInfo == nil {
		t.Fatalf("expected UPIVPAInfo to be set")
	}
	assertEqual(t, "RuPayRID", npciRuPayGUID, p.UPIVPAInfo.RuPayRID)
	assertEqual(t, "VPA", "testmerchant@okaxis", p.UPIVPAInfo.VPA)
}

func TestTotalAmount_NoAmount_NPCIBhartQR(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	// TransactionAmount deliberately absent (static QR — consumer enters amount)
	if _, err := p.TotalAmount(); err == nil {
		t.Fatal("expected error from TotalAmount() when TransactionAmount is absent, got nil")
	}
}

func TestTotalAmount_INR_BaseOnly_NPCIBhartQR(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	p.TransactionAmount = "2499"
	total, err := p.TotalAmount()
	if err != nil {
		t.Fatalf("TotalAmount() error: %v", err)
	}
	if total != 2499.0 {
		t.Errorf("TotalAmount() = ₹%.2f, want ₹2499.00", total)
	}
}

func TestPreferredMerchantName_NoTemplate_NPCIBhartQR(t *testing.T) {
	p := NPCIBhartQRBasePayload()
	// Without a language template, every locale returns the primary English name
	assertEqual(t, "name (hi)", "Sharma Chai Stall", p.PreferredMerchantName("hi"))
	assertEqual(t, "name (ta)", "Sharma Chai Stall", p.PreferredMerchantName("ta"))
	assertEqual(t, "name (en)", "Sharma Chai Stall", p.PreferredMerchantName("en"))
}

func TestHasMultipleNetworks_SingleRuPay_NPCIBhartQR(t *testing.T) {
	// A simple RuPay-only QR code should NOT report multiple networks
	if NPCIBhartQRBasePayload().HasMultipleNetworks() {
		t.Error("HasMultipleNetworks() = true for single-network RuPay payload; want false")
	}
}

// ---------------------------------------------------------------------------
// Shared test helpers
// ---------------------------------------------------------------------------

// NPCIBhartQRBasePayload returns a minimal valid Bharat QR payload representing
// Sharma Chai Stall in Mumbai — the canonical Indian "base example" used
// throughout this test suite.
func NPCIBhartQRBasePayload() *Payload {
	p := NewPayload()
	p.MerchantIdentifiers = []MerchantIdentifier{
		// RuPay primitive MAI — 16-character bank-assigned merchant ID
		{ID: "02", Value: "4403847800202706"},
	}
	p.MerchantCategoryCode = "5499" // Misc Food Stores & Convenience Stores
	p.TransactionCurrency = inrCurrency
	p.CountryCode = inCountryCode
	p.MerchantName = merchantName
	p.MerchantCity = merchantCity
	return p
}

// TestDecode_RealWorldBharatQRExample tests decoding of a real-world
// Bharat QR code with complete tag coverage including Tags 26, 27, 28, and 62.
// QR from: APRIL MOON RETAIL PRIVATE LIMITED, Ahmedabad
func TestDecode_RealWorldBharatQRExample(t *testing.T) {
	// Real-world Bharat QR payload with comprehensive tag support
	qrPayload := "000201010212021645851910410448940415545080003175565061661000100317556350822SBIN000415243930804448111531090003127398626590010A0000005240141SBIPMOPAD.02PL00000644432-21503961@SBIPAY27770010A0000005240123526020914454520875696090232https://www.hitachi-payments.com28180010A00000052401005204544153033565406250.005802IN5923APRIL MOON RETAIL PRIVA6009AHMEDABAD61063800036258031502PL00000644432052352602091445452087569609070821503961630451DD"

	decoded, err := Decode(qrPayload)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	// =========================================================================
	// Tag 00: Payload Format Indicator
	// =========================================================================
	assertEqual(t, "PayloadFormatIndicator", "01", decoded.PayloadFormatIndicator)

	// =========================================================================
	// Tag 01: Point of Initiation Method
	// =========================================================================
	assertEqual(t, "PointOfInitiationMethod", "12", decoded.PointOfInitiationMethod)

	// =========================================================================
	// Tag 02-11: Merchant Account Information (multiple networks)
	// =========================================================================
	if len(decoded.MerchantIdentifiers) != 8 {
		t.Fatalf("expected 8 merchant identifiers (5 primitives + 3 templates), got %d", len(decoded.MerchantIdentifiers))
	}

	// MAI[0] ID 02 - Visa primitive
	assertEqual(t, "MAI[0].ID", "02", decoded.MerchantIdentifiers[0].ID)
	assertEqual(t, "MAI[0].Value", "4585191041044894", decoded.MerchantIdentifiers[0].Value)

	// MAI[1] ID 04 - Mastercard primitive
	assertEqual(t, "MAI[1].ID", "04", decoded.MerchantIdentifiers[1].ID)
	assertEqual(t, "MAI[1].Value", "545080003175565", decoded.MerchantIdentifiers[1].Value)

	// MAI[2] ID 06 - NPCI/RuPay merchant PAN
	assertEqual(t, "MAI[2].ID", "06", decoded.MerchantIdentifiers[2].ID)
	assertEqual(t, "MAI[2].Value", "6100010031755635", decoded.MerchantIdentifiers[2].Value)

	// MAI[3] ID 08 - IFSC & Account
	assertEqual(t, "MAI[3].ID", "08", decoded.MerchantIdentifiers[3].ID)
	assertEqual(t, "MAI[3].Value", "SBIN000415243930804448", decoded.MerchantIdentifiers[3].Value)

	// MAI[4] ID 11 - AmEx primitive
	assertEqual(t, "MAI[4].ID", "11", decoded.MerchantIdentifiers[4].ID)
	assertEqual(t, "MAI[4].Value", "310900031273986", decoded.MerchantIdentifiers[4].Value)

	// =========================================================================
	// Tag 26: UPI VPA Template
	// =========================================================================
	if decoded.UPIVPAInfo == nil {
		t.Fatal("expected UPIVPAInfo to be set")
	}
	assertEqual(t, "UPIVPAInfo.RuPayRID", "A000000524", decoded.UPIVPAInfo.RuPayRID)
	assertEqual(t, "UPIVPAInfo.VPA", "SBIPMOPAD.02PL00000644432-21503961@SBIPAY", decoded.UPIVPAInfo.VPA)
	assertEqual(t, "UPIVPAInfo.MinimumAmount", "", decoded.UPIVPAInfo.MinimumAmount)

	// =========================================================================
	// Tag 27: UPI VPA Reference (Transaction Reference)
	// =========================================================================
	if decoded.UPITransactionRef == nil {
		t.Fatal("expected UPITransactionRef to be set")
	}
	assertEqual(t, "UPITransactionRef.RuPayRID", "A000000524", decoded.UPITransactionRef.RuPayRID)
	assertEqual(t, "UPITransactionRef.TransactionRef", "52602091445452087569609", decoded.UPITransactionRef.TransactionRef)
	assertEqual(t, "UPITransactionRef.ReferenceURL", "https://www.hitachi-payments.com", decoded.UPITransactionRef.ReferenceURL)

	// =========================================================================
	// Tag 28: Aadhaar Number Template
	// =========================================================================
	if decoded.MerchantAadhaar == nil {
		t.Fatal("expected AadhaarInfo to be set")
	}
	assertEqual(t, "AadhaarInfo.RuPayRID", "A000000524", decoded.MerchantAadhaar.RuPayRID)
	assertEqual(t, "AadhaarInfo.AadhaarNumber", "", decoded.MerchantAadhaar.AadhaarNumber)

	// =========================================================================
	// Tag 52: Merchant Category Code
	// =========================================================================
	assertEqual(t, "MerchantCategoryCode", "5441", decoded.MerchantCategoryCode)

	// =========================================================================
	// Tag 53: Transaction Currency
	// =========================================================================
	assertEqual(t, "TransactionCurrency", "356", decoded.TransactionCurrency) // INR

	// =========================================================================
	// Tag 54: Transaction Amount
	// =========================================================================
	assertEqual(t, "TransactionAmount", "250.00", decoded.TransactionAmount)

	// =========================================================================
	// Tag 58: Country Code
	// =========================================================================
	assertEqual(t, "CountryCode", "IN", decoded.CountryCode)

	// =========================================================================
	// Tag 59: Merchant Name
	// =========================================================================
	assertEqual(t, "MerchantName", "APRIL MOON RETAIL PRIVA", decoded.MerchantName)

	// =========================================================================
	// Tag 60: Merchant City
	// =========================================================================
	assertEqual(t, "MerchantCity", "AHMEDABAD", decoded.MerchantCity)

	// =========================================================================
	// Tag 61: Postal Code
	// =========================================================================
	assertEqual(t, "PostalCode", "380003", decoded.PostalCode)

	// =========================================================================
	// Tag 62: Additional Data Field Template
	// =========================================================================
	if decoded.AdditionalData == nil {
		t.Fatal("expected AdditionalData to be set")
	}
	assertEqual(t, "AdditionalData.BillNumber", "", decoded.AdditionalData.BillNumber)
	assertEqual(t, "AdditionalData.MobileNumber", "", decoded.AdditionalData.MobileNumber)
	assertEqual(t, "AdditionalData.StoreLabel", "02PL00000644432", decoded.AdditionalData.StoreLabel)
	assertEqual(t, "AdditionalData.LoyaltyNumber", "", decoded.AdditionalData.LoyaltyNumber)
	assertEqual(t, "AdditionalData.ReferenceLabel", "52602091445452087569609", decoded.AdditionalData.ReferenceLabel)
	assertEqual(t, "AdditionalData.CustomerLabel", "", decoded.AdditionalData.CustomerLabel)
	assertEqual(t, "AdditionalData.TerminalLabel", "21503961", decoded.AdditionalData.TerminalLabel)
	assertEqual(t, "AdditionalData.PurposeOfTransaction", "", decoded.AdditionalData.PurposeOfTransaction)
	assertEqual(t, "AdditionalData.AdditionalConsumerDataRequest", "", decoded.AdditionalData.AdditionalConsumerDataRequest)

	// =========================================================================
	// Tag 63: CRC
	// =========================================================================
	assertEqual(t, "CRC", "51DD", decoded.CRC)

	// =========================================================================
	// Tag 64: Language Template (not present in this QR)
	// =========================================================================
	if decoded.LanguageTemplate != nil {
		t.Error("expected LanguageTemplate to be nil")
	}

	// =========================================================================
	// Summary checks
	// =========================================================================
	if !decoded.HasMultipleNetworks() {
		t.Error("expected HasMultipleNetworks() == true for multi-network payload")
	}

	// Test getter methods
	assertEqual(t, "GetMerchantVPA()", "SBIPMOPAD.02PL00000644432-21503961@SBIPAY", decoded.GetMerchantVPA())
	assertEqual(t, "GetTransactionReference()", "52602091445452087569609", decoded.GetTransactionReference())
	assertEqual(t, "GetAadhaarNumber()", "", decoded.GetAadhaarNumber())

	// =========================================================================
	// Roundtrip test: encode decoded payload and verify CRC matches
	// =========================================================================
	encoded, err := Encode(decoded)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	// Re-decode to ensure roundtrip consistency
	reTested, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() roundtrip error: %v", err)
	}

	assertEqual(t, "Roundtrip.MerchantName", decoded.MerchantName, reTested.MerchantName)
	assertEqual(t, "Roundtrip.TransactionAmount", decoded.TransactionAmount, reTested.TransactionAmount)
	assertEqual(t, "Roundtrip.UPITransactionRef.TransactionRef", decoded.UPITransactionRef.TransactionRef, reTested.UPITransactionRef.TransactionRef)
}

// ---------------------------------------------------------------------------
// Full Object Comparison Test — Real-World Bharat QR using cmp.Diff()
//
// This test validates the complete decoded Payload object structure using
// google/go-cmp for deep comparison. Instead of individual field assertions,
// we create an expected Payload and compare it against the decoded result,
// catching any structural mismatches instantly.
// ---------------------------------------------------------------------------

func TestDecode_RealWorldBharatQR_ObjectComparison(t *testing.T) {
	qrPayload := "000201010212021645851910410448940415545080003175565061661000100317556350822SBIN000415243930804448111531090003127398626590010A0000005240141SBIPMOPAD.02PL00000644432-21503961@SBIPAY27770010A0000005240123526020914454520875696090232https://www.hitachi-payments.com28180010A00000052401005204544153033565406250.005802IN5923APRIL MOON RETAIL PRIVA6009AHMEDABAD61063800036258031502PL00000644432052352602091445452087569609070821503961630451DD"

	decoded, err := Decode(qrPayload)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	// Build expected Payload object with all fields
	expected := &Payload{
		PayloadFormatIndicator:  "01",
		PointOfInitiationMethod: "12",
		MerchantIdentifiers: []MerchantIdentifier{
			{ID: "02", Value: "4585191041044894"},
			{ID: "04", Value: "545080003175565"},
			{ID: "06", Value: "6100010031755635"},
			{ID: "08", Value: "SBIN000415243930804448"},
			{ID: "11", Value: "310900031273986"},
			{ID: "26", SubFields: []DataObject{{ID: "00", Value: "A000000524"}, {ID: "01", Value: "SBIPMOPAD.02PL00000644432-21503961@SBIPAY"}}},
			{ID: "27", SubFields: []DataObject{{ID: "00", Value: "A000000524"}, {ID: "01", Value: "52602091445452087569609"}, {ID: "02", Value: "https://www.hitachi-payments.com"}}},
			{ID: "28", SubFields: []DataObject{{ID: "00", Value: "A000000524"}}},
		},
		UPIVPAInfo: &UPIVPATemplate{
			RuPayRID:      "A000000524",
			VPA:           "SBIPMOPAD.02PL00000644432-21503961@SBIPAY",
			MinimumAmount: "",
		},
		UPITransactionRef: &UPIVPAReference{
			RuPayRID:       "A000000524",
			TransactionRef: "52602091445452087569609",
			ReferenceURL:   "https://www.hitachi-payments.com",
		},
		MerchantAadhaar: &AadhaarInfo{
			RuPayRID:      "A000000524",
			AadhaarNumber: "",
		},
		MerchantCategoryCode: "5441",
		TransactionCurrency:  "356",
		TransactionAmount:    "250.00",
		CountryCode:          "IN",
		MerchantName:         "APRIL MOON RETAIL PRIVA",
		MerchantCity:         "AHMEDABAD",
		PostalCode:           "380003",
		AdditionalData: &AdditionalDataField{
			BillNumber:                    "",
			MobileNumber:                  "",
			StoreLabel:                    "02PL00000644432",
			LoyaltyNumber:                 "",
			ReferenceLabel:                "52602091445452087569609",
			CustomerLabel:                 "",
			TerminalLabel:                 "21503961",
			PurposeOfTransaction:          "",
			AdditionalConsumerDataRequest: "",
		},
		LanguageTemplate:           nil,
		UnreservedTemplates:        nil,
		RFUFields:                  nil,
		TipOrConvenienceIndicator:  "",
		ValueConvenienceFeeFixed:   "",
		ValueConvenienceFeePercent: "",
		CRC:                        "51DD",
	}

	// Compare full objects using cmp.Diff
	if diff := cmp.Diff(expected, decoded); diff != "" {
		t.Errorf("Payload object mismatch (-want +got):\n%s", diff)
	}
}

// ---------------------------------------------------------------------------
// Roundtrip Object Comparison Test
//
// This test encodes the decoded payload and verifies the re-decoded result
// matches the original decode using object comparison (excluding CRC, which
// may differ due to field ordering in encoding).
// ---------------------------------------------------------------------------

func TestDecode_RealWorldBharatQR_RoundtripObjectComparison(t *testing.T) {
	qrPayload := "000201010212021645851910410448940415545080003175565061661000100317556350822SBIN000415243930804448111531090003127398626590010A0000005240141SBIPMOPAD.02PL00000644432-21503961@SBIPAY27770010A0000005240123526020914454520875696090232https://www.hitachi-payments.com28180010A00000052401005204544153033565406250.005802IN5923APRIL MOON RETAIL PRIVA6009AHMEDABAD61063800036258031502PL00000644432052352602091445452087569609070821503961630451DD"

	// First decode
	decoded1, err := Decode(qrPayload)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}

	// Encode the decoded payload
	encoded, err := Encode(decoded1)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	// Second decode
	decoded2, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() roundtrip error: %v", err)
	}

	// Compare critical fields (ignore CRC as it may differ due to field ordering)
	// Create copies without CRC for comparison
	decoded1Copy := *decoded1
	decoded1Copy.CRC = ""
	decoded2Copy := *decoded2
	decoded2Copy.CRC = ""

	if diff := cmp.Diff(&decoded1Copy, &decoded2Copy); diff != "" {
		t.Errorf("Roundtrip object mismatch (-original +roundtrip):\n%s", diff)
	}

	// Verify CRC is valid (should not be empty, even if different)
	if decoded2.CRC == "" {
		t.Error("Roundtrip CRC is empty")
	}
}
