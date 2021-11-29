package htbapi

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

type API struct {
	BaseURL      string
	Username     string
	Password     string
	Remember     bool
	Session      *http.Client
	Token        string
	RefreshToken string
	Is2FAEnabled bool
	TokenHas2FA  bool
}

type LoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type LoginResponse struct {
	Message LoginResponseMessage `json:"message"`
}

type LoginResponseMessage struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Is2FAEnabled bool   `json:"is2FAEnabled"`
	TokenHas2FA  bool   `json:"tokenHas2FA"`
}

type OTPBody struct {
	OneTimePassword string `json:"one_time_password"`
}

type OTPLoginResponse struct {
	Message string `json:"message"`
}

type JWTPayload struct {
	AUD string `json:"aud"`
	JTI string `json:"jti"`
	IAT int    `json:"iat"`
	NBF int    `json:"nbf"`
	EXP int64  `json:"exp"`
	SUB string `json:"sub"`
}

func New(u, p string, r bool) (*API, error) {
	a := &API{
		BaseURL:  "https://www.hackthebox.com/api/v4",
		Username: u,
		Password: p,
		Remember: r,
		Session:  &http.Client{},
	}

	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}
	a.Session.Jar = jar

	// Set proxy for http client if env DEBUG=TRUE
	if os.Getenv("DEBUG") == "TRUE" {
		proxy, err := url.Parse("http://127.0.0.1:8080")
		if err != nil {
			return nil, err
		}
		a.Session.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	return a, nil
}

func (a *API) Login() error {
	if err := a.doLogin(); err != nil {
		return err
	}

	if a.Is2FAEnabled {
		if err := a.doOTPLogin(); err != nil {
			return err
		}
	}

	return nil
}

func (a *API) doLogin() error {
	body := LoginBody{
		Email:    a.Username,
		Password: a.Password,
		Remember: a.Remember,
	}

	if body.Email == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter email: ")
		email, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		body.Email = strings.ReplaceAll(email, "\n", "")
	}

	if body.Password == "" {
		fmt.Print("Enter Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			panic(err)
		}

		body.Password = string(bytePassword)
		fmt.Println("")
	}

	jsonBody, err := json.Marshal(&body)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/login", a.BaseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.Session.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Problems logging in. Status: %d", resp.StatusCode)
	}

	var respMessage LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&respMessage); err != nil {
		return err
	}

	a.Token = respMessage.Message.AccessToken
	a.RefreshToken = respMessage.Message.RefreshToken
	a.Is2FAEnabled = respMessage.Message.Is2FAEnabled
	a.TokenHas2FA = respMessage.Message.TokenHas2FA

	return nil
}

func (a *API) doOTPLogin() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter OTP: ")
	otp, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	otp = strings.Replace(otp, "\n", "", -1)
	fmt.Println("")

	otpBody := OTPBody{
		OneTimePassword: otp,
	}

	jsonOTPBody, err := json.Marshal(otpBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/2fa/login", a.BaseURL), bytes.NewBuffer(jsonOTPBody))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.Session.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var otpResp OTPLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&otpResp); err != nil {
		return err
	}

	if !strings.Contains(otpResp.Message, "correct") {
		return fmt.Errorf("%s", "Problems with otp login")
	}

	return nil
}

func (a *API) doRefreshToken() error {
	url := fmt.Sprintf("%s/login/refresh", a.BaseURL)

	type refreshBody struct {
		RefreshToken string `json:"refresh_token"`
	}

	b := refreshBody{
		RefreshToken: a.RefreshToken,
	}

	jsonBody, err := json.Marshal(&b)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Token)
	req.Header.Add("Origin", "https://app.hackthebox.com")
	req.Header.Add("Referer", "https://app.hackthebox.com/")

	resp, err := a.Session.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Problems logging in. Status: %d", resp.StatusCode)
	}

	var respMessage LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&respMessage); err != nil {
		return err
	}

	if !respMessage.Message.TokenHas2FA {
		return fmt.Errorf("%s", "missing otp, please do new login")
	}

	a.Token = respMessage.Message.AccessToken
	a.RefreshToken = respMessage.Message.RefreshToken
	a.Is2FAEnabled = respMessage.Message.Is2FAEnabled
	a.TokenHas2FA = respMessage.Message.TokenHas2FA

	return nil
}

func JWTExpired(accessToken string) (bool, error) {
	payloads := strings.Split(accessToken, ".")
	data, err := base64.StdEncoding.DecodeString(payloads[1])
	if err != nil {
		return false, err
	}

	var jwtPayload JWTPayload

	if err := json.Unmarshal(data, &jwtPayload); err != nil {
		return false, err
	}

	t := time.Now().Unix()
	if t > jwtPayload.EXP {
		return true, nil
	} else {
		return false, nil
	}
}

func (a *API) LoadSessionFromCache(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, fmt.Errorf("error reading the cache file: %+v", err)
	}

	cache, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	var sessionCache LoginResponseMessage

	if err := json.Unmarshal(cache, &sessionCache); err != nil {
		return false, err
	}

	expired, err := JWTExpired(sessionCache.AccessToken)
	if err != nil {
		return false, err
	}

	if expired {
		return true, fmt.Errorf("%s", "cached session is expired. Please login again.")
	}

	a.Token = sessionCache.AccessToken
	a.RefreshToken = sessionCache.RefreshToken
	a.Is2FAEnabled = sessionCache.Is2FAEnabled

	return false, nil
}

func (a *API) DumpSessionToCache(path string) error {
	if a.Token == "" || a.RefreshToken == "" {
		return fmt.Errorf("%s", "there is no valid session yet")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Not there, create path
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
	}

	sessionCache := LoginResponseMessage{
		AccessToken:  a.Token,
		RefreshToken: a.RefreshToken,
		Is2FAEnabled: a.Is2FAEnabled,
	}

	file, err := json.MarshalIndent(sessionCache, "", "    ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, file, 0600); err != nil {
		return err
	}

	return nil
}

func (a *API) DoRequest(endpoint string, jsonData []byte, authorized bool) (io.ReadCloser, error) {
	method := "POST"
	if jsonData == nil {
		method = "GET"
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", a.BaseURL, endpoint), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Origin", "https://app.hackthebox.com")
	req.Header.Add("Referer", "https://app.hackthebox.com/")

	if authorized {
		expired, err := JWTExpired(a.Token)
		if err != nil {
			return nil, err

		}

		if expired {
			if err := a.doRefreshToken(); err != nil {
				return nil, err
			}
		}

		req.Header.Add("Authorization", "Bearer "+a.Token)
	}

	resp, err := a.Session.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
