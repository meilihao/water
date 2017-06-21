package water

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	_PATTERN_STATIC    byte = iota // /home
	_PATTERN_REGEXP                // /<id:int ~ [0-9]+>
	_PATTERN_HOLDER                // /<user>
	_PATTERN_MATCH_ALL             // /*
)

type node struct {
	parent   *node
	subNodes []*node
	endNodes []*node

	typ        byte
	pattern    string // origial pattern
	rawPattern string // pattern, not include : regexp,holder
	wildcards  []string
	reg        *regexp.Regexp

	handlers []Handler
}

func newTree() *node {
	return newNode(nil, "/", nil)
}

func newNode(parent *node, pattern string, handlers []Handler) *node {
	typ, rawPattern, wildcards, reg := analyzePattern(pattern)

	return &node{
		parent:     parent,
		typ:        typ,
		pattern:    pattern,
		rawPattern: rawPattern,
		wildcards:  wildcards,
		reg:        reg,
		handlers:   handlers,
	}
}

func getRawPattern(pattern string) string {
	if !strings.ContainsAny(pattern, "<>") {
		return pattern // _PATTERN_STATIC or _PATTERN_MATCH_ALL
	}

	startIdx := strings.Index(pattern, "<")   //start mark
	endIdx := strings.LastIndex(pattern, ">") //end mark
	if !(startIdx == 0 && endIdx > -1) {
		panic(fmt.Sprintf("invalid pattern[%s] without correct format.", pattern))
	}

	closeIdx := endIdx
	typeStartIdx := strings.Index(pattern, ":")
	if typeStartIdx > -1 {
		closeIdx = typeStartIdx
	}
	regStartIdx := strings.Index(pattern, "~")
	if typeStartIdx == -1 && regStartIdx > -1 {
		closeIdx = regStartIdx
	}

	return "<" + strings.TrimSpace(pattern[startIdx+1:closeIdx]) + ">"
}

func partternRegexp(pattern string) string {
	startIdx := strings.Index(pattern, "~")
	closeIdx := strings.Index(pattern, ">")

	return strings.TrimSpace(pattern[startIdx+1 : closeIdx])
}

func analyzePattern(pattern string) (typ byte, rawPattern string, wildcards []string, reg *regexp.Regexp) {
	if pattern != strings.TrimSpace(pattern) {
		panic(fmt.Sprintf("invalid pattern[%s],it may contain spaces, and so on. ", pattern))
	}

	rawPattern = getRawPattern(pattern)

	if pattern == "*" {
		typ = _PATTERN_MATCH_ALL
	} else if strings.Contains(pattern, "<") {
		wildcards = strings.Split(strings.TrimPrefix(strings.TrimSuffix(rawPattern, ">"), "<"), "+")

		if strings.Contains(pattern, "~") {
			typ = _PATTERN_REGEXP

			reg = regexp.MustCompile(partternRegexp(pattern))
		} else {
			typ = _PATTERN_HOLDER
		}
	}
	return typ, rawPattern, wildcards, reg
}

// --- build tree

func (n *node) add(pattern string, handlers []Handler) {
	if pattern == "/" {
		if len(n.handlers) != 0 {
			panic("root's node had set.")
		}

		n.handlers = handlers
		return
	}

	pattern = strings.TrimSuffix(pattern, "/")
	n.addNextSegment(pattern, handlers)
}

func (n *node) addNextSegment(pattern string, handlers []Handler) {
	pattern = strings.TrimPrefix(pattern, "/")

	i := strings.Index(pattern, "/")
	if i == -1 {
		n.addEndNode(pattern, handlers)
		return
	}
	n.addSubNode(pattern[:i], pattern[i+1:], handlers)
}

func (n *node) addEndNode(pattern string, handlers []Handler) {
	rawPattern := getRawPattern(pattern)
	for i := 0; i < len(n.endNodes); i++ {
		if n.endNodes[i].rawPattern == rawPattern { // added
			return
		}
	}

	end := newNode(n, pattern, handlers)

	i := 0
	for ; i < len(n.endNodes); i++ {
		if end.typ < n.endNodes[i].typ {
			break
		}
	}

	if i == len(n.endNodes) {
		n.endNodes = append(n.endNodes, end)
	} else {
		n.endNodes = append(n.endNodes[:i], append([]*node{end}, n.endNodes[i:]...)...)
	}
}

