package otto

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
)

func DownloadURL(theurl string) (err error) {

	// Make the Request
	resp, err := http.Get(theurl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Make sure the URL returns 200 OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the file
	//r, _ := http.NewRequest("GET", url, nil)
	// r.URL.Path
	parsed, err := url.ParseRequestURI(theurl)
	if err != nil {
		log.Fatal(err)
	}
	filename := path.Base(parsed.Path)
	fmt.Println("Filename:", filename)
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy contents to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
