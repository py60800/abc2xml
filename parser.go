// test
package abc2xml

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func (pctx *Abc2xml) warn(r *sReader, msg string) {
	c := "-"
	if r != nil {
		c = fmt.Sprint(r.idx + 1)
	}
	w := fmt.Sprintf("Warning(ln:%v,p:%v): %v", pctx.lineNumber+1, c, msg)
	fmt.Println(w)
	pctx.warnings = append(pctx.warnings, w)
}
func toInt(r rune) int {
	return int(r) - int('0')
}

// Bar *************************************************************************
func (pctx *Abc2xml) parseBar(r *sReader) {
	before := r.Eat(':')
	var after int
	var barType int
	b0 := r.Next()
	b1 := r.Next()
	switch {
	case b0 == '|' && b1 == ']':
		barType = BAR_THIN_THICK_DOUBLE
	case b0 == '|' && b1 == '|':
		barType = BAR_DOUBLE
	case b0 == '|':
		barType = BAR_SIMPLE
		r.UnRead()
	case b0 == '[' && b1 == '|':
		if r.Peek() == ']' {
			// Invisible bar => Ignore
			r.Next()
			return
		}
		barType = BAR_THICK_THIN_DOUBLE
	case b0 == '.' && b1 == '|':
		barType = BAR_DOT
		pctx.CMeasure.Append(&bar{
			Type: BAR_DOT,
		})
		return

	case b0 == '[' && unicode.IsDigit(b1):
		barType = BAR_SIMPLE
		r.UnRead()
	default:
		if before < 2 {
			pctx.warn(r, "Invalid bar")
			return
		}
		// :: alias for :|:
		before = 1
		after = 1
		r.UnRead() // b1
		r.UnRead() //b2
	}
	if after == 0 { // else ::
		after = r.Eat(':')
	}

	r.SkipSpace()
	endingNumber := 0
	c := r.Next()
	//fmt.Println(r)
	switch {
	case unicode.IsDigit(c):
		endingNumber = toInt(c)
	case c == '[':
		if n := r.Next(); unicode.IsDigit(n) {
			endingNumber = toInt(n)
		} else {
			r.UnRead()
			r.UnRead()
		}
	default:
		r.UnRead()
	}
	//fmt.Println("Bar:", pctx.IdxMesure, "Typ:", barType, "Before:", before, "After:", after, "Ending::", endingNumber)
	if pctx.CMeasure.IsEmpty() {
		endingType := ENDING_NONE
		if endingNumber > 0 {
			pctx.pendingVariant = endingNumber
			endingType = ENDING_START
		}

		// This is a new Measure => Set on Left
		pctx.CMeasure.leftBar = &bar{
			Before:       0,
			After:        after,
			Type:         barType,
			Ending:       endingNumber,
			EndingType:   endingType,
			LocationLeft: true,
		}
		return
	}
	// Finish current Measure
	if before > 0 || barType != BAR_SIMPLE {
		endingType := ENDING_NONE
		if pctx.pendingVariant > 0 {
			endingType = ENDING_END
		}
		pctx.CMeasure.rightBar = &bar{
			Before:       before,
			After:        0,
			Type:         barType,
			Ending:       pctx.pendingVariant,
			EndingType:   endingType,
			LocationLeft: false,
		}
		pctx.pendingVariant = 0
	}

	// Start a new Measure
	pctx.measureNew()

	if endingNumber > 0 || after > 0 || barType != BAR_SIMPLE {
		endingType := ENDING_NONE
		if endingNumber > 0 {
			pctx.pendingVariant = endingNumber
			endingType = ENDING_START
		}

		pctx.CMeasure.leftBar = &bar{
			Before:       0,
			After:        after,
			Type:         barType,
			Ending:       endingNumber,
			EndingType:   endingType,
			LocationLeft: true,
		}
	}
}

func (pctx *Abc2xml) parseModifier(r *sReader) (bool, int) {
	modifier := 0
	if r.Peek() == '=' {
		r.Next()
		return true, 0
	}
	if modifier = r.Eat('^'); modifier > 0 {
		return false, modifier
	}
	modifier = r.Eat('_')
	return false, -modifier
}

