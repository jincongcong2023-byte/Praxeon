package types

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// CardNumberLen is the total length of an identity card number.
const CardNumberLen = 16

// MaxAgentsPerCard is the maximum number of simultaneously active agent
// addresses that one Player Key may own.
const MaxAgentsPerCard = 10

// GenerateCardNumber generates a random 16-digit card number with a Luhn
// check digit as the last digit. The first 15 digits are random.
func GenerateCardNumber() (string, error) {
	payload := make([]byte, CardNumberLen-1)
	for i := range payload {
		digit, err := randomDigit()
		if err != nil {
			return "", fmt.Errorf("generate card entropy: %w", err)
		}
		payload[i] = byte('0' + digit)
	}
	check := luhnCheckDigit(string(payload))
	return fmt.Sprintf("%s%d", string(payload), check), nil
}

func randomDigit() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

// luhnCheckDigit computes the Luhn check digit for a payload (without the
// check digit). The check digit makes the full number valid under Luhn.
func luhnCheckDigit(payload string) int {
	sum := 0
	double := true // the digit before the check digit is doubled
	for i := len(payload) - 1; i >= 0; i-- {
		d := int(payload[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return (10 - sum%10) % 10
}

// ValidateCardNumber checks a full card number against the Luhn algorithm.
func ValidateCardNumber(number string) bool {
	if len(number) != CardNumberLen {
		return false
	}
	sum := 0
	double := false // the last digit is the check digit, not doubled
	for i := len(number) - 1; i >= 0; i-- {
		if number[i] < '0' || number[i] > '9' {
			return false
		}
		d := int(number[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}
