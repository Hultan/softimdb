package imdb

import (
	"errors"
	"fmt"
	"io/ioutil"
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
	movie.Title = doc.Find(".sc-b73cd867-0").Text()

	// Year
	year := doc.Find(".itZqyK").Text()
	movie.Year, _ = strconv.Atoi(year[:4])

	doc.Find("div.ipc-media--poster-l img.ipc-image").Each(func(x int, s *goquery.Selection) {
		imageSource, ok := s.Attr("src")
		if ok {
			imageData, _ := i.downloadFile(imageSource)
			movie.Image = imageData
			movie.HasImage = true
		}
	})

	// Rating
	ratingString := doc.Find(".jGRxWM").Text()
	rating, _ := strconv.ParseFloat(ratingString[:3], 32)
	movie.ImdbRating = float32(rating)

	// StoryLine
	movie.StoryLine = doc.Find("div.ipc-html-content div").First().Text()

	// Genres
	doc.Find("span.ipc-chip__text").Each(func(x int, s *goquery.Selection) {
		genreName := s.Text()
		if genreName == "Back to top" {
			// Ignore Back to top-button that occasionally shows up here
			return
		}
		// fmt.Println("GENRE:",genreName)
		genre := data.Tag{Name: genreName}
		if movie.Tags == nil {
			movie.Tags = []data.Tag{}
		}
		movie.Tags = append(movie.Tags, genre)
	})

	return true
}

func (i *Manager) downloadFile(url string) (*[]byte, error) {
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

	fileData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &fileData, err
}
