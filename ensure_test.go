package ensure_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/jackc/ensure"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecord(t *testing.T) {
	record := ensure.GetterSetterMap{"age": "abc"}
	errs := ensure.Record(record, func(r *ensure.RecordWithErrors) {
		r.Ensure("age", ensure.Int64())
	})
	require.Error(t, errs)
}

func TestNotNil(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{"foo", "foo", true},
		{nil, nil, false},
	}

	for i, tt := range tests {
		value, err := ensure.NotNil().Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequire(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{"foo", "foo", true},
		{"", nil, false},
		{nil, nil, false},
	}

	for i, tt := range tests {
		value, err := ensure.Require().Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestInt64(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{1, int64(1), true},
		{"1", int64(1), true},
		{" 2 ", int64(2), true},
		{float32(12345678), int64(12345678), true},
		{float64(1234567890), int64(1234567890), true},
		{"10.5", nil, false},
		{"abc", nil, false},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
	}

	for i, tt := range tests {
		value, err := ensure.Int64().Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestFloat64(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{1, float64(1), true},
		{"1", float64(1), true},
		{" 2 ", float64(2), true},
		{"10.5", float64(10.5), true},
		{"abc", nil, false},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
	}

	for i, tt := range tests {
		value, err := ensure.Float64().Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestFloat32(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{1, float32(1), true},
		{"1", float32(1), true},
		{" 2 ", float32(2), true},
		{"10.5", float32(10.5), true},
		{"abc", nil, false},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
	}

	for i, tt := range tests {
		value, err := ensure.Float32().Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{true, true, true},
		{false, false, true},
		{"true", true, true},
		{"t", true, true},
		{"false", false, true},
		{"f", false, true},
		{" true ", true, true},
		{"abc", nil, false},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
	}

	for i, tt := range tests {
		value, err := ensure.Bool().Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestTime(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{"foo", nil, false},
		{"2023-06-24", time.Date(2023, 6, 24, 0, 0, 0, 0, time.UTC), true},
		{"2023-06-24 20:41:50", time.Date(2023, 6, 24, 20, 41, 50, 0, time.UTC), true},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
	}

	for i, tt := range tests {
		value, err := ensure.Time("2006-01-02", "2006-01-02 15:04:05").Ensure(tt.value)
		if tt.expected == nil {
			assert.Nilf(t, value, "%d", i)
		} else {
			expectedTime := tt.expected.(time.Time)
			valueTime, ok := value.(time.Time)
			assert.Truef(t, ok, "%d", i)
			assert.Truef(t, expectedTime.Equal(valueTime), "%d", i)
		}
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestDecimal(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{decimal.NewFromInt(1), decimal.NewFromInt(1), true},
		{1, decimal.NewFromInt(1), true},
		{"10.5", decimal.NewFromFloat(10.5), true},
		{" 7.7 ", decimal.NewFromFloat(7.7), true},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
		{"abc", nil, false},
	}

	for i, tt := range tests {
		value, err := ensure.Decimal().Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestSliceRecord(t *testing.T) {
	elementEnsurer := ensure.NewRecordEnsurer(func(record *ensure.RecordWithErrors) {
		record.Ensure("n", ensure.Int32(), ensure.Require())
	})

	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{
			value:    []any{map[string]any{"n": 1}, map[string]any{"n": 2}},
			expected: []map[string]any{{"n": int32(1)}, {"n": int32(2)}},
			success:  true,
		},
		{
			value:    []any{map[string]any{"n": 1}, map[string]any{"n": "abc"}},
			expected: nil,
			success:  false,
		},
		{value: nil, expected: nil, success: true},
		{[]int32{1, 2, 3}, nil, false},
		{[]any{"1", "2", "3"}, nil, false},
		{[]any{"1", 2, "3"}, nil, false},
		{"abc", nil, false},
		{42, nil, false},
	}

	for i, tt := range tests {
		value, err := ensure.Slice[map[string]any](elementEnsurer).Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d: %v", i, err)
	}
}

func TestSliceInt32(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{[]int32{1, 2, 3}, []int32{1, 2, 3}, true},
		{[]any{"1", "2", "3"}, []int32{1, 2, 3}, true},
		{[]any{"1", 2, "3"}, []int32{1, 2, 3}, true},
		{value: nil, expected: nil, success: true},
		{"abc", nil, false},
		{42, nil, false},
	}

	for i, tt := range tests {
		value, err := ensure.Slice[int32](ensure.Int32()).Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestSliceString(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{[]string{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}, true},
		{[]any{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}, true},
		{value: nil, expected: nil, success: true},
		{"abc", nil, false},
	}

	for i, tt := range tests {
		value, err := ensure.Slice[string](ensure.SingleLineString()).Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestSingleLineString(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
		msg      string
	}{
		{value: "a", expected: "a", success: true, msg: "no changes"},
		{value: " a", expected: "a", success: true, msg: "trim left"},
		{value: "a ", expected: "a", success: true, msg: "trim right"},
		{value: " a ", expected: "a", success: true, msg: "trim both sides"},
		{value: "a\xfe\xffa", expected: "aa", success: true, msg: "invalid UTF-8"},
		{value: "a\u200Ba", expected: "a a", success: true, msg: "replace non-normal spaces"},
		{value: "a\ta", expected: "a a", success: true, msg: "replace control character"},
		{value: "a\r\n", expected: "a", success: true, msg: "trim happens after replaced control character"},
		{value: nil, expected: nil, success: true},
	}

	for i, tt := range tests {
		value, err := ensure.SingleLineString().Ensure(tt.value)
		assert.Equalf(t, tt.success, err == nil, "%d: %s", i, tt.msg)
		assert.Equalf(t, tt.expected, value, "%d: %s", i, tt.msg)
	}
}

func TestNilifyEmpty(t *testing.T) {
	type otherString string

	tests := []struct {
		value    any
		expected any
	}{
		{"foo", "foo"},
		{"", nil},
		{otherString(""), nil},
		{[]int{}, nil},
		{[]int{1}, []int{1}},
		{map[string]any{}, nil},
		{map[string]any{"foo": "bar"}, map[string]any{"foo": "bar"}},
		{nil, nil},
	}

	for i, tt := range tests {
		value, err := ensure.NilifyEmpty().Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.NoErrorf(t, err, "%d", i)
	}
}

func TestMinLen(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		length     int
		errMatcher *regexp.Regexp
	}{
		{"foo", "foo", 1, nil},
		{"f", "f", 1, nil},
		{"", nil, 1, regexp.MustCompile(`short`)},
		{1, nil, 1, regexp.MustCompile(`not a string`)},
		{[]int{1, 2, 3}, []int{1, 2, 3}, 1, nil},
		{[]int{}, nil, 1, regexp.MustCompile(`short`)},
		{map[string]any{}, nil, 1, regexp.MustCompile(`short`)},
		{map[string]any{"foo": "bar"}, map[string]any{"foo": "bar"}, 1, nil},
		{nil, nil, 1, nil},
	}

	for i, tt := range tests {
		value, err := ensure.MinLen(tt.length).Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			require.NoError(t, err, "%d", i)
		} else {
			require.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestMaxLen(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		length     int
		errMatcher *regexp.Regexp
	}{
		{"foo", "foo", 3, nil},
		{"f", "f", 3, nil},
		{"", "", 3, nil},
		{"abcd", nil, 3, regexp.MustCompile(`long`)},
		{1, nil, 3, regexp.MustCompile(`not a string`)},
		{[]int{1, 2, 3}, []int{1, 2, 3}, 3, nil},
		{[]int{1, 2, 3, 4}, nil, 3, regexp.MustCompile(`long`)},
		{map[string]any{"foo": "bar"}, map[string]any{"foo": "bar"}, 2, nil},
		{map[string]any{"foo": "bar", "baz": "quz"}, nil, 1, regexp.MustCompile(`long`)},
		{nil, nil, 1, nil},
	}

	for i, tt := range tests {
		value, err := ensure.MaxLen(tt.length).Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			require.NoError(t, err, "%d", i)
		} else {
			require.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestAllowStrings(t *testing.T) {
	tests := []struct {
		value         any
		allowedValues []string
		errMatcher    *regexp.Regexp
	}{
		{
			value:         "foo",
			allowedValues: []string{"foo", "bar"},
			errMatcher:    nil,
		},
		{
			value:         "quz",
			allowedValues: []string{"foo", "bar"},
			errMatcher:    regexp.MustCompile(`not allowed value`),
		},
	}

	for i, tt := range tests {
		value, err := ensure.AllowStrings(tt.allowedValues...).Ensure(tt.value)
		if tt.errMatcher == nil {
			assert.Equalf(t, tt.value, value, "%d", i)
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestExcludeStrings(t *testing.T) {
	tests := []struct {
		value          any
		excludedValues []string
		errMatcher     *regexp.Regexp
	}{
		{
			value:          "foo",
			excludedValues: []string{"foo", "bar"},
			errMatcher:     regexp.MustCompile(`not allowed value`),
		},
		{
			value:          "quz",
			excludedValues: []string{"foo", "bar"},
			errMatcher:     nil,
		},
	}

	for i, tt := range tests {
		value, err := ensure.ExcludeStrings(tt.excludedValues...).Ensure(tt.value)
		if tt.errMatcher == nil {
			assert.Equalf(t, tt.value, value, "%d", i)
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestLessThan(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		limit      any
		errMatcher *regexp.Regexp
	}{
		{decimal.NewFromInt(1), decimal.NewFromInt(1), decimal.NewFromInt(10), nil},
		{decimal.NewFromInt(10), nil, decimal.NewFromInt(10), regexp.MustCompile(`too large`)},
		{10, nil, 10, regexp.MustCompile(`too large`)},
		{32.5, nil, 10, regexp.MustCompile(`too large`)},
		{"11", nil, 10, regexp.MustCompile(`too large`)},
		{nil, nil, decimal.NewFromInt(10), nil},
	}

	for i, tt := range tests {
		value, err := ensure.LessThan(tt.limit).Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestLessThanOrEqual(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		limit      any
		errMatcher *regexp.Regexp
	}{
		{decimal.NewFromInt(1), decimal.NewFromInt(1), decimal.NewFromInt(10), nil},
		{decimal.NewFromInt(10), decimal.NewFromInt(10), decimal.NewFromInt(10), nil},
		{decimal.NewFromInt(11), nil, decimal.NewFromInt(10), regexp.MustCompile(`too large`)},
		{10, 10, 10, nil},
		{32.5, nil, 10, regexp.MustCompile(`too large`)},
		{"11", nil, 10, regexp.MustCompile(`too large`)},
		{nil, nil, decimal.NewFromInt(10), nil},
	}

	for i, tt := range tests {
		value, err := ensure.LessThanOrEqual(tt.limit).Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestGreaterThan(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		limit      any
		errMatcher *regexp.Regexp
	}{
		{decimal.NewFromInt(1), nil, decimal.NewFromInt(10), regexp.MustCompile(`too small`)},
		{decimal.NewFromInt(10), nil, decimal.NewFromInt(10), regexp.MustCompile(`too small`)},
		{decimal.NewFromInt(11), decimal.NewFromInt(11), decimal.NewFromInt(10), nil},
		{10, nil, 10, regexp.MustCompile(`too small`)},
		{32.5, 32.5, 10, nil},
		{"11", "11", 10, nil},
		{nil, nil, decimal.NewFromInt(10), nil},
	}

	for i, tt := range tests {
		value, err := ensure.GreaterThan(tt.limit).Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestGreaterThanOrEqual(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		limit      any
		errMatcher *regexp.Regexp
	}{
		{decimal.NewFromInt(1), nil, decimal.NewFromInt(10), regexp.MustCompile(`too small`)},
		{decimal.NewFromInt(10), decimal.NewFromInt(10), decimal.NewFromInt(10), nil},
		{decimal.NewFromInt(11), decimal.NewFromInt(11), decimal.NewFromInt(10), nil},
		{10, 10, 10, nil},
		{32.5, 32.5, 10, nil},
		{"11", "11", 10, nil},
		{nil, nil, decimal.NewFromInt(10), nil},
	}

	for i, tt := range tests {
		value, err := ensure.GreaterThanOrEqual(tt.limit).Ensure(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func BenchmarkRecordEnsurerEnsure(b *testing.B) {
	recordEnsurer := ensure.NewRecordEnsurer(func(record *ensure.RecordWithErrors) {
		record.Ensure("name", ensure.SingleLineString(), ensure.Require())
		record.Ensure("age", ensure.Int32(), ensure.GreaterThanOrEqual(0), ensure.LessThanOrEqual(125))
		record.Ensure("weight", ensure.Float32(), ensure.GreaterThanOrEqual(0), ensure.LessThanOrEqual(1000))
	})

	for i := 0; i < b.N; i++ {
		record := map[string]any{"name": "Adam", "age": "30", "weight": "80.5"}
		_, err := recordEnsurer.Ensure(record)
		if err != nil {
			b.Fatal(err)
		}
		if record["name"] != "Adam" {
			b.Fatal("name should not be changed")
		}
		if record["age"] != int32(30) {
			b.Fatal("age should have been parsed to int32(30)")
		}
		if record["weight"] != float32(80.5) {
			b.Fatal("weight should have been parsed to float32(80.5)")
		}
	}
}
