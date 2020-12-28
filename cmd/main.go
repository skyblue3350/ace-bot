package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ConsumerKey    string `yaml:"consumer_key"`
	ConsumerSecret string `yaml:"consumer_secret"`
	AccessToken    string `yaml:"access_token"`
	AccessSecret   string `yaml:"access_secret"`
}

const (
	textFile       = "text-file"
	configFile     = "config-file"
	checkFlag      = "check"
	dryRunFlag     = "dry-run"
	tweetMaxLength = 140
)

func main() {
	var cmd = &cobra.Command{
		Use:  "ace-bot",
		RunE: run,
	}

	flags := cmd.Flags()
	flags.StringP(textFile, "t", "", "path of the text file to tweet")
	flags.StringP(configFile, "c", "", "path of the text file to tweet")
	flags.Bool(checkFlag, false, "verify if can tweet")
	flags.Bool(dryRunFlag, false, "dry-run mode")

	cmd.MarkFlagRequired(textFile)

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func getTweets(textFilePath string) ([]string, error) {
	f, err := os.Open(textFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tweets := make([]string, 0, 1)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		tweets = append(tweets, scanner.Text())
	}

	return tweets, nil
}

func getTwitterClientConfig(configFilePath string) (Config, error) {
	config := Config{}

	buffer, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(buffer, &config)
	if err != nil {
		return config, err
	}
	return config, nil

}

func run(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	textFilePath, _ := flags.GetString(textFile)
	configFilePath, _ := flags.GetString(configFile)
	check, _ := flags.GetBool(checkFlag)
	dryRun, _ := flags.GetBool(dryRunFlag)

	tweets, err := getTweets(textFilePath)
	if err != nil {
		return err
	}

	// load config
	config := Config{}
	if configFilePath == "" {
		// load from env if path is empty
		config.ConsumerKey = os.Getenv("CONSUMER_KEY")
		config.ConsumerSecret = os.Getenv("CONSUMER_SECRET")
		config.AccessToken = os.Getenv("ACCESS_TOKEN")
		config.AccessSecret = os.Getenv("ACCESS_SECRET")
	} else {
		config, err = getTwitterClientConfig(configFilePath)
		if err != nil {
			return err
		}
	}

	if check {
		// verify if can tweet
		allPassFlag := true
		for index, tweet := range tweets {
			if tweetMaxLength < len(tweet) {
				allPassFlag = false
				fmt.Printf("Error: line: %d len: %d tweet: %s\n", index+1, len(tweet), tweet)
			}
		}

		if allPassFlag {
			fmt.Println("all passed")
		} else {
			return errors.New("There is a character string that cannot be tweeted")
		}

	} else {
		// random tweet
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(tweets))
		tweet := tweets[index]

		if dryRun {
			fmt.Print("dry-run mode: ")
		} else {
			oauthConfig := oauth1.NewConfig(config.ConsumerKey, config.ConsumerSecret)
			oauthToken := oauth1.NewToken(config.AccessToken, config.AccessSecret)

			httpClient := oauthConfig.Client(oauth1.NoContext, oauthToken)
			client := twitter.NewClient(httpClient)

			_, _, err := client.Statuses.Update(tweet, nil)
			if err != nil {
				return err
			}
		}
		fmt.Printf("tweet: %s", tweet)
	}

	return nil
}
