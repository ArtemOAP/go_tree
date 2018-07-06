package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Token    string  `json:"ref_code"`
	Code     string  `json:"-"`
	Children []*User `json:"children"`
}

var items []*User

func main() {

	db, err := sql.Open("mysql", "root:@/geo")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT `id`, COALESCE(`name`,''),COALESCE(`email`,''), COALESCE(`token`,''),COALESCE(`code`,'') FROM `users`")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	items := make([]*User, 0)
	for rows.Next() {
		item := new(User)
		err := rows.Scan(&item.Id, &item.Name, &item.Email, &item.Token, &item.Code)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	items = Tree(items)
	js_users, _ := json.Marshal(items)
	fs := http.FileServer(http.Dir("dir"))
	http.Handle("/img/", http.StripPrefix("/img/", fs))

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		if r.URL.Path != "/test" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Server", "A Go Web Server")
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js_users)

	})

	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		http.ServeFile(w, r, "home.html")

	})
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func Tree(items []*User) []*User {

	var temp []*User
	for _, ent := range items {

		if ent.Code == "" {
			fmt.Printf("admin %s,%s,%s\n", ent.Name, ent.Email, ent.Token)
			temp = getChild(items, ent.Token)
			if len(temp) > 0 {
				ent.Children = temp
				temp = nil
				temp = append(temp, ent)
			}
			break
		}
	}

	return temp
}
func getChild(items []*User, token string) []*User {
	var ch []*User
	for _, ent := range items {
		if ent.Code == token && token != "" {
			ch = append(ch, ent)
			if len(ch) > 0 {
				ent.Children = getChild(items, ent.Token)

			}
		}
	}
	return ch
}
