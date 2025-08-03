package torrnado

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

type authCoookes []*http.Cookie

type HttpMiddleware struct {
	rt         *RuTracker
	CookiesJar http.CookieJar
}

func (m *HttpMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode == http.StatusFound {
		m.CookiesJar.SetCookies(m.rt.RtRootUrl, resp.Cookies())
		m.rt.SessionCookies = &m.CookiesJar
		m.rt.ready = true
		m.rt.Status = "ready"
	}

	return resp, err
}

type RuTracker struct {
	RtRootUrl      *url.URL
	SessionCookies *http.CookieJar
	ready          bool
	Status         string
}

const (
	RT_ROOT      = "https://rutracker.org/forum/"
	RT_LOGIN_URL = "https://rutracker.org/forum/login.php"
	// RT_CATEGORIES_URL = "https://rutracker.org/forum/search.php"
	// RT_SEARCH_URL     = "https://rutracker.org/forum/tracker.php"
)

var ErrNotAuthenticated = errors.New("user is not authenicated. Status is NOT 302")

// ErrConfigPathMissing    = errors.New("config path is missing")
// ErrConfigFileNotExist   = errors.New("config file not found")
// ErrReadConfig           = errors.New("error while reading config")
// ErrAttchmentNotProvided = errors.New("ebook (attachment) is not provided")
// ErrQueryMissing         = errors.New("query is not provided")
// ErrRtPassword           = errors.New("password for RT is not provided")
// ErrEmptyCategoryLink    = errors.New("empty category link")

func NeedSource(config *Config) (*RuTracker, error) {
	tracker := &RuTracker{}

	err := tracker.login(config.Env[TORR_USER], config.Env[TORR_PSWD])
	if err != nil {
		return nil, err
	}

	return tracker, nil
}

func (rt *RuTracker) login(username, password string) error {
	url, err := url.Parse(RT_ROOT)
	if err != nil {
		return err
	}

	rt.RtRootUrl = url

	cookieJar, err := initCookieJar(rt)
	if err != nil {
		return err
	}
	mw := &HttpMiddleware{rt: rt}
	mw.CookiesJar = cookieJar

	creds := loginFormData(username, password)
	req, _ := http.NewRequest("POST", RT_LOGIN_URL, creds)
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	client := &http.Client{Transport: mw, Jar: mw.CookiesJar}

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	if !rt.ready {
		return ErrNotAuthenticated
	}

	return nil
}

func (rt *RuTracker) SaveTopicFile(url, filepath string) (int64, error) {
	const op = "tracker.save_topic_stream"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return  0, err
	}

	client := &http.Client{Jar: *rt.SessionCookies}

	resp, err := client.Do(req)
	if err != nil {
		return  0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%s: bad status code: %s", op, resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return 0, fmt.Errorf("%s: error creating file: %v", op, err)
	}
	defer out.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return 0, fmt.Errorf("%s: error copying data: %v", op, err)
	}

	return  n, nil
}

// func (t *RuTracker) logout() error {
// 	logout := logoutFormData()
// 	req, _ := http.NewRequest("POST", RT_LOGIN_URL, logout)
// 	req.Header.Set("content-type", "application/x-www-form-urlencoded")

// 	client := &http.Client{}
// 	_, err := client.Do(req)
// 	return err
// }

func initCookieJar(rt *RuTracker) (*cookiejar.Jar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:   "bb_guid",
		Value:  RandStringBytesMaskImpr(12),
		Path:   "/forum/",
		Domain: ".rutracker.org",
	}
	cookies = append(cookies, cookie)

	cookie = &http.Cookie{
		Name:   "bb_ssl",
		Value:  "1",
		Path:   "/forum/",
		Domain: ".rutracker.org",
	}
	cookies = append(cookies, cookie)

	cookie = &http.Cookie{
		Name:   "cf_clearance",
		Value:  "tqMeo_sSPESSzRHCnHuzuCCwELOvTrC_BDhfvwG0YZw-1723739304-1.0.1.1-0FrlWe78WtV4eGMmoEsG42.F7cXYulIN6L1ZKjIKV2ZSMD8K1cyqzBes.TGhYSQ3aPQIUrDzS_42H6JwQMYEtg",
		Path:   "/",
		Domain: ".rutracker.org",
	}
	cookies = append(cookies, cookie)
	jar.SetCookies(rt.RtRootUrl, cookies)

	return jar, nil
}

func loginFormData(username, password string) *strings.Reader {
	form := url.Values{}
	form.Add("login_username", username)
	form.Add("login_password", password)
	form.Add("login", "%C2%F5%EE%E4") // Вход (win1251, urlencoded)
	return strings.NewReader(form.Encode())
}

func logoutFormData() *strings.Reader {
	form := url.Values{}
	form.Add("logout", "1")
	return strings.NewReader(form.Encode())
}
