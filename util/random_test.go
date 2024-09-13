package util

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestRandomInt(t *testing.T) {
	// Similar logic to the previous example
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 1000; i++ {
		min := int64(1)
		max := int64(10)
		result := RandomInt(min, max)
		if result < min || result > max {
			t.Errorf("RandomInt(%d, %d) = %d; expected between %d and %d", min, max, result, min, max)
		}
	}
}

func TestRandomString(t *testing.T) {
	// Test string length
	for _, length := range []int{5, 10, 20} {
		result := RandomString(length)
		if len(result) != length {
			t.Errorf("RandomString(%d) length = %d; expected %d", length, len(result), length)
		}
	}

	// Test string content (basic check)
	for range [100]int{} {
		result := RandomString(10)
		if !strings.ContainsAny(result, alphabet) {
			t.Errorf("RandomString(10) contains characters outside alphabet")
		}
	}
}

func TestRandomOwner(t *testing.T) {
	// Similar logic to TestRandomString for length
	for range [100]int{} {
		result := RandomOwner()
		if len(result) != 6 {
			t.Errorf("RandomOwner() length = %d; expected 6", len(result))
		}
	}
}

func TestRandomMoney(t *testing.T) {
	// Test range of values
	for range [100]int{} {
		result := RandomMoney()
		if result < 0 || result > 1000 {
			t.Errorf("RandomMoney() = %d; expected between 0 and 1000", result)
		}
	}
}

func TestRandomCurrency(t *testing.T) {
	// Test if returned value is in the list of currencies
	expectedCurrencies := map[string]bool{
		"EUR": true,
		"USD": true,
		"CAD": true,
	}
	for range [100]int{} {
		result := RandomCurrency()
		if !expectedCurrencies[result] {
			t.Errorf("RandomCurrency() = %s; expected one of %v", result, expectedCurrencies)
		}
	}
}
