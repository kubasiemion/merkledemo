package tree

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
)

type Node struct {
	isLeaf   bool
	Data     []byte
	Children []*Node
	parent   *Node

	Hash    []byte
	Idx     int //position as sibilng, here 0 .. branchCount-1
	Lvl     int
	IsRoot  bool
	IsWrong bool
}

type Tree struct {
	Hash        hash.Hash
	Root        *Node
	BranchCount int
	Leaves      []*Node
}

func nodeWithChildrenString(nd *Node, i int) string {
	if nd == nil {
		return ""
	}
	s := "" //fmt.Sprintf("i:%v::%s\n", i, nd)
	for j, c := range nd.Children {

		s += fmt.Sprintf("i:%v::%s\n", j, c)
	}
	for j, c := range nd.Children {
		s += nodeWithChildrenString(c, j)
	}
	return s

}

func (nd *Node) String() string {
	s := fmt.Sprintf("level: %v index: %v isLeaf:%v children: %v\n", nd.Lvl, nd.Idx, nd.isLeaf, len(nd.Children))

	s += fmt.Sprintf(" %v %v\n", nd.Data, hex.EncodeToString(nd.Hash))
	return s
}

func (tr *Tree) String() string {
	s := tr.Root.String()
	s += nodeWithChildrenString(tr.Root, 0)

	return s
}

//https://graphics.stanford.edu/~seander/bithacks.html#RoundUpPowerOf2
func nextPowerOf2(i int) int { //int32 actually
	if i == 0 {
		return 1
	}
	i--
	i |= i >> 1
	i |= i >> 2
	i |= i >> 4
	i |= i >> 8
	i |= i >> 16
	i++
	return i

}

func NewTree(hf hash.Hash, data [][]byte, branchCount int) *Tree {
	if branchCount < 2 {
		branchCount = 2
	}

	if data == nil {
		data = [][]byte{{0}}
	}

	leaves := make([]*Node, 0)

	tr := new(Tree)
	tr.BranchCount = branchCount //nextPowerOf2(branchCount)
	tr.Hash = hf
	for i := 0; i < len(data); i++ {
		nod := new(Node)
		nod.isLeaf = true
		nod.Data = data[i]
		tr.Hash.Reset()
		tr.Hash.Write(data[i])
		nod.Hash = tr.Hash.Sum(nil)
		nod.Lvl = 0
		//nod.Idx = i % branchCount
		leaves = append(leaves, nod)

	}
	tr.Leaves = leaves

	tr.makeParents(leaves, 1)
	return tr
}

func (tr *Tree) makeParents(level []*Node, lvl int) {
	p := len(level)
	if p == 1 {
		tr.Root = level[0]
		tr.Root.IsRoot = true
		return
	}

	nextlvl := make([]*Node, 0)
	for i := 0; i < p; {
		parent := new(Node)
		parent.Lvl = lvl
		parent.Children = make([]*Node, 0)
		tr.Hash.Reset()
		for j := 0; i < p && j < tr.BranchCount; j++ {
			child := level[i]
			if child != nil {
				child.Idx = j
				tr.Hash.Write(child.Hash)
				child.parent = parent

			}

			parent.Children = append(parent.Children, child)
			i++
		}

		parent.Hash = tr.Hash.Sum(nil)
		nextlvl = append(nextlvl, parent)

	}

	tr.makeParents(nextlvl, lvl+1)

}

func cloneNode(org *Node) *Node {
	cp := new(Node)
	cp.Hash = org.Hash
	cp.Lvl = org.Lvl
	cp.Idx = org.Idx
	return cp
}

func (tr *Tree) GetProof(i int) (*Tree, error) {
	fmt.Println("Getting proof for element", i)

	proof := new(Tree)
	provednode := tr.Leaves[i]

	shadownode := cloneNode(provednode)
	shadownode.Data = provednode.Data
	shadownode.isLeaf = true

	root := GetPathToRoot(provednode, shadownode)

	proof.Hash = tr.Hash
	proof.Root = root

	return proof, nil
}

func GetPathToRoot(org *Node, cp *Node) *Node {

	if org.parent == nil {
		cp.IsRoot = true

		return cp
	}
	cppar := cloneNode(org.parent)

	for i, c := range org.parent.Children {
		if c == nil {
			continue
		}

		if i == cp.Idx {

			cppar.Children = append(cppar.Children, cp)
			cp.parent = cppar
		} else {
			cc := cloneNode(c)
			cppar.Children = append(cppar.Children, cc)
			cc.isLeaf = true
			cc.parent = cppar

		}

	}

	return GetPathToRoot(org.parent, cppar)

}

func VerifyProofConsistency(proof *Tree) bool {
	return IsSubtreeConsistent(proof.Root, proof.Hash)

}

func IsSubtreeConsistent(nd *Node, hash hash.Hash) bool {
	if nd.isLeaf || nd.Data != nil {
		if nd.Data != nil {
			hash.Reset()
			hash.Write(nd.Data)
			ht := hash.Sum(nil)
			loc := bytes.Equal(ht, nd.Hash)
			if !loc {
				fmt.Println("Wrong leaf!")
				return loc
			}
		}
		return true
	}
	hash.Reset()
	for _, c := range nd.Children {
		if c != nil {
			hash.Write(c.Hash)
		}
	}
	if !bytes.Equal(hash.Sum(nil), nd.Hash) {
		fmt.Println("Wrong node", nd)
		return false
	}
	for _, c := range nd.Children {
		if !IsSubtreeConsistent(c, hash) {
			return false
		}
	}
	return true
}

func (tr *Tree) NiceRoot() string {
	return hex.EncodeToString(tr.Root.Hash)
}
