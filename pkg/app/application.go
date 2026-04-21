package app

type Application interface {
	Init() error
	Run() error
	Stop() error
	Close() error
}