// Decoration ******************************************************************
var decoSet = ".~HLMOPSTuv"
var noteSet = "ABCDEFGabcdefgzZxX"
var noteAlter = "^_="
var validNoteStart map[rune]bool
var validDeco map[rune]bool
var validNote map[rune]bool

func string2BoolMap(str string) map[rune]bool {
	r := make(map[rune]bool)
	for _, c := range str {
		r[c] = true
	}
	return r
}
func init() {
	validNote = string2BoolMap(noteSet)
	validDeco = string2BoolMap(decoSet + "!")
	validNoteStart = string2BoolMap(decoSet + "!" + noteSet + noteAlter)
}

func (pctx *Abc2xml) parseDecoration(r *sReader) {
	deco := ""
	for {
		d := r.Next()
		switch {
		case d == '!':
			for {
				n := r.Next()
				if n == '!' || n == 0 {
					break
				}
			}
		case validDeco[d]:
			deco += string(d)
		default:
			r.UnRead()
			if deco != "" && pctx.CNote != nil {
				pctx.CNote.Decorate(deco)
			}
			return
		}
	}
}
func (n *note) appendNotation(i nItem) {
	n.Notations = append(n.Notations, i)
}

func (pctx *Abc2xml) parseStep(r *sReader) (bool, string, int) {
	if !validNote[r.Peek()] {
		return true, "", 0
	}
	octave := 0
	n := r.Next()
	if unicode.IsLower(n) {
		octave++
	}
	n = unicode.ToUpper(n)
	if n == 'Z' || n == 'X' {
		return true, "", 0 // rest
	}
	octave += r.Eat('\'')
	octave -= r.Eat(',')
	return false, string(n), octave

}

func (pctx *Abc2xml) parseDuration(r *sReader) (duration int, itype int, dots int) {
	duration = pctx.DefaultDuration
	i0 := r.ReadInt(1)
	duration *= i0

	if r.Peek() == '/' {
		cpt := r.Eat('/')
		switch cpt {
		case 0:
		//Done
		case 1:
			i1 := r.ReadInt(2)
			if i1 != 0 {
				duration /= i1
			}
		default:
			for i := 0; i < cpt; i++ {
				duration /= 2
			}
		}
	}
	itype, dots = pctx.determineNoteType(duration)
	return duration, itype, dots
}

func (pctx *Abc2xml) parseBaseNote(r *sReader) *baseNote {
	note := &baseNote{}
	pctx.parseDecoration(r)
	note.Natural, note.Modifier = pctx.parseModifier(r)
	note.IsRest, note.Step, note.Octave = pctx.parseStep(r)

	pctx.computeAlter(note)

	note.Duration, note.IType, note.DotCount = pctx.parseDuration(r)
	return note
}

func (pctx *Abc2xml) parseGracesNotes(r *sReader) {
	if r.Next() != '{' {
		pctx.warn(r, "Sequence Error Grace note")
	}
	for {
		switch r.Peek() {
		case 0:
			return
		case '}':
			r.Next()
			return
		default:
			if validNoteStart[r.Peek()] {
				g := &graceNote{}
				g.baseNote = *pctx.parseBaseNote(r)
				pctx.CMeasure.Append(g)
			} else {
				return
			}
		}
	}
}
func (pctx *Abc2xml) tieBegin(note *note) {
	note.TieStart = true
	note.appendNotation(&tie{true})
	pctx.PendingTie = true
}
func (pctx *Abc2xml) tieFinish(note *note) {
	note.TieStop = true
	note.appendNotation(&tie{false})
}

func (pctx *Abc2xml) parseNote(r *sReader) {
	note := &note{}
	pctx.CNote = note
	pctx.CMeasure.Append(note)
	note.baseNote = *pctx.parseBaseNote(r)
	if pctx.UnissonOnGoing {
		if pctx.UnissonStarted {
			note.InChord = 1 // mid
			note.Chord = true
		} else {
			note.InChord = 2 // start
		}
		pctx.UnissonStarted = true
	}
	if pctx.CTuplet != nil {
		pctx.tupletAdjust(pctx.CTuplet, pctx.CNote)
	}
	if pctx.PendingTie {
		pctx.tieFinish(pctx.CNote)
		pctx.PendingTie = false
	}
	if r.Peek() == '-' {
		pctx.tieBegin(note)
		r.Next()
	}
	if unicode.IsSpace(r.Peek()) {
		note.BeamBreak = true
		r.Next()
	}
	if !pctx.UnissonOnGoing {
		if pctx.BrokenRythm != 0 {
			note.ProcessBrokenRythm(pctx.BrokenRythm, false, pctx.BrokenRythmCount)
			pctx.BrokenRythm = 0
		}
	}

	if pctx.pendingSlur != nil {
		note.appendNotation(pctx.pendingSlur)
		pctx.pendingSlur = nil
	}
}

