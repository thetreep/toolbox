package phone

import (
	"errors"
	"fmt"

	"github.com/nyaruka/phonenumbers"
	"github.com/thetreep/toolbox/grammar"
)

type Format phonenumbers.PhoneNumberFormat

const (
	// E164 is the sanitized international format.
	// +33123456789 for french phone number.
	E164 Format = Format(phonenumbers.E164)

	// INTERNATIONAL is the international format with local formatting.
	// +33 1 23 45 67 89 for french phone number.
	INTERNATIONAL Format = Format(phonenumbers.INTERNATIONAL)

	// NATIONAL is the national format with local formatting.
	// 01 23 45 67 89 for french phone number.
	NATIONAL Format = Format(phonenumbers.NATIONAL)
)

var (
	// ErrInvalidCountry means that the provided iso code is wrong.
	ErrInvalidCountry = errors.New("iso country code is invalid")

	// ErrInvalidNumber means that the provided number is invalid.
	ErrInvalidNumber = errors.New("phone number is invalid")

	// ErrInvalidInternationalFormat means that the provided number is not in international format.
	ErrInvalidInternationalFormat = errors.New("phone number is not in international format")
)

type SanitizeMode uint8

const (
	SanitizeOff     SanitizeMode = iota // Keep the phone number as is
	SanitizeDefault                     // Removes only spaces and special characters, keep "+"
	SanitizeStrict                      // SanitizeStrict removes all characters except numbers.
)

// parseNumber is a helper function to parse and format a phone number.
func parseNumber(number, region string, format Format, sanitizeMode SanitizeMode) (string, error) {
	num, err := phonenumbers.Parse(number, region)
	if err != nil {
		if errors.Is(err, phonenumbers.ErrInvalidCountryCode) {
			return "", errors.Join(ErrInvalidCountry, err)
		}

		if errors.Is(err, phonenumbers.ErrNotANumber) ||
			errors.Is(err, phonenumbers.ErrTooShortNSN) {
			return "", errors.Join(ErrInvalidNumber, err)
		}

		return "", fmt.Errorf("cannot parse number: %w", err)
	}

	parsed := phonenumbers.Format(num, phonenumbers.PhoneNumberFormat(format))
	switch sanitizeMode {
	case SanitizeDefault:
		return grammar.SanitizePhone(parsed), nil
	case SanitizeStrict:
		return grammar.SanitizePhoneStrict(parsed), nil
	}

	return parsed, nil
}

// Parse parses a phone number from explicit region and returns it in the desired format.
func Parse(number, region string, format Format, sanitizeMode SanitizeMode) (string, error) {
	return parseNumber(number, region, format, sanitizeMode)
}

// ParseWithFallback parses a phone number and returns it in the desired format.
// If the region cannot be determined from the number, it uses the fallback region.
func ParseWithFallback(number, fallbackRegion string, format Format, sanitizeMode SanitizeMode) (string, error) {
	parsed, err := parseNumber(number, "ZZ", format, sanitizeMode)
	if err != nil && errors.Is(err, ErrInvalidCountry) {
		// Try parsing with the fallback region
		return parseNumber(number, fallbackRegion, format, sanitizeMode)
	}
	return parsed, err
}

// GetRegionFromInternationalNumber determines the region from an international phone number.
func GetRegionFromInternationalNumber(number string) (string, error) {
	parsedNumber, err := phonenumbers.Parse(number, phonenumbers.UNKNOWN_REGION)
	if err != nil {
		if errors.Is(err, phonenumbers.ErrInvalidCountryCode) {
			return "", ErrInvalidInternationalFormat
		}
		return "", fmt.Errorf("error parsing the number: %v", err)
	}

	// Check if the number is valid
	if !phonenumbers.IsValidNumber(parsedNumber) {
		return "", ErrInvalidNumber
	}

	// Get the region code
	regionCode := phonenumbers.GetRegionCodeForNumber(parsedNumber)
	if regionCode == "" {
		return "", fmt.Errorf("unable to determine the region for the number")
	}

	return regionCode, nil
}
