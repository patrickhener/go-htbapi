// Package htbapi provides a library to interact with the api/v4/ endpoint
// of hackthebox.com. It will let you do a lot of different things
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
	"time"
)

// API represents the connection details with hackthebox
// It will be provided with credentials and is the main interface to
// communicate with the api at /api/v4
type API struct {
	BaseURL      string
	Is2FAEnabled bool
	Password     string
	RefreshToken string
	Remember     bool
	Session      *http.Client
	Token        string
	TokenHas2FA  bool
	Username     string
}

// LoginBody is used to construct the json payload for /login
type LoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

// LoginResponse is used to construct the response for /login
type LoginResponse struct {
	Message LoginResponseMessage `json:"message"`
}

// LoginResponseMessage holds the login response data for /login
type LoginResponseMessage struct {
	AccessToken  string `json:"access_token"`
	Is2FAEnabled bool   `json:"is2FAEnabled"`
	RefreshToken string `json:"refresh_token"`
	TokenHas2FA  bool   `json:"tokenHas2FA"`
}

// OTPBody is used to construct the json payload for /2fa/login
type OTPBody struct {
	OneTimePassword string `json:"one_time_password"`
}

// OTPLoginResponse is used to construct the login response data for /2fa/login
type OTPLoginResponse struct {
	Message string `json:"message"`
}

// JWTPayload is used to construct the JWTToken data while parsed
type JWTPayload struct {
	AUD string `json:"aud"`
	JTI string `json:"jti"`
	IAT int    `json:"iat"`
	NBF int    `json:"nbf"`
	EXP int64  `json:"exp"`
	SUB string `json:"sub"`
}

// New will return an instantiated pointer to API.
// BaseURL is set statically to "https://www.hackthebox.com/api/v4".
// If DEBUG=TRUE is present in env http.Proxy will be set to http://127.0.0.1:8080.
// Also the connection will then ignore self signed certificates.
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
	a.Session.Timeout = 45 * time.Second
	a.Session.Transport = &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
	}

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
			TLSHandshakeTimeout: 10 * time.Second,
		}
	}

	return a, nil
}

// Login will handle the login to /login.
// It will also trigger 2FA login if needed.
// It is a wrapper function around DoLogin() and DoOTPLogin().
func (a *API) Login() error {
	if err := a.DoLogin(); err != nil {
		return err
	}

	if a.Is2FAEnabled {
		if err := a.DoOTPLogin(); err != nil {
			return err
		}
	}

	return nil
}

// DoLogin actually does the login request.
// If Email and Password are not set, it will prompt for it.
// It sets the Session details within the API struct after successful login.
func (a *API) DoLogin() error {
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
		bytePassword, err := GetPasswdMasked()
		if err != nil {
			return err
		}

		body.Password = string(bytePassword)
	}

	jsonBody, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	resp, err := a.DoRequest("/login", jsonBody, false, true)
	if err != nil {
		return err
	}
	defer resp.Close()

	var respMessage LoginResponse
	if err := json.NewDecoder(resp).Decode(&respMessage); err != nil {
		return err
	}

	a.Token = respMessage.Message.AccessToken
	a.RefreshToken = respMessage.Message.RefreshToken
	a.Is2FAEnabled = respMessage.Message.Is2FAEnabled
	a.TokenHas2FA = respMessage.Message.TokenHas2FA

	return nil
}

// DoOTPLogin will handle the 2FA OTP login. It will prompt for the login code.
func (a *API) DoOTPLogin() error {
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

	resp, err := a.DoRequest("/2fa/login", jsonOTPBody, true, true)
	if err != nil {
		return err
	}
	defer resp.Close()

	var otpResp OTPLoginResponse
	if err := json.NewDecoder(resp).Decode(&otpResp); err != nil {
		return err
	}

	if !strings.Contains(otpResp.Message, "correct") {
		return fmt.Errorf("%s", "Problems with otp login")
	}

	return nil
}

// DoRefreshToken will handle the renewal of the access_token. If it is expired
// it will pull a new one using the refresh_token
func (a *API) DoRefreshToken() error {
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

	resp, err := a.DoRequest("/login/refresh", jsonBody, true, true)
	if err != nil {
		return err
	}
	defer resp.Close()

	var respMessage LoginResponse
	if err := json.NewDecoder(resp).Decode(&respMessage); err != nil {
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

// JWTExpired is a helper function. It will take the access_token and parse
// the payload part of it. It will judge expiration based upon the 'exp' field
// in the payload. It will compare it to time.Now().Unix()
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

// LoadSessionFromCache will load a session cache file containing
// the access_token and the refresh_token. It takes a path where
// the file is stored.
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

// DumpSessionToCache will write the access_token and the refresh_token
// to disk to be read by LoadSessionFromCache. It will take a path
// where the file will be written to.
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

// DoRequest will send a request to the API endpoint. You provide the endpoint, jsonData or nil, if it will be authorized by using the Bearer Token and if it is supposed to be a POST request (otherwise it will be GET).
// It will return to you the io.ReadCloser of the responses body. Also it will throw an error if the
// HTTP Response code is other than 200.
func (a *API) DoRequest(endpoint string, jsonData []byte, authorized bool, post bool) (io.ReadCloser, error) {
	var method string
	if post {
		method = "POST"
	} else {
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
			if err := a.DoRefreshToken(); err != nil {
				return nil, err
			}
		}

		req.Header.Add("Authorization", "Bearer "+a.Token)
	}

	if jsonData != nil {
		req.Header.Add("Content-Type", "application/json")
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
