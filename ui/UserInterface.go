package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

type ButtonWidget struct {
	name    string
	x, y    int
	w       int
	label   string
	handler func(g *gocui.Gui, v *gocui.View) error
}

type WinWidget struct {
	name   string
	x1, y1 int
	x2, y2 int
	//handler func(g *gocui.Gui, v *gocui.View) error
}

func NewWinWidget(name string, x1, y1, x2, y2 int) *WinWidget {
	return &WinWidget{name: name, x1: x1, y1: y1, x2: x2, y2: y2}
}

func (w *WinWidget) Layout(g *gocui.Gui) error {
	_, err := g.SetView(w.name, w.x1, w.y1, w.x2, w.y2)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	return nil
}

func NewButtonWidget(name string, x, y int, label string, handler func(g *gocui.Gui, v *gocui.View) error) *ButtonWidget {
	return &ButtonWidget{name: name, x: x, y: y, w: len(label) + 1, label: label, handler: handler}
}

func (w *ButtonWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.x+w.w, w.y+2)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if _, err := g.SetCurrentView(w.name); err != nil {
			return err
		}
		if err := g.SetKeybinding(w.name, gocui.KeyEnter, gocui.ModNone, w.handler); err != nil {
			return err
		}
		fmt.Fprint(v, w.label)
	}
	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if _, err := g.SetView("side", -1, -1, int(0.2*float32(maxX)), maxY-5); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	}
	if _, err := g.SetView("main", int(0.2*float32(maxX)), -1, maxX, maxY-5); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	}
	if _, err := g.SetView("cmdline", -1, maxY-5, maxX, maxY); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	}
	return nil
}

func keybinding(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()
	maxX, maxY := g.Size()

	butup := NewButtonWidget("butup", 58, 7, "UP", nil)
	side := NewWinWidget("side", -1, -1, int(0.2*float32(maxX)), maxY-5)
	g.SetManagerFunc(layout)
	g.SetManager(butup, side)
	if err = keybinding(g); err != nil {
		panic(err.Error())
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
