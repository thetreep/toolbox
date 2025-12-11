package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPointerToOrNil(t *testing.T) {
	t.Run("int with true condition", func(t *testing.T) {
		result := PointerToOrNil(42, true)
		assert.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})

	t.Run("int with false condition", func(t *testing.T) {
		result := PointerToOrNil(42, false)
		assert.Nil(t, result)
	})

	t.Run("zero value with true condition", func(t *testing.T) {
		result := PointerToOrNil(0, true)
		assert.NotNil(t, result)
		assert.Equal(t, 0, *result)
	})

	t.Run("string with true condition", func(t *testing.T) {
		result := PointerToOrNil("hello", true)
		assert.NotNil(t, result)
		assert.Equal(t, "hello", *result)
	})

	t.Run("empty string with false condition", func(t *testing.T) {
		result := PointerToOrNil("", false)
		assert.Nil(t, result)
	})

	t.Run("float64 with condition", func(t *testing.T) {
		result := PointerToOrNil(3.14, true)
		assert.NotNil(t, result)
		assert.Equal(t, 3.14, *result)

		resultNil := PointerToOrNil(3.14, false)
		assert.Nil(t, resultNil)
	})

	t.Run("struct with true condition", func(t *testing.T) {
		type testStruct struct {
			x int
			y string
		}
		val := testStruct{x: 42, y: "test"}
		result := PointerToOrNil(val, true)
		assert.NotNil(t, result)
		assert.Equal(t, val, *result)
	})
}

func TestPointerToIf(t *testing.T) {
	t.Run("predicate returns true", func(t *testing.T) {
		result := PointerToIf(42, func(v int) bool { return v > 0 })
		assert.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})

	t.Run("predicate returns false", func(t *testing.T) {
		result := PointerToIf(-5, func(v int) bool { return v > 0 })
		assert.Nil(t, result)
	})

	t.Run("complex predicate - latitude range check valid", func(t *testing.T) {
		isValidLatitude := func(v float64) bool {
			return v >= -90 && v <= 90
		}
		result := PointerToIf(45.5, isValidLatitude)
		assert.NotNil(t, result)
		assert.Equal(t, 45.5, *result)
	})

	t.Run("complex predicate - latitude out of range", func(t *testing.T) {
		isValidLatitude := func(v float64) bool {
			return v >= -90 && v <= 90
		}
		result := PointerToIf(100.0, isValidLatitude)
		assert.Nil(t, result)
	})

	t.Run("string length predicate", func(t *testing.T) {
		isNotEmpty := func(s string) bool { return len(s) > 0 }
		result := PointerToIf("hello", isNotEmpty)
		assert.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		resultEmpty := PointerToIf("", isNotEmpty)
		assert.Nil(t, resultEmpty)
	})

	t.Run("zero value with predicate that accepts it", func(t *testing.T) {
		alwaysTrue := func(v int) bool { return true }
		result := PointerToIf(0, alwaysTrue)
		assert.NotNil(t, result)
		assert.Equal(t, 0, *result)
	})
}

func TestPointerToIfNonZero(t *testing.T) {
	t.Run("non-zero int", func(t *testing.T) {
		result := PointerToIfNonZero(42)
		assert.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})

	t.Run("zero int", func(t *testing.T) {
		result := PointerToIfNonZero(0)
		assert.Nil(t, result)
	})

	t.Run("non-empty string", func(t *testing.T) {
		result := PointerToIfNonZero("hello")
		assert.NotNil(t, result)
		assert.Equal(t, "hello", *result)
	})

	t.Run("empty string", func(t *testing.T) {
		result := PointerToIfNonZero("")
		assert.Nil(t, result)
	})

	t.Run("non-zero float64", func(t *testing.T) {
		result := PointerToIfNonZero(3.14)
		assert.NotNil(t, result)
		assert.Equal(t, 3.14, *result)
	})

	t.Run("zero float64", func(t *testing.T) {
		result := PointerToIfNonZero(0.0)
		assert.Nil(t, result)
	})

	t.Run("negative int is non-zero", func(t *testing.T) {
		result := PointerToIfNonZero(-5)
		assert.NotNil(t, result)
		assert.Equal(t, -5, *result)
	})

	t.Run("negative float is non-zero", func(t *testing.T) {
		result := PointerToIfNonZero(-3.14)
		assert.NotNil(t, result)
		assert.Equal(t, -3.14, *result)
	})
}

// Example tests for documentation

func ExamplePointerToOrNil() {
	type GeoLocation struct {
		Latitude  float64
		Longitude float64
	}

	geo := &GeoLocation{Latitude: 48.8566, Longitude: 2.3522}

	// Using PointerToOrNil for conditional pointer creation
	latPtr := PointerToOrNil(geo.Latitude, geo != nil && geo.Latitude != 0)

	if latPtr != nil {
		println(*latPtr) // 48.8566
	}
}

func ExamplePointerToIf() {
	// Validate latitude is in valid range
	isValidLatitude := func(lat float64) bool {
		return lat >= -90 && lat <= 90
	}

	validLat := PointerToIf(48.8566, isValidLatitude) // returns pointer
	invalidLat := PointerToIf(200.0, isValidLatitude) // returns nil

	println(validLat != nil)   // true
	println(invalidLat != nil) // false
}

func ExamplePointerToIfNonZero() {
	// Simple non-zero check
	name := PointerToIfNonZero("John") // returns pointer to "John"
	empty := PointerToIfNonZero("")    // returns nil
	count := PointerToIfNonZero(5)     // returns pointer to 5
	zero := PointerToIfNonZero(0)      // returns nil

	println(name != nil)  // true
	println(empty != nil) // false
	println(count != nil) // true
	println(zero != nil)  // false
}
