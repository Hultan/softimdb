package imdb_old

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"

	"github.com/hultan/softimdb/internal/data"
)

// IMDBManager represents a IMDB screen scraper.
type IMDBManager struct {
}

type IMDBMovie struct {
	Title            string
	Year             int
	Poster           []byte
	PosterHasChanged bool
	Rating           float64
	StoryLine        string
	Tags             []string
}

// ManagerNew creates a new IMDB IMDBManager
func ManagerNew() IMDBManager {
	return IMDBManager{}
}

// GetMovie fills in some IMDB information on the Movie instance passed.
func (i IMDBManager) GetMovie(movie *data.Movie) (*IMDBMovie, error) {
	doc, err := i.getGoQueryDocument(movie.ImdbUrl)
	if err != nil {
		return nil, err
	}

	info := i.parseGoQueryDocument(doc, movie)

	return info, nil
}

func (i IMDBManager) getGoQueryDocument(url string) (*goquery.Document, error) {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status))
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (i IMDBManager) parseGoQueryDocument(doc *goquery.Document, movie *data.Movie) *IMDBMovie {
	// Title
	title := getMovieTitle(doc)
	// Year
	year := getMovieYear(doc)
	// Get movie poster
	poster := getMoviePoster(doc)
	// Rating
	rating := getMovieRating(doc)
	// StoryLine
	storyLine := getMovieStoryLine(doc)
	// Genres
	genres := getMovieGenres(doc)

	info := &IMDBMovie{
		Title:     title,
		Year:      year,
		Poster:    poster,
		Rating:    rating,
		StoryLine: storyLine,
		Tags:      genres,
	}

	return info
}

func getMovieGenres(doc *goquery.Document) []string {
	var genres []string

	ok := false
	doc.Find("span.ipc-chip__text").Each(func(x int, s *goquery.Selection) {
		genreName := s.Text()
		if genreName == "Back to top" {
			// Ignore Back to top-button that occasionally shows up here
			// Skip this tag
			return
		}
		genres = append(genres, genreName)
		ok = true
	})
	if !ok {
		// We failed to retrieve at least one tag, abort
		return nil
	}
	return genres
}

func getMovieStoryLine(doc *goquery.Document) string {
	selection := doc.Find("div.ipc-html-content div")
	storyLine := selection.Text()
	if storyLine == "" {
		// We retrieved an invalid storyline, abort
		return ""
	}
	return storyLine
}

func getMovieRating(doc *goquery.Document) float64 {
	ratingString := doc.Find(".jGRxWM").Text()
	rating, err := strconv.ParseFloat(ratingString[:3], 32)
	if err != nil || rating < 1 || rating > 10 {
		return -1
	}
	return rating
}

func getMoviePoster(doc *goquery.Document) []byte {
	s := doc.Find("div.ipc-media--poster-l img.ipc-image").First()
	imageSource, ok := s.Attr("src")
	if !ok {
		return nil
	}
	imageData, err := downloadFile(imageSource)
	if err != nil {
		return nil
	}
	return imageData
}

func getMovieYear(doc *goquery.Document) int {
	year := doc.Find(".itZqyK").Text()
	yearInt, err := strconv.Atoi(year[:4])
	if err != nil || yearInt < 1900 || yearInt > 2100 {
		// We retrieved an invalid release year, abort
		return -1
	}
	return yearInt
}

func getMovieTitle(doc *goquery.Document) string {
	title := doc.Find(".sc-b73cd867-0").Text()
	if title == "" {
		// We failed to get the movie title, abort
		return ""
	}
	return title
}

func downloadFile(url string) ([]byte, error) {
	// Get the response bytes from the url
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = response.Body.Close()
	}()

	if response.StatusCode != 200 {
		return nil, errors.New("received non 200 response code")
	}

	// ioutil.ReadAll is deprecated
	fileData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return fileData, err
}
