package file

import (
	"path"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/go-msvc/errors"
	"github.com/go-msvc/logger"
)

var (
	watcherMutex sync.Mutex
	watcher      *fsnotify.Watcher
	watchedFiles = map[string]*watchedFile{}
	log          = logger.NewLogger("file-items")
)

type watchedFile struct {
	channels [][]chan bool
}

//bit values to identify file operations
type OperationBits uint32

const (
	Create OperationBits = 1
	Write  OperationBits = 2
	Remove OperationBits = 4
	Rename OperationBits = 8
	Chmod  OperationBits = 16
)

var (
	//warning: both types of operations here must be in same sequence
	allOperationBits   = []OperationBits{Create, Write, Remove, Rename, Chmod}
	allOperationValues = []fsnotify.Op{fsnotify.Create, fsnotify.Write, fsnotify.Remove, fsnotify.Rename, fsnotify.Chmod}
)

func init() {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()
	watcher, _ = fsnotify.NewWatcher()
	go func(watcher *fsnotify.Watcher) {
		for {
			select {
			case event := <-watcher.Events:
				{
					log.Debugf("event: %v", event)
					if wf, ok := watchedFiles[event.Name]; ok {
						operIndex, _ := operFromValue(event.Op)
						operChannels := wf.channels[operIndex]
						//loop through all channels on this file
						for _, ch := range operChannels {
							//do non-blocking write into the channel
							select {
							case ch <- true:
								log.Debugf("notified %s", event.Name) //may be more than once per event if multiple interested parties
							default:
								log.Debugf("some %s events not read", event.Name) //channel is blocking
							}
						}
					}
				}
			case err := <-watcher.Errors:
				{
					log.Errorf("error: %v", err)
				}
			} //select
		} //for
	}(watcher)
}

//Watcher for a specified file
//WARNING: if file does not exist, the parent will be watched, which may be inefficient if parent has many files
func Watcher(filename string, operationSet OperationBits) chan bool {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()

	wf, exists := watchedFiles[filename]
	if !exists {
		//not yet watching this file
		//try to watch it or its parent if it does not exist
		watchPath := filename
		for watchPath != "" {
			err := watcher.Add(watchPath)
			if err == nil {
				log.Debugf("watching %s", watchPath)
				break
			}
			log.Debugf("cannot watch %s: %+v", watchPath, err)
			if watchPath == "/" {
				panic(errors.Errorf("cannot watch %s", filename))
			}
			watchPath = path.Dir(watchPath)
		}
		wf = &watchedFile{
			channels: make([][]chan bool, len(allOperationBits)),
		}
		for operIndex, _ := range allOperationBits {
			wf.channels[operIndex] = make([]chan bool, 0)
		}
		watchedFiles[filename] = wf
	}

	//add channel for this user on selected operations
	ch := make(chan bool)
	for operIndex, oper := range allOperationBits {
		if operationSet&oper == oper {
			wf.channels[operIndex] = append(wf.channels[operIndex], ch)
		}
	}
	return ch
}

func operFromValue(op fsnotify.Op) (int, OperationBits) {
	for operIndex, operValue := range allOperationValues {
		if op == operValue {
			return operIndex, allOperationBits[operIndex]
		}
	}
	panic(errors.Errorf("unknown oper %v", op))
}
