package abc2xml

import (
	"fmt"
	"time"

	xml "github.com/subchen/go-xmldom"
)

const pi = `<!DOCTYPE score-partwise PUBLIC "-//Recordare//DTD MusicXML 3.0 Partwise//EN" "http://www.musicxml.org/dtds/partwise.dtd">`

func addPartList(n *xml.Node, partId string) {
	pl := n.CreateNode("part-list")
	sp := pl.CreateNode("score-part").SetAttributeValue("id", partId)
	sp.CreateNode("part-name")

}
func setTitle(n *xml.Node, title string) {
	w := n.CreateNode("work")
	t := w.CreateNode("work-title")
	t.Text = title
}

type divisions struct {
	D int
}

func (d *divisions) Gen(x *xml.Node) {
	dn := x.CreateNode("divisions")
	dn.Text = fmt.Sprint(d.D)
}

func (a *aKey) Gen(x *xml.Node) {
	kn := x.CreateNode("key")
	f := kn.CreateNode("fifths")
	f.Text = fmt.Sprint(a.Fifth)
	m := kn.CreateNode("mode")
	m.Text = a.Mode

}

/*
dashed
dotted
heavy
heavy-heavy
heavy-light
light-heavy
light-light
regular
short
tick
*/
var barStyle = map[int]string{
	BAR_THIN_THICK_DOUBLE: "light-heavy",
	BAR_THICK_THIN_DOUBLE: "heavy-light",
	//	BAR_SIMPLE:            "regular",
	BAR_DOUBLE: "light-light",
	BAR_DOT:    "dashed",
}

func SetAttributeAlt(x *xml.Node, name string, sel bool, v1, v2 string) {
	var v string
	if sel {
		v = v1
	} else {
		v = v2
	}
	x.SetAttributeValue(name, v)
}
func (b *bar) Gen(x *xml.Node) {

	blNode := x.CreateNode("barline")
	if b.Type != BAR_DOT {
		SetAttributeAlt(blNode, "location", b.LocationLeft, "left", "right")
	} else {
		blNode.SetAttributeValue("location", "middle")
	}
	if bs, ok := barStyle[b.Type]; ok && !b.LocationLeft {
		st := blNode.CreateNode("bar-style")
		st.Text = bs
	}
	if b.EndingType == ENDING_START || b.EndingType == ENDING_END {
		e := blNode.CreateNode("ending")
		e.SetAttributeValue("number", fmt.Sprint(b.Ending))
		SetAttributeAlt(e, "type", b.EndingType == ENDING_START, "start", "stop")
	}
	if b.Before > 0 {
		rp := blNode.CreateNode("repeat")
		rp.SetAttributeValue("direction", "backward")
	}
	if b.After > 0 {
		rp := blNode.CreateNode("repeat")
		rp.SetAttributeValue("direction", "forward")
	}
}
func (b *tChord) Gen(x *xml.Node) {
	h := x.CreateNode("harmony")
	r := h.CreateNode("root")
	rs := r.CreateNode("root-step")
	rs.Text = b.Root
	if b.Alter != 0 {
		a := r.CreateNode("root-alter")
		a.Text = fmt.Sprint(b.Alter)
	}
	k := h.CreateNode("kind")
	k.Text = b.Kind
}
func (b *tuplet) Gen(x *xml.Node) {

}

func (a *aTime) Gen(x *xml.Node) {
	tempo := x.CreateNode("time")
	if a.Symbol != "" {
		tempo.SetAttributeValue("symbol", a.Symbol)
	}
	beats := tempo.CreateNode("beats")
	beats.Text = a.StrBeats
	beatType := tempo.CreateNode("beat-type")
	beatType.Text = fmt.Sprint(a.BeatType)

}

func (t *tuplet) Decorate(x *xml.Node) {
	tm := x.CreateNode("time-modification")
	an := tm.CreateNode("actual-notes")
	nn := tm.CreateNode("normal-notes")
	an.Text = fmt.Sprint(t.n0)
	nn.Text = fmt.Sprint(t.n1)
}

// Note ************************************************************************
func (bNote *baseNote) GenBase(noteNode *xml.Node) {
	if bNote.IsRest {
		noteNode.CreateNode("rest")
	} else {
		pitch := noteNode.CreateNode("pitch")
		step := pitch.CreateNode("step")
		step.Text = bNote.Step
		if bNote.Alter != 0 || bNote.HasModifier {
			alter := pitch.CreateNode("alter")
			alter.Text = fmt.Sprint(bNote.Alter)
		}
		octave := pitch.CreateNode("octave")
		octave.Text = fmt.Sprint(bNote.Octave + 4)
	}

}

