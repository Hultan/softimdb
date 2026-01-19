package softimdb

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hultan/softimdb/internal/data"
)

type Movie struct {
	title     string
	subTitle  string
	storyLine string
	year      string
	myRating  int
	moviePath string
	runtime   int
	genres    string // Info field only
	persons   []data.Person
	size      int
	watchedAt *time.Time

	imdbRating string
	imdbUrl    string
	imdbId     string

	image           []byte
	imageHasChanged bool

	toWatch       bool
	pack          string
	needsSubtitle bool
}

func (m *Movie) fromDatabase(movie *data.Movie) {
	m.title = movie.Title
	m.subTitle = movie.SubTitle
	m.storyLine = movie.StoryLine
	m.year = fmt.Sprintf("%d", movie.Year)
	m.myRating = movie.MyRating
	m.moviePath = movie.MoviePath
	m.runtime = movie.Runtime
	m.size = movie.Size
	if movie.WatchedAt.Valid {
		m.watchedAt = &movie.WatchedAt.Time
	}
	m.toWatch = movie.ToWatch
	m.needsSubtitle = movie.NeedsSubtitle
	m.pack = movie.Pack
	m.imdbRating = fmt.Sprintf("%.1f", movie.ImdbRating)
	m.imdbUrl = movie.ImdbUrl
	m.imdbId = movie.ImdbID
	m.genres = getGenresString(movie.Genres)
	m.image = movie.Image
	m.imageHasChanged = false
}

func (m *Movie) toDatabase(movie *data.Movie) {
	movie.Title = m.title
	movie.SubTitle = m.subTitle
	movie.StoryLine = m.storyLine
	movie.MoviePath = m.moviePath
	movie.Pack = m.pack
	movie.Year = m.getYear()
	movie.MyRating = m.myRating
	movie.Runtime = m.runtime
	movie.ToWatch = m.toWatch
	movie.NeedsSubtitle = m.needsSubtitle
	movie.ImdbID = m.imdbId
	movie.ImdbUrl = m.imdbUrl
	movie.Size = m.size
	if m.watchedAt != nil {
		movie.WatchedAt.Time = *m.watchedAt
		movie.WatchedAt.Valid = true
	} else {
		movie.WatchedAt.Valid = false
	}

	movie.ImdbRating = m.getImdbRating()
	movie.Genres = m.getGenres(m.genres)
	for _, person := range m.persons {
		movie.Persons = append(movie.Persons, person)
	}
	if m.imageHasChanged {
		movie.HasImage = true
		movie.Image = m.image
	}
}

func (m *Movie) getImdbRating() float32 {
	rating, err := strconv.ParseFloat(m.imdbRating, 64)
	if err != nil {
		return 0.0
	}
	return float32(rating)
}

func (m *Movie) getYear() int {
	year, err := strconv.Atoi(m.year)
	if err != nil {
		return 0
	}
	return year
}

func (m *Movie) getGenres(genres string) []data.Genre {
	var result []data.Genre

	for _, item := range strings.Split(genres, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, data.Genre{Name: item})
		}
	}

	return result
}

func getGenresString(genres []data.Genre) string {
	result := ""
	for _, genre := range genres {
		if showPrivateGenres || !genre.IsPrivate {
			if result != "" {
				result += ","
			}
			result += genre.Name
		}
	}
	return result
}
