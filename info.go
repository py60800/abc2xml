// info
package abc2xml

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Key/Mode ********************************************************************

const keyConfig = `
7  C# A#m G#Mix D#Dor E#Phr F#Lyd B#Loc
6  F# D#m C#Mix G#Dor A#Phr BLyd E#Loc
5  B G#m F#Mix C#Dor D#Phr ELyd A#Loc
4  E C#m BMix F#Dor G#Phr ALyd D#Loc
3  A F#m EMix BDor C#Phr DLyd G#Loc
2  D Bm AMix EDor F#Phr GLyd C#Loc
1  G Em DMix ADor BPhr CLyd F#Loc
0 C Am GMix DDor EPhr FLyd BLoc
-1  F Dm CMix GDor APhr BbLyd ELoc
-2  Bb Gm FMix CDor DPhr EbLyd ALoc
-3  Eb Cm BbMix FDor GPhr AbLyd DLoc
-4  Ab Fm EbMix BbDor CPhr DbLyd GLoc
-5  Db Bbm AbMix EbDor FPhr GbLyd CLoc
-6  Gb Ebm DbMix AbDor BbPhr CbLyd FLoc
-7  Cb Abm GbMix DbDor EbPhr FbLyd BbLoc
`

var mod2Fifth map[string]int

func init() {
	mod2Fifth = make(map[string]int)
	kc := strings.Split(keyConfig, "\n")
	for _, lk := range kc {
		l := strings.Split(lk, " ")
		if len(l) > 1 {
			n, _ := strconv.Atoi(l[0])
			for _, s := range l[1:] {
				if s != "" {
					mod2Fifth[s] = n
				}
			}
		}
	}
}

/*
var xmlModes = map[string]string{
	"Mix": "mixolydian",
	"m":   "minor",
	"Dor": "dorian",
	"Phr": "phrygian",
	"Lyd": "lydian",
	"Loc": "locrian",
	"":    "major",
}
*/
// MuseScore seems confused
var xmlModes = map[string]string{
	"Mix": "major",
	"m":   "minor",
	"Dor": "minor",
	"Phr": "major",
	"Lyd": "major",
	"Loc": "mojor",
	"":    "major",
}

func (pctx *Abc2xml) parseKey(r *sReader) {
	r.SkipSpace()
	note := string(unicode.ToUpper(r.Next()))
	alter := r.Peek()
	switch alter {
	case '#':
		note += "#"
		r.Next()
	case 'b':
		note += "b"
		r.Next()
	}
	r.SkipSpace()

	mod := ""
	for i := 0; ; i++ {
		c := r.Next()
		if !unicode.IsLetter(c) {
			r.UnRead()
			break
		}
		switch i {
		case 0:
			mod += string(unicode.ToUpper(c))
		case 1, 2:
			mod += string(unicode.ToLower(c))
		default:
			// skip
		}
	}
	switch mod {
	case "Maj", "Ion":
		mod = ""
	case "Min", "Aeo", "M":
		mod = "m"
	}
	pctx.Fifth = mod2Fifth[note+mod]
	tmod, ok := xmlModes[mod]
	if !ok {
		tmod = "major"
	}
	//	fmt.Printf("Key: %v(%v) x:%v f:%v\n", note, mod, tmod, pctx.Fifth)
	pctx.CMeasure.AddAttribute(&aKey{
		Mode:  tmod,
		Fifth: pctx.Fifth,
	})
	if pctx.CPartition.Mode == "" {
		pctx.CPartition.Mode = note + mod
	}
}

// Title ***********************************************************************
func (pctx *Abc2xml) parseTitle(r *sReader) {
	if pctx.CPartition.Title == "" {
		pctx.CPartition.Title = strings.TrimSpace(r.Rest())
	}
}

// Title ***********************************************************************
func (pctx *Abc2xml) parseRythm(r *sReader) {
	if pctx.CPartition.Rythm == "" {
		pctx.CPartition.Rythm = strings.TrimSpace(r.Rest())
	}
}

// Title ***********************************************************************
func (pctx *Abc2xml) parseWords(r *sReader) {
	p := pctx.CPartition
	for j := len(p.Measures) - 1; j >= 0; j-- {
		m := p.Measures[j]
		if m.NewLine {
			m.UnderlineText = strings.TrimSpace(r.Rest())
			return
		}
	}
}

type iterNote struct {
	iMeasure int
	inote    int
	p        *partition
}

