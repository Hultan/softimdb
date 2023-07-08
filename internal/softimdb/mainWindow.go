package softimdb

import (
	_ "embed"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hultan/dialog"
	"log"
	"path"
	"strconv"
	"strings"

	"github.com/hultan/softimdb/internal/builder"
	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
)

//go:embed assets/application.png
var applicationIcon []byte

//go:embed assets/main.glade
var mainGlade string

const configFile = "/home/per/.config/softteam/softimdb/config.json"

type mainWindow struct {
	builder *builder.Builder

	application                           *gtk.Application
	window                                *gtk.ApplicationWindow
	movieList                             *gtk.FlowBox
	storyLineLabel                        *gtk.Label
	searchEntry                           *gtk.Entry
	searchButton                          *gtk.ToolButton
	popupMenu                             *popupMenu
	countLabel                            *gtk.Label
	database                              *data.Database
	config                                *config.Config
	menuNoTagItem                         *gtk.RadioMenuItem
	menuSortByName, menuSortByRating      *gtk.RadioMenuItem
	menuSortByYear, menuSortById          *gtk.RadioMenuItem
	menuSortAscending, menuSortDescending *gtk.RadioMenuItem

	movieWin    *movieWindow
	addMovieWin *addMovieWindow

	movies map[int]*data.Movie
}

var sortBy, sortOrder string
var searchGenreId int
var searchFor string

// NewMainWindow : Creates a new mainWindow object
func NewMainWindow() *mainWindow {
	mainForm := new(mainWindow)
	mainForm.movies = make(map[int]*data.Movie, 500)
	return mainForm
}

// Open : Opens the mainWindow window
func (m *mainWindow) Open(app *gtk.Application) {
	m.application = app

	// Create a new softBuilder
	b, err := builder.NewBuilder(mainGlade)
	if err != nil {
		log.Fatal(err)
	}
	m.builder = b

	// Get the main window from the glade file
	m.window = m.builder.GetObject("mainWindow").(*gtk.ApplicationWindow)

	// Set up main window
	m.window.SetApplication(app)
	m.window.SetTitle(fmt.Sprintf("%s - %s", applicationTitle, applicationVersion))
	m.window.Maximize()

	// Hook up the destroy event
	_ = m.window.Connect("destroy", m.closeMainWindow)
	_ = m.window.Connect("key-press-event", m.onKeyPressEvent)

	// Load config file
	cnf, err := config.LoadConfig(configFile)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	m.config = cnf

	// Open database
	m.database = data.DatabaseNew(false, cnf)

	// Misc GTK
	m.setupToolBar()
	m.popupMenu = newPopupMenu(m)
	m.popupMenu.setup()
	m.setupMenu(m.window)
	m.storyLineLabel = m.builder.GetObject("storyLineLabel").(*gtk.Label)
	versionLabel := m.builder.GetObject("versionLabel").(*gtk.Label)
	versionLabel.SetText("Version : " + applicationVersion)
	m.countLabel = m.builder.GetObject("countLabel").(*gtk.Label)

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

	m.refreshButtonClicked()

	m.window.ShowAll()
}

func (m *mainWindow) closeMainWindow() {
	m.database.CloseDatabase()
	m.window.Close()
	m.movieList = nil
	m.storyLineLabel = nil
	m.window = nil
	m.builder = nil
	m.application.Quit()
}

