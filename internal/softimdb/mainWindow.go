package softimdb

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hultan/dialog"

	"github.com/hultan/softimdb/internal/builder"
	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
)

//go:embed assets/application.png
var applicationIcon []byte

//go:embed assets/main.glade
var mainGlade string

//go:embed assets/toWatch.png
var toWatchIcon []byte

//go:embed assets/needsSubtitle.png
var needsSubtitleIcon []byte

//go:embed assets/softimdb.css
var mainCss string

type Sort struct {
	by, order string
}

type Search struct {
	genreId int
	forWhat string
}

type View struct {
	manager viewManager
	current ViewType
}

type GTK struct {
	application                           *gtk.Application
	window                                *gtk.ApplicationWindow
	movieList                             *gtk.FlowBox
	storyLineLabel                        *gtk.Label
	storyLineScrolledWindow               *gtk.ScrolledWindow
	searchEntry                           *gtk.Entry
	countLabel                            *gtk.Label
	menuNoGenreItem                       *gtk.RadioMenuItem
	menuSortByName, menuSortByRating      *gtk.RadioMenuItem
	menuSortByMyRating, menuSortByLength  *gtk.RadioMenuItem
	menuSortByYear, menuSortById          *gtk.RadioMenuItem
	menuSortAscending, menuSortDescending *gtk.RadioMenuItem
	genresSubMenu                         *gtk.Menu
	genresMenu                            *gtk.MenuItem
}

type MainWindow struct {
	builder     *builder.Builder
	database    *data.Database
	config      *config.Config
	popupMenu   *popupMenu
	movieWin    *movieWindow
	addMovieWin *addMovieWindow

	gtk    GTK
	search Search
	sort   Sort
	view   View

	movies map[int]*data.Movie
}

var (
	movieTitles       []string
	showPrivateGenres = true
)

// NewMainWindow : Creates a new MainWindow object
func NewMainWindow() *MainWindow {
	m := &MainWindow{}
	m.movies = make(map[int]*data.Movie, 2000)

	b, err := builder.NewBuilder(mainGlade)
	if err != nil {
		log.Fatal(err)
	}
	m.builder = b

	m.gtk.window = m.builder.GetObject("mainWindow").(*gtk.ApplicationWindow)
	m.gtk.window.SetTitle(fmt.Sprintf("%s - %s", applicationTitle, applicationVersion))
	m.gtk.window.Maximize()
	_ = m.gtk.window.Connect("destroy", m.onClose)
	_ = m.gtk.window.Connect("key-press-event", m.onKeyPressEvent)

	cnf, err := config.LoadConfig(configFile)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	m.config = cnf

	// Open the database after we have the config
	m.database = data.DatabaseNew(false, cnf)

	m.setupToolBar()
	m.popupMenu = newPopupMenu(m)
	m.popupMenu.setup()
	m.setupMenu(m.gtk.window)
	m.gtk.storyLineLabel = m.builder.GetObject("storyLineLabel").(*gtk.Label)
	m.gtk.storyLineScrolledWindow = m.builder.GetObject("storyLineScrolledWindow").(*gtk.ScrolledWindow)
	versionLabel := m.builder.GetObject("versionLabel").(*gtk.Label)
	versionLabel.SetText("Version : " + applicationVersion)
	m.gtk.countLabel = m.builder.GetObject("countLabel").(*gtk.Label)

	// Movie list
	m.gtk.movieList = m.builder.GetObject("movieList").(*gtk.FlowBox)
	m.gtk.movieList.SetSelectionMode(gtk.SELECTION_SINGLE)
	m.gtk.movieList.SetRowSpacing(listSpacing)
	m.gtk.movieList.SetColumnSpacing(listSpacing)
	m.gtk.movieList.SetMarginTop(listMargin)
	m.gtk.movieList.SetMarginBottom(listMargin)
	m.gtk.movieList.SetMarginStart(listMargin)
	m.gtk.movieList.SetMarginEnd(listMargin)
	m.gtk.movieList.SetActivateOnSingleClick(false)
	m.gtk.movieList.SetFocusOnClick(true)
	_ = m.gtk.movieList.Connect("selected-children-changed", m.onMovieListSelectionChanged)
	_ = m.gtk.movieList.Connect("child-activated", m.onMovieListDoubleClicked)

	m.view.manager = newViewManager(m)

	return m
}

