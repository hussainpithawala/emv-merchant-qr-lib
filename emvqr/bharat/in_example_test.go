// Runnable Example functions for the emvqr package, written in the context of
// India's Bharat QR standard (EMV QRCPS Merchant-Presented Mode v1.0 as
// adopted by NPCI, Visa, Mastercard and American Express).
//
// Every Example* function serves dual purpose:
//  1. It appears verbatim in pkg.go.dev package documentation.
//  2. go test verifies the // Output: comment automatically.
//
// Indian payment constants used throughout:
//
//	INR currency code : 356  (ISO 4217)
//	India country code: IN   (ISO 3166-1 alpha-2)
//	NPCI / RuPay AID  : A000000677010111
//	UPI VPA format    : <handle>@<bank-or-psp>
package bharat

import (
	"errors"
	"fmt"
	"testing"

	emvqr "github.com/hussainpithawala/emv-merchant-qr-lib/emvqr"
)

const (
	cityPune      = "Pune"
	cityDelhi     = "Delhi"
	cityBangalore = "Bangalore"
	cityChennai   = "Chennai"
)

// ---------------------------------------------------------------------------
// Decode
// ---------------------------------------------------------------------------

// ExampleDecode demonstrates parsing a Bharat QR payload that carries both a
// primitive RuPay Merchant Account Information (ID "02") and a UPI template
// (ID "26" with NPCI Globally Unique ID and merchant VPA).
//
// Scenario: Sharma Chai Stall, Mumbai — a street-side tea stall that accepts
// both RuPay card-present and UPI QR payments.
func ExampleDecode() {
	// Build a representative Bharat QR payload and encode it first,
	// so the Example is fully self-contained and testable.
	p := emvqr.NewPayload()
	p.MerchantIdentifiers = []emvqr.MerchantIdentifier{
		{ID: "02", Value: "4403847800202706"}, // RuPay primitive MAI
	}
	// UPI / NPCI template MAI
	_ = p.SetUPIVPATemplate("A000000677010111", "sharmachai@okaxis", "")
	p.MerchantCategoryCode = "5499" // Misc Food Stores
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "75"      // ₹75
	p.CountryCode = "IN"
	p.MerchantName = "Sharma Chai Stall"
	p.MerchantCity = cityPune

	raw, _ := emvqr.Encode(p)

	decoded, err := emvqr.Decode(raw)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(decoded.MerchantName)
	fmt.Println(decoded.MerchantCity)
	fmt.Println(decoded.TransactionCurrency)
	fmt.Println(decoded.TransactionAmount)
	fmt.Println(decoded.HasMultipleNetworks())
	// Output:
	// Sharma Chai Stall
	// Pune
	// 356
	// 75
	// true
}

// ---------------------------------------------------------------------------
// Encode
// ---------------------------------------------------------------------------

// ExampleEncode shows how to create a static Bharat QR code for a kirana
// (neighbourhood grocery) store.  No amount is embedded — the consumer enters
// the amount in their UPI or Bharat QR app after scanning.
//
// Scenario: Krishna Kirana Store, Delhi.
func ExampleEncode() {
	p := emvqr.NewPayload()
	_ = p.AddMerchantIdentifier("02", "4403847800209901") // RuPay MAI
	p.MerchantCategoryCode = "5411"                       // Grocery Stores
	p.TransactionCurrency = "356"                         // INR
	p.CountryCode = "IN"
	p.MerchantName = "Krishna Kirana Store"
	p.MerchantCity = cityDelhi

	raw, err := emvqr.Encode(p)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	// Verify it round-trips cleanly.
	decoded, _ := emvqr.Decode(raw)
	fmt.Println(decoded.MerchantName)
	fmt.Println(decoded.TransactionCurrency)
	fmt.Println(decoded.TransactionAmount == "") // no amount → consumer enters it
	// Output:
	// Krishna Kirana Store
	// 356
	// true
}

// ---------------------------------------------------------------------------
// SetFixedConvenienceFee
// ---------------------------------------------------------------------------

