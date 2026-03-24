package httpclient

import (
	"io"
	"net/http"
	"os"
)

const defaultTestServiceURL = "http://localhost:8000"

func CallHelloAPI() (string, error) {
	res, err := http.Get(getTestServiceURL())

	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	return string(body), nil
}

func getTestServiceURL() string {
	if url := os.Getenv("TEST_SERVICE_URL"); url != "" {
		return url
	}

	return defaultTestServiceURL
}
