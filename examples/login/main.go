package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/patrickhener/go-htbapi"
)

func main() {
	a, err := htbapi.New("", "", true)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat("/home/patrick/.htbapi/session.cache"); os.IsNotExist(err) {
		// No cached session, so login with password
		fmt.Println("No cached session")

		if err := a.Login(); err != nil {
			fmt.Println(err)
		}

		if err := a.DumpSessionToCache("/home/patrick/.htbapi/session.cache"); err != nil {
			panic(err)
		}
	} else {
		// Cached session, load from there
		fmt.Println("Found cached session - using it")

		expired, err := a.LoadSessionFromCache("/home/patrick/.htbapi/session.cache")
		if err != nil {
			panic(err)
		}

		if expired {
			// Need to login
			fmt.Println("Cached session is expired, login again")

			if err := a.Login(); err != nil {
				fmt.Println(err)
			}

		}
	}

	// Ready to use here
	machines, err := a.GetAllMachines(false)
	if err != nil {
		panic(err)
	}

	for i, m := range machines {
		fmt.Printf("[%02d] - %s\n", i+1, m.Name)
	}

	devzat, err := a.GetMachine("398")
	if err != nil {
		panic(err)
	}
	out, _ := json.MarshalIndent(devzat, "", "    ")
	fmt.Print(string(out))

	challenges, err := a.GetAllChallenges(false)
	if err != nil {
		panic(err)
	}

	for i, c := range challenges {
		fmt.Printf("[%3d] - %s\n", i+1, c.Name)
	}

	gunship, err := a.GetChallenge("245")
	if err != nil {
		panic(err)
	}
	out, _ := json.MarshalIndent(gunship, "", "    ")
	fmt.Print(string(out))

}
