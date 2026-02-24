# Bharat QR Code Specification - Complete Tag Reference (v4.0.1)

## Overview
The Bharat QR Code uses a **TLV (Tag-Length-Value)** format where each data element is encoded as:
- **Tag ID**: 2-digit numeric identifier
- **Length**: 2-digit numeric value (00-99)
- **Value**: Variable alphanumeric string

---

## Complete Tag Reference Table

| Tag | Name | Encoding | Length | Mandatory/Optional | Description |
|-----|------|----------|--------|-------------------|-------------|
| **00** | Payload Format Indicator | N | 02 | Mandatory | Version number (always "01") |
| **01** | Point of Initiation Method | N | 02 | Mandatory | Method & type indicator (e.g., "11"=QR static, "12"=QR dynamic) |
| **02** | Merchant Identifier (Visa) | ANS | Variable | One mandatory | Network 1 identifier (Visa) |
| **03** | Merchant Identifier (Visa) | ANS | Variable | Optional | Network 2 identifier (Visa) |
| **04** | Merchant Identifier (Mastercard) | ANS | Variable | Optional | Network 3 identifier (Mastercard) |
| **05** | Merchant Identifier (Mastercard) | ANS | Variable | Optional | Network 4 identifier (Mastercard) |
| **06** | Merchant Identifier (NPCI/RuPay) | ANS | Variable | Mandatory | 16-digit NPCI Merchant PAN |
| **07** | Merchant Identifier (NPCI/RuPay) | ANS | Variable | Optional | Reserved for NPCI use |
| **08** | Merchant Identifier (IFSC & Account) | N | Up to 37 | Mandatory | 11-digit IFSC + up to 26-digit account number |
| **09** | Reserved for Future Use | - | - | - | Reserved |
| **10** | Reserved for Future Use | - | - | - | Reserved |
| **11** | Merchant Identifier (AmEx) | ANS | Variable | Optional | Network identifier (American Express) |
| **12** | Merchant Identifier (AmEx) | ANS | Variable | Optional | Network identifier (American Express) |
| **13-25** | Merchant Identifier (Reserved) | ANS | Variable | Optional | Designated range for future networks |
| **26** | UPI VPA (Virtual Payment Address) | Template | Up to 99 | Conditional | Merchant VPA with RuPay RID and optional minimum amount |
| **27** | UPI VPA Reference (Transaction Reference) | Template | Up to 99 | Conditional (Dynamic QR) | **Transaction Reference ID (TR), URL, and RuPay RID** |
| **28** | Aadhaar Number | Template | Up to 99 | Conditional | Merchant Aadhaar number with RuPay RID |
| **29-51** | Merchant Identifiers (Template) | Template | Variable | Optional | Designated range for future identifiers |
| **52** | Merchant Category Code | N | 04 | Mandatory | ISO 18245 business category code (e.g., "5499") |
| **53** | Transaction Currency Code | N | 03 | Mandatory | ISO 4217 currency code (e.g., "356"=INR, "840"=USD) |
| **54** | Transaction Amount | AN | Up to 13 | Optional | Transaction amount with decimals (e.g., "99.12") |
| **55** | Tip or Convenience Indicator | N | 02 | Optional | Tip/fee type indicator ("01"=prompt, "02"=fixed, "03"=percentage) |
| **56** | Value of Convenience Fee Fixed | AN | Up to 13 | Conditional | Fixed convenience fee amount (when tag 55="02") |
| **57** | Value of Convenience Fee Percentage | AN | Up to 05 | Conditional | Percentage-based fee 0-100 (when tag 55="03") |
| **58** | Country Code | AN | 02 | Mandatory | ISO 3166-1 alpha-2 code (e.g., "IN", "US") |
| **59** | Merchant Name | ANS | Up to 23 | Mandatory | "Doing business as" name (max 25 chars in EMV, 23 in Bharat) |
| **60** | Merchant City | AN | Up to 15 | Mandatory | City of merchant operation |
| **61** | Postal Code | AN | Up to 10 | Mandatory | Zip/PIN code of merchant |
| **62** | Additional Data Field Template | ANS (Nested TLV) | Up to 99 | Conditional | Container for 9 standardized sub-fields |
| **63** | CRC | AN | 04 | Mandatory | CRC16-CCITT checksum (always last field) |
| **64** | Merchant Information Language Template | ANS (Nested TLV) | Up to 99 | Optional | Localized merchant name/city |
| **80-99** | Unreserved Templates | Template | Variable | Optional | Proprietary/domestic payment schemes |

