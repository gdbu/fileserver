package fileserver

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gdbu/atoms"
	"github.com/gdbu/filecacher"
)

var (
	noOptions       = []string{""}
	webpOptions     = []string{".webp", ""}
	gzipOptions     = []string{".gz", ""}
	webpGZipOptions = []string{".webp.gz", ".webp", ".gz", ""}
)

// New will return a new instance of fileserver
func New(dir string) (fp *FileServer, err error) {
	var f FileServer
	f.fc = filecacher.New(dir)
	fp = &f
	return
}

// FileServer manages the serving of files
type FileServer struct {
	fc *filecacher.FileCacher

	closed atoms.Bool
}

// serve will serve a file to a http.ResponseWriter
func (f *FileServer) serve(key string, w http.ResponseWriter) (err error) {
	ext := filepath.Ext(key)
	// Attempt to read the file with the given key
	err = f.fc.Read(key, func(r io.Reader) (err error) {
		// Set HTTP headers for file
		setHeaders(w, key, ext)
		// Write file bytes to HTTP response body
		_, err = io.Copy(w, r)
		return
	})

	return
}

// Serve will serve a file
func (f *FileServer) Serve(key string, res http.ResponseWriter, req *http.Request) (err error) {
	// Iterate through all key options
	// Note: See options var block within fileserver.go and/or getOptions func for more context
	for _, option := range getOptions(req) {
		// Attempt to serve current option
		if err = f.serve(getKey(key, option), res); err == filecacher.ErrFileNotFound {
			// Option has not been found, continue to the next option
			continue
		}

		// We've either received a nil error (hopefully!) or an error which is not ErrFileNotFound, return
		return
	}

	// If we made it to the end without finding the file
	return filecacher.ErrFileNotFound
}

// Close will close an instance of fileserver
func (f *FileServer) Close() (err error) {
	if !f.closed.Set(true) {
		return errors.New("is already closed")
	}

	return f.fc.Close()
}
