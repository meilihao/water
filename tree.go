package water

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type patternType int8

const (
	_PATTERN_STATIC    patternType = iota // /home
	_PATTERN_REGEXP                       // /<id:int ~ [0-9]+>
	_PATTERN_HOLDER                       // /<user>
	_PATTERN_MATCH_ALL                    // /*
)

type Tree struct {
	parent *Tree

	typ        patternType
	pattern    string
	rawPattern string //处理后的pattern
	wildcards  []string
	reg        *regexp.Regexp

	subtrees []*Tree
	leaves   []*Leaf
}

type Leaf struct {
	parent *Tree

	typ        patternType
	pattern    string
	rawPattern string
	wildcards  []string
	reg        *regexp.Regexp

	handlers []Handler
}

func NewTree() *Tree {
	return NewSubtree(nil, "")
}

func NewSubtree(parent *Tree, pattern string) *Tree {
	typ, rawPattern, wildcards, reg := checkPattern(pattern)
	return &Tree{parent, typ, pattern, rawPattern, wildcards, reg, make([]*Tree, 0, 5), make([]*Leaf, 0, 5)}
}

func NewLeaf(parent *Tree, pattern string, handlers []Handler) *Leaf {
	typ, rawPattern, wildcards, reg := checkPattern(pattern)
	return &Leaf{parent, typ, pattern, rawPattern, wildcards, reg, handlers}
}

func getPartternRegexp(pattern string) string {
	startIdx := strings.Index(pattern, "~")
	closeIdx := strings.Index(pattern, ">")

	return strings.TrimSpace(pattern[startIdx+1 : closeIdx])
}

func getRawPattern(pattern string) string {
	if !strings.ContainsAny(pattern, "<>") {
		return pattern // _PATTERN_STATIC and _PATTERN_MATCH_ALL
	}

	startIdx := strings.Index(pattern, "<") //start mark
	closeIdx := strings.IndexAny(pattern, ":~>")
	endIdx := strings.Index(pattern, ">") //end mark
	if !(startIdx == 0 && endIdx > -1 && closeIdx > -1) {
		panic(fmt.Sprintf("invalid pattern[%s] without correct format.", pattern))
	}

	return "<" + strings.TrimSpace(pattern[startIdx+1:closeIdx]) + ">"
}

func checkPattern(pattern string) (typ patternType, rawPattern string, wildcards []string, reg *regexp.Regexp) {
	if pattern != strings.TrimSpace(pattern) {
		panic(fmt.Sprintf("invalid pattern[%s],it may contain spaces, and so on. ", pattern))
	}

	rawPattern = getRawPattern(pattern)

	if pattern == "*" {
		typ = _PATTERN_MATCH_ALL
	} else if strings.Contains(pattern, "<") {
		wildcards = make([]string, 0, 1)
		wildcards = append(wildcards, strings.TrimPrefix(strings.TrimSuffix(rawPattern, ">"), "<"))

		if strings.Contains(pattern, "~") {
			typ = _PATTERN_REGEXP

			reg = regexp.MustCompile(getPartternRegexp(pattern))
		} else {
			typ = _PATTERN_HOLDER
		}
	}
	return typ, rawPattern, wildcards, reg
}

func (t *Tree) Add(pattern string, handlers []Handler) {
	t.addNextSegment(pattern, handlers)
}

func (t *Tree) addNextSegment(pattern string, handlers []Handler) {
	pattern = strings.TrimPrefix(pattern, "/")

	i := strings.Index(pattern, "/")
	if i == -1 { //(""||"/xxx")
		t.addLeaf(pattern, handlers)
		return
	}
	t.addSubtree(pattern[:i], pattern[i+1:], handlers)
}

func (t *Tree) addLeaf(pattern string, handlers []Handler) {
	rawPattern := getRawPattern(pattern)
	for i := 0; i < len(t.leaves); i++ {
		if t.leaves[i].rawPattern == rawPattern { //added
			return
		}
	}

	leaf := NewLeaf(t, pattern, handlers)
	i := 0
	for ; i < len(t.leaves); i++ {
		if leaf.typ < t.leaves[i].typ {
			break
		}
	}

	if i == len(t.leaves) {
		t.leaves = append(t.leaves, leaf)
	} else {
		t.leaves = append(t.leaves[:i], append([]*Leaf{leaf}, t.leaves[i:]...)...)
	}
}