// ExamplePayload_SetFixedConvenienceFee shows a restaurant QR code that
// automatically adds a fixed service charge to the food bill.
//
// Scenario: Spice Garden restaurant, Bangalore — adds a flat ₹50 service
// charge on every QR-initiated order.  MCC 5812 = Eating Places.
func ExamplePayload_SetFixedConvenienceFee() {
	p := emvqr.NewPayload()
	p.MerchantIdentifiers = []emvqr.MerchantIdentifier{
		{ID: "02", Value: "4403847800201111"},
	}
	p.MerchantCategoryCode = "5812" // Eating Places, Restaurants
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "500"     // ₹500 food bill
	p.CountryCode = "IN"
	p.MerchantName = "Spice Garden"
	p.MerchantCity = cityBangalore
	p.SetFixedConvenienceFee("50") // ₹50 service charge added automatically

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	total, _ := decoded.TotalAmount()
	fmt.Printf("₹%.2f\n", total) // ₹500 + ₹50
	// Output:
	// ₹550.00
}

// ---------------------------------------------------------------------------
// SetPercentageConvenienceFee
// ---------------------------------------------------------------------------

// ExamplePayload_SetPercentageConvenienceFee_India shows IRCTC's payment gateway fee
// applied as a percentage of the base ticket fare.
//
// Scenario: IRCTC Booking — Indian Railways charges 1.80% convenience fee on
// UPI / QR ticket purchases.  MCC 4111 = Transit / Railway.
//
// Calculation: ₹2500 × 1.80% = ₹45  →  total ₹2545.
func ExamplePayload_SetPercentageConvenienceFee() {
	p := emvqr.NewPayload()
	p.MerchantIdentifiers = []emvqr.MerchantIdentifier{
		{ID: "02", Value: "4403847800203333"},
	}
	p.MerchantCategoryCode = "4111" // Transit / Railway
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "2500"    // ₹2500 base ticket fare
	p.CountryCode = "IN"
	p.MerchantName = "IRCTC Booking"
	p.MerchantCity = "New Delhi"
	p.SetPercentageConvenienceFee("1.80") // 1.80% payment gateway fee

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	total, _ := decoded.TotalAmount()
	fmt.Printf("₹%.2f\n", total) // ₹2500 + ₹45
	// Output:
	// ₹2545.00
}

// ---------------------------------------------------------------------------
// SetPromptForTip
// ---------------------------------------------------------------------------

// ExamplePayload_SetPromptForTip_India shows a dhaba (roadside restaurant) QR code
// that asks the consumer app to display a tip/gratuity entry screen.
//
// Scenario: Punjab Da Dhaba, Amritsar — the merchant wishes to offer consumers
// the option to add a tip.  Per spec §3.4.2, the app must allow the consumer
// to choose no tip.
func ExamplePayload_SetPromptForTip() {
	p := emvqr.NewPayload()
	p.MerchantIdentifiers = []emvqr.MerchantIdentifier{
		{ID: "02", Value: "4403847800204444"},
	}
	p.MerchantCategoryCode = "5812" // Eating Places
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "800"     // ₹800 thali order
	p.CountryCode = "IN"
	p.MerchantName = "Punjab Da Dhaba"
	p.MerchantCity = "Amritsar"
	p.SetPromptForTip() // consumer app shows gratuity entry screen

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)
	fmt.Println(decoded.TipOrConvenienceIndicator) // "01" = prompt consumer
	// Output:
	// 01
}

// ---------------------------------------------------------------------------
// LoyaltyNumberRequired
// ---------------------------------------------------------------------------

// ExamplePayload_LoyaltyNumberRequired_India shows a D-Mart QR code that uses the
// PromptValue sentinel to make the consumer app request the customer's
// SmartBuy loyalty card number before payment.
//
// Scenario: D-Mart supermarket — SmartBuy loyalty programme integration.
func ExamplePayload_LoyaltyNumberRequired() {
	p := emvqr.NewPayload()
	p.MerchantIdentifiers = []emvqr.MerchantIdentifier{
		{ID: "02", Value: "4403847800205555"},
	}
	p.MerchantCategoryCode = "5411" // Grocery / Supermarket
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "1200"    // ₹1200 grocery basket
	p.CountryCode = "IN"
	p.MerchantName = "D-Mart"
	p.MerchantCity = "Pune"
	p.SetAdditionalData(func(adf *emvqr.AdditionalDataField) {
		// PromptValue ("***") signals the consumer app to show
		// "Enter SmartBuy card number" before confirming payment.
		adf.LoyaltyNumber = emvqr.PromptValue
	})

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)
	fmt.Println(decoded.LoyaltyNumberRequired()) // consumer app must prompt
	// Output:
	// true
}

