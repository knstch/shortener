package getmethod

import (
	. "github.com/knstch/shortener_url/internal/app/postMethod"
)

func GetMethod(url string, URLstorage Storage) string {
	for k, v := range URLstorage.Data {
		if v == url {
			return k
		}
	}
	return ""
}
