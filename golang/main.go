package main

import (
	"log"
	"net/http"

	"example.com/play/config"
	"example.com/play/logic"
	"example.com/play/repo/wecom/cron"
)

func main() {
	config.Init()
	cron.StartTokenRefresher(config.Config.WxCorpID, config.Config.WxAppSecret)
	http.HandleFunc("/", logic.CallbackHandler)
	log.Println("Server started on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
