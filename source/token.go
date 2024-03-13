package tokenize

import (
	"fmt"
	"os"
	"os/exec"
)

type Token string

type tokenizerNodePtr struct {
	Ptr    *tokenizerNode
	IsLink bool
}

type tokenizerNode struct {
	Next map[rune]tokenizerNodePtr

	SuffLink        *tokenizerNode
	ClosestTerminal *tokenizerNode
	Terminal        bool
}

func (n *tokenizerNode) Go(t *tokenizer, r rune) *tokenizerNode {
	if n == t.SuperRoot {
		return t.Root
	}
	if n.Next[r].Ptr == nil {
		n.Next[r] = tokenizerNodePtr{n.SuffLink.Go(t, r), true}
	}

	return n.Next[r].Ptr
}

type tokenizer struct {
	Root      *tokenizerNode
	SuperRoot *tokenizerNode
}

func (t *tokenizer) init(tokens []Token) {
	t.Root.SuffLink = t.SuperRoot

	for _, token := range tokens {
		t.insert(token)
	}
}

func (t *tokenizer) insert(token Token) {
	currNode := t.Root
	for _, r := range string(token) {
		if currNode.Next[r].Ptr == nil {
			currNode.Next[r] = tokenizerNodePtr{new(tokenizerNode), false}
		}

		currNode = currNode.Next[r].Ptr
	}

	currNode.Terminal = true
}

func (t *tokenizer) dump(filename string) {
	dumpFile, err := os.Open(filename + ".dot")
	if err != nil {
		panic(err)
	}

	dumped := make(map[*tokenizerNode]struct{})
	var dumpNodes func(n *tokenizerNode)
	dumpNodes = func(n *tokenizerNode) {
		if n == nil {
			return
		}

		lable := ""
		dumped[n] = struct{}{}
		switch n {
		case t.Root:
			lable = "root"
		case t.SuperRoot:
			lable = "super root"
		}

		fmt.Fprintf(dumpFile, "\tstruct%v [lable=\"%v\"];\n", n, lable)
		for _, next := range n.Next {
			if _, ok := dumped[next.Ptr]; ok {
				continue
			}

			dumpNodes(next.Ptr)
		}
	}
	var dumpEdges func(n *tokenizerNode)
	dumpEdges = func(n *tokenizerNode) {
		if n == nil {
			return
		}

		dumped[n] = struct{}{}

		for _, next := range n.Next {
			if _, ok := dumped[next.Ptr]; ok {
				continue
			}

			dumpEdges(next.Ptr)
		}
	}

	fmt.Fprint(dumpFile, "digraph g {\n\t{\n\t\tnode [];\n")
	dumpNodes(t.SuperRoot)
	dumped = make(map[*tokenizerNode]struct{})
	fmt.Fprint(dumpFile, "\t}\n")
	dumpEdges(t.SuperRoot)
	fmt.Fprint(dumpFile, "}\n")

	dumpFile.Close()
	exec.Command("dot", "-Tpng", filename+".dot", ">", filename+".png").Run()
}
