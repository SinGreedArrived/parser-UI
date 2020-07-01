package main

import (
	"log"

	"github.com/jroimartin/gocui"
)

type Binding struct {
	ViewName    []string
	Handler     func(*gocui.Gui, *gocui.View) error
	Key         interface{}
	Modifier    gocui.Modifier
	Description string
}

func KeyBinding(gui *gocui.Gui) {
	bindings := []*Binding{
		/*   Шаблон для биндинга клавиш
		{
			ViewName:	<<blank>>,
			Handler:	<<blank>>,
			Key:			<<blank>>,
			Modifier:	<<blank>>,
			Description:	<<blank>>,
		}
		*/
		{
			ViewName:    []string{""},
			Handler:     quit,
			Key:         gocui.KeyCtrlC,
			Modifier:    gocui.ModNone,
			Description: "Exit",
		},
		{
			ViewName:    []string{""},
			Handler:     nextView,
			Key:         gocui.KeyTab,
			Modifier:    gocui.ModNone,
			Description: "Change view",
		},
		{
			ViewName:    []string{""},
			Handler:     SaveRegexp,
			Key:         gocui.KeyCtrlS,
			Modifier:    gocui.ModNone,
			Description: "Save target and regexp",
		},
		{
			ViewName:    []string{"URL", "Regular", "Mask"},
			Handler:     clearLayout,
			Key:         'd',
			Modifier:    gocui.ModAlt,
			Description: "clear layout",
		},
		{
			ViewName:    []string{"Regular", "Mask"},
			Handler:     useRegexp,
			Key:         gocui.KeyEnter,
			Modifier:    gocui.ModNone,
			Description: "Use regexp to page",
		},
		{
			ViewName:    []string{"URL"},
			Handler:     DownloadPage,
			Key:         gocui.KeyEnter,
			Modifier:    gocui.ModNone,
			Description: "Download page",
		},
		{
			ViewName:    []string{"MainView"},
			Handler:     cursorDown,
			Key:         'j',
			Modifier:    gocui.ModNone,
			Description: "Cursor Down",
		},
		{
			ViewName:    []string{"MainView"},
			Handler:     cursorUp,
			Key:         'k',
			Modifier:    gocui.ModNone,
			Description: "Cursor Up",
		},
	}
	for _, bind := range bindings {
		for _, name := range bind.ViewName {
			if err := gui.SetKeybinding(name, bind.Key, bind.Modifier, bind.Handler); err != nil {
				log.Panicln(err, name)
			}
		}
	}
}
