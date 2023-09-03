package lyric

import (
	"sync"
	"time"
)

type Listener func(startTimeMs int64, content string, transContent string, last bool, index int)

type LRCTimer struct {
	file      *LRCFile
	transFile *TranslateLRCFile
	timer     chan time.Duration
	stop      chan struct{}
	listeners []Listener

	curIndex int
	l        sync.Mutex
}

func NewLRCTimer(file *LRCFile, transFile *TranslateLRCFile) *LRCTimer {
	return &LRCTimer{
		file:      file,
		transFile: transFile,
		timer:     make(chan time.Duration),
	}
}

func (t *LRCTimer) Timer() chan<- time.Duration {
	return t.timer
}

func (t *LRCTimer) AddListener(l Listener) {
	t.listeners = append(t.listeners, l)
}

func (t *LRCTimer) Start() {
	fragments := t.file.fragments

	if len(fragments) == 0 {
		return
	}

	t.Rewind()
	t.stop = make(chan struct{})
	var (
		current      = fragments[0]
		transContent = t.transFile.FindByTimeMs(current.StartTimeMs)
		isLast       = t.curIndex >= len(fragments)-1
	)
	for _, l := range t.listeners {
		l(current.StartTimeMs, current.Content, transContent, isLast, t.curIndex)
	}
	for {
		select {
		case <-t.stop:
			return
		case duration := <-t.timer:
			if isLast {
				break
			}

			if duration < time.Duration(fragments[t.curIndex].StartTimeMs)*time.Millisecond {
				for _, l := range t.listeners {
					go l(current.StartTimeMs, current.Content, transContent, isLast, t.curIndex)
				}
				continue
			}

			// locate after rewind
			for t.curIndex < len(fragments)-1 && duration >= time.Duration(fragments[t.curIndex+1].StartTimeMs)*time.Millisecond {
				t.l.Lock()
				t.curIndex++
				t.l.Unlock()
				current = fragments[t.curIndex]
				transContent = t.transFile.FindByTimeMs(current.StartTimeMs)
				isLast = t.curIndex > len(fragments)-1
			}

			for _, l := range t.listeners {
				go l(current.StartTimeMs, current.Content, transContent, isLast, t.curIndex)
			}
		}
	}
}

func (t *LRCTimer) IsStarted() bool {
	return t.timer != nil
}

func (t *LRCTimer) Stop() {
	t.l.Lock()
	defer t.l.Unlock()
	if t.stop != nil {
		close(t.stop)
		t.stop = nil
	}
	t.timer = nil
	t.listeners = nil
}

func (t *LRCTimer) Rewind() {
	t.l.Lock()
	defer t.l.Unlock()
	t.curIndex = 0
}

func (t *LRCTimer) GetLRCFragment(index int) (*LRCFragment, *LRCFragment) {
	if nil == t.file || index >= len(t.file.fragments) || index < 0 {
		return nil, nil
	}
	f := &t.file.fragments[index]
	transLyric := t.transFile.FindByTimeMs(f.StartTimeMs)

	return f, &LRCFragment{StartTimeMs: f.StartTimeMs, Content: transLyric}
}

func (t *LRCTimer) IsEmpty() bool {
	return nil == t.file || len(t.file.fragments) == 0
}
