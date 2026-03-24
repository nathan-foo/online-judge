package service

import (
	"github.com/nathan-foo/online-judge/gateway/internal/httpclient"
)

func GetHello() (string, error) {
	return httpclient.CallHelloAPI()
}
