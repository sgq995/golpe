package golpe_test

import (
	"os"

	"github.com/sgq995/golpe"
)

// Greeting is what the generator emits, line directives aside, for
//
//	template Greeting(name string, avatar element) {
//		<p class="greeting" title={name}>{avatar} Hello, {name}!</p>
//	}
func Greeting(name string, avatar golpe.Element) golpe.Element {
	return func(__golpeW golpe.Writer) error {
		if err := golpe.WriteString(__golpeW, `<p class="greeting" title=`); err != nil {
			return err
		}
		if err := golpe.WriteAttr(__golpeW, name); err != nil {
			return err
		}
		if err := golpe.WriteString(__golpeW, ">"); err != nil {
			return err
		}
		if err := golpe.Write(__golpeW, avatar); err != nil {
			return err
		}
		if err := golpe.WriteString(__golpeW, " Hello, "); err != nil {
			return err
		}
		if err := golpe.Write(__golpeW, name); err != nil {
			return err
		}
		if err := golpe.WriteString(__golpeW, "!</p>"); err != nil {
			return err
		}
		return nil
	}
}

// Templates are plain functions returning an Element; rendering one is a
// call with any io.Writer. Values are escaped, Elements are not.
func Example() {
	avatar := func(w golpe.Writer) error { return golpe.WriteString(w, `<img src="/me.png">`) }
	if err := Greeting(`Ada "<script>"`, avatar)(os.Stdout); err != nil {
		panic(err)
	}
	// Output: <p class="greeting" title="Ada &#34;&lt;script&gt;&#34;"><img src="/me.png"> Hello, Ada &#34;&lt;script&gt;&#34;!</p>
}

func ExampleWrite() {
	bold := func(w golpe.Writer) error { return golpe.WriteString(w, "<b>bold</b>") }
	for _, v := range []any{"<script>", " ", 42, " ", 2.5, " ", true, " ", bold, nil} {
		if err := golpe.Write(os.Stdout, v); err != nil {
			panic(err)
		}
	}
	// Output: &lt;script&gt; 42 2.5 true <b>bold</b>
}

// WriteAttr writes the quotes itself, so a value cannot break out of them.
func ExampleWriteAttr() {
	_ = golpe.WriteString(os.Stdout, "<a href=")
	_ = golpe.WriteAttr(os.Stdout, `"onmouseover=alert(1)`)
	_ = golpe.WriteString(os.Stdout, ">link</a>")
	// Output: <a href="&#34;onmouseover=alert(1)">link</a>
}
