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

const configFile = "~/.config/softteam/softimdb/config.json"

type MainWindow struct {
	builder *builder.Builder

	application                           *gtk.Application
	window                                *gtk.ApplicationWindow
	movieList                             *gtk.FlowBox
	storyLineLabel                        *gtk.Label
	storyLineScrolledWindow               *gtk.ScrolledWindow
	searchEntry                           *gtk.Entry
	searchButton                          *gtk.ToolButton
	clearSearchButton                     *gtk.ToolButton
	popupMenu                             *popupMenu
	countLabel                            *gtk.Label
	database                              *data.Database
	config                                *config.Config
	menuNoGenreItem                       *gtk.RadioMenuItem
	menuSortByName, menuSortByRating      *gtk.RadioMenuItem
	menuSortByMyRating, menuSortByLength  *gtk.RadioMenuItem
	menuSortByYear, menuSortById          *gtk.RadioMenuItem
	menuSortAscending, menuSortDescending *gtk.RadioMenuItem

	movieWin    *movieWindow
	addMovieWin *addMovieWindow

	movies map[int]*data.Movie
}

type View string

const (
	viewAll            View = "all"
	viewPacks               = "packs"
	viewToWatch             = "toWatch"
	viewNoRating            = "noRating"
	viewNeedsSubtitles      = "needsSubtitles"
)

var (
	sortBy, sortOrder = sortByName, sortAscending
	searchGenreId     = -1
	searchFor         = ""
	currentView       View
	view              viewManager
	movieTitles       []string
	showPrivateGenres = true
	genresSubMenu     *gtk.Menu
	genresMenu        *gtk.MenuItem
)

// NewMainWindow : Creates a new MainWindow object
func NewMainWindow() *MainWindow {
	m := &MainWindow{}
	m.movies = make(map[int]*data.Movie, 500)

	b, err := builder.NewBuilder(mainGlade)
	if err != nil {
		log.Fatal(err)
	}
	m.builder = b

	m.window = m.builder.GetObject("mainWindow").(*gtk.ApplicationWindow)
	m.window.SetTitle(fmt.Sprintf("%s - %s", applicationTitle, applicationVersion))
	m.window.Maximize()
	_ = m.window.Connect("destroy", m.onClose)
	_ = m.window.Connect("key-press-event", m.onKeyPressEvent)

	cnf, err := config.LoadConfig(configFile)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	m.config = cnf

	// Open database, after we have the config
	m.database = data.DatabaseNew(false, cnf)

	m.setupToolBar()
	m.popupMenu = newPopupMenu(m)
	m.popupMenu.setup()
	m.setupMenu(m.window)
	m.storyLineLabel = m.builder.GetObject("storyLineLabel").(*gtk.Label)
	m.storyLineScrolledWindow = m.builder.GetObject("storyLineScrolledWindow").(*gtk.ScrolledWindow)
	versionLabel := m.builder.GetObject("versionLabel").(*gtk.Label)
	versionLabel.SetText("Version : " + applicationVersion)
	m.countLabel = m.builder.GetObject("countLabel").(*gtk.Label)

	// Movie list
	m.movieList = m.builder.GetObject("movieList").(*gtk.FlowBox)
	m.movieList.SetSelectionMode(gtk.SELECTION_SINGLE)
	m.movieList.SetRowSpacing(listSpacing)
	m.movieList.SetColumnSpacing(listSpacing)
	m.movieList.SetMarginTop(listMargin)
	m.movieList.SetMarginBottom(listMargin)
	m.movieList.SetMarginStart(listMargin)
	m.movieList.SetMarginEnd(listMargin)
	m.movieList.SetActivateOnSingleClick(false)
	m.movieList.SetFocusOnClick(true)
	_ = m.movieList.Connect("selected-children-changed", m.onMovieListSelectionChanged)
	_ = m.movieList.Connect("child-activated", m.onMovieListDoubleClicked)

	view = newViewManager(m)

	return m
}