/*
(2	2 notes in the time of 3
(3	3 notes in the time of 2
(4	4 notes in the time of 3
(5	5 notes in the time of n
(6	6 notes in the time of 2
(7	7 notes in the time of n
(8	8 notes in the time of 3
(9	9 notes in the time of n
*/
var n0Ton1 = map[int]int{
	2: 3,
	3: 2,
	4: 3,
	6: 2,
	8: 3,
}

func (pctx *Abc2xml) parseTuplet(r *sReader) {
	if r.Next() != '(' {
		panic("Tuplet")
	}
	t := pctx.tupletNew()
	pctx.CMeasure.Append(t)

	t.n0 = r.ReadInt(0)
	if r.Peek() == ':' {
		r.Next()
		t.n1 = r.ReadInt(0)
		if r.Peek() == ':' {
			r.Next()
			t.n2 = r.ReadInt(0)
		}
	}
	if t.n1 == 0 {
		t.n1 = n0Ton1[t.n0]
	}
	if t.n2 == 0 {
		t.n2 = t.n0
	}

	r.SkipSpace()
	t.countDown = t.n2
}

/*
augmented	Triad: major third, augmented fifth.
augmented-seventh	Seventh: augmented triad, minor seventh.
diminished	Triad: minor third, diminished fifth.
diminished-seventh	Seventh: diminished triad, diminished seventh.
major	Triad: major third, perfect fifth.
major-11th	11th: major-ninth, perfect 11th.
major-13th	13th: major-11th, major 13th.
major-minor	Seventh: minor triad, major seventh.
major-ninth	Ninth: major-seventh, major ninth.
major-seventh	Seventh: major triad, major seventh.
major-sixth	Sixth: major triad, added sixth.
minor	Triad: minor third, perfect fifth.
minor-11th	11th: minor-ninth, perfect 11th.
minor-13th	13th: minor-11th, major 13th.
minor-ninth	Ninth: minor-seventh, major ninth.
minor-seventh	Seventh: minor triad, minor seventh.
minor-sixth	Sixth: minor triad, added sixth.
suspended-fourth	Suspended: perfect fourth, perfect fifth.
suspended-second	Suspended: major second, perfect fifth.
*/
/*
m or min        minor
maj             major
dim             diminished
aug or +        augmented
sus             suspended
7, 9 ...        7th, 9th, etc.
*/
var chordSuffix = map[int]string{
	2: "-second", 4: "-fourth", 6: "-sixth", 7: "-seventh",
	9: "-ninth", 11: "-11th", 13: "-13th"}

var chordPrefix = map[string]string{
	"min": "minor",
	"maj": "major",
	"dim": "diminished",
	"aug": "augmented",
	"+":   "augmented",
	"sus": "suspended",
}

func (pctx *Abc2xml) parseSlur(r *sReader) {
	c := r.Next()
	switch c {
	case '(':
		pctx.slurNew(true)
	case ')':
		s := pctx.slurNew(false)
		if pctx.CNote != nil {
			pctx.CNote.appendNotation(s)
		}
	}
}