// ---------------------------------------------------------------------------
// PreferredMerchantName  (Hindi language template)
// ---------------------------------------------------------------------------

// ExamplePayload_PreferredMerchantName_India shows a pharmacy QR code with an
// alternate Hindi name in the Merchant Information–Language Template (ID "64").
//
// Scenario: Raj Medical Store, Chennai — the English name "Raj Medical Store"
// is the required default; the Hindi name "राज मेडिकल" is surfaced only to
// consumers whose UPI app is set to Hindi.
func ExamplePayload_PreferredMerchantName() {
	p := emvqr.NewPayload()
	p.MerchantIdentifiers = []emvqr.MerchantIdentifier{
		{ID: "02", Value: "4403847800206666"},
	}
	p.MerchantCategoryCode = "5912" // Drug Stores and Pharmacies
	p.TransactionCurrency = "356"   // INR
	p.CountryCode = "IN"
	p.MerchantName = "Raj Medical Store" // English — always required
	p.MerchantCity = cityChennai
	p.SetLanguageTemplate("hi", "राज मेडिकल", "") // Hindi alternate; city not localised

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	fmt.Println(decoded.PreferredMerchantName("hi")) // Hindi-locale consumer
	fmt.Println(decoded.PreferredMerchantName("en")) // English-locale consumer
	fmt.Println(decoded.PreferredMerchantName("ta")) // Tamil-locale (fallback)
	fmt.Println(decoded.PreferredMerchantCity("hi")) // city not localised — English fallback
	// Output:
	// राज मेडिकल
	// Raj Medical Store
	// Raj Medical Store
	// Chennai
}

// ---------------------------------------------------------------------------
// HasMultipleNetworks  (Bharat QR multi-network)
// ---------------------------------------------------------------------------

// ExamplePayload_HasMultipleNetworks_India shows the canonical Bharat QR structure
// where a single QR code carries both a RuPay primitive MAI and a UPI template
// MAI, enabling acceptance on both networks from one scan.
//
// NPCI mandated UPI VPA to be present in all Bharat QR codes from Sep 2017.
func ExamplePayload_HasMultipleNetworks() {
	p := emvqr.NewPayload()

	// Network 1: RuPay — primitive MAI (bank-assigned 16-char merchant ID)
	_ = p.AddMerchantIdentifier("02", "4403847800202706")

	// Network 2: UPI / NPCI — UPI VPA Template with NPCI AID and merchant VPA
	_ = p.SetUPIVPATemplate("A000000677010111", "sharmachai@okaxis", "")

	p.MerchantCategoryCode = "5499" // Misc Food Stores
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "75"      // ₹75
	p.CountryCode = "IN"
	p.MerchantName = "Sharma Chai Stall"
	p.MerchantCity = "Mumbai"

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	fmt.Println(decoded.HasMultipleNetworks())
	// consumer app would show "Pay via RuPay card or UPI?" choice screen
	fmt.Println(decoded.MerchantIdentifiers[0].ID) // RuPay primitive
	fmt.Println(decoded.GetMerchantVPA())          // UPI VPA
	// Output:
	// true
	// 02
	// sharmachai@okaxis
}

// ---------------------------------------------------------------------------
// Decode — error handling
// ---------------------------------------------------------------------------