// Open : Opens the MainWindow window
func (m *MainWindow) Open(app *gtk.Application) {
	m.search.genreId = -1
	m.gtk.application = app
	m.gtk.window.SetApplication(app)
	m.gtk.window.ShowAll()
	m.view.manager.changeView(viewToWatch)
	m.gtk.storyLineScrolledWindow.Hide()

	var err error
	movieTitles, err = m.database.GetAllMovieTitles()
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
}

func (m *MainWindow) setupMenu(window *gtk.ApplicationWindow) {
	// File menu
	menuQuit := m.builder.GetObject("menuFileQuit").(*gtk.MenuItem)
	_ = menuQuit.Connect("activate", window.Close)

	// Help menu
	menuHelpAbout := m.builder.GetObject("menuHelpAbout").(*gtk.MenuItem)
	_ = menuHelpAbout.Connect("activate", m.onOpenAboutDialogClicked)

	// Sort menu
	m.setupSortMenuItem("menuSortByName", sortByName, sortAscending)
	m.setupSortMenuItem("menuSortByRating", sortByRating, sortDescending)
	m.setupSortMenuItem("menuSortByMyRating", sortByMyRating, sortDescending)
	m.setupSortMenuItem("menuSortByLength", sortByLength, sortDescending)
	m.setupSortMenuItem("menuSortByYear", sortByYear, sortDescending)
	m.setupSortMenuItem("menuSortById", sortById, sortAscending)

	// Sorting order radio items
	m.setupSortOrderMenuItem("menuSortAscending", sortAscending)
	m.setupSortOrderMenuItem("menuSortDescending", sortDescending)

	m.sort.by = sortByName
	m.sort.order = sortAscending

	// Genres menu
	m.gtk.genresMenu = m.builder.GetObject("menuGenres").(*gtk.MenuItem)
	m.fillGenresMenu()
}

// helper function to set up sorting menu items
func (m *MainWindow) setupSortMenuItem(name string, sortBy string, defaultOrder string) {
	menuItem := m.builder.GetObject(name).(*gtk.RadioMenuItem)

	menuItem.Connect("activate", func() {
		if menuItem.GetActive() {
			m.sort.by = sortBy
			m.sort.order = defaultOrder

			// Update order menu
			if defaultOrder == sortAscending {
				m.gtk.menuSortAscending.SetActive(true)
			} else {
				m.gtk.menuSortDescending.SetActive(true)
			}

			m.refresh(m.search, m.sort)
		}
	})

	// Store in m.gtk for future access
	switch name {
	case "menuSortByName":
		m.gtk.menuSortByName = menuItem
	case "menuSortByRating":
		m.gtk.menuSortByRating = menuItem
	case "menuSortByMyRating":
		m.gtk.menuSortByMyRating = menuItem
	case "menuSortByLength":
		m.gtk.menuSortByLength = menuItem
	case "menuSortByYear":
		m.gtk.menuSortByYear = menuItem
	case "menuSortById":
		m.gtk.menuSortById = menuItem
	}
}

func (m *MainWindow) setupSortOrderMenuItem(name string, order string) {
	menuItem := m.builder.GetObject(name).(*gtk.RadioMenuItem)

	menuItem.Connect("activate", func() {
		if menuItem.GetActive() {
			m.sort.order = order
			m.refresh(m.search, m.sort)
		}
	})

	// Store reference in m.gtk
	switch name {
	case "menuSortAscending":
		m.gtk.menuSortAscending = menuItem
	case "menuSortDescending":
		m.gtk.menuSortDescending = menuItem
	}
}

func (m *MainWindow) setupToolBar() {
	m.connectToolButton("quitButton", m.gtk.window.Close)
	m.connectToolButton("refreshButton", m.onRefreshButtonClicked)
	m.connectToolButton("playMovieButton", m.onPlayMovieClicked)
	m.connectToolButton("addButton", m.onOpenAddWindowClicked)
	m.connectToolButton("searchButton", m.onSearchButtonClicked)
	m.connectToolButton("clearSearchButton", m.onClearSearchButtonClicked)

	// Search entry
	m.gtk.searchEntry = m.builder.GetObject("searchEntry").(*gtk.Entry)
	_ = m.gtk.searchEntry.Connect("activate", m.onSearchButtonClicked)
}

