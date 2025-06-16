package reseed

import (
	"crypto/tls"
	"log"
	"net"

	"github.com/cretz/bine/tor"
	"github.com/go-i2p/i2pkeys"
	"github.com/go-i2p/onramp"
)

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return srv.Serve(newBlacklistListener(ln, srv.Blacklist))
}

func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	if srv.TLSConfig == nil {
		srv.TLSConfig = &tls.Config{}
	}

	if srv.TLSConfig.NextProtos == nil {
		srv.TLSConfig.NextProtos = []string{"http/1.1"}
	}

	var err error
	srv.TLSConfig.Certificates = make([]tls.Certificate, 1)
	srv.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(newBlacklistListener(ln, srv.Blacklist), srv.TLSConfig)
	return srv.Serve(tlsListener)
}

func (srv *Server) ListenAndServeOnionTLS(startConf *tor.StartConf, listenConf *tor.ListenConf, certFile, keyFile string) error {
	log.Println("Starting and registering OnionV3 HTTPS service, please wait a couple of minutes...")
	var err error
	srv.Onion, err = onramp.NewOnion("reseed")
	if err != nil {
		return err
	}
	srv.OnionListener, err = srv.Onion.ListenTLS()
	if err != nil {
		return err
	}
	log.Printf("Onionv3 server started on https://%v.onion\n", srv.OnionListener.Addr().String())

	return srv.Serve(srv.OnionListener)
}

func (srv *Server) ListenAndServeOnion(startConf *tor.StartConf, listenConf *tor.ListenConf) error {
	log.Println("Starting and registering OnionV3 HTTP service, please wait a couple of minutes...")
	var err error
	srv.Onion, err = onramp.NewOnion("reseed")
	if err != nil {
		return err
	}
	srv.OnionListener, err = srv.Onion.Listen()
	if err != nil {
		return err
	}
	log.Printf("Onionv3 server started on http://%v.onion\n", srv.OnionListener.Addr().String())

	return srv.Serve(srv.OnionListener)
}

func (srv *Server) ListenAndServeI2PTLS(samaddr string, I2PKeys i2pkeys.I2PKeys, certFile, keyFile string) error {
	log.Println("Starting and registering I2P HTTPS service, please wait a couple of minutes...")
	var err error
	srv.Garlic, err = onramp.NewGarlic("reseed-tls", samaddr, onramp.OPT_WIDE)
	if err != nil {
		return err
	}
	srv.I2PListener, err = srv.Garlic.ListenTLS()
	if err != nil {
		return err
	}
	log.Printf("I2P server started on https://%v\n", srv.I2PListener.Addr().(i2pkeys.I2PAddr).Base32())
	return srv.Serve(srv.I2PListener)
}

func (srv *Server) ListenAndServeI2P(samaddr string, I2PKeys i2pkeys.I2PKeys) error {
	log.Println("Starting and registering I2P service, please wait a couple of minutes...")
	var err error
	srv.Garlic, err = onramp.NewGarlic("reseed", samaddr, onramp.OPT_WIDE)
	if err != nil {
		return err
	}
	srv.I2PListener, err = srv.Garlic.Listen()
	if err != nil {
		return err
	}
	log.Printf("I2P server started on http://%v.b32.i2p\n", srv.I2PListener.Addr().(i2pkeys.I2PAddr).Base32())
	return srv.Serve(srv.I2PListener)
}
