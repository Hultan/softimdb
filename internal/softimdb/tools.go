package softimdb

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"image"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/disintegration/imaging"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/hultan/dialog"
)

func cleanString(text string) string {
	text = html.EscapeString(text)
	return text
}

func reportError(err error) {
	if err == nil {
		return
	}

	_, _ = fmt.Fprintln(os.Stderr, err)

	// Always make sure that the dialog is called from the main thread
	glib.IdleAdd(func() {
		_, _ = dialog.Title(applicationTitle).
			Text("An unkown error occured!").
			ExtraExpand(err.Error()).
			ExtraHeight(80).
			ErrorIcon().OkButton().Show()
	})
}

// openInNemo opens a new nemo instance with the specified folder opened
func openInNemo(path string) {
	openProcess("nemo", path)
}

// openProcess opens a command with the specified arguments
func openProcess(command string, args ...string) {
	cmd := exec.Command(command, args...)
	// This detaches the process from the parent so that
	// if SoftIMDB quits, Nemo remains open.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Detach from the parent process group
	}

	// Start the command without waiting for it to finish
	err := cmd.Start()
	if err != nil {
		log.Println("Failed to start process:", err)
		return
	}

	// Release the process so it doesn't become a zombie
	err = cmd.Process.Release()
	if err != nil {
		log.Println("Failed to release process:", err)
	}
}

func getIdFromUrl(url string) (string, error) {
	// Get the IMDB id from the URL.
	// Starts with tt and ends with 7 or 8 digits.
	re := regexp.MustCompile(`tt\d{7,8}`)
	matches := re.FindAll([]byte(url), -1)
	if len(matches) == 0 {
		err := errors.New("invalid imdb URL")
		return "", err
	}
	return string(matches[0]), nil
}

// https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func findMovieFile(path string) (string, error) {
	// Movie file extensions to look for
	var ext = []string{
		"mp4",
		"mkv",
		"avi",
		"webm",
	}

	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open path: %s", path)
	}

	files, err := f.Readdirnames(0)
	if err != nil {
		return "", fmt.Errorf("failed to read dir names: %s", path)
	}

	for _, file := range files {
		lower := strings.ToLower(file)
		for _, extension := range ext {
			if strings.HasSuffix(lower, extension) {
				return file, nil
			}
		}
	}

	return "", nil
}

func getSortBy(sort Sort) string {
	return fmt.Sprintf("%s %s", sort.by, sort.order)
}

func getEntryText(entry *gtk.Entry) string {
	text, err := entry.GetText()
	if err != nil {
		return ""
	}
	return text
}

// clearListBox : Clears a gtk.ListBox
func clearListBox(list *gtk.ListBox) {
	children := list.GetChildren()
	if children == nil {
		return
	}

	children.Foreach(func(item interface{}) {
		if widget, ok := item.(gtk.IWidget); ok {
			list.Remove(widget)
			if w, ok := widget.(*gtk.Widget); ok {
				w.Destroy()
			}
		}
	})
}

// ClearFlowBox clears a gtk.FlowBox
func clearFlowBox(list *gtk.FlowBox) {
	children := list.GetChildren()
	if children == nil {
		return
	}

	children.Foreach(func(item interface{}) {
		if widget, ok := item.(gtk.IWidget); ok {
			list.Remove(widget)
			if w, ok := widget.(*gtk.Widget); ok {
				w.Destroy()
			}
		}
	})
}

// getCorrectImageSize makes sure that the size of the image is 190x280 and returns it
func getCorrectImageSize(fileName string) []byte {
	data, err := os.ReadFile(fileName)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}

	pix, err := gdk.PixbufNewFromBytesOnly(data)
	if err != nil {
		reportError(err)
		log.Fatal(err)
	}
	width, height := pix.GetWidth(), pix.GetHeight()
	if width != imageWidth || height != imageHeight {
		return resizeImage(data)
	}

	return data
}

// resizeImage resizes the image to 190x280 and converts it to a JPG file
func resizeImage(imgData []byte) []byte {
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		reportError(err)
		return nil
	}

	imgResized := imaging.Resize(img, imageWidth, imageHeight, imaging.Lanczos)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, imgResized, &jpeg.Options{Quality: 85})
	if err != nil {
		reportError(err)
		return nil
	}
	return buf.Bytes()
}

// doesExist checks if the file exists and is accessible.
func doesExist(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		// No error, so the file exists
		return true
	}
	if os.IsNotExist(err) {
		// The file does not exist
		return false
	}
	// Other error types (e.g., permission issues) will also return false
	return false
}

func cleanTitle(title string) string {
	replacer := strings.NewReplacer(" ", "_", "/", "-")
	return replacer.Replace(title)
}

func saveMoviePoster(title string, poster []byte) (string, error) {
	filePath, err := getPosterFilePath(title)
	if err != nil {
		return "", fmt.Errorf("failed to get poster file path: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(poster))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	if err := saveJPEGImage(filePath, img, 100); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return filePath, nil
}

func getPosterFilePath(title string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	dir := filepath.Join(home, "Downloads")

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	filename := fmt.Sprintf("%s.jpg", cleanTitle(title))

	return filepath.Join(dir, filename), nil
}

func saveJPEGImage(filePath string, img image.Image, quality int) (err error) {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}

	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close file %s: %w", filePath, cerr)
		}
	}()

	jpegOptions := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(file, img, jpegOptions); err != nil {
		return fmt.Errorf("failed to encode JPEG to %s: %w", filePath, err)
	}

	return nil
}
