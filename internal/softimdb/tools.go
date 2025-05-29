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
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/nfnt/resize"

	"github.com/hultan/dialog"
)

func cleanString(text string) string {
	text = html.EscapeString(text)
	return text
}

func reportError(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	_, _ = dialog.Title(applicationTitle).Text(err.Error()).
		ErrorIcon().OkButton().Show()
}

// openInNemo : Opens a new nemo instance with the specified folder opened
func openInNemo(path string) {
	openProcess("nemo", path)
}

// openProcess : Opens a command with the specified arguments
func openProcess(command string, args ...string) {
	cmd := exec.Command(command, args...)
	// This detaches the process from the parent, so that
	// if SoftIMDB quits, Nemo remains open.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Detach from parent process group
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
		_, _ = dialog.Title("Failed to open file!").Text("Is the NAS unlocked?").WarningIcon().OkButton().Show()
		return "", err
	}
	files, err := f.Readdirnames(0)
	if err != nil {
		return "", err
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

func getSortBy(sortBy, sortOrder string) string {
	return fmt.Sprintf("%s %s", sortBy, sortOrder)
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
	var i uint = 0
	for i < children.Length() {
		widget, _ := children.NthData(i).(*gtk.Widget)
		list.Remove(widget)
		i++
	}
}

// ClearFlowBox : Clears a gtk.FlowBox
func clearFlowBox(list *gtk.FlowBox) {
	children := list.GetChildren()
	if children == nil {
		return
	}
	var i uint = 0
	for i < children.Length() {
		widget, _ := children.NthData(i).(*gtk.Widget)
		list.Remove(widget)
		i++
	}
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
	imgResized := resize.Resize(imageWidth, imageHeight, img, resize.Lanczos2)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, imgResized, &jpeg.Options{Quality: 85})
	if err != nil {
		reportError(err)
		return nil
	}
	return buf.Bytes()
}

func containsI(a, b string) bool {
	return strings.Contains(strings.ToLower(b), strings.ToLower(a))
}

func equalsI(a, b string) bool {
	return strings.ToLower(b) == strings.ToLower(a)
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

func saveMoviePoster(title string, poster []byte) (string, error) {
	home, _ := os.UserHomeDir()
	dir := path.Join(home, "Downloads")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	title = cleanTitle(title)

	// Define the full path for the image
	filePath := filepath.Join(dir, fmt.Sprintf("%s.jpg", title))

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Decode the image from the byte slice
	img, _, err := image.Decode(bytes.NewReader(poster))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	// Set the JPEG quality (0-100)
	jpegOptions := &jpeg.Options{Quality: 100}

	// Encode the image to the file
	if err := jpeg.Encode(file, img, jpegOptions); err != nil {
		return "", fmt.Errorf("failed to encode image: %v", err)
	}

	return filePath, nil
}

func cleanTitle(title string) string {
	t := strings.Replace(title, " ", "_", -1)
	t = strings.Replace(t, "/", "-", -1)
	return t
}
