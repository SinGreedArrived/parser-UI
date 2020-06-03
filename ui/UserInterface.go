package main

import (
	"log"

	"github.com/jroimartin/gocui"
)

func nextView(g *gocui.Gui, v *gocui.View) error {
	views := g.Views()
	for indView, view := range views {
		if g.CurrentView().Name() == view.Name() {
			if len(views)-1 == indView {
				if _, err := g.SetCurrentView(views[0].Name()); err != nil {
					return err
				}
				break
			} else {
				if _, err := g.SetCurrentView(views[indView+1].Name()); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func layout(g *gocui.Gui) error {
	maxX, _ := g.Size()
	if v, err := g.SetView("URL", 0, 0, maxX, 4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "URL"
		v.Editable = true
		v.Wrap = true
		v.SetCursor(1, 1)
		g.SetCurrentView("URL")
	}
	if v, err := g.SetView("Rname", 0, 5, maxX/4, 9); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Regexp - Name"
		v.Editable = true
		v.Wrap = true
		v.SetCursor(1, 1)
	}
	if v, err := g.SetView("Rtarget", 0, 10, maxX/4, 14); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Regexp - Target"
		v.Editable = true
		v.Wrap = true
		v.SetCursor(1, 1)
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

	g.Highlight = true
	g.Cursor = true
	g.Mouse = true
	g.SelFgColor = gocui.ColorGreen

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
