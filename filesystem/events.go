package filesystem

type EventType int

const (
	FOLDER_ADDED EventType = iota
	FOLDER_REMOVED
	FOLDER_CONTENT_MODIFIED
)

type Event struct {
	EventType   EventType
	AnimeFolder *AnimeFolder
}

func (e *Event) String() string {
	var str string
	switch e.EventType {
	case FOLDER_ADDED:
		str = "Anime folder added"
	case FOLDER_REMOVED:
		str = "Anime folder removed"
	case FOLDER_CONTENT_MODIFIED:
		str = "Anime folder content modified"
	}
	return str + ": " + e.AnimeFolder.String()
}
