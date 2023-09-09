package configurable

type Configurabler interface {
	Configure(opts ...ConfigurablerFn) error
	GetInt(key string) (int, error)
}

// ConfigurablerFn is function that adheres to the Configurabler interface.
type ConfigurablerFn func(t interface{}) error
