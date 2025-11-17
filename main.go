package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	pokecache "pokedexcli/internal"
	"slices"
	"strings"
	"time"
)

type pokedexItem struct {
	name   string
	height int
	weight int
	stats  string
	types  string
}

var pokedex map[string]pokedexItem

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	config := config{
		previous: "",
		next:     "https://pokeapi.co/api/v2/location-area/",
		cache:    *pokecache.NewCache(5 * time.Second),
		url:      "https://pokeapi.co/api/v2/location-area/",
	}
	pokedex = make(map[string]pokedexItem)
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
		"catch": {
			name:        "catch",
			description: "Catch a Pokemon and at it to your Pokedex",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a Pokemon from your Pokedex",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List all the names of the Pokemon that you caught",
			callback:    commandPokedex,
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

func commandPokedex(config *config) error {
	fmt.Println("Your Pokedex:")
	for _, value := range pokedex {
		fmt.Printf(" - %s\n", value.name)
	}

	return nil
}

func commandCatch(config *config) error {
	if len(config.args) == 0 {
		return fmt.Errorf("provide a pokemon name for the catch command")
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", config.args[0])

	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + config.args[0])
	if err != nil {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	var pokemonResp pokemonResp

	json.Unmarshal(body, &pokemonResp)

	caught := rand.Float64() > float64(pokemonResp.BaseExp)/1000
	if caught {
		fmt.Printf("%s was caught!\n", config.args[0])
		pokedex[config.args[0]] = pokedexItem{
			name:   config.args[0],
			height: pokemonResp.Height,
			weight: pokemonResp.Weight,
			stats:  formatStats(pokemonResp.Stats),
			types:  formatTypes(pokemonResp.Types),
		}
	} else {
		fmt.Printf("%s escaped!\n", config.args[0])
	}

	return nil
}

func formatStats(stats []statsResp) string {
	var result string

	for i := range stats {
		result += fmt.Sprintf("	-%s: %d \n", stats[i].Stat.Name, stats[i].BaseStat)
	}

	return result
}

func commandInspect(config *config) error {
	if len(config.args) == 0 {
		return fmt.Errorf("provide a pokemon name for the inspect command")
	}

	pokeName := config.args[0]

	pokemon, exists := pokedex[pokeName]

	if !exists {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n Height: %d\n Weight: %d\n Stats:\n%s Types:\n%s", pokemon.name, pokemon.height, pokemon.weight, pokemon.stats, pokemon.types)

	return nil
}

func formatTypes(types []typeResp) string {
	var result string

	for i := range types {
		result += fmt.Sprintf("	- %s\n", types[i].Type.Name)
	}

	return result
}

func commandExpore(config *config) error {
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

	for i := range locationResp.PokemonEnc {
		fmt.Println(locationResp.PokemonEnc[i].Pokemon.Name)
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
	Next     string
	Previous string
	Results  []location
}

type location struct {
	Name string
}

type config struct {
	next     string
	previous string
	cache    pokecache.Cache
	args     []string
	url      string
}

type locationResp struct {
	PokemonEnc []pokemonEnc `json:"pokemon_encounters"`
}

type pokemonEnc struct {
	Pokemon pokemon `json:"pokemon"`
}

type pokemon struct {
	Name string
}

type pokemonResp struct {
	BaseExp int `json:"base_experience"`
	Height  int
	Weight  int
	Stats   []statsResp
	Types   []typeResp
}

type statsResp struct {
	BaseStat int `json:"base_stat"`
	Stat     stat
}

type stat struct {
	Name string
}

type typeResp struct {
	Type typePoke
}

type typePoke struct {
	Name string
}
