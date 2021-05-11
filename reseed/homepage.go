package reseed

import (
	"io/ioutil"
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
    <title>This is an I2P Reseed Server</title>
    <link rel="stylesheet" href="style.css">
    <script src="script.js"></script>
  </head>
  <body>`)
var footer = []byte(`  </body>
</html>`)

var md = markdown.New(markdown.XHTMLOutput(true), markdown.HTML(true))

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

func (srv *Server) HandleARealBrowser(w http.ResponseWriter, r *http.Request) {
	if ContentPathError != nil {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	lang, _ := r.Cookie("lang")
	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(matcher, lang.String(), accept)
	base, _ := tag.Base()

	switch r.URL.Path {
	case "/style.css":
		w.Header().Set("Content-Type", "text/css")
		HandleAFile(w, "", "style.css")
	case "/script.js":
		w.Header().Set("Content-Type", "text/javascript")
		HandleAFile(w, "", "script.js")
	default:
		image := strings.Replace(r.URL.Path, "/", "", -1)
		if strings.HasPrefix(image, "images") {
			w.Header().Set("Content-Type", "image/png")
			HandleAFile(w, "images", strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/"), "images"))
		} else {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(header))
			HandleALocalizedFile(w, base.String())
			w.Write([]byte(`<form method="post" action="/i2pseeds" class="inline">
			<input type="hidden" name="onetime" value="` + srv.Acceptable() + `">
			<button type="submit" name="submit_param" value="submit_value" class="link-button">
			Bundle
			</button>
			</form>`))
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
			w.Write([]byte("Oops! Something went wrong handling your language. Please file a bug at https://i2pgit.org/idk/reseed-tools\n\t" + err.Error()))
			return
		}
		CachedDataPages[file] = f
		w.Write([]byte(CachedDataPages[file]))
	} else {
		w.Write(CachedDataPages[file])
	}
}

func HandleALocalizedFile(w http.ResponseWriter, dirPath string) {
	if _, prs := CachedLanguagePages[dirPath]; prs == false {
		dir := filepath.Join(BaseContentPath, "lang", dirPath)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			w.Write([]byte("Oops! Something went wrong handling your language. Please file a bug at https://i2pgit.org/idk/reseed-tools\n\t" + err.Error()))
		}
		var f []byte
		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".md") {
				return
			}
			trimmedName := strings.TrimSuffix(file.Name(), ".md")
			path := filepath.Join(dir, file.Name())
			b, err := ioutil.ReadFile(path)
			if err != nil {
				w.Write([]byte("Oops! Something went wrong handling your language. Please file a bug at https://i2pgit.org/idk/reseed-tools\n\t" + err.Error()))
				return
			}
			f = append(f, []byte(`<div id="`+trimmedName+`">`)...)
			f = append(f, []byte(md.RenderToString(b))...)
			f = append(f, []byte(`</div>`)...)

		}
		CachedLanguagePages[dirPath] = string(f)
		w.Write([]byte(CachedLanguagePages[dirPath]))
	} else {
		w.Write([]byte(CachedLanguagePages[dirPath]))
	}
}