var accidentalValues = map[int]string{
	-2: "flat-flat",
	-1: "flat",
	0:  "natural",
	1:  "sharp",
	2:  "sharp-sharp",
}

func (bNote *baseNote) AddType(noteNode *xml.Node) {
	voice := noteNode.CreateNode("voice")
	voice.Text = "1"

	typ := noteNode.CreateNode("type")
	typ.Text = noteTypes[bNote.IType]
	for i := 0; i < bNote.DotCount; i++ {
		noteNode.CreateNode("dot")
	}
	if mod, ok := accidentalValues[bNote.Alter]; ok && bNote.HasModifier {
		acc := noteNode.CreateNode("accidental")
		acc.Text = mod
	}

}
func (n *graceNote) Gen(x *xml.Node) {
	noteNode := x.CreateNode("note")
	noteNode.CreateNode("grace")
	n.GenBase(noteNode)
	n.AddType(noteNode)

}
func (n *note) Gen(x *xml.Node) {
	noteNode := x.CreateNode("note")
	if n.Chord {
		noteNode.CreateNode("chord")
	}
	n.GenBase(noteNode)
	if n.TieStop {
		noteNode.CreateNode("tie").SetAttributeValue("type", "stop")
	}

	dur := noteNode.CreateNode("duration")
	dur.Text = fmt.Sprint(n.Duration)
	if n.TieStart {
		noteNode.CreateNode("tie").SetAttributeValue("type", "start")
	}

	n.AddType(noteNode)
	if n.tuplet != nil {
		n.tuplet.Decorate(noteNode)
	}
	if n.Beam != BEAM_NONE {
		bv := []string{"none", "begin", "continue", "end"}
		bn := noteNode.CreateNode("beam")
		bn.SetAttributeValue("number", "1")
		bn.Text = bv[n.Beam]
	}

	n.GenNotation(noteNode)

	if n.Lyric != "" {
		l := noteNode.CreateNode("lyric").SetAttributeValue("default-y", "-50")
		text := l.CreateNode("text")
		text.Text = n.Lyric
	}

}

// Measure *********************************************************************
func addText(x *xml.Node, txt string) {
	d := x.CreateNode("direction").SetAttributeValue("placement", "bellow")
	dt := d.CreateNode("direction-type")
	w := dt.CreateNode("words").SetAttributeValue("font-size", "large")
	w.SetAttributeValue("default-y", "-20")
	w.Text = txt
}

func (m *measure) Gen(x *xml.Node) {
	if m.IsEmpty() {
		return
	}
	mn := x.CreateNode("measure")
	mn.SetAttributeValue("number", fmt.Sprint(m.Index))

	if m.NewLine {
		nl := mn.CreateNode("print")
		nl.SetAttributeValue("new-system", "yes")
	}
	if m.UnderlineText != "" {
		addText(mn, m.UnderlineText)

	}
	if len(m.Attributes) > 0 {
		ma := mn.CreateNode("attributes")
		for _, a := range m.Attributes {
			a.Gen(ma)
		}
	}
	if m.leftBar != nil {
		m.leftBar.Gen(mn)
	}

	for _, c := range m.Content {
		c.Gen(mn)
	}
	if m.rightBar != nil {
		m.rightBar.Gen(mn)
	}

}

// *****************************************************************************
func generateXml(p *partition) string {
	doc := xml.NewDocument("score-partwise")
	doc.Directives = append(doc.Directives, pi)
	setTitle(doc.Root, p.Title)

	id := doc.Root.CreateNode("identification")
	encoding := id.CreateNode("encoding")
	encoder := encoding.CreateNode("encoder")
	encoder.Text = "The Merry Encoder"

	supports := encoding.CreateNode("supports")
	supports.SetAttributeValue("attribute", "new-system")
	supports.SetAttributeValue("element", "print")
	supports.SetAttributeValue("type", "yes")
	supports.SetAttributeValue("value", "yes")
	date := encoding.CreateNode("encoding-date")
	date.Text = time.Now().Format("2006-01-02")
	partId := "P1"
	addPartList(doc.Root, partId)
	part := doc.Root.CreateNode("part").SetAttributeValue("id", partId)

	if len(p.Measures) > 0 {
		p.Measures[0].NewLine = false
	}

	for _, m := range p.Measures {
		m.Gen(part)
	}

	return doc.XMLPretty()
}
