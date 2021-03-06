package fetcher

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-git/go-billy/v5"

	"github.com/tie/modpacker/modpacker"
)

func curseURL(projectID, fileID int) string {
	u := "https://addons-ecs.forgesvc.net/api/v2/addon/%d/file/%d/download-url"
	return fmt.Sprintf(u, projectID, fileID)
}

func curseCachePath(fs billy.Basic, m modpacker.Mod) (dir, base string) {
	projectID := strconv.Itoa(m.ProjectID)
	fileID := strconv.Itoa(m.FileID)
	return fs.Join("curse", projectID), fileID
}

func curseFetchURL(c *http.Client, m modpacker.Mod) (string, error) {
	u := curseURL(m.ProjectID, m.FileID)
	resp, err := c.Get(u)
	if err != nil {
		return "", err
	}
	r := resp.Body
	defer func() {
		err := r.Close()
		if err != nil {
			log.Printf("close %q: %+v", u, err)
		}
	}()

	// Don’t read responses larger than 1KiB.
	lr := io.LimitReader(r, 1024)

	var b strings.Builder
	if _, err := io.Copy(&b, lr); err != nil {
		return "", err
	}
	rawurl := b.String()
	return rawurl, nil
}
