package abc2xml

import (
	"fmt"

	xml "github.com/subchen/go-xmldom"
)

type pItem interface {
	Gen(*xml.Node)
}
type nItem interface {
	Gen(*xml.Node)
	Kind() int
}

// Bar *************************************************************************

const (
	BAR_SIMPLE = iota
	BAR_THIN_THICK_DOUBLE
	BAR_DOUBLE
	BAR_THICK_THIN_DOUBLE
	BAR_DOT
)
const (
	ENDING_NONE = iota
	ENDING_START
	ENDING_MID
	ENDING_END
)

type bar struct {
	Before, After int
	Type          int
	Ending        int
	EndingType    int
	LocationLeft  bool
}

func (b *bar) String() string {
	return fmt.Sprintf("Bar %d/%d/%d ", b.Before, b.Type, b.After)
}

// Misc ************************************************************************
type tChord struct {
	Root   string
	Alter  int
	Kind   string
	Ending int
}

type aKey struct {
	Mode  string
	Fifth int
}

type aTime struct {
	Symbol   string
	StrBeats string
	Beats    int
	BeatType int
}

// Measure ********************************************************************
type measure struct {
	leftBar       *bar
	rightBar      *bar
	Index         int
	Content       []pItem
	Attributes    []pItem
	NewLine       bool
	UnderlineText string
}

func (m *measure) AddAttribute(a pItem) {
	m.Attributes = append(m.Attributes, a)
}

func (m *measure) String() string {
	s := fmt.Sprintf("M(%d):", m.Index)
	if len(m.Attributes) > 0 {
		s += "{"
		for _, a := range m.Attributes {
			s += " " + fmt.Sprint(a)
		}
		s += "}"
	}
	for i, c := range m.Content {
		p := ""
		if i != 0 {
			p = ","
		}
		s += fmt.Sprintf("%s%v", p, c)
	}
	return s
}
func (m *measure) IsEmpty() bool {
	return len(m.Content) == 0
}

// Note ************************************************************************
const (
	TUPLET_NONE = iota
	TUPLET_START
	TUPLET_MID
	TUPLET_END
)
const (
	BEAM_NONE = iota
	BEAM_START
	BEAM_MID
	BEAM_END
)
const (
	TIE_NONE = iota
	TIE_START
	TIE_END
)

type baseNote struct {
	Step     string
	IsRest   bool
	Octave   int
	Alter    int
	Modifier int
	Natural  bool
	IType    int
	DotCount int
	Duration int
}

type note struct {
	baseNote

	TieStart bool
	TieStop  bool

	Beam       int
	BeamBreak  bool
	tuplet     *tuplet
	tupletStep int

	Notations []nItem

	Chord   bool
	InChord int

	Lyric string
}

func (n *note) BeamAble() bool {
	return n.IType < quarterRank
}
func (n *note) String() string {
	no := ""
	for j := 0; j < n.Octave; j++ {
		no += "'"
	}
	for j := 0; j > n.Octave; j-- {
		no += ","
	}
	be := ""
	if n.BeamBreak {
		be = "]"
	}
	b := "_"
	if n.Beam != BEAM_NONE {
		b = "!"
	}
	return fmt.Sprintf("Note: %v%v d:%v %v %v", n.Step, no, n.Duration, be, b)
}

var sharp = []byte("FCGDAEB")
var flat = []byte("BEADGCF")

func (pctx *Abc2xml) computeAlter(n *baseNote) {

	if n.Natural || n.IsRest {
		return
	}
	note := n.Step[0]
	if pctx.Fifth > 0 {
		for i := 0; i < pctx.Fifth; i++ {
			if note == sharp[i] {
				n.Alter = 1
				break
			}
		}
	} else {
		for i := 0; i < -pctx.Fifth; i++ {
			if note == flat[i] {
				n.Alter = -1
			}
		}
	}
	n.Alter += n.Modifier
}

var noteTypes = []string{
	"1024th", "512th", "256th", "128th", "64th", "32nd", "16th",
	"eighth", "quarter", "half", "whole", "breve", "long", "maxima"}

const quarterRank = 8

