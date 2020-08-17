package fetcher

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/net/html"

	"github.com/go-git/go-billy/v5"

	"github.com/andybalholm/cascadia"

	"github.com/tie/modpacker/modpacker"
)

var ErrUnexpectedNode = errors.New("unexpected html node")

var optifineSel = cascadia.MustCompile("#Download > a")

func optifineURL(file string) string {
	u := "https://optifine.net/adloadx?f=%s"
	return fmt.Sprintf(u, url.QueryEscape(file))
}

func optifineCachePath(fs billy.Basic, m modpacker.Mod) (dir, base string) {
	return "optifine", m.File
}

func optifineFetchURL(c *http.Client, m modpacker.Mod) (string, error) {
	u := optifineURL(m.File)
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

	// Don’t read HTML pages larger than 1MiB.
	lr := io.LimitReader(r, 1024*1024)

	root, err := html.Parse(lr)
	if err != nil {
		return "", err
	}
	n := optifineSel.MatchFirst(root)
	if n.Type != html.ElementNode {
		err := ErrUnexpectedNode
		return "", err
	}
	if n.Namespace != "" || n.Data != "a" {
		err := ErrUnexpectedNode
		return "", err
	}
	for _, attr := range n.Attr {
		if attr.Namespace != "" {
			continue
		}
		if attr.Key != "href" {
			continue
		}
		rawurl := fmt.Sprintf("https://optifine.net/%s", attr.Val)
		return rawurl, nil
	}
	return "", ErrUnexpectedNode
}
