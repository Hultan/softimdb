package imdb

import (
	"testing"
)

func Test_replaceParameters(t *testing.T) {
	type args struct {
		url        string
		parameters []parameter
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"search movie",
			args{searchMoviesURL, []parameter{{"{APIKEY}", testApiKey}, {"{SEARCH}", "TestSearchString"}}},
			"https://imdb-api.com/en/API/SearchMovie/k_00000001/TestSearchString",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceParameters(tt.args.url, tt.args.parameters); got != tt.want {
				t.Errorf("replaceParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}
