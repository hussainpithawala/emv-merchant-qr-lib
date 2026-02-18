// Package emvqr — Indian / Bharat QR test suite.
//
// Bharat QR is India's interoperable QR payment standard, jointly developed
// by NPCI, Visa, Mastercard, and American Express and mandated by the Reserve
// Bank of India (RBI). It is built on the EMV QRCPS Merchant-Presented Mode
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
)

// ---------------------------------------------------------------------------
// Indian payment constants (Bharat QR context)
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

func TestCRC_KnownValue_India(t *testing.T) {
	// A freshly encoded payload should always pass its own CRC check.
	encoded, err := Encode(indianBasePayload())
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if _, err := Decode(encoded); err != nil {
		t.Fatalf("Decode() CRC error on freshly encoded Indian payload: %v", err)
	}
}

func TestCRC_Algorithm_India(t *testing.T) {
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

func TestRoundTrip_BaseExample_India(t *testing.T) {
	p := NewPayload()
	// Point of Initiation: "11" = static QR
	p.RFUFields = []DataObject{{ID: "01", Value: poiStatic}}
	p.MerchantAccountInfos = []MerchantAccountInfo{
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
	assertEqual(t, "TransactionCurrency", inrCurrency, decoded.TransactionCurrency)
	assertEqual(t, "CountryCode", inCountryCode, decoded.CountryCode)
	assertEqual(t, "MerchantName", merchantName, decoded.MerchantName)
	assertEqual(t, "MerchantCity", merchantCity, decoded.MerchantCity)
	assertEqual(t, "MerchantCategoryCode", "5499", decoded.MerchantCategoryCode)
	assertEqual(t, "PostalCode", "400001", decoded.PostalCode)

	// Point of Initiation (ID "01") lands in RFUFields since the library
	// does not yet model it as a named field.
	if len(decoded.RFUFields) == 0 {
		t.Fatal("expected RFUFields to contain Point of Initiation (ID 01)")
	}
	assertEqual(t, "PoI Method (static)", poiStatic, decoded.RFUFields[0].Value)
}

// ---------------------------------------------------------------------------
// 3.2  Transaction Amount — Krishna Kirana Store, Delhi
//
// A kirana (neighbourhood grocery) shop with a POS tablet that generates a
// fresh QR code for each bill.  The transaction amount (₹150) is embedded so
// the consumer app does not prompt for entry.  PIN code for Connaught Place,
// Delhi is included.
// ---------------------------------------------------------------------------

func TestRoundTrip_TransactionAmount_India(t *testing.T) {
	p := indianBasePayload()
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
	p.MerchantAccountInfos = []MerchantAccountInfo{
		// RuPay — primitive entry (bank-assigned merchant ID)
		{ID: "02", Value: "4403847800202706"},
	}
	// UPI/NPCI — template entry
	if err := p.AddTemplateMerchantAccount("26", npciRuPayGUID,
		DataObject{ID: "01", Value: "sharmachai@okaxis"}, // UPI VPA
	); err != nil {
		t.Fatalf("AddTemplateMerchantAccount: %v", err)
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
	if len(decoded.MerchantAccountInfos) != 2 {
		t.Fatalf("expected 2 MAIs, got %d", len(decoded.MerchantAccountInfos))
	}

	rupay := decoded.MerchantAccountInfos[0]
	assertEqual(t, "RuPay MAI ID", "02", rupay.ID)
	assertEqual(t, "RuPay merchant ID", "4403847800202706", rupay.Value)

	upi := decoded.MerchantAccountInfos[1]
	assertEqual(t, "UPI MAI ID", "26", upi.ID)
	assertEqual(t, "NPCI GUID", npciRuPayGUID, upi.GloballyUniqueID())
	assertEqual(t, "UPI VPA", "sharmachai@okaxis", upi.SubField("01"))
}

// ---------------------------------------------------------------------------
// 3.4.1  Fixed Convenience Fee — Spice Garden Restaurant, Bangalore
//
// A restaurant in Bangalore that adds a fixed ₹50 service charge to all
// QR-initiated orders.  MCC 5812 = Eating Places, Restaurants.
// ---------------------------------------------------------------------------

func TestRoundTrip_FixedConvenienceFee_India(t *testing.T) {
	p := indianBasePayload()
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
		t.Errorf("TotalAmount() = ₹%.2f, want ₹550.00", total)
	}
}

// ---------------------------------------------------------------------------
// 3.4.2  Prompt for Tip — Punjab Da Dhaba, Amritsar
//
// A traditional dhaba (roadside restaurant) in Amritsar where the merchant
// requests the consumer app to prompt for a tip/gratuity amount.
// The consumer is always allowed to choose no tip (per spec §3.4.2).
// ---------------------------------------------------------------------------

func TestRoundTrip_PromptForTip_India(t *testing.T) {
	p := indianBasePayload()
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
	p := indianBasePayload()
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
	p := indianBasePayload()
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

func TestRoundTrip_AdditionalData_AllFields_India(t *testing.T) {
	p := indianBasePayload()
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
	p := indianBasePayload()
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
	p := indianBasePayload()
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

func TestParseTLV_ValidInput_India(t *testing.T) {
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

func TestParseTLV_MultipleObjects_India(t *testing.T) {
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

func TestParseTLV_TruncatedData_India(t *testing.T) {
	_, err := parseTLV("5917Sharma") // declares 17 chars, only 6 present
	if err == nil {
		t.Fatal("expected error for truncated TLV, got nil")
	}
}

func TestParseTLV_TooShort_India(t *testing.T) {
	_, err := parseTLV("59") // fewer than 4 chars — no room for length field
	if err == nil {
		t.Fatal("expected error for input shorter than 4 chars, got nil")
	}
}

func TestEncodeTLV_RoundTrip_India(t *testing.T) {
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

func TestEncodeTLV_ValueTooLong_India(t *testing.T) {
	_, err := encodeTLV("59", strings.Repeat("A", 100))
	if err == nil {
		t.Fatal("expected error for value > 99 chars, got nil")
	}
}

// ---------------------------------------------------------------------------
// Validation Tests
// ---------------------------------------------------------------------------

func TestEncode_MissingRequiredFields_India(t *testing.T) {
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
			p := indianBasePayload()
			tc.fn(p)
			if _, err := Encode(p); err == nil {
				t.Errorf("expected validation error for %s, got nil", tc.name)
			}
		})
	}
}

func TestEncode_InvalidTipIndicator_India(t *testing.T) {
	p := indianBasePayload()
	p.TipOrConvenienceIndicator = "99"
	if _, err := Encode(p); err == nil {
		t.Fatal("expected error for invalid tip indicator value '99', got nil")
	}
}

func TestEncode_NilPayload_India(t *testing.T) {
	if _, err := Encode(nil); err == nil {
		t.Fatal("expected error encoding nil payload, got nil")
	}
}

func TestDecode_InvalidCRC_India(t *testing.T) {
	encoded, err := Encode(indianBasePayload())
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

func TestDecode_SkipCRCValidation_India(t *testing.T) {
	encoded, err := Encode(indianBasePayload())
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	corrupted := encoded[:len(encoded)-4] + "0000"
	if _, err = DecodeWithOptions(corrupted, DecodeOptions{SkipCRCValidation: true}); err != nil {
		t.Fatalf("expected no error with SkipCRCValidation=true, got: %v", err)
	}
}

func TestDecode_MalformedTLV_India(t *testing.T) {
	if _, err := DecodeWithOptions("0002", DecodeOptions{SkipCRCValidation: true}); err == nil {
		t.Fatal("expected error for truncated TLV, got nil")
	}
}

// ---------------------------------------------------------------------------
// Helper Method Tests
// ---------------------------------------------------------------------------

func TestAddPrimitiveMerchantAccount_AutoID_India(t *testing.T) {
	p := NewPayload()
	if err := p.AddPrimitiveMerchantAccount("", "4403847800202706"); err != nil {
		t.Fatalf("first AddPrimitiveMerchantAccount error: %v", err)
	}
	if err := p.AddPrimitiveMerchantAccount("", "4403847800202999"); err != nil {
		t.Fatalf("second AddPrimitiveMerchantAccount error: %v", err)
	}
	assertEqual(t, "first MAI auto-ID", "02", p.MerchantAccountInfos[0].ID)
	assertEqual(t, "second MAI auto-ID", "03", p.MerchantAccountInfos[1].ID)
}

func TestAddPrimitiveMerchantAccount_InvalidID_India(t *testing.T) {
	p := NewPayload()
	if err := p.AddPrimitiveMerchantAccount("01", "value"); err == nil {
		t.Fatal("expected error for reserved ID 01 (RFU), got nil")
	}
	if err := p.AddPrimitiveMerchantAccount("26", "value"); err == nil {
		t.Fatal("expected error for template-range ID 26, got nil")
	}
}

func TestAddTemplateMerchantAccount_AutoID_NPCI(t *testing.T) {
	p := NewPayload()
	if err := p.AddTemplateMerchantAccount("", npciRuPayGUID,
		DataObject{ID: "01", Value: "testmerchant@okaxis"},
	); err != nil {
		t.Fatalf("AddTemplateMerchantAccount error: %v", err)
	}
	assertEqual(t, "auto-assigned template MAI ID", "26", p.MerchantAccountInfos[0].ID)
	assertEqual(t, "NPCI GUID sub-field", npciRuPayGUID, p.MerchantAccountInfos[0].GloballyUniqueID())
}

func TestTotalAmount_NoAmount_India(t *testing.T) {
	p := indianBasePayload()
	// TransactionAmount deliberately absent (static QR — consumer enters amount)
	if _, err := p.TotalAmount(); err == nil {
		t.Fatal("expected error from TotalAmount() when TransactionAmount is absent, got nil")
	}
}

func TestTotalAmount_INR_BaseOnly_India(t *testing.T) {
	p := indianBasePayload()
	p.TransactionAmount = "2499"
	total, err := p.TotalAmount()
	if err != nil {
		t.Fatalf("TotalAmount() error: %v", err)
	}
	if total != 2499.0 {
		t.Errorf("TotalAmount() = ₹%.2f, want ₹2499.00", total)
	}
}

func TestPreferredMerchantName_NoTemplate_India(t *testing.T) {
	p := indianBasePayload()
	// Without a language template, every locale returns the primary English name
	assertEqual(t, "name (hi)", "Sharma Chai Stall", p.PreferredMerchantName("hi"))
	assertEqual(t, "name (ta)", "Sharma Chai Stall", p.PreferredMerchantName("ta"))
	assertEqual(t, "name (en)", "Sharma Chai Stall", p.PreferredMerchantName("en"))
}

func TestHasMultipleNetworks_SingleRuPay_India(t *testing.T) {
	// A simple RuPay-only QR code should NOT report multiple networks
	if indianBasePayload().HasMultipleNetworks() {
		t.Error("HasMultipleNetworks() = true for single-network RuPay payload; want false")
	}
}

// ---------------------------------------------------------------------------
// Shared test helpers
// ---------------------------------------------------------------------------

// indianBasePayload returns a minimal valid Bharat QR payload representing
// Sharma Chai Stall in Mumbai — the canonical Indian "base example" used
// throughout this test suite.
func indianBasePayload() *Payload {
	p := NewPayload()
	p.MerchantAccountInfos = []MerchantAccountInfo{
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