// ExampleDecode_errorHandling_India shows how to detect a corrupted Bharat QR code
// using the typed ErrCRCMismatch error.
//
// Scenario: a QR code sticker at a petrol pump has been partially obscured or
// tampered with; the consumer app should reject it and alert the user.
func ExampleDecode_errorHandling() {
	p := emvqr.NewPayload()
	p.MerchantIdentifiers = []emvqr.MerchantIdentifier{
		{ID: "02", Value: "4403847800207777"},
	}
	p.MerchantCategoryCode = "5541" // Service Stations / Petrol Pumps
	p.TransactionCurrency = "356"   // INR
	p.CountryCode = "IN"
	p.MerchantName = "HP Petrol Pump"
	p.MerchantCity = "Bangalore"

	raw, _ := emvqr.Encode(p)

	// Simulate physical damage / tampering by corrupting the last 4 chars (CRC)
	corrupted := raw[:len(raw)-4] + "DEAD"

	_, err := emvqr.Decode(corrupted)
	fmt.Println(errors.Is(err, emvqr.ErrCRCMismatch)) // consumer app rejects this QR
	// Output:
	// true
}

// ---------------------------------------------------------------------------
// Point of Initiation Method (Tag 01)
// ---------------------------------------------------------------------------

// ExamplePayload_SetPointOfInitiationMethod_India shows a dynamic Bharat QR
// for a food delivery platform where each order gets a unique QR code.
//
// Scenario: FreshFood online ordering — static QR at shop counter would be "11"
// (static QR), but here we use "12" for dynamic QR per order.
func ExamplePayload_SetPointOfInitiationMethod() {
	p := emvqr.NewPayload()
	p.MerchantIdentifiers = []emvqr.MerchantIdentifier{
		{ID: "02", Value: "4403847800208888"},
	}
	p.MerchantCategoryCode = "5814" // Fast Food Restaurants
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "350"     // ₹350 order
	p.CountryCode = "IN"
	p.MerchantName = "FreshFood"
	p.MerchantCity = "Hyderabad"
	_ = p.SetPointOfInitiationMethod("1", "2") // "12" = QR, dynamic

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	fmt.Println(decoded.PointOfInitiationMethod)
	// Output:
	// 12
}

// ---------------------------------------------------------------------------
// UPI VPA Reference with Transaction Reference (Tag 27) — Dynamic QRs
// ---------------------------------------------------------------------------

// ExamplePayload_SetUPIVPAReference_India shows a dynamic Bharat QR for an
// e-commerce platform where the transaction reference links to the order number.
//
// Scenario: BookStore online checkout — assigns each order a unique tracking ID.
// Tag 27 enables dynamic QR generation without regenerating all sub-networks.
func ExamplePayload_SetUPIVPAReference() {
	p := emvqr.NewPayload()
	p.MerchantIdentifiers = []emvqr.MerchantIdentifier{
		{ID: "02", Value: "4403847800209009"},
	}
	// UPI template (common in Bharat QR)
	_ = p.SetUPIVPATemplate("A000000677010111", "bookstore@okhdfcbank", "")
	p.MerchantCategoryCode = "5942" // Book Stores
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "599"     // ₹599 order total
	p.CountryCode = "IN"
	p.MerchantName = "BookStore Online"
	p.MerchantCity = "Delhi"
	_ = p.SetPointOfInitiationMethod("1", "2") // Dynamic QR
	_ = p.SetUPIVPAReference("ORD-2024-12345", "bookstore.in/12345")

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	fmt.Println(decoded.TransactionAmount)
	fmt.Println(decoded.GetTransactionReference())
	// Output:
	// 599
	// ORD-2024-12345
}

// ---------------------------------------------------------------------------
// Aadhaar Number Template (Tag 28)
// ---------------------------------------------------------------------------

// ExamplePayload_SetAadhaarNumber_India shows a Bharat QR for an IRDA-regulated
// insurance company that links purchases to Aadhaar-verified customer records.
//
// Scenario: HealthInsure Ltd — premium payment via Aadhaar-linked Bharat QR.
func ExamplePayload_SetAadhaarNumber() {
	p := emvqr.NewPayload()
	p.MerchantIdentifiers = []emvqr.MerchantIdentifier{
		{ID: "02", Value: "4403847800201010"},
	}
	p.MerchantCategoryCode = "6211" // Insurance Carriers
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "4999"    // ₹4999 premium
	p.CountryCode = "IN"
	p.MerchantName = "HealthInsure Ltd"
	p.MerchantCity = "Chennai"
	_ = p.SetAadhaarNumber("123456789012")

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	fmt.Println(decoded.GetAadhaarNumber())
	fmt.Println(len(decoded.GetAadhaarNumber()) == 12)
	// Output:
	// 123456789012
	// true
}

