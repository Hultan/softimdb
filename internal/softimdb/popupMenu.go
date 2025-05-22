package softimdb

import (
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/data"
)

type popupMenu struct {
	mainWindow *MainWindow
	popupMenu  *gtk.Menu

	popupGenres        *gtk.MenuItem
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

	p.popupGenres = p.mainWindow.builder.GetObject("popupGenres").(*gtk.MenuItem)
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

			genres, err := p.mainWindow.database.GetGenres()
			if err != nil {
				reportError(err)
				return
			}

			p.createGenreMenu(genres, movie, menu)
			p.popupGenres.SetSubmenu(menu)
			p.popupGenres.ShowAll()
			p.popupMenu.PopupAtPointer(event)
		},
	)

	p.popupOpenFolder.Connect(
		"activate", func() {
			p.mainWindow.onOpenFolderClicked()
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

func (p *popupMenu) createGenreMenu(genres []data.Genre, movie *data.Movie, menu *gtk.Menu) {
	for i := 0; i < len(genres); i++ {
		genre := genres[i]

		if showPrivateGenres || !genre.IsPrivate {
			item, err := gtk.CheckMenuItemNew()
			if err != nil {
				reportError(err)
				log.Fatal(err)
			}
			item.SetLabel(genre.Name)

			for i := 0; i < len(movie.Genres); i++ {
				if movie.Genres[i].Id == genre.Id {
					item.SetActive(true)
				}
			}

			menu.Add(item)

			p.addGenreActivateEvent(item, movie, &genre)
		}
	}
}

func (p *popupMenu) addGenre(movie *data.Movie, genre *data.Genre) {
	movie.Genres = append(movie.Genres, *genre)
}

func (p *popupMenu) removeGenre(movie *data.Movie, genre *data.Genre) {
	for i, t := range movie.Genres {
		if t.Id == genre.Id {
			movie.Genres = append(movie.Genres[:i], movie.Genres[i+1:]...)
		}
	}
}

func (p *popupMenu) addGenreActivateEvent(item *gtk.CheckMenuItem, movie *data.Movie, genre *data.Genre) {
	item.Connect(
		"activate", func() {
			if item.GetActive() {
				err := p.mainWindow.database.InsertMovieGenre(movie, genre)
				if err == nil {
					p.addGenre(movie, genre)
				}
			} else {
				err := p.mainWindow.database.RemoveMovieGenre(movie, genre)
				if err == nil {
					p.removeGenre(movie, genre)
				}
			}
		},
	)
}
