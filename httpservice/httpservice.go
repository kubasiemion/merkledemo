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
	"github.com/san-lab/commongo/gohttpservice"
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
	var err error
	data.SessionID, err = r.Cookie(gohttpservice.SessionIdName)
	fmt.Println(data.SessionID)
	if err != nil {
		fmt.Fprintln(w, err)
		return

	}
	path := r.URL.Path[1:]
	data.User, _, _ = r.BasicAuth()
	data.HeaderData = struct {
		User  string
		InTEE bool
	}{data.User, InTEE}

	r.ParseForm()

	switch path {
	case "loadtemplates":
		mh.Renderer.LoadTemplates("")
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

var partitionMemory = map[string]*Partition{}

func partition(data *templates.RenderData, rq *http.Request) {
	var part *Partition
	if data.SessionID != nil {
		part = partitionMemory[data.SessionID.Value]
	}

	if part == nil {
		part = &Partition{}
	}
	data.BodyData = part
	pstr := rq.FormValue("partitionstring")
	if len(pstr) == 0 {
		pstr = "1,2,3,4,5,6,7"
	}
	leafcountstr := rq.FormValue("branchcount")

	branchcount, err := strconv.Atoi(leafcountstr)

	if err != nil || branchcount == 0 {
		branchcount = 2
	}
	part.Branchcount = branchcount
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
	part.Set = intset
	part.PString = pstr
	part.Proofs = map[int]*tree.Tree{}

	// See if this is about a new proof
	proofidxstring := rq.FormValue("getproof")
	if part.Tree != nil && len(proofidxstring) > 0 {
		idx, err := strconv.Atoi(proofidxstring)
		if err != nil {
			data.Error = fmt.Sprint(err)
			return
		}
		part.DisplayTree, err = part.Tree.GetProof(idx)
		if err != nil {
			data.Error = fmt.Sprint(err)

		}
		part.Proofs[idx] = part.DisplayTree
		return
	}

	//Generate a new tree

	part.Solution, err = forcehalving.ForceHalve(intset)
	if err != nil {
		data.Error = fmt.Sprint(err)
		return
	}

	if len(part.Solution) > 0 {
		prev := 0
		part.Witness = []int{prev}
		for i, s := range part.Solution {
			prev += s * part.Set[i]
			part.Witness = append(part.Witness, prev)
		}
		part.ObfuscatedWitness = ObfuscatedWitness(part.Set, part.Solution)
		part.Tree = BuildTree(part.ObfuscatedWitness, branchcount)

	}
	part.Id = data.SessionID.Value
	partitionMemory[data.SessionID.Value] = part

	part.DisplayTree = part.Tree

}

type Partition struct {
	Id                string
	PString           string
	Set               []int
	Solution          []int
	Witness           []int
	ObfuscatedWitness []int
	Tree              *tree.Tree
	Proofs            map[int]*tree.Tree
	DisplayTree       *tree.Tree
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
