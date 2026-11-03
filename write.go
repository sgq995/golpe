package golpe

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// WriteString writes s to w unchanged and returns the write error, if any.
// Generated code writes static HTML with it; the generator has already
// escaped those bytes.
func WriteString(w Writer, s string) error {
	_, err := io.WriteString(w, s)
	return err
}

// Write renders v, the value of a content interpolation {v}, to w
// (DEC-004):
//
//   - nil renders nothing.
//   - An Element is called with w, and what it writes is not escaped: it
//     escapes its own content. So is a value of a named type whose
//     underlying type is func(io.Writer) error. A nil one renders nothing.
//   - A string or a []byte is written HTML-escaped.
//   - Any other value is formatted as fmt.Sprint(v) formats it, then
//     HTML-escaped. So fmt's rules apply: a fmt.Formatter wins over an
//     error, which wins over a fmt.Stringer; a nil pointer prints as <nil>;
//     a panic in a String or Error method is printed, not propagated.
//
// Write returns the first error from w or from the Element, unchanged.
func Write(w Writer, v any) error {
	switch v := v.(type) {
	case nil:
		return nil
	case Element:
		if v == nil {
			return nil
		}
		return v(w)
	case string:
		return writeText(w, v)
	}
	if el, ok := asElement(v); ok {
		if el == nil {
			return nil
		}
		return el(w)
	}
	return writeText(w, text(v))
}

// WriteAttr renders v, the value of an attribute interpolation name={v}, to
// w as a double-quoted attribute value, quotes included: the generator writes
// only `name=` before it. v is formatted as Write formats it, nil being the
// empty string, and escaped the same way, so it cannot break out of the
// quotes.
//
// An Element renders markup, which an attribute value cannot hold, so
// WriteAttr returns an error for one, nil or not, without calling it or
// writing anything. Otherwise it returns the first error from w, unchanged.
func WriteAttr(w Writer, v any) error {
	var s string
	switch v := v.(type) {
	case nil:
	case string:
		s = v
	case Element:
		return attrElementError(v)
	default:
		if _, ok := asElement(v); ok {
			return attrElementError(v)
		}
		s = text(v)
	}
	return writeAttrValue(w, s)
}

func attrElementError(v any) error {
	return fmt.Errorf("golpe: cannot use %T as an attribute value: it is an Element, which renders markup", v)
}

var elementType = reflect.TypeOf(Element(nil))

// asElement converts v to an Element if the underlying type of v's dynamic
// type is func(io.Writer) error. Callers check `case Element:` first; this
// catches the named types, such as `type Icon func(io.Writer) error`, which
// that case does not match (Element aliases the unnamed func type) and fmt
// would print as a code address.
func asElement(v any) (Element, bool) {
	t := reflect.TypeOf(v)
	if t == nil || t.Kind() != reflect.Func || !t.ConvertibleTo(elementType) {
		return nil, false
	}
	return reflect.ValueOf(v).Convert(elementType).Interface().(Element), true
}

// text formats v, which is neither nil, an Element nor a string, as
// fmt.Sprint does, except that a []byte is text rather than a list of
// numbers. Two shortcuts yield fmt's bytes faster (LAB-008 in PLAN-002;
// TestWriteMatchesFmt pins them):
//
//   - Predeclared numeric and boolean types cannot have methods, so strconv
//     formats them.
//   - An error or a fmt.Stringer gets its Error or String method called
//     directly, in fmt's order, unless it is a fmt.Formatter (fmt prefers
//     Format) or a reflect.Value (fmt prints the value it holds). If the
//     method panics, as one with a nil pointer receiver may, v goes to fmt,
//     which calls it again and prints <nil> or the panic in its usual way.
func text(v any) string {
	switch v := v.(type) {
	case []byte:
		return string(v)
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case fmt.Formatter, reflect.Value:
		// Left to fmt.
	case error:
		if s, ok := call(v.Error); ok {
			return s
		}
	case fmt.Stringer:
		if s, ok := call(v.String); ok {
			return s
		}
	}
	return fmt.Sprint(v)
}

// call returns method() and true, or false if method panics.
func call(method func() string) (s string, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return method(), true
}
