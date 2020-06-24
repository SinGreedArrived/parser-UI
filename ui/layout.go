package main

import (
	"github.com/jroimartin/gocui"
)

type layout struct {
	ViewName string
	x0       int
	y0       int
	x1       int
	y1       int
	Title    string
	Editable bool
	Wrap     bool
}

func Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	maxX--
	layouts := []*layout{
		{
			ViewName: "URL",
			x0:       0,
			y0:       0,
			x1:       maxX,
			y1:       2,
			Title:    "URL - source for parsing",
			Editable: true,
			Wrap:     false,
		},
		{
			ViewName: "Rname",
			x0:       0,
			y0:       3,
			x1:       maxX,
			y1:       5,
			Title:    "Regexp - Name",
			Editable: true,
			Wrap:     true,
		},
		{
			ViewName: "Rtarget",
			x0:       0,
			y0:       6,
			x1:       maxX,
			y1:       8,
			Title:    "Regexp - Target",
			Editable: true,
			Wrap:     true,
		},
		{
			ViewName: "MainView",
			x0:       0,
			y0:       9,
			x1:       maxX,
			y1:       maxY - 1,
			Title:    "Page Source",
			Editable: false,
			Wrap:     true,
		},
	}
	for _, view := range layouts {
		if v, err := g.SetView(view.ViewName, view.x0, view.y0, view.x1, view.y1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = view.Title
			v.Editable = view.Editable
			v.Wrap = view.Wrap
		}
	}
	return nil
}
