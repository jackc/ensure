package ensure

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gofrs/uuid/v5"
	"github.com/shopspring/decimal"
)

type FieldError struct {
	field string
	err   error
}

func (e *FieldError) Field() string {
	return e.field
}

func (e *FieldError) Unwrap() error {
	return e.err
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.field, e.err)
}

type RecordErrors struct {
	errors []*FieldError
}

// Add adds a new error to the validation errors for the given field. By convention, an empty string for field indicates
// a record-level error.
func (e *RecordErrors) Add(field string, err error) {
	e.errors = append(e.errors, &FieldError{field: field, err: err})
}

// Len returns the number of errors in the ValidationErrors.
func (e *RecordErrors) Len() int {
	if e == nil {
		return 0
	}

	return len(e.errors)
}

// On returns a []*ValidationError for the given field.
func (e *RecordErrors) On(field string) []*FieldError {
	if e == nil {
		return nil
	}

	var errs []*FieldError
	for _, e := range e.errors {
		if e.field == field {
			errs = append(errs, e)
		}
	}
	return errs
}

// All returns all errors.
func (e *RecordErrors) All() []*FieldError {
	if e == nil {
		return nil
	}

	return e.errors
}

// Unwrap unwraps all errors.
func (e *RecordErrors) Unwrap() []error {
	var errs []error
	for _, e := range e.errors {
		errs = append(errs, e)
	}

	return errs
}

// Error satisfies the error interface.
func (e *RecordErrors) Error() string {
	if len(e.errors) == 0 {
		return "BUG: Errors.Error() called with no errors"
	}

	sb := strings.Builder{}
	for i, e := range e.errors {
		if i > 0 {
			sb.WriteString(", ")
		}

		if e.field == "" {
			sb.WriteString(e.err.Error())
		} else {
			sb.WriteString(e.field)
			sb.WriteString(": ")
			sb.WriteString(e.err.Error())
		}
	}

	return sb.String()
}

type GetterSetter interface {
	Get(attribute string) any
	Set(attribute string, value any)
}

type GetterSetterMap map[string]any

func (m GetterSetterMap) Get(key string) any {
	return m[key]
}

func (m GetterSetterMap) Set(key string, value any) {
	m[key] = value
}

type RecordWithErrors struct {
	record GetterSetter
	errors *RecordErrors
}

func Record(record GetterSetter, fn EnsureRecordFunc) error {
	rwe := &RecordWithErrors{
		record: record,
	}

	fn(rwe)

	if errs := rwe.Errors(); errs.Len() > 0 {
		return errs
	}

	return nil
}

type RecordEnsurer struct {
	fn EnsureRecordFunc
}

func NewRecordEnsurer(fn EnsureRecordFunc) *RecordEnsurer {
	return &RecordEnsurer{
		fn: fn,
	}
}

func (re *RecordEnsurer) Ensure(value any) (any, error) {
	var record GetterSetter

	switch value := value.(type) {
	case GetterSetter:
		record = value
	case map[string]any:
		record = GetterSetterMap(value)
	default:
		return nil, errors.New("not a record")
	}

	err := Record(record, re.fn)
	if err != nil {
		return nil, err
	}

	return value, nil
}

type EnsureRecordFunc func(*RecordWithErrors)

func (r *RecordWithErrors) Add(field string, err error) {
	if r.errors == nil {
		r.errors = &RecordErrors{}
	}
	r.errors.Add(field, err)
}

func (r *RecordWithErrors) Get(field string) any {
	return r.record.Get(field)
}

func (r *RecordWithErrors) Set(field string, value any) {
	r.record.Set(field, value)
}

func (r *RecordWithErrors) Ensure(field string, ensurers ...Ensurer) {
	value := r.record.Get(field)
	for _, ensurer := range ensurers {
		var err error
		value, err = ensurer.Ensure(value)
		if err != nil {
			r.Add(field, err)
			return
		}
	}
	r.record.Set(field, value)
}

