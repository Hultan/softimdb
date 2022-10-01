package imdb2

type MovieResults struct {
	SearchType   string        `json:"searchType"`
	Expression   string        `json:"expression"`
	Results      []MovieResult `json:"results"`
	ErrorMessage string        `json:"errorMessage"`
}

type MovieResult struct {
	Id          string `json:"id"`
	ResultType  string `json:"resultType"`
	Image       string `json:"image"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Movie struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	Year      string `json:"year"`
	ImageURL  string `json:"image"`
	StoryLine string `json:"plot"`
	Genres    string `json:"genres"`
	Rating    string `json:"imDbRating"`

	ErrorMessage string `json:"errorMessage"`
}
