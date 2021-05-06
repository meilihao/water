// +build !extended

package water

import (
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter
	//http.CloseNotifier
	//http.Flusher
	//http.Hijacker
}
