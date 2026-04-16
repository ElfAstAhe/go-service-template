package container

// Container интерфейс контейнера
type Container interface {
	GetName() string

	Init() error
	Close() error

	Add(name string, instance any) error
	Remove(name string) error

	GetInstance(name string) (any, error)
	AllInstances() map[string]any
}
