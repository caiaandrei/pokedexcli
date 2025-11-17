package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	pokecache "pokedexcli/internal"
	"slices"
	"strings"
	"time"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	config := config{
		previous: "",
		next:     "https://pokeapi.co/api/v2/location-area/",
		cache:    *pokecache.NewCache(5 * time.Second),
		url:      "https://pokeapi.co/api/v2/location-area/",
	}
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		words := cleanupInput(scanner.Text())

		var firstWord string

		if len(words) > 0 {
			firstWord = strings.ToLower(words[0])
			config.args = words[1:]
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
		"explore": {
			name:        "explore",
			description: "Get a list of all the Pokémon from a location",
			callback:    commandExpore,
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

	cached, exists := config.cache.Get(config.previous)

	if exists {
		var locResp locationsResp

		err := json.Unmarshal(cached, &locResp)

		if err != nil {
			return nil
		}

		for i := range locResp.Results {
			fmt.Printf("%v\n", locResp.Results[i].Name)
		}
		return nil
	}

	return getLocations(config, "back")
}

func commandMap(config *config) error {

	cached, exists := config.cache.Get(config.next)

	if exists {
		var locResp locationsResp

		err := json.Unmarshal(cached, &locResp)

		if err != nil {
			return nil
		}

		for i := range locResp.Results {
			fmt.Printf("%v\n", locResp.Results[i].Name)
		}
		return nil
	}

	return getLocations(config, "forward")
}

func commandExpore(config *config) error {
	fmt.Println(config.args)
	if len(config.args) == 0 {
		return fmt.Errorf("provide a location for the explore command")
	}

	resp, err := http.Get(config.url + config.args[0])
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	var locationResp locationResp
	err = json.Unmarshal(body, &locationResp)
	if err != nil {
		return err
	}

	for i := range locationResp.PokemonResp {
		fmt.Println(locationResp.PokemonResp[i].Pokemon.Name)
	}

	return nil
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

	config.cache.Add(url, body)

	var locResp locationsResp

	err = json.Unmarshal(body, &locResp)

	if err != nil {
		return nil
	}

	for i := range locResp.Results {
		fmt.Printf("%v\n", locResp.Results[i].Name)
	}

	config.previous = locResp.Previous
	config.next = locResp.Next
	return nil
}

type locationsResp struct {
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
	cache    pokecache.Cache
	args     []string
	url      string
}

type locationResp struct {
	PokemonResp []pokemonResp `json:"pokemon_encounters"`
}

type pokemonResp struct {
	Pokemon pokemon `json:"pokemon"`
}

type pokemon struct {
	Name string
}