// Open : Opens the MainWindow window
func (m *MainWindow) Open(app *gtk.Application) {
	m.application = app
	m.window.SetApplication(app)
	m.window.ShowAll()
	view.changeView(viewToWatch)
	m.storyLineScrolledWindow.Hide()
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
	m.menuSortByName = m.builder.GetObject("menuSortByName").(*gtk.RadioMenuItem)
	m.menuSortByName.Connect(
		"activate", func() {
			if m.menuSortByName.GetActive() {
				sortBy = sortByName
				sortOrder = sortAscending
				m.menuSortAscending.SetActive(true)
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)
	m.menuSortByRating = m.builder.GetObject("menuSortByRating").(*gtk.RadioMenuItem)
	m.menuSortByRating.Connect(
		"activate", func() {
			if m.menuSortByRating.GetActive() {
				sortBy = sortByRating
				sortOrder = sortDescending
				m.menuSortDescending.SetActive(true)
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)
	m.menuSortByMyRating = m.builder.GetObject("menuSortByMyRating").(*gtk.RadioMenuItem)
	m.menuSortByMyRating.Connect(
		"activate", func() {
			if m.menuSortByMyRating.GetActive() {
				sortBy = sortByMyRating
				sortOrder = sortDescending
				m.menuSortDescending.SetActive(true)
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)
	m.menuSortByLength = m.builder.GetObject("menuSortByLength").(*gtk.RadioMenuItem)
	m.menuSortByLength.Connect(
		"activate", func() {
			if m.menuSortByLength.GetActive() {
				sortBy = sortByLength
				sortOrder = sortDescending
				m.menuSortDescending.SetActive(true)
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)
	m.menuSortByYear = m.builder.GetObject("menuSortByYear").(*gtk.RadioMenuItem)
	m.menuSortByYear.Connect(
		"activate", func() {
			if m.menuSortByYear.GetActive() {
				sortBy = sortByYear
				sortOrder = sortDescending
				m.menuSortDescending.SetActive(true)
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)
	m.menuSortById = m.builder.GetObject("menuSortById").(*gtk.RadioMenuItem)
	m.menuSortById.Connect(
		"activate", func() {
			if m.menuSortById.GetActive() {
				sortBy = sortById
				sortOrder = sortAscending
				m.menuSortAscending.SetActive(true)
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)

	m.menuSortAscending = m.builder.GetObject("menuSortAscending").(*gtk.RadioMenuItem)
	m.menuSortAscending.Connect(
		"activate", func() {
			if m.menuSortAscending.GetActive() {
				sortOrder = sortAscending
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)
	m.menuSortDescending = m.builder.GetObject("menuSortDescending").(*gtk.RadioMenuItem)
	m.menuSortDescending.Connect(
		"activate", func() {
			if m.menuSortDescending.GetActive() {
				sortOrder = sortDescending
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)

	sortBy = sortByName
	sortOrder = sortAscending

	// Genres menu
	genresMenu = m.builder.GetObject("menuGenres").(*gtk.MenuItem)
	m.fillGenresMenu()
}

func (m *MainWindow) setupToolBar() {
	// Quit button
	button := m.builder.GetObject("quitButton").(*gtk.ToolButton)
	_ = button.Connect("clicked", m.window.Close)

	// Refresh button
	button = m.builder.GetObject("refreshButton").(*gtk.ToolButton)
	_ = button.Connect("clicked", m.onRefreshButtonClicked)

	// Play button
	button = m.builder.GetObject("playMovieButton").(*gtk.ToolButton)
	_ = button.Connect("clicked", m.onPlayMovieClicked)

	// Add button
	button = m.builder.GetObject("addButton").(*gtk.ToolButton)
	_ = button.Connect("clicked", m.onOpenAddWindowClicked)

	// Search button
	m.searchButton = m.builder.GetObject("searchButton").(*gtk.ToolButton)
	_ = m.searchButton.Connect("clicked", m.onSearchButtonClicked)

	// Clear search button
	m.clearSearchButton = m.builder.GetObject("clearSearchButton").(*gtk.ToolButton)
	_ = m.clearSearchButton.Connect("clicked", m.onClearSearchButtonClicked)

	// Search entry
	m.searchEntry = m.builder.GetObject("searchEntry").(*gtk.Entry)
	_ = m.searchEntry.Connect("activate", m.onSearchButtonClicked)
}

func (m *MainWindow) fillMovieList(searchFor string, categoryId int, sortBy string) {
	movies, err := m.database.SearchMovies(string(currentView), searchFor, categoryId, sortBy)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	listHelper := &ListHelper{}
	clearFlowBox(m.movieList)
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
		m.movieList.Add(card)
	}
	m.updateCountLabel(len(movies))

}

func (m *MainWindow) getSelectedMovie() *data.Movie {
	children := m.movieList.GetSelectedChildren()
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
	id, err := strconv.Atoi(name[6:])
	if err != nil {
		return nil
	}
	return m.movies[id]
}

func (m *MainWindow) refresh(search string, categoryId int, sortBy string) {
	m.fillMovieList(search, categoryId, sortBy)
	m.movieList.ShowAll()
	if search == "" {
		m.searchEntry.SetText("")
	}
}

func (m *MainWindow) fillGenresMenu() {
	genres, _ := m.database.GetGenres()

	// Create and add genres menu
	sub, _ := gtk.MenuNew()
	genresSubMenu = sub
	genresMenu.SetSubmenu(sub)

	// No genre item
	m.menuNoGenreItem, _ = gtk.RadioMenuItemNewWithLabel(nil, "None")
	group, _ := m.menuNoGenreItem.GetGroup()
	m.menuNoGenreItem.SetActive(true)
	m.menuNoGenreItem.SetName("-1")
	sub.Add(m.menuNoGenreItem)
	m.menuNoGenreItem.Connect(
		"activate", func() {
			if m.menuNoGenreItem.GetActive() {
				m.searchGenre(m.menuNoGenreItem)
			}
		},
	)

	// Separator
	sep, _ := gtk.SeparatorMenuItemNew()
	sub.Add(sep)

	// Genre items
	for _, genre := range genres {
		if showPrivateGenres || !genre.IsPrivate {
			item, _ := gtk.RadioMenuItemNewWithLabel(group, genre.Name)
			item.SetName(strconv.Itoa(genre.Id))
			item.Connect(
				"activate", func() {
					if item.GetActive() {
						m.searchGenre(item)
					}
				},
			)
			sub.Add(item)
		}
	}
}

func (m *MainWindow) searchGenre(item *gtk.RadioMenuItem) {
	name, _ := item.GetName()
	i, _ := strconv.Atoi(name)
	searchGenreId = i
	m.refresh(searchFor, searchGenreId, getSortBy())
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
	m.countLabel.SetText(fmt.Sprintf("Number of videos : %d", i))
}

//
// Signal handlers
//

func (m *MainWindow) onClose() {
	m.database.CloseDatabase()
	m.window.Close()
	m.movieList = nil
	m.storyLineLabel = nil
	m.window = nil
	m.builder = nil
	m.movieWin = nil
	m.addMovieWin = nil
	m.application.Quit()
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

	info, err := newMovieInfoFromDatabase(selectedMovie)
	if err != nil {
		reportError(err)
		return
	}

	// Open movie dialog here
	if m.movieWin == nil {
		m.movieWin = newMovieWindow(m.builder, m.window, m.database, m.config)
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
	searchFor = ""
	searchGenreId = -1
	sortBy = sortByName
	sortOrder = sortAscending
	m.menuNoGenreItem.SetActive(true)
	m.menuSortByName.SetActive(true)
	m.menuSortAscending.SetActive(true)
	m.refresh(searchFor, searchGenreId, getSortBy())
}

func (m *MainWindow) onSearchButtonClicked() {
	search, err := m.searchEntry.GetText()
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	search = strings.Trim(search, " ")
	searchFor = search
	m.refresh(searchFor, searchGenreId, getSortBy())
}

func (m *MainWindow) onClearSearchButtonClicked() {
	searchFor = ""
	m.searchEntry.SetText("")
	m.refresh("", searchGenreId, getSortBy())
}

func (m *MainWindow) onKeyPressEvent(_ *gtk.ApplicationWindow, event *gdk.Event) {
	keyEvent := gdk.EventKeyNewFromEvent(event)

	special := (keyEvent.State() & gdk.MOD2_MASK) != 0 // Used for special keys like F5, DELETE, HOME in X11 etc
	ctrl := (keyEvent.State() & gdk.CONTROL_MASK) != 0

	if special {
		switch {
		case keyEvent.KeyVal() == gdk.KEY_F5:
			m.onRefreshButtonClicked()
		case keyEvent.KeyVal() == gdk.KEY_F6:
			m.onPlayMovieClicked()
		case keyEvent.KeyVal() == gdk.KEY_Escape:
			m.movieList.UnselectAll()
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
			m.searchEntry.GrabFocus()
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
	about.SetTransientFor(m.window)
	about.SetProgramName(applicationTitle)
	about.SetComments("An movie library application...")
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
		m.storyLineScrolledWindow.SetVisible(false)
		return
	}
	story := `<span font="Sans Regular 10" foreground="#d49c6b">` + cleanString(movie.StoryLine) + `</span>`
	m.storyLineLabel.SetMarkup(story)
	m.storyLineScrolledWindow.SetVisible(true)
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

	searchFor = "pack:" + movie.Pack
	searchGenreId = -1
	sortBy = sortByName
	sortOrder = sortAscending
	view.changeView(viewPacks)
	m.searchEntry.SetText(searchFor)
	m.menuNoGenreItem.SetActive(true)
	m.menuSortByName.SetActive(true)
	m.menuSortAscending.SetActive(true)
	m.refresh(searchFor, searchGenreId, getSortBy())
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
	m.refresh(searchFor, searchGenreId, getSortBy())

	genresSubMenu.Destroy()
	genresSubMenu = nil
	m.fillGenresMenu()
	genresSubMenu.ShowAll()
}
