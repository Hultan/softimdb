package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/hultan/softimdb/internal/config"
	"gorm.io/gorm"
)

var (
	srcDir       = "/home/per/temp/softimdb/fixed"
	db           *Database
	movies       map[int]*Movie
	targetWidth  = 190
	targetHeight = 280
)

func main() {
	const configFile = "~/.config/softteam/softimdb/config.json"
	cnf, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	// Open the database after we have the config
	db = DatabaseNew(cnf)
	defer db.CloseDatabase()

	m, err := db.SearchMovies("all", "", -1, "")
	movies = make(map[int]*Movie, len(m))
	for _, movie := range m {
		if movie != nil {
			movies[movie.ImageId] = movie
		}
	}

	//for id := 1; id < 1980; id++ {
	//	movie, ok := movies[id]
	//	if !ok {
	//		fmt.Printf("Deleted movie %d\n", id)
	//		continue
	//	}
	//	if doesCachedImageExists(id) {
	//		fmt.Printf("Cached image exist for movie (%d) %s\n", id, movie.Title)
	//		continue
	//	}
	//
	//	manager := imdb.ManagerNew()
	//	mov, err := manager.GetMovie(movie.URL)
	//	if err != nil {
	//		fmt.Printf("Failed to get movie %d from IMDB : %v\n", id, err)
	//		continue
	//	}
	//	dest := fmt.Sprintf("/home/per/temp/softimdb/fixed/%d.jpg", id)
	//	if err := os.WriteFile(dest, mov.Poster, 0o644); err != nil {
	//		fmt.Printf("Failed to write %s for movie %d: %v\n", dest, id, err)
	//		continue
	//	}
	//	err = resizeImage(dest)
	//	if err != nil {
	//		fmt.Printf("Failed to resize image %s : %v\n", dest, err)
	//		continue
	//	}
	//
	//	fmt.Printf("Sucessfully got image %s for (%d) %s\n", dest, id, mov.Title)
	//}
	//
	//return

	// Step 2 – walk the source directory.
	var processed, skipped int
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			// Propagate the error; stop walking.
			return walkErr
		}

		if err := processFile(path, info); err != nil {
			// Log the problem but keep going.
			log.Printf("⚠️  %v", err)
			skipped++
			return nil
		}
		processed++

		return nil
	})
	if err != nil {
		log.Fatalf("Error walking source directory: %v", err)
	}

	fmt.Printf("\nFinished. Processed %d images, skipped %d files.\n", processed, skipped)
}

func doesCachedImageExists(id int) bool {
	p := fmt.Sprintf("%s/%d.jpg", srcDir, id)

	_, err := os.Stat(p)
	if err == nil {
		return true
	}
	return false
}

// processFile loads an image, resizes it, and saves it to the destination.
func processFile(pathImage string, info os.FileInfo) error {
	if !isImage(info.Name()) {
		// Not an image we care about.
		return nil
	}

	// Load the image.
	img, err := imaging.Open(pathImage)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", pathImage, err)
	}

	fileWithExt := info.Name()
	ext := filepath.Ext(fileWithExt)
	fileName := strings.TrimSuffix(fileWithExt, ext) // → "123"
	id, _ := strconv.Atoi(fileName)
	movie := movies[id]

	if movie != nil {
		movie.HasImage = true
		movie.Image = imageToBytes(img)
		db.updateImage(movie)
	}

	return nil
}

// isImage checks the file extension against a whitelist of common image types.
func isImage(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tif", ".tiff":
		return true
	default:
		return false
	}
}

// updateImage replaces an image in the database.
func (d *Database) updateImage(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			//if err := db.Delete(&Poster{}, movie.ImageId).Error; err != nil {
			//	return fmt.Errorf("failed to delete old image: %w", err)
			//}
			//
			//image := &Poster{Data: movie.Image, Id: movie.ImageId}
			//if err := db.Create(image).Error; err != nil {
			//	return fmt.Errorf("failed to insert new image: %w", err)
			//}

			if err := db.Model(&Poster{}).
				Where("id = ?", movie.ImageId).
				Update("image", movie.Image).Error; err != nil {

				return fmt.Errorf("failed to update image: %w", err)
			}

			return nil
		},
	)

	// Check transaction error
	if err != nil {
		return err
	}

	return nil
}

func imageToBytes(img image.Image) []byte {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return nil
	}
	return buf.Bytes()
}

func resizeImage(path string) error {
	fmt.Printf("processing start for image: %s\n", path)

	// Load the image.
	img, err := imaging.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}

	// Resize – ignore aspect ratio (force exact dimensions).
	resized := imaging.Resize(img, targetWidth, targetHeight, imaging.Lanczos)

	// Preserve the original format when saving.
	err = imaging.Save(resized, path)
	if err != nil {
		return fmt.Errorf("failed to save image %s: %w", path, err)
	}

	fmt.Printf("processing done for image: %s\n", path)

	return nil
}
