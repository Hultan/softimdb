package imdb

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Manager represents a IMDB screen scraper.
type Manager struct {
	Errors []error
}

type Movie struct {
	Title   string
	Year    int
	Poster  []byte
	Rating  string
	Genres  []string
	Runtime int
}

// ManagerNew creates a new IMDB Manager
func ManagerNew() Manager {
	return Manager{}
}

// GetMovie fills in some IMDB information on the Movie instance passed.
func (i Manager) GetMovie(url string) (*Movie, error) {
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

	if len(i.Errors) > 0 {
		return info, i.Errors[0]
	}

	return info, nil
}

func (i Manager) getGoQueryDocument(url string) (*goquery.Document, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

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

func (i Manager) parseGoQueryDocument(doc *goquery.Document) *Movie {
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

	// Runtime
	runtime, err := getMovieRuntime(doc)
	if err != nil {
		i.Errors = append(i.Errors, err)
	}

	// Rating
	rating, err := getMovieRating(doc)
	if err != nil {
		i.Errors = append(i.Errors, err)
	}

	info := &Movie{
		Title:   title,
		Year:    year,
		Runtime: runtime,
		Rating:  rating,
	}

	return info
}

func getMovieRating(doc *goquery.Document) (string, error) {
	return doc.Find(".imUuxf").First().Text(), nil
}

func getMovieRuntime(doc *goquery.Document) (int, error) {
	runtimeString := doc.Find("ul.sc-ec65ba05-2").Children().Last().First().Text()
	runtime, err := calcRuntime(runtimeString)
	if err != nil {
		return -1, err
	}
	return runtime, nil
}

func calcRuntime(runtimeString string) (int, error) {
	items := strings.Split(runtimeString, " ")
	if len(items) != 1 && len(items) != 2 {
		return -1, errors.New(fmt.Sprintf("invalid IMDB runtime: %s", runtimeString))
	}

	var hours, minutes int
	var err error
	switch len(items) {
	case 1:
		if !strings.HasSuffix(items[0], "m") {
			return -1, errors.New(fmt.Sprintf("invalid IMDB runtime: %s", runtimeString))
		}
		minutes, err = strconv.Atoi(items[0][:len(items[0])-1])
		if err != nil {
			return -1, err
		}
	case 2:
		if !strings.HasSuffix(items[0], "h") {
			return -1, errors.New(fmt.Sprintf("invalid IMDB runtime: %s", runtimeString))
		}
		hours, err = strconv.Atoi(items[0][:len(items[0])-1])
		if err != nil {
			return -1, err
		}
		if !strings.HasSuffix(items[1], "m") {
			return -1, errors.New(fmt.Sprintf("invalid IMDB runtime: %s", runtimeString))
		}
		minutes, err = strconv.Atoi(items[1][:len(items[1])-1])
		if err != nil {
			return -1, err
		}
	}

	return hours*60 + minutes, nil
}

func getMovieYear(doc *goquery.Document) (int, error) {
	year := doc.Find("ul.sc-ec65ba05-2").First().First().Text()
	yearInt, err := strconv.Atoi(year[:4])
	if err != nil || yearInt < 1900 || yearInt > 2100 {
		// We retrieved an invalid release year, abort
		return -1, err
	}
	return yearInt, nil
}

func getMovieTitle(doc *goquery.Document) (string, error) {
	title := doc.Find(".hero__primary-text").Text()
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