func (m *MainWindow) connectToolButton(name string, handler func()) {
	button := m.builder.GetObject(name).(*gtk.ToolButton)
	_ = button.Connect("clicked", handler)
}

func (m *MainWindow) fillMovieList(search Search, sort Sort) {
	movies, err := m.database.SearchMovies(string(m.view.current), search.forWhat, search.genreId, getSortBy(sort))
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	listHelper := &ListHelper{}
	clearFlowBox(m.gtk.movieList)
	runtime.GC()

	cssProvider, _ := gtk.CssProviderNew()
	if err = cssProvider.LoadFromData(mainCss); err != nil {
		reportError(err)
		log.Fatal(err)
	}

	screen, err := gdk.ScreenGetDefault()
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	for i := range movies {
		movie := movies[i]
		m.movies[movie.Id] = movie
		card := listHelper.CreateMovieCard(movie)
		card.SetName("movie_" + strconv.Itoa(movie.Id))
		m.gtk.movieList.Add(card)
	}

	m.updateCountLabel(len(movies))
}

func (m *MainWindow) getSelectedMovie() *data.Movie {
	children := m.gtk.movieList.GetSelectedChildren()
	if children == nil {
		return nil
	}
	selected := children[0]
	frameObj, err := selected.GetChild()
	if err != nil {
		return nil
	}
	if frameObj == nil {
		return nil
	}
	frame, ok := frameObj.(*gtk.Frame)
	if !ok {
		return nil
	}
	name, err := frame.GetName()
	if err != nil {
		return nil
	}
	id, err := strconv.Atoi(name[6:]) // Name is movie_<id>
	if err != nil {
		return nil
	}
	return m.movies[id]
}

func (m *MainWindow) refresh(search Search, sort Sort) {
	m.fillMovieList(search, sort)
	m.gtk.movieList.ShowAll()
	if m.search.forWhat == "" {
		m.gtk.searchEntry.SetText("")
	}
}

func (m *MainWindow) fillGenresMenu() {
	genres, _ := m.database.GetGenres()

	// Create and add the genre menu
	sub, _ := gtk.MenuNew()
	m.gtk.genresSubMenu = sub
	m.gtk.genresMenu.SetSubmenu(sub)

	// No genre item (create a fake no genre item)
	genre := data.Genre{Id: -1, Name: "None"}
	m.gtk.menuNoGenreItem = m.addGenreMenu(sub, nil, genre)
	m.gtk.menuNoGenreItem.SetActive(true)
	group, _ := m.gtk.menuNoGenreItem.GetGroup()

	// Separator
	sep, _ := gtk.SeparatorMenuItemNew()
	sub.Add(sep)

	// Genre items
	for _, genre := range genres {
		if showPrivateGenres || !genre.IsPrivate {
			m.addGenreMenu(sub, group, genre)
		}
	}
}

func (m *MainWindow) addGenreMenu(sub *gtk.Menu, group *glib.SList, genre data.Genre) *gtk.RadioMenuItem {
	item, _ := gtk.RadioMenuItemNewWithLabel(group, genre.Name)
	item.SetName(strconv.Itoa(genre.Id))
	item.Connect(
		"activate", func() {
			if item.GetActive() {
				m.search.genreId = genre.Id
				m.refresh(m.search, m.sort)
			}
		},
	)
	sub.Add(item)
	return item
}

func (m *MainWindow) saveMovieInfo(movieInfo *movieInfo, movie *data.Movie) {
	movieInfo.toDatabase(movie)

	err := m.database.UpdateMovie(movie)
	if err != nil {
		reportError(err)
		return
	}

	if movieInfo.imageHasChanged {
		err = m.database.UpdateImage(movie, movieInfo.image)
		if err != nil {
			reportError(err)
			return
		}
		// TODO : Remove after update cache
		_, _ = dialog.
			Title("Restart needed!").
			Text("You need to restart the application to see the image change.").
			OkButton().
			Show()
	}

	if movie.SubTitle != "" {
		movieTitles = append(movieTitles, fmt.Sprintf("%s (%s)", movie.Title, movie.SubTitle))
	} else {
		movieTitles = append(movieTitles, movie.Title)
	}
}

