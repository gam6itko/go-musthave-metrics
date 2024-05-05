package sync

// Semaphore структура семафора
type Semaphore struct {
	semaCh chan struct{}
}

// NewSemaphore создает семафор с буферизованным каналом емкостью maxReq
func NewSemaphore(maxReq uint64) *Semaphore {
	return &Semaphore{
		semaCh: make(chan struct{}, maxReq),
	}
}

// когда горутина запускается, отправляем пустую структуру в канал semaCh
func (s *Semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

// когда горутина завершается, из канала semaCh убирается пустая структура
func (s *Semaphore) Release() {
	<-s.semaCh
}
