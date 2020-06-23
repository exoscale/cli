package sos_sync

// region Logger
type TaskLogger interface {
	Log(task Task)
}

// endregion

// region Memory
type MemoryTaskLogger struct {
	Tasks []Task
}

func NewMemoryTaskLogger() *MemoryTaskLogger {
	return &MemoryTaskLogger{
		Tasks: []Task{},
	}
}

func (memoryTaskLogger *MemoryTaskLogger) Log(task Task) {
	memoryTaskLogger.Tasks = append(memoryTaskLogger.Tasks, task)
}

// endregion
