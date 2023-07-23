package softimdb

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/data"
)

type ListHelper struct {
}

func ListHelperNew() *ListHelper {
	return new(ListHelper)
}

// CreateMovieCard creates a movie card (a gtk.Frame) to be placed in a gtk.FlowBox
func (l *ListHelper) CreateMovieCard(movie *data.Movie) *gtk.Frame {
	// TODO : Frame could be replaced with CSS border to the box?
	// Create a frame (for the border)
	frame, err := gtk.FrameNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	// Create an overlay, to allow us to overlay a toWatch image.
	overlay, err := gtk.OverlayNew()
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	frame.Add(overlay)

	box := createMovieBox(movie)
	overlay.AddOverlay(box)

	// Add to watch flag (if needed)
	if movie.ToWatch {
		toWatchImage := createToWatchOverlay()
		overlay.AddOverlay(toWatchImage)
	}

	imdbRating := createIMDBRatingOverlay(movie)
	overlay.AddOverlay(imdbRating)

	myRating := createMyRatingOverlay(movie)
	overlay.AddOverlay(myRating)

	// This is to make sure that all cards have an equal height of 430 (even if they have a small image)
	// and also to make sure that they have a minimal width, which makes the gtk.FlowBox to display four movies
	// per row.
	frame.SetSizeRequest(385, 430)

	return frame
}

// createMovieBox creates a gtk.Box that contains all the information about a single movie
func createMovieBox(movie *data.Movie) *gtk.Box {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	info := createMovieInfoBox(movie)
	box.PackStart(info, false, false, 5)

	// Image
	image := createMovieImage(movie)
	box.Add(image)

	// Genres
	label := createMovieGenresLabel(movie)
	box.Add(label)

	return box
}

// createMovieInfoBox creates a box containing movie title, subtitle and year
func createMovieInfoBox(movie *data.Movie) *gtk.Box {
	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	label, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	label.SetJustify(gtk.JUSTIFY_CENTER)
	label.SetMarkup(getMovieInfoMarkup(movie))
	box.PackStart(label, true, false, 5)

	return box
}

// getMovieInfoMarkup returns the markup for the title, subtitle for the movie
func getMovieInfoMarkup(movie *data.Movie) string {
	s := ""

	// Title
	s += `<span font="Sans Regular 16" foreground="#111111"><b>`
	s += cleanString(movie.Title)
	s += `</b></span>`
	s += "\n"

	// Subtitle
	s += `<span font="Sans Regular 12" foreground="#111111"><b>`
	s += cleanString(movie.SubTitle)
	s += `</b></span>`

	return s
}

// createMovieImage creates a gtk.Image for the movie
func createMovieImage(movie *data.Movie) *gtk.Image {
	if movie.Image == nil {
		image, err := gtk.ImageNew()
		if err != nil {
			reportError(err)
			log.Fatal(err)
		}
		return image
	}

	pixBuf, err := gdk.PixbufNewFromBytesOnly(movie.Image)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	image, err := gtk.ImageNewFromPixbuf(pixBuf)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	return image
}

// createMovieGenresLabel creates a gtk.Label containing the movie release year and all genres comma separated
func createMovieGenresLabel(movie *data.Movie) *gtk.Label {
	var s string

	for i := range movie.Tags {
		tag := movie.Tags[i]
		if s != "" {
			s += ", "
		}
		s += tag.Name
	}
	label, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	// Year
	year := `<span font="Sans Regular 10" foreground="#AAAAAA">`
	year += fmt.Sprintf("%v", movie.Year)
	year += `</span> - `

	s = `<span font="Sans Regular 10" foreground="#AAAAAA">` + s + `</span>`
	label.SetMarkup(year + s)

	return label
}

// createToWatchOverlay creates a gtk.Image containing the to watch image
func createToWatchOverlay() *gtk.Image {
	pixBuf, err := gdk.PixbufNewFromBytesOnly(toWatchIcon)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	image, err := gtk.ImageNewFromPixbuf(pixBuf)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	image.SetVAlign(gtk.ALIGN_START)
	image.SetHAlign(gtk.ALIGN_START)
	image.SetMarginStart(10)
	image.SetMarginTop(10)

	return image
}

// createIMDBRatingOverlay creates a gtk.Label containing IMDB rating
func createIMDBRatingOverlay(movie *data.Movie) *gtk.Label {
	label, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	label.SetJustify(gtk.JUSTIFY_CENTER)
	label.SetMarkup(getIMDBRatingMarkup(movie))

	label.SetHAlign(gtk.ALIGN_START)
	label.SetVAlign(gtk.ALIGN_END)
	label.SetMarginStart(10)
	label.SetMarginBottom(10)

	return label
}

// getIMDBRatingMarkup returns the markup for the IMDB rating
func getIMDBRatingMarkup(movie *data.Movie) string {
	s := `<span font="Sans Regular 12" foreground="#AAAA00">`
	s += fmt.Sprintf("Imdb rating : %v", movie.ImdbRating)
	s += `</span>   `

	return s
}

// createMyRatingOverlay creates a gtk.Label containing my rating
func createMyRatingOverlay(movie *data.Movie) *gtk.Label {
	label, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	if movie.MyRating == 0 {
		return label
	}

	label.SetJustify(gtk.JUSTIFY_CENTER)
	label.SetMarkup(getMyRatingMarkup(movie))

	label.SetHAlign(gtk.ALIGN_END)
	label.SetVAlign(gtk.ALIGN_END)
	label.SetMarginEnd(10)
	label.SetMarginBottom(10)

	return label
}

// getMyRatingMarkup returns the markup for my rating
func getMyRatingMarkup(movie *data.Movie) string {
	s := `<span font="Sans Regular 12" foreground="#AAAA00">`
	s += fmt.Sprintf("My rating: %v/5", movie.MyRating)
	s += `</span>`

	return s
}
