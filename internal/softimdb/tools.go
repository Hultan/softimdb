package softimdb

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/nfnt/resize"
	"html"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"

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
func openProcess(command string, args ...string) string {

	cmd := exec.Command(command, args...)
	// Forces the new process to detach from the GitDiscover process
	// so that it does not die when GitDiscover dies
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	err = cmd.Process.Release()
	if err != nil {
		log.Println(err)
	}

	return string(output)
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
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	files, err := f.Readdirnames(0)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if strings.HasSuffix(file, "mkv") {
			return file, nil
		}
	}
	return "", nil
}

func getSortBy() string {
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

// resizeImage resizes the image to 190x280 and converts it to a PNG file
func resizeImage(imgData []byte) []byte {
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		reportError(err)
		return nil
	}
	imgResized := resize.Resize(imageWidth, imageHeight, img, resize.Lanczos2)
	buf := new(bytes.Buffer)
	err = png.Encode(buf, imgResized)
	if err != nil {
		reportError(err)
		return nil
	}
	return buf.Bytes()
}
