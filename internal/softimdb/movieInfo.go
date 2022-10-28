package softimdb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hultan/softimdb/internal/data"
	"github.com/hultan/softimdb/internal/imdb"
)

type MovieInfo struct {
	title     string
	subTitle  string
	storyLine string
	year      string
	path      string

	imdbRating string
	imdbUrl    string
	imdbId     string

	tags            string // Info field only
	image           []byte
	imageHasChanged bool
}

func newMovieInfoFromImdb(movie *imdb.Movie) (*MovieInfo, error) {
	image, err := movie.GetPoster()
	if err != nil {
		return nil, err
	}

	return &MovieInfo{
		title:           movie.Title,
		storyLine:       movie.StoryLine,
		year:            movie.Year,
		imdbRating:      movie.Rating,
		imdbUrl:         movie.GetURL(),
		imdbId:          movie.Id,
		tags:            movie.Genres,
		image:           image,
		imageHasChanged: false,
	}, nil
}

func newMovieInfoFromDatabase(movie *data.Movie) (*MovieInfo, error) {
	return &MovieInfo{
		title:           movie.Title,
		subTitle:        movie.SubTitle,
		storyLine:       movie.StoryLine,
		year:            fmt.Sprintf("%d", movie.Year),
		path:            movie.MoviePath,
		imdbRating:      fmt.Sprintf("%.1f", movie.ImdbRating),
		imdbUrl:         movie.ImdbUrl,
		imdbId:          movie.ImdbID,
		tags:            getTagsString(movie.Tags),
		image:           movie.Image,
		imageHasChanged: false,
	}, nil
}

func (m *MovieInfo) toDatabase(movie *data.Movie) {
	movie.Title = m.title
	movie.SubTitle = m.subTitle
	movie.StoryLine = m.storyLine
	movie.MoviePath = m.path
	movie.Year = m.getYear()
	movie.ImdbID = m.imdbId
	movie.ImdbUrl = m.imdbUrl
	movie.ImdbRating = m.getImdbRating()
	movie.Tags = m.getTags(m.tags)
	if m.imageHasChanged {
		movie.HasImage = true
		movie.ImagePath = ""
		movie.Image = m.image
	}
}

func (m *MovieInfo) getImdbRating() float32 {
	rating, err := strconv.ParseFloat(m.imdbRating, 64)
	if err != nil {
		return 0.0
	}
	return float32(rating)
}

func (m *MovieInfo) getYear() int {
	year, err := strconv.Atoi(m.year)
	if err != nil {
		return 0
	}
	return year
}

func (m *MovieInfo) getTags(tags string) []data.Tag {
	var result []data.Tag
	tagItems := strings.Split(tags, ",")
	for _, item := range tagItems {
		result = append(result, data.Tag{Name: item})
	}
	return result
}

func getTagsString(tags []data.Tag) string {
	result := ""
	for _, tag := range tags {
		if result != "" {
			result += ","
		}
		result += tag.Name
	}
	return result
}
