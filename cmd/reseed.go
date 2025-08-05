package cmd

import (
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	//"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil"
	"github.com/cretz/bine/torutil/ed25519"
	"github.com/go-i2p/i2pkeys"
	"github.com/go-i2p/onramp"
	"github.com/go-i2p/sam3"
	"github.com/otiai10/copy"
	"github.com/rglonek/untar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"i2pgit.org/idk/reseed-tools/reseed"

	"github.com/go-i2p/checki2cp/getmeanetdb"
)

func getDefaultSigner() string {
	intentionalsigner := os.Getenv("RESEED_EMAIL")
	if intentionalsigner == "" {
		adminsigner := os.Getenv("MAILTO")
		if adminsigner != "" {
			return strings.Replace(adminsigner, "\n", "", -1)
		}
		return ""
	}
	return strings.Replace(intentionalsigner, "\n", "", -1)
}

func getHostName() string {
	hostname := os.Getenv("RESEED_HOSTNAME")
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	return strings.Replace(hostname, "\n", "", -1)
}

func providedReseeds() []string {
	reseedArg := viper.GetStringSlice("friends")
	reseed.AllReseeds = reseedArg
	return reseed.AllReseeds
}

// NewReseedCommand creates a new CLI command for starting a reseed server.
var reseedCmd = &cobra.Command{
	Use:   "reseed",
	Short: "Start a reseed server",
	Run: func(cmd *cobra.Command, args []string) {
		reseedAction()
	},
}

func init() {
	rootCmd.AddCommand(reseedCmd)
	ndb, err := getmeanetdb.WhereIstheNetDB()
	if err != nil {
		log.Fatal(err)
	}

	reseedCmd.PersistentFlags().String("signer", getDefaultSigner(), "Your su3 signing ID (ex. something@mail.i2p)")
	reseedCmd.PersistentFlags().String("tlsHost", getHostName(), "The public hostname used on your TLS certificate")
	reseedCmd.PersistentFlags().Bool("onion", false, "Present an onionv3 address")
	reseedCmd.PersistentFlags().Bool("singleOnion", false, "Use a faster, but non-anonymous single-hop onion")
	reseedCmd.PersistentFlags().String("onionKey", "onion.key", "Specify a path to an ed25519 private key for onion")
	reseedCmd.PersistentFlags().String("key", "", "Path to your su3 signing private key")
	reseedCmd.PersistentFlags().String("netdb", ndb, "Path to NetDB directory containing routerInfos")
	reseedCmd.PersistentFlags().Duration("routerInfoAge", 72*time.Hour, "Maximum age of router infos to include in reseed files (ex. 72h, 8d)")
	reseedCmd.PersistentFlags().String("tlsCert", "", "Path to a TLS certificate")
	reseedCmd.PersistentFlags().String("tlsKey", "", "Path to a TLS private key")
	reseedCmd.PersistentFlags().String("ip", "0.0.0.0", "IP address to listen on")
	reseedCmd.PersistentFlags().String("port", "8443", "Port to listen on")
	reseedCmd.PersistentFlags().Int("numRi", 77, "Number of routerInfos to include in each su3 file")
	reseedCmd.PersistentFlags().Int("numSu3", 50, "Number of su3 files to build (0 = automatic based on size of netdb)")
	reseedCmd.PersistentFlags().Duration("interval", 72*time.Hour, "Duration between SU3 cache rebuilds (ex. 12h, 15m)")
	reseedCmd.PersistentFlags().String("prefix", "", "Prefix path for the HTTP(S) server. (ex. /netdb)")
	reseedCmd.PersistentFlags().Bool("trustProxy", false,
		"If provided, we will trust the 'X-Forwarded-For' header in requests (ex. behind cloudflare)")
	reseedCmd.PersistentFlags().String("blacklist", "",
		"Path to a txt file containing a list of IPs to deny connections from.")
	reseedCmd.PersistentFlags().Duration("stats", 0, "Periodically print memory stats.")
	reseedCmd.PersistentFlags().Bool("i2p", false, "Listen for reseed request inside the I2P network")
	reseedCmd.PersistentFlags().Bool("yes", false, "Automatically answer 'yes' to self-signed SSL generation")
	reseedCmd.PersistentFlags().String("samaddr", "127.0.0.1:7656",
		"Use this SAM address to set up I2P connections for in-network reseed")
	reseedCmd.PersistentFlags().StringSlice("friends", reseed.AllReseeds,
		"Ping other reseed servers and display the result on the homepage to provide information about reseed uptime.")
	reseedCmd.PersistentFlags().String("share-peer", "",
		"Download the shared netDb content of another I2P router, over I2P")
	reseedCmd.PersistentFlags().String("share-password", "",
		"Password for downloading netDb content from another router. Required for share-peer to work.")
	reseedCmd.PersistentFlags().Bool("acme", false,
		"Automatically generate a TLS certificate with the ACME protocol, defaults to Let's Encrypt")
	reseedCmd.PersistentFlags().String(
		"acmeserver",
		"https://acme-staging-v02.api.letsencrypt.org/directory",
		"Use this server to issue a certificate with the ACME protocol")
	reseedCmd.PersistentFlags().Int("ratelimit", 4, "Maximum number of reseed bundle requests per-IP address, per-hour.")
	reseedCmd.PersistentFlags().Int("ratelimitweb", 40, "Maxiumum number of web-visits per-IP address, per-hour")

	viper.BindPFlags(reseedCmd.Flags())
}

