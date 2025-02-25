package worker

type MockedWorkerPool struct {
}

func NewMockedWorkerPool() *MockedWorkerPool {
	return &MockedWorkerPool{}
}
func (wp *MockedWorkerPool) Run() {

}

func (wp *MockedWorkerPool) Worker() {

}

func (wp *MockedWorkerPool) StartTrackingGame(task *WorkerTask) {

}
