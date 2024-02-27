package backend

import "gotty/pkg/webtty"

// Slave is webtty.Slave with some additional methods.
type Slave interface {
	webtty.Slave

	Close() error
}

// Factory Factory
type Factory interface {
	Name() string
	New(params map[string][]string) (Slave, error)
}
