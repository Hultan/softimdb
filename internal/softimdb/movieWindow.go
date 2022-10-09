package softimdb

import (
	"fmt"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/data"
	"github.com/hultan/softimdb/internal/imdb2"
	"github.com/hultan/softteam/framework"
)

// 1. Create a struct called MovieWindowManager
// 2. Make that handle the opening and saving the IMDB information/Movie

type MovieWindow struct {
	window           *gtk.Window
	titleEntry       *gtk.Entry
	yearEntry        *gtk.Entry
	storyLineEntry   *gtk.TextView
	ratingEntry      *gtk.Entry
	genresEntry      *gtk.Entry
	posterImage      *gtk.Image
	PosterHasChanged bool
	poster           []byte

	movieImdb    *imdb2.Movie
	movie        *data.Movie
	saveCallback func(*MovieWindow)
}

func NewMovieWindow(info *imdb2.Movie, saveCallback func(window *MovieWindow)) *MovieWindow {
	m := new(MovieWindow)
	m.saveCallback = saveCallback
	m.movieImdb = info
	return m
}

func NewMovieWindowFromMovie(movie *data.Movie, saveCallback func(window *MovieWindow)) *MovieWindow {
	m := new(MovieWindow)
	m.saveCallback = saveCallback
	m.movie = movie
	m.movieImdb = &imdb2.Movie{
		Id:           movie.ImdbID,
		Title:        movie.Title,
		Type:         "movie",
		Year:         fmt.Sprintf("%d", movie.Year),
		ImageURL:     "",
		StoryLine:    movie.StoryLine,
		Genres:       tagsToString(movie.Tags),
		Rating:       fmt.Sprintf("%.2f", movie.ImdbRating),
		ErrorMessage: "",
	}
	m.poster = *movie.Image
	return m
}

func tagsToString(tags []data.Tag) string {
	result := ""
	for _, tag := range tags {
		if result != "" {
			result += ","
		}
		result += tag.Name
	}
	return result
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
		m.titleEntry.SetText(m.movieImdb.Title)
		m.yearEntry.SetText(m.movieImdb.Year)
		buffer, err := gtk.TextBufferNew(nil)
		if err != nil {
			// TODO : Fix error handling
			panic(err)
		}
		buffer.SetText(m.movieImdb.StoryLine)
		m.storyLineEntry.SetBuffer(buffer)
		m.ratingEntry.SetText(m.movieImdb.Rating)
		m.genresEntry.SetText(m.movieImdb.Genres)
		if m.movieImdb.ImageURL == "" {
			// We are opening a movie from the database
			pix, err := gdk.PixbufNewFromBytesOnly(m.poster)
			if err != nil {
				// TODO : Fix error handling
				panic(err)
			}
			m.posterImage.SetFromPixbuf(pix)
		} else {
			// We are opening a movie from IMDB
			poster, err := m.movieImdb.GetPoster()
			if err != nil {
				// TODO : Fix error handling
				panic(err)
			}
			pix, err := gdk.PixbufNewFromBytesOnly(poster)
			if err != nil {
				// TODO : Fix error handling
				panic(err)
			}
			m.posterImage.SetFromPixbuf(pix)
		}
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
	m.movieImdb.Title = m.getEntryText(m.titleEntry)
	m.movieImdb.Year = m.getEntryText(m.yearEntry)
	m.movieImdb.Rating = m.getEntryText(m.ratingEntry)
	buffer, err := m.storyLineEntry.GetBuffer()
	if err != nil {
		panic(err)
	}
	storyLine, err := buffer.GetText(buffer.GetStartIter(), buffer.GetEndIter(), false)
	if err != nil {
		panic(err)
	}
	m.movieImdb.StoryLine = storyLine
	m.movieImdb.Genres = m.getEntryText(m.genresEntry)
	// Poster is set when clicking on the image

	m.saveCallback(m)

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

//
// func (m *MovieWindow) getGenres(text string) []string {
// 	var result []string
// 	genres := strings.Split(text, ",")
// 	for _, genre := range genres {
// 		result = append(result, genre)
// 	}
// 	return result
// }

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

	m.poster = file
	m.PosterHasChanged = true
}
