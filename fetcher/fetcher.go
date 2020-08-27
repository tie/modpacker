package fetcher

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/crypto/sha3"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"

	"github.com/tie/modpacker/modpacker"
)

var (
	ErrSumsMismatch     = errors.New("checksum mismatch")
	ErrUnknownModMethod = errors.New("unknown mod method")
)

type (
	cacheFunc func(billy.Basic, modpacker.Mod) (dir, base string)
	fetchFunc func(*http.Client, modpacker.Mod) (string, error)
)

func httpCachePath(fs billy.Basic, m modpacker.Mod) (dir, base string) {
	// Good enough is good enough.
	sum := sha1.Sum([]byte(m.File))
	hex := fmt.Sprintf("%x", sum)
	return "http", fs.Join(hex[:2], hex)
}

func httpFetchURL(c *http.Client, m modpacker.Mod) (string, error) {
	return m.File, nil
}

type Fetcher struct {
	Files  billy.Filesystem
	Client *http.Client
}

func (dl *Fetcher) Sums(m modpacker.Mod) ([]string, error) {
	switch m.Method {
	case modpacker.MethodCurse:
		return dl.sumsGeneric(m, curseCachePath, curseFetchURL)
	case modpacker.MethodOptifine:
		return dl.sumsGeneric(m, optifineCachePath, optifineFetchURL)
	case modpacker.MethodHTTP:
		return dl.sumsGeneric(m, httpCachePath, httpFetchURL)
	case modpacker.MethodFile:
		// TODO should we check files integrity?
		return nil, nil
	}
	return nil, ErrUnknownModMethod
}

func (dl *Fetcher) Cache(m modpacker.Mod) error {
	switch m.Method {
	case modpacker.MethodCurse:
		return dl.cacheGeneric(m, curseCachePath, curseFetchURL)
	case modpacker.MethodOptifine:
		return dl.cacheGeneric(m, optifineCachePath, optifineFetchURL)
	case modpacker.MethodHTTP:
		return dl.cacheGeneric(m, httpCachePath, httpFetchURL)
	case modpacker.MethodFile:
		return nil
	}
	return ErrUnknownModMethod
}

func (dl *Fetcher) Open(m modpacker.Mod) (billy.File, error) {
	switch m.Method {
	case modpacker.MethodCurse:
		return dl.downloadGeneric(m, curseCachePath, curseFetchURL)
	case modpacker.MethodOptifine:
		return dl.downloadGeneric(m, optifineCachePath, optifineFetchURL)
	case modpacker.MethodHTTP:
		return dl.downloadGeneric(m, httpCachePath, httpFetchURL)
	case modpacker.MethodFile:
		path := filepath.FromSlash(m.File)
		f, err := osfs.New("").Open(path)
		if err != nil {
			return nil, err
		}
		return f, err
	}
	return nil, ErrUnknownModMethod
}

func (dl *Fetcher) sumsGeneric(m modpacker.Mod, cachePath cacheFunc, fetchURL fetchFunc) ([]string, error) {
	err := dl.cacheGeneric(m, cachePath, fetchURL)
	if err != nil {
		return nil, err
	}
	dir, base := cachePath(dl.Files, m)
	return dl.readSums(dir, base)
}

func (dl *Fetcher) cacheGeneric(m modpacker.Mod, cachePath cacheFunc, fetchURL fetchFunc) error {
	dir, base := cachePath(dl.Files, m)
	_, err := dl.statData(dir, base)
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	rawurl, err := fetchURL(dl.Client, m)
	if err != nil {
		return err
	}
	return dl.downloadFile(rawurl, dir, base)
}

func (dl *Fetcher) downloadGeneric(m modpacker.Mod, cachePath cacheFunc, fetchURL fetchFunc) (billy.File, error) {
	dir, base := cachePath(dl.Files, m)
	f, err := dl.getFile(dir, base, m.Sums)
	if !errors.Is(err, os.ErrNotExist) {
		return f, err
	}
	rawurl, err := fetchURL(dl.Client, m)
	if err != nil {
		return nil, err
	}
	if err := dl.downloadFile(rawurl, dir, base); err != nil {
		return nil, err
	}
	return dl.getFile(dir, base, m.Sums)
}

func (dl *Fetcher) getFile(dir, base string, sums []string) (billy.File, error) {
	err := dl.verifySums(sums, dir, base)
	if err != nil {
		return nil, err
	}
	var f billy.File
	err = dl.withData(dir, base, os.O_RDONLY, func(ff billy.File) error {
		f = ff
		return nil
	})
	return f, err
}

