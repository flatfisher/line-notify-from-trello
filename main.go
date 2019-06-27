package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/", indexHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

//Card is a struct...
type Card struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func notify(msg string) error {
	URL := "https://notify-api.line.me/api/notify"
	u, err := url.ParseRequestURI(URL)
	if err != nil {
		return err
	}

	c := &http.Client{}
	form := url.Values{}
	form.Add("message", msg)
	body := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return err
	}

	token := os.Getenv("LINE_TOKEN")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	if _, err := c.Do(req); err != nil {
		return err
	}

	return nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	id := os.Getenv("TRELLO_LIST_ID")
	token := os.Getenv("TRELLO_TOKEN")
	URL := fmt.Sprintf("https://trello.com/1/lists/%s&token=%s&fields=name", id, token)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var c []Card
	err = json.Unmarshal(b, &c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var msg string
	for i := 0; i < len(c); i++ {
		msg += fmt.Sprintf("%s\n", c[i].Name)
	}

	// Notify to LINE
	if err := notify(msg); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprint(w, "Success")
}
