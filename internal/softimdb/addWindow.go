package softimdb

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
	"github.com/hultan/softimdb/internal/imdb"
	"github.com/hultan/softimdb/internal/nas"
	"github.com/hultan/softteam/framework"
)

type AddWindow struct {
	framework      *framework.Framework
	window         *gtk.Window
	list           *gtk.ListBox
	imdbUrlEntry   *gtk.Entry
	moviePathEntry *gtk.Entry

	database *data.Database
	config   *config.Config
	builder  *framework.GtkBuilder
}

var currentMovie *data.Movie
var currentMovieInfo *imdb.MovieInfo

func AddWindowNew(framework *framework.Framework) *AddWindow {
	a := new(AddWindow)
	a.framework = framework
	return a
}

func (a *AddWindow) OpenForm(builder *framework.GtkBuilder, database *data.Database, config *config.Config) {
	a.builder = builder
	if a.window == nil {
		// Get the extra window from glade
		addWindow := builder.GetObject("addWindow").(*gtk.Window)

		// Set up the extra window
		addWindow.SetTitle("Add movie window")
		addWindow.HideOnDelete()
		addWindow.SetModal(true)
		addWindow.SetKeepAbove(true)
		addWindow.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)

		// Hook up the destroy event
		_ = addWindow.Connect("delete-event", a.closeWindow)

		// Close button
		button := builder.GetObject("closeButton").(*gtk.Button)
		_ = button.Connect("clicked", a.closeWindow)

		// Ignore Path Button
		ignoreButton := builder.GetObject("ignorePathButton").(*gtk.Button)
		_ = ignoreButton.Connect("clicked", a.ignorePathButtonClicked)

		// Add Movie Button
		addMovieButton := builder.GetObject("addMovieButton").(*gtk.Button)
		_ = addMovieButton.Connect("clicked", a.addMovieButtonClicked)

		// IMDB Url and Movie Path entry
		entry := builder.GetObject("imdbEntry").(*gtk.Entry)
		a.imdbUrlEntry = entry
		entry = builder.GetObject("moviePathEntry").(*gtk.Entry)
		a.moviePathEntry = entry

		// Store reference to database and window
		a.database = database
		a.window = addWindow
		a.config = config
	}

	// Paths on NAS
	nasManager := nas.ManagerNew(a.database)
	moviePaths := nasManager.GetMovies(a.config)
	if moviePaths == nil {
		a.window.ShowAll()
		message := "Failed to access NAS, is it unlocked?"
		dialog := gtk.MessageDialogNew(nil, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, message)
		dialog.Run()
		dialog.Destroy()
		return
	}
	nasManager.Disconnect()

	// Paths list
	list := builder.GetObject("pathsList").(*gtk.ListBox)
	_ = list.Connect("row-activated", a.rowActivated)
	a.list = list
	a.framework.Gtk.ClearListBox(a.list)
	a.fillList(list, *moviePaths)

	// Show the window
	a.window.ShowAll()

	a.imdbUrlEntry.GrabFocus()
}

func (a *AddWindow) closeWindow() {
	a.window.Hide()
}

func (a *AddWindow) fillList(list *gtk.ListBox, paths []string) {
	for i := range paths {
		label, err := gtk.LabelNew(paths[i])
		label.SetHAlign(gtk.ALIGN_START)
		if err != nil {
			panic(err)
		}
		list.Add(label)
	}
	// Select the first row, this won't crash if
	// list is empty, since GetRowAtIndex returns
	// nil, and SelectRow can handle nil.
	row := list.GetRowAtIndex(0)
	list.SelectRow(row)
	a.rowActivated()
}

func (a *AddWindow) ignorePathButtonClicked() {
	row := a.list.GetSelectedRow()
	if row == nil {
		return
	}
	widget, err := row.GetChild()
	if err != nil {
		return
	}

	label, ok := widget.(*gtk.Label)
	if ok {
		path, err := label.GetText()
		if err != nil {
			return
		}
		ignorePath := data.IgnoredPath{Path: path}
		err = a.database.InsertIgnorePath(&ignorePath)
		if err != nil {
			return
		}

		a.list.Remove(row)
	}
}

func (a *AddWindow) addMovieButtonClicked() {
	url := a.getEntryText(a.imdbUrlEntry)
	if url == "" {
		message := "IMDB Url cannot be empty"
		dialog := gtk.MessageDialogNew(a.window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, message)
		dialog.Run()
		dialog.Destroy()
		return
	}
	moviePath := a.getEntryText(a.moviePathEntry)
	if moviePath == "" {
		message := "Movie path cannot be empty"
		dialog := gtk.MessageDialogNew(a.window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, message)
		dialog.Run()
		dialog.Destroy()
		return
	}

	manager := imdb.ManagerNew()
	currentMovie = &data.Movie{ImdbUrl: url, MoviePath: moviePath}
	var err error
	currentMovieInfo, err = manager.GetMovieInfo(currentMovie)
	if err != nil {
		message := fmt.Sprintf("Failed to retrieve movie information : \n\n%v", err)
		dialog := gtk.MessageDialogNew(a.window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, message)
		dialog.Run()
		dialog.Destroy()
		return
	}

	// Get IMDB id
	currentMovie.ImdbID = currentMovie.ImdbUrl[28 : 28+9]

	// Open movie dialog here
	movieDialog := MovieWindowNew(currentMovieInfo, a.saveMovieInfo)
	movieDialog.OpenForm(a.builder, a.window)
}

func (a *AddWindow) saveMovieInfo() {
	// Store data
	currentMovie.Title = currentMovieInfo.Title
	currentMovie.Year = currentMovieInfo.Year
	currentMovie.Image = &currentMovieInfo.Poster
	currentMovie.HasImage = true
	currentMovie.ImdbRating = float32(currentMovieInfo.Rating)
	currentMovie.StoryLine = currentMovieInfo.StoryLine
	currentMovie.Tags = a.getTags(currentMovieInfo.Tags)

	err := a.database.InsertMovie(currentMovie)
	if err != nil {
		panic(err)
	}

	// Get selected row and remove it
	row := a.list.GetSelectedRow()
	if row == nil {
		return
	}

	a.list.Remove(row)
	a.imdbUrlEntry.SetText("")
	a.moviePathEntry.SetText("")
}

func (a *AddWindow) getEntryText(entry *gtk.Entry) string {
	text, err := entry.GetText()
	if err != nil {
		return ""
	}
	return text
}

func (a *AddWindow) rowActivated() {
	row := a.list.GetSelectedRow()
	labelObj, err := row.GetChild()
	if err != nil {
		return
	}
	label := labelObj.(*gtk.Label)
	path, err := label.GetText()
	if err != nil {
		return
	}
	a.moviePathEntry.SetText(path)
}

func (a *AddWindow) getTags(tags []string) []data.Tag {
	var dataTags []data.Tag

	for _, tag := range tags {
		dataTag := data.Tag{Name: tag}
		dataTags = append(dataTags, dataTag)
	}

	return dataTags
}
