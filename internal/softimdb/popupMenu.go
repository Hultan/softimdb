package softimdb

import (
	"fmt"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/data"
)

type PopupMenu struct {
	mainWindow *MainWindow
	popupMenu  *gtk.Menu

	popupTags        *gtk.MenuItem
	popupOpenFolder  *gtk.MenuItem
	popupRefreshIMDB *gtk.MenuItem
	popupOpenIMDB    *gtk.MenuItem
	popupPlayMovie   *gtk.MenuItem
	popupUpdateImage *gtk.MenuItem
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
	p.popupOpenIMDB = p.mainWindow.builder.GetObject("popupOpenIMDBPage").(*gtk.MenuItem)
	p.popupRefreshIMDB = p.mainWindow.builder.GetObject("popupRefreshIMDB").(*gtk.MenuItem)
	p.popupPlayMovie = p.mainWindow.builder.GetObject("popupPlayMovie").(*gtk.MenuItem)
	p.popupUpdateImage = p.mainWindow.builder.GetObject("popupUpdateImage").(*gtk.MenuItem)

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
							err = p.mainWindow.database.InsertMovieTag(movie, &tag)
							if err == nil {
								p.addTag(movie, &tag)
							}
						} else {
							err = p.mainWindow.database.RemoveMovieTag(movie, &tag)
							if err == nil {
								p.removeTag(movie, &tag)
							}
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
		if movie == nil {
			return
		}
		p.mainWindow.openMovieDirectoryInNemo(movie)
	})

	p.popupOpenIMDB.Connect("activate", func() {
		p.mainWindow.openIMDB()
	})

	p.popupRefreshIMDB.Connect("activate", func() {
		p.mainWindow.refreshIMDB()
	})

	p.popupPlayMovie.Connect("activate", func() {
		movie := p.mainWindow.getSelectedMovie()
		if movie == nil {
			return
		}
		p.mainWindow.playMovie(movie)
	})

	p.popupUpdateImage.Connect("activate", func() {
		dialog, err := gtk.FileChooserDialogNewWith2Buttons("Choose new image...", p.mainWindow.window, gtk.FILE_CHOOSER_ACTION_OPEN, "Ok", gtk.RESPONSE_OK,
			"Cancel", gtk.RESPONSE_CANCEL)
		if err != nil {
			panic(err)
		}
		defer dialog.Destroy()

		response := dialog.Run()
		if response == gtk.RESPONSE_CANCEL {
			return
		}

		movie := p.mainWindow.getSelectedMovie()
		if movie == nil {
			return
		}

		fileName := dialog.GetFilename()
		file, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Printf("Could not read the file due to this %s error \n", err)
		}
		image := &data.Image{Data: &file}
		err = p.mainWindow.database.InsertImage(image)
		if err != nil {
			panic(err)
		}

		movie.ImageId = image.Id
		err = p.mainWindow.database.UpdateMovie(movie)
		if err != nil {
			panic(err)
		}
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