func CreateEepServiceKey() (i2pkeys.I2PKeys, error) {
	sam, err := sam3.NewSAM(viper.GetString("samaddr"))
	if err != nil {
		return i2pkeys.I2PKeys{}, err
	}
	defer sam.Close()
	k, err := sam.NewKeys()
	if err != nil {
		return i2pkeys.I2PKeys{}, err
	}
	return k, err
}

func LoadKeys(keysPath string) (i2pkeys.I2PKeys, error) {
	if _, err := os.Stat(keysPath); os.IsNotExist(err) {
		keys, err := CreateEepServiceKey()
		if err != nil {
			return i2pkeys.I2PKeys{}, err
		}
		file, err := os.Create(keysPath)
		if err != nil {
			return i2pkeys.I2PKeys{}, err
		}
		defer file.Close()
		err = i2pkeys.StoreKeysIncompat(keys, file)
		if err != nil {
			return i2pkeys.I2PKeys{}, err
		}
		return keys, nil
	} else if err == nil {
		file, err := os.Open(keysPath)
		if err != nil {
			return i2pkeys.I2PKeys{}, err
		}
		defer file.Close()
		keys, err := i2pkeys.LoadKeysIncompat(file)
		if err != nil {
			return i2pkeys.I2PKeys{}, err
		}
		return keys, nil
	} else {
		return i2pkeys.I2PKeys{}, err
	}
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func reseedAction() error {
	providedReseeds()
	netdbDir := viper.GetString("netdb")
	if netdbDir == "" {
		fmt.Println("--netdb is required")
		return fmt.Errorf("--netdb is required")
	}

	signerID := viper.GetString("signer")
	if signerID == "" || signerID == "you@mail.i2p" {
		fmt.Println("--signer is required")
		return fmt.Errorf("--signer is required")
	}
	if !strings.Contains(signerID, "@") {
		if !fileExists(signerID) {
			fmt.Println("--signer must be an email address or a file containing an email address.")
			return fmt.Errorf("--signer must be an email address or a file containing an email address.")
		}
		bytes, err := os.ReadFile(signerID)
		if err != nil {
			fmt.Println("--signer must be an email address or a file containing an email address.")
			return fmt.Errorf("--signer must be an email address or a file containing an email address.")
		}
		signerID = string(bytes)
	}
	if viper.GetString("share-peer") != "" {
		count := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		for i := range count {
			err := downloadRemoteNetDB(viper.GetString("share-peer"), viper.GetString("share-password"), viper.GetString("netdb"), viper.GetString("samaddr"))
			if err != nil {
				log.Println("Error downloading remote netDb,", err, "retrying in 10 seconds", i, "attempts remaining")
				time.Sleep(time.Second * 10)
			} else {
				break
			}
		}
		go getSupplementalNetDb(viper.GetString("share-peer"), viper.GetString("share-password"), viper.GetString("netdb"), viper.GetString("samaddr"))
	}

	var tlsCert, tlsKey string
	tlsHost := viper.GetString("tlsHost")
	onionTlsHost := ""
	var onionTlsCert, onionTlsKey string
	i2pTlsHost := ""
	var i2pTlsCert, i2pTlsKey string
	var i2pkey i2pkeys.I2PKeys

	if tlsHost != "" {
		onionTlsHost = tlsHost
		i2pTlsHost = tlsHost
		tlsKey = viper.GetString("tlsKey")
		// if no key is specified, default to the host.pem in the current dir
		if tlsKey == "" {
			tlsKey = tlsHost + ".pem"
			onionTlsKey = tlsHost + ".pem"
			i2pTlsKey = tlsHost + ".pem"
		}

		tlsCert = viper.GetString("tlsCert")
		// if no certificate is specified, default to the host.crt in the current dir
		if tlsCert == "" {
			tlsCert = tlsHost + ".crt"
			onionTlsCert = tlsHost + ".crt"
			i2pTlsCert = tlsHost + ".crt"
		}

		// prompt to create tls keys if they don't exist?
		auto := viper.GetBool("yes")
		ignore := viper.GetBool("trustProxy")
		if !ignore {
			// use ACME?
			acme := viper.GetBool("acme")
			if acme {
				acmeserver := viper.GetString("acmeserver")
				err := checkUseAcmeCert(tlsHost, signerID, acmeserver, &tlsCert, &tlsKey, auto)
				if nil != err {
					log.Fatalln(err)
				}
			} else {
				err := checkOrNewTLSCert(tlsHost, &tlsCert, &tlsKey, auto)
				if nil != err {
					log.Fatalln(err)
				}
			}
		}

	}

	if viper.GetBool("i2p") {
		var err error
		i2pkey, err = LoadKeys("reseed.i2pkeys")
		if err != nil {
			log.Fatalln(err)
		}
		if i2pTlsHost == "" {
			i2pTlsHost = i2pkey.Addr().Base32()
		}
		if i2pTlsHost != "" {
			// if no key is specified, default to the host.pem in the current dir
			if i2pTlsKey == "" {
				i2pTlsKey = i2pTlsHost + ".pem"
			}

			// if no certificate is specified, default to the host.crt in the current dir
			if i2pTlsCert == "" {
				i2pTlsCert = i2pTlsHost + ".crt"
			}

			// prompt to create tls keys if they don't exist?
			auto := viper.GetBool("yes")
			ignore := viper.GetBool("trustProxy")
			if !ignore {
				err := checkOrNewTLSCert(i2pTlsHost, &i2pTlsCert, &i2pTlsKey, auto)
				if nil != err {
					log.Fatalln(err)
				}
			}
		}
	}

	if viper.GetBool("onion") {
		var ok []byte
		var err error
		if _, err = os.Stat(viper.GetString("onionKey")); err == nil {
			ok, err = os.ReadFile(viper.GetString("onionKey"))
			if err != nil {
				log.Fatalln(err.Error())
			}
		} else {
			key, err := ed25519.GenerateKey(nil)
			if err != nil {
				log.Fatalln(err.Error())
			}
			ok = []byte(key.PrivateKey())
		}
		if onionTlsHost == "" {
			onionTlsHost = torutil.OnionServiceIDFromPrivateKey(ed25519.PrivateKey(ok)) + ".onion"
		}
		err = os.WriteFile(viper.GetString("onionKey"), ok, 0o644)
		if err != nil {
			log.Fatalln(err.Error())
		}
		if onionTlsHost != "" {
			// if no key is specified, default to the host.pem in the current dir
			if onionTlsKey == "" {
				onionTlsKey = onionTlsHost + ".pem"
			}

			// if no certificate is specified, default to the host.crt in the current dir
			if onionTlsCert == "" {
				onionTlsCert = onionTlsHost + ".crt"
			}

			// prompt to create tls keys if they don't exist?
			auto := viper.GetBool("yes")
			ignore := viper.GetBool("trustProxy")
			if !ignore {
				err := checkOrNewTLSCert(onionTlsHost, &onionTlsCert, &onionTlsKey, auto)
				if nil != err {
					log.Fatalln(err)
				}
			}
		}
	}

	reloadIntvl, err := time.ParseDuration(viper.GetString("interval"))
	if nil != err {
		fmt.Printf("'%s' is not a valid time interval.\n", reloadIntvl)
		return fmt.Errorf("'%s' is not a valid time interval.\n", reloadIntvl)
	}

	signerKey := viper.GetString("key")
	// if no key is specified, default to the signerID.pem in the current dir
	if signerKey == "" {
		signerKey = signerFile(signerID) + ".pem"
	}

	// load our signing privKey
	auto := viper.GetBool("yes")
	privKey, err := getOrNewSigningCert(&signerKey, signerID, auto)
	if nil != err {
		log.Fatalln(err)
	}

	// create a local file netdb provider
	routerInfoAge := viper.GetDuration("routerInfoAge")
	netdb := reseed.NewLocalNetDb(netdbDir, routerInfoAge)

	// create a reseeder
	reseeder := reseed.NewReseeder(netdb)
	reseeder.SigningKey = privKey
	reseeder.SignerID = []byte(signerID)
	reseeder.NumRi = viper.GetInt("numRi")
	reseeder.NumSu3 = viper.GetInt("numSu3")
	reseeder.RebuildInterval = reloadIntvl
	reseeder.Start()

	// create a server

	if viper.GetBool("onion") {
		log.Printf("Onion server starting\n")
		if tlsHost != "" && tlsCert != "" && tlsKey != "" {
			go reseedOnion(onionTlsCert, onionTlsKey, reseeder)
		} else {
			reseedOnion(onionTlsCert, onionTlsKey, reseeder)
		}
	}
	if viper.GetBool("i2p") {
		log.Printf("I2P server starting\n")
		if tlsHost != "" && tlsCert != "" && tlsKey != "" {
			go reseedI2P(i2pTlsCert, i2pTlsKey, i2pkey, reseeder)
		} else {
			reseedI2P(i2pTlsCert, i2pTlsKey, i2pkey, reseeder)
		}
	}
	if !viper.GetBool("trustProxy") {
		log.Printf("HTTPS server starting\n")
		reseedHTTPS(tlsCert, tlsKey, reseeder)
	} else {
		log.Printf("HTTP server starting on\n")
		reseedHTTP(reseeder)
	}
	return nil
}

func reseedHTTPS(tlsCert, tlsKey string, reseeder *reseed.ReseederImpl) {
	server := reseed.NewServer(viper.GetString("prefix"), viper.GetBool("trustProxy"))
	server.Reseeder = reseeder
	server.RequestRateLimit = viper.GetInt("ratelimit")
	server.WebRateLimit = viper.GetInt("ratelimitweb")
	server.Addr = net.JoinHostPort(viper.GetString("ip"), viper.GetString("port"))

	// load a blacklist
	blacklist := reseed.NewBlacklist()
	server.Blacklist = blacklist
	blacklistFile := viper.GetString("blacklist")
	if "" != blacklistFile {
		blacklist.LoadFile(blacklistFile)
	}

	// print stats once in a while
	if viper.GetDuration("stats") != 0 {
		go func() {
			var mem runtime.MemStats
			for range time.Tick(viper.GetDuration("stats")) {
				runtime.ReadMemStats(&mem)
				log.Printf("TotalAllocs: %d Kb, Allocs: %d Kb, Mallocs: %d, NumGC: %d", mem.TotalAlloc/1024, mem.Alloc/1024, mem.Mallocs, mem.NumGC)
			}
		}()
	}
	log.Printf("HTTPS server started on %s\n", server.Addr)
	if err := server.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
		log.Fatalln(err)
	}
}

