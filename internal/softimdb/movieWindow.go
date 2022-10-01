package softimdb

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/imdb"
	"github.com/hultan/softteam/framework"
)

type MovieWindow struct {
	window         *gtk.Window
	titleEntry     *gtk.Entry
	yearEntry      *gtk.Entry
	storyLineEntry *gtk.TextView
	ratingEntry    *gtk.Entry
	genresEntry    *gtk.Entry
	posterImage    *gtk.Image

	info         *imdb.MovieInfo
	saveCallback func()
}

func MovieWindowNew(info *imdb.MovieInfo, saveCallback func()) *MovieWindow {
	m := new(MovieWindow)
	m.saveCallback = saveCallback
	m.info = info
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
		m.yearEntry = builder.GetObject("yearEntry").(*gtk.Entry)
		m.storyLineEntry = builder.GetObject("storyLineTextView").(*gtk.TextView)
		m.ratingEntry = builder.GetObject("ratingEntry").(*gtk.Entry)
		m.genresEntry = builder.GetObject("genresEntry").(*gtk.Entry)
		m.posterImage = builder.GetObject("posterImage").(*gtk.Image)
		eventBox := builder.GetObject("imageEventBox").(*gtk.EventBox)
		eventBox.Connect("button-press-event", m.onImageClick)

		// Fill form with data
		m.titleEntry.SetText(m.info.Title)
		m.yearEntry.SetText(strconv.FormatInt(int64(m.info.Year), 10))
		buffer, err := gtk.TextBufferNew(nil)
		if err != nil {
			panic(err)
		}
		buffer.SetText(m.info.StoryLine)
		m.storyLineEntry.SetBuffer(buffer)
		m.ratingEntry.SetText(strconv.FormatFloat(m.info.Rating, 'f', 1, 64))
		m.genresEntry.SetText(m.getGenresText(m.info.Tags))
		pix, err := gdk.PixbufNewFromBytesOnly(m.info.Poster)
		if err != nil {
			panic(err)
		}
		m.posterImage.SetFromPixbuf(pix)
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
	// Title
	m.info.Title = m.getEntryText(m.titleEntry)

	// Year
	yearString := m.getEntryText(m.yearEntry)
	year, err := strconv.Atoi(yearString)
	if err != nil {
		message := "Invalid year"
		dialog := gtk.MessageDialogNew(m.window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, message)
		dialog.Run()
		dialog.Destroy()
		return
	}
	m.info.Year = year

	// Rating
	ratingString := m.getEntryText(m.ratingEntry)
	rating, err := strconv.ParseFloat(ratingString, 64)
	if err != nil || rating < 1 || rating > 10 {
		message := "Invalid rating"
		dialog := gtk.MessageDialogNew(m.window, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, message)
		dialog.Run()
		dialog.Destroy()
		return
	}
	m.info.Rating = rating

	// Story line
	buffer, err := m.storyLineEntry.GetBuffer()
	if err != nil {
		panic(err)
	}
	storyLine, err := buffer.GetText(buffer.GetStartIter(), buffer.GetEndIter(), false)
	if err != nil {
		panic(err)
	}
	m.info.StoryLine = storyLine

	// Genres
	genres := m.getGenres(m.getEntryText(m.genresEntry))
	m.info.Tags = genres

	// Poster is set when clicking on the image

	m.saveCallback()

	m.window.Hide()
}

func (m *MovieWindow) getGenresText(tags []string) string {
	result := ""
	for _, tag := range tags {
		if result != "" {
			result += ","
		}
		result += tag
	}
	return result
}

func (m *MovieWindow) getGenres(text string) []string {
	var result []string
	genres := strings.Split(text, ",")
	for _, genre := range genres {
		result = append(result, genre)
	}
	return result
}

func (m *MovieWindow) onImageClick() {
	dialog, err := gtk.FileChooserDialogNewWith2Buttons("Choose an image...", m.window, gtk.FILE_CHOOSER_ACTION_OPEN, "Ok", gtk.RESPONSE_OK,
		"Cancel", gtk.RESPONSE_CANCEL)
	if err != nil {
		panic(err)
	}
	defer dialog.Destroy()

	response := dialog.Run()
	if response == gtk.RESPONSE_CANCEL {
		return
	}

	fileName := dialog.GetFilename()
	file, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Could not read the file due to this %s error \n", err)
	}

	m.info.Poster = file
	m.info.PosterHasChanged = true
}
