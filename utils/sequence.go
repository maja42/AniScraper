package utils

import "sync"

type Sequence interface {
	Next() int
}

type sequence struct {
	mutex     sync.Mutex
	nextValue int
}

func NewSequenceGenerator(firstValue int) Sequence {
	return &sequence{
		nextValue: firstValue,
	}
}

func (seq *sequence) Next() int {
	seq.mutex.Lock()
	defer seq.mutex.Unlock()
	value := seq.nextValue
	seq.nextValue++
	return value
}
