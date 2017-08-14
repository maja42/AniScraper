package taskplanner

import (
	"fmt"
	"sync"

	"github.com/maja42/AniScraper/filesystem"
	"github.com/maja42/AniScraper/utils"
	"golang.org/x/net/context"
)

type TaskPlanner interface {
	// Start initializes the task planner
	Start(ctx context.Context) error

	// Wait blocks until all go routines (if any) have finished
	Wait()
}

type taskPlanner struct {
	animeLibrary   filesystem.AnimeLibrary
	incomingEvents <-chan filesystem.Event

	isRunning bool

	mutex  sync.RWMutex
	wg     sync.WaitGroup
	logger utils.Logger
}

func NewTaskPlanner(animeLibrary filesystem.AnimeLibrary, logger utils.Logger) TaskPlanner {
	taskPlanner := &taskPlanner{
		animeLibrary: animeLibrary,
		// libraryEvents: libraryEvents,
		logger: logger.New("TaskPlanner"),
	}

	return taskPlanner
}

func (t *taskPlanner) Start(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.isRunning {
		return fmt.Errorf("Already running")
	}
	t.isRunning = true

	t.incomingEvents = t.animeLibrary.Subscribe(ctx, true)

	t.wg.Add(1)
	go t.processEvents()
	return nil
}

func (t *taskPlanner) Wait() {
	t.wg.Wait()
}

func (t *taskPlanner) processEvents() {
	defer t.wg.Done()
	for event := range t.incomingEvents {
		t.processEvent(event)
	}
}

func (t *taskPlanner) processEvent(event filesystem.Event) {
	t.logger.Infof("processing event %q", event.String())
}
