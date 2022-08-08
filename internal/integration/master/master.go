package master

type MasterApi interface {
	RegisterNode(scheme string, host string, port int) error
}
