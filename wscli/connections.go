package main

import (
	"github.com/jroimartin/gocui"
)

type server struct {
	clients int
}

var cnxState = struct {
	cli            string
	commandHistory []string
	errMsg         string
	servers        []server
}{
	cli:            "",
	commandHistory: make([]string, 0),
	errMsg:         "",
	servers:        make([]server, 0),
}

func cnxLayout(g *gocui.Gui) error {

	setUniversalKeys(g)
	maxX, maxY := g.Size()
	servers, err := g.SetView("Servers", 0, 0, maxX/2, maxY-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	clients, err := g.SetView("Clients", maxX/2, 0, maxX, maxY-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	cli, err := g.SetView("CLI", 0, maxY-3, maxX, maxY)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	servers.Title = "Your servers"
	clients.Title = "Your clients"
	cli.Title = "CLI"
	if err := g.SetKeybinding("CLI", gocui.KeyEnter, gocui.ModNone, cnxCaptureCommand); err != nil {
		return err
	}
	_, err = g.SetCurrentView("CLI")
	if err != nil {
		return err
	}

	if cnxState.errMsg != "" {
		h := func(g *gocui.Gui, v *gocui.View) error {
			cnxState.errMsg = ""
			return nil
		}
		clients.Write([]byte(cnxState.errMsg))
		g.DeleteKeybinding("CLI", gocui.KeyEnter, gocui.ModNone)
		drawError(g, cnxState.errMsg, h)
	} else {
		g.DeleteView("Error")
		cli.Editable = true
		if x, y := cli.Cursor(); x == 0 && y == 0 {
			cli.Write([]byte(">"))
			cli.SetCursor(1, 0)
		}
		cnxState.cli = cli.Buffer()
		servers.Write([]byte(cnxState.cli))
	}
	return nil

}

func cnxCaptureCommand(g *gocui.Gui, v *gocui.View) error {
	if v.Buffer() == "\n" || v.Buffer() == "" {
		return nil
	}
	command := v.Buffer()
	v.Clear()
	v.SetCursor(0, 0)
	cnxProcessCommand(command)
	return nil
}

func cnxProcessCommand(command string) {
	l := len(command)
	if l == 0 {
		return
	}
	command = command[1:l]
	l = l - 1
	if l == 0 || command[0] == '\n' {
		return
	}
	command = command[0 : l-1]
	l = l - 1
	cnxSwitchCommand(command)
}
func cnxSwitchCommand(command string) error {
	cnxState.commandHistory = append(cnxState.commandHistory, command)
	if command == "SERVE" {

		errChan := make(chan error)
		clientChan := make(chan bool)
		activeChan := make(chan bool)
		newServer := &server{}
		cnxState.servers = append(cnxState.servers, server)
		go func(e chan error, c chan bool, a chan bool, ns *server) {
			for {
				select {
				case i := <-e:
					cnxState.ErrMsg = i.Error()
				case i := <-c:
					*ns.clients = *ns.clients + 1
				case i := <-a:
					break
				}
				a.Close()
				b.Close()
				c.Close()
			}
		}(errChan, clientChan, activeChan, newServer)
		go func(e chan error, c chan bool, a chan bool) {
			serve(DefaultHostPort, e, c)
			a <- true
		}(errChan, clientChan, activeChan)
	}
	return nil
}
