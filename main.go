package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"oauthserver_go/utils/crypto"
)

type Session struct {
	client_id             string
	state                 string
	scopes                string
	redirect_uri          string
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
	id       int
	name     string
	password string
}

type AuthCode struct {
	user         string
	client_id    string
	scopes       string
	redirect_uri string
	expires_at   int64
}

type Token struct {
	user  string
	client_id string
	scopes string
	expires_at int64
}

var clientInfo = Client{
	id:          "1234",
	name:        "test",
	redirectURL: "http://localhost:8080/callback",
	secret:      "secret",
}

var testUser = User{
	id:       1111,
	name:     "test",
	password: "hoge",
}

var tmpl *template.Template

var sessionList = make(map[string]Session)

var authCodeList = make(map[string]AuthCode)

func auth(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	session := Session{
		client_id:    query.Get("client_id"),
		state:        query.Get("state"),
		scopes:       query.Get("scope"),
		redirect_uri: query.Get("redirect_uri"),
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
		"ClientId": session.client_id,
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

	if username != testUser.name || password != testUser.password {
		log.Printf("%s coundn't login", username)
		w.Write([]byte("Login failed"))
		return
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%T", cookie)
	v := sessionList[cookie.Value]
	log.Print(v)

	authCodeid := crypto.SecureRandom()

	authCode := AuthCode{
		user:         username,
		client_id:    v.client_id,
		scopes:       v.scopes,
		redirect_uri: v.redirect_uri,
		expires_at:   time.Now().Unix() + 300, //単位は秒
	}

	authCodeList[authCodeid] = authCode

	// log.Printf("authCode: %T", authCode)
	log.Printf("authCode: %v", authCode)

	location := fmt.Sprintf("%s?code=%s&state=%s", v.redirect_uri, authCodeid, v.state)
	w.Header().Add("Location", location)
	w.WriteHeader(302)
}

func token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	query := r.Form
	required_params := []string{"grant_type", "code", "client_id", "redirect_uri"}

	// log.Printf("%T", query)
	// log.Print(query)
	// log.Print(query["grant_type"])
	// a, b := query["grant_type"]
	// log.Print(a)
	// log.Print(b)
	// log.Printf("%T", query["grant_type"])

	for _, v := range required_params {
		if _, ok := query[v]; !ok {
			log.Printf("%s is missiong", v)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("invalid request. %s is missing", v)))
			return
		}
	}

	if query.Get("grant_type") != "authorization_code" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request. Only authorization code flow is supported.\n"))
		return
	}
	log.Printf("%T", query.Get("code"))

	v, ok := authCodeList[query.Get("code")]
	if !ok {
		log.Println("Authrization code doesn't exist")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No authrization code.\n"))
		return
	}
	log.Print(v)

	if v.client_id != query.Get("client_id") {
		log.Println("client_id doesn't match")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid_request. Client id is invalid.\n"))
		return
	}

	if v.redirect_uri != query.Get("redirect_uri") {
		log.Println("redirect_uri doesn't match")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid_request. Redirect uri is invalid.\n"))
		return
	}

	if v.expires_at <  time.Now().Unix() {
		log.Println("Authrization code expired")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid_request. Authrization code expired.\n")))
		return
	}


	// log.Printf("%T", query)
	// log.Print(query)
	// log.Print(r)

	// for _, v := range required_params {
	// 	if query[]
	// }

}

func main() {
	var err error
	tmpl, err = template.ParseFiles("html/tpls.html")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%T", tmpl)
	tt := time.Now().Unix()
	t := time.Now()
	log.Printf("%T", tt)
	log.Print(tt)
	log.Printf("%T", t)
	log.Print(t)
	time.Sleep(time.Second * 1)
	ttr := time.Now().Unix()
	log.Print(ttr)

	http.HandleFunc("/auth", auth)
	http.HandleFunc("/authcheck", authcheck)
	http.HandleFunc("/token", token)
	http.ListenAndServe(":8080", nil)
}

/*
- クライアント: web アプリ https://client.example.com
  - リダイレクト URI: https://client.example.com/cb
- 認可サーバー:
  - 認可エンドポイント: https://server.example.com/authorize
  - トークンエンドポイント: https://server.example.com/token

https://zenn.dev/zaki_yama/articles/oauth2-authorization-code-grant-and-pkce#10.-%E8%AA%8D%E5%8F%AF%E3%82%B3%E3%83%BC%E3%83%89%E7%99%BA%E8%A1%8C
https://qiita.com/TakahikoKawasaki/items/e508a14ed960347cff11
*/