func (dl *Fetcher) readSums(dir, base string) ([]string, error) {
	sums := []string{}
	err := dl.withSums(dir, base, os.O_RDONLY, func(f billy.File) error {
		defer func() {
			cerr := f.Close()
			if cerr != nil {
				log.Printf("close %q: %+v", f.Name(), cerr)
			}
		}()
		s := bufio.NewScanner(f)
		for s.Scan() {
			sum := s.Text()
			sums = append(sums, sum)
		}
		return s.Err()
	})
	if err != nil {
		return nil, err
	}
	return sums, nil
}

func (dl *Fetcher) verifySums(sums []string, dir, base string) error {
	l := len(sums)
	if l <= 0 {
		return nil
	}
	sumsMap := make(map[string]struct{}, l)
	err := dl.withSums(dir, base, os.O_RDONLY, func(f billy.File) error {
		defer func() {
			cerr := f.Close()
			if cerr != nil {
				log.Printf("close %q: %+v", f.Name(), cerr)
			}
		}()
		s := bufio.NewScanner(f)
		for s.Scan() {
			sum := s.Text()
			sumsMap[sum] = struct{}{}
		}
		return s.Err()
	})
	if err != nil {
		return err
	}
	for _, sum := range sums {
		if _, ok := sumsMap[sum]; ok {
			continue
		}
		return ErrSumsMismatch
	}
	return nil
}

func (dl *Fetcher) downloadFile(rawurl, dir, base string) error {
	hashNames := []string{
		"md5",
		"sha1",
		"sha256",
		"keccak256",
	}
	hashes := []hash.Hash{
		md5.New(),
		sha1.New(),
		sha256.New(),
		sha3.New256(),
	}
	nhashes := len(hashes)
	flags := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	err := dl.withData(dir, base, flags, func(f billy.File) (err error) {
		defer func() {
			cerr := f.Close()
			if err == nil {
				err = cerr
			}
		}()
		l := nhashes
		ww := make([]io.Writer, l+1)
		for i, h := range hashes {
			ww[i] = h
		}
		ww[l] = f
		w := io.MultiWriter(ww...)
		return dl.fetchFile(w, rawurl)
	})
	if err != nil {
		return err
	}
	sums := make([]string, nhashes)
	for i, name := range hashNames {
		sums[i] = fmt.Sprintf("%s:%x", name, hashes[i].Sum(nil))
	}
	return dl.writeSums(dir, base, sums)
}

func (dl *Fetcher) fetchFile(w io.Writer, rawurl string) error {
	resp, err := dl.Client.Get(rawurl)
	if err != nil {
		return err
	}
	r := resp.Body
	defer func() {
		err := r.Close()
		if err != nil {
			log.Printf("close %q: %+v", rawurl, err)
		}
	}()
	if _, err := io.Copy(w, r); err != nil {
		return err
	}
	return nil
}

func (dl *Fetcher) writeSums(dir, base string, sums []string) error {
	flags := os.O_WRONLY | os.O_APPEND | os.O_CREATE
	return dl.withSums(dir, base, flags, func(f billy.File) (err error) {
		defer func() {
			cerr := f.Close()
			if err == nil {
				err = cerr
			}
		}()
		w := bufio.NewWriter(f)
		defer func() {
			ferr := w.Flush()
			if err == nil {
				err = ferr
			}
		}()
		for _, sum := range sums {
			_, err = fmt.Fprintf(w, "%s\r\n", sum)
			if err != nil {
				break
			}
		}
		return err
	})
}

func (dl *Fetcher) statData(dir, base string) (os.FileInfo, error) {
	return dl.statFile(dir, base, "dat")
}

func (dl *Fetcher) statFile(dir, base, ext string) (os.FileInfo, error) {
	fname := fmt.Sprintf("%s.%s", base, ext)
	fpath := dl.Files.Join(dir, fname)
	return dl.Files.Stat(fpath)
}

func (dl *Fetcher) withData(dir, base string, flag int, fn func(billy.File) error) error {
	return dl.withFile(dir, base, "dat", flag, fn)
}

func (dl *Fetcher) withSums(dir, base string, flag int, fn func(billy.File) error) error {
	return dl.withFile(dir, base, "sum", flag, fn)
}

func (dl *Fetcher) withFile(dir, base, ext string, flag int, fn func(billy.File) error) error {
	if err := dl.Files.MkdirAll(dir, 0755); err != nil {
		return err
	}
	fname := fmt.Sprintf("%s.%s", base, ext)
	fpath := dl.Files.Join(dir, fname)
	f, err := dl.Files.OpenFile(fpath, flag, 0644)
	if err != nil {
		return err
	}
	return fn(f)
}