func reseedHTTP(reseeder *reseed.ReseederImpl) {
	server := reseed.NewServer(viper.GetString("prefix"), viper.GetBool("trustProxy"))
	server.RequestRateLimit = viper.GetInt("ratelimit")
	server.WebRateLimit = viper.GetInt("ratelimitweb")
	server.Reseeder = reseeder
	server.Addr = net.JoinHostPort(viper.GetString("ip"), viper.GetString("port"))

	// load a blacklist
	blacklist := reseed.NewBlacklist()
	server.Blacklist = blacklist
	blacklistFile := viper.GetString("blacklist")
	if "" != blacklistFile {
		blacklist.LoadFile(blacklistFile)
	}

	// print stats once in a while
	if viper.GetDuration("stats") != 0 {
		go func() {
			var mem runtime.MemStats
			for range time.Tick(viper.GetDuration("stats")) {
				runtime.ReadMemStats(&mem)
				log.Printf("TotalAllocs: %d Kb, Allocs: %d Kb, Mallocs: %d, NumGC: %d", mem.TotalAlloc/1024, mem.Alloc/1024, mem.Mallocs, mem.NumGC)
			}
		}()
	}
	log.Printf("HTTP server started on %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}

func reseedOnion(onionTlsCert, onionTlsKey string, reseeder *reseed.ReseederImpl) {
	server := reseed.NewServer(viper.GetString("prefix"), viper.GetBool("trustProxy"))
	server.Reseeder = reseeder
	server.Addr = net.JoinHostPort(viper.GetString("ip"), viper.GetString("port"))

	// load a blacklist
	blacklist := reseed.NewBlacklist()
	server.Blacklist = blacklist
	blacklistFile := viper.GetString("blacklist")
	if "" != blacklistFile {
		blacklist.LoadFile(blacklistFile)
	}

	// print stats once in a while
	if viper.GetDuration("stats") != 0 {
		go func() {
			var mem runtime.MemStats
			for range time.Tick(viper.GetDuration("stats")) {
				runtime.ReadMemStats(&mem)
				log.Printf("TotalAllocs: %d Kb, Allocs: %d Kb, Mallocs: %d, NumGC: %d", mem.TotalAlloc/1024, mem.Alloc/1024, mem.Mallocs, mem.NumGC)
			}
		}()
	}
	port, err := strconv.Atoi(viper.GetString("port"))
	if err != nil {
		log.Fatalln(err.Error())
	}
	port += 1
	if _, err := os.Stat(viper.GetString("onionKey")); err == nil {
		ok, err := os.ReadFile(viper.GetString("onionKey"))
		if err != nil {
			log.Fatalln(err.Error())
		} else {
			if onionTlsCert != "" && onionTlsKey != "" {
				tlc := &tor.ListenConf{
					LocalPort:    port,
					Key:          ed25519.PrivateKey(ok),
					RemotePorts:  []int{443},
					Version3:     true,
					NonAnonymous: viper.GetBool("singleOnion"),
					DiscardKey:   false,
				}
				if err := server.ListenAndServeOnionTLS(nil, tlc, onionTlsCert, onionTlsKey); err != nil {
					log.Fatalln(err)
				}
			} else {
				tlc := &tor.ListenConf{
					LocalPort:    port,
					Key:          ed25519.PrivateKey(ok),
					RemotePorts:  []int{80},
					Version3:     true,
					NonAnonymous: viper.GetBool("singleOnion"),
					DiscardKey:   false,
				}
				if err := server.ListenAndServeOnion(nil, tlc); err != nil {
					log.Fatalln(err)
				}

			}
		}
	} else if os.IsNotExist(err) {
		tlc := &tor.ListenConf{
			LocalPort:    port,
			RemotePorts:  []int{80},
			Version3:     true,
			NonAnonymous: viper.GetBool("singleOnion"),
			DiscardKey:   false,
		}
		if err := server.ListenAndServeOnion(nil, tlc); err != nil {
			log.Fatalln(err)
		}
	}
	log.Printf("Onion server started on %s\n", server.Addr)
}

func reseedI2P(i2pTlsCert, i2pTlsKey string, i2pIdentKey i2pkeys.I2PKeys, reseeder *reseed.ReseederImpl) {
	server := reseed.NewServer(viper.GetString("prefix"), viper.GetBool("trustProxy"))
	server.RequestRateLimit = viper.GetInt("ratelimit")
	server.WebRateLimit = viper.GetInt("ratelimitweb")
	server.Reseeder = reseeder
	server.Addr = net.JoinHostPort(viper.GetString("ip"), viper.GetString("port"))

	// load a blacklist
	blacklist := reseed.NewBlacklist()
	server.Blacklist = blacklist
	blacklistFile := viper.GetString("blacklist")
	if "" != blacklistFile {
		blacklist.LoadFile(blacklistFile)
	}

	// print stats once in a while
	if viper.GetDuration("stats") != 0 {
		go func() {
			var mem runtime.MemStats
			for range time.Tick(viper.GetDuration("stats")) {
				runtime.ReadMemStats(&mem)
				log.Printf("TotalAllocs: %d Kb, Allocs: %d Kb, Mallocs: %d, NumGC: %d", mem.TotalAlloc/1024, mem.Alloc/1024, mem.Mallocs, mem.NumGC)
			}
		}()
	}
	port, err := strconv.Atoi(viper.GetString("port"))
	if err != nil {
		log.Fatalln(err.Error())
	}
	port += 1
	if i2pTlsCert != "" && i2pTlsKey != "" {
		if err := server.ListenAndServeI2PTLS(viper.GetString("samaddr"), i2pIdentKey, i2pTlsCert, i2pTlsKey); err != nil {
			log.Fatalln(err)
		}
	} else {
		if err := server.ListenAndServeI2P(viper.GetString("samaddr"), i2pIdentKey); err != nil {
			log.Fatalln(err)
		}
	}

	log.Printf("Onion server started on %s\n", server.Addr)
}

func getSupplementalNetDb(remote, password, path, samaddr string) {
	log.Println("Remote NetDB Update Loop")
	for {
		if err := downloadRemoteNetDB(remote, password, path, samaddr); err != nil {
			log.Println("Error downloading remote netDb", err)
			time.Sleep(time.Second * 30)
		} else {
			log.Println("Success downloading remote netDb", err)
			time.Sleep(time.Minute * 30)
		}
	}
}

func downloadRemoteNetDB(remote, password, path, samaddr string) error {
	var hremote string
	if !strings.HasPrefix("http://", remote) && !strings.HasPrefix("https://", remote) {
		hremote = "http://" + remote
	}
	if !strings.HasSuffix(hremote, ".tar.gz") {
		hremote += "/netDb.tar.gz"
	}
	url, err := url.Parse(hremote)
	if err != nil {
		return err
	}
	httpRequest := http.Request{
		URL:    url,
		Header: http.Header{},
	}
	garlic, err := onramp.NewGarlic("reseed-client", samaddr, onramp.OPT_WIDE)
	if err != nil {
		return err
	}

	defer garlic.Close()
	httpRequest.Header.Add(http.CanonicalHeaderKey("reseed-password"), password)
	httpRequest.Header.Add(http.CanonicalHeaderKey("x-user-agent"), reseed.I2pUserAgent)
	transport := http.Transport{
		Dial: garlic.Dial,
	}
	client := http.Client{
		Transport: &transport,
	}
	if resp, err := client.Do(&httpRequest); err != nil {
		return err
	} else {
		if bodyBytes, err := io.ReadAll(resp.Body); err != nil {
			return err
		} else {
			if err := os.WriteFile("netDb.tar.gz", bodyBytes, 0o644); err != nil {
				return err
			} else {
				dbPath := filepath.Join(path, "reseed-netDb")
				if err := untar.UntarFile("netDb.tar.gz", dbPath); err != nil {
					return err
				} else {
					// For example...
					opt := copy.Options{
						Skip: func(info os.FileInfo, src, dest string) (bool, error) {
							srcBase := filepath.Base(src)
							dstBase := filepath.Base(dest)
							if info.IsDir() {
								return false, nil
							}
							if srcBase == dstBase {
								log.Println("Ignoring existing RI", srcBase, dstBase)
								return true, nil
							}
							return false, nil
						},
					}
					if err := copy.Copy(dbPath, path, opt); err != nil {
						return err
					} else {
						if err := os.RemoveAll(dbPath); err != nil {
							return err
						} else {
							if err := os.RemoveAll("netDb.tar.gz"); err != nil {
								return err
							}
							return nil
						}
					}
				}
			}
		}
	}
}
