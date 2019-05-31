module github.com/pkedy/wasmtest

require (
	github.com/go-interpreter/wagon v0.4.0
	github.com/kr/pretty v0.1.0 // indirect
	github.com/perlin-network/life v0.0.0-20190521143330-57f3819c2df0
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/wasmerio/go-ext-wasm v0.0.0-20190529183953-15a798de9af3
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	google.golang.org/appengine v1.6.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/perlin-network/life v0.0.0-20190521143330-57f3819c2df0 => github.com/pkedy/life v0.0.0-20190530202850-c437dabbc556