func (r *RecordWithErrors) Errors() *RecordErrors {
	return r.errors
}

type Ensurer interface {
	Ensure(any) (any, error)
}

type EnsurerFunc func(any) (any, error)

func (fn EnsurerFunc) Ensure(v any) (any, error) {
	return fn(v)
}

func convertInt64(value any) (int64, error) {
	switch value := value.(type) {
	case int8:
		return int64(value), nil
	case uint8:
		return int64(value), nil
	case int16:
		return int64(value), nil
	case uint16:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case uint32:
		return int64(value), nil
	case int64:
		return int64(value), nil
	case uint64:
		if value > math.MaxInt64 {
			return 0, errors.New("greater than maximum allowed number")
		}
		return int64(value), nil
	case int:
		if int64(value) < math.MinInt64 {
			return 0, errors.New("less than minimum allowed number")
		}
		if int64(value) > math.MaxInt64 {
			return 0, errors.New("greater than maximum allowed number")
		}
		return int64(value), nil
	case uint:
		if uint64(value) > math.MaxInt64 {
			return 0, errors.New("greater than maximum allowed number")
		}
		return int64(value), nil
	case float32:
		if value < math.MinInt64 {
			return 0, errors.New("less than minimum allowed number")
		}
		if value > math.MaxInt64 {
			return 0, errors.New("greater than maximum allowed number")
		}
		if float32(int64(value)) != value {
			return 0, errors.New("not a valid number")
		}
		return int64(value), nil
	case float64:
		if value < math.MinInt64 {
			return 0, errors.New("less than minimum allowed number")
		}
		if value > math.MaxInt64 {
			return 0, errors.New("greater than maximum allowed number")
		}
		if float64(int64(value)) != value {
			return 0, errors.New("not a valid number")
		}
		return int64(value), nil
	}

	s := fmt.Sprintf("%v", value)
	s = strings.TrimSpace(s)

	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.New("not a valid number")
	}
	return num, nil
}

// Int64 returns a Ensurer that converts value to an int64. If value is nil or a blank string nil is returned.
func Int64() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		n, err := convertInt64(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertInt32(value any) (int32, error) {
	n, err := convertInt64(value)
	if err != nil {
		return 0, err
	}

	if n < math.MinInt32 {
		return 0, errors.New("less than minimum allowed number")
	}
	if n > math.MaxInt32 {
		return 0, errors.New("greater than maximum allowed number")
	}

	return int32(n), nil
}

// Int32 returns a Ensurer that converts value to an int32. If value is nil or a blank string nil is returned.
func Int32() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		n, err := convertInt32(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertFloat64(value any) (float64, error) {
	switch value := value.(type) {
	case int8:
		return float64(value), nil
	case uint8:
		return float64(value), nil
	case int16:
		return float64(value), nil
	case uint16:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case uint32:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case uint64:
		return float64(value), nil
	case int:
		return float64(value), nil
	case uint:
		return float64(value), nil
	case float32:
		return float64(value), nil
	case float64:
		return value, nil
	}

	s := fmt.Sprintf("%v", value)
	s = strings.TrimSpace(s)

	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errors.New("not a valid number")
	}
	return num, nil
}

// Float64 returns a Ensurer that converts value to an float64. If value is nil or a blank string nil is returned.
func Float64() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return value, nil
		}

		n, err := convertFloat64(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertFloat32(value any) (float32, error) {
	n, err := convertFloat64(value)
	if err != nil {
		return 0, err
	}

	if n < -math.MaxFloat32 {
		return 0, errors.New("less than minimum allowed number")
	}
	if n > math.MaxFloat32 {
		return 0, errors.New("greater than maximum allowed number")
	}

	return float32(n), nil
}

// Float32 returns a Ensurer that converts value to an float32. If value is nil or a blank string nil is
// returned.
func Float32() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return value, nil
		}

		n, err := convertFloat32(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

// Bool returns a Ensurer that converts value to a bool. If value is nil or a blank string nil is returned.
func Bool() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		switch value := value.(type) {
		case bool:
			return value, nil
		case string:
			value = strings.TrimSpace(value)
			b, err := strconv.ParseBool(value)
			if err != nil {
				return nil, err
			}
			return b, nil
		default:
			return nil, errors.New("not a valid boolean")
		}
	})
}

