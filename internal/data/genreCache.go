package data

type GenreCache struct {
	genres []*Genre
}

var genreCache *GenreCache

func genreCacheNew() *GenreCache {
	return &GenreCache{}
}

func (t *GenreCache) getByName(name string) *Genre {
	for _, genre := range t.genres {
		if genre.Name == name {
			return genre
		}
	}
	return nil
}

func (t *GenreCache) add(genre *Genre) {
	t.genres = append(t.genres, genre)
}

func (t *GenreCache) getById(id int) *Genre {
	for _, genre := range t.genres {
		if genre.Id == id {
			return genre
		}
	}
	return nil
}
