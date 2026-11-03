package golpe_test

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/sgq995/golpe"
)

// el returns an Element that writes s verbatim.
func el(s string) golpe.Element {
	return func(w golpe.Writer) error { return golpe.WriteString(w, s) }
}

// icon is a named Element type. Passed as any, its values have dynamic type
// icon, which a `case golpe.Element:` does not match.
type icon func(io.Writer) error

// badge is a named Element type that is also a fmt.Stringer: being an
// Element wins.
type badge func(io.Writer) error

func (badge) String() string { return "<not rendered>" }

type stringer string

func (s stringer) String() string { return string(s) }

// ptrStringer's String method dereferences its receiver, so calling it on a
// nil *ptrStringer panics; fmt prints <nil> instead.
type ptrStringer struct{ s string }

func (p *ptrStringer) String() string { return p.s }

type textError string

func (e textError) Error() string { return string(e) }

// ptrError's Error method dereferences its receiver, like ptrStringer.
type ptrError struct{ msg string }

func (e *ptrError) Error() string { return e.msg }

// valueStringer has a value receiver: calling String through a nil
// *valueStringer panics in the compiler-generated wrapper.
type valueStringer struct{ s string }

func (v valueStringer) String() string { return v.s }

// both is an error and a fmt.Stringer; fmt prefers Error.
type both struct{}

func (both) Error() string  { return "<error>" }
func (both) String() string { return "<stringer>" }

// formatter is a fmt.Formatter and a fmt.Stringer; fmt prefers Format.
type formatter struct{}

func (formatter) Format(f fmt.State, verb rune) { fmt.Fprintf(f, "<%c>", verb) }
func (formatter) String() string                { return "<stringer>" }

type panicky struct{}

func (panicky) String() string { panic("<boom>") }

type celsius float64

func (c celsius) String() string { return strconv.FormatFloat(float64(c), 'f', 1, 64) + " <°C>" }

type (
	label string
	count int
	raw   []byte
)

func TestWriterIsAnAlias(t *testing.T) {
	// Generated closures are spelled func(__golpeW __golpe.Writer) error.
	// Such a closure is a golpe.Element (and a plain func(io.Writer) error)
	// only if Writer is an alias of io.Writer, not a type of its own.
	generated := func(w golpe.Writer) error { return nil }
	var _ golpe.Element = generated
	var _ func(io.Writer) error = generated
	var _ golpe.Writer = io.Writer(nil)
}

