package web

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var distFS embed.FS

func RegisterStatic(r *gin.Engine) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	// Static assets (e.g. /assets/*.js, /assets/*.css from Vite build)
	// Serve dist/assets at /assets — index.html references /assets/index-xxx.js
	assetsSub, err := fs.Sub(sub, "assets")
	if err != nil {
		panic(err)
	}
	r.StaticFS("/assets", http.FS(assetsSub))

	// index.html for SPA (/, and any unknown route fall back)
	r.GET("/", func(c *gin.Context) {
		serveIndex(c, sub)
	})

	r.NoRoute(func(c *gin.Context) {
		// Let /api/** be handled earlier; only SPA fallback here
		if len(c.Request.URL.Path) >= 5 && c.Request.URL.Path[:5] == "/api/" {
			c.Status(http.StatusNotFound)
			return
		}
		serveIndex(c, sub)
	})
}

func serveIndex(c *gin.Context, sub fs.FS) {
	file, err := sub.Open("index.html")
	if err != nil {
		c.String(http.StatusNotFound, "index.html not found")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to read index.html")
		return
	}

	http.ServeContent(c.Writer, c.Request, "index.html", fsStatModTime(file), bytes.NewReader(data))
}

// fsStatModTime gets modification time from fs.File if possible, otherwise returns zero.
func fsStatModTime(f fs.File) (t time.Time) {
	if info, err := f.Stat(); err == nil {
		return info.ModTime()
	}
	return
}
