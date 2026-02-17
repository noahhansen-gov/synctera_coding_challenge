package api_test

import (
	"testing"

	"github.com/synctera/tech-challenge/internal/api"
)

// --- ParseIntOrDefault ---

// Test: TestParseIntOrDefault_emptyStringReturnsDefault
// What: ParseIntOrDefault returns the caller-supplied default when given an empty string
// Input: s="", defaultVal=42
// Output: 42
func TestParseIntOrDefault_emptyStringReturnsDefault(t *testing.T) {
	if got := api.ParseIntOrDefault("", 42); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

// Test: TestParseIntOrDefault_validIntegerParsed
// What: ParseIntOrDefault parses a valid integer string and returns it
// Input: s="100", defaultVal=0
// Output: 100
func TestParseIntOrDefault_validIntegerParsed(t *testing.T) {
	if got := api.ParseIntOrDefault("100", 0); got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

// Test: TestParseIntOrDefault_invalidStringReturnsDefault
// What: ParseIntOrDefault returns the default when the string cannot be parsed as an integer
// Input: s="abc", defaultVal=5
// Output: 5
func TestParseIntOrDefault_invalidStringReturnsDefault(t *testing.T) {
	if got := api.ParseIntOrDefault("abc", 5); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
}

// Test: TestParseIntOrDefault_negativeIntegerParsed
// What: ParseIntOrDefault correctly handles negative integer strings
// Input: s="-10", defaultVal=0
// Output: -10
func TestParseIntOrDefault_negativeIntegerParsed(t *testing.T) {
	if got := api.ParseIntOrDefault("-10", 0); got != -10 {
		t.Errorf("expected -10, got %d", got)
	}
}

// --- ParseDateOrNil ---

// Test: TestParseDateOrNil_emptyStringReturnsNil
// What: ParseDateOrNil treats an empty string as "no filter provided" and returns nil with no error
// Input: dateStr=""
// Output: nil pointer, nil error
func TestParseDateOrNil_emptyStringReturnsNil(t *testing.T) {
	result, err := api.ParseDateOrNil("")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

// Test: TestParseDateOrNil_validDateParsed
// What: ParseDateOrNil parses a YYYY-MM-DD string into a *time.Time
// Input: dateStr="2024-01-15"
// Output: pointer to time.Time{year:2024, month:1, day:15}, nil error
func TestParseDateOrNil_validDateParsed(t *testing.T) {
	result, err := api.ParseDateOrNil("2024-01-15")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Year() != 2024 || result.Month() != 1 || result.Day() != 15 {
		t.Errorf("expected 2024-01-15, got %v", result)
	}
}

// Test: TestParseDateOrNil_invalidDateReturnsError
// What: ParseDateOrNil returns an error for completely non-date strings
// Input: dateStr="not-a-date"
// Output: nil pointer, non-nil error
func TestParseDateOrNil_invalidDateReturnsError(t *testing.T) {
	_, err := api.ParseDateOrNil("not-a-date")
	if err == nil {
		t.Error("expected error for invalid date, got nil")
	}
}

// Test: TestParseDateOrNil_wrongFormatReturnsError
// What: ParseDateOrNil only accepts YYYY-MM-DD; other formats (MM/DD/YYYY) are rejected
// Input: dateStr="01/15/2024"
// Output: nil pointer, non-nil error
func TestParseDateOrNil_wrongFormatReturnsError(t *testing.T) {
	_, err := api.ParseDateOrNil("01/15/2024")
	if err == nil {
		t.Error("expected error for wrong date format, got nil")
	}
}

// --- ParseAndValidateDateFilters ---

// Test: TestParseAndValidateDateFilters_noFilters
// What: ParseAndValidateDateFilters returns nil,nil,nil when both inputs are empty strings
// Input: startDateStr="", endDateStr=""
// Output: nil start, nil end, nil error
func TestParseAndValidateDateFilters_noFilters(t *testing.T) {
	start, end, err := api.ParseAndValidateDateFilters("", "")
	if err != nil || start != nil || end != nil {
		t.Errorf("expected nil,nil,nil - got %v,%v,%v", start, end, err)
	}
}

// Test: TestParseAndValidateDateFilters_startOnly
// What: ParseAndValidateDateFilters handles a start date with no end date
// Input: startDateStr="2024-01-01", endDateStr=""
// Output: non-nil start, nil end, nil error
func TestParseAndValidateDateFilters_startOnly(t *testing.T) {
	start, end, err := api.ParseAndValidateDateFilters("2024-01-01", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start == nil {
		t.Fatal("expected non-nil start")
	}
	if end != nil {
		t.Errorf("expected nil end, got %v", end)
	}
}

// Test: TestParseAndValidateDateFilters_endOnly
// What: ParseAndValidateDateFilters handles an end date with no start date
// Input: startDateStr="", endDateStr="2024-06-30"
// Output: nil start, non-nil end, nil error
func TestParseAndValidateDateFilters_endOnly(t *testing.T) {
	start, end, err := api.ParseAndValidateDateFilters("", "2024-06-30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != nil {
		t.Errorf("expected nil start, got %v", start)
	}
	if end == nil {
		t.Fatal("expected non-nil end")
	}
}

// Test: TestParseAndValidateDateFilters_validRange
// What: ParseAndValidateDateFilters accepts a range where start < end
// Input: startDateStr="2024-01-01", endDateStr="2024-12-31"
// Output: nil error
func TestParseAndValidateDateFilters_validRange(t *testing.T) {
	_, _, err := api.ParseAndValidateDateFilters("2024-01-01", "2024-12-31")
	if err != nil {
		t.Errorf("expected nil error for valid range, got %v", err)
	}
}

// Test: TestParseAndValidateDateFilters_equalDates
// What: ParseAndValidateDateFilters accepts start == end (single-day range)
// Input: startDateStr="2024-06-01", endDateStr="2024-06-01"
// Output: nil error
func TestParseAndValidateDateFilters_equalDates(t *testing.T) {
	_, _, err := api.ParseAndValidateDateFilters("2024-06-01", "2024-06-01")
	if err != nil {
		t.Errorf("expected nil error for equal dates, got %v", err)
	}
}

// Test: TestParseAndValidateDateFilters_startAfterEndReturnsError
// What: ParseAndValidateDateFilters rejects a range where start > end
// Input: startDateStr="2024-12-31", endDateStr="2024-01-01"
// Output: non-nil error
func TestParseAndValidateDateFilters_startAfterEndReturnsError(t *testing.T) {
	_, _, err := api.ParseAndValidateDateFilters("2024-12-31", "2024-01-01")
	if err == nil {
		t.Error("expected error when start > end, got nil")
	}
}

// Test: TestParseAndValidateDateFilters_invalidStartReturnsError
// What: ParseAndValidateDateFilters returns an error when the start date string is not parseable
// Input: startDateStr="bad-date", endDateStr="2024-12-31"
// Output: non-nil error
func TestParseAndValidateDateFilters_invalidStartReturnsError(t *testing.T) {
	_, _, err := api.ParseAndValidateDateFilters("bad-date", "2024-12-31")
	if err == nil {
		t.Error("expected error for invalid start date, got nil")
	}
}

// Test: TestParseAndValidateDateFilters_invalidEndReturnsError
// What: ParseAndValidateDateFilters returns an error when the end date string is not parseable
// Input: startDateStr="2024-01-01", endDateStr="bad-date"
// Output: non-nil error
func TestParseAndValidateDateFilters_invalidEndReturnsError(t *testing.T) {
	_, _, err := api.ParseAndValidateDateFilters("2024-01-01", "bad-date")
	if err == nil {
		t.Error("expected error for invalid end date, got nil")
	}
}

// --- ParseAndValidateAmountFilters ---

// Test: TestParseAndValidateAmountFilters_noFilters
// What: ParseAndValidateAmountFilters returns nil,nil,nil when both inputs are empty
// Input: minAmountStr="", maxAmountStr=""
// Output: nil min, nil max, nil error
func TestParseAndValidateAmountFilters_noFilters(t *testing.T) {
	min, max, err := api.ParseAndValidateAmountFilters("", "")
	if err != nil || min != nil || max != nil {
		t.Errorf("expected nil,nil,nil - got %v,%v,%v", min, max, err)
	}
}

// Test: TestParseAndValidateAmountFilters_minOnly
// What: ParseAndValidateAmountFilters handles min amount with no max
// Input: minAmountStr="100", maxAmountStr=""
// Output: min=100, nil max, nil error
func TestParseAndValidateAmountFilters_minOnly(t *testing.T) {
	min, max, err := api.ParseAndValidateAmountFilters("100", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if min == nil || *min != 100 {
		t.Errorf("expected min=100, got %v", min)
	}
	if max != nil {
		t.Errorf("expected nil max, got %v", max)
	}
}

// Test: TestParseAndValidateAmountFilters_maxOnly
// What: ParseAndValidateAmountFilters handles max amount with no min
// Input: minAmountStr="", maxAmountStr="500"
// Output: nil min, max=500, nil error
func TestParseAndValidateAmountFilters_maxOnly(t *testing.T) {
	min, max, err := api.ParseAndValidateAmountFilters("", "500")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if min != nil {
		t.Errorf("expected nil min, got %v", min)
	}
	if max == nil || *max != 500 {
		t.Errorf("expected max=500, got %v", max)
	}
}

// Test: TestParseAndValidateAmountFilters_validRange
// What: ParseAndValidateAmountFilters accepts a range where min < max
// Input: minAmountStr="100", maxAmountStr="500"
// Output: nil error
func TestParseAndValidateAmountFilters_validRange(t *testing.T) {
	_, _, err := api.ParseAndValidateAmountFilters("100", "500")
	if err != nil {
		t.Errorf("expected nil for valid range, got %v", err)
	}
}

// Test: TestParseAndValidateAmountFilters_equalAmounts
// What: ParseAndValidateAmountFilters accepts min == max (exact amount filter)
// Input: minAmountStr="100", maxAmountStr="100"
// Output: nil error
func TestParseAndValidateAmountFilters_equalAmounts(t *testing.T) {
	_, _, err := api.ParseAndValidateAmountFilters("100", "100")
	if err != nil {
		t.Errorf("expected nil for equal min/max, got %v", err)
	}
}

// Test: TestParseAndValidateAmountFilters_minGreaterThanMaxReturnsError
// What: ParseAndValidateAmountFilters rejects a range where min > max
// Input: minAmountStr="500", maxAmountStr="100"
// Output: non-nil error
func TestParseAndValidateAmountFilters_minGreaterThanMaxReturnsError(t *testing.T) {
	_, _, err := api.ParseAndValidateAmountFilters("500", "100")
	if err == nil {
		t.Error("expected error when min > max, got nil")
	}
}

// Test: TestParseAndValidateAmountFilters_invalidMinReturnsError
// What: ParseAndValidateAmountFilters returns an error when min_amount is not numeric
// Input: minAmountStr="abc", maxAmountStr="100"
// Output: non-nil error
func TestParseAndValidateAmountFilters_invalidMinReturnsError(t *testing.T) {
	_, _, err := api.ParseAndValidateAmountFilters("abc", "100")
	if err == nil {
		t.Error("expected error for invalid min_amount, got nil")
	}
}

// Test: TestParseAndValidateAmountFilters_invalidMaxReturnsError
// What: ParseAndValidateAmountFilters returns an error when max_amount is not numeric
// Input: minAmountStr="100", maxAmountStr="xyz"
// Output: non-nil error
func TestParseAndValidateAmountFilters_invalidMaxReturnsError(t *testing.T) {
	_, _, err := api.ParseAndValidateAmountFilters("100", "xyz")
	if err == nil {
		t.Error("expected error for invalid max_amount, got nil")
	}
}