// Time returns a Ensurer that converts value to a time.Time using formats. If value is nil or a blank string nil is returned.
func Time(formats ...string) Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		switch value := value.(type) {
		case time.Time:
			return value, nil
		case string:
			for _, format := range formats {
				t, err := time.Parse(format, value)
				if err == nil {
					return t, nil
				}
			}
		}

		return nil, errors.New("not a valid time")
	})
}

// UUID returns a Ensurer that converts value to a uuid.UUID. If value is nil or a blank string nil is returned.
func UUID() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		var uuidValue uuid.UUID
		var err error

		if value, ok := value.([]byte); ok {
			uuidValue, err = uuid.FromBytes(value)
			return uuidValue, err
		}

		s := fmt.Sprintf("%v", value)
		uuidValue, err = uuid.FromString(s)
		return uuidValue, err
	})
}

func convertDecimal(value any) (decimal.Decimal, error) {
	switch value := value.(type) {
	case decimal.Decimal:
		return value, nil
	case int64:
		return decimal.NewFromInt(value), nil
	case int:
		return decimal.NewFromInt(int64(value)), nil
	case int32:
		return decimal.NewFromInt32(value), nil
	case float32:
		return decimal.NewFromFloat32(value), nil
	case float64:
		return decimal.NewFromFloat(value), nil
	case string:
		value = strings.TrimSpace(value)
		return decimal.NewFromString(value)
	default:
		s := fmt.Sprintf("%v", value)
		s = strings.TrimSpace(s)
		return decimal.NewFromString(s)
	}
}

// Decimal returns a Ensurer that converts value to a decimal.Decimal. If value is nil or a blank string nil is
// returned.
func Decimal() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		n, err := convertDecimal(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertString(value any) string {
	switch value := value.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	}

	return fmt.Sprint(value)
}

// String returns a Ensurer that converts value to a string. If value is nil then nil is returned. It does not
// perform any normalization. In almost all cases, SingleLineString or MultiLineString should be used instead.
func String() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return value, nil
		}

		return convertString(value), nil
	})
}

type sliceElementError struct {
	Index int
	Err   error
}

type sliceElementErrors []sliceElementError

func (e sliceElementErrors) Error() string {
	sb := &strings.Builder{}
	for i, ee := range e {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(sb, "Element %d: %v", ee.Index, ee.Err)
	}
	return sb.String()
}

// Slice returns a Ensurer that converts value to a []T. value must be a []T or []any. If value is nil then nil
// is returned.
func Slice[T any](elementEnsurer Ensurer) Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		switch value := value.(type) {
		case []T:
			return value, nil
		case []any:
			ts := make([]T, len(value))
			var elErrs sliceElementErrors
			for i := range value {
				element, err := elementEnsurer.Ensure(value[i])
				if err != nil {
					elErrs = append(elErrs, sliceElementError{Index: i, Err: err})
				}
				if element, ok := element.(T); ok {
					ts[i] = element
				} else {
					var zero T
					elErrs = append(elErrs, sliceElementError{Index: i, Err: fmt.Errorf("not a %T", zero)})
				}
			}

			if elErrs != nil {
				return nil, elErrs
			}

			return ts, nil
		}

		return nil, fmt.Errorf("cannot convert to slice")
	})
}

// NotNil returns a Ensurer that fails if value is nil.
func NotNil() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return nil, errors.New("cannot be nil")
		}
		return value, nil
	})
}

// Require returns a Ensurer that returns an error if value is nil or "".
func Require() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		if value == nil || value == "" {
			return nil, fmt.Errorf("cannot be nil or empty")
		}

		return value, nil
	})
}

