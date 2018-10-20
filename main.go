package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"sort"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/fasmide/capture-all-the-scripts/server"
	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

var (
	port = flag.Int("port", 22, "specify listen port")
)

var (
	activeConnView *gocui.View
	statsView      *gocui.View
	logView        *gocui.View
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

	started := time.Now()
	// TODO: make sure items in this list are
	// removed when clients disconnect
	lastWritten := make(map[string]int)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	// This goroutine refreshes state
	go func() {
		for {
			s := server.State()
			sort.Sort(byStarted(s.Connections))

			g.Update(func(g *gocui.Gui) error {

				// update active connections
				activeConnView.Clear()
				activeBytes := 0
				for _, item := range s.Connections {
					written := item.Written()
					perSecond := 0

					if last, exists := lastWritten[item.Remote]; exists {
						perSecond = written - last
					}
					lastWritten[item.Remote] = written

					color.New(color.FgGreen).Fprintf(activeConnView, "%11s: %7s: %7s/sec %s\n",
						time.Now().Sub(item.Started).Truncate(time.Second).String(),
						humanize.Bytes(uint64(written)),
						humanize.Bytes(uint64(perSecond*2)), // times two, as we update twice pr second
						item.Remote,
					)
					activeBytes += written

				}
				activeConnView.Title = fmt.Sprintf("(%d) Active connections", len(s.Connections))

				// update stats
				statsView.Clear()
				fmt.Fprintf(statsView, " Total conns: %d\n Total bytes: %s\n Uptime:      %s\n",
					s.TotalConnections,
					humanize.Bytes(uint64(s.BytesSent+activeBytes)),
					time.Now().Sub(started).Truncate(time.Second),
				)

				fmt.Fprintf(statsView, " Routines:    %d\n",
					runtime.NumGoroutine(),
				)

				v, _ := mem.VirtualMemory()

				fmt.Fprintf(statsView, " Used Memory: %.2f%%\n", v.UsedPercent)

				l, _ := load.Avg()
				fmt.Fprintf(statsView, " Load:        %.2f / %.2f / %.2f\n", l.Load1, l.Load5, l.Load15)

				return nil
			})

			time.Sleep(time.Millisecond * 500)
		}
	}()

	// This goroutine receives events
	go func() {
		eventSlice := make([]string, 0, 50)
		for {
			event := <-events
			eventSlice = append(eventSlice, event)
			if len(eventSlice) >= 30 {
				eventSlice = eventSlice[len(eventSlice)-29 : 30]
			}
			g.Update(func(g *gocui.Gui) error {
				logView.Clear()
				for _, s := range eventSlice {
					color.New(color.FgYellow).Fprintf(logView, "%s\n", s)
				}
				return nil
			})

		}
	}()

	// Gui Mainloop
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	var err error
	if activeConnView, err = g.SetView("activeconnections", 0, 0, maxX-35, maxY/2-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		activeConnView.Wrap = false
		activeConnView.Title = "(%d) Active Connections"
	}

	if statsView, err = g.SetView("stats", maxX-35, 0, maxX-1, maxY/2-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		statsView.Title = "Stats"
	}

	if logView, err = g.SetView("log", 0, maxY/2-1, maxX-1, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		logView.Title = "Log"
		logView.Autoscroll = true
	}

	return nil
}

type byStarted []*server.Connection

func (s byStarted) Len() int           { return len(s) }
func (s byStarted) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byStarted) Less(i, j int) bool { return s[i].Started.Before(s[j].Started) }
