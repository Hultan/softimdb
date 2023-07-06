package main

import (
	"fmt"
	"os"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/softimdb/internal/softimdb"
)

const (
	ApplicationId    = "se.softteam.softimdb"
	ApplicationFlags = glib.APPLICATION_FLAGS_NONE
)

func main() {
	// Initialize gtk
	gtk.Init(&os.Args)

	// Create a new application
	application, err := gtk.ApplicationNew(ApplicationId, ApplicationFlags)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create GTK Application : %v", err)
	}

	mainForm := softimdb.NewMainWindow()
	// Hook up the activate event handler
	_ = application.Connect("activate", mainForm.Open)

	// Start the application (and exit when it is done)
	os.Exit(application.Run(nil))
}
