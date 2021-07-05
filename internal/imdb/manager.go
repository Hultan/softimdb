package imdb

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hultan/softimdb/internal/data"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Manager struct {

}

func ManagerNew() *Manager {
	imdb := new(Manager)
	return imdb
}

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
	movie.Title = doc.Find("h1.TitleHeader__TitleText-sc-1wu6n3d-0").Text() //First().Clone().Children().Remove().End()

	//if movie.Title=="" {
	//	return false
	//}

	// Year
	year := doc.Find("a.rgaOW").First().Text()
	movie.Year, _ = strconv.Atoi(year)

	doc.Find("div.ipc-media--poster-l img.ipc-image").Each(func(x int, s *goquery.Selection) {
		imageSource, ok := s.Attr("src")
		if ok {
			imageData,_ := i.downloadFile(imageSource)
			movie.Image = imageData
			movie.HasImage = true
		}
	})

	// Rating
	rating, _ := strconv.ParseFloat(doc.Find("span.AggregateRatingButton__RatingScore-sc-1ll29m0-1").First().Text(), 32)
	movie.ImdbRating = float32(rating)

	// StoryLine
	movie.StoryLine = doc.Find("div.ipc-html-content div").First().Text()

	// Genres
	doc.Find("div.ipc-metadata-list-item__content-container ul li a").Each(func(x int, s *goquery.Selection) {
		tagSource, ok := s.Attr("href")
		if ok && strings.Contains(tagSource, "genres") {
			genreName := s.Text()
			//fmt.Println("GENRE:",genreName)
			genre := data.Tag{Name:genreName}
			if movie.Tags==nil {
				movie.Tags = []data.Tag{}
			}
			movie.Tags = append(movie.Tags, genre)
		}
	})

	return true
}

func (i *Manager) downloadFile(url string) (*[]byte, error) {
	//Get the response bytes from the url
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
