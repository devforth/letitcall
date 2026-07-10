package httpapi

import (
	"errors"
	"io/fs"
	"net/http"
)

func (s *Server) serveAvatar(w http.ResponseWriter, r *http.Request) {
	file, err := s.avatars.Open(r.PathValue("filename"))
	if errors.Is(err, fs.ErrNotExist) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		internalError(w, err, "open avatar")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		internalError(w, err, "inspect avatar")
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