// ---------------------------------------------------------------------------
// UPI VPA Info (Tag 26) — Typed Access
// ---------------------------------------------------------------------------

// TestUPIVPAInfoAccess demonstrates accessing the merchant's UPI
// Virtual Payment Address using the typed getter method.
//
// Scenario: Street vendor — accessing merchant VPA without manually parsing SubFields.
func TestUPIVPAInfoAccess(t *testing.T) {
	p := emvqr.NewPayload()
	_ = p.AddMerchantIdentifier("02", "4403847800201111")
	_ = p.SetUPIVPATemplate("A000000677010111", "vendor@upi", "")
	p.MerchantCategoryCode = "5499"
	p.TransactionCurrency = "356"
	p.CountryCode = "IN"
	p.MerchantName = "Street Vendor"
	p.MerchantCity = "Mumbai"

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	// Access merchant VPA using typed getter
	fmt.Println(decoded.GetMerchantVPA())
	fmt.Println(decoded.UPIVPAInfo != nil)
	fmt.Println(decoded.UPIVPAInfo.VPA)
	// Verify
	if decoded.GetMerchantVPA() != "vendor@upi" {
		t.Errorf("expected vendor@upi, got %s", decoded.GetMerchantVPA())
	}
	if decoded.UPIVPAInfo == nil {
		t.Error("expected UPIVPAInfo to be set")
	}
	if decoded.UPIVPAInfo.VPA != "vendor@upi" {
		t.Errorf("expected vendor@upi, got %s", decoded.UPIVPAInfo.VPA)
	}
}

// ---------------------------------------------------------------------------
// UPI VPA Info with Minimum Amount (Tag 26, SubTag 02)
// ---------------------------------------------------------------------------

// TestUPIVPAMinimumAmount shows dynamic Bharat QR where a merchant
// specifies a minimum purchase amount for UPI transactions.
//
// Scenario: E-commerce platform — minimum order value ₹100 for free shipping.
func TestUPIVPAMinimumAmount(t *testing.T) {
	p := emvqr.NewPayload()
	_ = p.AddMerchantIdentifier("02", "4403847800202222")
	_ = p.SetUPIVPATemplate("A000000677010111", "shop@okhdfcbank", "100.00")
	p.MerchantCategoryCode = "5411"
	p.TransactionCurrency = "356"
	p.CountryCode = "IN"
	p.MerchantName = "ECommerce Shop"
	p.MerchantCity = "Bangalore"
	_ = p.SetPointOfInitiationMethod("1", "2") // Dynamic QR

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	// Verify
	if decoded.GetMinimumAmount() != "100.00" {
		t.Errorf("expected 100.00, got %s", decoded.GetMinimumAmount())
	}
	if decoded.UPIVPAInfo.MinimumAmount != "100.00" {
		t.Errorf("expected 100.00, got %s", decoded.UPIVPAInfo.MinimumAmount)
	}
}

// ---------------------------------------------------------------------------
// UPI Transaction Reference with URL (Tag 27) — Typed Access
// ---------------------------------------------------------------------------

// TestUPITransactionRefWithURL demonstrates a dynamic QR that
// links to an invoice URL along with the transaction reference.
//
// Scenario: SaaS platform — payment QR includes order number and invoice link.
func TestUPITransactionRefWithURL(t *testing.T) {
	p := emvqr.NewPayload()
	_ = p.AddMerchantIdentifier("02", "4403847800203333")
	_ = p.SetUPIVPATemplate("A000000677010111", "saas@upi", "")
	p.MerchantCategoryCode = "7372" // Software Services
	p.TransactionCurrency = "356"
	p.TransactionAmount = "999"
	p.CountryCode = "IN"
	p.MerchantName = "SaaS Platform"
	p.MerchantCity = "Pune"
	_ = p.SetPointOfInitiationMethod("1", "2") // Dynamic QR
	_ = p.SetUPIVPAReference("SUB-2024-98765", "platform.com/invoice/98765")

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	// Verify
	if decoded.GetTransactionReference() != "SUB-2024-98765" {
		t.Errorf("expected SUB-2024-98765, got %s", decoded.GetTransactionReference())
	}
	if decoded.UPITransactionRef.ReferenceURL != "platform.com/invoice/98765" {
		t.Errorf("expected platform.com/invoice/98765, got %s", decoded.UPITransactionRef.ReferenceURL)
	}
	if decoded.UPITransactionRef == nil {
		t.Error("expected UPITransactionRef to be set")
	}
}

