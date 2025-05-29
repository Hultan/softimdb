package softimdb

const (
	applicationTitle     = "SoftImdb"
	applicationVersion   = "v 2.8.8"
	applicationCopyRight = "©SoftTeam AB, 2025"
	listMargin           = 3
	listSpacing          = 0
	imageWidth           = 190
	imageHeight          = 280
)

const (
	sortByName     = "title"
	sortByRating   = "imdb_rating"
	sortByMyRating = "my_rating"
	sortByYear     = "year"
	sortById       = "id"
	sortByLength   = "length"
)

const (
	sortAscending  = "asc"
	sortDescending = "desc"
)

type View string

const (
	viewAll            View = "all"
	viewPacks               = "packs"
	viewToWatch             = "toWatch"
	viewNoRating            = "noRating"
	viewNeedsSubtitles      = "needsSubtitles"
)

const configFile = "~/.config/softteam/softimdb/config.json"
