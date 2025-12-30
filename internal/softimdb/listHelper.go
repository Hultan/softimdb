package softimdb

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/data"
)

type ListHelper struct {
}

// CreateMovieCard creates a movie card (a gtk.Frame) to be placed in a gtk.FlowBox
func (l *ListHelper) CreateMovieCard(movie *data.Movie) *gtk.Frame {
	// Create a gtk.Frame to contain the movie card and provide it with a border
	frame, err := gtk.FrameNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	// Create an overlay to allow us to overlay a toWatch image.
	overlay, err := gtk.OverlayNew()
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	frame.Add(overlay)

	// CSS
	box := createMovieBox(movie)
	overlay.AddOverlay(box)

	// Add to the watch icon (if needed)
	if movie.ToWatch {
		toWatchImage := createToWatchOverlay()
		overlay.AddOverlay(toWatchImage)
	}

	// Add to needsSubtitle flag (if needed)
	if movie.NeedsSubtitle {
		needsSubtitleImage := createNeedsSubtitleOverlay()
		overlay.AddOverlay(needsSubtitleImage)
	}

	imdbRating := createIMDBRatingOverlay(movie)
	overlay.AddOverlay(imdbRating)

	myRating := createMyRatingOverlay(movie)
	overlay.AddOverlay(myRating)

	pack := createPackOverlay(movie)
	overlay.AddOverlay(pack)

	// This is to make sure that all cards have an equal height of 480 (even if they have a small image)
	// and also to make sure that they have a minimal width, which makes the gtk.FlowBox to display four movies
	// per row.
	overlay.SetSizeRequest(385, 480)

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

	// Runtime
	label := createRuntimeLabel(movie)
	box.Add(label)

	// Image
	image := createMovieImage(movie)
	box.Add(image)

	// Genres
	label = createMovieGenresLabel(movie)
	box.Add(label)

	// This does not work well when the movie is selected.
	//
	//if movie.Pack != "" {
	//	boxContext, err := box.GetStyleContext()
	//	if err != nil {
	//		reportError(err)
	//		log.Fatal(err)
	//	}
	//
	//	boxContext.AddClass("packBackground")
	//}

	return box
}

// createMovieInfoBox creates a box containing the movie title, year and subtitle
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
	label.SetSizeRequest(0, 48)
	box.PackStart(label, true, false, 5)
	box.SetName("titleBox")

	return box
}

