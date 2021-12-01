package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/patrickhener/go-htbapi"
)

func main() {
	// Handle Login
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	cachePath := filepath.Join(home, ".htbapi", "session.cache")

	a, err := htbapi.New("", "", true)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		// No cached session, so login with password and OTP if needed
		fmt.Println("No cached session")

		if err := a.Login(); err != nil {
			fmt.Println(err)
		}

		if err := a.DumpSessionToCache(cachePath); err != nil {
			panic(err)
		}
	} else {
		// Cached session, load from there
		fmt.Println("Found cached session - using it")

		expired, err := a.LoadSessionFromCache(cachePath)
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

	/////////////////////////////////////////////////////////////////////////////
	// Query challenges and single challenge
	/////////////////////////////////////////////////////////////////////////////

	challenges, err := a.GetAllChallenges(false)
	if err != nil {
		panic(err)
	}

	for i, c := range challenges {
		fmt.Printf("[%3d: ID %d] - %s\n", i+1, c.ID, c.Name)
	}

	gunship, err := a.GetChallenge(245)
	if err != nil {
		panic(err)
	}

	fmt.Printf("The maker of challenge '%s, Category: %s' is '%s'\n", gunship.Name, gunship.CategoryName, gunship.CreatorName)

	/////////////////////////////////////////////////////////////////////////////
}
