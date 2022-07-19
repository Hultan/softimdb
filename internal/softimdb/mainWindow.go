package softimdb

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
	imdb2 "github.com/hultan/softimdb/internal/imdb"
	"github.com/hultan/softteam/framework"
)

const configFile = "/home/per/.config/softteam/softimdb/config.json"

type MainWindow struct {
	builder   *framework.GtkBuilder
	framework *framework.Framework

	application    *gtk.Application
	window         *gtk.ApplicationWindow
	aboutDialog    *gtk.AboutDialog
	addWindow      *AddWindow
	movieList      *gtk.FlowBox
	storyLineLabel *gtk.Label
	searchEntry    *gtk.Entry
	searchButton   *gtk.ToolButton
	popupMenu      *PopupMenu

	database *data.Database
	config   *config.Config

	movies map[int]*data.Movie
}

var sortBy, sortOrder, tagIndex int
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
	fw := framework.NewFramework()
	m.framework = fw
	builder, err := fw.Gtk.CreateBuilder("main.glade")
	if err != nil {
		panic(err)
	}
	m.builder = builder

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

	// Open database
	m.database = data.DatabaseNew(false)

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

	// // Status bar
	// statusBar := m.builder.getObject("main_window_status_bar").(*gtk.Statusbar)
	// statusBar.Push(statusBar.GetContextId("gtk-startup"), "gtk-startup : version 0.1.0")

	// Fill movie list box
	m.refreshButtonClicked()

	// Load config file
	cnf, err := config.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}
	m.config = cnf

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
	menuSortByName.Connect("activate", func() {
		if menuSortByName.GetActive() {
			fmt.Println("Sort by name")
			sortBy = sortByName
			m.refresh(searchFor, tagIndex, m.getSortBy())
		}
	})
	menuSortByRating.Connect("activate", func() {
		if menuSortByRating.GetActive() {
			fmt.Println("Sort by rating")
			sortBy = sortByRating
			m.refresh(searchFor, tagIndex, m.getSortBy())
		}
	})
	menuSortByYear.Connect("activate", func() {
		if menuSortByYear.GetActive() {
			fmt.Println("Sort by year")
			sortBy = sortByYear
			m.refresh(searchFor, tagIndex, m.getSortBy())
		}
	})

	menuSortAscending = m.builder.GetObject("menuSortAscending").(*gtk.RadioMenuItem)
	menuSortDescending := m.builder.GetObject("menuSortDescending").(*gtk.RadioMenuItem)
	menuSortAscending.Connect("activate", func() {
		if menuSortAscending.GetActive() {
			fmt.Println("Sort ascending")
			sortOrder = sortAscending
			m.refresh(searchFor, tagIndex, m.getSortBy())
		}
	})
	menuSortDescending.Connect("activate", func() {
		if menuSortDescending.GetActive() {
			fmt.Println("Sort descending")
			sortOrder = sortDescending
			m.refresh(searchFor, tagIndex, m.getSortBy())
		}
	})

	sortBy = sortByName
	sortOrder = sortAscending

	// Tags menu
	menuTags := m.builder.GetObject("menuTags").(*gtk.MenuItem)
	m.fillTagsMenu(menuTags)

	// Tools menu
	menuToolsRefresh := m.builder.GetObject("mnuToolsRefreshIMDBData").(*gtk.MenuItem)
	_ = menuToolsRefresh.Connect("activate", m.refreshIMDB)
	menuToolsOpenIMDB := m.builder.GetObject("mnuToolsIOpenIMDB").(*gtk.MenuItem)
	_ = menuToolsOpenIMDB.Connect("activate", m.openIMDB)

}

