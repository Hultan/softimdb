package softimdb

import (
	"errors"
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/hultan/dialog"
	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
	"github.com/hultan/softimdb/internal/nas"
)

type addMovieWindow struct {
	mainWindow     *MainWindow
	window         *gtk.Window
	list           *gtk.ListBox
	moviePathEntry *gtk.Entry
	database       *data.Database
	config         *config.Config
}

func newAddMovieWindow(m *MainWindow, db *data.Database, cfg *config.Config) *addMovieWindow {
	a := &addMovieWindow{
		mainWindow: m,
		database:   db,
		config:     cfg,
	}

	a.window = m.builder.GetObject("addWindow").(*gtk.Window)
	a.window.SetTitle("Add movie window")
	//a.window.SetModal(true)
	a.window.SetKeepAbove(true)
	a.window.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)
	a.window.HideOnDelete()

	button := m.builder.GetObject("closeButton").(*gtk.Button)
	_ = button.Connect("clicked", func() {
		a.window.Hide()
	})
	ignoreButton := m.builder.GetObject("ignorePathButton").(*gtk.Button)
	_ = ignoreButton.Connect("clicked", a.onIgnorePathButtonClicked)
	addMovieButton := m.builder.GetObject("addMovieButton").(*gtk.Button)
	_ = addMovieButton.Connect("clicked", a.onAddMovieButtonClicked)

	entry := m.builder.GetObject("moviePathEntry").(*gtk.Entry)
	a.moviePathEntry = entry

	a.list = m.builder.GetObject("pathsList").(*gtk.ListBox)
	_ = a.list.Connect("row-activated", a.onRowActivated)

	return a
}

func (a *addMovieWindow) open() {
	clearListBox(a.list)
	label, err := gtk.LabelNew("Looking for new videos...please wait...")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	label.SetHAlign(gtk.ALIGN_START)

	a.list.Add(label)
	a.window.ShowAll()
	a.window.QueueDraw()

	go a.findNewMovies()
}

func (a *addMovieWindow) findNewMovies() {
	// Find new paths on NAS
	nasManager := nas.ManagerNew(a.database)
	moviePaths, err := nasManager.GetMovies(a.config)
	if err != nil {
		_, _ = dialog.Title("Error").
			ErrorIcon().
			Text("Failed to access NAS or database...").
			ExtraExpand(err.Error()).
			OkButton().
			Show()

		return
	}

	clearListBox(a.list)

	if len(moviePaths) <= 0 {
		label, err := gtk.LabelNew("No new movies found...")
		if err != nil {
			reportError(err)
			log.Fatal(err)
		}
		label.SetHAlign(gtk.ALIGN_START)
		a.list.Add(label)
	} else {
		a.fillList(a.list, moviePaths)
	}

	a.window.ShowAll()
	a.window.QueueDraw()
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
	// the list is empty, since GetRowAtIndex returns
	// nil, and SelectRow can handle nil.
	row := list.GetRowAtIndex(0)
	list.SelectRow(row)
	a.onRowActivated()
}

func (a *addMovieWindow) windowClosed(r gtk.ResponseType, info *movieInfo, movie *data.Movie) {
	switch r {
	case gtk.RESPONSE_ACCEPT:
		// Save movie
		a.insertMovie(info, movie)
	case gtk.RESPONSE_CANCEL:
		// Cancel dialog
	default:
		// gtk.RESPONSE_REJECT should not happen from add movie window
		// Unknown response
		// Handle as cancel
	}
}

func (a *addMovieWindow) insertMovie(info *movieInfo, _ *data.Movie) {
	newMovie := &data.Movie{}
	info.toDatabase(newMovie)

	if err := a.database.InsertMovie(newMovie); err != nil {
		reportError(fmt.Errorf("failed to insert movie: %w", err))
		return
	}

	row := a.list.GetSelectedRow()
	if row == nil {
		reportError(errors.New("no row selected in movie list"))
		return
	}

	a.list.Remove(row)
	a.moviePathEntry.SetText("")
}

func (a *addMovieWindow) onIgnorePathButtonClicked() {
	row := a.list.GetSelectedRow()
	if row == nil {
		return
	}
	widget, err := row.GetChild()
	if err != nil {
		reportError(fmt.Errorf("failed to get child widget from row: %w", err))
		return
	}

	label, ok := widget.(*gtk.Label)
	if !ok {
		reportError(errors.New("expected widget to be a *gtk.Label"))
		return
	}

	path, err := label.GetText()
	if err != nil {
		reportError(fmt.Errorf("failed to get text from label: %w", err))
		return
	}

	response, err := dialog.Title(applicationTitle).
		Text("Ignore folder?").
		ExtraExpandf("Are you sure you want to ignore the folder '%s'?", path).
		ExtraHeight(70).
		QuestionIcon().YesNoButtons().Show()
	if err != nil {
		reportError(fmt.Errorf("failed to show dialog: %w", err))
		return
	}
	if response == gtk.RESPONSE_NO {
		return
	}

	// Save to DB
	ignorePath := data.IgnoredPath{Path: path}
	if err := a.database.InsertIgnorePath(&ignorePath); err != nil {
		reportError(fmt.Errorf("failed to insert ignore path: %w", err))
		return
	}

	a.list.Remove(row)
}

func (a *addMovieWindow) onAddMovieButtonClicked() {
	moviePath := getEntryText(a.moviePathEntry)
	if moviePath == "" {
		_, err := dialog.Title(applicationTitle).Text("Movie path cannot be empty").
			ErrorIcon().OkButton().Show()
		if err != nil {
			fmt.Printf("Error : %s", err)
		}

		return
	}

	info := &movieInfo{moviePath: moviePath}

	// Open the movie dialog here
	if a.mainWindow.movieWin == nil {
		a.mainWindow.movieWin = newMovieWindow(a.mainWindow.builder, a.window, a.database, a.config)
	}
	a.mainWindow.movieWin.open(info, nil, a.windowClosed)
}

func (a *addMovieWindow) onRowActivated() {
	row := a.list.GetSelectedRow()
	if row == nil {
		return
	}

	labelObj, err := row.GetChild()
	if err != nil {
		reportError(errors.New("failed to get child widget from row"))
		return
	}

	label, ok := labelObj.(*gtk.Label)
	if !ok {
		reportError(errors.New("expected labelObj to be a *gtk.Label"))
		return
	}

	path, err := label.GetText()
	if err != nil {
		reportError(fmt.Errorf("failed to get text from label: %w", err))
		return
	}

	a.moviePathEntry.SetText(path)
}
func (a *addMovieWindow) getGenres(genres []string) []data.Genre {
	dataGenres := make([]data.Genre, len(genres))

	for i, genre := range genres {
		dataGenres[i] = data.Genre{Name: genre}
	}

	return dataGenres
}
