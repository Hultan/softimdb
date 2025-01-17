package imdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

const constYearSelector = "div:has(h1[data-testid='hero__pageTitle']) > ul.ipc-inline-list > li:nth-child([CHILD])"

const (
	Director int = iota
	Writer
	Actor
)

// Manager represents a IMDB screen scraper.
type Manager struct {
	Errors []error
}

type PersonImdb struct {
	Name string
	Type int
}

type MovieImdb struct {
	// Done
	Title     string
	Year      int
	Rating    string
	Runtime   int
	StoryLine string
	Genres    []string
	Persons   []PersonImdb
	Poster    []byte
}

// ManagerNew creates a new IMDB Manager
func ManagerNew() *Manager {
	return &Manager{}
}

// GetMovie fills in some IMDB information on the Movie instance passed.
func (m *Manager) GetMovie(url string) (*MovieImdb, error) {
	// Clear errors
	m.Errors = nil

	// Get GoQuery document from URL
	doc, err := m.getGoQueryDocument(url)
	if err != nil {
		m.Errors = append(m.Errors, err)
		return nil, err
	}

	// Parse GoQuery document
	info := m.parseGoQueryDocument(doc)

	if len(m.Errors) > 0 {
		return info, m.Errors[0]
	}

	return info, nil
}

func (m *Manager) getGoQueryDocument(url string) (*goquery.Document, error) {
	// Create a ChromeDP allocator with User-Agent header and flags
	allocatorCtx, cancelAllocator := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
			chromedp.Flag("headless", false),
		)...,
	)
	defer cancelAllocator()

	// Create a ChromeDP context with extended timeout
	ctx, cancel := chromedp.NewContext(allocatorCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Set a timeout for the context to ensure it doesn't hang
	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	// Get the Html from the Url
	pageHTML, err := m.scrapeUrl(url, ctx)
	if err != nil {
		return nil, err
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(pageHTML))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (m *Manager) scrapeUrl(url string, ctx context.Context) (string, error) {
	var pageHTML string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),

		// Click the "Accept Cookies" button using data-testid
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Click(`button[data-testid="accept-button"]`, chromedp.ByQuery).Do(ctx)
			if err != nil {
				log.Println("No cookie popup detected or click failed, continuing...")
			}
			return nil
		}),

		// Scroll the page smaller increments to trigger content loading
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < 20; i++ { // Try up to 20 scrolls
				err := chromedp.Evaluate(`window.scrollBy(0, 1200);`, nil).Do(ctx)
				if err != nil {
					log.Printf("Scroll attempt %d failed: %v\n", i+1, err)
				}
				time.Sleep(3 * time.Second) // Allow time for content to load

				// Check if the storyline is now visible after each scroll
				var isVisible bool
				err = chromedp.Evaluate(`document.querySelector('div[data-testid="storyline-plot-summary"]') !== null`, &isVisible).Do(ctx)
				if err != nil {
					log.Println("Error checking storyline visibility:", err)
				}
				if isVisible {
					log.Println("Storyline is now visible after scrolling!")
					break
				}
			}
			return nil
		}),

		// Capture the full page HTML after scrolling for parsing with GoQuery
		chromedp.OuterHTML("html", &pageHTML),
	)

	return pageHTML, err
}

func (m *Manager) parseGoQueryDocument(doc *goquery.Document) *MovieImdb {
	// Title
	title, err := m.getMovieTitle(doc)
	if err != nil {
		m.Errors = append(m.Errors, err)
	}

	// Year
	year, err := m.getMovieYear(doc)
	if err != nil {
		m.Errors = append(m.Errors, err)
	}

	// Runtime
	runtime, err := m.getMovieRuntime(doc)
	if err != nil {
		m.Errors = append(m.Errors, err)
	}

	// Rating
	rating, err := m.getMovieRating(doc)
	if err != nil {
		m.Errors = append(m.Errors, err)
	}

	// StoryLine
	storyLine, err := m.getMovieStoryLine(doc)
	if err != nil {
		m.Errors = append(m.Errors, err)
	}

	// Genres
	genres, err := m.getMovieGenres(doc)
	if err != nil {
		m.Errors = append(m.Errors, err)
	}

	// Persons
	people, err := m.getMoviePeople(doc)
	if err != nil {
		m.Errors = append(m.Errors, err)
	}

	// Poster
	poster, err := m.getMoviePoster(doc)
	if err != nil {
		m.Errors = append(m.Errors, err)
	}

	info := &MovieImdb{
		Title:     title,
		Year:      year,
		Runtime:   runtime,
		Rating:    rating,
		StoryLine: storyLine,
		Genres:    genres,
		Persons:   people,
		Poster:    poster,
	}

	return info
}

func (m *Manager) getMoviePoster(doc *goquery.Document) ([]byte, error) {
	src, ok := doc.Find(`img[width="190"]`).Attr("src")
	if ok {
		imageData, err := m.downloadFile(src)
		if err != nil {
			return nil, err
		}
		return imageData, nil
	}
	return nil, errors.New("couldn't find movie poster")
}