func convertSlice(value any, converters []Ensurer) (any, error) {
	v := value
	var err error

	for _, vc := range converters {
		v, err = vc.Ensure(v)
		if err != nil {
			break
		}
	}

	return v, err
}

func IfNotNil(converters ...Ensurer) Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return value, nil
		}

		return convertSlice(value, converters)
	})
}

// SingleLineString returns a Ensurer that converts a string value to a normalized string. If value is nil then nil is
// returned. If value is not a string then an error is returned.
//
// It performs the following operations:
//   - Remove any invalid UTF-8
//   - Replace non-printable characters with standard space
//   - Remove spaces from left and right
func SingleLineString() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		if s, ok := value.(string); ok {
			s = strings.ToValidUTF8(s, "")
			s = strings.Map(func(r rune) rune {
				if unicode.IsPrint(r) {
					return r
				} else {
					return ' '
				}
			}, s)
			s = strings.TrimSpace(s)

			return s, nil
		}

		return nil, errors.New("not a string")
	})
}

// MultiLineString returns a Ensurer that converts a string value to a normalized string. If value is nil then nil is
// returned. If value is not a string then an error is returned.
//
// It performs the following operations:
//   - Remove any invalid UTF-8
//   - Replace characters that are not graphic or space with standard space
func MultiLineString() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		if s, ok := value.(string); ok {
			s = strings.ToValidUTF8(s, "")
			s = strings.Map(func(r rune) rune {
				if unicode.IsGraphic(r) || unicode.IsSpace(r) {
					return r
				} else {
					return ' '
				}
			}, s)

			return s, nil
		}

		return nil, errors.New("not a string")
	})
}

// normalizeForParsing prepares value for parsing. If the value is not a string it is returned. Otherwise, space is
// trimmed from both sides of the string. If the string is now empty then nil is returned. Otherwise, the string is
// returned.
func normalizeForParsing(value any) any {
	if s, ok := value.(string); ok {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		return s
	}
	return value
}

// NilifyEmpty converts strings, slices, and maps where len(value) == 0 to nil. Any other value not modified.
func NilifyEmpty() Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		n, ok := tryLen(value)
		if ok && n == 0 {
			return nil, nil
		}
		return value, nil
	})
}

func requireStringTest(test func(string) bool, failErr error) Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		s, ok := value.(string)
		if !ok {
			return nil, errors.New("not a string")
		}

		if test(s) {
			return s, nil
		}

		return nil, failErr
	})
}

func tryLen(value any) (n int, ok bool) {
	s, ok := value.(string)
	if ok {
		return len(s), true
	}

	refval := reflect.ValueOf(value)
	switch refval.Kind() {
	case reflect.String, reflect.Slice, reflect.Map:
		return refval.Len(), true
	}

	return 0, false
}

// MinLen returns a Ensurer that fails if len(value) < min. value must be a string, slice, or map. nil is
// returned unmodified.
func MinLen(min int) Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryLen(value)
		if !ok {
			return nil, errors.New("not a string, slice or map")
		}

		if n < min {
			return nil, fmt.Errorf("too short")
		}

		return value, nil
	})
}

// MaxLen returns a Ensurer that fails if len(value) > max. value must be a string, slice, or map. nil is
// returned unmodified.
func MaxLen(max int) Ensurer {
	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryLen(value)
		if !ok {
			return nil, errors.New("not a string, slice or map")
		}

		if n > max {
			return nil, fmt.Errorf("too long")
		}

		return value, nil
	})
}

// AllowStrings returns a Ensurer that returns an error unless value is one of the allowedItems. If value is nil
// then nil is returned. If value is not a string then an error is returned.
func AllowStrings(allowedItems ...string) Ensurer {
	set := make(map[string]struct{}, len(allowedItems))
	for _, item := range allowedItems {
		set[item] = struct{}{}
	}

	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return value, nil
		}

		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("not allowed value")
		}

		if _, ok := set[s]; !ok {
			return nil, fmt.Errorf("not allowed value")
		}

		return value, nil
	})
}

