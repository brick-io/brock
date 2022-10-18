package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/rs/xid"

	"go.onebrick.io/brock"
)

func main() {
	var (
		o2 *server.Server = oauth2Server()

		GET                = http.MethodGet
		GET_POST           = http.MethodGet + "," + http.MethodPost
		GET_PUT_POST_PATCH = http.MethodGet + "," + http.MethodPut + "," + http.MethodPost + "," + http.MethodPatch
	)

	mux := brock.HTTP.Mux()
	mux.Handle(GET, "/", handleWrite(http.StatusOK, []byte{}))
	mux.Handle(GET, "/favicon.ico", handleWrite(http.StatusOK, []byte{}))
	mux.Handle(GET_POST, "/authentication", handleLogin())
	mux.Handle(GET_POST, "/oauth2/consent", handleConsent())
	mux.Handle(GET_POST, "/oauth2/access_token", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := o2.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}
	}))
	mux.Handle(GET_POST, "/oauth2/authorization", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := o2.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}
	}))
	mux.Handle(GET_PUT_POST_PATCH, "/me", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// _,_=brock.Printf("\n  %s %s\n", r.Method, "/me")
		switch r.Method {
		case http.MethodGet: // render
			handleMeGet(o2).ServeHTTP(w, r)
		case http.MethodPut: // register
			handleMePut(o2).ServeHTTP(w, r)
		case http.MethodPost: // login
			handleMePost(o2).ServeHTTP(w, r)
		case http.MethodPatch: // update
			handleMePatch(o2).ServeHTTP(w, r)
		}
	}))

	srv2 := http.Server{Addr: ":9096", Handler: mux}

	go func() { doclient() }()

	log.Fatal(srv2.ListenAndServe())
}

//nolint:gochecknoglobals
var (
	ErrUnimplemented = brock.Errorf("Unimplemented")
	ErrNotFound      = brock.Errorf("Not Found")

	seal = brock.Crypto.NaCl.Box.SealWithSharedKey
	open = brock.Crypto.NaCl.Box.OpenWithSharedKey
	// norm = strings.NewReplacer(
	// 	"=", "_",
	// 	"/", "_",
	// 	"+", "_",
	// 	"-", "_",
	// ).Replace

	cookie_key_user_token = "bk_user"

	// pub_client,
	// pvt_client, _ = brock.Crypto.NaCl.Box.Generate()
	// pub_client_b64 = strings.TrimRight(btoa(pub_client[:]), "=")
	// pvt_client_b64 = strings.TrimRight(btoa(pvt_client[:]), "=")

	clientID       = xid.New()
	pub_client_b64 = clientID.String()
	pvt_client_b64 = Cipher{expand(clientID)}.String()
	pub, pvt, key  = func() (pub, pvt, key *[32]byte) {
		if p, err := atob(string(_keypair)); err == nil && len(p) == 64 {
			pub, pvt = new([32]byte), new([32]byte)
			_, _ = copy(pub[:], p[:32]), copy(pvt[:], p[32:])
			key = brock.Crypto.NaCl.Box.SharedKey(pub, pvt)
		}

		return
	}()

	_ = func() struct{} {
		_, _ = brock.Println("KEYPAIR: ", btoa(append(pub[:], pvt[:]...)))
		_, _ = brock.Println("CLIENT ID: ", pub_client_b64)
		_, _ = brock.Println("CLIENT SE: ", pvt_client_b64)
		var p []byte
		_ = Cipher{&p}.UnmarshalJSON([]byte(pvt_client_b64))
		_, _ = brock.Println("CLIENT OK: ", bytes.Equal(expand(clientID), p))

		return struct{}{}
	}()

	_ embed.FS
	//go:embed _keypair.bin
	_keypair []byte
	//go:embed _html.consent.html
	_html_consent []byte
	//go:embed _html.login.html
	_html_login []byte
	// //go:embed _sql.select_client.sql
	// _sql_select_client string
	//go:embed _sql.select_token.sql
	_sql_select_token string
)

func expand(id xid.ID) []byte {
	p := clientID.Bytes()
	p = append(p, clientID.Machine()...)
	p = append(p, []byte(brock.Sprint(clientID.Pid()))...)
	p = append(p, []byte(clientID.Time().String())...)
	p = append(p, []byte(brock.Sprint(clientID.Counter()))...)
	return p
}

func handleConsent() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		// _,_=brock.Println("\n  handleConsent FORM", err, r.Form)
		_ = err

		q, consent := r.URL.Query(), r.Form.Get("consent")
		if consent == "cancel" {
			r.Form.Del("a")
			q.Del("a")
		}
		r.URL.RawQuery = q.Encode()

		a, err := getAction(r.URL.Query().Get("a"))
		next := a.Get("next")
		if next == "" {
			next = "/"
		}
		_ = err
		// _,_=brock.Println("\n  next:", next)
		if consent == "agree" || consent == "cancel" {
			http.Redirect(w, r, next, http.StatusTemporaryRedirect)

			return
		}

		clientID, scope := "", ""
		if u, err := url.Parse(next); err == nil {
			clientID = u.Query().Get("client_id")
			scope = u.Query().Get("scope")
		}
		scopes, scopesStr := strings.Split(scope, " "), ""
		for _, scope := range scopes {
			scopesStr += "<li>" + scope + "</li>"
		}
		scopesStr = "<ul>" + scopesStr + "</ul>"

		clientName := ""
		_ = clientID
		// if info, err := new(_storage_client).GetByID(r.Context(), client_id); err == nil {
		// 	if info, ok := info.(*client_info); ok {
		// 		client_name = info.Name
		// 	}
		// }

		// u, err := getUser(r)
		// if err != nil || u.ID == "" {
		// 	if len(a) > 0 && a.Get("next") != "" {
		// 		action := "?a=" + Cipher{"next:" + r.URL.String()}.String()
		// 		http.Redirect(w, r, "/authentication"+action, http.StatusTemporaryRedirect)
		// 		// _,_=brock.Println("\n  handleConsent", err, a.Get("next"))
		// 		return
		// 	}
		// }
		body := `
		<h2>Authorize ` + clientName + `</h2>
		<p>Hi ` + "u.Name" + `, if you click Agree then ` + clientName + ` will have access to:</p>
		` + scopesStr + `
		`

		text := bytes.ReplaceAll(_html_consent, []byte("$body"), []byte(body))
		handleWrite(http.StatusOK, text).ServeHTTP(w, r)
	})
}

