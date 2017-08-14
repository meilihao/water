package benchmark

import (
	"net/http/httptest"
	"testing"

	"github.com/meilihao/water"
)

func BenchmarkDirectCall(b *testing.B) {
	var p = &water.Context{}
	p.ResponseWriter = httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		p.ResponseWriter.Header().Set(water.HeaderContentType, water.MIMEApplicationXML)
	}
}

func BenchmarkEmbedCall(b *testing.B) {
	var p = &water.Context{}
	p.ResponseWriter = httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		p.Header().Set(water.HeaderContentType, water.MIMEApplicationXML)
	}
}

// EmbedCall ≈ DirectCall, EmbedCall is better.
// ⋊> ~/g/g/s/g/m/w/a on master ⨯ go test -bench=.                                                                                                                                    18:02:49
// goos: linux
// goarch: amd64
// pkg: github.com/meilihao/water/a
// BenchmarkDirectCall-4   	20000000	        73.2 ns/op
// BenchmarkEmbedCall-4    	20000000	        71.8 ns/op
// PASS
// ok  	github.com/meilihao/water/a	3.057s
// ⋊> ~/g/g/s/g/m/w/a on master ⨯ go test -bench=.                                                                                                                                    18:03:16
// goos: linux
// goarch: amd64
// pkg: github.com/meilihao/water/a
// BenchmarkDirectCall-4   	20000000	        70.7 ns/op
// BenchmarkEmbedCall-4    	20000000	        69.1 ns/op
// PASS
// ok  	github.com/meilihao/water/a	2.953s
// ⋊> ~/g/g/s/g/m/w/a on master ⨯ go test -bench=.                                                                                                                                    18:03:24
// goos: linux
// goarch: amd64
// pkg: github.com/meilihao/water/a
// BenchmarkDirectCall-4   	20000000	        71.3 ns/op
// BenchmarkEmbedCall-4    	20000000	        71.9 ns/op
// PASS
// ok  	github.com/meilihao/water/a	3.033s
// ⋊> ~/g/g/s/g/m/w/a on master ⨯ go test -bench=.                                                                                                                                    18:03:31
// goos: linux
// goarch: amd64
// pkg: github.com/meilihao/water/a
// BenchmarkDirectCall-4   	20000000	        72.5 ns/op
// BenchmarkEmbedCall-4    	20000000	        71.5 ns/op
// PASS
// ok  	github.com/meilihao/water/a	3.053s
