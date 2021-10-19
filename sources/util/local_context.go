package util

import "os"

var LocalContext string

func init() {
	hn, err := os.Hostname()

	if err != nil {
		LocalContext = "UNKNOWN"
	}

	LocalContext = hn
}
