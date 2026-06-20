package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Shadw-Developer/tamk/pkg/logger"
	"github.com/fsnotify/fsnotify"
)

type FileChangeHandler func(path string)

type ExtSet map[string]bool

var WatchExtensions = ExtSet{
	".html": true, ".css": true, ".js": true,
	".json": true, ".png": true, ".jpg": true,
	".jpeg": true, ".svg": true, ".webp": true,
	".xml": true, ".kt": true,
}

var IgnoreDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"secret":       true,
	".idea":        true,
}

type Watcher struct {
	watcher  *fsnotify.Watcher
	handler  FileChangeHandler
	debounce time.Duration
	events   chan string
	done     chan struct{}
	mu       sync.Mutex
	running  bool
	timer    *time.Timer
	pending  map[string]struct{}
	wg       sync.WaitGroup
}

func New(handler FileChangeHandler, debounce time.Duration) *Watcher {
	if debounce == 0 {
		debounce = 500 * time.Millisecond
	}
	return &Watcher{
		handler:  handler,
		debounce: debounce,
		events:   make(chan string, 256),
		done:     make(chan struct{}),
		pending:  make(map[string]struct{}),
	}
}

func (w *Watcher) Start(watchPath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.watcher = watcher

	if err := filepath.WalkDir(watchPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if IgnoreDirs[filepath.Base(path)] {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		watcher.Close()
		return err
	}

	w.mu.Lock()
	w.running = true
	w.mu.Unlock()

	w.wg.Add(2)
	go w.processEvents()
	go w.watchLoop()

	return nil
}

func (w *Watcher) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	w.mu.Unlock()

	close(w.done)
	if w.watcher != nil {
		w.watcher.Close()
	}
	w.wg.Wait()
}

func (w *Watcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

func (w *Watcher) watchLoop() {
	defer w.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Watcher watchLoop panicked", "recover", r)
		}
	}()
	if w.watcher == nil {
		return
	}
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				ext := strings.ToLower(filepath.Ext(event.Name))
				if WatchExtensions[ext] {
					w.queueEvent(event.Name)
				}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			logger.Warn("Watcher error", "error", err)
		case <-w.done:
			return
		}
	}
}

func (w *Watcher) queueEvent(name string) {
	select {
	case w.events <- name:
	default:
	}
}

func (w *Watcher) processEvents() {
	defer w.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Watcher panicked", "recover", r)
		}
	}()
	if w.watcher == nil {
		return
	}

	w.timer = time.NewTimer(w.debounce)
	w.timer.Stop()

	for {
		select {
		case path := <-w.events:
			w.mu.Lock()
			w.pending[path] = struct{}{}
			w.timer.Reset(w.debounce)
			w.mu.Unlock()

		case <-w.timer.C:
			w.mu.Lock()
			pending := w.pending
			w.pending = make(map[string]struct{})
			w.mu.Unlock()

			for path := range pending {
				w.handler(path)
			}

		case <-w.done:
			w.timer.Stop()
			return
		}
	}
}
