package phone

import (
	"github.com/nyaruka/phonenumbers"
	"github.com/pkg/errors"
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
)

// Parse parses a given phone number from a country to a specified format.
func Parse(number, isoCountry string, format Format, sanitize bool) (string, error) {
	num, err := phonenumbers.Parse(number, isoCountry)
	if err != nil {
		if errors.Is(err, phonenumbers.ErrInvalidCountryCode) {
			return "", ErrInvalidCountry
		}

		if errors.Is(err, phonenumbers.ErrNotANumber) ||
			errors.Is(err, phonenumbers.ErrTooShortNSN) {
			return "", ErrInvalidNumber
		}

		return "", errors.Wrap(err, "cannot parse number")
	}

	parsed := phonenumbers.Format(num, phonenumbers.PhoneNumberFormat(format))
	if sanitize {
		return grammar.SanitizePhone(parsed), nil
	}

	return parsed, nil
}
