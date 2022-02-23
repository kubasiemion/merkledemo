package tree

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"html/template"
)

type Net struct {
	Nodes template.JS
	Edges template.JS
}

func (tr *Tree) VisNet() Net {
	return Net{tr.VisNodes(), tr.VisEdges()}
}

func visAddWithChildren(p *Node, s *string) {
	if p != nil {
		*s += "{"
		color := "99ffff"
		if p.isLeaf {
			color = "33ff99"
		}
		if len(p.Data) > 0 {
			color = "fee9b4"
		}
		if p.IsRoot {
			color = "ffA050"
		}

		label := hex.EncodeToString(p.Hash)[:5] + "..." + hex.EncodeToString(p.Hash)[59:]
		if p.Data != nil {
			label = fmt.Sprint(int32(binary.BigEndian.Uint32(p.Data))) + "\\n" + label
		}
		*s += fmt.Sprintf("id:\"%s\", label: \"%s\", margin: { top: 15, right: 15, bottom: 15, left: 15 }, color: \"#%s\", heightConstraint:15,",
			p.VisId(), label, color)

		*s += "},"
		for _, c := range p.Children {
			visAddWithChildren(c, s)
		}
	}
}

func (tr *Tree) VisEdges() template.JS {
	s := "["
	arrowsFromCildren(tr.Root, &s)
	s += "]"
	return template.JS(s)
}

func arrowsFromCildren(p *Node, s *string) {
	for _, c := range p.Children {
		if c == nil {
			continue
		}
		*s += fmt.Sprintf("{from: \"%s\", to: \"%s\"},", c.VisId(), p.VisId())
		arrowsFromCildren(c, s)
	}
}

func (n *Node) VisId() string {
	s := fmt.Sprintf("%03d", n.Idx)
	if n.parent != nil {
		s += n.parent.VisId()
	}
	return s
}

func (tr *Tree) VisNodes() template.JS {
	s := "["
	visAddWithChildren(tr.Root, &s)

	s += "]"
	return template.JS(s)
}
