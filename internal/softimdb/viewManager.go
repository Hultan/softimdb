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
			m.view.current = viewAll
			w.viewPacksButton.SetActive(false)
			w.viewToWatchButton.SetActive(false)
			w.viewNoRatingButton.SetActive(false)
			w.viewNeedsSubtitlesButton.SetActive(false)
			m.refresh(m.search, getSortBy(m.sort))
		}
	})
	_ = w.viewPacksButton.Connect("toggled", func() {
		if w.viewPacksButton.GetActive() {
			m.view.current = viewPacks
			w.viewAllButton.SetActive(false)
			w.viewToWatchButton.SetActive(false)
			w.viewNoRatingButton.SetActive(false)
			w.viewNeedsSubtitlesButton.SetActive(false)
			m.refresh(m.search, getSortBy(m.sort))
		}
	})
	_ = w.viewToWatchButton.Connect("toggled", func() {
		if w.viewToWatchButton.GetActive() {
			m.view.current = viewToWatch
			w.viewAllButton.SetActive(false)
			w.viewPacksButton.SetActive(false)
			w.viewNoRatingButton.SetActive(false)
			w.viewNeedsSubtitlesButton.SetActive(false)
			m.refresh(m.search, getSortBy(m.sort))
		}
	})
	_ = w.viewNoRatingButton.Connect("toggled", func() {
		if w.viewNoRatingButton.GetActive() {
			m.view.current = viewNoRating
			w.viewAllButton.SetActive(false)
			w.viewPacksButton.SetActive(false)
			w.viewToWatchButton.SetActive(false)
			w.viewNeedsSubtitlesButton.SetActive(false)
			m.refresh(m.search, getSortBy(m.sort))
		}
	})
	_ = w.viewNeedsSubtitlesButton.Connect("toggled", func() {
		if w.viewNeedsSubtitlesButton.GetActive() {
			m.view.current = viewNeedsSubtitles
			w.viewAllButton.SetActive(false)
			w.viewPacksButton.SetActive(false)
			w.viewToWatchButton.SetActive(false)
			w.viewNoRatingButton.SetActive(false)
			m.refresh(m.search, getSortBy(m.sort))
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
