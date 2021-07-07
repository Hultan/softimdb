package softimdb

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/hultan/softimdb/internal/data"
	imdb2 "github.com/hultan/softimdb/internal/imdb"
	"github.com/hultan/softimdb/internal/nas"
)

type AddWindow struct {
	Window         *gtk.Window
	list           *gtk.ListBox
	imdbUrlEntry   *gtk.Entry
	moviePathEntry *gtk.Entry
}

func AddWindowNew() *AddWindow {
	return new(AddWindow)
}

func (a *AddWindow) OpenForm(builder *SoftBuilder) {
	if a.Window == nil {
		// Get the extra window from glade
		addWindow := builder.getObject("addWindow").(*gtk.Window)

		// Set up the extra window
		addWindow.SetTitle("Add movie window")
		addWindow.HideOnDelete()
		addWindow.SetModal(true)
		addWindow.SetKeepAbove(true)
		addWindow.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)

		// Hook up the destroy event
		_ = addWindow.Connect("delete-event", a.closeWindow)

		// Close button
		button := builder.getObject("closeButton").(*gtk.Button)
		_ = button.Connect("clicked", a.closeWindow)

		// Ignore Path Button
		ignoreButton := builder.getObject("ignorePathButton").(*gtk.Button)
		_ = ignoreButton.Connect("clicked", a.ignorePathButtonClicked)

		// Add Movie Button
		addMovieButton := builder.getObject("addMovieButton").(*gtk.Button)
		_ = addMovieButton.Connect("clicked", a.addMovieButtonClicked)

		// IMDB Url and Movie Path entry
		entry := builder.getObject("imdbEntry").(*gtk.Entry)
		a.imdbUrlEntry = entry
		entry = builder.getObject("moviePathEntry").(*gtk.Entry)
		a.moviePathEntry = entry

		a.Window = addWindow
	}

	// Paths on NAS
	nasManager := nas.ManagerNew()
	moviePaths := nasManager.GetMovies()

	// Paths list
	list := builder.getObject("pathsList").(*gtk.ListBox)
	//_ = button.Connect("clicked", a.closeWindow)
	a.list = list
	a.clearList()
	a.fillList(list, *moviePaths)

	// Show the window
	a.Window.ShowAll()
}

func (a *AddWindow) closeWindow() {
	a.Window.Hide()
}

func (a *AddWindow) fillList(list *gtk.ListBox, paths []string) {
	for i := range paths {
		label, err := gtk.LabelNew(paths[i])
		if err != nil {
			panic(err)
		}
		list.Add(label)
	}
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
		db := data.DatabaseNew(false)
		ignorePath := data.IgnoredPath{Path: path}
		err = db.InsertIgnorePath(&ignorePath)
		if err != nil {
			return
		}
		db.CloseDatabase()

		a.list.Remove(row)
	}
}

func (a *AddWindow) addMovieButtonClicked() {
	url := a.getEntryText(a.imdbUrlEntry)
	if url == "" {
		return
	}
	moviePath := a.getEntryText(a.moviePathEntry)
	imdb := imdb2.ManagerNew()
	movie := data.Movie{ImdbUrl: url, MoviePath: moviePath}
	err := imdb.GetMovieInfo(&movie)
	if err != nil {
		panic(err)
	}

	db := data.DatabaseNew(false)
	err = db.InsertMovie(&movie)
	if err != nil {
		panic(err)
	}
	db.CloseDatabase()

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

func (a *AddWindow) clearList() {
	children := a.list.GetChildren()
	if children==nil {
		return
	}
	var i uint = 0
	for ;i<children.Length(); {
		widget, _ := children.NthData(i).(*gtk.Widget)
		a.list.Remove(widget)
		i++
	}
}