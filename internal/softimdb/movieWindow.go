package softimdb

import (
	"fmt"
	_ "image/jpeg"
	"log"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/hultan/dialog"

	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/builder"
	"github.com/hultan/softimdb/internal/data"
)

type movieWindow struct {
	window                   *gtk.Window
	pathEntry                *gtk.Entry
	imdbUrlEntry             *gtk.Entry
	titleEntry               *gtk.Entry
	subTitleEntry            *gtk.Entry
	yearEntry                *gtk.Entry
	myRatingEntry            *gtk.Entry
	toWatchCheckButton       *gtk.CheckButton
	needsSubtitleCheckButton *gtk.CheckButton
	storyLineEntry           *gtk.TextView
	ratingEntry              *gtk.Entry
	genresEntry              *gtk.Entry
	packEntry                *gtk.Entry
	posterImage              *gtk.Image
	runtimeEntry             *gtk.Entry
	deleteButton             *gtk.Button

	movieInfo *movieInfo
	movie     *data.Movie

	db *data.Database

	closeCallback func(gtk.ResponseType, *movieInfo, *data.Movie)
}

var isNew bool

func newMovieWindow(builder *builder.Builder, parent gtk.IWindow, db *data.Database) *movieWindow {
	m := &movieWindow{}

	m.db = db

	m.window = builder.GetObject("movieWindow").(*gtk.Window)
	m.window.SetTitle("Movie info window")
	m.window.SetTransientFor(parent)
	m.window.SetModal(true)
	m.window.SetKeepAbove(true)
	m.window.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)
	m.window.HideOnDelete()

	button := builder.GetObject("okButton").(*gtk.Button)
	_ = button.Connect("clicked", func() {
		if !m.saveMovie() {
			return
		}
		m.window.Hide()
		m.closeCallback(gtk.RESPONSE_ACCEPT, m.movieInfo, m.movie)
	})

	button = builder.GetObject("cancelButton").(*gtk.Button)
	_ = button.Connect("clicked", func() {
		m.window.Hide()
		m.closeCallback(gtk.RESPONSE_CANCEL, nil, nil)
	})

	button = builder.GetObject("deleteButton").(*gtk.Button)
	_ = button.Connect("clicked", func() {
		if m.deleteMovie() {
			m.window.Hide()
			m.closeCallback(gtk.RESPONSE_REJECT, nil, m.movie)
		}
	})
	m.deleteButton = button

	m.imdbUrlEntry = builder.GetObject("imdbUrlEntry").(*gtk.Entry)
	m.pathEntry = builder.GetObject("pathEntry").(*gtk.Entry)
	m.titleEntry = builder.GetObject("titleEntry").(*gtk.Entry)
	m.subTitleEntry = builder.GetObject("subTitleEntry").(*gtk.Entry)
	m.yearEntry = builder.GetObject("yearEntry").(*gtk.Entry)
	m.myRatingEntry = builder.GetObject("myRatingEntry").(*gtk.Entry)
	m.toWatchCheckButton = builder.GetObject("toWatchCheckButton").(*gtk.CheckButton)
	m.needsSubtitleCheckButton = builder.GetObject("needsSubtitleCheckButton").(*gtk.CheckButton)
	m.storyLineEntry = builder.GetObject("storyLineTextView").(*gtk.TextView)
	m.ratingEntry = builder.GetObject("ratingEntry").(*gtk.Entry)
	m.genresEntry = builder.GetObject("genresEntry").(*gtk.Entry)
	m.packEntry = builder.GetObject("packEntry").(*gtk.Entry)
	m.posterImage = builder.GetObject("posterImage").(*gtk.Image)
	m.runtimeEntry = builder.GetObject("runtimeEntry").(*gtk.Entry)
	eventBox := builder.GetObject("imageEventBox").(*gtk.EventBox)
	eventBox.Connect("button-press-event", m.onImageClick)

	m.titleEntry.Connect("focus-out-event", m.onTitleFocusOut)

	return m
}

