package proxy

import "net/http"

type Plugin struct {
	Entries []CallEntry
}

type Handler func()

type CallEntry struct {
	Name string
	Handler Handler
}

var P = Plugin{
	[]CallEntry{
		{"dfsdf", func() {

		}},
	},
}

func buildServer() *http.Server {
	return &http.Server{
		Addr: "",
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			req.GetBody()
		}),
	}
}

func Serve() {
	http.
}