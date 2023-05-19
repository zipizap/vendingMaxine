package web

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/labstack/echo/v4"
	"gopkg.in/fsnotify.v1"
)

type TemplateRenderer struct {
	templates *template.Template
	lock      sync.RWMutex
}

// Render renders the template with the given name
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.templates.ExecuteTemplate(w, name, data)
}

// Watch for changes in template files and reload them
func watchTemplates(watcher *fsnotify.Watcher, templatesDir string, renderer *TemplateRenderer) {
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".html") {
			err = watcher.Add(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		slog.Error(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				slog.Info("Reloading templates...")
				renderer.lock.Lock()
				renderer.templates = template.Must(template.ParseGlob(filepath.Join(templatesDir, "*.html")))
				renderer.lock.Unlock()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Errorf("Watcher error:", err)
		}
	}
}