---

## Transaction Reference (Tag 27) - Detailed Structure

**Tag 27** is used for **Dynamic QR codes** and contains transaction-specific reference information.

### Sub-tags under Tag 27:

| Sub-Tag | Name | Length | Description |
|---------|------|--------|-------------|
| **00** | RuPay RID | 10 | Constant value: `A000000524` (mandatory) |
| **01** | Transaction Reference (TR) | Min 4, Max 35 | **Transaction Reference ID** - Can include: order number, subscription number, Bill ID, booking ID, insurance renewal reference, etc. Min 4 digits, max 35 digits. Specific to UPI transactions only. |
| **02** | URL (Reference URL) | Up to 26 | Optional URL for transaction details (invoice, bill copy, order details, ticket details, etc.). Must be related to the particular transaction. Must NOT be used for unsolicited information. |

### Transaction Reference (TR) Details:

- **Min Length**: 4 digits
- **Max Length**: 35 digits
- **Format**: Numeric or alphanumeric
- **Purpose**: Used to identify and track specific transactions
- **Examples**:
  - Order number: `ORDER123456`
  - Booking ID: `BOOK2024001234`
  - Subscription number: `SUB0987654321`
  - Bill ID: `BILL20240223001`
  - Insurance reference: `INS-POL-2024-0001`

### Example Tag 27 Encoding:

```
27 35 00 10 A000000524 01 17 ORDER20240223001234 02 04 http://...
```

Breaking down:
- `27` = Tag 27
- `35` = Total length of value (35 characters)
- `00 10 A000000524` = Sub-tag 00, length 10, value "A000000524"
- `01 17 ORDER20240223001234` = Sub-tag 01, length 17, value "ORDER20240223001234"
- `02 04 http://...` = Sub-tag 02, length 4+, optional URL

---

## Additional Data Field Template (Tag 62) - Sub-tags

| Sub-Tag | Item | Max Length | Description |
|---------|------|-----------|-------------|
| **01** | Bill Number | 26 | Invoice number or bill number |
| **02** | Mobile Number | 26 | Phone number for top-up/bill payment (or "***" to prompt) |
| **03** | Store ID | 26 | Distinctive number associated with store/outlet |
| **04** | Loyalty Number | 26 | Loyalty card number (or "***" to prompt) |
| **05** | Reference ID | 26 | Any value as defined by merchant/acquirer to identify transaction |
| **06** | Consumer ID | 26 | Subscriber ID for subscription services or enrollment number |
| **07** | Terminal ID | 26 | Distinctive number associated with terminal/POS counter (mandatory for all RuPay merchants) |
| **08** | Purpose of Transaction | 26 | Transaction Note (TN) - e.g., "Airtime", "Data", "International Package", "Grocery" |
| **09** | Additional Consumer Data Request | 26 | Codes: "A"=address, "M"=mobile, "E"=email |
| **09-49** | Reserved | - | Reserved for future use |
| **50-99** | Dynamically Allocable | - | Available for proprietary use |

**Total Constraint**: Combined length of all sub-fields ≤ 99 characters

**Prompt Value**: Use "***" to signal consumer app to prompt for input (e.g., "02" for mobile number, "04" for loyalty card entry)

---

## Merchant Information Language Template (Tag 64) - Sub-tags

| Sub-Tag | Field | Max Length | Description |
|---------|-------|-----------|-------------|
| **00** | Language Preference | Variable | ISO 639-1 language code (e.g., "hi" for Hindi, "ta" for Tamil, "es" for Spanish) |
| **01** | Merchant Name | 23 | Localized merchant name in specified language |
| **02** | Merchant City | 15 | Localized city name in specified language |

**Important**: English (Tags 59, 60) is always required and remains the default fallback.

---

## UPI VPA (Tag 26) - Sub-tags

| Sub-Tag | Field | Length | Description |
|---------|-------|--------|-------------|
| **00** | RuPay RID | 10 | Constant value: `A000000524` (mandatory) |
| **01** | VPA (Virtual Payment Address) | Up to 50 | Merchant's UPI handle (e.g., `merchant@okaxis`, `shop@upi`) - Uniquely defined by bank during onboarding |
| **02** | Minimum Amount (MAM) | Variable | Minimum Amount to be paid if different from actual amount (Tag 54). Only for Dynamic QR. Limited to 2 decimal places for UPI. |