func handleLogin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, err := getUser(r); err != nil || u.ID == "" {
			if a, err := getAction(r.URL.Query().Get("a")); len(a) > 0 && a.Get("next") != "" {
				_ = err
				// _,_=brock.Println("\n  handleLogin", err, a.Get("next"))
			}
		}
		msg := []byte("$message")
		text := bytes.ReplaceAll(_html_login, msg,
			[]byte(``),
		)
		text1 := bytes.ReplaceAll(_html_login, msg,
			[]byte(`<blockquote> empty username or password </blockquote>`),
		)
		text2 := bytes.ReplaceAll(_html_login, msg,
			[]byte(`<blockquote> username not found or password mismatch </blockquote>`),
		)
		switch r.Method {
		case http.MethodGet:
			handleWrite(http.StatusOK, text).ServeHTTP(w, r)
		case http.MethodPost:
			err := r.ParseForm()
			_ = err
			// _,_=brock.Println("\n  handleLogin FORM", err, r.Form)
			un, pw := r.Form.Get("username"), r.Form.Get("password")
			if len(un) < 1 || len(pw) < 1 {
				handleWrite(http.StatusBadRequest, text1).ServeHTTP(w, r)

				return
			}
			if !(un == "stevejobs" && pw == "@ppL3") {
				handleWrite(http.StatusBadRequest, text2).ServeHTTP(w, r)

				return
			}

			if old, _ := r.Cookie(cookie_key_user_token); true {
				if old == nil {
					old = new(http.Cookie)
				}
				old.Name = cookie_key_user_token
				old.Value = Cipher{User{un, "Steve Jobs"}}.String()

				old.Expires = time.Now().Add(1 * time.Hour)
				old.Secure = true
				old.HttpOnly = true
				old.SameSite = http.SameSiteStrictMode
				http.SetCookie(w, old)
			}

			if a, err := getAction(r.Form.Get("a")); err == nil && len(a) > 0 && a.Get("next") != "" {
				http.Redirect(w, r, a.Get("next"), http.StatusTemporaryRedirect)

				return
			}
		}
	})
}

func handleWrite(c int, p []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(c)
		_, _ = w.Write(p)
	})
}

func handleMeGet(o2 *server.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWrite(http.StatusOK, []byte{}).ServeHTTP(w, r)
	})
}

func handleMePut(o2 *server.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWrite(http.StatusOK, []byte{}).ServeHTTP(w, r)
	})
}

func handleMePost(o2 *server.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWrite(http.StatusOK, []byte{}).ServeHTTP(w, r)
	})
}

func handleMePatch(o2 *server.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWrite(http.StatusOK, []byte{}).ServeHTTP(w, r)
	})
}

// getAction from query param, not using hash/fragment because by default
// browser don't send this information to the server.
func getAction(a string) (u url.Values, err error) {
	if s := ""; a != "" {
		u, err = make(url.Values), Cipher{&s}.UnmarshalJSON([]byte(a))

		for _, v := range []string{"next"} {
			if strings.HasPrefix(s, v+":") {
				u.Set(v, s[len(v)+1:])

				return
			}
		}
	} else {
		err = ErrNotFound
	}
	return
}

// getUser from cookie.
func getUser(r *http.Request) (u User, err error) {
	now, c := time.Now(), (*http.Cookie)(nil)

	if c, err = r.Cookie(cookie_key_user_token); err != nil {
		//
	} else if c.Expires.After(now) {
		err = brock.Errorf("expire")
	} else if c.Value == "" {
		err = brock.Errorf("empty")
	} else {
		err = Cipher{&u}.UnmarshalJSON([]byte(c.Value))
	}
	// _,_=brock.Printf("\n  getUser err:[%s] u:[%#v]\n", err, u)
	return
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Cipher struct{ any }

func (c Cipher) String() string {
	p, _ := c.MarshalJSON()

	return string(p)
}

func (c Cipher) MarshalJSON() ([]byte, error) {
	p, err := brock.JSON.Marshal(c.any)

	return []byte(btoa(seal(p, key))), err
}

func (c Cipher) UnmarshalJSON(p []byte) error {
	p, _ = atob(string(p))
	if p, ok := open(p, key); ok && len(p) > 0 {
		return brock.JSON.Unmarshal(p, c.any)
	}
	return nil
}

func btoa(p []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(p), "=")
}

func atob(s string) ([]byte, error) {
	for len(s)%8 != 0 {
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
