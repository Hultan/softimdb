package softimdb

import (
	"bytes"
	"fmt"
	"github.com/hultan/dialog"
	"github.com/nfnt/resize"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"strconv"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/builder"
	"github.com/hultan/softimdb/internal/data"
)

type MovieWindow struct {
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
	posterImage        *gtk.Image

	movieInfo *MovieInfo
	movie     *data.Movie

	saveCallback func(*MovieInfo, *data.Movie)
}

func NewMovieWindow(info *MovieInfo, movie *data.Movie, saveCallback func(*MovieInfo, *data.Movie)) *MovieWindow {
	if info == nil {
		info = &MovieInfo{}
	}
	return &MovieWindow{movieInfo: info, movie: movie, saveCallback: saveCallback}
}

func (m *MovieWindow) OpenForm(builder *builder.Builder, parent gtk.IWindow) {
	if m.window == nil {
		// Get the extra window from glade
		movieWindow := builder.GetObject("movieWindow").(*gtk.Window)

		// Set up the extra window
		movieWindow.SetTitle("Movie info window")
		movieWindow.HideOnDelete()
		movieWindow.SetTransientFor(parent)
		movieWindow.SetModal(true)
		movieWindow.SetKeepAbove(true)
		movieWindow.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)

		// Hook up the destroy event
		_ = movieWindow.Connect("delete-event", m.closeWindow)

		// Buttons
		button := builder.GetObject("okButton").(*gtk.Button)
		_ = button.Connect("clicked", m.okButtonClicked)
		button = builder.GetObject("cancelButton").(*gtk.Button)
		_ = button.Connect("clicked", m.closeWindow)

		// Entries and images
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
		m.posterImage = builder.GetObject("posterImage").(*gtk.Image)
		eventBox := builder.GetObject("imageEventBox").(*gtk.EventBox)
		eventBox.Connect("button-press-event", m.onImageClick)

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
			panic(err)
		}
		buffer.SetText(m.movieInfo.storyLine)
		m.storyLineEntry.SetBuffer(buffer)
		m.ratingEntry.SetText(m.movieInfo.imdbRating)
		m.genresEntry.SetText(m.movieInfo.tags)

		// Poster
		if m.movieInfo.image != nil {
			m.updateImage(m.movieInfo.image)
		}

		// Store reference to database and window
		m.window = movieWindow
	}

	// Show the window
	m.window.ShowAll()

	m.imdbUrlEntry.GrabFocus()
}

func (m *MovieWindow) closeWindow() {
	m.window.Hide()
}

func (m *MovieWindow) getEntryText(entry *gtk.Entry) string {
	text, err := entry.GetText()
	if err != nil {
		return ""
	}
	return text
}

func (m *MovieWindow) okButtonClicked() {
	// Fill fields
	m.movieInfo.path = m.getEntryText(m.pathEntry)
	m.movieInfo.imdbUrl = m.getEntryText(m.imdbUrlEntry)
	id, err := getIdFromUrl(m.movieInfo.imdbUrl)
	if err != nil {
		msg := fmt.Sprintf("Failed to retrieve IMDB id from url : %s", err)
		dialog.Title("Invalid IMDB url...").Text(msg).ErrorIcon().OkButton().Show()
		return
	}
	m.movieInfo.imdbId = id
	m.movieInfo.title = m.getEntryText(m.titleEntry)
	m.movieInfo.subTitle = m.getEntryText(m.subTitleEntry)
	m.movieInfo.year = m.getEntryText(m.yearEntry)
	ratingText := m.getEntryText(m.myRatingEntry)
	rating, err := strconv.Atoi(ratingText)
	if err != nil || rating < 0 || rating > 5 {
		msg := fmt.Sprintf("Invalid my rating : %s (error : %s)", ratingText, err)
		dialog.Title("Invalid my rating...").Text(msg).ErrorIcon().OkButton().Show()
		return
	}
	m.movieInfo.myRating = rating
	m.movieInfo.toWatch = m.toWatchCheckButton.GetActive()
	m.movieInfo.imdbRating = m.getEntryText(m.ratingEntry)
	buffer, err := m.storyLineEntry.GetBuffer()
	if err != nil {
		reportError(err)
		panic(err)
	}
	storyLine, err := buffer.GetText(buffer.GetStartIter(), buffer.GetEndIter(), false)
	if err != nil {
		reportError(err)
		panic(err)
	}
	m.movieInfo.storyLine = storyLine
	m.movieInfo.tags = m.getEntryText(m.genresEntry)
	// Poster is set when clicking on the image

	m.saveCallback(m.movieInfo, m.movie)

	m.window.Hide()
}

func (m *MovieWindow) onImageClick() {
	dlg, err := gtk.FileChooserDialogNewWith2Buttons(
		"Choose an image...", m.window, gtk.FILE_CHOOSER_ACTION_OPEN, "Ok", gtk.RESPONSE_OK,
		"Cancel", gtk.RESPONSE_CANCEL,
	)
	if err != nil {
		reportError(err)
		panic(err)
	}
	defer dlg.Destroy()

	response := dlg.Run()
	if response == gtk.RESPONSE_CANCEL {
		return
	}

	fileName := dlg.GetFilename()
	fileData, err := os.ReadFile(fileName)
	if err != nil {
		reportError(err)
		return
	}

	fileData = m.checkImageSize(fileData)
	if fileData == nil || len(fileData) == 0 {
		return
	}
	m.updateImage(fileData)
	m.movieInfo.image = fileData
	m.movieInfo.imageHasChanged = true
}

// updateImage updates the GtkImage
func (m *MovieWindow) updateImage(image []byte) {
	// Image size: 190x280
	pix, err := gdk.PixbufNewFromBytesOnly(image)
	if err != nil {
		reportError(err)
		panic(err)
	}
	m.posterImage.SetFromPixbuf(pix)
}

// checkImageSize makes sure that the size of the image is 190x280 and returns it
func (m *MovieWindow) checkImageSize(data []byte) []byte {
	pix, err := gdk.PixbufNewFromBytesOnly(data)
	if err != nil {
		reportError(err)
		panic(err)
	}
	width, height := pix.GetWidth(), pix.GetHeight()
	if width != imageWidth || height != imageHeight {
		return m.resizeImage(data)
	}

	return data
}

// resizeImage resizes the image to 190x280 and converts it to a PNG file
func (m *MovieWindow) resizeImage(imgData []byte) []byte {
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		reportError(err)
		return nil
	}
	imgResized := resize.Resize(imageWidth, imageHeight, img, resize.Lanczos2)
	buf := new(bytes.Buffer)
	err = png.Encode(buf, imgResized)
	if err != nil {
		reportError(err)
		return nil
	}
	return buf.Bytes()
}
