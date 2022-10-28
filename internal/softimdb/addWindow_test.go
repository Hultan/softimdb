package softimdb

import (
	"testing"
)

func TestAddWindow_getIdFromUrl(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{""}, "", true},
		{"url #1", args{"https://www.imdb.com/title/tt12593682/?ref_=nv_sr_srsg_0"}, "tt12593682", false},
		{"url #2", args{"imdb.com/title/tt12593682/?ref_=nv_sr_srsg_0"}, "tt12593682", false},
		{"url #3", args{"tt12593682/?ref_=nv_sr_srsg_0"}, "tt12593682", false},
		{"not enough numbers", args{"https://www.imdb.com/title/tt125"}, "", true},
		{"ll instead of tt", args{"https://www.imdb.com/title/ll12593682"}, "", true},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				a := AddWindow{}
				got, err := a.getIdFromUrl(tt.args.url)
				if (err != nil) != tt.wantErr {
					t.Errorf("getIdFromUrl() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("getIdFromUrl() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
