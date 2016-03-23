package water

import (
	"fmt"
	"sync"
)

// Group is used internally to configure router.
// 用于成组地配置路由
type Group struct {
	parent  *Group
	befores []interface{}
	pattern string
	router  *Router
}

func newGroup() *Group {
	return &Group{
		parent:  nil,
		befores: make([]interface{}, 0),
		router:  nil,
	}
}

// Before adds middleware to the group
// 给Group添加中间件
func (g *Group) Before(handlers ...interface{}) {
	g.befores = append(g.befores, handlers...)
}

func (g *Group) Any(pattern string, handlers ...interface{}) {
	g.handle("Any", pattern, handlers)
}

func (g *Group) Get(pattern string, handlers ...interface{}) {
	g.handle("GET", pattern, handlers)
}

func (g *Group) Post(pattern string, handlers ...interface{}) {
	g.handle("POST", pattern, handlers)
}

func (g *Group) Delete(pattern string, handlers ...interface{}) {
	g.handle("DELETE", pattern, handlers)
}

func (g *Group) Put(pattern string, handlers ...interface{}) {
	g.handle("PUT", pattern, handlers)
}

func (g *Group) Patch(pattern string, handlers ...interface{}) {
	g.handle("PATCH", pattern, handlers)
}

func (g *Group) Options(pattern string, handlers ...interface{}) {
	g.handle("OPTIONS", pattern, handlers)
}

func (g *Group) Head(pattern string, handlers ...interface{}) {
	g.handle("HEAD", pattern, handlers)
}

func (g *Group) Trace(pattern string, handlers ...interface{}) {
	g.handle("TRACE", pattern, handlers)
}

// change group to a route and add the route to Router.
// 将group解析成Route并向Router注册.
func (g *Group) handle(method, pattern string, handlers []interface{}) {
	if !(pattern == "" || checkSplitPattern(pattern)) {
		panic(fmt.Sprintf("invalid g.%s pattern : [%s]", method, pattern))
	}

	patternPath := pattern
	handerChain := make([]interface{}, 0)
	handerChain = append(handerChain, handlers...)

	//recursive group to get a route
	for {
		patternPath = g.pattern + patternPath
		if len(g.befores) > 0 {
			tmpHandlerChain := make([]interface{}, 0, len(g.befores)+len(handerChain))
			tmpHandlerChain = append(tmpHandlerChain, g.befores...)
			tmpHandlerChain = append(tmpHandlerChain, handerChain...)
			handerChain = tmpHandlerChain
		}

		if g.parent != nil {
			g = g.parent
		} else {
			break
		}
	}

	// add the route to Router
	g.router.handle(method, patternPath, handerChain)
}

// add a group to another group
// group嵌套
func (g *Group) Group(pattern string, fn func(*Group)) {
	if !checkSplitPattern(pattern) {
		panic(fmt.Sprintf("invalid g.Group pattern : [%s]", pattern))
	}
	// 保存当前使用的group,避免下面递归导致group变化
	currentGroup := g

	// 递归检查group pattern是否重复
	patternPath := pattern
	for {
		patternPath = g.pattern + patternPath
		if g.parent != nil {
			g = g.parent
		} else {
			break
		}
	}

	if g.router.groupMap.isExist(patternPath) {
		panic(fmt.Sprintf("double g.Group pattern : [%s]", patternPath))
	}
	g.router.groupMap.add(patternPath)

	subGroup := newGroup()
	subGroup.parent = currentGroup
	subGroup.pattern = pattern
	subGroup.router = currentGroup.router

	fn(subGroup)
}

// add a group to Router.
// 向Router注册group
func (r *Router) Group(pattern string, fn func(*Group)) {
	if !(pattern == "/" || checkSplitPattern(pattern)) {
		panic(fmt.Sprintf("invalid r.Group pattern : [%s]", pattern))
	}

	if r.groupMap.isExist(pattern) {
		panic(fmt.Sprintf("double r.Group pattern : [%s]", pattern))
	}
	r.groupMap.add(pattern)

	g := newGroup()
	g.parent = nil
	g.pattern = pattern
	g.router = r

	fn(g)
}

// groupMap represents a thread-safe map for group pattern.
// 用于检查group pattern是否重复(冲突)
type groupMap struct {
	groups map[string]bool
	lock   sync.RWMutex
}

func newGroupMap() *groupMap {
	gm := &groupMap{
		groups: make(map[string]bool),
	}

	return gm
}

func (gm *groupMap) isExist(pattern string) bool {
	gm.lock.RLock()
	defer gm.lock.RUnlock()

	return gm.groups[pattern]
}

func (gm *groupMap) add(pattern string) {
	gm.lock.Lock()
	defer gm.lock.Unlock()

	gm.groups[pattern] = true
}
