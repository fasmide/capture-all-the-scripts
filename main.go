package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/fasmide/capture-all-the-scripts/server"
	ui "github.com/gizak/termui"
)

var (
	port = flag.Int("port", 22, "specify listen port")
)

func main() {

	flag.Parse()

	listenPath := fmt.Sprintf("0.0.0.0:%d", *port)
	eventChan := make(chan string)

	server := server.SSH{Path: listenPath, Events: eventChan}
	go server.Listen()

	log.Printf("listening on %s", listenPath)

	gui(&server, eventChan)
}

func gui(server *server.SSH, events chan string) {
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	strs := []string{
		"[0] github.com/gizak/termui",
		"[1] [你好，世界](fg-blue)",
		"[2] [こんにちは世界](fg-red)",
		"[3] [color output](fg-white,bg-green)",
		"[4] output.go",
		"[5] random_out.go",
		"[6] dashboard.go",
		"[7] nsf/termbox-go"}

	ls := ui.NewList()
	ls.Items = strs
	ls.ItemFgColor = ui.ColorYellow
	ls.BorderLabel = "(%d) Active Connections"
	ls.Height = 7
	ls.Y = 0

	par := ui.NewPar("Total conns: \t%d\nTotal bytes: \t%d")
	par.Height = 4
	par.BorderLabel = "Stats"

	log := ui.NewList()
	log.ItemFgColor = ui.ColorYellow
	log.BorderLabel = "Log"
	log.Height = 7

	var renderLock sync.Mutex

	// build layout
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(6, 0, ls),
			ui.NewCol(6, 0, par),
		),
		ui.NewRow(
			ui.NewCol(12, 0, log),
		),
	)

	// calculate layout
	ui.Body.Align()

	ui.Render(ui.Body)

	ui.Handle("q", func(ui.Event) {
		ui.StopLoop()
	})

	go func() {
		for {
			renderLock.Lock()
			s := server.State()
			sort.Sort(SortByStarted(s.Connections))
			list := make([]string, len(s.Connections))
			activeBytes := 0
			for i, item := range s.Connections {
				list[i] = fmt.Sprintf("%11s: %7s: %s", time.Now().Sub(item.Started).Truncate(time.Second), humanize.Bytes(uint64(item.BytesSent)), item.Remote)
				activeBytes += item.BytesSent
			}
			ls.Items = list
			ls.BorderLabel = fmt.Sprintf("(%d) Active connections", len(list))

			par.Text = fmt.Sprintf("Total conns: \t%d\nTotal bytes: \t%s", s.TotalConnections, humanize.Bytes(uint64(s.BytesSent+activeBytes)))
			ui.Render(ui.Body)
			renderLock.Unlock()

			time.Sleep(time.Millisecond * 500)
		}
	}()

	go func() {
		for {
			event := <-events
			renderLock.Lock()
			log.Items = append([]string{event}, log.Items...)
			// cap the log to 100 entries
			if len(log.Items) >= 100 {
				log.Items = log.Items[:100]
			}
			renderLock.Unlock()
		}
	}()

	ui.Handle("<Resize>", func(e ui.Event) {
		renderLock.Lock()
		defer renderLock.Unlock()
		payload := e.Payload.(ui.Resize)
		ui.Body.Width = payload.Width
		height := payload.Height / 2
		ls.Height = height
		log.Height = height
		ui.Body.Align()
		ui.Clear()
		ui.Render(ui.Body)
	})
	ui.Loop()
}

type SortByStarted []*server.Connection

func (s SortByStarted) Len() int           { return len(s) }
func (s SortByStarted) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortByStarted) Less(i, j int) bool { return s[i].Started.Before(s[j].Started) }
