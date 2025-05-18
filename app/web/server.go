package web

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/meesooqa/files2tg/app/job"
	"github.com/meesooqa/files2tg/app/send"
)

type Server struct {
	TemplLocationPattern string
	TemplStaticLocation  string
	JobQueue             job.JobQueuer
	TelegramClient       send.Client

	httpServer *http.Server
	templates  *template.Template
}

func (s *Server) Run(ctx context.Context, port int) {
	log.Printf("[INFO] starting server on port %d", port)

	if s.TemplLocationPattern == "" {
		s.TemplLocationPattern = "app/web/templates/*"
	}
	log.Printf("[DEBUG] loading templates from %s", s.TemplLocationPattern)
	s.templates = template.Must(template.ParseGlob(s.TemplLocationPattern))

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           s.router(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second, // HTTPResponseTimeout
		IdleTimeout:       60 * time.Second,
	}

	err := s.httpServer.ListenAndServe()
	log.Printf("[WARN] http server terminated, %s", err)
}

func (s *Server) router() http.Handler {
	mux := http.NewServeMux()

	// Static
	if s.TemplStaticLocation == "" {
		s.TemplStaticLocation = "app/web/static"
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.TemplStaticLocation))))

	// Route
	mux.HandleFunc("/", s.getIndexPageCtrl)
	mux.HandleFunc("/status", s.getStatusPageCtrl)
	mux.HandleFunc("/send", s.send)

	return mux
}
