package main

import (
	"flag"
	"log"
	"time"

	"github.com/jroimartin/gocui"
)

var DefaultHostPort = flag.String("h", "localhost:2222", "set a default host and port")

var debounce time.Duration = 50 * time.Millisecond
var ctlCDebouncer time.Time

func setUniversalKeys(g *gocui.Gui) {

	// set keybinding on tab to see all possible commands based on typed
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(cnxLayout)

	ch := make(chan bool)
	go func() {
		if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
			log.Panicln(err)
		}
		ch <- true
	}()
	<-ch
}
func drawError(g *gocui.Gui, msg string, h func(g *gocui.Gui, v *gocui.View) error) (*gocui.View, error) {

	maxX, maxY := g.Size()
	errorMsg, err := g.SetView("Error", maxX/2-(len(msg)/2)-3, maxY/2-2, maxX/2+(len(msg)+2)+2, maxY/2+2)
	errorMsg.Title = "Error"
	if err := g.SetKeybinding("Error", gocui.KeyEnter, gocui.ModNone, h); err != nil {
		return nil, err
	}
	if err := g.SetKeybinding("Error", gocui.KeyCtrlC, gocui.ModNone, h); err != nil {
		return nil, err
	}
	g.SetCurrentView("Error")
	errorMsg.Clear()
	_, err = errorMsg.Write([]byte("  \n   " + msg))
	return errorMsg, err
}

func quit(g *gocui.Gui, v *gocui.View) error {
	if time.Now().Before(ctlCDebouncer.Add(debounce)) {
		return nil
	}
	ctlCDebouncer = time.Now()
	if cli, _ := g.View("CLI"); cli == v {
		if v.Buffer() == ">\n" {
			return gocui.ErrQuit
		}
		v.Clear()
		v.SetCursor(0, 0)
		return nil
	}
	g.SetCurrentView("CLI")
	return nil
}
