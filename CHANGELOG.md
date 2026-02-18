# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2024-11-01

### Added
- Initial release implementing EMV QRCPS Merchant-Presented Mode v1.0.
- `Decode` / `DecodeWithOptions` — parse raw QR Code strings into typed `Payload` structs.
- `Encode` / `EncodeWithOptions` — serialise `Payload` structs to QR Code strings.
- CRC16-CCITT validation on decode; automatic CRC computation on encode.
- Full support for all mandatory and optional top-level data objects (IDs `00`–`64`).
- Primitive Merchant Account Information (IDs `02`–`25`).
- Template Merchant Account Information (IDs `26`–`51`) with nested TLV sub-fields.
- Tip or Convenience Indicator with fixed fee, percentage fee, and consumer-prompted tip modes.
- Additional Data Field Template (ID `62`) with all nine specified sub-fields.
- Merchant Information – Language Template (ID `64`) for localised merchant name/city.
- Unreserved Templates (IDs `80`–`99`) for proprietary/domestic payment schemes.
- Convenience helpers: `TotalAmount`, `LoyaltyNumberRequired`, `MobileNumberRequired`,
  `PreferredMerchantName`, `PreferredMerchantCity`, `HasMultipleNetworks`.
- Fluent setters: `SetFixedConvenienceFee`, `SetPercentageConvenienceFee`, `SetPromptForTip`,
  `SetAdditionalData`, `SetLanguageTemplate`.
- `AddPrimitiveMerchantAccount` and `AddTemplateMerchantAccount` with automatic ID assignment.
- `DecodeOptions.SkipCRCValidation` for testing with partial payloads.
- Typed errors: `ErrCRCMismatch`, `ErrInvalidTLV`, `ErrInvalidLength`, `ErrMissingRequired`, `ParseError`.
- Zero external dependencies.
- GitHub Actions CI across Go 1.25.

[Unreleased]: https://github.com/hussainpithawala/emv-merchant-qr-lib/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/hussainpithawala/emv-merchant-qr-lib/releases/tag/v1.0.0
