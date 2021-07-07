package main

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hultan/softimdb/internal/softimdb"
	"os"
)

const (
	ApplicationId    = "se.softteam.softimdb"
	ApplicationFlags = glib.APPLICATION_FLAGS_NONE
)

func main() {
	// Create a new application
	application, err := gtk.ApplicationNew(ApplicationId, ApplicationFlags)
	softimdb.ErrorCheckWithPanic(err, "Failed to create GTK Application")

	mainForm := softimdb.NewMainWindow()
	// Hook up the activate event handler
	_ = application.Connect("activate", mainForm.OpenMainWindow)

	// Start the application (and exit when it is done)
	os.Exit(application.Run(nil))
}
