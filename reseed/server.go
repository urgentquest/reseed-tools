package reseed

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-i2p/onramp"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	throttled "github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store"
)

const (
	I2pUserAgent = "Wget/1.11.4"
)

type Server struct {
	*http.Server

	Reseeder  *ReseederImpl
	Blacklist *Blacklist

	ServerListener net.Listener

	// I2P Listener
	Garlic      *onramp.Garlic
	I2PListener net.Listener

	// Tor Listener
	OnionListener net.Listener
	Onion         *onramp.Onion

	RequestRateLimit int
	WebRateLimit     int
	acceptables      map[string]time.Time
}

func NewServer(prefix string, trustProxy bool) *Server {
	config := &tls.Config{
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		CurvePreferences: []tls.CurveID{tls.CurveP384, tls.CurveP521}, // default CurveP256 removed
	}
	h := &http.Server{TLSConfig: config}
	server := Server{Server: h, Reseeder: nil}

	th := throttled.RateLimit(throttled.PerHour(4), &throttled.VaryBy{RemoteAddr: true}, store.NewMemStore(200000))
	thw := throttled.RateLimit(throttled.PerHour(30), &throttled.VaryBy{RemoteAddr: true}, store.NewMemStore(200000))

	middlewareChain := alice.New()
	if trustProxy {
		middlewareChain = middlewareChain.Append(proxiedMiddleware)
	}

	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write(nil); nil != err {
			log.Println(err)
		}
	})

	mux := http.NewServeMux()
	mux.Handle("/", middlewareChain.Append(disableKeepAliveMiddleware, loggingMiddleware, thw.Throttle, server.browsingMiddleware).Then(errorHandler))
	mux.Handle(prefix+"/i2pseeds.su3", middlewareChain.Append(disableKeepAliveMiddleware, loggingMiddleware, verifyMiddleware, th.Throttle).Then(http.HandlerFunc(server.reseedHandler)))
	server.Handler = mux

	return &server
}

// See use of crypto/rand on:
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // 52 possibilities
	letterIdxBits = 6                                                      // 6 bits to represent 64 possibilities / indexes
	letterIdxMask = 1<<letterIdxBits - 1                                   // All 1-bits, as many as letterIdxBits
)

func SecureRandomAlphaString() string {
	length := 16
	result := make([]byte, length)
	bufferSize := int(float64(length) * 1.3)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			randomBytes = SecureRandomBytes(bufferSize)
		}
		if idx := int(randomBytes[j%length] & letterIdxMask); idx < len(letterBytes) {
			result[i] = letterBytes[idx]
			i++
		}
	}
	return string(result)
}

// SecureRandomBytes returns the requested number of bytes using crypto/rand
func SecureRandomBytes(length int) []byte {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatal("Unable to generate random bytes")
	}
	return randomBytes
}

//

func (srv *Server) Address() string {
	addrs := make(map[string]string)
	if srv.I2PListener != nil {
		addrs["i2p"] = srv.I2PListener.Addr().String()
	}
	if srv.OnionListener != nil {
		addrs["onion"] = srv.OnionListener.Addr().String()
	}
	if srv.Server != nil {
		addrs["tcp"] = srv.Server.Addr
	}
	return fmt.Sprintf("%v", addrs)
}

func (srv *Server) Acceptable() string {
	if srv.acceptables == nil {
		srv.acceptables = make(map[string]time.Time)
	}
	if len(srv.acceptables) > 50 {
		for val := range srv.acceptables {
			srv.CheckAcceptable(val)
		}
		for val := range srv.acceptables {
			if len(srv.acceptables) < 50 {
				break
			}
			delete(srv.acceptables, val)
		}
	}
	acceptme := SecureRandomAlphaString()
	srv.acceptables[acceptme] = time.Now()
	return acceptme
}

func (srv *Server) CheckAcceptable(val string) bool {
	if srv.acceptables == nil {
		srv.acceptables = make(map[string]time.Time)
	}
	if timeout, ok := srv.acceptables[val]; ok {
		checktime := time.Since(timeout)
		if checktime > (4 * time.Minute) {
			delete(srv.acceptables, val)
			return false
		}
		delete(srv.acceptables, val)
		return true
	}
	return false
}

func (srv *Server) reseedHandler(w http.ResponseWriter, r *http.Request) {
	var peer Peer
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		peer = Peer(ip)
	} else {
		peer = Peer(r.RemoteAddr)
	}

	su3Bytes, err := srv.Reseeder.PeerSu3Bytes(peer)
	if nil != err {
		http.Error(w, "500 Unable to serve su3", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=i2pseeds.su3")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(su3Bytes)), 10))

	io.Copy(w, bytes.NewReader(su3Bytes))
}

func disableKeepAliveMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Version", Version)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, next)
}

func (srv *Server) browsingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if srv.CheckAcceptable(r.FormValue("onetime")) {
			srv.reseedHandler(w, r)
		}
		if I2pUserAgent != r.UserAgent() {
			srv.HandleARealBrowser(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func verifyMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if I2pUserAgent != r.UserAgent() {
			http.Error(w, "403 Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func proxiedMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if prior, ok := r.Header["X-Forwarded-For"]; ok {
			r.RemoteAddr = prior[0]
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
