package ui

import (
	"github.com/rivo/tview"
)

type LogWindow struct {
	*tview.TextView
}

type Log interface {
	Log(s string)
}

func NewLogWindow() *LogWindow {
	log := &LogWindow{
		TextView: tview.NewTextView(),
	}

	return log
}

func (log *LogWindow) Log(s string) {
	log.TextView.Write([]byte(s))
}