func (pctx *Abc2xml) determineNoteType(duration int) (itype int, dotCount int) {
	cDur := pctx.Divisions * 32 // maximaDuration

	for j := len(noteTypes) - 1; j > 0; j-- {
		if duration >= cDur {
			itype = j
			break
		}
		cDur /= 2
	}
	switch {
	case duration == cDur:
		return itype, 0
	case duration == cDur+cDur/2:
		return itype, 1
	case duration == cDur+cDur/2+cDur/4:
		return itype, 2
	default:
		pctx.warn(nil, fmt.Sprintf("Unexpected note duration (%d/%1.3f)",
			duration, float64(duration)/float64(pctx.Divisions)))
		return itype, 0
	}

}

type graceNote struct {
	baseNote
}

// Decorations  **********************************************************************

type decoration struct {
	Text string
}

// Tuplet **********************************************************************
type tuplet struct {
	n0, n1, n2 int
	Beam       int
	countDown  int
}

func (t *tuplet) String() string {
	return fmt.Sprintf("T%d:%d:%d", t.n0, t.n1, t.n2)
}
func (pctx *Abc2xml) tupletNew() *tuplet {
	t := new(tuplet)
	pctx.CTuplet = t
	return t
}

func (pctx *Abc2xml) tupletAdjust(t *tuplet, n *note) {
	n.tuplet = t
	switch t.countDown {
	case t.n2:
		n.tupletStep = TUPLET_START
		n.appendNotation(&nTuplet{TUPLET_START})
	case 1:
		n.tupletStep = TUPLET_END
		n.appendNotation(&nTuplet{TUPLET_END})
		pctx.CTuplet = nil
	default:
		n.tupletStep = TUPLET_MID
	}
	t.countDown--
	if t.countDown == t.n2 {
		n.tupletStep = TUPLET_START
	}

	switch t.n0 {
	case 2:
		n.Duration = (n.Duration * 3) / 2
	case 3:
		n.Duration = (n.Duration * 2) / 3
	default:
		pctx.warn(nil, "not implemented")
	}

}

// Partition *******************************************************************
type partition struct {
	Title    string
	Rythm    string
	Mode     string
	Measures []*measure
}

func (p *partition) String() string {
	r := fmt.Sprintf("Partition %d\n", len(p.Measures))
	for _, m := range p.Measures {
		r += "\t" + m.String() + "\n"
	}
	return r
}
func (pctx *Abc2xml) measureNew() *measure {
	p := pctx.CPartition
	if pctx.CMeasure != nil && pctx.CMeasure.IsEmpty() {
		pctx.CMeasure.Content = make([]pItem, 0)
	} else {
		m := new(measure)
		pctx.IdxMesure++
		m.Index = pctx.IdxMesure
		p.Measures = append(p.Measures, m)
		pctx.CMeasure = m
	}
	pctx.CNote = nil
	return pctx.CMeasure
}
func (m *measure) Append(item pItem) {
	m.Content = append(m.Content, item)
}

func (pctx *Abc2xml) partitionNew() *partition {
	p := new(partition)
	pctx.CPartition = p
	pctx.DefaultDuration = pctx.Divisions / 2
	pctx.measureNew()
	pctx.CMeasure.AddAttribute(&divisions{D: pctx.Divisions})

	return p
}

func (p *partition) beamResolve() {
	for _, m := range p.Measures {
		beam := false
		var lastNote *note
		closeBeam := func() {
			if lastNote != nil {
				switch lastNote.Beam {
				case BEAM_START:
					lastNote.Beam = BEAM_NONE
				case BEAM_MID:
					lastNote.Beam = BEAM_END
				}

			}
		}
		for _, e := range m.Content {
			if n, ok := e.(*note); ok {

				if n.BeamAble() {
					switch {
					case !beam && n.BeamBreak:
						n.Beam = BEAM_NONE
					case !beam && !n.BeamBreak:
						beam = true
						n.Beam = BEAM_START
					case beam && !n.BeamBreak:
						n.Beam = BEAM_MID
					case beam && n.BeamBreak:
						beam = false
						n.Beam = BEAM_END
					}
				} else {
					beam = false
					closeBeam()
				}
				lastNote = n
			}
		}
		closeBeam()
	}
}

const (
	DefaultQuarterDuration = 120
)
