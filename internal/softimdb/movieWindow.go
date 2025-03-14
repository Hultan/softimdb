package softimdb

import (
	"fmt"
	_ "image/jpeg"
	"log"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

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

	movieInfo *movieInfo
	movie     *data.Movie

	config *config.Config
	db     *data.Database

	closeCallback func(gtk.ResponseType, *movieInfo, *data.Movie)
}

var scrapeImdbOnce bool
var showSimilarOnce bool

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
		m.closeCallback(gtk.RESPONSE_ACCEPT, m.movieInfo, m.movie)
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
			m.closeCallback(gtk.RESPONSE_REJECT, nil, m.movie)
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

	eventBox := builder.GetObject("imageEventBox").(*gtk.EventBox)
	eventBox.Connect("button-press-event", m.onImageClick)
	m.titleEntry.Connect("focus-out-event", m.onTitleEntryFocusOut)
	m.imdbUrlEntry.Connect("focus-out-event", m.onIMDBEntryFocusOut)

	return m
}

func (m *movieWindow) open(info *movieInfo, movie *data.Movie, closeCallback func(gtk.ResponseType, *movieInfo,
	*data.Movie)) {

	if info == nil {
		info = &movieInfo{}
		// TODO : Disable cast and crew tab here
	}

	scrapeImdbOnce = false
	showSimilarOnce = false
	if info.title == "" {
		//  New movie
		info.toWatch = true
		info.needsSubtitle = !m.hasSubtitles(info.path)
	} else {
		// Edit movie
		scrapeImdbOnce = true
		showSimilarOnce = true
	}

	// We show the window here as well, since we are going to the
	// database to load person below, and that causes a short delay
	m.window.ShowAll()

	if movie != nil {
		// Load persons for movie (they are no longer loaded in the main load)
		persons, err := m.db.GetPersonsForMovie(movie)
		if err != nil {
			return
		}
		movie.Persons = persons
	}

	m.movie = movie
	m.movieInfo = info
	m.closeCallback = closeCallback
	m.deleteButton.SetSensitive(movie != nil)

	// Fill form with data
	m.imdbUrlEntry.SetText(m.movieInfo.imdbUrl)
	m.pathEntry.SetText(m.movieInfo.path)
	m.titleEntry.SetText(m.movieInfo.title)
	m.subTitleEntry.SetText(m.movieInfo.subTitle)
	m.yearEntry.SetText(fmt.Sprintf("%d", m.movieInfo.getYear()))
	m.myRatingEntry.SetText(fmt.Sprintf("%d", m.movieInfo.myRating))
	m.toWatchCheckButton.SetActive(m.movieInfo.toWatch)
	m.needsSubtitleCheckButton.SetActive(m.movieInfo.needsSubtitle)
	buffer, err := gtk.TextBufferNew(nil)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	buffer.SetText(m.movieInfo.storyLine)
	m.storyLineEntry.SetBuffer(buffer)
	m.ratingEntry.SetText(m.movieInfo.imdbRating)
	m.genresEntry.SetText(m.movieInfo.genres)
	m.packEntry.SetText(m.movieInfo.pack)
	m.runtimeEntry.SetText(strconv.Itoa(m.movieInfo.runtime))

	if m.movieInfo.image == nil {
		m.posterImage.Clear()
	} else {
		m.updateImage(m.movieInfo.image)
	}

	if m.movie != nil {
		m.fillCastAndCrewPage()
	}
	m.movieStack.SetVisibleChildName("MoviePage")

	m.window.ShowAll()
	m.imdbUrlEntry.GrabFocus()
}

func (m *movieWindow) saveMovie() bool {
	// Fill fields
	m.movieInfo.path = getEntryText(m.pathEntry)
	m.movieInfo.imdbUrl = getEntryText(m.imdbUrlEntry)
	id, err := getIdFromUrl(m.movieInfo.imdbUrl)
	if err != nil {
		msg := fmt.Sprintf("Failed to retrieve IMDB id from url : %s", err)
		_, err = dialog.Title("Invalid IMDB url...").Text(msg).ErrorIcon().OkButton().Show()

		if err != nil {
			fmt.Printf("Error : %s", err)
		}

		return false
	}
	m.movieInfo.imdbId = id
	m.movieInfo.title = getEntryText(m.titleEntry)
	m.movieInfo.subTitle = getEntryText(m.subTitleEntry)
	m.movieInfo.pack = getEntryText(m.packEntry)
	m.movieInfo.year = getEntryText(m.yearEntry)

	ratingText := getEntryText(m.myRatingEntry)
	rating, err := strconv.Atoi(ratingText)
	legalRating := rating >= 0 && rating <= 5
	if err != nil || !legalRating {
		msg := fmt.Sprintf("Invalid my rating : %s (error : %s)", ratingText, err)
		_, err = dialog.Title("Invalid my rating...").Text(msg).ErrorIcon().OkButton().Show()

		if err != nil {
			fmt.Printf("Error : %s", err)
		}

		return false
	}
	m.movieInfo.myRating = rating

	runtimeText := getEntryText(m.runtimeEntry)
	runtime, err := strconv.Atoi(runtimeText)
	legalRuntime := runtime == -1 || runtime > 0
	if err != nil || !legalRuntime {
		msg := fmt.Sprintf("Invalid runtime : %s (error : %s)", runtimeText, err)
		_, err = dialog.Title("Invalid runtime...").Text(msg).ErrorIcon().OkButton().Show()

		if err != nil {
			fmt.Printf("Error : %s", err)
		}

		return false
	}
	m.movieInfo.runtime = runtime

	m.movieInfo.toWatch = m.toWatchCheckButton.GetActive()
	m.movieInfo.needsSubtitle = m.needsSubtitleCheckButton.GetActive()
	m.movieInfo.imdbRating = getEntryText(m.ratingEntry)
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
	m.movieInfo.storyLine = storyLine
	m.movieInfo.genres = getEntryText(m.genresEntry)
	// Poster is set when clicking on the image

	//for _, person := range m.movie.Persons {
	//	p := data.Person{
	//		Name: person.Name,
	//		Type: person.Type,
	//	}
	//	m.movieInfo.persons = append(m.movieInfo.persons, p)
	//}

	return true
}

