package htbapi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Machine will represent the data of a machine either in lab or release arena
type Machine struct {
	Active                int             `json:"active"`
	AuthUserFirstRootTime string          `json:"authUserFirstRootTime"`
	AuthUserFirstUserTime string          `json:"authUserFirstUserTime"`
	AuthUserHasReviewed   bool            `json:"auth_user_has_reviewed"`
	AuthUserInRootOwns    bool            `json:"auth_user_in_root_owns"`
	AuthUserInUserOwns    bool            `json:"auth_user_in_user_owns"`
	Avatar                string          `json:"avatar"`
	AvatarThumbUrl        string          `json:"avatar_thumb_url"`
	Difficulty            int             `json:"difficulty"`
	DifficultyText        string          `json:"difficultyText"`
	EasyMonth             int             `json:"easy_month"`
	ExpiresAt             string          `json:"expires_at"`
	FeedbackForChart      DifficultyChart `json:"feedbackForChart"`
	FirstRootBloodTime    string          `json:"firstRootBloodTime"`
	FirstUserBloodTime    string          `json:"firstUserBloodTime"`
	Free                  bool            `json:"free"`
	ID                    int             `json:"id"`
	IP                    string          `json:"ip"`
	IsCompleted           bool            `json:"isCompleted"`
	IsSpawning            bool            `json:"isSpawning"`
	IsTodo                bool            `json:"isTodo"`
	LabServer             string          `json:"lab_server"`
	Lifespan              int             `json:"lifespan"`
	Maker                 Maker           `json:"maker"`
	Maker2                Maker           `json:"maker2"`
	Name                  string          `json:"name"`
	OS                    string          `json:"os"`
	PlayInfo              PlayInfo        `json:"playInfo"`
	Points                int             `json:"points"`
	Recommended           int             `json:"recommended"`
	Release               time.Time       `json:"release"`
	Retired               int             `json:"retired"`
	RootBlood             BloodInfo       `json:"rootBlood"`
	RootBloodAvatar       string          `json:"rootBloodAvatar"`
	RootOwnsCount         int             `json:"root_owns_count"`
	SpFlag                int             `json:"sp_flag"`
	Stars                 string          `json:"stars"`
	StaticPoints          int             `json:"static_points"`
	Type                  string          `json:"type"`
	UserBlood             BloodInfo       `json:"userBlood"`
	UserBloodAvatar       string          `json:"userBloodAvatar"`
	UserOwnsCount         int             `json:"user_owns_count"`
	Voted                 bool            `json:"voted"`
	Voting                bool            `json:"voting"`
}

// MachineInstance is a wrapper around machine and vpn server info
type MachineInstance struct {
	IP      string
	Machine Machine
	Server  string
}

// PlayInfo will represent data of an active machine
type PlayInfo struct {
	ActivePlayerCount int    `json:"active_player_count"`
	ExpiresAt         string `json:"expires_at"`
	IsActive          bool   `json:"isActive"`
	IsSpawend         bool   `json:"isSpawned"`
	IsSpawning        bool   `json:"isSpawning"`
}

// Maker will hold data about a box creator
type Maker struct {
	Avatar      string `json:"avatar"`
	ID          int    `json:"id"`
	IsRespected bool   `json:"isRespected"`
	Name        string `json:"name"`
}