// getMovieInfoMarkup returns the markup for the movie title, year and subtitle.
func getMovieInfoMarkup(movie *data.Movie) string {
	var s strings.Builder

	// Title & Year
	mov := &Movie{}
	mov.fromDatabase(movie)
	if calculateBitrate(mov) > bitRateWarning {
		s.WriteString(fmt.Sprintf(`<span font="Sans Regular 13" foreground="#ff7f7f"><b>%s</b></span>`, cleanString(movie.Title)))
		s.WriteString(fmt.Sprintf(`<span font="Sans Regular 13" foreground="#ff7f7f"> (%d)</span>`, movie.Year))
	} else {
		s.WriteString(fmt.Sprintf(`<span font="Sans Regular 13" foreground="#f1e3ae"><b>%s</b></span>`, cleanString(movie.Title)))
		s.WriteString(fmt.Sprintf(`<span font="Sans Regular 13" foreground="#f1e3ae"> (%d)</span>`, movie.Year))
	}

	// Subtitle
	if movie.SubTitle != "" {
		s.WriteString("\n")
		s.WriteString(fmt.Sprintf(`<span font="Sans Regular 12" foreground="#908868"><b>%s</b></span>`, cleanString(movie.SubTitle)))
	}

	return s.String()
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

// createRuntimeLabel creates a gtk.Label containing the runtime in hours and minutes
func createRuntimeLabel(movie *data.Movie) *gtk.Label {
	var s string
	var b strings.Builder

	if movie.Runtime == -1 {
		s = "Runtime : unknown"
	} else {
		t := movie.Runtime
		h, m := t/60, t%60
		s = fmt.Sprintf("Runtime : %dh %dm", h, m)
	}
	label, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	b.WriteString(`<span font="Sans Regular 10" foreground="#AAAAAA">`)
	b.WriteString(s)
	b.WriteString(`</span>`)

	label.SetMarkup(b.String())

	return label
}

// createMovieGenresLabel creates a gtk.Label containing the movie release year and all genres comma separated
func createMovieGenresLabel(movie *data.Movie) *gtk.Label {
	var s string
	var b strings.Builder

	for i := range movie.Genres {
		genre := movie.Genres[i]
		if showPrivateGenres || !genre.IsPrivate {
			if s != "" {
				s += ", "
			}
			s += genre.Name
		}
	}
	label, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	b.WriteString(`<span font="Sans Regular 10" foreground="#AAAAAA">`)
	b.WriteString(s)
	b.WriteString(`</span>`)

	label.SetMarkup(b.String())

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
	image.SetMarginStart(23)
	image.SetMarginTop(100)

	return image
}

// createNeedsSubtitleOverlay creates a gtk.Image containing the NeedsSubtitle image
func createNeedsSubtitleOverlay() *gtk.Image {
	pixBuf, err := gdk.PixbufNewFromBytesOnly(needsSubtitleIcon)
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
	image.SetMarginStart(23)
	image.SetMarginTop(160)

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
	label.SetName("imdbRatingBox")

	label.SetHAlign(gtk.ALIGN_START)
	label.SetVAlign(gtk.ALIGN_END)
	label.SetMarginStart(10)
	label.SetMarginBottom(10)

	return label
}

// getIMDBRatingMarkup returns the markup for the IMDB rating
func getIMDBRatingMarkup(movie *data.Movie) string {
	var b strings.Builder

	b.WriteString(`<span font="Sans Regular 12" foreground="#f1e3ae">IMDB : `)
	b.WriteString(fmt.Sprintf("%v", movie.ImdbRating))
	b.WriteString(`</span>`)

	return b.String()
}

// createMyRatingOverlay creates a gtk.Label containing my rating
func createMyRatingOverlay(movie *data.Movie) *gtk.Label {
	label, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	label.SetJustify(gtk.JUSTIFY_CENTER)
	label.SetMarkup(getMyRatingMarkup(movie))
	label.SetName("myRatingBox")

	label.SetHAlign(gtk.ALIGN_END)
	label.SetVAlign(gtk.ALIGN_END)
	label.SetMarginEnd(10)
	label.SetMarginBottom(10)

	return label
}

// getMyRatingMarkup returns the markup for my rating
func getMyRatingMarkup(movie *data.Movie) string {
	var b strings.Builder

	if movie.MyRating == 0 {
		b.WriteString(`<span font="Sans Regular 12" foreground="#666666">`)
		b.WriteString(fmt.Sprintf("My rating:    "))
		b.WriteString(`</span>`)
	} else {
		b.WriteString(`<span font="Sans Regular 12" foreground="#f1e3ae">`)
		b.WriteString(fmt.Sprintf("My rating: %v/5", movie.MyRating))
		b.WriteString(`</span>`)
	}

	return b.String()
}

// createPackOverlay creates a gtk.Label containing the pack name
func createPackOverlay(movie *data.Movie) *gtk.Label {
	label, err := gtk.LabelNew("")
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	if movie.Pack == "" {
		return label
	}

	label.SetJustify(gtk.JUSTIFY_CENTER)
	label.SetMarkup(getPackMarkup(movie))
	label.SetName("packLabel")

	label.SetHAlign(gtk.ALIGN_END)
	label.SetVAlign(gtk.ALIGN_START)
	label.SetMarginTop(92)

	label.SetAngle(-90)

	context, err := label.GetStyleContext()
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	context.AddClass("packLabel")

	return label
}

// getPackMarkup returns the markup for my rating
func getPackMarkup(movie *data.Movie) string {
	var b strings.Builder

	b.WriteString(`<span font="Sans Regular 12" foreground="#141103">`)
	b.WriteString(fmt.Sprintf("Pack: %s", movie.Pack))
	b.WriteString(`</span>`)

	return b.String()
}

func calculateBitrate(movie *Movie) int {
	if movie.runtime <= 0 {
		return 0 // avoid division by zero
	}

	// Convert everything to float64 to keep fractional precision.
	sizeBits := float64(movie.size) * 8        // bits
	durationSec := float64(movie.runtime) * 60 // seconds
	bitrateBps := sizeBits / durationSec       // bits per second
	bitrateKbps := bitrateBps / 1000           // kilobits per second

	return int(math.Round(bitrateKbps))
}