func (m *movieWindow) deleteMovie() bool {
	var title string
	if len(m.movieInfo.title) <= 30 {
		title = m.movieInfo.title
	} else {
		title = m.movieInfo.title[:27] + "..."
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
	m.movieInfo.image = fileData
	m.movieInfo.imageHasChanged = true
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

func (m *movieWindow) onTitleEntryFocusOut() {
	if showSimilarOnce {
		return
	}
	showSimilarOnce = true

	title, err := m.titleEntry.GetText()
	if err != nil {
		return
	}

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

	url, err := m.imdbUrlEntry.GetText()
	if err != nil {
		return
	}

	if url == "" {
		return
	}

	// TODO : Ask question
	manager := imdb.ManagerNew()
	movieImdb, err := manager.GetMovie(url)

	if movieImdb != nil {
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
			log.Fatal(err)
		}
		buffer.SetText(movieImdb.StoryLine)
		m.storyLineEntry.SetBuffer(buffer)

		// Movie poster
		fileName, err := saveMoviePoster(movieImdb.Title, movieImdb.Poster)
		if err != nil {
			reportError(err)
			return
		} else {
			fileData := getCorrectImageSize(fileName)
			if fileData == nil || len(fileData) == 0 {
				return
			}
			m.updateImage(fileData)
			m.movieInfo.image = fileData
			m.movieInfo.imageHasChanged = true
		}

		var p data.Person
		for _, person := range movieImdb.Persons {
			p.Name = person.Name
			p.Type = data.PersonType(person.Type)
			m.movieInfo.persons = append(m.movieInfo.persons, p)
		}
	}

	if err != nil {
		txt := ""
		for _, err := range manager.Errors {
			txt += err.Error() + "\n"
		}
		_, _ = dialog.Title("Errors while retrieving IMDB data...").Text(txt).WarningIcon().OkButton().Show()
	}

	scrapeImdbOnce = true
}

func (m *movieWindow) showSimilarMovies(movies []movie) {
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

type movie struct {
	distance int
	title    string
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
	// Clear the list before adding new items
	m.castAndCrewList.GetChildren().Foreach(func(item interface{}) {
		m.castAndCrewList.Remove(item.(gtk.IWidget))
	})

	// Director(s)
	label := getLabel("Director(s)", true)
	m.castAndCrewList.Add(label)

	for _, person := range m.movie.Persons {
		if person.Type == data.Director {
			m.castAndCrewList.Add(getLabel(person.Name, false))
		}
	}

	label = getLabel("", false)
	m.castAndCrewList.Add(label)

	// Writer(s)
	label = getLabel("Writer(s)", true)
	m.castAndCrewList.Add(label)

	for _, person := range m.movie.Persons {
		if person.Type == data.Writer {
			m.castAndCrewList.Add(getLabel(person.Name, false))
		}
	}

	label = getLabel("", false)
	m.castAndCrewList.Add(label)

	// Actor(s)
	label = getLabel("Actor(s)", true)
	m.castAndCrewList.Add(label)

	for _, person := range m.movie.Persons {
		if person.Type == data.Actor {
			m.castAndCrewList.Add(getLabel(person.Name, false))
		}
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

func findSimilarMovies(newTitle string, existingTitles []string, maxReturnedMovies int) []movie {
	var movies []movie

	if newTitle == "" {
		return []movie{}
	}

	for _, existingTitle := range existingTitles {
		l := compareTitles(newTitle, existingTitle)
		movies = append(movies, movie{
			l,
			existingTitle,
		})
	}
	slices.SortFunc(movies, func(a, b movie) int {
		if a.distance > b.distance {
			return -1
		} else if a.distance == b.distance {
			return 0
		}
		return 1
	})
	//fmt.Println(movies)
	return movies[:maxReturnedMovies]
}

// Compute similarity between two movie titles
func compareTitles(title1, title2 string) int {
	// Split titles into words
	words1 := strings.Fields(strings.ToLower(title1))
	words2 := strings.Fields(strings.ToLower(title2))

	// Matrix to track best word matches
	wordMatches := make([]float64, len(words1))

	// Compare each word in title1 with the most similar word in title2
	for i, w1 := range words1 {
		bestScore := 0.0
		for _, w2 := range words2 {
			// Calculate Levenshtein distance for each word
			dist := levenshtein.DistanceForStrings([]rune(w1), []rune(w2), levenshtein.DefaultOptions)
			// Convert distance to similarity score (normalized to 0-1)
			score := 1 - float64(dist)/float64(len(w1)+len(w2)) // Normalized similarity
			if score > bestScore {
				bestScore = score
			}
		}
		wordMatches[i] = bestScore // Store best match
	}

	// Average word match similarity
	wordScore := 0.0
	for _, s := range wordMatches {
		wordScore += s
	}
	wordScore /= float64(len(words1))

	// Levenshtein distance on full titles (normalized to 0-1)
	fullDist := levenshtein.DistanceForStrings([]rune(title1), []rune(title2), levenshtein.DefaultOptions)
	fullTitleScore := 1 - float64(fullDist)/float64(len(title1)+len(title2))

	// Weighted similarity: 30% word-based, 70% full-string Levenshtein
	finalScore := (wordScore * 0.3) + (fullTitleScore * 0.7)

	return int(finalScore * 1000)
}
