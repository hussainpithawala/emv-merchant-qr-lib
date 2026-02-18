# emv-merchant-qr-lib

[![Go Reference](https://pkg.go.dev/badge/github.com/hussainpithawala/emv-merchant-qr-lib.svg)](https://pkg.go.dev/github.com/hussainpithawala/emv-merchant-qr-lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/hussainpithawala/emv-merchant-qr-lib)](https://goreportcard.com/report/github.com/hussainpithawala/emv-merchant-qr-lib)
[![CI](https://github.com/hussainpithawala/emv-merchant-qr-lib/actions/workflows/ci.yml/badge.svg)](https://github.com/hussainpithawala/emv-merchant-qr-lib/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A pure-Go encoder/decoder library for **EMV® QR Code payloads** as defined in the
[EMV QR Code Specification for Payment Systems (EMV QRCPS) – Merchant-Presented Mode v1.0](https://www.emvco.com/emv-technologies/qrcps/).

## Features

- 🔍 **Decode** any EMV merchant QR Code string into a typed Go struct
- ✍️ **Encode** a structured payload back to a valid QR Code string
- 🔐 **CRC16-CCITT** validation on decode; automatic computation on encode
- 💳 **Multi-network** support – primitive (IDs `02`–`25`) and template (IDs `26`–`51`) Merchant Account Information
- 💰 **Tip & convenience fees** – fixed amount, percentage, and consumer-prompted tip
- 🗂️ **Additional Data Fields** – bill number, loyalty number, store label, reference label, and more
- 🌐 **Alternate language** template for localised merchant name and city display
- 🔧 **Unreserved Templates** (IDs `80`–`99`) for domestic/proprietary payment schemes
- Zero external dependencies

## Installation

```bash
go get github.com/hussainpithawala/emv-merchant-emvqr
```

Requires **Go 1.21** or later.

## Quick Start

```go
import emvqr "github.com/hussainpithawala/emv-merchant-emvqr"
```

### Decode a QR Code string

```go
raw := "000201021640001234567890125204525153038405802US5911ABC Hammers6008New York63047222"

payload, err := emvqr.Decode(raw)
if err != nil {
    log.Fatal(err)
}

fmt.Println(payload.MerchantName)        // ABC Hammers
fmt.Println(payload.MerchantCity)        // New York
fmt.Println(payload.TransactionCurrency) // 840
```

### Encode a QR Code string

```go
p := emvqr.NewPayload()
p.MerchantAccountInfos = []emvqr.MerchantAccountInfo{
    {ID: "02", Value: "4000123456789012"},
}
p.MerchantCategoryCode = "5251"
p.TransactionCurrency  = "840"
p.TransactionAmount    = "10"
p.CountryCode          = "US"
p.MerchantName         = "ABC Hammers"
p.MerchantCity         = "New York"

raw, err := emvqr.Encode(p)
if err != nil {
    log.Fatal(err)
}
fmt.Println(raw)
// 00020102164000123456789012520452515303840540210580...6304XXXX
```

---

## Usage Examples

### Base Example (static sticker QR)

A minimal QR code suitable for printing on a sticker or poster. The consumer
enters the amount in their payment app.

```go
p := emvqr.NewPayload()
if err := p.AddPrimitiveMerchantAccount("02", "4000123456789012"); err != nil {
    log.Fatal(err)
}
p.MerchantCategoryCode = "5251"
p.TransactionCurrency  = "840"
p.CountryCode          = "US"
p.MerchantName         = "ABC Hammers"
p.MerchantCity         = "New York"

raw, _ := emvqr.Encode(p)
```

### Transaction Amount Provided (dynamic QR per transaction)

The merchant POS generates a new QR code for each transaction embedding the amount.
The consumer app must **not** allow the consumer to alter the amount.

```go
p.TransactionAmount = "10"
raw, _ := emvqr.Encode(p)
```

### Multiple Payment Networks

```go
p := emvqr.NewPayload()

// EMV network (primitive MAI)
p.AddPrimitiveMerchantAccount("02", "4000123456789012")

// Domestic / non-EMV network (template MAI)
p.AddTemplateMerchantAccount("26", "D15600000000",
    emvqr.DataObject{ID: "01", Value: "A93FO3230QDJ8F93845K"},
)

p.MerchantCategoryCode = "5251"
p.TransactionCurrency  = "840"
p.TransactionAmount    = "10"
p.CountryCode          = "US"
p.MerchantName         = "ABC Hammers"
p.MerchantCity         = "New York"

raw, _ := emvqr.Encode(p)

// On decode, detect multiple networks:
decoded, _ := emvqr.Decode(raw)
if decoded.HasMultipleNetworks() {
    // Prompt consumer to choose a card
}
```

### Fixed Convenience Fee

```go
p.TransactionAmount = "50"
p.SetFixedConvenienceFee("10.75") // sets indicator "02" + fixed fee field

raw, _ := emvqr.Encode(p)

decoded, _ := emvqr.Decode(raw)
total, _ := decoded.TotalAmount()
fmt.Printf("$%.2f\n", total) // $60.75
```

### Consumer-Prompted Tip

```go
p.TransactionAmount = "50"
p.SetPromptForTip() // sets indicator "01"

raw, _ := emvqr.Encode(p)
// Consumer QR app prompts the consumer to enter a tip amount.
// Must allow consumer to choose no tip (per spec).
```

### Percentage Convenience Fee

```go
p.TransactionAmount = "3000"
p.SetPercentageConvenienceFee("3.00") // 3% fee

raw, _ := emvqr.Encode(p)

decoded, _ := emvqr.Decode(raw)
total, _ := decoded.TotalAmount()
fmt.Printf("$%.2f\n", total) // $3090.00
```

### Additional Data Fields

```go
p.SetAdditionalData(func(adf *emvqr.AdditionalDataField) {
    adf.BillNumber     = "INV-2024-001"
    adf.StoreLabel     = "ABC Hammers – Downtown"
    adf.ReferenceLabel = "ORDER-9987"

    // Use PromptValue ("***") to make the consumer app prompt the consumer
    adf.LoyaltyNumber  = emvqr.PromptValue
    adf.MobileNumber   = emvqr.PromptValue
})

raw, _ := emvqr.Encode(p)

decoded, _ := emvqr.Decode(raw)
if decoded.LoyaltyNumberRequired() {
    // Show "Enter your loyalty number" UI
}
if decoded.MobileNumberRequired() {
    // Show "Enter mobile number to credit" UI
}
```

### Alternate Language Template

```go
p.MerchantName = "ABC Hammers"   // English (default)
p.SetLanguageTemplate("es", "ABC Martillos", "") // Spanish name, city falls back

raw, _ := emvqr.Encode(p)

decoded, _ := emvqr.Decode(raw)

// Consumer app picks the name based on the user's preferred language:
fmt.Println(decoded.PreferredMerchantName("es")) // ABC Martillos
fmt.Println(decoded.PreferredMerchantName("fr")) // ABC Hammers  (fallback)
fmt.Println(decoded.PreferredMerchantCity("es")) // New York     (city not localised)
```

### Unreserved Templates (IDs 80–99)

```go
p.UnreservedTemplates = []emvqr.UnreservedTemplate{
    {
        ID:               "80",
        GloballyUniqueID: "COM.EXAMPLE.PAY0001",
        SubFields: []emvqr.DataObject{
            {ID: "01", Value: "merchant-token-xyz"},
        },
    },
}
```

---

## API Reference

### Top-level Functions

| Function | Description |
|---|---|
| `Decode(raw string) (*Payload, error)` | Decode a raw QR string; validates CRC |
| `DecodeWithOptions(raw string, opts DecodeOptions) (*Payload, error)` | Decode with custom options (e.g., skip CRC) |
| `Encode(p *Payload) (string, error)` | Encode a payload; appends computed CRC |
| `EncodeWithOptions(p *Payload, opts EncodeOptions) (string, error)` | Encode with custom options |

### Payload Methods

| Method | Description |
|---|---|
| `AddPrimitiveMerchantAccount(id, value string) error` | Add a primitive MAI (IDs `02`–`25`) |
| `AddTemplateMerchantAccount(id, guid string, extra ...DataObject) error` | Add a template MAI (IDs `26`–`51`) |
| `SetFixedConvenienceFee(amount string)` | Configure fixed convenience fee |
| `SetPercentageConvenienceFee(percent string)` | Configure percentage-based fee |
| `SetPromptForTip()` | Configure consumer tip prompt |
| `SetAdditionalData(fn func(*AdditionalDataField))` | Set additional data fields |
| `SetLanguageTemplate(lang, name, city string)` | Set alternate language template |
| `TotalAmount() (float64, error)` | Compute base + convenience fee total |
| `LoyaltyNumberRequired() bool` | Reports if app should prompt for loyalty number |
| `MobileNumberRequired() bool` | Reports if app should prompt for mobile number |
| `PreferredMerchantName(lang string) string` | Name in the given language (with fallback) |
| `PreferredMerchantCity(lang string) string` | City in the given language (with fallback) |
| `HasMultipleNetworks() bool` | Reports if multiple MAI entries are present |

### Constants

```go
// Tip or Convenience Indicator values
emvqr.TipIndicatorPromptConsumer      // "01" — consumer enters tip
emvqr.TipIndicatorFixedConvenienceFee // "02" — fixed fee added automatically
emvqr.TipIndicatorPercentageFee       // "03" — percentage fee added automatically

// Additional Data Field sentinel — signals the consumer app to prompt for input
emvqr.PromptValue // "***"
```

---

## Data Object ID Reference

| ID | Field | Notes |
|---|---|---|
| `00` | Payload Format Indicator | Always `"01"` |
| `02`–`25` | Merchant Account Information (primitive) | Up to 24 payment networks |
| `26`–`51` | Merchant Account Information (template) | Non-EMV / domestic networks |
| `52` | Merchant Category Code | ISO 18245 |
| `53` | Transaction Currency | ISO 4217 numeric |
| `54` | Transaction Amount | Optional; absent = consumer enters amount |
| `55` | Tip or Convenience Indicator | `01`, `02`, or `03` |
| `56` | Value of Convenience Fee Fixed | Used when `55` = `02` |
| `57` | Value of Convenience Fee Percentage | Used when `55` = `03` |
| `58` | Country Code | ISO 3166-1 alpha-2 |
| `59` | Merchant Name | Max 25 chars |
| `60` | Merchant City | Max 15 chars |
| `61` | Postal Code | Optional |
| `62` | Additional Data Field Template | See sub-fields below |
| `63` | CRC | CRC16-CCITT, always last |
| `64` | Merchant Information – Language Template | Optional localisation |
| `80`–`99` | Unreserved Templates | Proprietary/domestic use |

**Additional Data Field (ID `62`) sub-fields:**

| Sub-ID | Field |
|---|---|
| `01` | Bill Number |
| `02` | Mobile Number |
| `03` | Store Label |
| `04` | Loyalty Number |
| `05` | Reference Label |
| `06` | Customer Label |
| `07` | Terminal Label |
| `08` | Purpose of Transaction |
| `09` | Additional Consumer Data Request (`A`=address, `M`=mobile, `E`=email) |

---

## Error Handling

```go
payload, err := emvqr.Decode(raw)
if err != nil {
    switch {
    case errors.Is(err, emvqr.ErrCRCMismatch):
        // QR code is corrupted or tampered
    case errors.Is(err, emvqr.ErrInvalidTLV):
        // Malformed TLV structure
    case errors.Is(err, emvqr.ErrMissingRequired):
        // Required field absent (encode-time validation)
    }
}
```

---

## Spec Compliance Notes

- The **Payload Format Indicator** (ID `00`) must always be the first field; this library enforces field ordering on encode.
- The **CRC** (ID `63`) must always be the last field; appended automatically on encode, verified before any parsing on decode.
- When `TransactionAmount` is present, the consumer app **must not** allow the consumer to alter it.
- When the **Tip or Convenience Indicator** is `"02"` (fixed fee) or `"03"` (percentage), the consumer app must add the fee automatically.
- When the **Tip or Convenience Indicator** is `"01"` (prompt for tip), the consumer app must allow the consumer to choose no tip.
- The **Merchant Information – Language Template** is non-normative guidance; English (`ID "59"`, `ID "60"`) remains the required default.

---

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License — see [LICENSE](LICENSE) for full text.

## Disclaimer

EMV® is a registered trademark of EMVCo, LLC. This library is an independent open-source implementation and is not affiliated with or endorsed by EMVCo.