func (pctx *Abc2xml) parseChord(r *sReader) {
	r.Next()
	step := unicode.ToUpper(r.Next())
	a := r.Next()
	var alter int
	switch a {
	case '#':
		alter = 1
	case 'b', 'B':
		alter = -1
	default:
		r.UnRead()
	}

	letters := ""
	digits := ""
loop:
	for {
		c := r.Next()
		switch {
		case c == 0:
			return
		case unicode.IsLetter(c):
			letters = letters + string(c)
		case unicode.IsDigit(c):
			digits = digits + string(c)
		case c == '"':
			break loop
		default:
		}
	}
	kind := "major"
	if letters == "m" {
		kind = "minor"
	} else {
		letters = strings.ToLower(letters)
		for k, v := range chordPrefix {
			if strings.Contains(letters, k) {
				kind = v
				break
			}
		}
	}
	n, _ := strconv.Atoi(digits)
	if s, ok := chordSuffix[n]; ok {
		kind += s
	}
	fmt.Printf("Chord:%v(%v)%v %v(%v)\n", string(step), alter, letters, kind, n)
	pctx.CMeasure.Append(&tChord{
		Root:  string(step),
		Alter: alter,
		Kind:  kind,
	})

}
func (n *note) ProcessBrokenRythm(code rune, first bool, count int) {
	switch {
	case code == '>' && first, code == '<' && !first:
		for i := 0; i < count; i++ {
			n.Duration = (n.Duration * 3) / 2
			n.DotCount++
		}
	case code == '>' && !first, code == '<' && first:
		for i := 0; i < count; i++ {
			n.Duration = n.Duration / 2
			n.IType--
		}
	}
}

func (pctx *Abc2xml) iterChordBack(f func(*note)) {
	for i := len(pctx.CMeasure.Content) - 1; i >= 0; i-- {
		if n, ok := pctx.CMeasure.Content[i].(*note); ok {
			switch n.InChord {
			case 0:
				return
			case 1:
				f(n)
			case 2:
				f(n)
				return
			}
		}
	}
}

func (pctx *Abc2xml) parseBroken(r *sReader) {

	pctx.BrokenRythm = r.Peek()
	pctx.BrokenRythmCount = r.Eat(pctx.BrokenRythm)

	if pctx.CNote == nil {
		pctx.BrokenRythm = 0
		return
	}
	if pctx.CNote.InChord == 0 {
		pctx.CNote.ProcessBrokenRythm(pctx.BrokenRythm, true, pctx.BrokenRythmCount)
	} else {
		pctx.iterChordBack(func(n *note) {
			n.ProcessBrokenRythm(pctx.BrokenRythm, true, pctx.BrokenRythmCount)
		})
	}
}
func (pctx *Abc2xml) parseFract(r *sReader, d1, d2 int) (int, int) {
	r.SkipSpace()
	v1 := r.ReadInt(d1)
	r.SkipSpace()
	if r.Next() != '/' {
		pctx.warn(r, "Bad Fract")
		return v1, d2
	}
	return v1, r.ReadInt(d2)
}
func (pctx *Abc2xml) parseInlineInfo(r *sReader) {
	r.Next()
	t := r.Next()
	if r.Next() != ':' {
		pctx.warn(r, "Bad Inline")
	}
	r.SkipSpace()
	switch t {
	case 'M':
		pctx.parseTempo(r)
	case 'L':
		pctx.parseUnit(r)
	case 'K':
		pctx.parseKey(r)
	default:
		pctx.warn(r, "Bad Inline")
	}
	for {
		c := r.Next()
		if c == ']' || c == 0 {
			return
		}
	}
}
func (pctx *Abc2xml) parseUnisson(r *sReader) {
	pctx.UnissonOnGoing = true
	pctx.UnissonStarted = false
	defer func() {
		pctx.UnissonOnGoing = false
	}()
	r.Next() // '['
	for {
		c := r.Peek()
		switch {
		case unicode.IsSpace(c):
			if pctx.CNote != nil {
				pctx.CNote.BeamBreak = true
			}
			r.Next()
		case c == 0:
			return
		case c == ']':
			r.Next()
			d, _, _ := pctx.parseDuration(r)
			if d != pctx.DefaultDuration {
				pctx.iterChordBack(func(n *note) {
					n.Duration = (n.Duration * d) / pctx.DefaultDuration
					n.IType, n.DotCount = pctx.determineNoteType(n.Duration)
				})
			}
			if pctx.BrokenRythm != 0 {
				pctx.iterChordBack(func(n *note) {
					n.ProcessBrokenRythm(pctx.BrokenRythm, false, pctx.BrokenRythmCount)
				})
				pctx.BrokenRythm = 0
			}
			return
		case validNoteStart[c]:
			pctx.parseNote(r)
		default:
			r.Next()
		}
	}
}

