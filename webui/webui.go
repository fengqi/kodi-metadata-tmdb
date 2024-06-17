package webui

import (
	"fengqi/kodi-metadata-tmdb-cli/config"
	"fengqi/kodi-metadata-tmdb-cli/utils"
	"net/http"
)

func RunWebui(config *config.Config) {
	if !config.Webui.Enable {
		return
	}

	// Routes
	http.HandleFunc("/", hello)

	// Start server
	utils.Logger.InfoF("webui started on http://%s", config.Webui.Listen)
	if err := http.ListenAndServe(config.Webui.Listen, nil); err != nil {
		panic(err)
	}
}

// Handler
func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}
