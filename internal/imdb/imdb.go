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

// Manager represents a IMDB screen scraper.
type Manager struct {
	Errors []error
}

type Movie struct {
	// Done
	Title   string
	Year    int
	Rating  string
	Runtime int

	// Not working
	StoryLine string
	Poster    []byte
	Genres    []string
}

// ManagerNew creates a new IMDB Manager
func ManagerNew() *Manager {
	return &Manager{}
}

// GetMovie fills in some IMDB information on the Movie instance passed.
func (i *Manager) GetMovie(url string) (*Movie, error) {
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

func (i *Manager) getGoQueryDocument(url string) (*goquery.Document, error) {
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
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

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
				time.Sleep(1 * time.Second) // Allow time for content to load

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

func (i *Manager) parseGoQueryDocument(doc *goquery.Document) *Movie {
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

	// StoryLine
	storyLine, err := getMovieStoryLine(doc)
	if err != nil {
		i.Errors = append(i.Errors, err)
	}

	info := &Movie{
		Title:     title,
		Year:      year,
		Runtime:   runtime,
		Rating:    rating,
		StoryLine: storyLine,
	}

	return info
}

func getMovieStoryLine(doc *goquery.Document) (string, error) {
	fmt.Println(doc.Html())
	storyLine := doc.Find(`div[data-testid="storyline-plot-summary"]`).Text()
	if storyLine == "" {
		// We retrieved an invalid storyLine, abort
		return "", errors.New("story line is empty")
	}
	return storyLine, nil
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
