package softimdb

import (
	"errors"
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"html"
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
	dialog.Title(applicationTitle).Text(err.Error()).
		ErrorIcon().OkButton().Show()
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
