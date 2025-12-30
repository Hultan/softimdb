package softimdb

import (
	"fmt"
	_ "image/jpeg"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hultan/dialog"
	"github.com/hultan/softimdb/internal/config"
	"github.com/texttheater/golang-levenshtein/levenshtein"

	"github.com/hultan/softimdb/internal/builder"
	"github.com/hultan/softimdb/internal/data"
	"github.com/hultan/softimdb/internal/imdb"
)

type movieWindow struct {
	window                   *gtk.Window
	pathEntry                *gtk.Entry
	imdbUrlEntry             *gtk.Entry
	titleEntry               *gtk.Entry
	subTitleEntry            *gtk.Entry
	yearEntry                *gtk.Entry
	myRatingEntry            *gtk.Entry
	toWatchCheckButton       *gtk.CheckButton
	needsSubtitleCheckButton *gtk.CheckButton
	storyLineEntry           *gtk.TextView
	ratingEntry              *gtk.Entry
	genresEntry              *gtk.Entry
	packEntry                *gtk.Entry
	posterImage              *gtk.Image
	runtimeEntry             *gtk.Entry
	deleteButton             *gtk.Button
	castAndCrewList          *gtk.ListBox
	movieStack               *gtk.Stack
	bitrateLabel             *gtk.Label

	guiMovie  *Movie
	dataMovie *data.Movie

	config *config.Config
	db     *data.Database

	closeCallback func(gtk.ResponseType, *Movie, *data.Movie)
}

type similarMovie struct {
	distance int
	title    string
}

var scrapeImdbOnce bool
var showSimilarOnce bool

const bitRateWarning = 8000

func newMovieWindow(builder *builder.Builder, parent gtk.IWindow, db *data.Database,
	config *config.Config) *movieWindow {
	m := &movieWindow{}

	m.db = db
	m.config = config
	m.window = builder.GetObject("movieWindow").(*gtk.Window)
	m.window.SetTitle("Movie info window")
	m.window.SetTransientFor(parent)
	m.window.SetModal(true)
	m.window.SetKeepAbove(true)
	m.window.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)
	m.window.HideOnDelete()

	button := builder.GetObject("okButton").(*gtk.Button)
	_ = button.Connect("clicked", func() {
		if !m.saveMovie() {
			return
		}
		m.window.Hide()
		m.closeCallback(gtk.RESPONSE_ACCEPT, m.guiMovie, m.dataMovie)
	})

	button = builder.GetObject("cancelButton").(*gtk.Button)
	_ = button.Connect("clicked", func() {
		m.window.Hide()
		m.closeCallback(gtk.RESPONSE_CANCEL, nil, nil)
	})

	button = builder.GetObject("deleteButton").(*gtk.Button)
	_ = button.Connect("clicked", func() {
		if m.deleteMovie() {
			m.window.Hide()
			m.closeCallback(gtk.RESPONSE_REJECT, nil, m.dataMovie)
		}
	})
	m.deleteButton = button

	m.imdbUrlEntry = builder.GetObject("imdbUrlEntry").(*gtk.Entry)
	m.pathEntry = builder.GetObject("pathEntry").(*gtk.Entry)
	m.titleEntry = builder.GetObject("titleEntry").(*gtk.Entry)
	m.subTitleEntry = builder.GetObject("subTitleEntry").(*gtk.Entry)
	m.yearEntry = builder.GetObject("yearEntry").(*gtk.Entry)
	m.myRatingEntry = builder.GetObject("myRatingEntry").(*gtk.Entry)
	m.toWatchCheckButton = builder.GetObject("toWatchCheckButton").(*gtk.CheckButton)
	m.needsSubtitleCheckButton = builder.GetObject("needsSubtitleCheckButton").(*gtk.CheckButton)
	m.storyLineEntry = builder.GetObject("storyLineTextView").(*gtk.TextView)
	m.ratingEntry = builder.GetObject("ratingEntry").(*gtk.Entry)
	m.genresEntry = builder.GetObject("genresEntry").(*gtk.Entry)
	m.packEntry = builder.GetObject("packEntry").(*gtk.Entry)
	m.posterImage = builder.GetObject("posterImage").(*gtk.Image)
	m.runtimeEntry = builder.GetObject("runtimeEntry").(*gtk.Entry)
	m.castAndCrewList = builder.GetObject("castAndCrewList").(*gtk.ListBox)
	m.movieStack = builder.GetObject("movieStack").(*gtk.Stack)
	m.bitrateLabel = builder.GetObject("bitrateLabel").(*gtk.Label)

	eventBox := builder.GetObject("imageEventBox").(*gtk.EventBox)
	eventBox.Connect("button-press-event", m.onImageClick)
	m.titleEntry.Connect("focus-out-event", m.onTitleEntryFocusOut)
	m.imdbUrlEntry.Connect("focus-out-event", m.onIMDBEntryFocusOut)

	return m
}

