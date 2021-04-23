package water

import (
	"fmt"
	"testing"
)

func TestRouterWithGinStype(t *testing.T) {
	router := Default()

	v1 := router.Group("/v1", func(c *Context) {
		fmt.Println("/v1中间件")
	})
	{
		v1.POST("/login", _t)

		v2 := v1.Group("/v2")
		{
			v2.POST("/login", _t)
		}
	}

	w := router.Handler()

	fmt.Println("\n\n", "Raw Router Tree:")
	w.PrintRawRouter()
}

func _t(c *Context) {
}
