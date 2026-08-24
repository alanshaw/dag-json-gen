package typegen

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// corpus of strings exercising every escape path
func escapeCorpus() []string {
	corpus := []string{
		"",
		"hello",
		`with "quotes" and \backslashes\`,
		"tabs\tnewlines\nreturns\rbell\bformfeed\f",
		"html <script>alert('&')</script>",
		"unicode: 你好 привет 🎸 é",
		"line seps: \u2028 and \u2029",
		"del char: \x7f",
		"invalid utf8: \xff",
		"truncated rune: a\xc3",
		"lone continuation: \x80\x81",
		"mixed \xf0\x9f valid 🎸 invalid \xff end",
	}
	// every single ASCII byte as a 1-char string
	for c := 0; c < 0x80; c++ {
		corpus = append(corpus, string(rune(c)))
	}
	return corpus
}

func TestWriteStringDifferential(t *testing.T) {
	var buf bytes.Buffer
	w := NewDagJsonWriter(&buf)
	check := func(t *testing.T, s string) {
		t.Helper()
		buf.Reset()
		if err := w.WriteString(s); err != nil {
			t.Fatalf("WriteString(%q): %v", s, err)
		}
		want, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("json.Marshal(%q): %v", s, err)
		}
		if !bytes.Equal(buf.Bytes(), want) {
			t.Errorf("WriteString mismatch for %q:\n got: %s\nwant: %s", s, buf.Bytes(), want)
		}
	}
	for _, s := range escapeCorpus() {
		check(t, s)
	}
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 10000; i++ {
		val, ok := quick.Value(reflect.TypeFor[string](), r)
		if !ok {
			t.Fatal("failed to generate string")
		}
		check(t, val.String())
	}
}

func TestUnescapeJSONStringRoundTrip(t *testing.T) {
	check := func(t *testing.T, s string) {
		t.Helper()
		quoted, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("json.Marshal(%q): %v", s, err)
		}
		var want string
		if err := json.Unmarshal(quoted, &want); err != nil {
			t.Fatalf("json.Unmarshal(%s): %v", quoted, err)
		}
		got, err := unescapeJSONString(nil, quoted[1:len(quoted)-1])
		if err != nil {
			t.Fatalf("unescapeJSONString(%s): %v", quoted, err)
		}
		if string(got) != want {
			t.Errorf("unescape mismatch for %s:\n got: %q\nwant: %q", quoted, got, want)
		}
	}
	for _, s := range escapeCorpus() {
		check(t, s)
	}
	r := rand.New(rand.NewSource(43))
	for i := 0; i < 10000; i++ {
		val, ok := quick.Value(reflect.TypeFor[string](), r)
		if !ok {
			t.Fatal("failed to generate string")
		}
		check(t, val.String())
	}
}

// TestUnescapeJSONStringForms feeds raw escape forms directly (as the
// tokenizer would emit them) and compares outcome with json.Unmarshal of the
// same quoted input: both must error, or both succeed with equal values.
func TestUnescapeJSONStringForms(t *testing.T) {
	// inputs must not contain a raw unescaped quote (the tokenizer would have
	// terminated the string there)
	forms := []string{
		`\"`, `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`,
		`\u0041`, `\u00e9`, `\u4f60`, `\uD834\uDD1E`, `\ud834\udd1e`,
		`\uD834`, `\uDD1E`, `\uD834\uD834`, `\uD834x`, `\uD834\n`,
		`\uFFFD`, `\u0000`, `\u001f`,
		`\x`, `\q`, `\u`, `\u12`, `\u12G4`, `\`,
		"raw control \x01 char", "\x1f", "tab\there",
		"plain", "unicode 🎸", "invalid \xff utf8",
		`double \\\\ backslash`, `esc at end \n`,
	}
	for _, in := range forms {
		var want string
		wantErr := json.Unmarshal([]byte(`"`+in+`"`), &want)
		got, gotErr := unescapeJSONString(nil, []byte(in))
		if (wantErr != nil) != (gotErr != nil) {
			t.Errorf("error mismatch for %q: got err %v, json err %v", in, gotErr, wantErr)
			continue
		}
		if gotErr == nil && string(got) != want {
			t.Errorf("value mismatch for %q:\n got: %q\nwant: %q", in, got, want)
		}
	}
}
