package ignorelist

import (
	"sync"

	"github.com/egfanboy/mediapire-media-host/internal/utils"
)

/**
* Represents a list of files to ignore from fs events during runtime.
* Expectation is during operations where we are changing files we don't want the fs service to
* be reacting to these since the service will handle them. This ensures that most events in
* the fs service are from outside the scope of mediapire. IE: user manually adding/deleting files from disk
 */
type IgnoreList interface {
	AddFile(string)
	RemoveFile(string)
	IsFileIgnored(string) bool
}

type ignoreList struct{}

var (
	list = utils.NewConcurrentMap[string, bool]()
	once sync.Once
	il   *ignoreList
)

func (l ignoreList) AddFile(f string) {
	list.Add(f, true)
}

func (l ignoreList) RemoveFile(f string) {
	if _, ok := list.GetKey(f); ok {
		list.Delete(f)
	}
}

func (l ignoreList) IsFileIgnored(f string) bool {
	_, ok := list.GetKey(f)

	return ok
}

func GetIgnoreList() IgnoreList {
	once.Do(func() {
		il = &ignoreList{}
	})

	return il
}
