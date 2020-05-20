package reseed

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/golang-commonmark/markdown"
	"golang.org/x/text/language"
)

var SupportedLanguages = []language.Tag{
	language.English,
}
var CachedLanguagePages = map[string]string{}
var CachedDataPages = map[string][]byte{}

var BaseContentPath, ContentPathError = ContentPath()

var matcher = language.NewMatcher(SupportedLanguages)

var header = []byte(`<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>title</title>
    <link rel="stylesheet" href="style.css">
    <script src="script.js"></script>
  </head>
  <body>`)
var footer = []byte(`  </body>
</html>`)

var md = markdown.New(markdown.XHTMLOutput(true))

func ContentPath() (string, error) {
	exPath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	//exPath := filepath.Dir(ex)
	if _, err := os.Stat(filepath.Join(exPath, "content")); err != nil {
		return "", err
	}
	return filepath.Join(exPath, "content"), nil
}

func HandleARealBrowser(w http.ResponseWriter, r *http.Request) {
	if ContentPathError != nil {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	lang, err := r.Cookie("lang")
	if err != nil {
		log.Printf("Language cookie error: %s\n")
	}
	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, lang.String(), accept)
	base, _ := tag.Base()

	switch r.URL.Path {
	case "/style.css":
		//		log.Printf("Showing CSS %s %s", BaseContentPath, r.URL.Path)
		w.Header().Set("Content-Type", "text/css")
		HandleAFile(w, "", "style.css")
	case "/script.js":
		//		log.Printf("Showing JAVASCRIPT %s %s", BaseContentPath, r.URL.Path)
		w.Header().Set("Content-Type", "text/javascript")
		HandleAFile(w, "", "script.js")
	default:
		image := strings.Replace(r.URL.Path, "/", "", -1)
		if strings.HasPrefix(image, "images") {
			//			log.Printf("Showing IMAGE %s %s", BaseContentPath, r.URL.Path)
			w.Header().Set("Content-Type", "image/png")
			HandleAFile(w, "images", strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/"), "images"))
		} else {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(header))
			HandleALocalizedFile(w, base.String(), "homepage.md")
			w.Write([]byte(footer))
		}
	}
}

func HandleAFile(w http.ResponseWriter, dirPath, file string) {
	file = filepath.Join(dirPath, file)
	if _, prs := CachedDataPages[file]; prs == false {
		path := filepath.Join(BaseContentPath, file)
		f, err := ioutil.ReadFile(path)
		if err != nil {
			w.Write([]byte("Oops! Something went wrong handling your language. Please file a bug at https://github.com/eyedeekay/i2p-tools-1\n\t" + err.Error()))
		}
		CachedDataPages[file] = f
		w.Write([]byte(CachedDataPages[file]))
	} else {
		w.Write(CachedDataPages[file])
	}
}

func HandleALocalizedFile(w http.ResponseWriter, dirPath, file string) {
	file = filepath.Join(dirPath, file)
	if _, prs := CachedLanguagePages[file]; prs == false {
		path := filepath.Join(BaseContentPath, "lang", file)
		log.Printf("Showing HYPERTEXT %s", path)
		f, err := ioutil.ReadFile(path)
		if err != nil {
			w.Write([]byte("Oops! Something went wrong handling your language. Please file a bug at https://github.com/eyedeekay/i2p-tools-1\n\t" + err.Error()))
		}
		log.Printf("%b\n", f)
		CachedLanguagePages[file] = md.RenderToString(f)
		w.Write([]byte(CachedLanguagePages[file]))
	} else {
		w.Write([]byte(CachedLanguagePages[file]))
	}
}
