package softimdb

import (
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hultan/dialog"
	"github.com/hultan/softimdb/internal/builder"
	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
	"github.com/hultan/softimdb/internal/nas"
	"log"
)

type addMovieWindow struct {
	window         *gtk.Window
	list           *gtk.ListBox
	moviePathEntry *gtk.Entry
	database       *data.Database
	config         *config.Config
	builder        *builder.Builder
}

func newAddMovieWindow() *addMovieWindow {
	return new(addMovieWindow)
}

func (a *addMovieWindow) openForm(builder *builder.Builder, database *data.Database, config *config.Config) {
	a.builder = builder
	if a.window == nil {
		// Get the extra window from glade
		addWindow := builder.GetObject("addWindow").(*gtk.Window)

		// Set up the extra window
		addWindow.SetTitle("Add movie window")
		addWindow.SetModal(true)
		addWindow.SetKeepAbove(true)
		addWindow.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)

		// Hook up the destroy event
		_ = addWindow.Connect("delete-event", a.onCloseWindow)

		// Close button
		button := builder.GetObject("closeButton").(*gtk.Button)
		_ = button.Connect("clicked", a.onCloseWindow)

		// Ignore Path Button
		ignoreButton := builder.GetObject("ignorePathButton").(*gtk.Button)
		_ = ignoreButton.Connect("clicked", a.onIgnorePathButtonClicked)

		// Add Movie Button
		addMovieButton := builder.GetObject("addMovieButton").(*gtk.Button)
		_ = addMovieButton.Connect("clicked", a.onAddMovieButtonClicked)

		entry := builder.GetObject("moviePathEntry").(*gtk.Entry)
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

		_, err := dialog.Title("Error").
			ErrorIcon().
			Text("Failed to access NAS, is it unlocked?").
			Show()

		if err != nil {
			fmt.Printf("Error : %s", err)
		}
		return
	}

	// Paths list
	a.list = builder.GetObject("pathsList").(*gtk.ListBox)
	_ = a.list.Connect("row-activated", a.rowActivated)
	clearListBox(a.list)
	a.fillList(a.list, *moviePaths)

	// Show the window
	a.window.Present()
}

func (a *addMovieWindow) onCloseWindow() {
	a.window.Destroy()
}

func (a *addMovieWindow) fillList(list *gtk.ListBox, paths []string) {
	for i := range paths {
		label, err := gtk.LabelNew(paths[i])
		if err != nil {
			reportError(err)
			log.Fatal(err)
		}
		label.SetHAlign(gtk.ALIGN_START)
		list.Add(label)
	}
	// Select the first row, this won't crash if
	// list is empty, since GetRowAtIndex returns
	// nil, and SelectRow can handle nil.
	row := list.GetRowAtIndex(0)
	list.SelectRow(row)
	a.rowActivated()
}

func (a *addMovieWindow) onIgnorePathButtonClicked() {
	msg := "Are you sure you want to ignore this folder?"
	response, _ := dialog.Title(applicationTitle).Text(msg).
		QuestionIcon().YesNoButtons().Show()
	if response == gtk.RESPONSE_NO {
		return
	}

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

func (a *addMovieWindow) onAddMovieButtonClicked() {
	moviePath := a.getEntryText(a.moviePathEntry)
	if moviePath == "" {
		_, err := dialog.Title(applicationTitle).Text("Movie path cannot be empty").
			ErrorIcon().OkButton().Show()

		if err != nil {
			fmt.Printf("Error : %s", err)
		}

		return
	}

	info := &movieInfo{
		path: moviePath,
	}

	// Open movie dialog here
	win := newMovieWindow(info, nil, a.saveMovieInfo)
	win.openForm(a.builder, a.window)
}

func (a *addMovieWindow) saveMovieInfo(info *movieInfo, _ *data.Movie) {
	newMovie := &data.Movie{}
	info.toDatabase(newMovie)
	err := a.database.InsertMovie(newMovie)
	if err != nil {
		reportError(err)
		return
	}

	// Get selected row and remove it
	row := a.list.GetSelectedRow()
	if row == nil {
		reportError(err)
		return
	}

	a.list.Remove(row)
	a.moviePathEntry.SetText("")
}

func (a *addMovieWindow) getEntryText(entry *gtk.Entry) string {
	text, err := entry.GetText()
	if err != nil {
		return ""
	}
	return text
}

func (a *addMovieWindow) rowActivated() {
	row := a.list.GetSelectedRow()
	if row == nil {
		return
	}
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

func (a *addMovieWindow) getTags(tags []string) []data.Tag {
	var dataTags []data.Tag

	for _, tag := range tags {
		dataTag := data.Tag{Name: tag}
		dataTags = append(dataTags, dataTag)
	}

	return dataTags
}

// clearListBox : Clears a gtk.ListBox
func clearListBox(list *gtk.ListBox) {
	children := list.GetChildren()
	if children == nil {
		return
	}
	var i uint = 0
	for i < children.Length() {
		widget, _ := children.NthData(i).(*gtk.Widget)
		list.Remove(widget)
		i++
	}
}
