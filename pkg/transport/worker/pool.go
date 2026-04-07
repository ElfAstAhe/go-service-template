package worker

type Pool[D any] interface {
	CommonWorker

	Push(data D)
	TryPush(data D) bool
	Len() int
	Capacity() int
}
