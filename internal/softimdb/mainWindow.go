package softimdb

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/builder"
	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
)

//go:embed assets/application.png
var applicationIcon []byte

//go:embed assets/main.glade
var mainGlade string

const configFile = "/home/per/.config/softteam/softimdb/config.json"

type MainWindow struct {
	builder *builder.Builder

	application    *gtk.Application
	window         *gtk.ApplicationWindow
	aboutDialog    *gtk.AboutDialog
	addWindow      *AddWindow
	movieList      *gtk.FlowBox
	storyLineLabel *gtk.Label
	searchEntry    *gtk.Entry
	searchButton   *gtk.ToolButton
	popupMenu      *PopupMenu
	countLabel     *gtk.Label
	movieWindow    *MovieWindow
	database       *data.Database
	config         *config.Config

	movies map[int]*data.Movie
}

var sortBy, sortOrder, selectedGenreId int
var searchFor string
var noneItem, menuSortByName, menuSortAscending *gtk.RadioMenuItem

// NewMainWindow : Creates a new MainWindow object
func NewMainWindow() *MainWindow {
	mainForm := new(MainWindow)
	mainForm.movies = make(map[int]*data.Movie, 500)
	return mainForm
}

// OpenMainWindow : Opens the MainWindow window
func (m *MainWindow) OpenMainWindow(app *gtk.Application) {
	m.application = app

	// Create a new softBuilder
	m.builder = builder.NewBuilder(mainGlade)

	// Get the main window from the glade file
	m.window = m.builder.GetObject("mainWindow").(*gtk.ApplicationWindow)

	// Set up main window
	m.window.SetApplication(app)
	m.window.SetTitle(fmt.Sprintf("%s - %s", applicationTitle, applicationVersion))
	m.window.Maximize()

	// Hook up the destroy event
	_ = m.window.Connect("destroy", m.closeMainWindow)
	_ = m.window.Connect("key-press-event", m.keyPressEvent)

	// StoryLine label
	m.storyLineLabel = m.builder.GetObject("storyLineLabel").(*gtk.Label)

	// Load config file
	cnf, err := config.LoadConfig(configFile)
	if err != nil {
		reportError(err)
		panic(err)
	}
	m.config = cnf

	// Open database
	m.database = data.DatabaseNew(false, cnf)

	// Toolbar
	m.setupToolBar()

	// Popup menu
	m.popupMenu = NewPopupMenu(m)
	m.popupMenu.Setup()

	// Menu
	m.setupMenu(m.window)

	// MovieList
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
	_ = m.movieList.Connect("selected-children-changed", m.selectionChanged)
	_ = m.movieList.Connect("child-activated", m.movieClicked)

	// Status bar
	versionLabel := m.builder.GetObject("versionLabel").(*gtk.Label)
	versionLabel.SetText("Version : " + applicationVersion)
	m.countLabel = m.builder.GetObject("countLabel").(*gtk.Label)

	// Fill movie list box
	m.refreshButtonClicked()

	// Show the main window
	m.window.ShowAll()
}

func (m *MainWindow) closeMainWindow() {
	m.database.CloseDatabase()
	m.window.Close()
	if m.addWindow != nil {
		m.addWindow.window.Close()
	}
	if m.aboutDialog != nil {
		m.aboutDialog.Close()
	}

	m.movieList = nil
	m.storyLineLabel = nil
	m.addWindow = nil
	m.aboutDialog = nil
	m.window = nil
	m.builder = nil

	m.application.Quit()
}