---

## Aadhaar Number (Tag 28) - Sub-tags

| Sub-Tag | Field | Length | Description |
|---------|-------|--------|-------------|
| **00** | RuPay RID | 10 | Constant value: `A000000524` (mandatory) |
| **01** | Aadhaar Number | 12 | 12-digit Aadhaar number for merchant account linking |

---

## Point of Initiation Method (Tag 01) - Values

**Format**: 2-digit code (XY)

| First Digit (X) | Method | Second Digit (Y) | Type |
|---|---|---|---|
| 1 | QR | 1 | Static |
| 1 | QR | 2 | Dynamic |
| 2 | BLE (Bluetooth Low Energy) | 1 | Static |
| 2 | BLE | 2 | Dynamic |
| 3 | NFC | 1 | Static |
| 3 | NFC | 2 | Dynamic |
| 4-9 | Reserved | 3-9 | Reserved |

**Common Examples**:
- `"11"` = QR static (printed sticker, reused across transactions)
- `"12"` = QR dynamic (generated per transaction)

---

## Tip or Convenience Indicator (Tag 55) - Values

| Value | Meaning | Associated Field | Behavior |
|-------|---------|-------------------|----------|
| **"01"** | Prompt Consumer for Tip | None | Consumer app shows tip entry screen; must allow "no tip" option |
| **"02"** | Fixed Convenience Fee | Tag 56 | Fee added automatically to transaction amount |
| **"03"** | Percentage Convenience Fee | Tag 57 | Percentage applied to base amount |

---

## NPCI Merchant ID (Tag 06) - Structure

The 16-digit Merchant PAN is composed of:

| Position | Length | Content | Example |
|----------|--------|---------|---------|
| 1-6 | 6 digits | Acquirer ID (assigned by NPCI) | `123456` |
| 7 | 1 digit | Filler (always "0") | `0` |
| 8-15 | 8 digits | Running Sequence Number (generated by acquirer) | `00000001` |
| 16 | 1 digit | Check Digit (Luhn algorithm) | Calculated |

**Example Formation**:
- Acquirer ID: `123456`
- Filler: `0`
- Running Sequence: `00000001`
- Check Digit: `X` (Luhn calculated)
- **Result**: `123456000000001X`

---

## IFSC & Account Number (Tag 08) - Structure

| Component | Length | Description |
|-----------|--------|-------------|
| IFSC Code | 11 digits | Bank's Indian Financial System Code |
| Account Number | Up to 26 digits | Merchant's bank account number (right-justified) |
| **Total** | Up to 37 digits | 11 IFSC + up to 26 account digits |

**Example**: `HDFC0001234001234567890123456789012`
- IFSC: `HDFC0001234` (11 digits)
- Account: `001234567890123456789012` (24 digits in this example)

---

## Data Encoding Rules

| Encoding | Allowed Characters | Example | Notes |
|----------|-------------------|---------|-------|
| **N** (Numeric) | 0-9 only | "356", "01", "52" | Decimal digits only |
| **AN** (Alphanumeric) | 0-9, A-Z, space, $ % * + - . / : | "5499", "100.00", "11.95" | No special symbols except listed |
| **ANS** (Alphanumeric String) | Any ISO 18004:2006 characters | "Sharma Chai Stall", "राज मेडिकल" | Includes letters, numbers, symbols, non-ASCII |

**Character Encoding**:
- All values encoded as **UTF-8**
- Length field counts **bytes** (not characters) for multi-byte text (Hindi, Tamil, etc.)

---

## Transaction Amount (Tag 54) - Rules

- **Format**: Alphanumeric with decimal point
- **Valid Examples**: "100.00", "99.85", "99.333", "99.3456"
- **UPI Limit**: Max 2 decimal places (e.g., "99.12")
- **Invalid Values**: Zero is NOT valid
- **Consumer App Rule**: When Tag 54 is present, consumer cannot alter the amount
- **When Absent**: Consumer enters amount in app (static QR)

---

## Convenience Fee Fixed (Tag 56) - Rules

- **Format**: Same as Transaction Amount (alphanumeric with decimal)
- **Examples**: "100.00", "50.50", "25.00"
- **UPI Limit**: Max 2 decimal places
- **Used When**: Tag 55 = "02"
- **Invalid Values**: Zero is NOT valid

