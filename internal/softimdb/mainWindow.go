package softimdb

import (
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hultan/softimdb/internal/data"
	"os"
	"os/exec"
	"strconv"
)

const applicationTitle = "SoftImdb"
const applicationVersion = "v 1.00"
const applicationCopyRight = "©SoftTeam AB, 2020"
const listMargin = 3
const listSpacing = 0

type MainWindow struct {
	Window         *gtk.ApplicationWindow
	builder        *SoftBuilder
	AboutDialog    *gtk.AboutDialog
	AddWindow      *gtk.Window
	MovieList      *gtk.FlowBox
	Movies         map[int]*data.Movie
	StoryLineLabel *gtk.Label
}

// NewMainWindow : Creates a new MainWindow object
func NewMainWindow() *MainWindow {
	mainForm := new(MainWindow)
	mainForm.Movies = make(map[int]*data.Movie, 500)
	return mainForm
}

// OpenMainWindow : Opens the MainWindow window
func (m *MainWindow) OpenMainWindow(app *gtk.Application) {
	// Initialize gtk
	gtk.Init(&os.Args)

	// Create a new softBuilder
	m.builder = SoftBuilderNew("main.glade")

	// Get the main window from the glade file
	m.Window = m.builder.getObject("mainWindow").(*gtk.ApplicationWindow)

	// Set up main window
	m.Window.SetApplication(app)
	m.Window.SetTitle(fmt.Sprintf("%s - %s", applicationTitle, applicationVersion))

	// Hook up the destroy event
	_ = m.Window.Connect("destroy", m.Window.Close)

	// StoryLine label
	m.StoryLineLabel = m.builder.getObject("storyLineLabel").(*gtk.Label)

	// Toolbar
	m.setupToolBar()

	// Menu
	m.setupMenu(m.Window)

	// MovieList
	m.MovieList = m.builder.getObject("movieList").(*gtk.FlowBox)
	m.MovieList.SetSelectionMode(gtk.SELECTION_SINGLE)
	m.MovieList.SetRowSpacing(listSpacing)
	m.MovieList.SetColumnSpacing(listSpacing)
	m.MovieList.SetMarginTop(listMargin)
	m.MovieList.SetMarginBottom(listMargin)
	m.MovieList.SetMarginStart(listMargin)
	m.MovieList.SetMarginEnd(listMargin)
	m.MovieList.SetActivateOnSingleClick(false)
	_ = m.MovieList.Connect("selected-children-changed", m.selectionChanged)
	_ = m.MovieList.Connect("child-activated", m.movieClicked)
	m.FillMovieList()

	//// Status bar
	//statusBar := m.builder.getObject("main_window_status_bar").(*gtk.Statusbar)
	//statusBar.Push(statusBar.GetContextId("gtk-startup"), "gtk-startup : version 0.1.0")

	// Show the main window
	m.Window.ShowAll()
}

func (m *MainWindow) setupMenu(window *gtk.ApplicationWindow) {
	menuQuit := m.builder.getObject("menuFileQuit").(*gtk.MenuItem)
	_ = menuQuit.Connect("activate", window.Close)

	menuHelpAbout := m.builder.getObject("menuHelpAbout").(*gtk.MenuItem)
	_ = menuHelpAbout.Connect("activate", m.openAboutDialog)
}

func (m *MainWindow) FillMovieList() {
	db := data.DatabaseNew(false)
	defer db.CloseDatabase()

	movies, err := db.GetAllMovies()
	if err != nil {
		panic(err)
	}

	listHelper := ListHelperNew()

	for i := range movies {
		movie := movies[i]
		m.Movies[movie.Id] = movie
		frame := listHelper.GetMovieCard(movie)
		m.MovieList.Add(frame)
		frame.SetName("frame_" + strconv.Itoa(movie.Id))
	}
}

func (m *MainWindow) selectionChanged(_ *gtk.FlowBox) {
	movie := m.getSelectedMovie()
	story := `<span font="Sans Regular 10" foreground="#d49c6b">` + cleanString(movie.StoryLine) + `</span>`
	m.StoryLineLabel.SetMarkup(story)
}

func (m *MainWindow) movieClicked(_ *gtk.FlowBox) {
	movie := m.getSelectedMovie()
	m.openMovieDirectoryInNemo(movie)
}

func (m *MainWindow) getSelectedMovie() *data.Movie {
	selected := m.MovieList.GetSelectedChildren()[0]
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
	return m.Movies[id]
}

func (m *MainWindow) openMovieDirectoryInNemo(movie *data.Movie) {
	path := "smb://192.168.1.100/Videos/" + movie.MoviePath
	command := exec.Command("nemo", path)
	_ = command.Run()
}

func (m *MainWindow) setupToolBar() {
	// Quit button
	button := m.builder.getObject("quitButton").(*gtk.ToolButton)
	_ = button.Connect("clicked", m.Window.Close)

	// Add button
	button = m.builder.getObject("addButton").(*gtk.ToolButton)
	_ = button.Connect("clicked", m.openAddWindow)
}

func (m *MainWindow) openAboutDialog() {
	if m.AboutDialog == nil {
		about := m.builder.getObject("aboutDialog").(*gtk.AboutDialog)

		about.SetDestroyWithParent(true)
		about.SetTransientFor(m.Window)
		about.SetProgramName(applicationTitle)
		about.SetComments("An application...")
		about.SetVersion(applicationVersion)
		about.SetCopyright(applicationCopyRight)

		resource := ResourcesNew()
		image, err := gdk.PixbufNewFromFile(resource.GetResourcePath("application.png"))
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

		m.AboutDialog = about
	}

	m.AboutDialog.Present()
}

func (m *MainWindow) openAddWindow() {
	addWindow := AddWindowNew()
	addWindow.OpenForm(m.builder)
}
