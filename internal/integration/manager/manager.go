package manager

type ManagerApi interface {
	RegisterNode(scheme string, host string, port int) error
}
