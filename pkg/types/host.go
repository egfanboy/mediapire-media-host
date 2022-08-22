package types

type Host interface {
	Scheme() string
	Port() int
	Host() string
}

type HttpHost struct {
	hostPort int
	hostIP   string
}

func (h HttpHost) Scheme() string {
	return "http"
}

func (h HttpHost) Port() int {
	return h.hostPort
}

func (h HttpHost) Host() string {
	return h.hostIP
}

func NewHttpHost(host string, port int) HttpHost {
	return HttpHost{hostIP: host, hostPort: port}
}
