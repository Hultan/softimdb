package softimdb

import (
	"log"
	"path"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/data"
)

type popupMenu struct {
	mainWindow *MainWindow
	popupMenu  *gtk.Menu

	popupTags          *gtk.MenuItem
	popupOpenFolder    *gtk.MenuItem
	popupOpenIMDB      *gtk.MenuItem
	popupOpenMovieInfo *gtk.MenuItem
	popupOpenPack      *gtk.MenuItem
	popupPlayMovie     *gtk.MenuItem
}

func newPopupMenu(window *MainWindow) *popupMenu {
	menu := new(popupMenu)
	menu.mainWindow = window
	return menu
}

func (p *popupMenu) setup() {
	p.popupMenu = p.mainWindow.builder.GetObject("popupMenu").(*gtk.Menu)

	p.popupTags = p.mainWindow.builder.GetObject("popupTags").(*gtk.MenuItem)
	p.popupOpenFolder = p.mainWindow.builder.GetObject("popupOpenFolder").(*gtk.MenuItem)
	p.popupOpenIMDB = p.mainWindow.builder.GetObject("popupOpenIMDBPage").(*gtk.MenuItem)
	p.popupOpenMovieInfo = p.mainWindow.builder.GetObject("popupOpenMovieInfo").(*gtk.MenuItem)
	p.popupOpenPack = p.mainWindow.builder.GetObject("popupOpenPack").(*gtk.MenuItem)
	p.popupPlayMovie = p.mainWindow.builder.GetObject("popupPlayMovie").(*gtk.MenuItem)

	p.setupEvents()
}

func (p *popupMenu) setupEvents() {
	_ = p.mainWindow.window.Connect(
		"button-release-event", func(window *gtk.ApplicationWindow, event *gdk.Event) {
			buttonEvent := gdk.EventButtonNewFromEvent(event)
			if buttonEvent.Button() != gdk.BUTTON_SECONDARY {
				return
			}

			movie := p.mainWindow.getSelectedMovie()
			if movie == nil {
				return
			}

			p.popupOpenPack.SetSensitive(false)
			if movie.Pack != "" {
				p.popupOpenPack.SetSensitive(true)
			}

			menu, err := gtk.MenuNew()
			if err != nil {
				reportError(err)
				log.Fatal(err)
			}

			tags, err := p.mainWindow.database.GetTags()
			if err != nil {
				reportError(err)
				return
			}

			p.createTagMenu(tags, movie, menu)
			p.popupTags.SetSubmenu(menu)
			p.popupTags.ShowAll()
			p.popupMenu.PopupAtPointer(event)
		},
	)

	p.popupOpenFolder.Connect(
		"activate", func() {
			movie := p.mainWindow.getSelectedMovie()
			if movie == nil {
				return
			}
			openInNemo(path.Join(p.mainWindow.config.RootDir, movie.MoviePath))
		},
	)

	p.popupOpenMovieInfo.Connect(
		"activate", func() {
			p.mainWindow.onEditMovieInfoClicked()
		},
	)

	p.popupOpenPack.Connect(
		"activate", func() {
			p.mainWindow.onOpenPackClicked()
		},
	)

	p.popupOpenIMDB.Connect(
		"activate", func() {
			p.mainWindow.onOpenIMDBClicked()
		},
	)

	p.popupPlayMovie.Connect(
		"activate", func() {
			p.mainWindow.onPlayMovieClicked()
		},
	)
}

func (p *popupMenu) createTagMenu(tags []data.Tag, movie *data.Movie, menu *gtk.Menu) {
	for i := 0; i < len(tags); i++ {
		tag := tags[i]
		item, err := gtk.CheckMenuItemNew()
		if err != nil {
			reportError(err)
			log.Fatal(err)
		}
		item.SetLabel(tag.Name)

		for i := 0; i < len(movie.Tags); i++ {
			if movie.Tags[i].Id == tag.Id {
				item.SetActive(true)
			}
		}

		menu.Add(item)

		p.addTagActivateEvent(item, movie, &tag)
	}
}

func (p *popupMenu) addTag(movie *data.Movie, tag *data.Tag) {
	movie.Tags = append(movie.Tags, *tag)
}

func (p *popupMenu) removeTag(movie *data.Movie, tag *data.Tag) {
	for i, t := range movie.Tags {
		if t.Id == tag.Id {
			movie.Tags = append(movie.Tags[:i], movie.Tags[i+1:]...)
		}
	}
}

func (p *popupMenu) addTagActivateEvent(item *gtk.CheckMenuItem, movie *data.Movie, tag *data.Tag) {
	item.Connect(
		"activate", func() {
			if item.GetActive() {
				err := p.mainWindow.database.InsertMovieTag(movie, tag)
				if err == nil {
					p.addTag(movie, tag)
				}
			} else {
				err := p.mainWindow.database.RemoveMovieTag(movie, tag)
				if err == nil {
					p.removeTag(movie, tag)
				}
			}
		},
	)
}