// ---------------------------------------------------------------------------
// UPI VPA Info — SubFields Access (Generic)
// ---------------------------------------------------------------------------

// TestUPIVPAInfoSubFields demonstrates accessing UPI VPA
// template data via the generic SubFields array in MerchantIdentifiers.
//
// Scenario: Generic QR decoder — accessing template data through the hybrid structure.
func TestUPIVPAInfoSubFields(t *testing.T) {
	p := emvqr.NewPayload()
	_ = p.AddMerchantIdentifier("02", "4403847800204444")
	_ = p.SetUPIVPATemplate("A000000677010111", "retail@upi", "")
	p.MerchantCategoryCode = "5411"
	p.TransactionCurrency = "356"
	p.CountryCode = "IN"
	p.MerchantName = "Retail Store"
	p.MerchantCity = "Delhi"

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	// Find UPI VPA template (ID "26") in MerchantIdentifiers
	found := false
	for _, mi := range decoded.MerchantIdentifiers {
		if mi.ID == "26" {
			found = true
			if len(mi.SubFields) != 2 {
				t.Errorf("expected 2 SubFields, got %d", len(mi.SubFields))
			}
			// SubFields should contain: RuPayRID (00) and VPA (01)
			expectedSubTags := map[string]string{
				"00": "A000000677010111",
				"01": "retail@upi",
			}
			for _, sf := range mi.SubFields {
				if expected, ok := expectedSubTags[sf.ID]; ok {
					if sf.Value != expected {
						t.Errorf("SubTag %s: expected %s, got %s", sf.ID, expected, sf.Value)
					}
				}
			}
		}
	}
	if !found {
		t.Error("expected to find UPI VPA template (ID 26) in MerchantIdentifiers")
	}
}

// ---------------------------------------------------------------------------
// Multi-Network with UPI VPA and Transaction Reference
// ---------------------------------------------------------------------------

// TestMultiNetworkUPIDynamic shows a Bharat QR that supports
// both RuPay (primitive) and UPI (template) with dynamic transaction reference.
//
// Scenario: Logistics company — dynamic QR for each shipment with order tracking.
func TestMultiNetworkUPIDynamic(t *testing.T) {
	p := emvqr.NewPayload()
	_ = p.AddMerchantIdentifier("02", "4403847800205555")
	_ = p.SetUPIVPATemplate("A000000677010111", "logistics@icici", "")
	p.MerchantCategoryCode = "4215" // Courier Services
	p.TransactionCurrency = "356"
	p.TransactionAmount = "249"
	p.CountryCode = "IN"
	p.MerchantName = "FastShip Logistics"
	p.MerchantCity = "Chennai"
	_ = p.SetPointOfInitiationMethod("1", "2") // Dynamic QR
	_ = p.SetUPIVPAReference("SHIP-2024-456789", "fastship.in/track/456789")

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	// Verify
	if decoded.MerchantName != "FastShip Logistics" {
		t.Errorf("expected FastShip Logistics, got %s", decoded.MerchantName)
	}
	if !decoded.HasMultipleNetworks() {
		t.Error("expected HasMultipleNetworks to be true")
	}
	// Should have: 02 (RuPay primitive), 26 (UPI VPA template), 27 (UPI VPA Reference template)
	if len(decoded.MerchantIdentifiers) != 3 {
		t.Errorf("expected 3 MerchantIdentifiers (02, 26, 27), got %d", len(decoded.MerchantIdentifiers))
	}
	if decoded.GetMerchantVPA() != "logistics@icici" {
		t.Errorf("expected logistics@icici, got %s", decoded.GetMerchantVPA())
	}
	if decoded.GetTransactionReference() != "SHIP-2024-456789" {
		t.Errorf("expected SHIP-2024-456789, got %s", decoded.GetTransactionReference())
	}
}
