package data

// ImageCache represents an image cache that loads the images from the local filesystem.
type ImageCache struct {
	data map[int][]byte
}

// imageCacheNew creates a new ImageCache.
func imageCacheNew() *ImageCache {
	cache := new(ImageCache)
	cache.data = make(map[int][]byte, 2000)

	return cache
}

// save saves the image to the cache.
func (i *ImageCache) save(index int, image []byte) {
	if i.load(index) != nil {
		return
	}

	i.data[index] = image
}

// load loads the image from the cache.
func (i *ImageCache) load(index int) []byte {
	value, ok := i.data[index]
	if ok {
		return value
	}
	return nil
}
