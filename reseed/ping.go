package reseed

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Ping requests an `.su3` from another reseed server and return true if
// the reseed server is alive If the reseed server is not alive, returns
// false and the status of the request as an error
func Ping(url string) (bool, error) {
	if strings.HasSuffix(url, "i2pseeds.su3") {
		url = url + "i2pseeds.su3"
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", i2pUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("%s", resp.Status)
	}
	return true, nil
}

func PingWriteContent(url string) error {
	date := time.Now().Format("2006-01-02")
	path := strings.Replace(url, "http://", "", 1)
	path = strings.Replace(path, "https://", "", 1)
	path = strings.Replace(path, "/", "", -1)
	path = filepath.Join(BaseContentPath, path+"-"+date+".ping")
	if _, err := os.Stat(path); err != nil {
		result, err := Ping(url)
		if result {
			log.Printf("Ping: %s OK", url)
			err := ioutil.WriteFile(path, []byte("Alive: Status OK"), 0644)
			return err
		} else {
			log.Printf("Ping: %s %s", url, err)
			err := ioutil.WriteFile(path, []byte("Dead: "+err.Error()), 0644)
			return err
		}
	}
	return nil
}

//TODO: make this a configuration option
var AllReseeds = []string{
	"https://banana.incognet.io/",
	"https://i2p.novg.net/",
	"https://i2pseed.creativecowpat.net:8443/",
	"https://reseed.diva.exchange/",
	"https://reseed.i2pgit.org/",
	"https://reseed.memcpy.io/",
	"https://reseed.onion.im/",
	"https://reseed2.i2p.net/",
}

func PingEverybody() []string {
	var nonerrs []string
	for _, url := range AllReseeds {
		err := PingWriteContent(url)
		if err == nil {
			nonerrs = append(nonerrs, url)
		} else {
			nonerrs = append(nonerrs, err.Error()+"-"+url)
		}
	}
	return nonerrs
}

// Get a list of all files ending in ping in the BaseContentPath
func GetPingFiles() ([]string, error) {
	var files []string
	err := filepath.Walk(BaseContentPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".ping") {
			files = append(files, path)
		}
		return nil
	})
	if len(files) == 0 {
		return nil, fmt.Errorf("No ping files found")
	}
	return files, err
}

func ReadOut(w http.ResponseWriter) {
	pinglist, err := GetPingFiles()
	if err == nil {
		fmt.Fprintf(w, "<h3>Reseed Server Statuses</h3>")
		fmt.Fprintf(w, "<p>This feature is experimental and may not always provide accurate results.</p>")
		fmt.Fprintf(w, "</div><p><ul>")
		for _, file := range pinglist {
			ping, err := ioutil.ReadFile(file)
			host := strings.Replace(file, ".ping", "", 1)
			host = filepath.Base(host)
			if err == nil {
				fmt.Fprintf(w, "<li><strong>%s</strong> - %s</li>\n", host, ping)
			} else {
				fmt.Fprintf(w, "<li><strong>%s</strong> - No ping file found</li>\n", host)
			}
		}
		fmt.Fprintf(w, "</ul></p></div>")
	} else {
		fmt.Fprintf(w, "<h4>No ping files found, check back later for reseed stats</h4>")
	}
}