func (m *MainWindow) fillMovieList(searchFor string, categoryId int, sortBy string) {
	movies, err := m.database.GetAllMovies(searchFor, categoryId, sortBy)
	if err != nil {
		panic(err)
	}

	listHelper := ListHelperNew()
	m.framework.Gtk.ClearFlowBox(m.movieList)

	for i := range movies {
		movie := movies[i]
		m.movies[movie.Id] = movie
		frame := listHelper.GetMovieCard(movie)
		m.movieList.Add(frame)
		frame.SetName("frame_" + strconv.Itoa(movie.Id))
	}
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
	m.openMovieDirectoryInNemo(movie)
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
	path := fmt.Sprintf("smb://%s/%s/%s", m.config.Nas, m.config.Folder, movie.MoviePath)
	// path := "smb://192.168.1.100/Videos/" + movie.MoviePath
	m.framework.Process.OpenInNemo(path)
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

		image, err := gdk.PixbufNewFromFile(m.framework.Resource.GetResourcePath("application.png"))
		if err == nil {
			about.SetLogo(image)
		}

		about.SetModal(true)
		about.SetPosition(gtk.WIN_POS_CENTER)

		_ = about.Connect("response", func(dialog *gtk.AboutDialog, responseId gtk.ResponseType) {
			if responseId == gtk.RESPONSE_CANCEL || responseId == gtk.RESPONSE_DELETE_EVENT {
				about.Hide()
			}
		})

		m.aboutDialog = about
	}

	m.aboutDialog.ShowAll()
}

func (m *MainWindow) openAddWindowClicked() {
	if m.addWindow == nil {
		m.addWindow = AddWindowNew(m.framework)
	}

	m.addWindow.OpenForm(m.builder, m.database, m.config)
}

func (m *MainWindow) refreshButtonClicked() {
	searchFor = ""
	tagIndex = -1
	sortBy = sortByName
	sortOrder = sortAscending
	noneItem.SetActive(true)
	menuSortByName.SetActive(true)
	menuSortAscending.SetActive(true)
	m.refresh(searchFor, tagIndex, m.getSortBy())
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
		panic(err)
	}
	search = strings.Trim(search, " ")
	searchFor = search
	m.refresh(searchFor, tagIndex, m.getSortBy())
}

func (m *MainWindow) keyPressEvent(_ *gtk.ApplicationWindow, event *gdk.Event) {
	keyEvent := gdk.EventKeyNewFromEvent(event)

	ctrl := (keyEvent.State() & gdk.CONTROL_MASK) > 0

	// Catch CTRL + f
	switch {
	case keyEvent.KeyVal() == gdk.KEY_f && ctrl:
		m.searchEntry.GrabFocus()
	case keyEvent.KeyVal() == gdk.KEY_F5 && ctrl:
		m.refreshIMDB()
	}
	if keyEvent.KeyVal() == gdk.KEY_f && ctrl {

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
		item.Connect("activate", func() {
			if item.GetActive() {
				name, _ := item.GetName()
				i, _ := strconv.Atoi(name)
				tagIndex = i
				m.refresh(searchFor, tagIndex, m.getSortBy())
			}
		})
		sub.Add(item)
	}
	noneItem.Connect("activate", func() {
		if noneItem.GetActive() {
			name, _ := noneItem.GetName()
			i, _ := strconv.Atoi(name)
			tagIndex = i
			m.refresh(searchFor, tagIndex, m.getSortBy())
		}
	})
}

func (m *MainWindow) playMovie(movie *data.Movie) {
	go func() {
		path := fmt.Sprintf("smb://%s/%s/%s", m.config.Nas, m.config.Folder, movie.MoviePath)
		cmd := fmt.Sprintf("find %s -type f -exec du -h {} + | sort -r | head -n1", path)
		file, err := m.executeCommand("bash", "-c", cmd)
		if err != nil {
			panic(err)
		}
		m.framework.Process.Open("smplayer", file)
	}()
}

func (m *MainWindow) executeCommand(command string, arguments ...string) (string, error) {
	cmd := exec.Command(command, arguments...)

	// set the output to our variable
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(out), nil
}

func (m *MainWindow) refreshIMDB() {
	v := m.getSelectedMovie()
	if v == nil {
		return
	}

	imdb := &imdb2.Manager{}
	err := imdb.GetMovieInfo(v)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = m.database.UpdateMovie(v)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (m *MainWindow) openIMDB() {
	v := m.getSelectedMovie()
	if v == nil {
		return
	}

	openbrowser(v.ImdbUrl)
}

// https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func openbrowser(url string) {
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