func (m *mainWindow) setupMenu(window *gtk.ApplicationWindow) {
	// File menu
	menuQuit := m.builder.GetObject("menuFileQuit").(*gtk.MenuItem)
	_ = menuQuit.Connect("activate", window.Close)

	// Help menu
	menuHelpAbout := m.builder.GetObject("menuHelpAbout").(*gtk.MenuItem)
	_ = menuHelpAbout.Connect("activate", m.openAboutDialog)

	// Sort menu
	m.menuSortByName = m.builder.GetObject("menuSortByName").(*gtk.RadioMenuItem)
	m.menuSortByName.Connect(
		"activate", func() {
			if m.menuSortByName.GetActive() {
				sortBy = sortByName
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)
	m.menuSortByRating = m.builder.GetObject("menuSortByRating").(*gtk.RadioMenuItem)
	m.menuSortByRating.Connect(
		"activate", func() {
			if m.menuSortByRating.GetActive() {
				sortBy = sortByRating
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)
	m.menuSortByYear = m.builder.GetObject("menuSortByYear").(*gtk.RadioMenuItem)
	m.menuSortByYear.Connect(
		"activate", func() {
			if m.menuSortByYear.GetActive() {
				sortBy = sortByYear
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)
	m.menuSortById = m.builder.GetObject("menuSortById").(*gtk.RadioMenuItem)
	m.menuSortById.Connect(
		"activate", func() {
			if m.menuSortById.GetActive() {
				sortBy = sortById
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

	// Tags menu
	menuTags := m.builder.GetObject("menuTags").(*gtk.MenuItem)
	m.fillTagsMenu(menuTags)
}

func (m *mainWindow) fillMovieList(searchFor string, categoryId int, sortBy string) {
	movies, err := m.database.GetAllMovies(searchFor, categoryId, sortBy)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	listHelper := ListHelperNew()
	clearFlowBox(m.movieList)

	for i := range movies {
		movie := movies[i]
		m.movies[movie.Id] = movie
		frame := listHelper.CreateMovieCard(movie)
		m.movieList.Add(frame)
		frame.SetName("frame_" + strconv.Itoa(movie.Id))
	}

	m.updateCountLabel(len(movies))
}

func (m *mainWindow) onMovieListSelectionChanged(_ *gtk.FlowBox) {
	movie := m.getSelectedMovie()
	if movie == nil {
		return
	}
	story := `<span font="Sans Regular 10" foreground="#d49c6b">` + cleanString(movie.StoryLine) + `</span>`
	m.storyLineLabel.SetMarkup(story)
}

func (m *mainWindow) onMovieListDoubleClicked(_ *gtk.FlowBox) {
	movie := m.getSelectedMovie()
	if movie == nil {
		return
	}
	m.editMovieInfo()
}

func (m *mainWindow) getSelectedMovie() *data.Movie {
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

func (m *mainWindow) setupToolBar() {
	// Quit button
	button := m.builder.GetObject("quitButton").(*gtk.ToolButton)
	_ = button.Connect("clicked", m.window.Close)

	// Refresh button
	button = m.builder.GetObject("refreshButton").(*gtk.ToolButton)
	_ = button.Connect("clicked", m.refreshButtonClicked)

	// Add button
	button = m.builder.GetObject("addButton").(*gtk.ToolButton)
	_ = button.Connect("clicked", m.openAddWindowClicked)

	// Search button
	m.searchButton = m.builder.GetObject("searchButton").(*gtk.ToolButton)
	_ = m.searchButton.Connect("clicked", m.searchButtonClicked)

	// Search entry
	m.searchEntry = m.builder.GetObject("searchEntry").(*gtk.Entry)
	_ = m.searchEntry.Connect("activate", m.searchButtonClicked)

	// Sort by buttons
	sortByNameButton := m.builder.GetObject("sortByName").(*gtk.ToolButton)
	_ = sortByNameButton.Connect(
		"clicked", func() {
			sortBy = sortByName
			sortOrder = sortAscending
			m.refresh("", -1, getSortBy())
		},
	)

	sortByIdButton := m.builder.GetObject("sortById").(*gtk.ToolButton)
	_ = sortByIdButton.Connect(
		"clicked", func() {
			sortBy = sortById
			sortOrder = sortDescending
			m.refresh("", -1, getSortBy())
		},
	)
}

func (m *mainWindow) openAboutDialog() {
	about := m.builder.GetObject("aboutDialog").(*gtk.AboutDialog)

	about.SetDestroyWithParent(true)
	about.SetTransientFor(m.window)
	about.SetProgramName(applicationTitle)
	about.SetComments("An application...")
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

func (m *mainWindow) openAddWindowClicked() {
	if m.addMovieWin == nil {
		m.addMovieWin = newAddMovieWindow(m, m.database, m.config)
	}
	m.addMovieWin.openForm()
}

func (m *mainWindow) refreshButtonClicked() {
	searchFor = ""
	searchGenreId = -1
	sortBy = sortByName
	sortOrder = sortAscending
	m.menuNoTagItem.SetActive(true)
	m.menuSortByName.SetActive(true)
	m.menuSortAscending.SetActive(true)
	m.refresh(searchFor, searchGenreId, getSortBy())
}

func (m *mainWindow) refresh(search string, categoryId int, sortBy string) {
	m.fillMovieList(search, categoryId, sortBy)
	m.movieList.ShowAll()
	if search == "" {
		m.searchEntry.SetText("")
	}
}

func (m *mainWindow) searchButtonClicked() {
	search, err := m.searchEntry.GetText()
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	search = strings.Trim(search, " ")
	searchFor = search
	m.refresh(searchFor, searchGenreId, getSortBy())
}

func (m *mainWindow) onKeyPressEvent(_ *gtk.ApplicationWindow, event *gdk.Event) {
	keyEvent := gdk.EventKeyNewFromEvent(event)

	ctrl := (keyEvent.State() & gdk.CONTROL_MASK) > 0

	switch {
	case keyEvent.KeyVal() == gdk.KEY_f && ctrl:
		m.searchEntry.GrabFocus()
	case keyEvent.KeyVal() == gdk.KEY_a && ctrl:
		m.openAddWindowClicked()
	case (keyEvent.KeyVal() == gdk.KEY_q || keyEvent.KeyVal() == gdk.KEY_Q) && ctrl:
		m.closeMainWindow()
	}
}

func (m *mainWindow) fillTagsMenu(menu *gtk.MenuItem) {
	tags, _ := m.database.GetTags()
	sub, _ := gtk.MenuNew()
	menu.SetSubmenu(sub)
	m.menuNoTagItem, _ = gtk.RadioMenuItemNewWithLabel(nil, "None")
	group, _ := m.menuNoTagItem.GetGroup()
	m.menuNoTagItem.SetActive(true)
	m.menuNoTagItem.SetName("-1")
	sub.Add(m.menuNoTagItem)
	sep, _ := gtk.SeparatorMenuItemNew()
	sub.Add(sep)

	for _, tag := range tags {
		item, _ := gtk.RadioMenuItemNewWithLabel(group, tag.Name)
		item.SetName(strconv.Itoa(tag.Id))
		item.Connect(
			"activate", func() {
				if item.GetActive() {
					name, _ := item.GetName()
					i, _ := strconv.Atoi(name)
					searchGenreId = i
					m.refresh(searchFor, searchGenreId, getSortBy())
				}
			},
		)
		sub.Add(item)
	}
	m.menuNoTagItem.Connect(
		"activate", func() {
			if m.menuNoTagItem.GetActive() {
				name, _ := m.menuNoTagItem.GetName()
				i, _ := strconv.Atoi(name)
				searchGenreId = i
				m.refresh(searchFor, searchGenreId, getSortBy())
			}
		},
	)
}

func (m *mainWindow) playMovie(movie *data.Movie) {
	go func() {
		moviePath := fmt.Sprintf("%s/%s", m.config.RootDir, movie.MoviePath)
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

func (m *mainWindow) editMovieInfo() {
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
		m.movieWin = newMovieWindow(m.builder, m.window)
	}

	m.movieWin.open(info, selectedMovie, m.saveMovieInfo)
}

func (m *mainWindow) saveMovieInfo(movieInfo *movieInfo, movie *data.Movie) {
	movieInfo.toDatabase(movie)

	err := m.database.UpdateMovie(movie, true)
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
}

func (m *mainWindow) openIMDB() {
	v := m.getSelectedMovie()
	if v == nil {
		return
	}

	openBrowser(v.ImdbUrl)
}

func (m *mainWindow) updateCountLabel(i int) {
	m.countLabel.SetText(fmt.Sprintf("Number of videos : %d", i))
}