func (m *MainWindow) deleteMovie(movie *data.Movie) {
	err := m.database.DeleteMovie(m.config.RootDir, movie)
	var pathError *fs.PathError
	if errors.Is(err, pathError) {
		moviePath := path.Join(m.config.RootDir, movie.MoviePath)
		msg := fmt.Sprintf("Failed to delete movie from NAS. "+
			"Some directories or files might need to be removed manually from path='%s'.", moviePath)

		_, _ = dialog.Title("Failed to delete movie").Text(msg).ExtraExpand(err.Error()).
			ErrorIcon().OkButton().Show()
		return
	} else if err != nil {
		//reportError(err)
		return
	}

	_, _ = dialog.Title("Movie deleted...").Text("The movie has been deleted!").
		InfoIcon().OkButton().Show()
}

func (m *MainWindow) updateCountLabel(i int) {
	m.gtk.countLabel.SetText(fmt.Sprintf("Number of videos : %d", i))
}

//
// Signal handlers
//

func (m *MainWindow) onClose() {
	m.database.CloseDatabase()
	m.gtk.window.Close()
	m.gtk.movieList = nil
	m.gtk.storyLineLabel = nil
	m.gtk.window = nil
	m.builder = nil
	m.movieWin = nil
	m.addMovieWin = nil
	m.gtk.application.Quit()
}

func (m *MainWindow) onOpenIMDBClicked() {
	v := m.getSelectedMovie()
	if v == nil {
		return
	}

	openBrowser(v.ImdbUrl)
}

func (m *MainWindow) onPlayMovieClicked() {
	go func() {
		movie := m.getSelectedMovie()
		if movie == nil {
			return
		}

		moviePath := path.Join(m.config.RootDir, movie.MoviePath)
		movieName, err := findMovieFile(moviePath)
		if err != nil {
			reportError(err)
			return
		}
		if movieName == "" {
			openInNemo(path.Join(m.config.RootDir, movie.MoviePath))
			return
		}
		moviePath = path.Join(moviePath, movieName)
		openProcess("smplayer", moviePath)
	}()
}

func (m *MainWindow) onEditMovieInfoClicked() {
	selectedMovie := m.getSelectedMovie()
	if selectedMovie == nil {
		return
	}

	info := &movieInfo{}
	info.fromDatabase(selectedMovie)

	// Open the movie dialog here
	if m.movieWin == nil {
		m.movieWin = newMovieWindow(m.builder, m.gtk.window, m.database, m.config)
	}

	m.movieWin.open(info, selectedMovie, m.onWindowClosed)
}

func (m *MainWindow) onOpenAddWindowClicked() {
	if m.addMovieWin == nil {
		m.addMovieWin = newAddMovieWindow(m, m.database, m.config)
	}
	m.addMovieWin.open()
}

func (m *MainWindow) onRefreshButtonClicked() {
	m.search.forWhat = ""
	m.search.genreId = -1
	m.sort.by = sortByName
	m.sort.order = sortAscending
	m.gtk.menuNoGenreItem.SetActive(true)
	m.gtk.menuSortByName.SetActive(true)
	m.gtk.menuSortAscending.SetActive(true)
	m.refresh(m.search, m.sort)
}

func (m *MainWindow) onSearchButtonClicked() {
	search := getEntryText(m.gtk.searchEntry)
	m.search.forWhat = strings.Trim(search, " ")
	m.refresh(m.search, m.sort)
}

func (m *MainWindow) onClearSearchButtonClicked() {
	m.search.forWhat = ""
	m.gtk.searchEntry.SetText("")
	m.refresh(m.search, m.sort)
}