var lInLine = map[rune]bool{
	'L': true,
	'M': true,
	'K': true,
}

func (pctx *Abc2xml) parseContent(str string) bool {
	parser := sReaderNew(str)

	pctx.CMeasure.NewLine = !pctx.LineContinuation
	pctx.LineContinuation = false

	pctx.CNote = nil
	for {
		r := parser.Peek()
		r1 := parser.PeekN(1)
		r2 := parser.PeekN(2)
		switch {
		case r == rune(0):
			return true
		case r == '>', r == '<':
			pctx.parseBroken(parser)
		case r == '"':
			pctx.parseChord(parser)
		case r == '[' && lInLine[r1] && r2 == ':':
			pctx.parseInlineInfo(parser)
		case r == '.' && r1 == '|',
			r == '|',
			r == '[' && r1 == '|',
			r == ':',
			r == '[' && unicode.IsDigit(r1):
			fmt.Println("ParseBar:", string(r), string(r1))
			pctx.parseBar(parser)
		case r == '[':
			pctx.parseUnisson(parser)
		case unicode.IsSpace(r):
			// Beam end
			if pctx.CNote != nil {
				pctx.CNote.BeamBreak = true
			}
			parser.Next()
		case r == '(' && unicode.IsDigit(r1):
			pctx.parseTuplet(parser)
		case r == '(', r == ')':
			pctx.parseSlur(parser)
		case validNoteStart[r]:
			pctx.parseNote(parser)
		case r == '{':
			pctx.parseGracesNotes(parser)
		case r == '-':
			if pctx.CNote != nil {
				pctx.tieBegin(pctx.CNote)
			}
			parser.Next()
		case r == '\\':
			pctx.LineContinuation = true
			parser.Next()
			parser.SkipSpace()
		default:
			pctx.warn(parser, fmt.Sprintf("Unexpected:%v", string(r)))
			parser.Next()
		}
	}

	return true
}

type Abc2xml struct {
	CPartition *partition
	IdxMesure  int
	CMeasure   *measure
	CNote      *note

	LineContinuation bool

	CTuplet *tuplet

	Beats    int
	BeatType int
	Fifth    int

	Divisions       int // Ticks per QuarterNote
	DefaultDuration int // for Quarter

	pendingVariant int

	BrokenRythm      rune
	BrokenRythmCount int

	PendingTie bool

	slurLevel   int
	pendingSlur *slur

	UnissonOnGoing bool
	UnissonStarted bool

	warnings     []string
	lineNumber   int
	AbortRequest bool
}

func Abc2xmlNew() *Abc2xml {
	parser := &Abc2xml{}
	parser.SetDivisions(DefaultQuarterDuration * 2)
	return parser
}
func (pctx *Abc2xml) SetDivisions(divisions int) {
	pctx.Divisions = divisions
	pctx.DefaultDuration = divisions / 2 // L:1/8
}
func (pctx *Abc2xml) Warnings() []string {
	return pctx.warnings
}
func (pctx *Abc2xml) Run(abc string) (string, error) {
	lines := strings.Split(abc, "\n")
	p := pctx.partitionNew()
	started := false
	for i, str := range lines {
		if pctx.AbortRequest {
			return "", fmt.Errorf("Failed to parse ABC")
		}
		pctx.lineNumber = i
		str = strings.TrimSpace(strings.TrimSuffix(str, "\n"))
		if started && len(str) == 0 {
			break
		}
		if len(str) != 0 {
			started = true
		}

		parser := sReaderNew(str) // Include comment strip

		if unicode.IsLetter(parser.Peek()) && parser.PeekN(1) == ':' {
			pctx.processInfo(parser)
		} else {
			pctx.parseContent(str)
		}
	}
	p.beamResolve()
	xml := generateXml(p)

	return xml, nil

}
func (pctx *Abc2xml) GetTitle() string {
	if pctx.CPartition != nil {
		return pctx.CPartition.Title
	}
	return ""
}
func (pctx *Abc2xml) GetRythm() string {
	if pctx.CPartition != nil {
		return pctx.CPartition.Rythm
	}
	return ""
}
func (pctx *Abc2xml) GetMod() string {
	if pctx.CPartition != nil {
		return pctx.CPartition.Mode
	}
	return ""
}
