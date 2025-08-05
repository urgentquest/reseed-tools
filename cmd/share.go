package cmd

import (

	//"flag"

	"archive/tar"
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-i2p/checki2cp/getmeanetdb"
	"github.com/go-i2p/onramp"
)

// NewShareCommand creates a new CLI Command for sharing the netDb over I2P with a password.
// Can be used to combine the local netDb with the netDb of a remote I2P router.
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Start a netDb sharing server",
	Run: func(cmd *cobra.Command, args []string) {
		shareAction()
	},
}

func init() {
	ndb, err := getmeanetdb.WhereIstheNetDB()
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.AddCommand(shareCmd)

	shareCmd.Flags().String("signer", getDefaultSigner(), "Your su3 signing ID (ex. something@mail.i2p)")
	shareCmd.Flags().String("key", "", "Path to your su3 signing private key")
	shareCmd.Flags().String("netdb", ndb, "Path to NetDB directory containing routerInfos")
	shareCmd.Flags().String("samaddr", "127.0.0.1:7656",
		"Use this SAM address to set up I2P connections for in-network sharing")
	shareCmd.Flags().String("share-password", "",
		"Share the contents of your netDb directory privately over I2P as a tar.gz archive. Will fail is password is blank.",
	)

	viper.BindPFlags(shareCmd.Flags())
}

type sharer struct {
	http.FileSystem
	http.Handler
	Path     string
	Password string
}

func (s *sharer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, ok := r.Header[http.CanonicalHeaderKey("reseed-password")]
	if !ok {
		return
	}
	if p[0] != s.Password {
		return
	}
	log.Println("Path", r.URL.Path)
	if strings.HasSuffix(r.URL.Path, "tar.gz") {
		log.Println("Serving netdb")
		archive, err := walker(s.Path)
		if err != nil {
			return
		}
		w.Write(archive.Bytes())
		return
	}
	s.Handler.ServeHTTP(w, r)
}

func Sharer(netDbDir, password string) *sharer {
	fileSystem := &sharer{
		FileSystem: http.Dir(netDbDir),
		Path:       netDbDir,
		Password:   password,
	}
	fileSystem.Handler = http.FileServer(fileSystem.FileSystem)
	return fileSystem
}

func shareAction() error {
	netDbDir, err := filepath.Abs(viper.GetString("netdb"))
	if err != nil {
		return err
	}
	httpFs := Sharer(netDbDir, viper.GetString("share-password"))
	garlic, err := onramp.NewGarlic("reseed", viper.GetString("samaddr"), onramp.OPT_WIDE)
	if err != nil {
		return err
	}
	defer garlic.Close()

	garlicListener, err := garlic.Listen()
	if err != nil {
		return err
	}
	defer garlicListener.Close()

	return http.Serve(garlicListener, httpFs)
}

func walker(netDbDir string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsDir() {
			return nil
		}
		new_path := path[len(netDbDir):]
		if len(new_path) == 0 {
			return nil
		}
		fr, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fr.Close()
		if h, err := tar.FileInfoHeader(info, new_path); err != nil {
			log.Fatalln(err)
		} else {
			h.Name = new_path
			if err = tw.WriteHeader(h); err != nil {
				log.Fatalln(err)
			}
		}
		if _, err := io.Copy(tw, fr); err != nil {
			log.Fatalln(err)
		}
		return nil
	}
	if err := filepath.Walk(netDbDir, walkFn); err != nil {
		return nil, err
	}
	return &buf, nil
}