func (m *movieWindow) open(guiMovie *Movie, dataMovie *data.Movie, closeCallback func(gtk.ResponseType, *Movie,
	*data.Movie)) {

	if guiMovie == nil {
		guiMovie = &Movie{}
		// TODO : Disable cast and crew tab here
	}

	scrapeImdbOnce = false
	showSimilarOnce = false
	if guiMovie.title == "" {
		//  New movie
		guiMovie.toWatch = true
		guiMovie.needsSubtitle = !m.hasSubtitles(guiMovie.moviePath)
	} else {
		// Edit movie
		scrapeImdbOnce = true
		showSimilarOnce = true
	}

	// We show the window here as well, since we are going to the
	// database to load the person below, and that causes a short delay
	m.window.ShowAll()

	if dataMovie != nil {
		// Load persons for the movie (they are no longer loaded in the main load)
		persons, err := m.db.GetPersonsForMovie(dataMovie)
		if err != nil {
			return
		}
		dataMovie.Persons = persons
	}

	m.dataMovie = dataMovie
	m.guiMovie = guiMovie
	m.closeCallback = closeCallback
	m.deleteButton.SetSensitive(dataMovie != nil)

	m.fillForm()

	m.window.ShowAll()
	m.imdbUrlEntry.GrabFocus()
}

func (m *movieWindow) fillForm() {
	// Fill form with data
	m.imdbUrlEntry.SetText(m.guiMovie.imdbUrl)
	m.pathEntry.SetText(m.guiMovie.moviePath)
	m.titleEntry.SetText(m.guiMovie.title)
	m.subTitleEntry.SetText(m.guiMovie.subTitle)
	m.yearEntry.SetText(fmt.Sprintf("%d", m.guiMovie.getYear()))
	m.myRatingEntry.SetText(fmt.Sprintf("%d", m.guiMovie.myRating))
	m.toWatchCheckButton.SetActive(m.guiMovie.toWatch)
	m.needsSubtitleCheckButton.SetActive(m.guiMovie.needsSubtitle)
	buffer, err := gtk.TextBufferNew(nil)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	buffer.SetText(m.guiMovie.storyLine)
	m.storyLineEntry.SetBuffer(buffer)
	m.ratingEntry.SetText(m.guiMovie.imdbRating)
	m.genresEntry.SetText(m.guiMovie.genres)
	m.packEntry.SetText(m.guiMovie.pack)
	m.runtimeEntry.SetText(strconv.Itoa(m.guiMovie.runtime))

	// New movies will have size = 0, so fix that
	if m.guiMovie.size == 0 {
		m.guiMovie.size = m.getMovieSize(m.guiMovie)
	}
	m.bitrateLabel.SetMarkup(calculateBitrateString(m.guiMovie))

	if m.guiMovie.image == nil {
		m.posterImage.Clear()
	} else {
		m.updateImage(m.guiMovie.image)
	}

	if m.dataMovie != nil {
		m.fillCastAndCrewPage()
	}
	m.movieStack.SetVisibleChildName("MoviePage")
}

func (m *movieWindow) getMovieSize(movie *Movie) int {
	dir := path.Join(m.config.RootDir, movie.moviePath)
	file, err := findMovieFile(dir)
	if err != nil {
		return 0
	}
	info, err := os.Stat(path.Join(dir, file))
	if err != nil {
		return 0
	}

	return int(info.Size())
}

func calculateBitrateString(movie *Movie) string {
	p := message.NewPrinter(language.Swedish)
	size := calculateBitrate(movie)
	var format string
	if size > bitRateWarning {
		format = "Bitrate (est): <span foreground=\"red\">%d kbps</span>"
	} else {
		format = "Bitrate (est): %d kbps"
	}
	return p.Sprintf(format, size)
}

func (m *movieWindow) saveMovie() bool {
	if !m.fillAndValidateBasicFields() {
		return false
	}
	if !m.fillAndValidateRatings() {
		return false
	}
	if !m.fillStorylineAndGenres() {
		return false
	}
	return true
}

