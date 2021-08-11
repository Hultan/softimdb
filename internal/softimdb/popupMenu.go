package softimdb

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hultan/softimdb/internal/data"
)

type PopupMenu struct {
	mainWindow *MainWindow
	popupMenu  *gtk.Menu

	popupTags               *gtk.MenuItem
	popupOpenFolder               *gtk.MenuItem
}

func NewPopupMenu(window *MainWindow) *PopupMenu {
	menu := new(PopupMenu)
	menu.mainWindow = window
	return menu
}

func (p *PopupMenu) Setup() {
	p.popupMenu = p.mainWindow.builder.GetObject("popupMenu").(*gtk.Menu)

	p.popupTags = p.mainWindow.builder.GetObject("popupTags").(*gtk.MenuItem)
	p.popupOpenFolder = p.mainWindow.builder.GetObject("popupOpenFolder").(*gtk.MenuItem)

	p.setupEvents()
}

func (p *PopupMenu) setupEvents() {
	_ = p.mainWindow.window.Connect("button-release-event", func(window *gtk.ApplicationWindow, event *gdk.Event) {
		buttonEvent := gdk.EventButtonNewFromEvent(event)
		if buttonEvent.Button() == gdk.BUTTON_SECONDARY {
			movie := p.mainWindow.getSelectedMovie()
			if movie == nil {
				return
			}

			menu, err := gtk.MenuNew()
			if err != nil {
				panic(err)
			} else {
				tags, err := p.mainWindow.database.GetTags()
				if err != nil {
					panic(err)
				}
				for i := 0; i < len(tags); i++ {
					tag := tags[i]
					item, err := gtk.CheckMenuItemNew()
					if err != nil {
						panic(err)
					}
					item.SetLabel(tag.Name)

					for i := 0; i < len(movie.Tags); i++ {
						if movie.Tags[i].Id == tag.Id {
							item.SetActive(true)
						}
					}

					menu.Add(item)
					item.Connect("activate", func() {
						if item.GetActive() {
							p.mainWindow.database.InsertMovieTag(movie, &tag)
							p.addTag(movie, &tag)
						} else {
							p.mainWindow.database.RemoveMovieTag(movie, &tag)
							p.removeTag(movie, &tag)
						}
					})
				}
				p.popupTags.SetSubmenu(menu)
				p.popupTags.ShowAll()
			}
			p.popupMenu.PopupAtPointer(event)
		}
	})

	p.popupOpenFolder.Connect("activate", func() {
		movie := p.mainWindow.getSelectedMovie()
		if movie== nil {
			return
		}
		p.mainWindow.openMovieDirectoryInNemo(movie)
	})
}

func (p *PopupMenu) addTag(movie *data.Movie, tag *data.Tag) {
	movie.Tags = append(movie.Tags, *tag)
}

func (p *PopupMenu) removeTag(movie *data.Movie, tag *data.Tag) {
	for i, t := range movie.Tags {
		if t.Id == tag.Id {
			movie.Tags = append(movie.Tags[:i], movie.Tags[i+1:]...)
		}
	}
}