func (m *Manager) getMovieGenres(doc *goquery.Document) ([]string, error) {
	var genres []string

	// Genres
	doc.Find(`li[data-testid="storyline-genres"]`).Each(func(i int, s *goquery.Selection) {
		// Find the <ul> within this <li> and iterate over its <li> children
		s.Find("ul li").Each(func(j int, genre *goquery.Selection) {
			// Get the text of each genre and append it to the genres slice
			genres = append(genres, genre.Text())
		})
	})
	return genres, nil
}

func (m *Manager) getMovieStoryLine(doc *goquery.Document) (string, error) {
	storyLine := doc.Find(`div[data-testid="storyline-plot-summary"]`).Text()
	if storyLine == "" {
		// We retrieved an invalid storyLine, abort
		return "", errors.New("story line is empty")
	}
	return storyLine, nil
}

func (m *Manager) getMovieRating(doc *goquery.Document) (string, error) {
	return doc.Find(`div[data-testid="hero-rating-bar__aggregate-rating__score"]`).Find("span").First().Text(), nil
}

func (m *Manager) getMovieRuntime(doc *goquery.Document) (int, error) {
	// Initialize a variable to hold the runtime
	var runtimeString string

	// Use a specific selector to find the <ul> that is a child of the <div> containing the <h1>
	doc.Find("div:has(h1[data-testid='hero__pageTitle']) > ul.ipc-inline-list > li").Each(func(i int, s *goquery.Selection) {
		// Check if this <li> does not have any <a> elements
		if s.Find("a").Length() == 0 {
			// Get the text of the <li>
			runtimeString = s.Text()
		}
	})

	runtime, err := m.calcRuntime(runtimeString)
	if err != nil {
		return -1, err
	}
	return runtime, nil
}

func (m *Manager) calcRuntime(runtimeString string) (int, error) {
	items := strings.Split(runtimeString, " ")
	if len(items) != 1 && len(items) != 2 {
		return -1, errors.New(fmt.Sprintf("invalid IMDB runtime: %s", runtimeString))
	}

	var hours, minutes int
	var err error
	switch len(items) {
	case 1:
		if strings.HasSuffix(items[0], "m") {
			minutes, err = strconv.Atoi(items[0][:len(items[0])-1])
		} else if strings.HasSuffix(items[0], "h") {
			hours, err = strconv.Atoi(items[0][:len(items[0])-1])
		} else {
			return -1, errors.New(fmt.Sprintf("invalid IMDB runtime: %s", runtimeString))
		}
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

func (m *Manager) getMovieYear(doc *goquery.Document) (yearInt int, err error) {
	for i := 1; i <= 2; i++ {
		selector := strings.Replace(constYearSelector, "[CHILD]", strconv.Itoa(i), -1)
		year := doc.Find(selector).Text()
		if yearInt, err = m.parseYear(year); err == nil {
			break
		}
	}

	if err != nil {
		return -1, err
	}

	return yearInt, nil
}

func (m *Manager) parseYear(year string) (int, error) {
	year = strings.TrimSpace(year)
	if len(year) > 4 {
		year = year[:4]
	}
	yearInt, err := strconv.Atoi(year)
	if err != nil || yearInt < 1900 || yearInt > 2100 {
		// We retrieved an invalid release year, abort
		return -1, err
	}

	return yearInt, nil
}

func (m *Manager) getMovieTitle(doc *goquery.Document) (string, error) {
	title := doc.Find("h1[data-testid='hero__pageTitle'] > span.hero__primary-text").Text()

	// Trim any whitespace from the extracted text
	title = strings.TrimSpace(title)

	if title == "" {
		// We failed to get the movie title, abort
		return "", errors.New("title is empty")
	}
	return title, nil
}

func (m *Manager) downloadFile(url string) ([]byte, error) {
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

func (m *Manager) getMoviePeople(doc *goquery.Document) ([]PersonImdb, error) {
	var persons []PersonImdb

	// Directors
	doc.Find(".sc-70a366cc-3 li.ipc-metadata-list__item:contains('Director') ." +
		"ipc-metadata-list-item__list-content-item--link").Each(
		func(i int, s *goquery.Selection) {
			name := s.Text()
			persons = append(persons, PersonImdb{
				Name: name,
				Type: Director,
			})
		})

	// Writers
	doc.Find(".sc-70a366cc-3 li.ipc-metadata-list__item:contains('Writer') ." +
		"ipc-metadata-list-item__list-content-item--link").Each(
		func(i int, s *goquery.Selection) {
			name := s.Text()
			persons = append(persons, PersonImdb{
				Name: name,
				Type: Writer,
			})
		})

	// Actors
	doc.Find(".sc-cd7dc4b7-5 .sc-cd7dc4b7-1").Each(
		func(i int, s *goquery.Selection) {
			name := s.Text()
			persons = append(persons, PersonImdb{
				Name: name,
				Type: Actor,
			})
		})

	return persons, nil
}
