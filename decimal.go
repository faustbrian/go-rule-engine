package ruleengine

import "strings"

// maxExponentDigits is an absolute construction guard independent of Limits:
// eighteen significant exponent digits always fit int64 after the coefficient
// adjustment, so hostile exponents cannot overflow before budgets are checked.
const maxExponentDigits = 18

// decimalValue stores the caller's lexical text and its normalized exact
// value: (-1)^negative * digits * 10^exponent. digits has no leading or
// trailing zeros and is empty for zero, which is never negative. Comparison
// uses the normalized form, so no arbitrary-precision arithmetic or float
// conversion is needed and the cost is linear in the significant digits.
type decimalValue struct {
	text     string
	negative bool
	digits   string
	exponent int64
}

// Decimal validates text against the RFC 8259 JSON number grammar and returns
// an exact decimal value that retains text for DecimalText and canonical JSON.
// Leading zeros, a leading plus, missing digits, whitespace, NaN and infinities
// are rejected; -0 is accepted and equals 0. Exponents with more than 18
// significant digits are rejected here; MaxDecimalBytes and MaxDecimalExponent
// are enforced by context construction, compilation and JSON parsing.
func Decimal(text string) (Value, error) {
	parsed, ok := parseDecimal(text)
	if !ok {
		return Value{}, newError(CodeInvalidFact, "decimal text is invalid")
	}
	return Value{kind: KindDecimal, data: parsed}, nil
}

// DecimalText returns the original lexical text and whether the kind matches.
func (v Value) DecimalText() (string, bool) {
	value, ok := v.data.(decimalValue)
	return value.text, ok && v.kind == KindDecimal
}

func parseDecimal(text string) (decimalValue, bool) {
	index := 0
	negative := false
	if index < len(text) && text[index] == '-' {
		negative = true
		index++
	}
	integerStart := index
	index = skipDigits(text, index)
	integer := text[integerStart:index]
	if integer == "" || (len(integer) > 1 && integer[0] == '0') {
		return decimalValue{}, false
	}
	fraction := ""
	if index < len(text) && text[index] == '.' {
		fractionStart := index + 1
		index = skipDigits(text, fractionStart)
		fraction = text[fractionStart:index]
		if fraction == "" {
			return decimalValue{}, false
		}
	}
	var exponent int64
	if index < len(text) && (text[index] == 'e' || text[index] == 'E') {
		index++
		exponentNegative := false
		if index < len(text) && (text[index] == '+' || text[index] == '-') {
			exponentNegative = text[index] == '-'
			index++
		}
		exponentStart := index
		index = skipDigits(text, exponentStart)
		digits := text[exponentStart:index]
		if digits == "" {
			return decimalValue{}, false
		}
		digits = strings.TrimLeft(digits, "0")
		if len(digits) > maxExponentDigits {
			return decimalValue{}, false
		}
		for _, digit := range []byte(digits) {
			exponent = exponent*10 + int64(digit-'0')
		}
		if exponentNegative {
			exponent = -exponent
		}
	}
	if index != len(text) {
		return decimalValue{}, false
	}
	coefficient := strings.TrimLeft(integer+fraction, "0")
	exponent -= int64(len(fraction))
	trimmed := strings.TrimRight(coefficient, "0")
	exponent += int64(len(coefficient) - len(trimmed))
	if trimmed == "" {
		return decimalValue{text: text}, true
	}
	return decimalValue{text: text, negative: negative, digits: trimmed, exponent: exponent}, true
}

func skipDigits(text string, index int) int {
	for index < len(text) && text[index] >= '0' && text[index] <= '9' {
		index++
	}
	return index
}

// adjustedExponent is the power of ten of the most significant digit; zero
// reports zero. It is the magnitude bounded by MaxDecimalExponent.
func (value decimalValue) adjustedExponent() int64 {
	if value.digits == "" {
		return 0
	}
	return value.exponent + int64(len(value.digits)) - 1
}

func (value decimalValue) sign() int {
	if value.digits == "" {
		return 0
	}
	if value.negative {
		return -1
	}
	return 1
}

func decimalsEqual(left, right decimalValue) bool {
	return left.negative == right.negative && left.digits == right.digits && left.exponent == right.exponent
}

func compareDecimals(left, right decimalValue) int {
	leftSign, rightSign := left.sign(), right.sign()
	if leftSign != rightSign {
		return compare(int64(leftSign), int64(rightSign))
	}
	if leftSign == 0 {
		return 0
	}
	magnitude := compare(left.adjustedExponent(), right.adjustedExponent())
	if magnitude == 0 {
		// Equal leading powers align the digit strings; without trailing zeros
		// a longer equal-prefix coefficient has an extra nonzero digit.
		magnitude = strings.Compare(left.digits, right.digits)
	}
	return magnitude * leftSign
}
