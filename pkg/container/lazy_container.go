package container

type Provider func(name string) (any, error)

type LazyContainer interface {
	Container

	RegisterProvider(name string, provider Provider) error
	UnregisterProvider(name string) error

	Unregister(name string) error

	AllProviders() map[string]Provider
}