func TestWrite(t *testing.T) {
	var (
		nilElement golpe.Element
		nilIcon    icon
		nilPtr     *ptrStringer
	)
	tests := []struct {
		name string
		v    any
		want string
	}{
		{"nil", nil, ""},
		{"nil Element", nilElement, ""},
		{"Element", el("<b>bold</b>"), "<b>bold</b>"},
		{"func literal", func(w io.Writer) error { return golpe.WriteString(w, "<i>") }, "<i>"},
		{"named Element type", icon(el("<svg/>")), "<svg/>"},
		{"nil named Element type", nilIcon, ""},
		{"named Element type with a String method", badge(el("<i>new</i>")), "<i>new</i>"},
		{"string", "<script>", "&lt;script&gt;"},
		{"special characters", `& < > " '`, "&amp; &lt; &gt; &#34; &#39;"},
		{"empty string", "", ""},
		{"named string type", label("<b>"), "&lt;b&gt;"},
		{"Stringer", stringer("<em>"), "&lt;em&gt;"},
		{"pointer Stringer", &ptrStringer{"a&b"}, "a&amp;b"},
		{"nil pointer Stringer", nilPtr, "&lt;nil&gt;"},
		{"error", textError(`"quoted"`), "&#34;quoted&#34;"},
		{"error before Stringer", both{}, "&lt;error&gt;"},
		{"Formatter before Stringer", formatter{}, "&lt;v&gt;"},
		{"[]byte", []byte("<p>"), "&lt;p&gt;"},
		{"nil []byte", []byte(nil), ""},
		{"int", 42, "42"},
		{"negative int", -7, "-7"},
		{"int64", int64(math.MinInt64), "-9223372036854775808"},
		{"uint8", uint8(255), "255"},
		{"uint64", uint64(math.MaxUint64), "18446744073709551615"},
		{"float64", 3.25, "3.25"},
		{"float32", float32(0.1), "0.1"},
		{"large float", 1e21, "1e+21"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"named int type", count(3), "3"},
		{"struct", struct{ A, B string }{"<", ">"}, "{&lt; &gt;}"},
		{"slice", []string{"a", "<b>"}, "[a &lt;b&gt;]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := golpe.Write(&buf, tt.v); err != nil {
				t.Fatalf("Write(%#v): unexpected error %v", tt.v, err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("Write(%#v) wrote %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

// fmtValues are rendered by Write and WriteAttr exactly as fmt.Sprint formats
// them, then escaped. They cover every fast path (which must agree with fmt
// byte for byte) and fmt's own rules: Formatter, then error, then Stringer;
// <nil> for nil pointers; recovered panics; a reflect.Value prints the value
// it holds, not its String method.
var fmtValues = []any{
	reflect.ValueOf(42), reflect.ValueOf("<v>"), reflect.ValueOf(stringer("<s>")),
	(*ptrError)(nil), &ptrError{"<e>"}, valueStringer{"<vs>"}, (*valueStringer)(nil),
	0, 1, -1, 255, 256, math.MaxInt, math.MinInt,
	int8(math.MinInt8), int16(-1000), int32('x'), int64(1) << 53,
	uint(7), uint8(200), uint16(math.MaxUint16), uint32(math.MaxUint32), uint64(1) << 63, uintptr(9),
	0.0, math.Copysign(0, -1), 1.0, -2.5, 0.1, 1e6, 1e20, 1e21, 1e-4, 1e-5, 123456789.0,
	math.MaxFloat64, math.SmallestNonzeroFloat64, math.Inf(1), math.Inf(-1), math.NaN(),
	float32(1.1), float32(-0.3), float32(1e7), float32(math.MaxFloat32), float32(math.SmallestNonzeroFloat32),
	complex(1, -2), true, false,
	stringer("<x>"), textError("e&"), both{}, formatter{}, panicky{}, (*ptrStringer)(nil),
	celsius(21.5), label(`"l"`), count(-4), raw("<r>"),
	struct{}{}, []int{1, 2}, map[string]int{"<": 1},
}

func TestWriteMatchesFmt(t *testing.T) {
	for _, v := range fmtValues {
		want := html.EscapeString(fmt.Sprint(v))
		var buf bytes.Buffer
		if err := golpe.Write(&buf, v); err != nil {
			t.Errorf("Write(%T %v): unexpected error %v", v, v, err)
			continue
		}
		if got := buf.String(); got != want {
			t.Errorf("Write(%T %v) wrote %q, want %q", v, v, got, want)
		}
	}
}

func TestWriteAttr(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want string
	}{
		{"nil", nil, `""`},
		{"empty string", "", `""`},
		{"string", "/docs", `"/docs"`},
		// SPEC-002 acceptance: href={"\"onmouseover=x"} renders href="&#34;onmouseover=x".
		{"cannot break out of the quotes", `"onmouseover=x`, `"&#34;onmouseover=x"`},
		{"special characters", `& < > " '`, `"&amp; &lt; &gt; &#34; &#39;"`},
		{"single quote", "it's", `"it&#39;s"`},
		{"int", 42, `"42"`},
		{"bool", false, `"false"`},
		{"float", 0.5, `"0.5"`},
		{"Stringer", stringer("a<b"), `"a&lt;b"`},
		{"error", textError("x>y"), `"x&gt;y"`},
		{"[]byte", []byte(`"b"`), `"&#34;b&#34;"`},
		{"nil pointer", (*ptrStringer)(nil), `"&lt;nil&gt;"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := golpe.WriteAttr(&buf, tt.v); err != nil {
				t.Fatalf("WriteAttr(%#v): unexpected error %v", tt.v, err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("WriteAttr(%#v) wrote %s, want %s", tt.v, got, tt.want)
			}
		})
	}
}

func TestWriteAttrMatchesFmt(t *testing.T) {
	for _, v := range fmtValues {
		want := `"` + html.EscapeString(fmt.Sprint(v)) + `"`
		var buf bytes.Buffer
		if err := golpe.WriteAttr(&buf, v); err != nil {
			t.Errorf("WriteAttr(%T %v): unexpected error %v", v, v, err)
			continue
		}
		if got := buf.String(); got != want {
			t.Errorf("WriteAttr(%T %v) wrote %s, want %s", v, v, got, want)
		}
	}
}

// An Element renders markup, which an attribute value cannot hold, so
// WriteAttr rejects one (nil or not) with an error, without calling it and
// without writing anything.
func TestWriteAttrRejectsElements(t *testing.T) {
	called := false
	spy := func(w io.Writer) error { called = true; return golpe.WriteString(w, "<b>") }
	var nilElement golpe.Element
	values := []any{golpe.Element(spy), nilElement, icon(spy), icon(nil), badge(spy)}
	for _, v := range values {
		var buf bytes.Buffer
		err := golpe.WriteAttr(&buf, v)
		if err == nil {
			t.Errorf("WriteAttr(%T): no error, wrote %q", v, buf.String())
			continue
		}
		if !strings.Contains(err.Error(), "attribute") || !strings.Contains(err.Error(), fmt.Sprintf("%T", v)) {
			t.Errorf("WriteAttr(%T): error %q should name the type and say it cannot be an attribute value", v, err)
		}
		if buf.Len() != 0 {
			t.Errorf("WriteAttr(%T) wrote %q, want nothing", v, buf.String())
		}
	}
	if called {
		t.Error("WriteAttr called an Element")
	}
}

// An Element is called with the writer Write was given, so it can use the
// writer's other methods (http.Flusher, io.StringWriter, …).
func TestWritePassesTheWriterThrough(t *testing.T) {
	var buf bytes.Buffer
	var got io.Writer
	spy := func(w io.Writer) error { got = w; return nil }
	for _, v := range []any{golpe.Element(spy), icon(spy)} {
		got = nil
		if err := golpe.Write(&buf, v); err != nil {
			t.Fatal(err)
		}
		if got != io.Writer(&buf) {
			t.Errorf("Write(%T) called the Element with %#v, want the *bytes.Buffer it was given", v, got)
		}
	}
}

func TestWriteReturnsElementErrors(t *testing.T) {
	errElement := errors.New("element failed")
	failing := func(io.Writer) error { return errElement }
	for _, v := range []any{golpe.Element(failing), icon(failing)} {
		if err := golpe.Write(io.Discard, v); err != errElement {
			t.Errorf("Write(%T) = %v, want the Element's error %v", v, err, errElement)
		}
	}
}

var errWrite = errors.New("write failed")

// failingWriter lets writes 1…failAt-1 through and fails write failAt and
// every later one with errWrite. It counts every write attempted.
type failingWriter struct {
	failAt int
	writes int
	out    bytes.Buffer
}

func (f *failingWriter) Write(p []byte) (int, error) {
	f.writes++
	if f.writes >= f.failAt {
		return 0, errWrite
	}
	return f.out.Write(p)
}

// assertStopsAtFirstError renders r once into a writer that never fails, to
// count its writes, then once per write n with a writer failing at write n:
// r must return that writer's error, make no write after it, and have written
// a prefix of the full output.
func assertStopsAtFirstError(t *testing.T, name string, r golpe.Element) {
	t.Helper()
	full := &failingWriter{failAt: math.MaxInt}
	if err := r(full); err != nil {
		t.Fatalf("%s: unexpected error %v", name, err)
	}
	if full.writes == 0 {
		t.Fatalf("%s: wrote nothing", name)
	}
	for n := 1; n <= full.writes; n++ {
		fw := &failingWriter{failAt: n}
		if err := r(fw); err != errWrite {
			t.Errorf("%s, failing at write %d of %d: got error %v, want %v", name, n, full.writes, err, errWrite)
		}
		if fw.writes != n {
			t.Errorf("%s, failing at write %d of %d: %d writes attempted, want %d (rendering must stop at the first error)", name, n, full.writes, fw.writes, n)
		}
		if !strings.HasPrefix(full.out.String(), fw.out.String()) {
			t.Errorf("%s, failing at write %d: wrote %q, not a prefix of %q", name, n, fw.out.String(), full.out.String())
		}
	}
}

func TestWriterErrorsAreReturned(t *testing.T) {
	twoWrites := func(w io.Writer) error {
		if err := golpe.WriteString(w, "<b>"); err != nil {
			return err
		}
		return golpe.WriteString(w, "</b>")
	}
	calls := []struct {
		name string
		call golpe.Element
	}{
		{"WriteString", func(w golpe.Writer) error { return golpe.WriteString(w, "<p>") }},
		{"Write string", func(w golpe.Writer) error { return golpe.Write(w, "<p>") }},
		{"Write []byte", func(w golpe.Writer) error { return golpe.Write(w, []byte("p")) }},
		{"Write int", func(w golpe.Writer) error { return golpe.Write(w, 42) }},
		{"Write Stringer", func(w golpe.Writer) error { return golpe.Write(w, stringer("s")) }},
		{"Write Element", func(w golpe.Writer) error { return golpe.Write(w, golpe.Element(twoWrites)) }},
		{"Write named Element", func(w golpe.Writer) error { return golpe.Write(w, icon(twoWrites)) }},
		{"WriteAttr", func(w golpe.Writer) error { return golpe.WriteAttr(w, "a&b") }},
		{"WriteAttr empty", func(w golpe.Writer) error { return golpe.WriteAttr(w, "") }},
		{"WriteAttr int", func(w golpe.Writer) error { return golpe.WriteAttr(w, 7) }},
	}
	for _, c := range calls {
		assertStopsAtFirstError(t, c.name, c.call)
	}
}

// A template lowered the way SPEC-002 shows (every write and Element call
// checked) stops at the first failing write and returns its error, wherever
// in the tree that write happens.
func TestFirstWriteErrorStopsRendering(t *testing.T) {
	item := func(v any) golpe.Element {
		return func(w golpe.Writer) error {
			if err := golpe.WriteString(w, "<li>"); err != nil {
				return err
			}
			if err := golpe.Write(w, v); err != nil {
				return err
			}
			return golpe.WriteString(w, "</li>")
		}
	}
	list := func(title string, items ...any) golpe.Element {
		return func(w golpe.Writer) error {
			if err := golpe.WriteString(w, "<ul title="); err != nil {
				return err
			}
			if err := golpe.WriteAttr(w, title); err != nil {
				return err
			}
			if err := golpe.WriteString(w, ">"); err != nil {
				return err
			}
			for _, v := range items {
				if err := golpe.Write(w, item(v)); err != nil {
					return err
				}
			}
			if err := golpe.WriteString(w, "</ul>"); err != nil {
				return err
			}
			return nil
		}
	}
	page := list(`a"b`, "x<", 1, 2.5, el("<b>y</b>"), icon(el("<i>z</i>")), stringer("s"), []byte("q"), nil)
	assertStopsAtFirstError(t, "page", page)

	var buf bytes.Buffer
	if err := page(&buf); err != nil {
		t.Fatal(err)
	}
	const want = `<ul title="a&#34;b"><li>x&lt;</li><li>1</li><li>2.5</li><li><b>y</b></li><li><i>z</i></li><li>s</li><li>q</li><li></li></ul>`
	if got := buf.String(); got != want {
		t.Errorf("page rendered\n%s\nwant\n%s", got, want)
	}
}

// LAB-008 (PLAN-002, Laboratories): the cost of runtime dispatch (DEC-004).
//
// Each "dispatch" benchmark renders a value through golpe.Write or
// golpe.WriteAttr, as the generator emits it; its "direct" twin renders the
// same value with the code a type-directed generator would emit instead
// (io.WriteString of html.EscapeString or strconv output, or a direct Element
// call). Both write into a bytes.Buffer through a golpe.Writer, like
// generated code. Inputs are package variables so nothing is constant-folded.
//
//	go test -run '^$' -bench 'Write|Page' -benchmem
var (
	benchText     = "An ordinary product description."
	benchHTML     = `Tom & Jerry's "<b>Show</b>"`
	benchInt      = 31337
	benchFloat    = 19.99
	benchBool     = true
	benchStringer = stringer("in stock")
	benchElement  = el("<b>bold</b>")
)

func benchRender(b *testing.B, render golpe.Element) {
	b.Helper()
	var buf bytes.Buffer
	if err := render(&buf); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(buf.Len()))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := render(&buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	cases := []struct {
		name             string
		dispatch, direct golpe.Element
	}{
		{"string",
			func(w golpe.Writer) error { return golpe.Write(w, benchText) },
			func(w golpe.Writer) error { return golpe.WriteString(w, html.EscapeString(benchText)) }},
		{"escaped-string",
			func(w golpe.Writer) error { return golpe.Write(w, benchHTML) },
			func(w golpe.Writer) error { return golpe.WriteString(w, html.EscapeString(benchHTML)) }},
		{"int",
			func(w golpe.Writer) error { return golpe.Write(w, benchInt) },
			func(w golpe.Writer) error { return golpe.WriteString(w, strconv.Itoa(benchInt)) }},
		{"float",
			func(w golpe.Writer) error { return golpe.Write(w, benchFloat) },
			func(w golpe.Writer) error {
				return golpe.WriteString(w, strconv.FormatFloat(benchFloat, 'g', -1, 64))
			}},
		{"bool",
			func(w golpe.Writer) error { return golpe.Write(w, benchBool) },
			func(w golpe.Writer) error { return golpe.WriteString(w, strconv.FormatBool(benchBool)) }},
		{"Stringer",
			func(w golpe.Writer) error { return golpe.Write(w, benchStringer) },
			func(w golpe.Writer) error { return golpe.WriteString(w, html.EscapeString(benchStringer.String())) }},
		{"Element",
			func(w golpe.Writer) error { return golpe.Write(w, benchElement) },
			func(w golpe.Writer) error {
				if benchElement == nil {
					return nil
				}
				return benchElement(w)
			}},
		{"attr-string",
			func(w golpe.Writer) error { return golpe.WriteAttr(w, benchText) },
			func(w golpe.Writer) error { return directAttr(w, html.EscapeString(benchText)) }},
	}
	for _, c := range cases {
		b.Run(c.name+"/dispatch", func(b *testing.B) { benchRender(b, c.dispatch) })
		b.Run(c.name+"/direct", func(b *testing.B) { benchRender(b, c.direct) })
	}
}

// directAttr writes an already escaped attribute value with the same three
// writes WriteAttr makes, so the comparison isolates the dispatch.
func directAttr(w golpe.Writer, escaped string) error {
	if err := golpe.WriteString(w, `"`); err != nil {
		return err
	}
	if err := golpe.WriteString(w, escaped); err != nil {
		return err
	}
	return golpe.WriteString(w, `"`)
}

// The synthetic page: a product table, lowered twice by hand the way the
// generator would lower
//
//	template Row(p product) {
//		<tr class="row"><td>{p.ID}</td><td><a href={p.URL}>{p.Name}</a></td><td>{p.Price}</td><td>{p.Status}</td></tr>
//	}
//	template Page(title string, ps []product) {
//		<!doctype html><html><head><title>{title}</title></head><body><h1>{title}</h1>
//		<table><thead><tr><th>ID</th><th>Name</th><th>Price</th><th>Status</th></tr></thead>
//		<tbody>for _, p := range ps { {Row(p)} }</tbody></table></body></html>
//	}
//
// Page has 21 nodes (elements, text, interpolations, the for) and each row 12
// (6 elements, 4 text interpolations, 1 attribute interpolation and the
// {Row(p)} itself), so 82 rows make 21 + 82×12 = 1005 nodes. Every row
// interpolates an int, two strings (one as an attribute), a float64, a
// fmt.Stringer and an Element: an interpolation-heavy page.
type product struct {
	ID     int
	Name   string
	URL    string
	Price  float64
	Status stringer
}

const benchRows = 82

var benchTitle = "Products & prices"

var benchProducts = func() []product {
	ps := make([]product, benchRows)
	for i := range ps {
		id := 1000 + i
		ps[i] = product{
			ID:     id,
			Name:   "Widget " + strconv.Itoa(id),
			URL:    "/products/" + strconv.Itoa(id) + "?ref=list",
			Price:  9.99 + float64(i),
			Status: stringer("in stock"),
		}
		if i%10 == 0 {
			ps[i].Name = "Nuts & Bolts " + strconv.Itoa(id)
			ps[i].Status = "sold out"
		}
	}
	return ps
}()

const (
	pageHead = `<!doctype html><html><head><title>`
	pageH1   = `</title></head><body><h1>`
	pageThs  = `</h1><table><thead><tr><th>ID</th><th>Name</th><th>Price</th><th>Status</th></tr></thead><tbody>`
	pageEnd  = `</tbody></table></body></html>`
	rowStart = `<tr class="row"><td>`
	rowLink  = `</td><td><a href=`
	rowName  = `>`
	rowPrice = `</a></td><td>`
	rowCell  = `</td><td>`
	rowEnd   = `</td></tr>`
)

func dispatchRow(p product) golpe.Element {
	return func(w golpe.Writer) error {
		if err := golpe.WriteString(w, rowStart); err != nil {
			return err
		}
		if err := golpe.Write(w, p.ID); err != nil {
			return err
		}
		if err := golpe.WriteString(w, rowLink); err != nil {
			return err
		}
		if err := golpe.WriteAttr(w, p.URL); err != nil {
			return err
		}
		if err := golpe.WriteString(w, rowName); err != nil {
			return err
		}
		if err := golpe.Write(w, p.Name); err != nil {
			return err
		}
		if err := golpe.WriteString(w, rowPrice); err != nil {
			return err
		}
		if err := golpe.Write(w, p.Price); err != nil {
			return err
		}
		if err := golpe.WriteString(w, rowCell); err != nil {
			return err
		}
		if err := golpe.Write(w, p.Status); err != nil {
			return err
		}
		if err := golpe.WriteString(w, rowEnd); err != nil {
			return err
		}
		return nil
	}
}

func dispatchPage(title string, ps []product) golpe.Element {
	return func(w golpe.Writer) error {
		if err := golpe.WriteString(w, pageHead); err != nil {
			return err
		}
		if err := golpe.Write(w, title); err != nil {
			return err
		}
		if err := golpe.WriteString(w, pageH1); err != nil {
			return err
		}
		if err := golpe.Write(w, title); err != nil {
			return err
		}
		if err := golpe.WriteString(w, pageThs); err != nil {
			return err
		}
		for _, p := range ps {
			if err := golpe.Write(w, dispatchRow(p)); err != nil {
				return err
			}
		}
		if err := golpe.WriteString(w, pageEnd); err != nil {
			return err
		}
		return nil
	}
}

func directRow(p product) golpe.Element {
	return func(w golpe.Writer) error {
		if err := golpe.WriteString(w, rowStart); err != nil {
			return err
		}
		if err := golpe.WriteString(w, strconv.Itoa(p.ID)); err != nil {
			return err
		}
		if err := golpe.WriteString(w, rowLink); err != nil {
			return err
		}
		if err := directAttr(w, html.EscapeString(p.URL)); err != nil {
			return err
		}
		if err := golpe.WriteString(w, rowName); err != nil {
			return err
		}
		if err := golpe.WriteString(w, html.EscapeString(p.Name)); err != nil {
			return err
		}
		if err := golpe.WriteString(w, rowPrice); err != nil {
			return err
		}
		if err := golpe.WriteString(w, strconv.FormatFloat(p.Price, 'g', -1, 64)); err != nil {
			return err
		}
		if err := golpe.WriteString(w, rowCell); err != nil {
			return err
		}
		if err := golpe.WriteString(w, html.EscapeString(p.Status.String())); err != nil {
			return err
		}
		if err := golpe.WriteString(w, rowEnd); err != nil {
			return err
		}
		return nil
	}
}

func directPage(title string, ps []product) golpe.Element {
	return func(w golpe.Writer) error {
		if err := golpe.WriteString(w, pageHead); err != nil {
			return err
		}
		if err := golpe.WriteString(w, html.EscapeString(title)); err != nil {
			return err
		}
		if err := golpe.WriteString(w, pageH1); err != nil {
			return err
		}
		if err := golpe.WriteString(w, html.EscapeString(title)); err != nil {
			return err
		}
		if err := golpe.WriteString(w, pageThs); err != nil {
			return err
		}
		for _, p := range ps {
			if row := directRow(p); row != nil {
				if err := row(w); err != nil {
					return err
				}
			}
		}
		if err := golpe.WriteString(w, pageEnd); err != nil {
			return err
		}
		return nil
	}
}

// The two lowerings of the LAB-008 page must render the same bytes, or the
// benchmark compares different work.
func TestLab008PageLoweringsAgree(t *testing.T) {
	var dispatch, direct bytes.Buffer
	if err := dispatchPage(benchTitle, benchProducts)(&dispatch); err != nil {
		t.Fatal(err)
	}
	if err := directPage(benchTitle, benchProducts)(&direct); err != nil {
		t.Fatal(err)
	}
	if dispatch.String() != direct.String() {
		t.Fatalf("lowerings differ:\ndispatch %q\ndirect   %q", dispatch.String(), direct.String())
	}
	for _, want := range []string{
		`<title>Products &amp; prices</title>`,
		`<tr class="row"><td>1000</td><td><a href="/products/1000?ref=list">Nuts &amp; Bolts 1000</a></td><td>9.99</td><td>sold out</td></tr>`,
	} {
		if !strings.Contains(dispatch.String(), want) {
			t.Errorf("page does not contain %s", want)
		}
	}
	if got := strings.Count(dispatch.String(), "<tr class=\"row\">"); got != benchRows {
		t.Errorf("page has %d rows, want %d", got, benchRows)
	}
}

func BenchmarkPage(b *testing.B) {
	b.Run("dispatch", func(b *testing.B) { benchRender(b, dispatchPage(benchTitle, benchProducts)) })
	b.Run("direct", func(b *testing.B) { benchRender(b, directPage(benchTitle, benchProducts)) })
}
