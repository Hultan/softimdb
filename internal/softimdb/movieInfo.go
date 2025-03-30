package softimdb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hultan/softimdb/internal/data"
)

type movieInfo struct {
	title     string
	subTitle  string
	storyLine string
	year      string
	myRating  int
	moviePath string
	runtime   int
	genres    string // Info field only
	persons   []data.Person

	imdbRating string
	imdbUrl    string
	imdbId     string

	image           []byte
	imageHasChanged bool

	toWatch       bool
	pack          string
	needsSubtitle bool
}

func newMovieInfoFromDatabase(movie *data.Movie) (*movieInfo, error) {
	return &movieInfo{
		title:           movie.Title,
		subTitle:        movie.SubTitle,
		storyLine:       movie.StoryLine,
		year:            fmt.Sprintf("%d", movie.Year),
		myRating:        movie.MyRating,
		moviePath:       movie.MoviePath,
		runtime:         movie.Runtime,
		toWatch:         movie.ToWatch,
		needsSubtitle:   movie.NeedsSubtitle,
		pack:            movie.Pack,
		imdbRating:      fmt.Sprintf("%.1f", movie.ImdbRating),
		imdbUrl:         movie.ImdbUrl,
		imdbId:          movie.ImdbID,
		genres:          getGenresString(movie.Genres),
		image:           movie.Image,
		imageHasChanged: false,
	}, nil
}

func (m *movieInfo) toDatabase(movie *data.Movie) {
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

func (m *movieInfo) getImdbRating() float32 {
	rating, err := strconv.ParseFloat(m.imdbRating, 64)
	if err != nil {
		return 0.0
	}
	return float32(rating)
}

func (m *movieInfo) getYear() int {
	year, err := strconv.Atoi(m.year)
	if err != nil {
		return 0
	}
	return year
}

func (m *movieInfo) getGenres(genres string) []data.Genre {
	var result []data.Genre
	genres = strings.Trim(genres, " ")
	genreItems := strings.Split(genres, ",")

	// Avoid inserting empty genres
	if len(genreItems) == 0 ||
		(len(genreItems) == 1 && strings.Trim(genreItems[0], " \n\t") == "") {
		return result
	}

	for _, item := range genreItems {
		result = append(result, data.Genre{Name: item})
	}
	return result
}

func getGenresString(genres []data.Genre) string {
	result := ""
	for _, genre := range genres {
		if result != "" {
			result += ","
		}
		result += genre.Name
	}
	return result
}
