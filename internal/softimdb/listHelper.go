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

	// This is to make sure that all cards have an equal height of 430 (even if they have a small image)
	// and also to make sure that they have a minimal width, which makes the flowbox to display four movies
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
	label := createGenresLabel(movie)
	box.Add(label)

	return box
}

// createMovieInfoBox creates a box containing movie title, year and IMDB rating
func createMovieInfoBox(movie *data.Movie) *gtk.Box {
	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	titleLabel, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	titleLabel.SetJustify(gtk.JUSTIFY_CENTER)
	titleLabel.SetMarkup(getMarkup(movie))
	box.PackStart(titleLabel, true, false, 5)

	return box
}

// createGenresLabel creates a gtk.Label containing all genres comma separated
func createGenresLabel(movie *data.Movie) *gtk.Label {
	var str string
	for i := range movie.Tags {
		tag := movie.Tags[i]
		if str != "" {
			str += ", "
		}
		str += tag.Name
	}
	label, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	str = `<span font="Sans Regular 10" foreground="#DDDDDD">` + str + `</span>`
	label.SetMarkup(str)

	return label
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

// getMarkup returns the markup for the title, subtitle, year & ratings for the movie
func getMarkup(movie *data.Movie) string {
	s := ""

	// Title
	s += `<span font="Sans Regular 14" foreground="#111111"><b>`
	s += cleanString(movie.Title)
	s += `</b></span>`
	s += "\n"

	// Subtitle
	s += `<span font="Sans Regular 10" foreground="#111111"><b>`
	s += cleanString(movie.SubTitle)
	s += `</b></span>`
	s += "\n"

	// Year & ratings
	s += `<span font="Sans Regular 10" foreground="#DDDDDD">`
	s += fmt.Sprintf("%v", movie.Year)
	s += `</span> - <span font="Sans Regular 10" foreground="#AAAA00">`
	s += fmt.Sprintf("Imdb rating : %v", movie.ImdbRating)
	s += `</span>`
	if movie.MyRating > 0 {
		s += ` - <span font="Sans Regular 10" foreground="#DDDDDD">`
		s += fmt.Sprintf("My rating: %v/5", movie.MyRating)
		s += `</span>`
	}

	return s
}
