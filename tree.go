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

	typ           byte
	pattern       string // origial pattern
	parsedPattern string // pattern, only include regexp/holder's parms
	wildcards     []string
	reg           *regexp.Regexp

	handlers []Handler

	matchNode *route // only for leaf node
}

// tree:
// - ""    // /
// - about // /about
// - api
//   - xxx // /api/xxx
func newTree() *node {
	return newNode(nil, "")
}

func newNode(parent *node, pattern string) *node {
	typ, parsedPattern, wildcards, reg := analyzePattern(pattern)

	return &node{
		parent:        parent,
		typ:           typ,
		pattern:       pattern,
		parsedPattern: parsedPattern,
		wildcards:     wildcards,
		reg:           reg,
	}
}

// getparsedPattern return parsedPattern
// abc -> abc
// *abc -> *abc
// <id:Int> -> id
// <id ~ 70|80> -> id
func getparsedPattern(pattern string) string {
	if strings.HasPrefix(pattern, "*") {
		return pattern // _PATTERN_MATCH_ALL
	}

	if !strings.ContainsAny(pattern, "<>") {
		return pattern // _PATTERN_STATIC
	}

	startIdx := strings.Index(pattern, "<")   //start mark
	endIdx := strings.LastIndex(pattern, ">") //end mark
	if !(startIdx == 0 && endIdx == len(pattern)-1) {
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

	return strings.TrimSpace(pattern[startIdx+1 : closeIdx])
}

// partternRegexp return repexp, and check match count
func partternRegexp(pattern string, want int) (string, error) {
	if want == 0 {
		return "", fmt.Errorf("named args need >= 1")
	}

	startIdx := strings.Index(pattern, "~")
	closeIdx := strings.Index(pattern, ">")

	regExp := strings.TrimSpace(pattern[startIdx+1 : closeIdx])
	if strings.ContainsAny(pattern, "()") && !(strings.Count(regExp, "(") == want && strings.Count(regExp, "(") == strings.Count(regExp, ")")) {
		return "", fmt.Errorf("regexp group count not match want(%d)", want)
	}

	return regExp, nil
}

func analyzePattern(pattern string) (typ byte, parsedPattern string, wildcards []string, reg *regexp.Regexp) {
	if pattern != strings.TrimSpace(pattern) {
		panic(fmt.Sprintf("invalid pattern[%s],it may contain spaces, and so on.", pattern))
	}

	parsedPattern = getparsedPattern(pattern)

	if strings.HasPrefix(parsedPattern, "*") {
		typ = _PATTERN_MATCH_ALL
		parsedPattern = strings.TrimPrefix(parsedPattern, "*")
	} else if strings.Contains(pattern, "<") {
		wildcards = getWildcards(parsedPattern)

		if strings.Contains(pattern, "~") {
			typ = _PATTERN_REGEXP

			regExp, err := partternRegexp(pattern, len(wildcards))
			if err != nil {
				panic(fmt.Sprintf("invalid regexp pattern[%s], err: %s", pattern, err.Error()))
			}
			reg = regexp.MustCompile(regExp)
		} else {
			typ = _PATTERN_HOLDER
		}
	}
	return typ, parsedPattern, wildcards, reg
}

func getWildcards(pattern string) []string {
	ls := strings.Split(pattern, ",")
	for i := range ls {
		ls[i] = strings.TrimSpace(ls[i])
	}

	return ls
}

// --- build tree
// add same type route match order use add order
func (n *node) add(pattern string, handlers []Handler) *node {
	pattern = strings.TrimSuffix(pattern, "/")
	return n.addNextSegment(pattern, handlers)
}

func (n *node) addNextSegment(pattern string, handlers []Handler) *node {
	pattern = strings.TrimPrefix(pattern, "/")

	i := strings.Index(pattern, "/")
	if i == -1 {
		return n.addEndNode(pattern, handlers)
	}

	return n.addSubNode(pattern[:i], pattern[i+1:], handlers)
}

func (n *node) addEndNode(pattern string, handlers []Handler) *node {
	for i := 0; i < len(n.endNodes); i++ {
		if n.endNodes[i].pattern == pattern { // added
			return n.endNodes[i]
		}
	}

	end := newNode(n, pattern)
	end.handlers = handlers

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

	return end
}

func (n *node) addSubNode(segment, pattern string, handlers []Handler) *node {
	for i := 0; i < len(n.subNodes); i++ {
		if n.subNodes[i].pattern == segment {
			return n.subNodes[i].addNextSegment(pattern, handlers)
		}
	}

	sub := newNode(n, segment)

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

	return sub.addNextSegment(pattern, handlers)
}

// --- match uri
func (n *node) Match(uri string) (*node, Params) {
	uri = strings.TrimPrefix(uri, "/")
	uri = strings.TrimSuffix(uri, "/")
	params := make(Params)
	node := n.matchNextSegment(0, uri, params)
	return node, params
}

// globLevel is for _PATTERN_MATCH_ALL route order: 0..n
func (n *node) matchNextSegment(globLevel int, uri string, params Params) *node {
	i := strings.Index(uri, "/")
	if i == -1 {
		return n.matchEndNode(globLevel, uri, params)
	}
	return n.matchSubNode(globLevel, uri[:i], uri[i+1:], params)
}

func (n *node) matchEndNode(globLevel int, uri string, params Params) *node {
	for i := 0; i < len(n.endNodes); i++ {
		switch n.endNodes[i].typ {
		case _PATTERN_STATIC:
			if n.endNodes[i].pattern == uri {
				return n.endNodes[i]
			}
		case _PATTERN_REGEXP:
			results := n.endNodes[i].reg.FindStringSubmatch(uri)
			// Number of results and wildcasrd should be exact same
			if len(results)-1 != len(n.endNodes[i].wildcards) {
				continue
			}

			for j := 0; j < len(n.endNodes[i].wildcards); j++ {
				params[n.endNodes[i].wildcards[j]] = results[j+1]
			}
			return n.endNodes[i]
		case _PATTERN_HOLDER:
			if n.endNodes[i].parsedPattern != "_" {
				params[n.endNodes[i].parsedPattern] = uri
			}
			return n.endNodes[i]
		case _PATTERN_MATCH_ALL:
			if n.endNodes[i].parsedPattern != "" {
				if n.endNodes[i].parsedPattern != "_" {
					params[n.endNodes[i].parsedPattern] = uri
				}
			} else {
				params["*"+strconv.Itoa(globLevel)] = uri
			}

			return n.endNodes[i]
		}
	}

	return nil
}

func (n *node) matchSubNode(globLevel int, segment, uri string, params Params) *node {
	for i := 0; i < len(n.subNodes); i++ {
		switch n.subNodes[i].typ {
		case _PATTERN_STATIC:
			if n.subNodes[i].pattern == segment {
				if end := n.subNodes[i].matchNextSegment(globLevel, uri, params); end != nil {
					return end
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
			if end := n.subNodes[i].matchNextSegment(globLevel, uri, params); end != nil {
				return end
			}
		case _PATTERN_HOLDER:
			if end := n.subNodes[i].matchNextSegment(globLevel, uri, params); end != nil {
				if n.subNodes[i].parsedPattern != "_" {
					params[n.subNodes[i].parsedPattern] = segment
				}
				return end
			}
		case _PATTERN_MATCH_ALL:
			if end := n.subNodes[i].matchNextSegment(globLevel+1, uri, params); end != nil {
				if n.subNodes[i].parsedPattern != "" {
					if n.subNodes[i].parsedPattern != "_" {
						params[n.subNodes[i].parsedPattern] = segment + "/" + uri
					}
				} else {
					params["*"+strconv.Itoa(globLevel)] = segment
				}
				return end
			}
		}
	}

	if len(n.endNodes) > 0 { //for match "/*"
		end := n.endNodes[len(n.endNodes)-1]
		if end.typ == _PATTERN_MATCH_ALL {
			if end.parsedPattern != "" {
				if end.parsedPattern != "_" {
					params[end.parsedPattern] = segment + "/" + uri
				}
			} else {
				params["*"+strconv.Itoa(globLevel)] = segment + "/" + uri
			}
			return end
		}
	}

	return nil
}
