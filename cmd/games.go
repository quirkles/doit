/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

type GameData struct {
	Date string
	Timestamp int32
	Teams struct {
		Home struct {
			Name string
		}
		Away struct {
			Name string
		}
	}
	Scores struct {
		Home struct {
			Total int32
		}
		Away struct {
			Total int32
		}
	}
}

type NbaApiResponse struct {
		Response []GameData
}

func getGames(dateStr string, apiKey string) []GameData {
	url := fmt.Sprintf("https://api-basketball.p.rapidapi.com/games?season=2020-2021&league=12&date=%s&timezone=est", dateStr)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("x-rapidapi-key", apiKey)
	req.Header.Add("x-rapidapi-host", "api-basketball.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err)
	}
	result := NbaApiResponse{}
	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	games := result.Response

	sort.SliceStable(games, func(i, j int) bool {
		return games[i].Timestamp < games[j].Timestamp
	})
	return games
}

// gamesCmd represents the games command
var gamesCmd = &cobra.Command{
	Use:   "games",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		validDates := [3]string{"today", "tomorrow", "yesterday"}
		rapidApiKey := viper.GetString("rapidapikey")
		if len(rapidApiKey) == 0 {
			fmt.Fprintln(os.Stderr, `To use the nba games command you need a rapid api key
you can get one by signing up https://rapidapi.com/api-sports/api/api-nba/pricing
Once you have one, set it in the config with: config set rapidApiKey your-key-here`)
			return
		}
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Enter a date to check the weather for.")
			return
		}
		date := args[0]
		isValidDate := itemExists(validDates, date)
		if !isValidDate {
			fmt.Fprintln(os.Stderr, date, "is not a valid date to check the weather for.")
			return
		}
		dateStr := ""
		switch date {
		case "today":
			dateStr = time.Now().Format("2006-01-02")
		case "tomorrow":
			dateStr = time.Now().Add(time.Hour * 24).Format("2006-01-02")
		case "yesterday":
			dateStr = time.Now().Add(time.Hour * -24).Format("2006-01-02")
		}

		fmt.Fprintln(os.Stdout, dateStr)

		gameData := getGames(dateStr, rapidApiKey)
		for _, item := range gameData {
			t, _ := time.Parse(time.RFC3339, item.Date)
			zone, _ := time.Now().Zone()
			if zone == "EDT" {
				t = t.Add(time.Hour * 1)
			}
			fmt.Fprintln(os.Stdout, t.Format(time.Kitchen), ":", item.Teams.Away.Name, item.Scores.Away.Total, "@", item.Scores.Home.Total, item.Teams.Home.Name)
		}
	},
}

func init() {
	nbaCmd.AddCommand(gamesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gamesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gamesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
