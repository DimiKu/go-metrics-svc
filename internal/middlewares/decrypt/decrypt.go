package decrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"net/http"
)

func DecryptMiddleware(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			encrypted, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			decrypted, err := rsa.DecryptOAEP(
				sha256.New(),
				rand.Reader,
				privateKey,
				encrypted,
				nil,
			)
			if err != nil {
				http.Error(w, "Decryption failed", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.ContentLength = int64(len(decrypted))

			next.ServeHTTP(w, r)
		})
	}
}
