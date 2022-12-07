package softimdb

import (
	"errors"
	"fmt"
	"regexp"

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
	imdbIdEntry    *gtk.Entry
	moviePathEntry *gtk.Entry
	movieWindow    *MovieWindow
	database       *data.Database
	config         *config.Config
	builder        *framework.GtkBuilder
}

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
		_ = entry.Connect("changed", a.imdbURLChanged)
		entry = builder.GetObject("imdbIdEntry").(*gtk.Entry)
		a.imdbIdEntry = entry
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
	a.list = builder.GetObject("pathsList").(*gtk.ListBox)
	_ = a.list.Connect("row-activated", a.rowActivated)
	a.framework.Gtk.ClearListBox(a.list)
	a.fillList(a.list, *moviePaths)

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
			reportError(err)
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
	msg := "Are you sure you want to ignore this folder?"
	response := a.framework.Gtk.Title(applicationTitle).Text(msg).
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

func (a *AddWindow) addMovieButtonClicked() {

	url := a.getEntryText(a.imdbUrlEntry)
	if url == "" {
		a.framework.Gtk.Title(applicationTitle).Text("IMDB Url cannot be empty").
			ErrorIcon().OkButton().Show()
		return
	}
	moviePath := a.getEntryText(a.moviePathEntry)
	if moviePath == "" {
		a.framework.Gtk.Title(applicationTitle).Text("Movie path cannot be empty").
			ErrorIcon().OkButton().Show()
		return
	}

	key, err := imdb.NewApiKeyManagerFromStandardPath()
	if err != nil {
		a.framework.Gtk.Title(applicationTitle).Text("Failed to create new api key manager").
			ErrorIcon().OkButton().Show()
		return
	}

	id := a.getEntryText(a.imdbIdEntry)
	if id == "" {
		a.framework.Gtk.Title(applicationTitle).Text("IMDB id cannot be empty").
			ErrorIcon().OkButton().Show()
		return
	}
	manager := imdb.NewImdb(key)
	info, err := manager.Title(id)
	if err != nil {
		a.framework.Gtk.Title(applicationTitle).
			Text(fmt.Sprintf("Failed to retrieve movie information : \n\n%v", err)).
			ErrorIcon().OkButton().Show()
		return
	}

	movie, err := newMovieInfoFromImdb(info)
	if err != nil {
		reportError(err)
		panic(err)
	}

	if len(movie.image) > 0 {
		movie.imageHasChanged = true
	}

	// Open movie dialog here
	win := NewMovieWindow(movie, nil, a.saveMovieInfo)
	win.OpenForm(a.builder, a.window)
	a.movieWindow = win
}

func (a *AddWindow) saveMovieInfo(info *MovieInfo, _ *data.Movie) {
	newMovie := &data.Movie{}
	info.toDatabase(newMovie)
	newMovie.MoviePath = a.getEntryText(a.moviePathEntry)
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
	a.imdbUrlEntry.SetText("")
	a.imdbIdEntry.SetText("")
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

func (a *AddWindow) getTags(tags []string) []data.Tag {
	var dataTags []data.Tag

	for _, tag := range tags {
		dataTag := data.Tag{Name: tag}
		dataTags = append(dataTags, dataTag)
	}

	return dataTags
}

func (a *AddWindow) getIdFromUrl(url string) (string, error) {
	// Get the IMDB id from the URL.
	// Starts with tt and ends with 7 or 8 digits.
	re := regexp.MustCompile(`tt\d{7,8}`)
	matches := re.FindAll([]byte(url), -1)
	if len(matches) == 0 {
		err := errors.New("invalid imdb URL")
		return "", err
	}
	return string(matches[0]), nil
}

func (a *AddWindow) imdbURLChanged() {
	text, err := a.imdbUrlEntry.GetText()
	if err != nil {
		reportError(err)
		panic(err)
	}
	if text == "" {
		return
	}
	id, err := a.getIdFromUrl(text)
	if err != nil {
		reportError(err)
		return
	}
	a.imdbIdEntry.SetText(id)
}
