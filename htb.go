package htbapi

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

type API struct {
	BaseURL  string
	Username string
	Password string
	Remember bool
	Session  *http.Client
	Token    string
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
}

type OTPBody struct {
	OneTimePassword string `json:"one_time_password"`
}

type OTPLoginResponse struct {
	Message string `json:"message"`
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
	url := fmt.Sprintf("%s/login", a.BaseURL)

	body := LoginBody{
		Email:    a.Username,
		Password: a.Password,
		Remember: a.Remember,
	}

	jsonBody, err := json.Marshal(&body)
	if err != nil {
		return err
	}

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

	if respMessage.Message.Is2FAEnabled {
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

	}

	return nil
}