func (m *movieWindow) fillAndValidateBasicFields() bool {
	m.guiMovie.moviePath = getEntryText(m.pathEntry)
	m.guiMovie.imdbUrl = getEntryText(m.imdbUrlEntry)

	id, err := getIdFromUrl(m.guiMovie.imdbUrl)
	if err != nil {
		m.showValidationError("Invalid IMDB url", fmt.Sprintf("Failed to retrieve IMDB id from url: %s", err))
		return false
	}
	m.guiMovie.imdbId = id

	m.guiMovie.title = getEntryText(m.titleEntry)
	m.guiMovie.subTitle = getEntryText(m.subTitleEntry)
	m.guiMovie.pack = getEntryText(m.packEntry)
	m.guiMovie.year = getEntryText(m.yearEntry)

	m.guiMovie.toWatch = m.toWatchCheckButton.GetActive()
	m.guiMovie.needsSubtitle = m.needsSubtitleCheckButton.GetActive()
	m.guiMovie.imdbRating = getEntryText(m.ratingEntry)
	return true
}

func (m *movieWindow) fillAndValidateRatings() bool {
	ratingText := getEntryText(m.myRatingEntry)
	rating, err := strconv.Atoi(ratingText)
	if err != nil || rating < 0 || rating > 5 {
		m.showValidationError("Invalid my rating", fmt.Sprintf("Invalid rating: %s (error: %s)", ratingText, err))
		return false
	}
	m.guiMovie.myRating = rating

	runtimeText := getEntryText(m.runtimeEntry)
	runtime, err := strconv.Atoi(runtimeText)
	if err != nil || (runtime != -1 && runtime <= 0) {
		m.showValidationError("Invalid runtime", fmt.Sprintf("Invalid runtime: %s (error: %s)", runtimeText, err))
		return false
	}
	m.guiMovie.runtime = runtime

	return true
}

func (m *movieWindow) fillStorylineAndGenres() bool {
	buffer, err := m.storyLineEntry.GetBuffer()
	if err != nil {
		reportError(err)
		return false
	}
	storyLine, err := buffer.GetText(buffer.GetStartIter(), buffer.GetEndIter(), false)
	if err != nil {
		reportError(err)
		return false
	}
	m.guiMovie.storyLine = storyLine
	m.guiMovie.genres = getEntryText(m.genresEntry)
	return true
}

func (m *movieWindow) showValidationError(title, message string) {
	_, err := dialog.Title(title).Text(message).ErrorIcon().OkButton().Show()
	if err != nil {
		fmt.Printf("Dialog error: %s\n", err)
	}
}

//func (m *movieWindow) saveMovie() bool {
//	// Fill fields
//	m.guiMovie.moviePath = getEntryText(m.pathEntry)
//	m.guiMovie.imdbUrl = getEntryText(m.imdbUrlEntry)
//	id, err := getIdFromUrl(m.guiMovie.imdbUrl)
//	if err != nil {
//		msg := fmt.Sprintf("Failed to retrieve IMDB id from url : %s", err)
//		_, err = dialog.Title("Invalid IMDB url...").Text(msg).ErrorIcon().OkButton().Show()
//
//		if err != nil {
//			fmt.Printf("Error : %s", err)
//		}
//
//		return false
//	}
//	m.guiMovie.imdbId = id
//	m.guiMovie.title = getEntryText(m.titleEntry)
//	m.guiMovie.subTitle = getEntryText(m.subTitleEntry)
//	m.guiMovie.pack = getEntryText(m.packEntry)
//	m.guiMovie.year = getEntryText(m.yearEntry)
//
//	ratingText := getEntryText(m.myRatingEntry)
//	rating, err := strconv.Atoi(ratingText)
//	legalRating := rating >= 0 && rating <= 5
//	if err != nil || !legalRating {
//		msg := fmt.Sprintf("Invalid my rating : %s (error : %s)", ratingText, err)
//		_, err = dialog.Title("Invalid my rating...").Text(msg).ErrorIcon().OkButton().Show()
//
//		if err != nil {
//			fmt.Printf("Error : %s", err)
//		}
//
//		return false
//	}
//	m.guiMovie.myRating = rating
//
//	runtimeText := getEntryText(m.runtimeEntry)
//	runtime, err := strconv.Atoi(runtimeText)
//	legalRuntime := runtime == -1 || runtime > 0
//	if err != nil || !legalRuntime {
//		msg := fmt.Sprintf("Invalid runtime : %s (error : %s)", runtimeText, err)
//		_, err = dialog.Title("Invalid runtime...").Text(msg).ErrorIcon().OkButton().Show()
//
//		if err != nil {
//			fmt.Printf("Error : %s", err)
//		}
//
//		return false
//	}
//	m.guiMovie.runtime = runtime
//
//	m.guiMovie.toWatch = m.toWatchCheckButton.GetActive()
//	m.guiMovie.needsSubtitle = m.needsSubtitleCheckButton.GetActive()
//	m.guiMovie.imdbRating = getEntryText(m.ratingEntry)
//	buffer, err := m.storyLineEntry.GetBuffer()
//	if err != nil {
//		reportError(err)
//		return false
//	}
//	storyLine, err := buffer.GetText(buffer.GetStartIter(), buffer.GetEndIter(), false)
//	if err != nil {
//		reportError(err)
//		return false
//	}
//	m.guiMovie.storyLine = storyLine
//	m.guiMovie.genres = getEntryText(m.genresEntry)
//
//	return true
//}

