package main

import (
	"fmt"
	"net/http"
)

const addForm = `
<html><body>
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
</html></body>
`

var store = NewURLStore("store.gob")

func main() {
	defer store.close()
	http.HandleFunc("/", Redirect)
	http.HandleFunc("/add", Add)
	http.ListenAndServe(":3000", nil)
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	url := store.Get(key)
	if url == "" {
		http.Redirect(w, r, "/add", http.StatusFound)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func Add(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	url := r.FormValue("url")
	if url == "" {
		fmt.Fprint(w, addForm)
		return
	}

	key := store.Put(url)

	fmt.Fprintf(w, "%s", key)
}