func (n *node) addSubNode(segment, pattern string, handlers []Handler) {
	rawSegment := getRawPattern(segment)
	for i := 0; i < len(n.subNodes); i++ {
		if n.subNodes[i].rawPattern == rawSegment {
			n.subNodes[i].addNextSegment(pattern, handlers)
			return
		}
	}

	sub := newNode(n, segment, nil)

	i := 0
	for ; i < len(n.subNodes); i++ {
		if sub.typ < n.subNodes[i].typ {
			break
		}
	}

	if i == len(n.subNodes) {
		n.subNodes = append(n.subNodes, sub)
	} else {
		n.subNodes = append(n.subNodes[:i], append([]*node{sub}, n.subNodes[i:]...)...)
	}

	sub.addNextSegment(pattern, handlers)
}

// --- match uri

func (n *node) Match(uri string) ([]Handler, Params, bool) {
	// no method
	if n == nil {
		return nil, nil, false
	}

	if uri == "/" {
		return n.handlers, nil, true
	}

	uri = strings.TrimPrefix(uri, "/")
	uri = strings.TrimSuffix(uri, "/")
	params := make(Params)
	handle, ok := n.matchNextSegment(0, uri, params)
	return handle, params, ok
}

func (n *node) matchNextSegment(globLevel int, uri string, params Params) ([]Handler, bool) {
	i := strings.Index(uri, "/")
	if i == -1 {
		return n.matchEndNode(globLevel, uri, params)
	}
	return n.matchSubNode(globLevel, uri[:i], uri[i+1:], params)
}

func (n *node) matchEndNode(globLevel int, uri string, params Params) ([]Handler, bool) {
	for i := 0; i < len(n.endNodes); i++ {
		switch n.endNodes[i].typ {
		case _PATTERN_STATIC:
			if n.endNodes[i].pattern == uri {
				return n.endNodes[i].handlers, true
			}
		case _PATTERN_REGEXP:
			results := n.endNodes[i].reg.FindStringSubmatch(uri)
			if len(results)-1 != len(n.endNodes[i].wildcards) {
				continue
			}

			for j := 0; j < len(n.endNodes[i].wildcards); j++ {
				params[n.endNodes[i].wildcards[j]] = results[j+1]
			}
			return n.endNodes[i].handlers, true
		case _PATTERN_HOLDER:
			params[n.endNodes[i].wildcards[0]] = uri
			return n.endNodes[i].handlers, true
		case _PATTERN_MATCH_ALL:
			params["*"] = uri
			params["*"+strconv.Itoa(globLevel)] = uri
			return n.endNodes[i].handlers, true
		}
	}
	return nil, false
}

func (n *node) matchSubNode(globLevel int, segment, uri string, params Params) ([]Handler, bool) {
	for i := 0; i < len(n.subNodes); i++ {
		switch n.subNodes[i].typ {
		case _PATTERN_STATIC:
			if n.subNodes[i].pattern == segment {
				if handlers, ok := n.subNodes[i].matchNextSegment(globLevel, uri, params); ok {
					return handlers, true
				}
			}
		case _PATTERN_REGEXP:
			results := n.subNodes[i].reg.FindStringSubmatch(segment)
			if len(results)-1 != len(n.subNodes[i].wildcards) {
				continue
			}

			for j := 0; j < len(n.subNodes[i].wildcards); j++ {
				params[n.subNodes[i].wildcards[j]] = results[j+1]
			}
			if handle, ok := n.subNodes[i].matchNextSegment(globLevel, uri, params); ok {
				return handle, true
			}
		case _PATTERN_HOLDER:
			if handlers, ok := n.subNodes[i].matchNextSegment(globLevel+1, uri, params); ok {
				params[n.subNodes[i].wildcards[0]] = segment
				return handlers, true
			}
		case _PATTERN_MATCH_ALL:
			if handlers, ok := n.subNodes[i].matchNextSegment(globLevel+1, uri, params); ok {
				params["*"+strconv.Itoa(globLevel)] = segment
				return handlers, true
			}
		}
	}

	if len(n.endNodes) > 0 { //for match "/*"
		end := n.endNodes[len(n.endNodes)-1]
		if end.typ == _PATTERN_MATCH_ALL {
			params["*"] = segment + "/" + uri
			params["*"+strconv.Itoa(globLevel)] = segment + "/" + uri
			return end.handlers, true
		}
	}

	return nil, false
}