func (t *Tree) addSubtree(segment, pattern string, handlers []Handler) {
	rawSegment := getRawPattern(segment)
	for i := 0; i < len(t.subtrees); i++ {
		if t.subtrees[i].rawPattern == rawSegment {
			t.subtrees[i].addNextSegment(pattern, handlers)
			return
		}
	}

	subtree := NewSubtree(t, segment)
	i := 0
	for ; i < len(t.subtrees); i++ {
		if subtree.typ < t.subtrees[i].typ {
			break
		}
	}

	if i == len(t.subtrees) {
		t.subtrees = append(t.subtrees, subtree)
	} else {
		t.subtrees = append(t.subtrees[:i], append([]*Tree{subtree}, t.subtrees[i:]...)...)
	}
	subtree.addNextSegment(pattern, handlers)
}

func (t *Tree) Match(url string) ([]Handler, Params, bool) {
	url = strings.TrimPrefix(url, "/")

	params := make(Params, 0, 3)
	handlers, ok := t.matchNextSegment(0, url, &params)
	return handlers, params, ok
}

func (t *Tree) matchNextSegment(globLevel int, url string, params *Params) ([]Handler, bool) {
	i := strings.Index(url, "/")
	if i == -1 {
		return t.matchLeaf(globLevel, url, params)
	}
	return t.matchSubtree(globLevel, url[:i], url[i+1:], params)
}

func (t *Tree) matchLeaf(globLevel int, url string, params *Params) ([]Handler, bool) {
	for i := 0; i < len(t.leaves); i++ {
		switch t.leaves[i].typ {
		case _PATTERN_STATIC:
			if t.leaves[i].rawPattern == url {
				return t.leaves[i].handlers, true
			}
		case _PATTERN_REGEXP:
			if t.leaves[i].reg.MatchString(url) {
				*params = append(*params, Param{Name: t.leaves[i].wildcards[0], Value: url})
				return t.leaves[i].handlers, true
			}
		case _PATTERN_HOLDER:
			*params = append(*params, Param{Name: t.leaves[i].wildcards[0], Value: url})
			return t.leaves[i].handlers, true
		case _PATTERN_MATCH_ALL:
			*params = append(*params, Param{Name: "*", Value: url})
			*params = append(*params, Param{Name: "*" + strconv.Itoa(globLevel), Value: url})
			return t.leaves[i].handlers, true
		}
	}
	return nil, false
}

func (t *Tree) matchSubtree(globLevel int, segment, url string, params *Params) ([]Handler, bool) {
	for i := 0; i < len(t.subtrees); i++ {
		switch t.subtrees[i].typ {
		case _PATTERN_STATIC:
			if t.subtrees[i].rawPattern == segment {
				if handlers, ok := t.subtrees[i].matchNextSegment(globLevel, url, params); ok {
					return handlers, true
				}
			}
		case _PATTERN_REGEXP:
			if t.subtrees[i].reg.MatchString(segment) {
				*params = append(*params, Param{Name: t.subtrees[i].wildcards[0], Value: segment})

				if handlers, ok := t.subtrees[i].matchNextSegment(globLevel, url, params); ok {
					return handlers, true
				}
			}
		case _PATTERN_HOLDER:
			if handlers, ok := t.subtrees[i].matchNextSegment(globLevel+1, url, params); ok {
				*params = append(*params, Param{Name: t.subtrees[i].wildcards[0], Value: segment})
				return handlers, true
			}
		case _PATTERN_MATCH_ALL:
			if handlers, ok := t.subtrees[i].matchNextSegment(globLevel+1, url, params); ok {
				*params = append(*params, Param{Name: "*" + strconv.Itoa(globLevel), Value: segment})
				return handlers, true
			}
		}
	}

	if len(t.leaves) > 0 { //for match "/*"
		leaf := t.leaves[len(t.leaves)-1]
		if leaf.typ == _PATTERN_MATCH_ALL {
			*params = append(*params, Param{Name: "*", Value: segment + "/" + url})
			*params = append(*params, Param{Name: "*" + strconv.Itoa(globLevel), Value: segment + "/" + url})
			return leaf.handlers, true
		}
	}
	return nil, false
}
