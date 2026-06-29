package portal

import (
	"math"
	"testing"

	"pgregory.net/rapid"
)

// **Validates: Requirements 17.8**
// Property 27: Burn Rate Calculation
// The burn rate is always the average of the last 3 monthly expenditures
// (or fewer if < 3 months available), is always >= 0 when all expenditures
// are >= 0, equals the value when 3 equal months are provided, and returns 0
// for empty expenditures.

const burnRateEpsilon = 1e-9

func floatsAlmostEqual(a, b float64) bool {
	return math.Abs(a-b) < burnRateEpsilon
}

func TestProperty27_BurnRateIsAverageOfLastThreeMonths(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a slice of 3 to 24 months of non-negative expenditures
		length := rapid.IntRange(3, 24).Draw(t, "length")
		expenditures := make([]float64, length)
		for i := 0; i < length; i++ {
			expenditures[i] = rapid.Float64Range(0, 10_000_000).Draw(t, "expenditure")
		}

		burnRate := CalculateBurnRate(expenditures)

		// Property: burn rate is the average of the last 3 months
		last3Sum := expenditures[length-1] + expenditures[length-2] + expenditures[length-3]
		expected := last3Sum / 3.0

		if !floatsAlmostEqual(burnRate, expected) {
			t.Fatalf("burn rate mismatch: got %v, expected average of last 3 months = %v (values: %v, %v, %v)",
				burnRate, expected,
				expenditures[length-3], expenditures[length-2], expenditures[length-1])
		}
	})
}

func TestProperty27_BurnRateNonNegativeForNonNegativeExpenditures(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a slice of 1 to 24 months of non-negative expenditures
		length := rapid.IntRange(1, 24).Draw(t, "length")
		expenditures := make([]float64, length)
		for i := 0; i < length; i++ {
			expenditures[i] = rapid.Float64Range(0, 10_000_000).Draw(t, "expenditure")
		}

		burnRate := CalculateBurnRate(expenditures)

		// Property: burn rate is always >= 0 when all expenditures are >= 0
		if burnRate < 0 {
			t.Fatalf("burn rate is negative (%f) for non-negative expenditures: %v",
				burnRate, expenditures)
		}
	})
}

func TestProperty27_BurnRateEqualsValueForThreeEqualMonths(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a single non-negative value used for exactly 3 months
		value := rapid.Float64Range(0, 10_000_000).Draw(t, "value")
		expenditures := []float64{value, value, value}

		burnRate := CalculateBurnRate(expenditures)

		// Property: for exactly 3 months of equal expenditures, burn rate equals that value
		if !floatsAlmostEqual(burnRate, value) {
			t.Fatalf("burn rate should equal %v for three equal months, got %v",
				value, burnRate)
		}
	})
}

func TestProperty27_BurnRateEmptyExpendituresReturnsZero(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Empty expenditures should always return 0
		burnRate := CalculateBurnRate([]float64{})

		if burnRate != 0 {
			t.Fatalf("burn rate should be 0 for empty expenditures, got %f", burnRate)
		}
	})
}

func TestProperty27_BurnRateFewerThanThreeMonths(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate 1 or 2 months of non-negative expenditures
		length := rapid.IntRange(1, 2).Draw(t, "length")
		expenditures := make([]float64, length)
		for i := 0; i < length; i++ {
			expenditures[i] = rapid.Float64Range(0, 10_000_000).Draw(t, "expenditure")
		}

		burnRate := CalculateBurnRate(expenditures)

		// Property: when fewer than 3 months available, burn rate is the average of what's available
		sum := 0.0
		for _, v := range expenditures {
			sum += v
		}
		expected := sum / float64(length)

		if !floatsAlmostEqual(burnRate, expected) {
			t.Fatalf("burn rate mismatch for %d months: got %v, expected %v (expenditures: %v)",
				length, burnRate, expected, expenditures)
		}
	})
}
