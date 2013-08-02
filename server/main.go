package main

import (
	"fmt"
	"github.com/calmh/mole/scrypto"
	"net/http"
	"strings"
)

const (
	errorMissingHeader = -1001 - iota
	errorIncorrectAHFormat
	errorIncorrectAH
)

var key *scrypto.Key

func init() {
	keystr := "ACDSHcyfzZp0KOykolRKWg55DXKuWWR645mbJzH+EKqZKAAgNTEvPtCr2p+eCz8wfsvwaS1UbULXXQLQ1oumpUpmG1IAIBU8KJFeu9BJ8TzLOFPpx/QpIpa4fXNnXYlmgH+dnN+s"
	key = scrypto.KeyFromString(keystr)
}

func httpError(w http.ResponseWriter, httpCode, intCode int, message string) {
	http.Error(w, fmt.Sprintf(`{"code": %d, "message": "%s"}`, intCode, message), httpCode)
}

func authenticated(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Header)
		ah, ok := r.Header["X-Mole-Authentication"]
		if !ok {
			httpError(w, 403, errorMissingHeader, "Missing authentication header")
			return
		}

		fs := strings.Split(ah[0], ";")
		if len(fs) != 2 {
			httpError(w, 403, errorIncorrectAHFormat, "Incorrect authentication header format")
			return
		}

		cookie := fs[0]
		sign := fs[1]

		ok, e := key.Verify([]byte(cookie), sign)
		if !ok || e != nil {
			httpError(w, 403, errorIncorrectAH, "Incorrect authentication signature")
			return
		}

		fn(w, r)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Header)
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	fmt.Println(key.Sign([]byte("hello")))
	http.HandleFunc("/", authenticated(handler))
	http.ListenAndServe(":8080", nil)
}
