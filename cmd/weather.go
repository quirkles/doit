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
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/spf13/cobra"
)

// weatherCmd represents the weather command

func getOutboundIP() string {
	url := "https://api.ipify.org?format=text" // we are using a pulib IP API, we're using ipify here, below are some others
	// https://www.ipify.org
	// http://myexternalip.com
	// http://api.ident.me
	// http://whatismyipaddress.com/api
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(ip)
}

type CityCountryCode struct {
	City         string
	Country_code string
	Region_code  string
}

func getCityCountryCode(ip string, apiKey string) CityCountryCode {
	url := fmt.Sprintf("http://api.ipstack.com/%s?access_key=%s&format=1", ip, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	result := CityCountryCode{}
	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return result
}

type WeatherDetailsForTime struct {
	Dt   int64
	Main struct {
		Temp       float32
		Feels_like float32
	}
	Weather []struct {
		Description string
	}
}

type WeatherResponse struct {
	List []WeatherDetailsForTime
}

func getWeather(regionData CityCountryCode, apiKey string) WeatherResponse {
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?q=%s,%s,%s&appid=%s&cnt=8&units=metric", regionData.City, regionData.Region_code, regionData.Country_code, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	result := WeatherResponse{}
	jsonErr := json.Unmarshal(body, &result)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return result
}

var weatherCmd = &cobra.Command{
	Use:   "weather",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		validDates := [3]string{"now", "later", "tomorrow"}
		weatherApiKey := viper.GetString("weatherapikey")
		ipLocationApiKey := viper.GetString("iplocationapikey")
		if len(weatherApiKey) == 0 {
			fmt.Fprintln(os.Stderr, `To use the weather command you need to get an api key from openweathermap.
You can get one for free at https://home.openweathermap.org/users/sign_up
Once you have one, set it in the config with: config set weatherApiKey your-key-here`)
			return
		}
		if len(ipLocationApiKey) == 0 {
			fmt.Fprintln(os.Stderr, `To use the weather command you need to get an api key ipstack.
You can get one for free at https://ipstack.com/product
Once you have one, set it in the config with: config set ipLocationApiKey your-key-here`)
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
		ip := getOutboundIP()
		cityCountryCode := getCityCountryCode(ip, ipLocationApiKey)
		weatherResponse := getWeather(cityCountryCode, weatherApiKey)
		var weatherDetailsList []WeatherDetailsForTime
		switch date {
		case "now":
			weatherDetailsList = weatherResponse.List[0:2]
		case "later":
			weatherDetailsList = weatherResponse.List[2:4]
		case "tomorrow":
			weatherDetailsList = weatherResponse.List[4:len(weatherResponse.List)]
		}
		fmt.Fprintln(os.Stdout, "In", cityCountryCode.City)
		for _, item := range weatherDetailsList {
			tm := time.Unix(item.Dt, 0)
			fmt.Fprintln(os.Stdout, "At", tm.Format(time.Kitchen), "its going to be", item.Main.Temp, "feeling like", item.Main.Feels_like)
		}
	},
}

func itemExists(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)

	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}

	return false
}

func init() {
	rootCmd.AddCommand(weatherCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// weatherCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// weatherCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