func (m *MainWindow) setupMenu(window *gtk.ApplicationWindow) {
	// File menu
	menuQuit := m.builder.GetObject("menuFileQuit").(*gtk.MenuItem)
	_ = menuQuit.Connect("activate", window.Close)

	// Help menu
	menuHelpAbout := m.builder.GetObject("menuHelpAbout").(*gtk.MenuItem)
	_ = menuHelpAbout.Connect("activate", m.openAboutDialog)

	// Sort menu
	menuSortByName = m.builder.GetObject("menuSortByName").(*gtk.RadioMenuItem)
	menuSortByRating := m.builder.GetObject("menuSortByRating").(*gtk.RadioMenuItem)
	menuSortByYear := m.builder.GetObject("menuSortByYear").(*gtk.RadioMenuItem)
	menuSortById := m.builder.GetObject("menuSortById").(*gtk.RadioMenuItem)
	menuSortByName.Connect(
		"activate", func() {
			if menuSortByName.GetActive() {
				fmt.Println("Sort by name")
				sortBy = sortByName
				m.refresh(searchFor, selectedGenreId, m.getSortBy())
			}
		},
	)
	menuSortByRating.Connect(
		"activate", func() {
			if menuSortByRating.GetActive() {
				fmt.Println("Sort by rating")
				sortBy = sortByRating
				m.refresh(searchFor, selectedGenreId, m.getSortBy())
			}
		},
	)
	menuSortByYear.Connect(
		"activate", func() {
			if menuSortByYear.GetActive() {
				fmt.Println("Sort by year")
				sortBy = sortByYear
				m.refresh(searchFor, selectedGenreId, m.getSortBy())
			}
		},
	)
	menuSortById.Connect(
		"activate", func() {
			if menuSortById.GetActive() {
				fmt.Println("Sort by id")
				sortBy = sortById
				m.refresh(searchFor, selectedGenreId, m.getSortBy())
			}
		},
	)

	menuSortAscending = m.builder.GetObject("menuSortAscending").(*gtk.RadioMenuItem)
	menuSortDescending := m.builder.GetObject("menuSortDescending").(*gtk.RadioMenuItem)
	menuSortAscending.Connect(
		"activate", func() {
			if menuSortAscending.GetActive() {
				fmt.Println("Sort ascending")
				sortOrder = sortAscending
				m.refresh(searchFor, selectedGenreId, m.getSortBy())
			}
		},
	)
	menuSortDescending.Connect(
		"activate", func() {
			if menuSortDescending.GetActive() {
				fmt.Println("Sort descending")
				sortOrder = sortDescending
				m.refresh(searchFor, selectedGenreId, m.getSortBy())
			}
		},
	)

	sortBy = sortByName
	sortOrder = sortAscending

	// Tags menu
	menuTags := m.builder.GetObject("menuTags").(*gtk.MenuItem)
	m.fillTagsMenu(menuTags)

	// Tools menu
	menuToolsOpenIMDB := m.builder.GetObject("mnuToolsIOpenIMDB").(*gtk.MenuItem)
	_ = menuToolsOpenIMDB.Connect("activate", m.openIMDB)
}