func (m *movieWindow) deleteMovie() bool {
	var title string
	if len(m.guiMovie.title) <= 30 {
		title = m.guiMovie.title
	} else {
		title = m.guiMovie.title[:27] + "..."
	}
	msg := fmt.Sprintf("Do you want to delete the movie '%s'?", title)
	response, err := dialog.Title("Delete movie...").Text(msg).YesNoButtons().Width(450).
		WarningIcon().Show()

	if err != nil {
		return false
	}

	if response == gtk.RESPONSE_YES {
		return true
	}

	return false
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

func (m *movieWindow) showSimilarMovies(movies []similarMovie) {
	titles := ""

	for _, mov := range movies {
		titles += fmt.Sprintf("%s\n", mov.title)
	}
	titles = strings.TrimSuffix(titles, "\n")

	if titles != "" {
		_, _ = dialog.Title("Similar movies...").
			Text("There are similar movies in the DB:").
			ExtraHeight(90).
			ExtraExpand(titles).
			WarningIcon().
			OkButton().
			Show()
	}
}

//
//func (m *movieWindow) findSimilarMovies(newTitle string) []movie {
//	var movies []movie
//
//	for _, existingTitle := range movieTitles {
//		l := gstr.Levenshtein(existingTitle, newTitle, 1, 3, 1)
//		if containsI(newTitle, existingTitle) {
//			l = 2
//		}
//		if containsI(existingTitle, newTitle) {
//			l = 2
//		}
//		if equalsI(newTitle, existingTitle) {
//			l = 1
//		}
//		movies = append(movies, movie{
//			l,
//			existingTitle,
//		})
//	}
//	slices.SortFunc(movies, func(a, b movie) int {
//		return a.distance - b.distance
//	})
//	return movies[:10]
//}

func (m *movieWindow) hasSubtitles(dir string) bool {
	movieFile, err := findMovieFile(filepath.Join(m.config.RootDir, dir))
	if err != nil {
		return false
	}
	movieFile = strings.TrimSuffix(movieFile, filepath.Ext(movieFile))

	srtPath := filepath.Join(m.config.RootDir, dir, movieFile+".srt")
	subPath := filepath.Join(m.config.RootDir, dir, movieFile+".sub")

	if doesExist(srtPath) || doesExist(subPath) {
		return true
	}

	return false
}

func (m *movieWindow) fillCastAndCrewPage() {
	// Clear the list before refreshing the list
	m.castAndCrewList.GetChildren().Foreach(func(item interface{}) {
		m.castAndCrewList.Remove(item.(gtk.IWidget))
	})

	m.addPeopleSection("Director(s)", data.Director, true)
	m.addPeopleSection("Writer(s)", data.Writer, true)
	m.addPeopleSection("Actor(s)", data.Actor, false)
}

func (m *movieWindow) addPeopleSection(title string, personType data.PersonType, addSpacer bool) {
	// Section title
	m.castAndCrewList.Add(getLabel(title, true))

	// People of this type
	for _, person := range m.dataMovie.Persons {
		if person.Type == personType {
			m.castAndCrewList.Add(getLabel(person.Name, false))
		}
	}

	if addSpacer {
		// Spacer
		m.castAndCrewList.Add(getLabel("", false))
	}
}

func getLabel(text string, header bool) *gtk.Label {
	label, err := gtk.LabelNew("")
	if err != nil {
		log.Fatal(err)
	}
	size := 15
	weight := "normal"
	color := "#f1e3ae"
	if header {
		size = 18
		weight = "bold"
		color = "#91834e"
	}
	m := fmt.Sprintf("<span foreground='%s' weight='%s' size='%dpt'>%s</span>", color, weight, size, text)
	label.SetMarkup(m)
	return label
}

func findSimilarMovies(newTitle string, existingTitles []string, maxReturned int) []similarMovie {
	var similar []similarMovie

	if newTitle == "" {
		return []similarMovie{}
	}

	for _, title := range existingTitles {
		score := compareTitles(newTitle, title)
		similar = append(similar, similarMovie{
			score,
			title,
		})
	}

	sort.Slice(similar, func(i, j int) bool {
		return similar[i].distance > similar[j].distance
	})

	if len(similar) > maxReturned {
		return similar[:maxReturned]
	}
	return similar
}

// Compute similarity between two movie titles
func compareTitles(title1, title2 string) int {
	// Split titles into words
	words1 := strings.Fields(strings.ToLower(title1))
	words2 := strings.Fields(strings.ToLower(title2))

	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}

	// Compare each word in title1 with the most similar word in title2
	totalWordScore := 0.0
	for _, w1 := range words1 {
		best := 0.0
		for _, w2 := range words2 {
			// Calculate Levenshtein distance for each word
			dist := levenshtein.DistanceForStrings([]rune(w1), []rune(w2), levenshtein.DefaultOptions)
			// Convert distance to similarity score (normalized to 0-1)
			score := 1 - float64(dist)/float64(len(w1)+len(w2)) // Normalized similarity
			if score > best {
				best = score
			}
		}
		totalWordScore += best
	}

	wordScore := totalWordScore / float64(len(words1))

	// Levenshtein distance on full titles (normalized to 0-1)
	fullDist := levenshtein.DistanceForStrings([]rune(title1), []rune(title2), levenshtein.DefaultOptions)
	fullTitleScore := 1 - float64(fullDist)/float64(len(title1)+len(title2))

	// Weighted similarity: 30% word-based, 70% full-string Levenshtein
	finalScore := (wordScore * 0.3) + (fullTitleScore * 0.7)

	return int(finalScore * 1000)
}

