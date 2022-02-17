package httpservice

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kubasiemion/merkledemo/forcehalving"
	"github.com/kubasiemion/merkledemo/tree"
	"github.com/san-lab/commongo/gohttpservice/templates"
)

var InTEE bool

type myHandler struct {
	Renderer *templates.Renderer
}

func NewHandler() *myHandler {
	mh := new(myHandler)
	mh.Renderer = templates.NewRenderer()
	return mh
}

func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	data := new(templates.RenderData)
	path := r.URL.Path[1:]
	data.User, _, _ = r.BasicAuth()
	data.HeaderData = struct {
		User  string
		InTEE bool
	}{data.User, InTEE}

	r.ParseForm()

	switch path {
	case "loadtemplates":
		mh.Renderer.LoadTemplates()
	case "partition":
		partition(data, r)
	default:
		data.TemplateName = "home"
		//data.BodyData = t1{mh.chamber, 0, playercount, playercount}
	}

	mh.Renderer.RenderResponse(w, *data)
	if r.URL.Path[1:] == "EXIT" {
		os.Exit(0)
	}
}

var lastPartition = Partition{}

func partition(data *templates.RenderData, rq *http.Request) {

	pstr := rq.FormValue("partitionstring")
	leafcountstr := rq.FormValue("branchcount")
	branchcount, err := strconv.Atoi(leafcountstr)
	if err != nil {
		branchcount = 2
	}
	reg := regexp.MustCompile(`[; ,]`)
	sint := reg.Split(pstr, -1)
	intset := []int{}
	for _, s := range sint {
		if len(s) > 0 {
			i, e := strconv.Atoi(strings.Trim(s, " \n"))
			if e != nil {
				data.Error = fmt.Sprint(e)
				return
			}
			intset = append(intset, i)
		}

	}
	prt := Partition{Set: intset, PString: pstr}

	proofidxstring := rq.FormValue("getproof")
	if lastPartition.Tree != nil && len(proofidxstring) > 0 {
		idx, err := strconv.Atoi(proofidxstring)
		if err != nil {
			data.Error = fmt.Sprint(err)
			return
		}
		proofTree, err := lastPartition.Tree.GetProof(idx)
		if err != nil {
			data.Error = fmt.Sprint(err)
			return
		}
		fmt.Println("Proof consistent:", tree.VerifyProofConsistency(proofTree))
		prt = lastPartition
		prt.Tree = proofTree

	} else {
		prt.Solution, err = forcehalving.ForceHalve(intset)
		if err != nil {
			data.Error = fmt.Sprint(err)
			return
		}

		if len(prt.Solution) > 0 {
			prev := 0
			prt.Witness = []int{prev}
			for i, s := range prt.Solution {
				prev += s * prt.Set[i]
				prt.Witness = append(prt.Witness, prev)
			}
			prt.ObfuscatedWitness = ObfuscatedWitness(prt.Set, prt.Solution)
			prt.Tree = BuildTree(prt.ObfuscatedWitness, branchcount)
			lastPartition = prt
		}

		prt.Branchcount = branchcount

	}
	data.BodyData = prt

}

type Partition struct {
	PString           string
	Set               []int
	Solution          []int
	Witness           []int
	ObfuscatedWitness []int
	Tree              *tree.Tree
	Proofs            [][][]byte
	Branchcount       int
}

func ObfuscatedWitness(set, solution []int) []int {
	pad := int(rand.Int31n(200))
	sign := 2*rand.Int()&1 - 1
	ret := []int{pad}
	for i := range set {
		pad += set[i] * solution[i] * sign
		ret = append(ret, pad)
	}
	return ret
}

func BuildTree(obWi []int, leafCount int) *tree.Tree {
	bb := [][]byte{}
	for i := range obWi {
		idx := make([]byte, 4)
		binary.BigEndian.PutUint32(idx, uint32(obWi[i]))
		bb = append(bb, idx)
	}
	return tree.NewTree(sha256.New(), bb, leafCount)
}
