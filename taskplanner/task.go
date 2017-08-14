package taskplanner

import (
	"github.com/maja42/AniScraper/filesystem"
	"golang.org/x/net/context"
)

type TaskType int

const (
	PROCESS_ANIME_FOLDER TaskType = iota
)

type Task struct {
	TaskType    TaskType
	AnimeFolder *filesystem.AnimeFolder

	Ctx        context.Context
	cancelFunc context.CancelFunc
}

func (t *Task) String() string {
	var str string
	switch t.TaskType {
	case PROCESS_ANIME_FOLDER:
		str = "Process anime folder"
	}
	return str + ": " + t.AnimeFolder.String()
}
