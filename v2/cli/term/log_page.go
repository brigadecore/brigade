package term

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/armon/circbuf"
	"github.com/brigadecore/brigade/sdk/v3/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const logPageName = "log"

type logPage struct {
	*page
	logText  *tview.TextView
	logBuf   *circbuf.Buffer
	maxBytes int64
}

func newLogPage(
	apiClient core.APIClient,
	app *tview.Application,
	router *pageRouter,
) *logPage {
	l := &logPage{
		page:    newPage(apiClient, app, router),
		logText: tview.NewTextView().SetDynamicColors(true),
	}

	l.maxBytes = 65535
	l.logText.SetBorder(true).SetTitle("Logs (<-/Del) Quit")

	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	l.page.Flex = tview.NewFlex().
		AddItem(
			nil, // Spacer to help create the illusion of a floating window
			0,
			1,     // Proportionate height-- 1 unit
			false, // Don't bring into focus
		).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(
					nil, // Spacer to help create the illusion of a floating window
					0,
					1,     // Proportionate width-- 1 unit
					false, // Don't bring into focus
				).
				AddItem(
					l.logText,
					25, //  Fixed width
					0,
					false,
				).
				AddItem(
					nil, // Spacer to help create the illusion of a floating window
					0,
					1,     // Proportionate width-- 1 unit
					false, // Don't bring into focus
				),
			85, // Fixed height
			0,
			false, // Don't bring into focus
		).
		AddItem(
			nil, // Spacer to help create the illusion of a floating window
			0,
			1,     // Proportionate height-- 1 unit
			false, // Don't bring into focus
		)

	return l
}

func (l *logPage) load(ctx context.Context, eventID string, jobID string) {

	l.logText.Clear()
	l.app.SetFocus(l.logText)
	l.logText.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		switch evt.Key() {
		case // Back
			tcell.KeyLeft,
			tcell.KeyDelete,
			tcell.KeyBackspace,
			tcell.KeyBackspace2:
			l.app.Sync()
			if jobID == "" {
				l.router.loadEventPage(eventID)
			} else {
				l.router.loadJobPage(eventID, jobID)
			}
		}
		return evt
	})

	var err error
	l.logBuf, err = circbuf.NewBuffer(l.maxBytes)
	if err != nil {
		l.logText.SetText(err.Error())
		return
	}

	go l.streamLogsToBuffer(ctx, eventID, jobID)
	go l.writeLogs(ctx)
}

// refresh refreshes Event info and associated Jobs and repaints the page.
func (l *logPage) refresh(ctx context.Context, eventID string, jobID string) {
}

// nolint: lll
func (l *logPage) streamLogsToBuffer(ctx context.Context, eventID string, jobID string) {
	l.logBuf.Reset()
	var logsSelector core.LogsSelector
	if jobID == "" {
		logsSelector = core.LogsSelector{}
	} else {
		logsSelector = core.LogsSelector{Job: jobID}
	}
	logEntryCh, errCh, err := l.apiClient.Events().Logs().Stream(
		ctx,
		eventID,
		&logsSelector,
		&core.LogStreamOptions{Follow: true},
	)
	if err != nil {
		l.logText.SetText(err.Error())
	}

	for {
		select {
		case logEntry, ok := <-logEntryCh:
			if ok {
				if _, err = l.logBuf.Write([]byte(logEntry.Message)); err != nil {
					l.logText.SetText(err.Error())
					break
				}
				if _, err = l.logBuf.Write([]byte("\n")); err != nil {
					l.logText.SetText(err.Error())
					break
				}
			} else {
				// logEntryCh was closed, but want to keep looping through this select
				// in case there are pending errors on the errCh still. nil channels are
				// never readable, so we'll just nil out logEntryCh and move on.
				logEntryCh = nil
			}
		case err, ok := <-errCh:
			if ok {
				// TODO: Remove and handle this
				log.Println(err)
			}
			// errCh was closed, but want to keep looping through this select in case
			// there are pending messages on the logEntryCh still. nil channels are
			// never readable, so we'll just nil out errCh and move on.
			errCh = nil
		case <-ctx.Done():
			return
		}
		// If BOTH logEntryCh and errCh were closed, we're done.
		if logEntryCh == nil && errCh == nil {
			break
		}
	}
}

func (l *logPage) writeLogs(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if l.logBuf.TotalWritten() > l.maxBytes {
				l.logText.SetText(
					fmt.Sprintf("(Previous text omitted)\n %s", l.logBuf.String()),
				)
			} else {
				l.logText.SetText(l.logBuf.String())
			}
			l.app.Draw()
		case <-ctx.Done():
			return
		}
	}
}
