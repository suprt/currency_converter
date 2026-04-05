package handler

import "errors"

func validateCurrencyCode(code string) error {
	if len(code) != 3 {
		return errors.New("currency code must be 3 characters")
	}
	for _, c := range code {
		if c < 'A' || c > 'Z' {
			return errors.New("currency code must contain only uppercase Latin letters")
		}
	}
	return nil
}
