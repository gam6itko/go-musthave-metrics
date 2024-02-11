package sync

type NullSemaphore struct {
}

func (ths NullSemaphore) Acquire() {
	// do nothing
}

func (ths NullSemaphore) Release() {
	// do nothing
}
