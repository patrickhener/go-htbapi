package htbapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Machine struct {
	ID                  int              `json:"id"`
	Name                string           `json:"name"`
	OS                  string           `json:"os"`
	Points              int              `json:"points"`
	StaticPoints        int              `json:"static_points"`
	Release             time.Time        `json:"release"`
	UserOwnsCount       int              `json:"user_owns_count"`
	RootOwnsCount       int              `json:"root_owns_count"`
	AuthUserInUserOwns  bool             `json:"auth_user_in_user_owns"`
	AuthUserInRootOwns  bool             `json:"auth_user_in_root_owns"`
	IsTodo              bool             `json:"isTodo"`
	AuthUserHasReviewed bool             `json:"auth_user_has_reviewed"`
	Stars               string           `json:"stars"`
	Difficulty          int              `json:"difficulty"`
	FeedbackForChart    FeedbackForChart `json:"feedbackForChart"`
	Avatar              string           `json:"avatar"`
	DifficultyText      string           `json:"difficultyText"`
	PlayInfo            PlayInfo         `json:"playInfo"`
	Free                bool             `json:"free"`
	Maker               Maker            `json:"maker"`
	Maker2              Maker            `json:"maker2"`
	Recommended         int              `json:"recommended"`
	SpFlag              int              `json:"sp_flag"`
	EasyMonth           int              `json:"easy_month"`
	IP                  string           `json:"ip"`
}

type FeedbackForChart struct {
	CounterCake      int `json:"counterCake"`
	CounterVeryEasy  int `json:"counterVeryEasy"`
	CounterEasy      int `json:"counterEasy"`
	CounterTooEasy   int `json:"counterTooEasy"`
	CounterMedium    int `json:"counterMedium"`
	CounterBitHard   int `json:"counterBitHard"`
	CounterHard      int `json:"counterHard"`
	CounterTooHard   int `json:"counterTooHard"`
	CounterExHard    int `json:"counterExHard"`
	CounterBrainFuck int `json:"counterBrainFuck"`
}

type PlayInfo struct {
	IsSpawend         bool   `json:"isSpawned"`
	IsSpawning        bool   `json:"isSpawning"`
	IsActive          bool   `json:"isActive"`
	ActivePlayerCount int    `json:"active_player_count"`
	ExpiresAt         string `json:"expires_at"`
}

type Maker struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	IsRespected bool   `json:"isRespected"`
}

type GetActiveMachinesResponse struct {
	Info []Machine `json:"info"`
}

func (a *API) GetActiveMachines() ([]Machine, error) {
	url := fmt.Sprintf("%s/machine/list", a.BaseURL)

	bearer := fmt.Sprintf("Bearer %s", a.Token)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", bearer)
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Origin", "https://app.hackthebox.com")
	req.Header.Add("Referer", "https://app.hackthebox.com/")

	resp, err := a.Session.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Problems fetching active machines. Status: %d", resp.StatusCode)
	}

	var respMessage GetActiveMachinesResponse
	if err := json.NewDecoder(resp.Body).Decode(&respMessage); err != nil {
		return nil, err
	}

	return respMessage.Info, nil
}
