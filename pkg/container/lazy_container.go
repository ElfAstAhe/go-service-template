package container

type Provider func() (any, error)

type LazyContainer interface {
	Container

	// RegisterProvider register non runnable provider
	RegisterProvider(name string, provider Provider) error
	// RegisterRunnableProvider register runnable provider
	RegisterRunnableProvider(name string, provider Provider) error
	// UnregisterProvider unregister any registered provider
	UnregisterProvider(name string) error
	// Unregister remove provider and instance from lists, errors ignored
	Unregister(name string) error
	// AllProviders return all registered providers
	AllProviders() map[string]Provider
}