func (m *MainWindow) onKeyPressEvent(_ *gtk.ApplicationWindow, event *gdk.Event) {
	keyEvent := gdk.EventKeyNewFromEvent(event)

	special := (keyEvent.State() & gdk.MOD2_MASK) != 0 // Used for special keys like F5, DELETE, HOME in X11.
	ctrl := (keyEvent.State() & gdk.CONTROL_MASK) != 0

	if special {
		switch {
		case keyEvent.KeyVal() == gdk.KEY_F5:
			m.onRefreshButtonClicked()
		case keyEvent.KeyVal() == gdk.KEY_F6:
			m.onPlayMovieClicked()
		case keyEvent.KeyVal() == gdk.KEY_Escape:
			m.gtk.movieList.UnselectAll()
		}
	}
	if ctrl {
		switch {
		case keyEvent.KeyVal() == gdk.KEY_i:
			m.onOpenIMDBClicked()
		case keyEvent.KeyVal() == gdk.KEY_p:
			m.onOpenPackClicked()
		case keyEvent.KeyVal() == gdk.KEY_h:
			m.onShowHidePrivateClicked()
		case keyEvent.KeyVal() == gdk.KEY_f:
			m.gtk.searchEntry.GrabFocus()
		case keyEvent.KeyVal() == gdk.KEY_a:
			m.onOpenAddWindowClicked()
		case keyEvent.KeyVal() == gdk.KEY_q:
			m.onClose()
		case keyEvent.KeyVal() == gdk.KEY_o:
			m.onOpenFolderClicked()
		}
	}
}

func (m *MainWindow) onOpenAboutDialogClicked() {
	about := m.builder.GetObject("aboutDialog").(*gtk.AboutDialog)

	about.SetDestroyWithParent(true)
	about.SetTransientFor(m.gtk.window)
	about.SetProgramName(applicationTitle)
	about.SetComments("A movie library application...")
	about.SetVersion(applicationVersion)
	about.SetCopyright(applicationCopyRight)

	image, err := gdk.PixbufNewFromBytesOnly(applicationIcon)
	if err == nil {
		about.SetLogo(image)
	}

	about.SetModal(true)

	_ = about.Connect(
		"response", func(dialog *gtk.AboutDialog, responseId gtk.ResponseType) {
			if responseId == gtk.RESPONSE_CANCEL || responseId == gtk.RESPONSE_DELETE_EVENT {
				about.Hide()
			}
		},
	)

	about.ShowAll()
}

func (m *MainWindow) onMovieListSelectionChanged(_ *gtk.FlowBox) {
	movie := m.getSelectedMovie()
	if movie == nil {
		m.gtk.storyLineScrolledWindow.SetVisible(false)
		return
	}
	story := `<span font="Sans Regular 10" foreground="#d49c6b">` + cleanString(movie.StoryLine) + `</span>`
	m.gtk.storyLineLabel.SetMarkup(story)
	m.gtk.storyLineScrolledWindow.SetVisible(true)
}

func (m *MainWindow) onMovieListDoubleClicked(_ *gtk.FlowBox) {
	movie := m.getSelectedMovie()
	if movie == nil {
		return
	}
	m.onEditMovieInfoClicked()
}

func (m *MainWindow) onOpenPackClicked() {
	movie := m.getSelectedMovie()
	if movie == nil {
		return
	}

	m.search.forWhat = "pack:" + movie.Pack
	m.search.genreId = -1
	m.sort.by = sortByName
	m.sort.order = sortAscending
	m.view.manager.changeView(viewPacks)
	m.gtk.searchEntry.SetText(m.search.forWhat)
	m.gtk.menuNoGenreItem.SetActive(true)
	m.gtk.menuSortByName.SetActive(true)
	m.gtk.menuSortAscending.SetActive(true)
	m.refresh(m.search, m.sort)
}

func (m *MainWindow) onOpenFolderClicked() {
	movie := m.getSelectedMovie()
	if movie == nil {
		return
	}
	openInNemo(path.Join(m.config.RootDir, movie.MoviePath))
}

func (m *MainWindow) onWindowClosed(r gtk.ResponseType, info *movieInfo, movie *data.Movie) {
	switch r {
	case gtk.RESPONSE_ACCEPT:
		// Save movie
		m.saveMovieInfo(info, movie)
	case gtk.RESPONSE_CANCEL:
		// Cancel dialog
	case gtk.RESPONSE_REJECT:
		// Delete movie
		m.deleteMovie(movie)
	default:
		// Unknown response
		// Handle as cancel
	}
}

func (m *MainWindow) onShowHidePrivateClicked() {
	showPrivateGenres = !showPrivateGenres
	m.refresh(m.search, m.sort)

	m.gtk.genresSubMenu.Destroy()
	m.gtk.genresSubMenu = nil
	m.fillGenresMenu()
	m.gtk.genresSubMenu.ShowAll()
}
