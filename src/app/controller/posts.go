package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Post struct {
	Uid      int    `json:"uid"`
	Text     string `json:"text"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Favorite bool   `json:"favorite"`
}

func Posts(w http.ResponseWriter, r *http.Request) {
	posts := []Post{}
	// you'd use a real database here
	file, err := ioutil.ReadFile("posts.json")
	if err != nil {
		log.Println("Error reading posts.json:", err)
		panic(err)
	}
	fmt.Printf("file: %s\n", string(file))
	err = json.Unmarshal(file, &posts)
	if err != nil {
		log.Println("Error unmarshalling posts.json:", err)
	}

	bs, err := json.Marshal(posts)
	if err != nil {
		ReturnError(w, err)
		return
	}
	fmt.Fprint(w, string(bs))
}