func (m *MainWindow) fillMovieList(searchFor string, categoryId int, sortBy string) {
	movies, err := m.database.GetAllMovies(searchFor, categoryId, sortBy)
	if err != nil {
		reportError(err)
		panic(err)
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

func (m *MainWindow) selectionChanged(_ *gtk.FlowBox) {
	movie := m.getSelectedMovie()
	if movie == nil {
		return
	}
	story := `<span font="Sans Regular 10" foreground="#d49c6b">` + cleanString(movie.StoryLine) + `</span>`
	m.storyLineLabel.SetMarkup(story)
}

func (m *MainWindow) movieClicked(_ *gtk.FlowBox) {
	movie := m.getSelectedMovie()
	if movie == nil {
		return
	}
	m.editMovieInfo()
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

func (m *MainWindow) openMovieDirectoryInNemo(movie *data.Movie) {
	openInNemo(fmt.Sprintf("%s/%s", m.config.RootDir, movie.MoviePath))
}

func (m *MainWindow) setupToolBar() {
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
			m.refresh("", -1, m.getSortBy())
		},
	)

	sortByIdButton := m.builder.GetObject("sortById").(*gtk.ToolButton)
	_ = sortByIdButton.Connect(
		"clicked", func() {
			sortBy = sortById
			sortOrder = sortDescending
			m.refresh("", -1, m.getSortBy())
		},
	)
}

func (m *MainWindow) openAboutDialog() {
	if m.aboutDialog == nil {
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
		about.SetPosition(gtk.WIN_POS_CENTER)

		_ = about.Connect(
			"response", func(dialog *gtk.AboutDialog, responseId gtk.ResponseType) {
				if responseId == gtk.RESPONSE_CANCEL || responseId == gtk.RESPONSE_DELETE_EVENT {
					about.Hide()
				}
			},
		)

		m.aboutDialog = about
	}

	m.aboutDialog.ShowAll()
}

func (m *MainWindow) openAddWindowClicked() {
	if m.addWindow == nil {
		m.addWindow = AddWindowNew()
	}

	m.addWindow.OpenForm(m.builder, m.database, m.config)
}

func (m *MainWindow) refreshButtonClicked() {
	searchFor = ""
	selectedGenreId = -1
	sortBy = sortByName
	sortOrder = sortAscending
	noneItem.SetActive(true)
	menuSortByName.SetActive(true)
	menuSortAscending.SetActive(true)
	m.refresh(searchFor, selectedGenreId, m.getSortBy())
}

func (m *MainWindow) refresh(search string, categoryId int, sortBy string) {
	m.fillMovieList(search, categoryId, sortBy)
	m.movieList.ShowAll()
	if search == "" {
		m.searchEntry.SetText("")
	}
}

func (m *MainWindow) searchButtonClicked() {
	search, err := m.searchEntry.GetText()
	if err != nil {
		reportError(err)
		panic(err)
	}
	search = strings.Trim(search, " ")
	searchFor = search
	m.refresh(searchFor, selectedGenreId, m.getSortBy())
}

func (m *MainWindow) keyPressEvent(_ *gtk.ApplicationWindow, event *gdk.Event) {
	keyEvent := gdk.EventKeyNewFromEvent(event)

	ctrl := (keyEvent.State() & gdk.CONTROL_MASK) > 0

	// Catch CTRL + f
	switch {
	case keyEvent.KeyVal() == gdk.KEY_f && ctrl:
		m.searchEntry.GrabFocus()
	case keyEvent.KeyVal() == gdk.KEY_a && ctrl:
		m.openAddWindowClicked()
	}
}

func (m *MainWindow) getSortBy() string {
	var sort = ""

	switch sortBy {
	case sortByName:
		sort = "title"
	case sortByRating:
		sort = "imdb_rating"
	case sortByYear:
		sort = "year"
	case sortById:
		sort = "id"
	}

	switch sortOrder {
	case sortAscending:
		sort += " asc"
	case sortDescending:
		sort += " desc"
	}

	return sort
}

func (m *MainWindow) fillTagsMenu(menu *gtk.MenuItem) {
	tags, _ := m.database.GetTags()
	sub, _ := gtk.MenuNew()
	menu.SetSubmenu(sub)
	noneItem, _ = gtk.RadioMenuItemNewWithLabel(nil, "None")
	group, _ := noneItem.GetGroup()
	noneItem.SetActive(true)
	noneItem.SetName("-1")
	sub.Add(noneItem)
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
					selectedGenreId = i
					m.refresh(searchFor, selectedGenreId, m.getSortBy())
				}
			},
		)
		sub.Add(item)
	}
	noneItem.Connect(
		"activate", func() {
			if noneItem.GetActive() {
				name, _ := noneItem.GetName()
				i, _ := strconv.Atoi(name)
				selectedGenreId = i
				m.refresh(searchFor, selectedGenreId, m.getSortBy())
			}
		},
	)
}

