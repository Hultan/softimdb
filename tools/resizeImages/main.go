// resize_images.go
//
// Resize every image in /home/per/temp/softimdb to 190x280 pixels
// and store the result in /home/per/temp/softimdb/fixed.
//
// Supported formats: JPEG, PNG, GIF, BMP, TIFF.
// Requires Go 1.20+ (or any recent version) and the imaging package.
//
// To build:
//
//	go get github.com/disintegration/imaging
//	go build -o resize_images resize_images.go
//
// To run:
//
//	./resize_images
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// ---------------------------------------------------------------------------
// Configuration – change these if you ever need to reuse the program.
// ---------------------------------------------------------------------------
var (
	srcDir       = "/home/per/temp/softimdb"
	dstDir       = "/home/per/temp/softimdb/fixed"
	targetWidth  = 190
	targetHeight = 280
)

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

// ensureDst makes sure the destination directory exists.
func ensureDst() error {
	info, err := os.Stat(dstDir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dstDir, 0755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", dstDir)
	}
	return nil
}

// processFile loads an image, resizes it, and saves it to the destination.
func processFile(path string, info os.FileInfo) error {
	if info.IsDir() {
		// Skip sub‑directories (the program assumes a flat layout).
		return nil
	}
	if !isImage(info.Name()) {
		// Not an image we care about.
		return nil
	}

	fmt.Printf("processing start: %s\n", path)

	// Load the image.
	img, err := imaging.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}

	// Resize – ignore aspect ratio (force exact dimensions).
	resized := imaging.Resize(img, targetWidth, targetHeight, imaging.Lanczos)

	// Destination path (same filename, new folder).
	dstPath := filepath.Join(dstDir, info.Name())

	// Preserve the original format when saving.
	err = imaging.Save(resized, dstPath)
	if err != nil {
		return fmt.Errorf("failed to save %s: %w", dstPath, err)
	}

	fmt.Printf("processing done: %s\n", path)

	return nil
}

func main() {
	// Step 1 – make sure the output folder exists.
	if err := ensureDst(); err != nil {
		log.Fatalf("Could not prepare destination folder: %v", err)
	}

	// Step 2 – walk the source directory.
	var processed, skipped int
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			// Propagate the error; stop walking.
			return walkErr
		}
		// Skip the destination folder itself if it already exists inside srcDir.
		if path == dstDir {
			return filepath.SkipDir
		}
		if err := processFile(path, info); err != nil {
			// Log the problem but keep going.
			log.Printf("⚠️  %v", err)
			skipped++
			return nil
		}
		if isImage(info.Name()) {
			fmt.Printf("✅ %s → %s\n", info.Name(), dstDir)
			processed++
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking source directory: %v", err)
	}

	fmt.Printf("\nFinished. Processed %d images, skipped %d files.\n", processed, skipped)
}
