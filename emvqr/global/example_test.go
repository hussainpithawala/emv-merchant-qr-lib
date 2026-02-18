package global

import (
	"errors"
	"fmt"

	emvqr "github.com/hussainpithawala/emv-merchant-qr-lib/emvqr"
)

const merchantName = "ABC Hammers"
const merchantCity = "New York"

// ExampleDecode shows how to parse a raw EMV QR Code string.
func ExampleDecode() {
	p := emvqr.NewPayload()
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{{ID: "02", Value: "4000123456789012"}}
	p.MerchantCategoryCode = "5251"
	p.TransactionCurrency = "840"
	p.TransactionAmount = "10"
	p.CountryCode = "US"

	p.MerchantName = merchantName
	p.MerchantCity = merchantCity

	raw, _ := emvqr.Encode(p)

	decoded, err := emvqr.Decode(raw)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(decoded.MerchantName)
	fmt.Println(decoded.MerchantCity)
	fmt.Println(decoded.TransactionCurrency)
	// Output:
	// ABC Hammers
	// New York
	// 840
}

// ExampleEncode shows how to build and encode a basic merchant QR payload.
func ExampleEncode() {
	p := emvqr.NewPayload()
	_ = p.AddPrimitiveMerchantAccount("02", "4000123456789012")
	p.MerchantCategoryCode = "5251"
	p.TransactionCurrency = "840"
	p.CountryCode = "US"
	p.MerchantName = merchantName
	p.MerchantCity = merchantCity

	raw, err := emvqr.Encode(p)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	decoded, _ := emvqr.Decode(raw)
	fmt.Println(decoded.MerchantName)
	// Output:
	// ABC Hammers
}

// ExamplePayload_SetFixedConvenienceFee demonstrates a fixed convenience fee.
func ExamplePayload_SetFixedConvenienceFee() {
	p := emvqr.NewPayload()
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{{ID: "02", Value: "4000123456789012"}}
	p.MerchantCategoryCode = "5812"
	p.TransactionCurrency = "840"
	p.TransactionAmount = "50"
	p.CountryCode = "US"
	p.MerchantName = "XYZ Restaurant"
	p.MerchantCity = "Miami"
	p.SetFixedConvenienceFee("10.75")

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	total, _ := decoded.TotalAmount()
	fmt.Printf("$%.2f\n", total)
	// Output:
	// $60.75
}

// ExamplePayload_SetPercentageConvenienceFee demonstrates a percentage-based fee.
func ExamplePayload_SetPercentageConvenienceFee() {
	p := emvqr.NewPayload()
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{{ID: "02", Value: "4000123456789012"}}
	p.MerchantCategoryCode = "9311"
	p.TransactionCurrency = "840"
	p.TransactionAmount = "3000"
	p.CountryCode = "US"
	p.MerchantName = "National Tax Service"
	p.MerchantCity = "eCommerce"
	p.SetPercentageConvenienceFee("3.00")

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	total, _ := decoded.TotalAmount()
	fmt.Printf("$%.2f\n", total)
	// Output:
	// $3090.00
}

// ExamplePayload_SetPromptForTip demonstrates the tip prompt indicator.
func ExamplePayload_SetPromptForTip() {
	p := emvqr.NewPayload()
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{{ID: "02", Value: "4000123456789012"}}
	p.MerchantCategoryCode = "5812"
	p.TransactionCurrency = "840"
	p.TransactionAmount = "50"
	p.CountryCode = "US"
	p.MerchantName = "XYZ Restaurant"
	p.MerchantCity = "Miami"
	p.SetPromptForTip()

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)
	fmt.Println(decoded.TipOrConvenienceIndicator)
	// Output:
	// 01
}

// ExamplePayload_LoyaltyNumberRequired shows how to detect the loyalty-number prompt.
func ExamplePayload_LoyaltyNumberRequired() {
	p := emvqr.NewPayload()
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{{ID: "02", Value: "4000123456789012"}}
	p.MerchantCategoryCode = "5251"
	p.TransactionCurrency = "840"
	p.TransactionAmount = "10"
	p.CountryCode = "US"
	p.MerchantName = merchantName
	p.MerchantCity = merchantCity
	p.SetAdditionalData(func(adf *emvqr.AdditionalDataField) {
		adf.LoyaltyNumber = emvqr.PromptValue
	})

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)
	fmt.Println(decoded.LoyaltyNumberRequired())
	// Output:
	// true
}

// ExamplePayload_PreferredMerchantName shows language-aware name resolution.
func ExamplePayload_PreferredMerchantName() {
	p := emvqr.NewPayload()
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{{ID: "02", Value: "4000123456789012"}}
	p.MerchantCategoryCode = "5251"
	p.TransactionCurrency = "840"
	p.CountryCode = "US"
	p.MerchantName = merchantName
	p.MerchantCity = merchantCity
	p.SetLanguageTemplate("es", "ABC Martillos", "")

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)

	fmt.Println(decoded.PreferredMerchantName("es")) // Spanish consumer
	fmt.Println(decoded.PreferredMerchantName("fr")) // French consumer (fallback)
	// Output:
	// ABC Martillos
	// ABC Hammers
}

// ExamplePayload_HasMultipleNetworks shows multi-network detection.
func ExamplePayload_HasMultipleNetworks() {
	p := emvqr.NewPayload()
	_ = p.AddPrimitiveMerchantAccount("02", "4000123456789012")
	_ = p.AddTemplateMerchantAccount("26", "D15600000000",
		emvqr.DataObject{ID: "01", Value: "A93FO3230QDJ8F93845K"},
	)
	p.MerchantCategoryCode = "5251"
	p.TransactionCurrency = "840"
	p.TransactionAmount = "10"
	p.CountryCode = "US"
	p.MerchantName = merchantName
	p.MerchantCity = merchantCity

	raw, _ := emvqr.Encode(p)
	decoded, _ := emvqr.Decode(raw)
	fmt.Println(decoded.HasMultipleNetworks())
	// Output:
	// true
}

// ExampleDecode_errorHandling shows typed error inspection.
func ExampleDecode_errorHandling() {
	p := emvqr.NewPayload()
	p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{{ID: "02", Value: "4000123456789012"}}
	p.MerchantCategoryCode = "5251"
	p.TransactionCurrency = "840"
	p.CountryCode = "US"
	p.MerchantName = merchantName
	p.MerchantCity = merchantCity

	raw, _ := emvqr.Encode(p)

	// Corrupt the CRC
	corrupted := raw[:len(raw)-4] + "DEAD"

	_, err := emvqr.Decode(corrupted)
	fmt.Println(errors.Is(err, emvqr.ErrCRCMismatch))
	// Output:
	// true
}
