package softimdb

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hultan/softimdb/internal/data"
	"github.com/hultan/softteam/framework"
	"os"
	"strconv"
	"strings"
)

type MainWindow struct {
	builder        *framework.GtkBuilder
	framework      *framework.Framework

	window         *gtk.ApplicationWindow
	aboutDialog    *gtk.AboutDialog
	addWindow      *AddWindow
	movieList      *gtk.FlowBox
	storyLineLabel *gtk.Label
	searchEntry    *gtk.Entry
	searchButton   *gtk.ToolButton

	database *data.Database

	movies map[int]*data.Movie
}

// NewMainWindow : Creates a new MainWindow object
func NewMainWindow() *MainWindow {
	mainForm := new(MainWindow)
	mainForm.movies = make(map[int]*data.Movie, 500)
	return mainForm
}

// OpenMainWindow : Opens the MainWindow window
func (m *MainWindow) OpenMainWindow(app *gtk.Application) {
	// Initialize gtk
	gtk.Init(&os.Args)

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

	// Hook up the destroy event
	_ = m.window.Connect("destroy", m.closeMainWindow)
	_ = m.window.Connect("key-press-event", m.keyPressEvent)

	// StoryLine label
	m.storyLineLabel = m.builder.GetObject("storyLineLabel").(*gtk.Label)

	// Toolbar
	m.setupToolBar()

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
	_ = m.movieList.Connect("selected-children-changed", m.selectionChanged)
	_ = m.movieList.Connect("child-activated", m.movieClicked)

	//// Status bar
	//statusBar := m.builder.getObject("main_window_status_bar").(*gtk.Statusbar)
	//statusBar.Push(statusBar.GetContextId("gtk-startup"), "gtk-startup : version 0.1.0")

	// Open database
	m.database = data.DatabaseNew(false)

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
}

func (m *MainWindow) setupMenu(window *gtk.ApplicationWindow) {
	menuQuit := m.builder.GetObject("menuFileQuit").(*gtk.MenuItem)
	_ = menuQuit.Connect("activate", window.Close)

	menuHelpAbout := m.builder.GetObject("menuHelpAbout").(*gtk.MenuItem)
	_ = menuHelpAbout.Connect("activate", m.openAboutDialog)
}

func (m *MainWindow) fillMovieList(searchFor string) {
	movies, err := m.database.GetAllMovies(searchFor)
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
	path := "smb://192.168.1.100/Videos/" + movie.MoviePath
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

	m.addWindow.OpenForm(m.builder, m.database)
}

func (m *MainWindow) refreshButtonClicked() {
	m.refresh("")
}

func (m *MainWindow) refresh(searchFor string) {
	m.fillMovieList(searchFor)
	m.movieList.ShowAll()
	if searchFor == "" {
		m.searchEntry.SetText("")
	}
}

func (m *MainWindow) searchButtonClicked() {
	searchFor, err := m.searchEntry.GetText()
	if err != nil {
		panic(err)
	}
	searchFor = strings.Trim(searchFor, " ")
	m.refresh(searchFor)
}

func (m *MainWindow) keyPressEvent(_ *gtk.ApplicationWindow, event *gdk.Event) {
	keyEvent := gdk.EventKeyNewFromEvent(event)

	ctrl := (keyEvent.State() & gdk.CONTROL_MASK) > 0

	// Catch CTRL + s
	if keyEvent.KeyVal() == gdk.KEY_s && ctrl {
		m.searchEntry.GrabFocus()
	}
}
