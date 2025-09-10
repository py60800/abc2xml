// reader.go
package abc2xml

import (
	"fmt"
	"unicode"
)

type sReader struct {
	data []rune
	idx  int
}

func sReaderNew(str string) *sReader {
	p := &sReader{
		data: []rune(str),
		idx:  0,
	}
	// Strip comment
	escape := false
	for i, c := range p.data {
		if !escape && c == '%' {
			p.data = p.data[:i]
			break
		}
		escape = c == '\\'
	}
	return p
}
func (r *sReader) Next() (c rune) {
	/*defer func() {
		fmt.Print(string(c))
	}()
	*/
	if r.idx >= len(r.data) {
		r.idx++

		return rune(0)
	}
	res := r.data[r.idx]
	r.idx++
	return res
}
func (r *sReader) Peek() rune {
	if r.idx >= len(r.data) {
		return rune(0)
	}
	res := r.data[r.idx]
	return res
}
func (r *sReader) PeekN(d int) rune {
	if r.idx+d >= len(r.data) {
		return rune(0)
	}
	return r.data[r.idx+d]
}
func (r *sReader) UnRead() {
	if r.idx == 0 {
		return
	}
	r.idx--
}
func (r *sReader) Rest() string {
	return string(r.data[r.idx:])
}
func (r *sReader) SkipSpace() {
	for {
		c := r.Next()
		if !unicode.IsSpace(c) {
			r.UnRead()
			return
		}
	}
}
func (r *sReader) String() string {
	return fmt.Sprintf("Parser:[%d]%v ... %v", r.idx, string(r.data[:min(r.idx, len(r.data))]), string(r.data[min(r.idx, len(r.data)):]))
}
func (r *sReader) Eat(c rune) int {
	count := 0
	for r.Next() == c {
		count++
	}
	r.UnRead()
	return count
}
func (r *sReader) ReadInt(defaultValue int) int {
	found := false
	val := 0
	for {
		d := r.Next()
		if unicode.IsDigit(d) {
			found = true
			val = val*10 + int(d-'0')
		} else {
			r.UnRead()
			break
		}
	}
	if found {
		return val
	} else {
		return defaultValue
	}
}
