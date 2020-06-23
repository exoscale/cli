package sos_sync

type Task struct {
	Action int
	File   string
	Size   int64
}

const (
	UploadAction = iota
	DeleteAction
)
