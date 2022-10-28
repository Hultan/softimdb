package softimdb

import (
	"fmt"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/data"
	"github.com/hultan/softteam/framework"
)

type MovieWindow struct {
	window         *gtk.Window
	titleEntry     *gtk.Entry
	subTitleEntry  *gtk.Entry
	yearEntry      *gtk.Entry
	storyLineEntry *gtk.TextView
	ratingEntry    *gtk.Entry
	genresEntry    *gtk.Entry
	posterImage    *gtk.Image

	movieInfo *MovieInfo
	movie     *data.Movie

	saveCallback func(*MovieInfo, *data.Movie)
}

func NewMovieWindow(info *MovieInfo, movie *data.Movie, saveCallback func(*MovieInfo, *data.Movie)) *MovieWindow {
	m := new(MovieWindow)
	m.movieInfo = info
	m.movie = movie
	m.saveCallback = saveCallback
	return m
}

func (m *MovieWindow) OpenForm(builder *framework.GtkBuilder, parent gtk.IWindow) {
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
		m.titleEntry = builder.GetObject("titleEntry").(*gtk.Entry)
		m.subTitleEntry = builder.GetObject("subTitleEntry").(*gtk.Entry)
		m.yearEntry = builder.GetObject("yearEntry").(*gtk.Entry)
		m.storyLineEntry = builder.GetObject("storyLineTextView").(*gtk.TextView)
		m.ratingEntry = builder.GetObject("ratingEntry").(*gtk.Entry)
		m.genresEntry = builder.GetObject("genresEntry").(*gtk.Entry)
		m.posterImage = builder.GetObject("posterImage").(*gtk.Image)
		eventBox := builder.GetObject("imageEventBox").(*gtk.EventBox)
		eventBox.Connect("button-press-event", m.onImageClick)

		// Fill form with data
		m.titleEntry.SetText(m.movieInfo.title)
		m.subTitleEntry.SetText(m.movieInfo.subTitle)
		m.yearEntry.SetText(fmt.Sprintf("%d", m.movieInfo.getYear()))
		buffer, err := gtk.TextBufferNew(nil)
		if err != nil {
			// TODO : Fix error handling
			panic(err)
		}
		buffer.SetText(m.movieInfo.storyLine)
		m.storyLineEntry.SetBuffer(buffer)
		m.ratingEntry.SetText(m.movieInfo.imdbRating)
		m.genresEntry.SetText(m.movieInfo.tags)
		m.genresEntry.SetEditable(false)

		// Poster
		m.updateImage(m.movieInfo.image)

		// Store reference to database and window
		m.window = movieWindow
	}

	// Show the window
	m.window.ShowAll()

	m.titleEntry.GrabFocus()
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
	m.movieInfo.title = m.getEntryText(m.titleEntry)
	m.movieInfo.subTitle = m.getEntryText(m.subTitleEntry)
	m.movieInfo.year = m.getEntryText(m.yearEntry)
	m.movieInfo.imdbRating = m.getEntryText(m.ratingEntry)
	buffer, err := m.storyLineEntry.GetBuffer()
	if err != nil {
		panic(err)
	}
	storyLine, err := buffer.GetText(buffer.GetStartIter(), buffer.GetEndIter(), false)
	if err != nil {
		panic(err)
	}
	m.movieInfo.storyLine = storyLine

	// TODO : Fix editing of tags/genres
	// m.movie.Tags = m.getEntryText(m.genresEntry)

	// Poster is set when clicking on the image

	m.saveCallback(m.movieInfo, m.movie)
	m.window.Hide()
}

func (m *MovieWindow) onImageClick() {
	dialog, err := gtk.FileChooserDialogNewWith2Buttons(
		"Choose an image...", m.window, gtk.FILE_CHOOSER_ACTION_OPEN, "Ok", gtk.RESPONSE_OK,
		"Cancel", gtk.RESPONSE_CANCEL,
	)
	if err != nil {
		panic(err)
	}
	defer dialog.Destroy()

	response := dialog.Run()
	if response == gtk.RESPONSE_CANCEL {
		return
	}

	fileName := dialog.GetFilename()
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Could not read the file due to this %s error \n", err)
		return
	}

	m.updateImage(bytes)
	m.movieInfo.image = bytes
	m.movieInfo.imageHasChanged = true
}

func (m *MovieWindow) updateImage(image []byte) {
	pix, err := gdk.PixbufNewFromBytesOnly(image)
	if err != nil {
		// TODO : Fix error handling
		panic(err)
	}
	m.posterImage.SetFromPixbuf(pix)
}