func (m *movieWindow) open(info *movieInfo, movie *data.Movie, closeCallback func(gtk.ResponseType, *movieInfo,
	*data.Movie)) {

	if info == nil {
		info = &movieInfo{}
	}

	isNew = false
	if info.title == "" {
		//  New movie
		isNew = true
	}

	m.movie = movie
	m.movieInfo = info
	m.closeCallback = closeCallback
	m.deleteButton.SetSensitive(movie != nil)

	// Fill form with data
	m.imdbUrlEntry.SetText(m.movieInfo.imdbUrl)
	m.pathEntry.SetText(m.movieInfo.path)
	m.titleEntry.SetText(m.movieInfo.title)
	m.subTitleEntry.SetText(m.movieInfo.subTitle)
	m.yearEntry.SetText(fmt.Sprintf("%d", m.movieInfo.getYear()))
	m.myRatingEntry.SetText(fmt.Sprintf("%d", m.movieInfo.myRating))
	m.toWatchCheckButton.SetActive(m.movieInfo.toWatch)
	m.needsSubtitleCheckButton.SetActive(m.movieInfo.needsSubtitle)
	buffer, err := gtk.TextBufferNew(nil)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	buffer.SetText(m.movieInfo.storyLine)
	m.storyLineEntry.SetBuffer(buffer)
	m.ratingEntry.SetText(m.movieInfo.imdbRating)
	m.genresEntry.SetText(m.movieInfo.genres)
	m.packEntry.SetText(m.movieInfo.pack)
	m.runtimeEntry.SetText(strconv.Itoa(m.movieInfo.runtime))

	if m.movieInfo.image == nil {
		m.posterImage.Clear()
	} else {
		m.updateImage(m.movieInfo.image)
	}

	m.window.ShowAll()
	m.imdbUrlEntry.GrabFocus()
}

func (m *movieWindow) saveMovie() bool {
	// Fill fields
	m.movieInfo.path = getEntryText(m.pathEntry)
	m.movieInfo.imdbUrl = getEntryText(m.imdbUrlEntry)
	id, err := getIdFromUrl(m.movieInfo.imdbUrl)
	if err != nil {
		msg := fmt.Sprintf("Failed to retrieve IMDB id from url : %s", err)
		_, err = dialog.Title("Invalid IMDB url...").Text(msg).ErrorIcon().OkButton().Show()

		if err != nil {
			fmt.Printf("Error : %s", err)
		}

		return false
	}
	m.movieInfo.imdbId = id
	m.movieInfo.title = getEntryText(m.titleEntry)
	m.movieInfo.subTitle = getEntryText(m.subTitleEntry)
	m.movieInfo.pack = getEntryText(m.packEntry)
	m.movieInfo.year = getEntryText(m.yearEntry)

	ratingText := getEntryText(m.myRatingEntry)
	rating, err := strconv.Atoi(ratingText)
	legalRating := rating >= 0 && rating <= 5
	if err != nil || !legalRating {
		msg := fmt.Sprintf("Invalid my rating : %s (error : %s)", ratingText, err)
		_, err = dialog.Title("Invalid my rating...").Text(msg).ErrorIcon().OkButton().Show()

		if err != nil {
			fmt.Printf("Error : %s", err)
		}

		return false
	}
	m.movieInfo.myRating = rating

	runtimeText := getEntryText(m.runtimeEntry)
	runtime, err := strconv.Atoi(runtimeText)
	legalRuntime := runtime == -1 || runtime > 0
	if err != nil || !legalRuntime {
		msg := fmt.Sprintf("Invalid runtime : %s (error : %s)", runtimeText, err)
		_, err = dialog.Title("Invalid runtime...").Text(msg).ErrorIcon().OkButton().Show()

		if err != nil {
			fmt.Printf("Error : %s", err)
		}

		return false
	}
	m.movieInfo.runtime = runtime

	m.movieInfo.toWatch = m.toWatchCheckButton.GetActive()
	m.movieInfo.needsSubtitle = m.needsSubtitleCheckButton.GetActive()
	m.movieInfo.imdbRating = getEntryText(m.ratingEntry)
	buffer, err := m.storyLineEntry.GetBuffer()
	if err != nil {
		reportError(err)
		return false
	}
	storyLine, err := buffer.GetText(buffer.GetStartIter(), buffer.GetEndIter(), false)
	if err != nil {
		reportError(err)
		return false
	}
	m.movieInfo.storyLine = storyLine
	m.movieInfo.genres = getEntryText(m.genresEntry)
	// Poster is set when clicking on the image

	return true
}

