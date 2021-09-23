package phone

import (
	"errors"

	"github.com/nyaruka/phonenumbers"
	"github.com/thetreep/toolbox/grammar"
)

type Format phonenumbers.PhoneNumberFormat

const (
	E164          Format = Format(phonenumbers.E164)
	INTERNATIONAL        = Format(phonenumbers.INTERNATIONAL)
	NATIONAL             = Format(phonenumbers.NATIONAL)
	RFC3966              = Format(phonenumbers.RFC3966)
)

var (
	ErrInvalidCountry = errors.New("iso country code is invalid")
	ErrInvalidNumber  = errors.New("phone number is invalid")
)

// Parse parses a given phone number from a country to a specified format.
func Parse(number string, isoCountry string, format Format, sanitize bool) (string, error) {
	num, err := phonenumbers.Parse(number, isoCountry)
	switch err {
	case nil:
		// do nothing
	case phonenumbers.ErrInvalidCountryCode:
		return "", ErrInvalidCountry
	case phonenumbers.ErrNotANumber, phonenumbers.ErrTooShortNSN:
		return "", ErrInvalidNumber
	default:
		return "", err
	}

	parsed := phonenumbers.Format(num, phonenumbers.PhoneNumberFormat(format))
	if sanitize {
		return grammar.SanitizePhone(parsed), nil
	}

	return parsed, nil
}
