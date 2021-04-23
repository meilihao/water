package water

import (
	"fmt"
	"sort"
	"strings"
)

var (
	_PREFIX_BRANCH = "├──"  //树枝
	_PREFIX_TRUNK  = "│   " //树干
	_PREFIX_LEAF   = "└──"  //叶子
	_PREFIX_EMP    = "    " //空
)

func printRawRoute(prefix string, node *Router) {
	if node.method == "" {
		fmt.Printf("%s %s\n", prefix, node.pattern)
	} else {
		fmt.Printf("%s %s [%-7s : %d]\n", prefix, node.pattern, node.method, countHandlersForRawRouter(node))
	}
}

// 在e.rootRouter上递归统计len(handlers)
func countHandlersForRawRouter(r *Router) int {
	if r.parent != nil {
		return len(r.befores) + len(r.handlers) + countHandlersForRawRouter(r.parent)
	}

	return len(r.befores) + len(r.handlers)
}

func printRawRouter(nodes []*Router, prefix string) {
	if prefix != "" {
		prefix = strings.Replace(prefix, _PREFIX_LEAF, _PREFIX_EMP, -1)
	}

	for i, n := 0, len(nodes); i < n; i++ {
		if i == n-1 { // leaf
			printRawRoute(prefix+_PREFIX_LEAF, nodes[i])

			if len(nodes[i].sub) > 0 {
				printRawRouter(nodes[i].sub, prefix+_PREFIX_LEAF)
			}
		} else { //树枝
			printRawRoute(prefix+_PREFIX_BRANCH, nodes[i])

			if len(nodes[i].sub) > 0 {
				printRawRouter(nodes[i].sub, prefix+_PREFIX_TRUNK)
			}
		}
	}
}

func (e *Engine) PrintRawRouter() {
	if e.rootRouter == nil {
		panic("print router: no raw router.")
	}

	printRawRouter(e.rootRouter.sub, "")
}

// print routes by method
// order by uri
func (e *Engine) PrintRawRoutes(method string) {
	method, _ = checkMethod(method)
	routes := e.routeStore.routeMap[method]
	if len(routes) == 0 {
		fmt.Printf("%s\n", "no route")
		return
	}

	list := make([]string, 0, len(routes))
	for k := range routes {
		list = append(list, k)
	}

	sort.Strings(list)

	for _, v := range list {
		route := routes[v]

		// count(route.handlers) + uri
		fmt.Printf("(%2d) %s\n", len(route.handlers), v)
	}
}

// order by add router order
func (e *Engine) PrintRawAllRoutes() {
	for _, v := range e.routeStore.routeSlice {
		// count(router.handlers) + uri
		fmt.Printf("(%7s) %s : %d\n", v.method, v.uri, len(v.handlers))
	}
}

// print release router tree by method
// len(tree.handlers) includes middleware
// TODO print handler name in []handler
func (e *Engine) PrintRouterTree(method string) {
	_, idx := checkMethod(method)
	tree := e.routers[idx]
	if tree == nil {
		fmt.Printf("%s\n", "no route")
		return
	}

	fmt.Printf("%s [%d]\n", "/", len(tree.handlers))
	printTreeNode(tree, "")
}

func printNode(prefix string, node *node, isLeaf bool) {
	if isLeaf {
		fmt.Printf("%s %s [%d]\n", prefix, node.pattern, len(node.handlers))
	} else {
		fmt.Printf("%s %s\n", prefix, node.pattern)
	}
}

func printTreeNode(node *node, prefix string) {
	if prefix != "" {
		prefix = strings.Replace(prefix, _PREFIX_LEAF, _PREFIX_EMP, -1)
	}

	n_ends := len(node.endNodes)
	for i, n := 0, len(node.subNodes); i < n; i++ {
		if i == n-1 {
			if n_ends > 0 {
				printNode(prefix+_PREFIX_BRANCH, node.subNodes[i], false)

				printTreeNode(node.subNodes[i], prefix+_PREFIX_TRUNK)
			} else {
				printNode(prefix+_PREFIX_LEAF, node.subNodes[i], false)

				printTreeNode(node.subNodes[i], prefix+_PREFIX_EMP)
			}
		} else {
			printNode(prefix+_PREFIX_BRANCH, node.subNodes[i], false)

			printTreeNode(node.subNodes[i], prefix+_PREFIX_TRUNK)
		}
	}
	for i, n := 0, len(node.endNodes); i < n; i++ {
		if i == n-1 { // leaf
			printNode(prefix+_PREFIX_LEAF, node.endNodes[i], true)
		} else {
			printNode(prefix+_PREFIX_BRANCH, node.endNodes[i], true)
		}
	}
}
