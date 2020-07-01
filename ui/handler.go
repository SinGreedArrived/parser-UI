package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/jroimartin/gocui"
)

var buff *target

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

func writeLayout(nameView, data string, g *gocui.Gui) error {
	g.Update(func(g *gocui.Gui) error {
		view, err := g.View(nameView)
		if err != nil {
			StatusWrite(g, fmt.Sprintf("Can't GetView:: %s", err))
			return nil
		}
		clearLayout(g, view)
		fmt.Fprint(view, data)
		return nil
	})
	return nil
}

func readLayout(nameView string, g *gocui.Gui) string {
	view, err := g.View(nameView)
	if err != nil {
		StatusWrite(g, fmt.Sprintf("Can't GetView:: %s", err))
		return ""
	}
	Btext, _ := ioutil.ReadAll(view)
	view.Rewind()
	text := string(Btext)
	return strings.TrimSpace(text)
}

func StatusWrite(g *gocui.Gui, message string) error {
	writeLayout("StatusBar", message, g)
	return nil
}

func DownloadPage(g *gocui.Gui, v *gocui.View) error {
	StatusWrite(g, "")
	url := readLayout(v.Name(), g)
	trg := database.GetTarget(url)
	if trg == nil {
		trg = new(target)
		trg.Url = url
		trg.Cur = "0"
	}
	body, err := trg.GetData()
	if err != nil {
		StatusWrite(g, err.Error())
		return nil
	}
	writeLayout("MainView", string(body), g)
	rgx := database.GetRegexp(url)
	if rgx != nil {
		writeLayout("Regular", rgx.Exp, g)
		writeLayout("Mask", rgx.Mask, g)
	} else {
		writeLayout("Regular", "", g)
		writeLayout("Mask", "", g)
	}
	buff = trg
	return nil
}

func SaveRegexp(g *gocui.Gui, v *gocui.View) error {
	database.CreateTarget(buff.Url, "0")
	rgx := readLayout("Regular", g)
	msk := readLayout("Mask", g)
	database.CreateRegexp(buff.Url, rgx, msk)
	StatusWrite(g, "Saved data...")
	return nil
}

func useRegexp(g *gocui.Gui, v *gocui.View) error {
	StatusWrite(g, "Use Regexp on Page: "+buff.Url)
	body, err := buff.GetData()
	if err != nil {
		StatusWrite(g, fmt.Sprint(err))
		return nil
	}

	rgx := readLayout("Regular", g)
	msk := readLayout("Mask", g)

	regex := regexp.MustCompile(rgx)
	res := regex.ReplaceAllString(string(body), msk)
	writeLayout("MainView", res, g)
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

// Exit
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
