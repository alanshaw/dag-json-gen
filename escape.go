package typegen

import (
	"errors"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

const hexDigits = "0123456789abcdef"

// writeEscapedJSONString writes s as a quoted JSON string, streaming through
// the writer's fixed scratch array so no heap allocation occurs. The output
// is byte-for-byte identical to encoding/json.Marshal: HTML characters
// (< > &) and U+2028/U+2029 are \u-escaped, and invalid UTF-8 is replaced
// with U+FFFD.
func (d *DagJsonWriter) writeEscapedJSONString(s string) error {
	buf := d.scratch[:0]
	buf = append(buf, '"')
	var i, n int
	for n < len(s) {
		// scan a run of characters that pass through verbatim
		var special bool
		var rn int
		for n < len(s) {
			if c := s[n]; c < utf8.RuneSelf {
				if c < 0x20 || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' {
					special = true
					rn = 1
					break
				}
				n++
			} else {
				var r rune
				r, rn = utf8.DecodeRuneInString(s[n:])
				if (r == utf8.RuneError && rn == 1) || r == '\u2028' || r == '\u2029' {
					special = true
					break
				}
				n += rn
			}
		}
		// bulk-copy the run through the scratch
		for run := s[i:n]; len(run) > 0; {
			space := len(d.scratch) - len(buf)
			if space == 0 {
				if _, err := d.w.Write(buf); err != nil {
					return err
				}
				buf = d.scratch[:0]
				space = len(d.scratch)
			}
			k := min(space, len(run))
			buf = append(buf, run[:k]...)
			run = run[k:]
		}
		if !special {
			break
		}
		// room for the longest single emission (a 6-byte \uXXXX escape)
		if len(buf)+6 > len(d.scratch) {
			if _, err := d.w.Write(buf); err != nil {
				return err
			}
			buf = d.scratch[:0]
		}
		if c := s[n]; c < utf8.RuneSelf {
			buf = appendEscapedByte(buf, c)
		} else if r, _ := utf8.DecodeRuneInString(s[n:]); r == '\u2028' || r == '\u2029' {
			buf = append(buf, '\\', 'u', '2', '0', '2', hexDigits[r&0xf])
		} else { // invalid UTF-8
			buf = append(buf, "\ufffd"...)
		}
		n += rn
		i = n
	}
	if len(buf) == len(d.scratch) {
		if _, err := d.w.Write(buf); err != nil {
			return err
		}
		buf = d.scratch[:0]
	}
	buf = append(buf, '"')
	_, err := d.w.Write(buf)
	return err
}

func appendEscapedByte(dst []byte, c byte) []byte {
	switch c {
	case '"', '\\':
		return append(dst, '\\', c)
	case '\b':
		return append(dst, '\\', 'b')
	case '\f':
		return append(dst, '\\', 'f')
	case '\n':
		return append(dst, '\\', 'n')
	case '\r':
		return append(dst, '\\', 'r')
	case '\t':
		return append(dst, '\\', 't')
	default: // other control characters, and < > & (HTML escaping)
		return append(dst, '\\', 'u', '0', '0', hexDigits[c>>4], hexDigits[c&0xf])
	}
}

// unescapeJSONString appends the decoded value of src — the raw, still-escaped
// content of a JSON string without its surrounding quotes — to dst. Semantics
// match encoding/json.Unmarshal: invalid escapes and raw control characters
// are errors, while unpaired surrogates and invalid UTF-8 are replaced with
// U+FFFD.
func unescapeJSONString(dst, src []byte) ([]byte, error) {
	for i := 0; i < len(src); {
		switch c := src[i]; {
		case c == '\\':
			i++
			if i >= len(src) {
				return nil, errors.New("unexpected end of string escape")
			}
			e := src[i]
			i++
			switch e {
			case '"', '\\', '/':
				dst = append(dst, e)
			case 'b':
				dst = append(dst, '\b')
			case 'f':
				dst = append(dst, '\f')
			case 'n':
				dst = append(dst, '\n')
			case 'r':
				dst = append(dst, '\r')
			case 't':
				dst = append(dst, '\t')
			case 'u':
				r, ok := parseHex4(src[i:])
				if !ok {
					return nil, errors.New("invalid \\u escape in string")
				}
				i += 4
				if utf16.IsSurrogate(r) {
					r2 := rune(-1)
					if i+6 <= len(src) && src[i] == '\\' && src[i+1] == 'u' {
						if v, ok := parseHex4(src[i+2:]); ok {
							r2 = v
						}
					}
					if dec := utf16.DecodeRune(r, r2); dec != utf8.RuneError {
						i += 6
						dst = utf8.AppendRune(dst, dec)
						continue
					}
					r = utf8.RuneError // unpaired surrogate
				}
				dst = utf8.AppendRune(dst, r)
			default:
				return nil, fmt.Errorf("invalid escape character %q in string", e)
			}
		case c < 0x20:
			return nil, fmt.Errorf("invalid control character %q in string", c)
		case c < utf8.RuneSelf:
			dst = append(dst, c)
			i++
		default:
			r, rn := utf8.DecodeRune(src[i:])
			if r == utf8.RuneError && rn == 1 { // invalid UTF-8
				dst = append(dst, "\ufffd"...)
			} else {
				dst = append(dst, src[i:i+rn]...)
			}
			i += rn
		}
	}
	return dst, nil
}

func parseHex4(src []byte) (rune, bool) {
	if len(src) < 4 {
		return 0, false
	}
	var r rune
	for _, c := range src[:4] {
		switch {
		case c >= '0' && c <= '9':
			c = c - '0'
		case c >= 'a' && c <= 'f':
			c = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			c = c - 'A' + 10
		default:
			return 0, false
		}
		r = r<<4 | rune(c)
	}
	return r, true
}