func (m *movieWindow) deleteMovie() bool {
	response, err := dialog.Title("Delete movie...").Text("Do you want to delete this movie?").YesNoButtons().
		WarningIcon().Show()

	if err != nil {
		return false
	}

	if response == gtk.RESPONSE_YES {
		return true
	}

	return false
}

func (m *movieWindow) onImageClick() {
	dlg, err := gtk.FileChooserDialogNewWith2Buttons(
		"Choose an image...", nil, gtk.FILE_CHOOSER_ACTION_OPEN, "Ok", gtk.RESPONSE_OK,
		"Cancel", gtk.RESPONSE_CANCEL,
	)
	if err != nil {
		reportError(err)
		return
	}
	defer dlg.Destroy()

	home, err := os.UserHomeDir()
	if err == nil {
		dir := path.Join(home, "Downloads")
		_ = dlg.SetCurrentFolder(dir)
	}

	response := dlg.Run()
	if response == gtk.RESPONSE_CANCEL {
		return
	}

	fileData := getCorrectImageSize(dlg.GetFilename())
	if fileData == nil || len(fileData) == 0 {
		return
	}
	m.updateImage(fileData)
	m.movieInfo.image = fileData
	m.movieInfo.imageHasChanged = true
}

// updateImage updates the GtkImage
func (m *movieWindow) updateImage(image []byte) {
	// Image size: 190x280
	pix, err := gdk.PixbufNewFromBytesOnly(image)
	if err != nil {
		reportError(err)
		return
	}
	m.posterImage.SetFromPixbuf(pix)
}

func (m *movieWindow) onTitleFocusOut() {
	if !isNew {
		return
	}

	title, err := m.titleEntry.GetText()
	if err != nil {
		return
	}

	if title == "" {
		return
	}

	m.showSimilarMovies(m.findSimilarMovies(title))
}

func (m *movieWindow) showSimilarMovies(movies []movie) {
	titles := ""

	for i, mov := range movies {
		if i > 5 {
			break
		}
		titles += fmt.Sprintf("%s\n", mov.title)
	}

	if titles != "" {
		_, _ = dialog.Title("Similar movies...").Text("There are similar movies in the DB:\n\n" + titles).
			WarningIcon().OkButton().Show()
	}
}

type movie struct {
	distance int
	title    string
}

func (m *movieWindow) findSimilarMovies(title string) []movie {
	var movies []movie

	for _, movieTitle := range movieTitles {
		l := gstr.Levenshtein(movieTitle, title, 1, 3, 1)
		if containsI(title, movieTitle) {
			l = 2
		}
		if containsI(movieTitle, title) {
			l = 2
		}
		if equalsI(title, movieTitle) {
			l = 1
		}
		movies = append(movies, movie{l, movieTitle})
	}
	slices.SortFunc(movies, func(a, b movie) int {
		return a.distance - b.distance
	})
	//
	//fmt.Println("-----------------------------------")
	//for i := 0; i < 10; i++ {
	//	fmt.Println(movies[i])
	//}
	return movies[:5]
}

func containsI(a, b string) bool {
	return strings.Contains(strings.ToLower(b), strings.ToLower(a))
}

func equalsI(a, b string) bool {
	return strings.ToLower(b) == strings.ToLower(a)
}
