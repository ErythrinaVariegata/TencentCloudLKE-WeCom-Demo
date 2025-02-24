package cron

import (
	"sync"
	"time"
)

const (
	WxTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
)

var (
	accessToken string
	tokenMutex  sync.RWMutex
	timer       *time.Timer
)

type AccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