---

## Convenience Fee Percentage (Tag 57) - Rules

- **Format**: Whole integers 0-100 with decimal point
- **Valid Examples**: "1.80", "11.95", "3.00", "10.50"
- **Range**: 0 to 100 (percent)
- **Used When**: Tag 55 = "03"
- **Invalid Values**: 0 and 100 are NOT valid
- **Application**: Percentage applied to base transaction amount (Tag 54)

---

## CRC Calculation (Tag 63)

**Algorithm**: CRC16-CCITT (ISO/IEC 3309 compliant)

**Parameters**:
- Polynomial: `0x1021`
- Initial Value: `0xFFFF`
- Input Reflection: None
- Output Reflection: None
- XOR Output: `0x0000`

**Computation Process**:
1. Serialize all TLV objects in order (excluding CRC field itself)
2. Append "6304" (CRC field tag and length) to the string
3. Compute CRC16-CCITT over combined bytes
4. Convert result to 4-character uppercase hexadecimal
5. Append as final TLV: "63" + "04" + crc_hex

**Always Last**: CRC field must be the last field in the QR code

---

## Mandatory vs Optional Fields Summary

### Mandatory Fields (must be present in every valid QR):
- Tag 00: Payload Format Indicator
- At least one Tag from 02-08: Merchant Account Information
- Tag 52: Merchant Category Code
- Tag 53: Transaction Currency Code
- Tag 58: Country Code
- Tag 59: Merchant Name
- Tag 60: Merchant City
- Tag 63: CRC

### Conditional Mandatory:
- Tag 01: Point of Initiation Method (mandatory in specification)
- Tag 61: Postal Code (mandatory per spec)
- Tag 06 or 08: NPCI Merchant ID or IFSC+Account (at least one mandatory)

### Optional Fields:
- Tag 54: Transaction Amount (if absent, consumer enters amount)
- Tag 55: Tip or Convenience Indicator
- Tag 56: Convenience Fee Fixed (required if Tag 55="02")
- Tag 57: Convenience Fee Percentage (required if Tag 55="03")
- Tag 62: Additional Data Field Template
- Tag 64: Merchant Information Language Template
- Tags 80-99: Unreserved/Proprietary Templates

---

## Static vs Dynamic QR Codes

### Static QR (Tag 01 = "11"):
- Same QR code printed/reused across multiple transactions
- Typically no transaction-specific data
- Tag 27 (Transaction Reference) NOT used
- Consumer enters amount in payment app
- Tag 54 (Transaction Amount) typically absent

### Dynamic QR (Tag 01 = "12"):
- Generated per transaction
- Contains transaction-specific reference
- Tag 27 (Transaction Reference) included
- Tag 54 (Transaction Amount) typically present
- Merchant can mandate amount or allow modification

---

## Key Implementation Notes

1. **TLV Ordering**: While specification prefers ordered fields, consumer app must be able to read in any sequence
2. **Length Validation**: Each tag's value length must not exceed specified maximum
3. **Reserved Tags**: Cannot be used by acquirers/issuers without joint NPCI-Visa-Mastercard agreement
4. **UTF-8 Encoding**: Multi-byte characters (Hindi, Tamil) counted in bytes, not characters
5. **CRC Validation**: Must be performed on decode; CRC mismatch indicates corruption or tampering
6. **Transaction Reference (Tag 27 Sub-01)**: Critical for transaction tracking and reconciliation in UPI dynamic QR scenarios

---

## Example Bharat QR Payloads

### Example 1: Simple Static QR (RuPay + UPI)
```
000201011102164403847800202706530356538 02IN5917Sharma Chai Stall6006Mumbai6304XXXX
```

### Example 2: Dynamic QR with Transaction Reference
```
000201021640038478002027065204521153035654100.005802IN5917Sharma Chai Stall6006Mumbai27...
26A000000524...27...01...ORDERID20240223001...
```

### Example 3: With Additional Data and Language Template
```
000201011102164403847800202706530356538 02IN5917Sharma Chai Stall6006Mumbai620262...640260...
```

---

## References

- **Specification**: Bharat QR Code Specification v4.0.1
- **CRC Algorithm**: ISO/IEC 3309
- **Currency Code**: ISO 4217
- **Country Code**: ISO 3166-1
- **MCC**: ISO 18245
- **QR Code Standard**: ISO/IEC 18004:2006
