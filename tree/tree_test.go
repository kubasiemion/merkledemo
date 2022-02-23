package tree

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestNewTree(t *testing.T) {
	fmt.Println("testing...")
	data := [][]byte{{0, 0, 0, 1}, {0, 0, 0, 2}, nil, {0, 0, 0, 3}, {0, 0, 0, 7}}
	hf := sha256.New()
	tr := NewTree(hf, data, 2)
	fmt.Println(tr.VisNodes())
	fmt.Println()
	//fmt.Println(tr.VisEdges())
	fmt.Println(tr)

	fmt.Println("root", hex.EncodeToString(tr.Root.Hash))

	proof, _ := tr.GetProof(3)
	fmt.Println("Consistent:", VerifyProofConsistency(proof))

}
