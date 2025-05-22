package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/meesooqa/files2tg/app/finder"
	"github.com/meesooqa/files2tg/app/job"
)

func (s *Server) getIndexPageCtrl(w http.ResponseWriter, r *http.Request) {
	statuses := s.JobQueue.GetJobsStatuses()
	s.templates.Execute(w, statuses)
}

func (s *Server) getStatusPageCtrl(w http.ResponseWriter, r *http.Request) {
	statuses := s.JobQueue.GetJobsStatuses()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		http.Error(w, "JSON Encoding error", http.StatusInternalServerError)
	}
}

func (s *Server) send(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.addJobsChunk()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) addJobsChunk() {
	// TODO move "dir" to config
	dir := "var/files"

	videoInfoProvider := finder.NewVideoInfoProvider()
	filesProvider := finder.NewProvider(videoInfoProvider)
	files, err := filesProvider.GetListFilesSorted(dir, ".")
	if err != nil {
		fmt.Printf("GetListFilesSorted: %v\n", err)
		return
	}

	s.JobQueue.Clear()
	for i, file := range files {
		// fmt.Printf("  %s — %s\n", file.Name, file.ModTime.Format(time.RFC3339))
		// jobId := uuid.New().String()
		jobId := fmt.Sprintf("%s-%s", file.Name, file.ModTime.Format(time.RFC3339))

		// TODO stars 10
		stars := 10
		if i%10 == 0 {
			stars = 0
		}

		s.JobQueue.AddJob(job.SendVideoJob{
			BaseJob:        job.BaseJob{ID: jobId},
			TelegramClient: s.TelegramClient,
			File:           file,
			Stars:          stars,
		})
	}
}