// Player will hold data of a htb player
type Player struct {
	Avatar string `json:"avatar"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
}

// GetMachinesResponse will be used to construct the response to /machine/list endpoint
type GetMachinesResponse struct {
	Machines []Machine `json:"info"`
}

// GetMachineResponse will be used to construct the response to /machine/profile/<id> endpoint
type GetMachineRepsonse struct {
	Machine Machine `json:"info"`
}

// BloodInfo will hold information about machines first blood
type BloodInfo struct {
	BloodDifference string `json:"blood_difference"`
	CreatedAt       string `json:"created_at"`
	User            Player `json:"user"`
}

// SpawnMachineResponse will be used to construct the response to /vm/spawn or /release_arena/spawn
type SpawnMachineResponse struct {
	Message string `json:"message"`
	Success int    `json:"success"`
}

// SpawnedMachineInfoResponse will be used to construct the response to /machine/active or /release_arena/active
type SpawnedMachineInfoResponse struct {
	Info Machine `json:"info"`
}

// Submission will represent submission details for submitting flags to /machine/own
type Submission struct {
	Difficulty int    `json:"difficulty"`
	Flag       string `json:"flag"`
	ID         int    `json:"id"`
}

// SubmissionResponse will be used to construct the response to /machine/own
type SubmissionResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Success string `json:"success"`
}

// GetAllMachines will get you a list of machines either active ones when choosing retired=false or
// retired ones if choosing retired=true
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

	body, _, err := a.DoRequest(endpoint, nil, true, false)
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

// GetMachine will get you a machine by id
func (a *API) GetMachine(id int) (Machine, error) {
	sID := strconv.Itoa(id)

	body, _, err := a.DoRequest(fmt.Sprintf("/machine/profile/%s", sID), nil, true, false)
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

// GetReleaseArenaMachine will get you the machine currently in release arena
func (a *API) GetReleaseArenaMachine() (Machine, error) {
	machines, err := a.GetAllMachines(false)
	if err != nil {
		return Machine{}, err
	}

	raServer, err := a.GetCurrentVPNServer("release_arena")
	if err != nil {
		return Machine{}, err
	}

	for _, m := range machines {
		if m.ID == raServer.Machine.ID {
			return m, nil
		}
	}

	return Machine{}, fmt.Errorf("%s", "No release arena machine found")
}

// Spawn machine will spawn a machine and give you the machine instance.
// You can choose if you want to spawn a release arena machine or a lab machine.
func (m *Machine) SpawnMachine(a *API, releaseArena bool) (MachineInstance, error) {
	switch releaseArena {
	case true:
		body, _, err := a.DoRequest("/release_arena/spawn", nil, true, true)
		if err != nil {
			return MachineInstance{}, err
		}
		defer body.Close()

		var resp SpawnMachineResponse
		if err := json.NewDecoder(body).Decode(&resp); err != nil {
			return MachineInstance{}, err
		}

		if resp.Success != 1 {
			return MachineInstance{}, fmt.Errorf("cannot spawn machine in release arena: %s", resp.Message)
		}

		mi, err := a.GetSpawnedMachineInstance(true)
		if err != nil {
			return MachineInstance{}, err
		}

		return mi, nil

	case false:
		type jsonBody struct {
			MachineID int `json:"machine_id"`
		}

		b := jsonBody{
			MachineID: m.ID,
		}

		j, err := json.Marshal(&b)
		if err != nil {
			return MachineInstance{}, err
		}

		resp, _, err := a.DoRequest("/vm/spawn", j, true, true)
		if err != nil {
			return MachineInstance{}, err
		}
		defer resp.Close()

		mi, err := a.GetSpawnedMachineInstance(false)
		if err != nil {
			return MachineInstance{}, err
		}

		return mi, nil

	default:
	}

	return MachineInstance{}, nil
}

// GetSpawnedMachineInstance will return the Machine Instance of the spawned machine either in
// release arena or in the lab.
func (a *API) GetSpawnedMachineInstance(releaseArena bool) (MachineInstance, error) {
	switch releaseArena {
	case true:
		mi := MachineInstance{}
		infoBody, _, err := a.DoRequest("/release_arena/active", nil, true, false)
		if err != nil {
			return MachineInstance{}, err
		}
		defer infoBody.Close()

		var info SpawnedMachineInfoResponse
		if err := json.NewDecoder(infoBody).Decode(&info); err != nil {
			return MachineInstance{}, err
		}
		mi.IP = info.Info.IP

		// Grab current vpn server
		raServer, err := a.GetCurrentVPNServer("release_arena")
		if err != nil {
			return MachineInstance{}, err
		}
		mi.Server = raServer.AssignedServer.FriendlyName
		mi.Machine = raServer.Machine

		return mi, nil
	case false:
		mi := MachineInstance{}

		infoBody, _, err := a.DoRequest("/machine/active", nil, true, false)
		if err != nil {
			return MachineInstance{}, err
		}
		defer infoBody.Close()

		var info SpawnedMachineInfoResponse
		if err := json.NewDecoder(infoBody).Decode(&info); err != nil {
			return MachineInstance{}, err
		}

		ma, err := a.GetMachine(info.Info.ID)
		if err != nil {
			return MachineInstance{}, err
		}

		mi.IP = ma.IP

		// Grab current vpn server
		labServer, err := a.GetCurrentVPNServer("lab")
		if err != nil {
			return MachineInstance{}, err
		}
		mi.Server = labServer.AssignedServer.FriendlyName
		mi.Machine = ma

		return mi, nil

	default:
	}
	return MachineInstance{}, nil
}

// Stop will stop the currently running machine instance
func (mi *MachineInstance) Stop(a *API, releaseArena bool) (bool, error) {
	switch releaseArena {
	case true:
		respBody, _, err := a.DoRequest("/release_arena/terminate", nil, true, true)
		if err != nil {
			return false, err
		}
		defer respBody.Close()

		var resp SpawnMachineResponse
		if err := json.NewDecoder(respBody).Decode(&resp); err != nil {
			return false, fmt.Errorf("did not terminate machine: %+v", resp.Message)
		}

	case false:
		type jsonBody struct {
			MachineID int `json:"machine_id"`
		}

		b := jsonBody{
			MachineID: mi.Machine.ID,
		}

		j, err := json.Marshal(&b)
		if err != nil {
			return false, err
		}

		respBody, _, err := a.DoRequest("/vm/terminate", j, true, true)
		if err != nil {
			return false, err
		}
		defer respBody.Close()

		return true, nil

	default:
	}

	return false, nil
}

// Submit will submit a flag to the currently running machine instance. We will have to provide diffuculty from 1 to 10 and the flag and we need to either choose releaseArena true or false accordingly
func (mi *MachineInstance) Submit(a *API, flag string, difficulty int, releaseArena bool) (bool, SubmissionResponse, error) {
	sr := SubmissionResponse{}
	if difficulty < 1 || difficulty > 10 {
		return false, sr, fmt.Errorf("%s", "Difficulty has to be between 1 and 10")
	}

	submission := Submission{
		ID:         mi.Machine.ID,
		Flag:       flag,
		Difficulty: difficulty * 10,
	}

	jsonData, err := json.Marshal(submission)
	if err != nil {
		return false, sr, err
	}

	var endpoint string
	switch releaseArena {
	case true:
		endpoint = "/release_arena/own"
	case false:
		endpoint = "/machine/own"
	}

	resp, code, err := a.DoRequest(endpoint, jsonData, true, true)
	if err != nil {
		return false, sr, err
	}
	defer resp.Close()

	var submissionResponse SubmissionResponse
	if err := json.NewDecoder(resp).Decode(&submissionResponse); err != nil {
		return false, sr, err
	}

	if code == 400 || submissionResponse.Status == 400 || submissionResponse.Message == "Incorrect Flag!" {
		return false, submissionResponse, nil
	}

	return true, submissionResponse, nil
}
