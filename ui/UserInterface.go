package main

import (
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

var (
	logger   *log.Logger
	PageBody string
)

func main() {
	file, err := os.OpenFile("logger.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		panic(err)

	}
	logger = log.New(file, "", log.Ltime|log.Ldate|log.Lshortfile)

	logger.Println("Start UI...")
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen

	logger.Println("Set layout manager function...")
	g.SetManagerFunc(Layout)
	KeyBinding(g)
	g.Update(func(g *gocui.Gui) error {
		if _, err := g.SetCurrentView("URL"); err != nil {
			return err
		}
		return nil
	})
	logger.Println("Run mainloop...")
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
