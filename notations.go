package abc2xml

import (
	"fmt"

	xml "github.com/subchen/go-xmldom"
)

const (
	NOTA_BASE = iota
	NOTA_TECHNICAL
	NOTA_ORNAMENTS
	NOTA_ARTICULATIONS
)

/*
.       staccato mark
~       Irish roll
H       fermata
L       accent or emphasis
M       lowermordent
O       coda
P       uppermordent
S       segno
T       trill
u       up-bow
v       down-bow
*/
/*
<accent>
<breath-mark>
<caesura>
<detached-legato>
<doit>
<falloff>
<other-articulation>
<plop>
<scoop>
<soft-accent>
<spiccato>
<staccatissimo>
<staccato>
<stress>
<strong-accent>
<tenuto>
<unstress>

// Ornaments
<accidental-mark>
<delayed-inverted-turn>
<delayed-turn>
<haydn>
<inverted-mordent>
<inverted-turn>
<inverted-vertical-turn>
<mordent>
<other-ornament>
<schleifer>
<shake>
<tremolo>
<trill-mark>
<turn>
<vertical-turn>
<wavy-line>
*/
// Roll
type roll struct{}

func (*roll) Kind() int {
	return NOTA_ORNAMENTS
}
func (*roll) Gen(x *xml.Node) {
	x.CreateNode("turn")
}

// trill
type trill struct{}

func (*trill) Kind() int {
	return NOTA_ORNAMENTS
}
func (*trill) Gen(x *xml.Node) {
	x.CreateNode("trill-mark")
}

// Bow *************************************************************************
type bow struct {
	code rune
}

func (b *bow) Kind() int {
	return NOTA_TECHNICAL
}
func (t *bow) Gen(x *xml.Node) {
	if t.code == 'v' {
		x.CreateNode("down-bow")
	} else {
		x.CreateNode("up-bow")
	}
}

// Tie *************************************************************************

type tie struct {
	start bool
}

func (t *tie) Kind() int {
	return NOTA_BASE
}
func (t *tie) Gen(x *xml.Node) {
	tn := x.CreateNode("tied")
	SetAttributeAlt(tn, "type", t.start, "start", "stop")
}

// Staccato ********************************************************************
type staccato struct{}

func (a *staccato) Kind() int {
	return NOTA_ARTICULATIONS
}
func (s *staccato) Gen(x *xml.Node) {
	x.CreateNode("staccato")
}

type accent struct{}

func (a *accent) Kind() int {
	return NOTA_ARTICULATIONS
}
func (s *accent) Gen(x *xml.Node) {
	x.CreateNode("accent")
}

func (n *note) GenNotation(x *xml.Node) {
	if len(n.Notations) == 0 {
		return
	}
	type nota struct {
		kind int
		val  string
	}
	notas := []nota{
		nota{NOTA_BASE, ""},
		nota{NOTA_ARTICULATIONS, "articulations"},
		nota{NOTA_ORNAMENTS, "ornaments"},
		nota{NOTA_TECHNICAL, "technical"},
	}
	notations := x.CreateNode("notations")
	for _, N := range notas {
		var auxNode *xml.Node
		for _, elem := range n.Notations {
			if elem.Kind() == N.kind {
				if N.kind == NOTA_BASE {
					elem.Gen(notations)
				} else {
					if auxNode == nil {
						auxNode = notations.CreateNode(N.val)
					}
					elem.Gen(auxNode)
				}
			}
		}
	}
}

// Ntuplet Helper for notation
type nTuplet struct {
	Type int
}

func (nt *nTuplet) Kind() int {
	return NOTA_BASE
}
func nTupletNew(t int) *nTuplet {
	return &nTuplet{
		Type: t,
	}
}
func (nt *nTuplet) Gen(x *xml.Node) {
	if nt.Type == TUPLET_MID {
		return
	}
	tuplet := x.CreateNode("tuplet")
	switch nt.Type {
	case TUPLET_START:
		tuplet.SetAttributeValue("type", "start")
		tuplet.SetAttributeValue("bracket", "yes")
	case TUPLET_END:
		tuplet.SetAttributeValue("type", "stop")
	}
}

// Slur ************************************************************************
// Slurs ***********************************************************************

type slur struct {
	Start bool
	Level int
}

func (*slur) Kind() int {
	return NOTA_BASE
}
func (pctx *Abc2xml) slurNew(start bool) *slur {
	s := &slur{
		Start: start,
	}
	if start {
		pctx.pendingSlur = s
		pctx.slurLevel++
		s.Level = pctx.slurLevel
	} else {
		s.Level = pctx.slurLevel
		pctx.slurLevel--

	}
	return s
}
func (s *slur) Gen(x *xml.Node) {
	sn := x.CreateNode("slur")
	sn.SetAttributeValue("number", fmt.Sprint(s.Level))
	SetAttributeAlt(sn, "type", s.Start, "start", "stop")
}

func (n *note) Decorate(decorations string) {
	for _, s := range decorations {
		switch s {
		case '.':
			n.appendNotation(&staccato{})
		case '~':
			n.appendNotation(&roll{})
		case 'T':
			n.appendNotation(&trill{})
		case 'u', 'v':
			n.appendNotation(&bow{s})
		case 'L':
			n.appendNotation(&accent{})

		}
	}
}
