package worker

type Scheduler interface {
	Start() error
	Stop() error
}