func (p *partition) iterNoteNewFromStartOfLine() iterNote {
	for idxMeasure := len(p.Measures) - 1; idxMeasure >= 0; idxMeasure-- {
		m := p.Measures[idxMeasure]
		if m.NewLine {
			return iterNote{
				iMeasure: idxMeasure,
				inote:    0,
				p:        p,
			}
		}
	}
	return iterNote{
		iMeasure: len(p.Measures),
		p:        p,
	}
}
func (iter *iterNote) NextMeasure() {
	iter.iMeasure++
	iter.inote = 0
}
func (iter *iterNote) Next() *note {
	for {
		if iter.iMeasure >= len(iter.p.Measures) {
			return nil
		}
		if iter.inote >= len(iter.p.Measures[iter.iMeasure].Content) {
			iter.iMeasure++
			iter.inote = 0
		} else {
			if n, ok := iter.p.Measures[iter.iMeasure].Content[iter.inote].(*note); ok {
				iter.inote++
				return n

			} else {
				iter.inote++
			}
		}
	}
}
func (pctx *Abc2xml) parseWords2(r *sReader) {
	iterNote := pctx.CPartition.iterNoteNewFromStartOfLine()
	txt := ""
	setLyric := func() {
		if n := iterNote.Next(); n != nil {
			n.Lyric = txt

		}
	}
	r.SkipSpace()

	for {
		c := r.Next()
		switch {
		case c == 0:
			setLyric()
			return
		case unicode.IsSpace(c):
			setLyric()
			txt = ""
			r.SkipSpace()

		case c == '-':
			setLyric()
			txt = ""
		case c == '_', c == '*':
			iterNote.Next()

		case c == '|':
			iterNote.NextMeasure()
		case c == '\\':
			if t := r.Next(); t == 0 {
				setLyric()
				return
			}
			fallthrough
		default:
			txt = txt + string(c)
		}

	}
}

/*
-	(hyphen) break between syllables within a word
_	(underscore) previous syllable is to be held for an extra note
*	one note is skipped (i.e. * is equivalent to a blank syllable)
~	appears as a space; aligns multiple words under one note
\-	appears as hyphen; aligns multiple syllables under one note
|	advances to the next bar
*/

// Unit ************************************************************************
func (pctx *Abc2xml) parseUnit(r *sReader) {
	a, b := pctx.parseFract(r, 1, 8)
	if a != 1 || b == 0 {
		pctx.warn(r, "Odd Unit")
		return
	}
	// should check validity

	pctx.DefaultDuration = (pctx.Divisions * 4) / b
	fmt.Println("ParseUnit:", a, b, pctx.Divisions, pctx.DefaultDuration)
}

// Tempo ***********************************************************************
func (pctx *Abc2xml) parseTempo(r *sReader) {
	atime := &aTime{
		Symbol:   "",
		StrBeats: "4",
		Beats:    4,
		BeatType: 4,
	}
	defer func() {
		pctx.Beats = atime.Beats
		pctx.BeatType = atime.BeatType
		pctx.CMeasure.AddAttribute(atime)
	}()

	r.SkipSpace()

	if r.Peek() == 'C' {
		r.Next()
		if r.Peek() == '|' {
			r.Next()
			atime.Symbol = "cut"
			atime.Beats = 2
			atime.BeatType = 2
			return
		} else {
			atime.Symbol = "common"
			return
		}
	}

	beats := 0
	sBeats := ""
	if r.Peek() == '(' {
		r.Next()
	}
	var pfx string
loop:
	for {
		r.SkipSpace()
		n := r.ReadInt(4)
		beats += n
		sBeats += pfx + strconv.Itoa(n)
		r.SkipSpace()
		switch r.Next() {
		case '+':
			pfx = "+"
		case ')':
			break loop
		case '/':
			r.UnRead()
			break loop
		default:
			pctx.warn(r, "Meter syntax error")
			return

		}
	}
	r.SkipSpace()
	if r.Next() != '/' {
		pctx.warn(r, "Meter syntax error")
	}
	r.SkipSpace()
	atime.Beats = beats
	atime.StrBeats = sBeats
	atime.BeatType = r.ReadInt(4)

}

// *****************************************************************************
func (pctx *Abc2xml) processInfo(r *sReader) {
	t := r.Next()
	r.Next() // Skip ':'
	switch t {
	case 'T':
		pctx.parseTitle(r)
	case 'R':
		pctx.parseRythm(r)
	case 'M':
		pctx.parseTempo(r)
	case 'L':
		pctx.parseUnit(r)
	case 'K':
		pctx.parseKey(r)
	case 'W':
		pctx.parseWords(r)
	case 'w':
		pctx.parseWords2(r)
	case 'V':
		pctx.warn(nil, "Multiple voices not supported")
		pctx.AbortRequest = true
	default:
		pctx.warn(nil, fmt.Sprintf("\"%v\": Directive ignored", string(t)))
	}
}
