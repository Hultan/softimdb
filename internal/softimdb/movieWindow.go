package softimdb

import (
	"fmt"
	"github.com/hultan/dialog"
	_ "image/jpeg"
	"log"
	"strconv"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/builder"
	"github.com/hultan/softimdb/internal/data"
)

type movieWindow struct {
	window             *gtk.Window
	pathEntry          *gtk.Entry
	imdbUrlEntry       *gtk.Entry
	titleEntry         *gtk.Entry
	subTitleEntry      *gtk.Entry
	yearEntry          *gtk.Entry
	myRatingEntry      *gtk.Entry
	toWatchCheckButton *gtk.CheckButton
	storyLineEntry     *gtk.TextView
	ratingEntry        *gtk.Entry
	genresEntry        *gtk.Entry
	packEntry          *gtk.Entry
	posterImage        *gtk.Image

	movieInfo *movieInfo
	movie     *data.Movie

	saveCallback func(*movieInfo, *data.Movie)
}

func newMovieWindow(builder *builder.Builder, parent gtk.IWindow) *movieWindow {
	m := &movieWindow{}

	m.window = builder.GetObject("movieWindow").(*gtk.Window)
	m.window.SetTitle("Movie info window")
	m.window.SetTransientFor(parent)
	m.window.SetModal(true)
	m.window.SetKeepAbove(true)
	m.window.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)
	m.window.HideOnDelete()

	button := builder.GetObject("okButton").(*gtk.Button)
	_ = button.Connect("clicked", func() {
		m.saveMovie()
		m.window.Hide()
	})
	button = builder.GetObject("cancelButton").(*gtk.Button)
	_ = button.Connect("clicked", func() {
		m.window.Hide()
	})

	m.imdbUrlEntry = builder.GetObject("imdbUrlEntry").(*gtk.Entry)
	m.pathEntry = builder.GetObject("pathEntry").(*gtk.Entry)
	m.titleEntry = builder.GetObject("titleEntry").(*gtk.Entry)
	m.subTitleEntry = builder.GetObject("subTitleEntry").(*gtk.Entry)
	m.yearEntry = builder.GetObject("yearEntry").(*gtk.Entry)
	m.myRatingEntry = builder.GetObject("myRatingEntry").(*gtk.Entry)
	m.toWatchCheckButton = builder.GetObject("toWatchCheckButton").(*gtk.CheckButton)
	m.storyLineEntry = builder.GetObject("storyLineTextView").(*gtk.TextView)
	m.ratingEntry = builder.GetObject("ratingEntry").(*gtk.Entry)
	m.genresEntry = builder.GetObject("genresEntry").(*gtk.Entry)
	m.packEntry = builder.GetObject("packEntry").(*gtk.Entry)
	m.posterImage = builder.GetObject("posterImage").(*gtk.Image)
	eventBox := builder.GetObject("imageEventBox").(*gtk.EventBox)
	eventBox.Connect("button-press-event", m.onImageClick)

	return m
}

func (m *movieWindow) open(info *movieInfo, movie *data.Movie, saveCallback func(*movieInfo, *data.Movie)) {
	if info == nil {
		info = &movieInfo{}
	}
	m.movie = movie
	m.movieInfo = info
	m.saveCallback = saveCallback

	// Fill form with data
	m.imdbUrlEntry.SetText(m.movieInfo.imdbUrl)
	m.pathEntry.SetText(m.movieInfo.path)
	m.titleEntry.SetText(m.movieInfo.title)
	m.subTitleEntry.SetText(m.movieInfo.subTitle)
	m.yearEntry.SetText(fmt.Sprintf("%d", m.movieInfo.getYear()))
	m.myRatingEntry.SetText(fmt.Sprintf("%d", m.movieInfo.myRating))
	m.toWatchCheckButton.SetActive(m.movieInfo.toWatch)
	buffer, err := gtk.TextBufferNew(nil)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	buffer.SetText(m.movieInfo.storyLine)
	m.storyLineEntry.SetBuffer(buffer)
	m.ratingEntry.SetText(m.movieInfo.imdbRating)
	m.genresEntry.SetText(m.movieInfo.tags)
	m.packEntry.SetText(m.movieInfo.pack)
	if m.movieInfo.image == nil {
		m.posterImage.Clear()
	} else {
		m.updateImage(m.movieInfo.image)
	}

	m.window.ShowAll()
	m.imdbUrlEntry.GrabFocus()
}

func (m *movieWindow) saveMovie() {
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

		return
	}
	m.movieInfo.imdbId = id
	m.movieInfo.title = getEntryText(m.titleEntry)
	m.movieInfo.subTitle = getEntryText(m.subTitleEntry)
	m.movieInfo.pack = getEntryText(m.packEntry)
	m.movieInfo.year = getEntryText(m.yearEntry)
	ratingText := getEntryText(m.myRatingEntry)
	rating, err := strconv.Atoi(ratingText)
	if err != nil || rating < 0 || rating > 5 {
		msg := fmt.Sprintf("Invalid my rating : %s (error : %s)", ratingText, err)
		_, err = dialog.Title("Invalid my rating...").Text(msg).ErrorIcon().OkButton().Show()

		if err != nil {
			fmt.Printf("Error : %s", err)
		}

		return
	}
	m.movieInfo.myRating = rating
	m.movieInfo.toWatch = m.toWatchCheckButton.GetActive()
	m.movieInfo.imdbRating = getEntryText(m.ratingEntry)
	buffer, err := m.storyLineEntry.GetBuffer()
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	storyLine, err := buffer.GetText(buffer.GetStartIter(), buffer.GetEndIter(), false)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	m.movieInfo.storyLine = storyLine
	m.movieInfo.tags = getEntryText(m.genresEntry)
	// Poster is set when clicking on the image

	m.saveCallback(m.movieInfo, m.movie)
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
