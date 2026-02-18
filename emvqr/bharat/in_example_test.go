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

	emvqr "github.com/hussainpithawala/emv-merchant-qr-lib/emvqr"
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
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{
		{ID: "02", Value: "4403847800202706"}, // RuPay primitive MAI
	}
	// UPI / NPCI template MAI
	_ = p.AddTemplateMerchantAccount("26", "A000000677010111",
		emvqr.DataObject{ID: "01", Value: "sharmachai@okaxis"}, // UPI VPA
	)
	p.MerchantCategoryCode = "5499" // Misc Food Stores
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "75"      // ₹75
	p.CountryCode = "IN"
	p.MerchantName = "Sharma Chai Stall"
	p.MerchantCity = "Mumbai"

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
	// Mumbai
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
	_ = p.AddPrimitiveMerchantAccount("02", "4403847800209901") // RuPay MAI
	p.MerchantCategoryCode = "5411"                             // Grocery Stores
	p.TransactionCurrency = "356"                               // INR
	p.CountryCode = "IN"
	p.MerchantName = "Krishna Kirana Store"
	p.MerchantCity = "Delhi"

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
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{
		{ID: "02", Value: "4403847800201111"},
	}
	p.MerchantCategoryCode = "5812" // Eating Places, Restaurants
	p.TransactionCurrency = "356"   // INR
	p.TransactionAmount = "500"     // ₹500 food bill
	p.CountryCode = "IN"
	p.MerchantName = "Spice Garden"
	p.MerchantCity = "Bangalore"
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
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{
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
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{
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
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{
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
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{
		{ID: "02", Value: "4403847800206666"},
	}
	p.MerchantCategoryCode = "5912" // Drug Stores and Pharmacies
	p.TransactionCurrency = "356"   // INR
	p.CountryCode = "IN"
	p.MerchantName = "Raj Medical Store" // English — always required
	p.MerchantCity = "Chennai"
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
	_ = p.AddPrimitiveMerchantAccount("02", "4403847800202706")

	// Network 2: UPI / NPCI — template MAI with NPCI AID and merchant VPA
	_ = p.AddTemplateMerchantAccount("26", "A000000677010111",
		emvqr.DataObject{ID: "01", Value: "sharmachai@okaxis"},
	)

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
	fmt.Println(decoded.MerchantAccountInfos[0].ID)             // RuPay primitive
	fmt.Println(decoded.MerchantAccountInfos[1].ID)             // UPI template
	fmt.Println(decoded.MerchantAccountInfos[1].SubField("01")) // UPI VPA
	// Output:
	// true
	// 02
	// 26
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
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{
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
