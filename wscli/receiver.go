package main

import (
	"github.com/jroimartin/gocui"
)

func receiverLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	binary, err := g.SetView("Binary", 0, 0, maxX/3, maxY-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	text, err := g.SetView("Text", maxX/3, 0, 2*maxX/3, maxY-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	ping, err := g.SetView("Ping", 2*maxX/3, 0, maxX, maxY/2-3)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	pong, err := g.SetView("Pong", 2*maxX/3, maxY/2-5, maxX, maxY-3)
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
	_ = binary
	_ = text
	_ = ping
	_ = pong
	_ = cli
	return nil
}
