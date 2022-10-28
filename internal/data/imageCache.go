package data

// ImageCache represents an image cache that loads the images from the local filesystem.
type ImageCache struct {
	data map[int][]byte
}

// ImageCacheNew creates a new ImageCache.
func ImageCacheNew() *ImageCache {
	cache := new(ImageCache)
	cache.data = make(map[int][]byte, 1000)

	return cache
}

// Save saves the image to the cache.
func (i *ImageCache) Save(index int, image []byte) {
	if i.Load(index) != nil {
		return
	}

	i.data[index] = image
}

// Load loads the image from the cache.
func (i *ImageCache) Load(index int) []byte {
	value, ok := i.data[index]
	if ok {
		return value
	}
	return nil
}
