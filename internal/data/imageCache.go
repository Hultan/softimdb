package data

type ImageCache struct {
	data map[int]*[]byte
}

func ImageCacheNew() *ImageCache {
	cache := new(ImageCache)
	cache.data = make(map[int]*[]byte,500)

	return cache
}

func (i *ImageCache) Save(index int, image *[]byte) {
	if i.Load(index) != nil {
		return
	}

	i.data[index] = image
}

func (i *ImageCache) Load(index int) *[]byte {
	value, ok := i.data[index]
	if ok {
		return value
	}
	return nil
}

