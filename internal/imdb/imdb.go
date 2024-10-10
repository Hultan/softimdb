package imdb_old

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// IMDBManager represents a IMDB screen scraper.
type IMDBManager struct {
	Errors []error
}

type IMDBMovie struct {
	Title     string
	Year      int
	Poster    []byte
	Rating    float64
	StoryLine string
	Genres    []string
}

// ManagerNew creates a new IMDB IMDBManager
func ManagerNew() IMDBManager {
	return IMDBManager{}
}

// GetMovie fills in some IMDB information on the Movie instance passed.
func (i IMDBManager) GetMovie(url string) (*IMDBMovie, error) {
	// Clear errors
	i.Errors = nil

	// Get GoQuery document from URL
	doc, err := i.getGoQueryDocument(url)
	if err != nil {
		i.Errors = append(i.Errors, err)
		return nil, err
	}

	// Parse GoQuery document
	info := i.parseGoQueryDocument(doc)

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

func (i IMDBManager) parseGoQueryDocument(doc *goquery.Document) *IMDBMovie {
	// Title
	title, err := getMovieTitle(doc)
	if err != nil {
		i.Errors = append(i.Errors, err)
	}

	// Year
	year, err := getMovieYear(doc)
	if err != nil {
		i.Errors = append(i.Errors, err)
	}

	// Get movie poster
	poster, err := getMoviePoster(doc)
	if err != nil {
		i.Errors = append(i.Errors, err)
	}

	// Rating
	rating, err := getMovieRating(doc)
	if err != nil {
		i.Errors = append(i.Errors, err)
	}

	// StoryLine
	storyLine, err := getMovieStoryLine(doc)
	if err != nil {
		i.Errors = append(i.Errors, err)
	}

	// Genres
	genres, err := getMovieGenres(doc)
	if err != nil {
		i.Errors = append(i.Errors, err)
	}

	info := &IMDBMovie{
		Title:     title,
		Year:      year,
		Poster:    poster,
		Rating:    rating,
		StoryLine: storyLine,
		Genres:    genres,
	}

	return info
}

func getMovieGenres(doc *goquery.Document) ([]string, error) {
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
		return nil, errors.New("could not find genre")
	}
	return genres, nil
}

func getMovieStoryLine(doc *goquery.Document) (string, error) {
	selection := doc.Find("div.ipc-html-content div")
	storyLine := selection.Text()
	if storyLine == "" {
		// We retrieved an invalid storyline, abort
		return "", errors.New("story line is empty")
	}
	return storyLine, nil
}

func getMovieRating(doc *goquery.Document) (float64, error) {
	ratingString := doc.Find(".jGRxWM").Text()
	rating, err := strconv.ParseFloat(ratingString[:3], 32)
	if err != nil {
		return -1, err
	}
	if rating < 1 || rating > 10 {
		return -1, errors.New(fmt.Sprintf("invalid IMDB rating: %s", ratingString))
	}
	return rating, nil
}

func getMoviePoster(doc *goquery.Document) ([]byte, error) {
	s := doc.Find("div.ipc-media--poster-l img.ipc-image").First()
	imageSource, ok := s.Attr("src")
	if !ok {
		return nil, errors.New("could not find image source")
	}
	imageData, err := downloadFile(imageSource)
	if err != nil {
		return nil, errors.New("could not download image")
	}
	return imageData, nil
}

func getMovieYear(doc *goquery.Document) (int, error) {
	year := doc.Find(".itZqyK").Text()
	yearInt, err := strconv.Atoi(year[:4])
	if err != nil || yearInt < 1900 || yearInt > 2100 {
		// We retrieved an invalid release year, abort
		return -1, err
	}
	return yearInt, nil
}

func getMovieTitle(doc *goquery.Document) (string, error) {
	title := doc.Find(".sc-b73cd867-0").Text()
	if title == "" {
		// We failed to get the movie title, abort
		return "", errors.New("title is empty")
	}
	return title, nil
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
