package main

import (
	"fmt"
	"io/ioutil"

	"github.com/jroimartin/gocui"
)

// Change select view
func nextView(g *gocui.Gui, v *gocui.View) error {
	if g.CurrentView() == nil {
		if _, err := g.SetCurrentView(layouts[0].Title); err != nil {
			return err
		}
		return nil
	}
	oldView := g.CurrentView().Name()
	//views := g.Views()
	views := layouts
	for indView, view := range views {
		if g.CurrentView().Name() == view.Name() {
			if len(views)-1 == indView {
				if _, err := g.SetCurrentView(views[0].Name()); err != nil {
					return err
				}
				logger.Printf("Change select view %s to %s", oldView, g.CurrentView().Name())
				return nil
			} else {
				if _, err := g.SetCurrentView(views[indView+1].Name()); err != nil {
					return err
				}
				if !views[indView+1].Selected {
					nextView(g, v)
				}
				logger.Printf("Change select view %s to %s", oldView, g.CurrentView().Name())
				return nil
			}
		}
	}
	return nil
}

// Clear text in view
func clearLayout(g *gocui.Gui, v *gocui.View) error {
	v.Clear()
	v.SetCursor(0, 0)
	logger.Printf("Clear all text in view: %s", v.Name())
	return nil
}

func StatusWrite(g *gocui.Gui, message string) error {
	st, _ := g.View("StatusBar")
	clearLayout(g, st)
	fmt.Fprintf(st, message)
	return nil
}

func clearStatusBar(g *gocui.Gui) error {
	st, _ := g.View("StatusBar")
	clearLayout(g, st)
	return nil
}

func DownloadPage(g *gocui.Gui, v *gocui.View) error {
	clearStatusBar(g)
	Burl, _ := ioutil.ReadAll(v)
	v.Rewind()
	url := string(Burl)
	target := database.GetTarget(url)
	if target == nil {
		target = database.CreateTarget(url, "0")
	}
	body, err := target.GetBody()
	if err != nil {
		StatusWrite(g, err.Error())
		return nil
	}
	view, err := g.View("MainView")
	if err != nil {
		StatusWrite(g, fmt.Sprintf("func:handler.DownloadPage() GetView:: %s", err))
	}
	clearLayout(g, view)
	fmt.Fprintf(view, "%s", string(body))
	view, err = g.View("Regular")
	if err != nil {
		StatusWrite(g, fmt.Sprintf("func:handler.DownloadPage() GetView:: %s", err))
	}
	rgx := database.GetRegexp(url)
	if rgx != nil {
		clearLayout(g, view)
		fmt.Fprintf(view, "%s", rgx.Exp)
	}
	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func useRegexp(g *gocui.Gui, v *gocui.View) error {

	return nil
}

// Exit
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
