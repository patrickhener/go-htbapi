package htbapi

import (
	"encoding/json"
	"fmt"
)

type Challenge struct {
	ID                   int             `json:"id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	URLName              string          `json:"url_name"`
	Retired              int             `json:"retired"`
	Difficulty           string          `json:"difficulty"`
	AVGDifficulty        int             `json:"avg_difficulty"`
	Points               string          `json:"points"`
	StaticPoints         string          `json:"static_points"`
	DifficultyChart      DifficultyChart `json:"difficulty_chart"`
	DifficultyChartArray []int           `json:"difficulty_chart_arr"`
	Solves               int             `json:"solves"`
	Likes                int             `json:"likes"`
	Dislikes             int             `json:"dislikes"`
	ReleaseDate          string          `json:"release_date"`
	IsCompleted          bool            `json:"isCompleted"`
	ChallengeCategoryID  int             `json:"challenge_category_id"`
	CategoryName         string          `json:"category_name"`
	LikeByAuthUser       bool            `json:"likeByAuthUser"`
	DislikeByAuthUser    bool            `json:"dislikeByAuthUser"`
	AuthUserSolve        bool            `json:"authUserSolve"`
	AuthUserSolveTime    string          `json:"authUserSolveTime"`
	IsActive             bool            `json:"isActive"`
	IsTodo               bool            `json:"isTodo"`
	Recommended          int             `json:"recommended"`
	FirstBloodUser       string          `json:"first_blood_user"`
	FirstBloodUserID     int             `json:"first_blood_user_id"`
	FirstBloodTime       string          `json:"first_blood_time"`
	FirstBloodUserAvatar string          `json:"first_blood_user_avatar"`
	CreatorID            int             `json:"creator_id"`
	CreatorName          string          `json:"creator_name"`
	CreatorAvatar        string          `json:"creator_avatar"`
	IsRespected          bool            `json:"isRespected"`
	Creator2ID           int             `json:"creator2_id"`
	Creator2Name         string          `json:"creator2_name"`
	Creator2Avatar       string          `json:"creator2_avatar"`
	IsRespected2         bool            `json:"isRespected2"`
	Download             bool            `json:"download"`
	SHA256               string          `json:"sha256"`
	Docker               bool            `json:"docker"`
	DockerIP             string          `json:"docker_ip"`
	DockerPort           int             `json:"docker_port"`
}

type DifficultyChart struct {
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

type GetChallengesResponse struct {
	Challenges []Challenge `json:"challenges"`
}

type GetChallengeRepsonse struct {
	Challenge Challenge `json:"challenge"`
}

func (a *API) GetAllChallenges(retired bool) ([]Challenge, error) {
	var endpoint string
	switch retired {
	case true:
		endpoint = "/challenge/list/retired"
	case false:
		endpoint = "/challenge/list"
	default:
		return nil, nil
	}

	body, err := a.DoRequest(endpoint, nil, true)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var respMessage GetChallengesResponse
	if err := json.NewDecoder(body).Decode(&respMessage); err != nil {
		return nil, err
	}

	return respMessage.Challenges, nil
}

func (a *API) GetChallenge(id string) (Challenge, error) {
	body, err := a.DoRequest(fmt.Sprintf("/challenge/info/%s", id), nil, true)
	if err != nil {
		return Challenge{}, err
	}
	defer body.Close()

	var respMessage GetChallengeRepsonse
	if err := json.NewDecoder(body).Decode(&respMessage); err != nil {
		return Challenge{}, err
	}

	return respMessage.Challenge, nil
}
