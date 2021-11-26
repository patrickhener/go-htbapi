package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/patrickhener/go-htbapi"
	"golang.org/x/term"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	fmt.Println("")

	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}

	password := string(bytePassword)
	fmt.Println("")

	a, err := htbapi.New(strings.TrimSpace(username), password, true)
	if err != nil {
		panic(err)
	}

	if err := a.Login(); err != nil {
		panic(err)
	}

	fmt.Println("[+] Login successful")

}
