package softimdb

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/data"
)

type ListHelper struct {
}

func ListHelperNew() *ListHelper {
	return new(ListHelper)
}

func (l *ListHelper) CreateMovieCard(movie *data.Movie) *gtk.Frame {
	// Create a frame (for the border)
	frame, err := gtk.FrameNew("")
	if err != nil {
		reportError(err)
		panic(err)
	}

	// Create the card (box)
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	if err != nil {
		reportError(err)
		panic(err)
	}
	box.SetBorderWidth(10)
	frame.Add(box)
	// Title, Year and IMDB rating
	nameBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	if err != nil {
		reportError(err)
		panic(err)
	}
	titleLabel, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		panic(err)
	}
	titleLabel.SetJustify(gtk.JUSTIFY_CENTER)
	titleLabel.SetMarkup(getMarkup(movie))
	nameBox.PackStart(titleLabel, true, false, 5)
	box.PackStart(nameBox, false, false, 5)

	// Image
	if movie.Image == nil {
		image, err := gtk.ImageNew()
		if err != nil {
			reportError(err)
			panic(err)
		}
		box.Add(image)
	} else {
		pixBuf, err := gdk.PixbufNewFromBytesOnly(movie.Image)
		if err != nil {
			reportError(err)
			panic(err)
		}
		image, err := gtk.ImageNewFromPixbuf(pixBuf)
		if err != nil {
			reportError(err)
			panic(err)
		}
		box.Add(image)
	}

	// Genres
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
		panic(err)
	}
	str = `<span font="Sans Regular 10" foreground="#DDDDDD">` + str + `</span>`
	label.SetMarkup(str)
	box.Add(label)

	return frame
}

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
