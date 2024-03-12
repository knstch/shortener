// Пакет certconstructor отвечает за создание автосертификата.
package certconstructor

import "golang.org/x/crypto/acme/autocert"

// NewCert возвращает TLS-сертификат.
func NewCert(website ...string) *autocert.Manager {
	return &autocert.Manager{
		Cache:      autocert.DirCache("cache-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(website...),
	}
}
