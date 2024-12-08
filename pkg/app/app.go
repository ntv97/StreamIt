// Package app manages main application server.
package app

import (
        "fmt"
        "html/template"
        "log"
        "net"
        "net/http"
        "path"

        "github.com/gorilla/mux"
        "github.com/ntv97/streamit/pkg/media"

)

// App represents main application.
type App struct {
        Config    *Config
        Library   *media.Library
        Templates *template.Template
        Listener  net.Listener
        Router    *mux.Router
}

// NewApp returns a new instance of App from Config.
func NewApp(cfg *Config) (*App, error) {
        if cfg == nil {
                cfg = DefaultConfig()
        }
        a := &App{
                Config: cfg,
        }
        // Setup Library
        a.Library = media.NewLibrary()
        // Setup Listener
        ln, err := newListener(cfg.Server)
        if err != nil {
                return nil, err
        }
        a.Listener = ln
        // Setup Templates
        a.Templates = template.Must(template.ParseGlob("templates/*"))
        // Setup Router
        r := mux.NewRouter().StrictSlash(true)
        r.HandleFunc("/", a.indexHandler).Methods("GET")
        r.HandleFunc("/v/{id}.mp4", a.videoHandler).Methods("GET")
        r.HandleFunc("/t/{id}", a.thumbHandler).Methods("GET")
        r.HandleFunc("/v/{id}", a.pageHandler).Methods("GET")
        fsHandler := http.StripPrefix(
                "/static/",
                http.FileServer(http.Dir("./static/")),
        )
        r.PathPrefix("/static/").Handler(fsHandler).Methods("GET")
        a.Router = r
	return a, nil
}

// Run imports the library and starts server.
func (a *App) Run() error {
        for _, pc := range a.Config.Library {
                p := &media.Path{
                        Path:   pc.Path,
                }
                err := a.Library.AddPath(p)
                if err != nil {
                        return err
                }
                err = a.Library.Import(p)
                if err != nil {
                        return err
                }
        }
        return http.Serve(a.Listener, a.Router)
}

// HTTP handler for /
func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
        log.Printf("/")
        pl := a.Library.Playlist()
        if len(pl) > 0 {
                fmt.Println("len(pl>0), /v/+pl[0].ID", pl[0].ID);
                http.Redirect(w, r, "/v/"+pl[0].ID, 302)
        } else {
                fmt.Println("len(pl==0")
                a.Templates.ExecuteTemplate(w, "index.html", &struct {
                        Playing  *media.Video
                        Playlist media.Playlist
                }{
                        Playing:  &media.Video{ID: ""},
                        Playlist: a.Library.Playlist(),
                })
        }
}

// HTTP handler for /v/id
func (a *App) pageHandler(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        prefix, ok := vars["prefix"]
        if ok {
                id = path.Join(prefix, id)
        }
        log.Printf("/v/%s", id)
        playing, ok := a.Library.Videos[id]
        if !ok {
                a.Templates.ExecuteTemplate(w, "index.html", &struct {
                        Playing  *media.Video
                        Playlist media.Playlist
                }{
                        Playing:  &media.Video{ID: ""},
                        Playlist: a.Library.Playlist(),
                })
                return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        a.Templates.ExecuteTemplate(w, "index.html", &struct {
                Playing  *media.Video
                Playlist media.Playlist
        }{
                Playing:  playing,
                Playlist: a.Library.Playlist(),
        })
}

// HTTP handler for /v/id.mp4
func (a *App) videoHandler(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        prefix, ok := vars["prefix"]
        if ok {
                id = path.Join(prefix, id)
        }
        log.Printf("/v/%s", id)
        m, ok := a.Library.Videos[id]
        if !ok {
                return
        }
        title := m.Title
        disposition := "attachment; filename=\"" + title + ".mp4\""
        w.Header().Set("Content-Disposition", disposition)
        w.Header().Set("Content-Type", "video/mp4")
        http.ServeFile(w, r, m.Path)
}

// HTTP handler for /t/id
func (a *App) thumbHandler(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        prefix, ok := vars["prefix"]
        if ok {
                id = path.Join(prefix, id)
        }
        log.Printf("/t/%s", id)
        m, ok := a.Library.Videos[id]
        if !ok {
                return
        }
        w.Header().Set("Cache-Control", "public, max-age=7776000")
        if m.ThumbType == "" {
                w.Header().Set("Content-Type", "image/jpeg")
                http.ServeFile(w, r, "static/defaulticon.jpg")
        } else {
                w.Header().Set("Content-Type", m.ThumbType)
                w.Write(m.Thumb)
        }
}
