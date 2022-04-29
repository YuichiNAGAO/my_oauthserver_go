package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"oauthserver_go/utils/crypto"
)

func homeHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the home page!")
}

type Session struct {
	client                string
	state                 string
	scopes                string
	redirectUri           string
	code_challenge        string
	code_challenge_method string
	// OIDC用
	nonce string
	// IDトークンを払い出すか否か、trueならIDトークンもfalseならOAuthでトークンだけ払い出す
	oidc bool
}

type Client struct {
	id          string
	name        string
	redirectURL string
	secret      string
}

type User struct {
	id string
}

var clientInfo = Client{
	id:          "1234",
	name:        "test",
	redirectURL: "http://localhost:8080/callback",
	secret:      "secret",
}

var tmpl *template.Template

var sessionList = make(map[string]Session)

func auth(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	session := Session{
		client:      query.Get("client_id"),
		state:       query.Get("state"),
		scopes:      query.Get("scope"),
		redirectUri: query.Get("redirect_uri"),
	}
	requiredParameter := []string{"response_type", "client_id", "redirect_uri"}
	fmt.Printf("%T\n", requiredParameter)
	fmt.Printf("%T\n", query)
	// a := query["client_id"]
	// b := query.Get("client_id")

	// fmt.Printf("%T\n", a)
	// fmt.Printf("%T\n", b)

	// fmt.Println(query["client_id"])
	// fmt.Println(query.Get("client_id"))

	log.Println(session)

	log.Println(query)
	for i, v := range requiredParameter {
		log.Println(i, v)
	}

	fmt.Printf("%T\n", w)

	for _, v := range requiredParameter {
		if _, ok := query[v]; !ok {
			log.Printf("%s is missing", v)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("invalid_request. %s is missing", v)))
			return
		}

	}
	if query.Get("client_id") != clientInfo.id {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("client_id does not match"))
		return
	}

	if query.Get("response_type") != "code" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("response_type = %s is not supported", query.Get("response_type"))))
		return
	}

	sessionid := crypto.SecureRandom()
	// if !query.Has(v) {
	// 	log.Printf("%s is missing", v)
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte(fmt.Sprintf("invalid_request. %s is missing", v)))
	// 	return
	// }

	log.Print(sessionid)
	log.Printf("%T", sessionid)

	sessionList[sessionid] = session

	cookie := &http.Cookie{
		Name:  "session",
		Value: sessionid,
	}

	http.SetCookie(w, cookie)

	log.Printf("%T", tmpl)

	m := map[string]string{
		"ClientId": session.client,
		"Scope":    session.scopes,
	}

	err := tmpl.Execute(w, m)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("return login page...")

}

func authcheck(w http.ResponseWriter, r *http.Request) {
	log.Print(r.FormValue("username"))
	log.Print(r.FormValue("password"))

	username := r.FormValue("username")
	password := r.FormValue("password")

}

func main() {
	var err error
	tmpl, err = template.ParseFiles("html/tpls.html")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%T", tmpl)

	http.HandleFunc("/", homeHandle)
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/authcheck", authcheck)
	http.ListenAndServe(":8080", nil)
}
