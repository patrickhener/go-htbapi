package htbapi

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Challenge represents information about a Challenge
type Challenge struct {
	AuthUserSolve        bool            `json:"authUserSolve"`
	AuthUserSolveTime    string          `json:"authUserSolveTime"`
	AVGDifficulty        int             `json:"avg_difficulty"`
	CategoryName         string          `json:"category_name"`
	ChallengeCategoryID  int             `json:"challenge_category_id"`
	Creator2Avatar       string          `json:"creator2_avatar"`
	Creator2ID           int             `json:"creator2_id"`
	Creator2Name         string          `json:"creator2_name"`
	CreatorAvatar        string          `json:"creator_avatar"`
	CreatorID            int             `json:"creator_id"`
	CreatorName          string          `json:"creator_name"`
	Description          string          `json:"description"`
	Difficulty           string          `json:"difficulty"`
	DifficultyChart      DifficultyChart `json:"difficulty_chart"`
	DifficultyChartArray []int           `json:"difficulty_chart_arr"`
	DislikeByAuthUser    bool            `json:"dislikeByAuthUser"`
	Dislikes             int             `json:"dislikes"`
	Docker               bool            `json:"docker"`
	DockerIP             string          `json:"docker_ip"`
	DockerPort           int             `json:"docker_port"`
	Download             bool            `json:"download"`
	FirstBloodTime       string          `json:"first_blood_time"`
	FirstBloodUser       string          `json:"first_blood_user"`
	FirstBloodUserAvatar string          `json:"first_blood_user_avatar"`
	FirstBloodUserID     int             `json:"first_blood_user_id"`
	ID                   int             `json:"id"`
	IsActive             bool            `json:"isActive"`
	IsCompleted          bool            `json:"isCompleted"`
	IsRespected          bool            `json:"isRespected"`
	IsRespected2         bool            `json:"isRespected2"`
	IsTodo               bool            `json:"isTodo"`
	LikeByAuthUser       bool            `json:"likeByAuthUser"`
	Likes                int             `json:"likes"`
	Name                 string          `json:"name"`
	Points               string          `json:"points"`
	Recommended          int             `json:"recommended"`
	ReleaseDate          string          `json:"release_date"`
	Retired              int             `json:"retired"`
	SHA256               string          `json:"sha256"`
	Solves               int             `json:"solves"`
	StaticPoints         string          `json:"static_points"`
	URLName              string          `json:"url_name"`
}

// DifficultyChart is the rating system
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

// GetChallengesResponse is used to construct the response to /challenge/list
type GetChallengesResponse struct {
	Challenges []Challenge `json:"challenges"`
}

// GetChallengeResponse is used to construct the response to /challenge/info/<id>
type GetChallengeRepsonse struct {
	Challenge Challenge `json:"challenge"`
}

// GetAllChallenges will return you all challenges either retired=true or retired=false (the active ones)
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

	body, err := a.DoRequest(endpoint, nil, true, false)
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

// GetChallenge will return you a certain challenge by id
func (a *API) GetChallenge(id int) (Challenge, error) {
	sID := strconv.Itoa(id)

	body, err := a.DoRequest(fmt.Sprintf("/challenge/info/%s", sID), nil, true, false)
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