// ExcludeStrings returns a Ensurer that returns an error if value is one of the excludedItems. If value is nil
// then nil is returned. If value is not a string then an error is returned.
func ExcludeStrings(excludedItems ...string) Ensurer {
	set := make(map[string]struct{}, len(excludedItems))
	for _, item := range excludedItems {
		set[item] = struct{}{}
	}

	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return value, nil
		}

		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("not allowed value")
		}

		if _, ok := set[s]; ok {
			return nil, fmt.Errorf("not allowed value")
		}

		return value, nil
	})
}

func tryDecimal(value any) (n decimal.Decimal, ok bool) {
	var strValue string
	switch value := value.(type) {
	case decimal.Decimal:
		return value, true
	case int32:
		return decimal.NewFromInt32(value), true
	case int64:
		return decimal.NewFromInt(value), true
	case int:
		return decimal.NewFromInt(int64(value)), true
	case float32:
		return decimal.NewFromFloat32(value), true
	case float64:
		return decimal.NewFromFloat(value), true
	case string:
		strValue = value
	default:
		strValue = fmt.Sprint(value)
	}

	n, err := decimal.NewFromString(strValue)
	if err != nil {
		return decimal.Zero, false
	}

	return n, true
}

// LessThan returns a Ensurer that fails unless value < x. x must be convertable to a decimal number or LessThan
// panics. value must be convertable to a decimal number. nil is returned unmodified.
func LessThan(x any) Ensurer {
	dx, ok := tryDecimal(x)
	if !ok {
		panic(fmt.Errorf("%v is not convertable to a decimal number", x))
	}

	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryDecimal(value)
		if !ok {
			return nil, fmt.Errorf("not a number")
		}

		if !n.LessThan(dx) {
			return nil, fmt.Errorf("too large")
		}

		return value, nil
	})
}

// LessThanOrEqual returns a Ensurer that fails unless value <= x. x must be convertable to a decimal number or
// LessThanOrEqual panics. value must be convertable to a decimal number. nil is returned unmodified.
func LessThanOrEqual(x any) Ensurer {
	dx, ok := tryDecimal(x)
	if !ok {
		panic(fmt.Errorf("%v is not convertable to a decimal number", x))
	}

	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryDecimal(value)
		if !ok {
			return nil, fmt.Errorf("not a number")
		}

		if !n.LessThanOrEqual(dx) {
			return nil, fmt.Errorf("too large")
		}

		return value, nil
	})
}

// GreaterThan returns a Ensurer that fails unless value > x. x must be convertable to a decimal number or
// GreaterThan panics. value must be convertable to a decimal number. nil is returned unmodified.
func GreaterThan(x any) Ensurer {
	dx, ok := tryDecimal(x)
	if !ok {
		panic(fmt.Errorf("%v is not convertable to a decimal number", x))
	}

	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryDecimal(value)
		if !ok {
			return nil, fmt.Errorf("not a number")
		}

		if !n.GreaterThan(dx) {
			return nil, fmt.Errorf("too small")
		}

		return value, nil
	})
}

// GreaterThanOrEqual returns a Ensurer that fails unless value >= x. x must be convertable to a decimal number
// or GreaterThanOrEqual panics. value must be convertable to a decimal number. nil is returned unmodified.
func GreaterThanOrEqual(x any) Ensurer {
	dx, ok := tryDecimal(x)
	if !ok {
		panic(fmt.Errorf("%v is not convertable to a decimal number", x))
	}

	return EnsurerFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryDecimal(value)
		if !ok {
			return nil, fmt.Errorf("not a number")
		}

		if !n.GreaterThanOrEqual(dx) {
			return nil, fmt.Errorf("too small")
		}

		return value, nil
	})
}
