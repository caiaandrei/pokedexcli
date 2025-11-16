package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	config := config{
		previous: "",
		next:     "https://pokeapi.co/api/v2/location-area/",
	}
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		words := cleanupInput(scanner.Text())

		var firstWord string

		if len(words) > 0 {
			firstWord = strings.ToLower(words[0])
		} else {
			firstWord = ""
		}

		commands := getCommands()
		command, ok := commands[firstWord]

		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		_ = command.callback(&config)
	}
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

func getCommands() map[string]cliCommand {

	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the names of 20 location areas in the Pokemon world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 locations areas in the Pokemon world",
			callback:    commandMapb,
		},
	}
}

func cleanupInput(text string) []string {

	trimed := strings.TrimSpace(text)
	words := strings.Split(trimed, " ")
	words = slices.DeleteFunc(words, func(word string) bool {
		return len(strings.TrimSpace(word)) == 0
	})

	return words
}

func commandExit(config *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *config) error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	commands := getCommands()

	for _, command := range commands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	fmt.Println()
	return nil
}

func commandMapb(config *config) error {
	return getLocations(config, "back")
}

func commandMap(config *config) error {

	return getLocations(config, "forward")
}

func getLocations(config *config, direction string) error {

	var url string

	if direction == "back" {
		if config.previous == "" {
			fmt.Println("You're on the first page")
			return nil
		}
		url = config.previous
	} else {
		url = config.next
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		return err
	}

	var locResp locationResp

	err = json.Unmarshal(body, &locResp)

	if err != nil {
		return nil
	}

	for i := range locResp.Results {
		fmt.Printf("%v\n", locResp.Results[i].Name)
	}

	config.next = locResp.Next
	config.previous = locResp.Previous

	return nil
}

type locationResp struct {
	// count   int
	Next     string
	Previous string
	Results  []location
}

type location struct {
	Name string
	//Url  string
}

type config struct {
	next     string
	previous string
}
