package softimdb

import "github.com/gotk3/gotk3/gtk"

type viewManager struct {
	viewAllButton, viewPacksButton        *gtk.ToggleToolButton
	viewToWatchButton, viewNoRatingButton *gtk.ToggleToolButton
	viewNeedsSubtitlesButton              *gtk.ToggleToolButton
}

func newViewManager(m *MainWindow) viewManager {
	w := viewManager{}

	w.viewAllButton = m.builder.GetObject("viewAll").(*gtk.ToggleToolButton)
	w.viewPacksButton = m.builder.GetObject("viewPacks").(*gtk.ToggleToolButton)
	w.viewToWatchButton = m.builder.GetObject("viewToWatch").(*gtk.ToggleToolButton)
	w.viewNoRatingButton = m.builder.GetObject("viewNoRating").(*gtk.ToggleToolButton)
	w.viewNeedsSubtitlesButton = m.builder.GetObject("viewNeedsSubtitles").(*gtk.ToggleToolButton)

	_ = w.viewAllButton.Connect("toggled", func() {
		if w.viewAllButton.GetActive() {
			currentView = viewAll
			w.viewPacksButton.SetActive(false)
			w.viewToWatchButton.SetActive(false)
			w.viewNoRatingButton.SetActive(false)
			w.viewNeedsSubtitlesButton.SetActive(false)
			m.refresh(searchFor, searchGenreId, getSortBy())
		}
	})
	_ = w.viewPacksButton.Connect("toggled", func() {
		if w.viewPacksButton.GetActive() {
			currentView = viewPacks
			w.viewAllButton.SetActive(false)
			w.viewToWatchButton.SetActive(false)
			w.viewNoRatingButton.SetActive(false)
			w.viewNeedsSubtitlesButton.SetActive(false)
			m.refresh(searchFor, searchGenreId, getSortBy())
		}
	})
	_ = w.viewToWatchButton.Connect("toggled", func() {
		if w.viewToWatchButton.GetActive() {
			currentView = viewToWatch
			w.viewAllButton.SetActive(false)
			w.viewPacksButton.SetActive(false)
			w.viewNoRatingButton.SetActive(false)
			w.viewNeedsSubtitlesButton.SetActive(false)
			m.refresh(searchFor, searchGenreId, getSortBy())
		}
	})
	_ = w.viewNoRatingButton.Connect("toggled", func() {
		if w.viewNoRatingButton.GetActive() {
			currentView = viewNoRating
			w.viewAllButton.SetActive(false)
			w.viewPacksButton.SetActive(false)
			w.viewToWatchButton.SetActive(false)
			w.viewNeedsSubtitlesButton.SetActive(false)
			m.refresh(searchFor, searchGenreId, getSortBy())
		}
	})
	_ = w.viewNeedsSubtitlesButton.Connect("toggled", func() {
		if w.viewNeedsSubtitlesButton.GetActive() {
			currentView = viewNeedsSubtitles
			w.viewAllButton.SetActive(false)
			w.viewPacksButton.SetActive(false)
			w.viewToWatchButton.SetActive(false)
			w.viewNoRatingButton.SetActive(false)
			m.refresh(searchFor, searchGenreId, getSortBy())
		}
	})

	return w
}

func (w viewManager) changeView(view View) {
	w.viewAllButton.SetActive(false)
	w.viewToWatchButton.SetActive(false)
	w.viewPacksButton.SetActive(false)
	w.viewNoRatingButton.SetActive(false)
	w.viewNeedsSubtitlesButton.SetActive(false)

	switch view {
	case viewAll:
		w.viewAllButton.SetActive(true)
	case viewToWatch:
		w.viewToWatchButton.SetActive(true)
	case viewPacks:
		w.viewPacksButton.SetActive(true)
	case viewNoRating:
		w.viewNoRatingButton.SetActive(true)
	case viewNeedsSubtitles:
		w.viewNeedsSubtitlesButton.SetActive(true)
	}
}
