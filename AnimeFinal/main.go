package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

type Manga struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Image       string  `json:"image"`
	Background  string  `json:"background"`
	Description string  `json:"description"`
	Rating      float64 `json:"rating"`
	Opening     string  `json:"opening"`
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

type Rating struct {
	UserName   string `json:"user_name"`
	MangaTitle string `json:"manga_title"`
	Rating     int    `json:"rating"`
}

type MiddleData struct {
	User   User
	Mangas []Manga
}

type MangaMiddleData struct {
	Manga Manga
	User  User
}

var user User

func index(writer http.ResponseWriter, request *http.Request) {
	html, err := template.ParseFiles("htmls/index.html")
	if err != nil {
		log.Fatal(err)
	}

	indexMiddleData := MiddleData{
		User:   user,
		Mangas: []Manga{},
	}

	file, err := ioutil.ReadFile("db/mangas.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &indexMiddleData.Mangas)
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(writer, indexMiddleData)
}

func anime(writer http.ResponseWriter, request *http.Request) {
	html, err := template.ParseFiles("htmls/anime.html")
	if err != nil {
		log.Fatal(err)
	}

	title := request.FormValue("title")
	var allMangas []Manga
	mangaData := MangaMiddleData{
		Manga: Manga{},
		User:  user,
	}

	file, err := ioutil.ReadFile("db/mangas.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &allMangas)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range allMangas {
		if m.Title == title {
			mangaData.Manga = m
		}
	}

	html.Execute(writer, mangaData)
}

func signup(writer http.ResponseWriter, request *http.Request) {
	html, err := template.ParseFiles("htmls/signup.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(writer, user)
}

func signupFunc(writer http.ResponseWriter, request *http.Request) {
	name := request.FormValue("name")
	password := request.FormValue("password")
	repassword := request.FormValue("repassword")

	if password == repassword {
		file, _ := ioutil.ReadFile("db/users.json")
		var allUsers []User
		_ = json.Unmarshal(file, &allUsers)
		for _, temp := range allUsers {
			if temp.Name == name {
				http.Redirect(writer, request, "/", http.StatusSeeOther)
				return
			}
		}

		allUsers = append(allUsers, User{
			Name:     name,
			Password: password,
			Type:     "weeb",
		})
		file, _ = json.MarshalIndent(allUsers, "", " ")
		_ = ioutil.WriteFile("db/users.json", file, 0644)
	}

	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

func signin(writer http.ResponseWriter, request *http.Request) {
	html, err := template.ParseFiles("htmls/signin.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(writer, user)
}

func signinFunc(writer http.ResponseWriter, request *http.Request) {
	name := request.FormValue("name")
	password := request.FormValue("password")

	file, _ := ioutil.ReadFile("db/users.json")
	var allUsers []User
	_ = json.Unmarshal(file, &allUsers)

	for _, temp := range allUsers {
		if temp.Name == name && temp.Password == password {
			user = temp
			break
		}
	}

	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

func calculateRating(ratings []int) float64 {
	total := 0
	for _, r := range ratings {
		total += r
	}
	var floatTotal = float64(total)
	return floatTotal / float64(len(ratings))
}

func updateRatings(ratings []Rating) {
	file, _ := ioutil.ReadFile("db/mangas.json")
	var mangas []Manga
	_ = json.Unmarshal(file, &mangas)

	connections := map[string][]int{}

	for _, temp := range ratings {
		connections[temp.MangaTitle] = append(connections[temp.MangaTitle], temp.Rating)
	}

	var updatedMangas []Manga
	for _, manga := range mangas {
		if val, ok := connections[manga.Title]; ok {
			manga.Rating = calculateRating(val)
		}
		updatedMangas = append(updatedMangas, manga)
	}

	file1, _ := json.MarshalIndent(updatedMangas, "", " ")
	_ = ioutil.WriteFile("db/mangas.json", file1, 0644)
}

func rate(writer http.ResponseWriter, request *http.Request) {
	rating, _ := strconv.Atoi(request.FormValue("rating"))
	title := request.FormValue("title")
	if user.Name != "" && user.Type == "weeb" {
		file, _ := ioutil.ReadFile("db/ratings.json")
		var ratings []Rating
		_ = json.Unmarshal(file, &ratings)

		for _, temp := range ratings {
			if temp.UserName == user.Name && temp.MangaTitle == title {
				http.Redirect(writer, request, "/", http.StatusSeeOther)
				return
			}
		}

		ratings = append(ratings, Rating{
			UserName:   user.Name,
			MangaTitle: title,
			Rating:     rating,
		})

		updateRatings(ratings)

		file, _ = json.MarshalIndent(ratings, "", " ")
		_ = ioutil.WriteFile("db/ratings.json", file, 0644)
	}
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

func signout(writer http.ResponseWriter, request *http.Request) {
	user = User{
		Name:     "",
		Password: "",
		Type:     "",
	}
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

func new(writer http.ResponseWriter, request *http.Request) {
	html, err := template.ParseFiles("htmls/new.html")
	if err != nil {
		log.Fatal(err)
	}

	html.Execute(writer, user)
}

func newFunc(writer http.ResponseWriter, request *http.Request) {
	manga := Manga{
		Title:       request.FormValue("title"),
		Author:      request.FormValue("author"),
		Image:       request.FormValue("image"),
		Background:  request.FormValue("background"),
		Description: request.FormValue("description"),
		Rating:      0,
		Opening:     request.FormValue("opening"),
	}

	file, _ := ioutil.ReadFile("db/mangas.json")
	var allMangas []Manga
	_ = json.Unmarshal(file, &allMangas)
	for _, temp := range allMangas {
		if temp.Title == manga.Title {
			http.Redirect(writer, request, "/", http.StatusSeeOther)
			return
		}
	}

	allMangas = append(allMangas, manga)
	file, _ = json.MarshalIndent(allMangas, "", " ")
	_ = ioutil.WriteFile("db/mangas.json", file, 0644)

	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

func about(writer http.ResponseWriter, request *http.Request) {

}

func aboutFunc(writer http.ResponseWriter, request *http.Request) {

}

func embeddedDirectories() {
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))
	http.Handle("/footer/", http.StripPrefix("/footer/", http.FileServer(http.Dir("./footer/"))))
	http.Handle("/header/", http.StripPrefix("/header/", http.FileServer(http.Dir("./header/"))))
	http.Handle("/manga/", http.StripPrefix("/manga/", http.FileServer(http.Dir("./manga/"))))
	http.Handle("/db/", http.StripPrefix("/db/", http.FileServer(http.Dir("./db/"))))
}

func main() {
	embeddedDirectories()
	http.HandleFunc("/", index)
	http.HandleFunc("/anime/", anime)
	http.HandleFunc("/signup/", signup)
	http.HandleFunc("/signup", signupFunc)
	http.HandleFunc("/signin/", signin)
	http.HandleFunc("/signin", signinFunc)
	http.HandleFunc("/rate", rate)
	http.HandleFunc("/signout", signout)
	http.HandleFunc("/new/", new)
	http.HandleFunc("/new", newFunc)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
