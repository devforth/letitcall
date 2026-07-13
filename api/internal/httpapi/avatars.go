package httpapi

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
)

func (s *Server) serveAvatar(w http.ResponseWriter, r *http.Request) {
	s.serveImage(w, r, "avatar", s.avatars.Open)
}

func (s *Server) serveLogo(w http.ResponseWriter, r *http.Request) {
	s.serveImage(w, r, "logo", s.logos.Open)
}

func (s *Server) serveImage(w http.ResponseWriter, r *http.Request, noun string, open func(string) (*os.File, error)) {
	file, err := open(r.PathValue("filename"))
	if errors.Is(err, fs.ErrNotExist) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		internalError(w, err, "open "+noun)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		internalError(w, err, "inspect "+noun)
		return
	}
	if !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}
