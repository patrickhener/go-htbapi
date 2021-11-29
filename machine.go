package htbapi

import (
	"encoding/json"
	"fmt"
	"time"
)

type Machine struct {
	ID                    int             `json:"id"`
	Name                  string          `json:"name"`
	OS                    string          `json:"os"`
	Active                int             `json:"active"`
	Retired               int             `json:"retired"`
	IsCompleted           bool            `json:"isCompleted"`
	Points                int             `json:"points"`
	StaticPoints          int             `json:"static_points"`
	Release               time.Time       `json:"release"`
	UserOwnsCount         int             `json:"user_owns_count"`
	RootOwnsCount         int             `json:"root_owns_count"`
	AuthUserInUserOwns    bool            `json:"auth_user_in_user_owns"`
	AuthUserInRootOwns    bool            `json:"auth_user_in_root_owns"`
	AuthUserFirstUserTime string          `json:"authUserFirstUserTime"`
	AuthUserFirstRootTime string          `json:"authUserFirstRootTime"`
	IsTodo                bool            `json:"isTodo"`
	AuthUserHasReviewed   bool            `json:"auth_user_has_reviewed"`
	Stars                 string          `json:"stars"`
	Difficulty            int             `json:"difficulty"`
	FeedbackForChart      DifficultyChart `json:"feedbackForChart"`
	Avatar                string          `json:"avatar"`
	DifficultyText        string          `json:"difficultyText"`
	PlayInfo              PlayInfo        `json:"playInfo"`
	Free                  bool            `json:"free"`
	Maker                 Maker           `json:"maker"`
	Maker2                Maker           `json:"maker2"`
	Recommended           int             `json:"recommended"`
	SpFlag                int             `json:"sp_flag"`
	EasyMonth             int             `json:"easy_month"`
	IP                    string          `json:"ip"`
	UserBlood             BloodInfo       `json:"userBlood"`
	UserBloodAvatar       string          `json:"userBloodAvatar"`
	RootBlood             BloodInfo       `json:"rootBlood"`
	RootBloodAvatar       string          `json:"rootBloodAvatar"`
	FirstUserBloodTime    string          `json:"firstUserBloodTime"`
	FirstRootBloodTime    string          `json:"firstRootBloodTime"`
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

type Player struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type GetMachinesResponse struct {
	Machines []Machine `json:"info"`
}

type GetMachineRepsonse struct {
	Machine Machine `json:"info"`
}

type BloodInfo struct {
	User            Player `json:"user"`
	CreatedAt       string `json:"created_at"`
	BloodDifference string `json:"blood_difference"`
}

func (a *API) GetAllMachines(retired bool) ([]Machine, error) {
	var endpoint string
	switch retired {
	case true:
		endpoint = "/machine/list/retired"
	case false:
		endpoint = "/machine/list"
	default:
		return nil, nil
	}

	body, err := a.DoRequest(endpoint, nil, true)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var respMessage GetMachinesResponse
	if err := json.NewDecoder(body).Decode(&respMessage); err != nil {
		return nil, err
	}

	return respMessage.Machines, nil
}

func (a *API) GetMachine(id string) (Machine, error) {
	body, err := a.DoRequest(fmt.Sprintf("/machine/profile/%s", id), nil, true)
	if err != nil {
		return Machine{}, err
	}
	defer body.Close()

	var respMessage GetMachineRepsonse
	if err := json.NewDecoder(body).Decode(&respMessage); err != nil {
		return Machine{}, err
	}

	return respMessage.Machine, nil
}
