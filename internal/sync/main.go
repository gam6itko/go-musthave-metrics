package sync

type ISemaphore interface {
	Acquire()
	Release()
}
