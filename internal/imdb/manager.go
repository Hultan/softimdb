package imdb

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"

	"github.com/hultan/softimdb/internal/data"
)

// Manager represents a IMDB screen scraper.
type Manager struct {
}

// ManagerNew creates a new IMDB Manager
func ManagerNew() *Manager {
	imdb := new(Manager)
	return imdb
}

// GetMovieInfo fills in some IMDB information on the Movie instance passed.
func (i *Manager) GetMovieInfo(movie *data.Movie) error {
	doc, err := i.getDocument(movie.ImdbUrl)
	if err != nil {
		return err
	}

	result := i.parseDocument(doc, movie)
	if !result {
		return errors.New("failed to retrieve movie information")
	}

	return nil
}

func (i *Manager) getDocument(url string) (*goquery.Document, error) {
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

func (i *Manager) parseDocument(doc *goquery.Document, movie *data.Movie) bool {
	// Title
	title, ok := getMovieTitle(doc)
	if !ok {
		// We failed to get the movie title, abort
		return false
	}

	// Year
	year, ok := getMovieYear(doc)
	if !ok {
		// We retrieved an invalid release year, abort
		return false
	}

	// Get movie poster
	poster, ok := getMoviePoster(doc)
	if !ok {
		// We failed to retrieve an image, abort
		return false
	}

	// Rating
	rating, ok := getMovieRating(doc)
	if !ok {
		return false
	}

	// StoryLine
	storyLine, ok := getMovieStoryLine(doc)
	if !ok {
		// We retrieved an invalid storyline, abort
		return false
	}

	// Genres
	genres, ok := getMovieGenres(doc)
	if !ok {
		// We failed to retrieve at least one tag, abort
		return false
	}

	// Store data
	movie.Title = title
	movie.Year = year
	movie.Image = &poster
	movie.HasImage = true
	movie.ImdbRating = rating
	movie.StoryLine = storyLine
	movie.Tags = genres

	return true
}

func getMovieGenres(doc *goquery.Document) ([]data.Tag, bool) {
	var genres []data.Tag

	ok := false
	doc.Find("span.ipc-chip__text").Each(func(x int, s *goquery.Selection) {
		genreName := s.Text()
		if genreName == "Back to top" {
			// Ignore Back to top-button that occasionally shows up here
			// Skip this tag
			return
		}
		genre := data.Tag{Name: genreName}
		genres = append(genres, genre)
		ok = true
	})
	if !ok {
		// We failed to retrieve at least one tag, abort
		return nil, false
	}
	return genres, true
}

func getMovieStoryLine(doc *goquery.Document) (string, bool) {
	storyLine := doc.Find("div.ipc-html-content div").First().Text()
	if storyLine == "" {
		// We retrieved an invalid storyline, abort
		return "", false
	}
	return storyLine, true
}

func getMovieRating(doc *goquery.Document) (float32, bool) {
	ratingString := doc.Find(".jGRxWM").Text()
	rating, err := strconv.ParseFloat(ratingString[:3], 32)
	if err != nil || rating < 1 || rating > 10 {
		return -1, false
	}
	return float32(rating), true
}

func getMoviePoster(doc *goquery.Document) ([]byte, bool) {
	s := doc.Find("div.ipc-media--poster-l img.ipc-image").First()
	imageSource, ok := s.Attr("src")
	if !ok {
		return nil, false
	}
	imageData, err := downloadFile(imageSource)
	if err != nil {
		return nil, false
	}
	return imageData, true
}

func getMovieYear(doc *goquery.Document) (int, bool) {
	year := doc.Find(".itZqyK").Text()
	yearInt, err := strconv.Atoi(year[:4])
	if err != nil || yearInt < 1900 || yearInt > 2100 {
		// We retrieved an invalid release year, abort
		return -1, false
	}
	return yearInt, true
}

func getMovieTitle(doc *goquery.Document) (string, bool) {
	title := doc.Find(".sc-b73cd867-0").Text()
	if title == "" {
		// We failed to get the movie title, abort
		return "", false
	}
	return title, true
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
