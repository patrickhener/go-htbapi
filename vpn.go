package htbapi

import (
	"encoding/json"
	"fmt"
)

// Connections represents the connection details of all vpn endpoints
// within the HTB VPN environment
type Connections struct {
	Status bool                 `json:"status"`
	Data   map[string]VPNServer `json:"data"`
}

// VPNServer represents a single vpn endpoint within the HTB VPN environment
type VPNServer struct {
	CanAccess                bool           `json:"can_access"`
	LocationTypeFriendlyName string         `json:"location_type_friendly"`
	AssignedServer           AssignedServer `json:"assigned_server"`
	Available                bool           `json:"available"`
	Machine                  Machine        `json:"machine"`
}

// AssignedServer has further information about a VPNServer
type AssignedServer struct {
	ID             int    `json:"id"`
	FriendlyName   string `json:"friendly_name"`
	CurrentClients int    `json:"current_clients"`
	Location       string `json:"location"`
}

// GetCurrentVPNServer will give you VPNServer information by giving it one of the possible endpoints
// (also see EnumVPNEndpoints)
func (a *API) GetCurrentVPNServer(search string) (VPNServer, error) {
	vs := VPNServer{}

	found := false
	for _, i := range EnumVPNEndpoints {
		if search == i {
			found = true
		}
	}
	if !found {
		return vs, fmt.Errorf("you have to specify a valid vpn endpoint. Those are: %+v", EnumVPNEndpoints)
	}

	connectionsBody, _, err := a.DoRequest("/connections", nil, true, false)
	if err != nil {
		return VPNServer{}, err
	}
	defer connectionsBody.Close()

	var connections Connections
	if err := json.NewDecoder(connectionsBody).Decode(&connections); err != nil {
		return vs, err
	}

	return connections.Data[search], nil
}