func (m *movieWindow) onTitleEntryFocusOut() {
	if showSimilarOnce {
		return
	}
	showSimilarOnce = true

	title := getEntryText(m.titleEntry)
	if title == "" {
		return
	}

	similar := findSimilarMovies(title, movieTitles, 5)
	m.showSimilarMovies(similar)
}

func (m *movieWindow) onIMDBEntryFocusOut() {
	if scrapeImdbOnce {
		return
	}

	url := getEntryText(m.imdbUrlEntry)
	if url == "" {
		return
	}

	manager := imdb.ManagerNew()
	movieImdb, err := manager.GetMovie(url)

	if err != nil {
		txt := ""
		for _, err := range manager.Errors {
			txt += err.Error() + "\n"
		}
		_, _ = dialog.Title("Errors while retrieving IMDB data...").
			Text(txt).WarningIcon().OkButton().Show()
	}

	if m.createMovieInfo(movieImdb) {
		return
	}

	scrapeImdbOnce = true
}

func (m *movieWindow) createMovieInfo(movieImdb *imdb.MovieImdb) bool {
	if movieImdb == nil {
		return true
	}

	m.titleEntry.SetText(movieImdb.Title)
	m.yearEntry.SetText(strconv.Itoa(movieImdb.Year))
	m.ratingEntry.SetText(movieImdb.Rating)
	m.runtimeEntry.SetText(strconv.Itoa(movieImdb.Runtime))
	genres := strings.Join(movieImdb.Genres, ", ")
	m.genresEntry.SetText(genres)

	// Story line
	buffer, err := gtk.TextBufferNew(nil)
	if err != nil {
		reportError(err)
		return true
	}
	buffer.SetText(movieImdb.StoryLine)
	m.storyLineEntry.SetBuffer(buffer)

	// Movie poster
	fileName, err := saveMoviePoster(movieImdb.Title, movieImdb.Poster)
	if err != nil {
		reportError(err)
		return true
	}

	fileData := getCorrectImageSize(fileName)
	if fileData == nil || len(fileData) == 0 {
		return true
	}
	m.updateImage(fileData)
	m.guiMovie.image = fileData
	m.guiMovie.imageHasChanged = true

	var p data.Person
	for _, person := range movieImdb.Persons {
		p.Name = person.Name
		p.Type = data.PersonType(person.Type)
		m.guiMovie.persons = append(m.guiMovie.persons, p)
	}
	return false
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

	home, err := os.UserHomeDir()
	if err == nil {
		dir := path.Join(home, "Downloads")
		_ = dlg.SetCurrentFolder(dir)
	}

	response := dlg.Run()
	if response == gtk.RESPONSE_CANCEL {
		return
	}

	fileData := getCorrectImageSize(dlg.GetFilename())
	if fileData == nil || len(fileData) == 0 {
		return
	}
	m.updateImage(fileData)
	m.guiMovie.image = fileData
	m.guiMovie.imageHasChanged = true
}