func (m *MainWindow) playMovie(movie *data.Movie) {
	go func() {
		moviePath := fmt.Sprintf("%s/%s", m.config.RootDir, movie.MoviePath)
		movieName, err := m.getMovieName(moviePath)
		if err != nil {
			reportError(err)
			return
		}
		if movieName == "" {
			m.openMovieDirectoryInNemo(movie)
			return
		}
		moviePath = path.Join(moviePath, movieName)
		openProcess("smplayer", moviePath)
	}()
}

func (m *MainWindow) getMovieName(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	files, err := f.Readdirnames(0)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if strings.HasSuffix(file, "mkv") {
			return file, nil
		}
	}
	return "", nil
}

func (m *MainWindow) executeCommand(command string, arguments ...string) (string, error) {
	cmd := exec.Command(command, arguments...)

	// set the output to our variable
	out, err := cmd.Output()
	if err != nil {
		reportError(err)
		return "", err
	}

	return string(out), nil
}

func (m *MainWindow) editMovieInfo() {
	selectedMovie := m.getSelectedMovie()
	if selectedMovie == nil {
		return
	}

	movieInfo, err := newMovieInfoFromDatabase(selectedMovie)
	if err != nil {
		reportError(err)
		return
	}

	// Open movie dialog here
	win := NewMovieWindow(movieInfo, selectedMovie, m.saveMovieInfo)
	win.OpenForm(m.builder, m.window)
	m.movieWindow = win
}

func (m *MainWindow) saveMovieInfo(movieInfo *MovieInfo, movie *data.Movie) {
	movieInfo.toDatabase(movie)

	err := m.database.UpdateMovie(movie, true)
	if err != nil {
		reportError(err)
		return
	}

	if movieInfo.imageHasChanged {
		image, err := m.database.GetImage(movie.ImageId)
		if err != nil {
			reportError(err)
			return
		}
		image.Data = movieInfo.image
		err = m.database.UpdateImage(image)
		if err != nil {
			reportError(err)
			return
		}
	}

	m.movieWindow.window.Destroy()
	m.movieWindow = nil
}

func (m *MainWindow) openIMDB() {
	v := m.getSelectedMovie()
	if v == nil {
		return
	}

	m.openBrowser(v.ImdbUrl)
}

// https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func (m *MainWindow) openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func (m *MainWindow) updateCountLabel(i int) {
	m.countLabel.SetText(fmt.Sprintf("Number of videos : %d", i))
}

func (m *MainWindow) getTags(tags []string) []data.Tag {
	var dataTags []data.Tag

	for _, tag := range tags {
		dataTag := data.Tag{Name: tag}
		dataTags = append(dataTags, dataTag)
	}

	return dataTags
}

// ClearFlowBox : Clears a gtk.FlowBox
func clearFlowBox(list *gtk.FlowBox) {
	children := list.GetChildren()
	if children == nil {
		return
	}
	var i uint = 0
	for i < children.Length() {
		widget, _ := children.NthData(i).(*gtk.Widget)
		list.Remove(widget)
		i++
	}
}

// openInNemo : Opens a new nemo instance with the specified folder opened
func openInNemo(path string) {
	openProcess("nemo", path)
}

// openProcess : Opens a command with the specified arguments
func openProcess(command string, args ...string) string {

	cmd := exec.Command(command, args...)
	// Forces the new process to detach from the GitDiscover process
	// so that it does not die when GitDiscover dies
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	err = cmd.Process.Release()
	if err != nil {
		log.Println(err)
	}

	return string(output)
}

func getIdFromUrl(url string) (string, error) {
	// Get the IMDB id from the URL.
	// Starts with tt and ends with 7 or 8 digits.
	re := regexp.MustCompile(`tt\d{7,8}`)
	matches := re.FindAll([]byte(url), -1)
	if len(matches) == 0 {
		err := errors.New("invalid imdb URL")
		return "", err
	}
	return string(matches[0]), nil
}
