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

func (l *ListHelper) GetMovieCard(movie *data.Movie) *gtk.Frame {
	// Create a frame (for the border)
	frame, err := gtk.FrameNew("")
	if err != nil {
		panic(err)
	}
	// Create the card (box)
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	if err != nil {
		return nil
	}
	box.SetBorderWidth(10)
	frame.Add(box)
	// Title, Year and IMDB rating
	nameBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 5)
	if err != nil {
		panic(err)
	}
	titleLabel, err := gtk.LabelNew("")
	if err != nil {
		panic(err)
	}
	titleLabel.SetMarkup(`<span font="Sans Regular 14" foreground="#d49c6b"><b>` + cleanString(movie.Title) + `</b></span> 
<span font="Sans Regular 10" foreground="#d49c6b"><b>` + cleanString(movie.SubTitle) + `</b></span>
<span font="Sans Regular 10" foreground="#967c61">` + fmt.Sprintf("%v", movie.Year) + ` - ` + fmt.Sprintf("Imdb : %v", movie.ImdbRating) + `</span>`	)
	nameBox.PackStart(titleLabel, true, false, 5)
	box.PackStart(nameBox, false, false, 5)

	// Image
	pixBuf, err := gdk.PixbufNewFromBytesOnly(*movie.Image)
	if err != nil {
		panic(err)
	}
	image, err := gtk.ImageNewFromPixbuf(pixBuf)
	if err != nil {
		panic(err)
	}
	box.Add(image)

	// Genres
	var str string
	for i := range movie.Tags {
		tag := movie.Tags[i]
		if str!=""{
			str += ", "
		}
		str += tag.Name
	}
	label, err := gtk.LabelNew("")
	if err != nil {
		panic(err)
	}
	str = `<span font="Sans Regular 10" foreground="#d49c6b">` + str + `</span>`
	label.SetMarkup(str)
	box.Add(label)

	return frame
}
