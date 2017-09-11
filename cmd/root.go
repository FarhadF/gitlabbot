// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var cfgFile string
var (
	dbHost        string
	dbPort        int
	dbName        string
	dbUser        string
	dbPassword    string
	gitlabBase    string
	gitlabToken   string
	lgtmTreashold int
	gitlabBot     string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gitlabbot",
	Short: "Gitlab Munger",
	Long: `Golang implementation of Gitlab Munger, with following features:
- Commenting
- Merging
- LGTM Counts
Using Gitlab database as well as webhook on the project.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: rootCmd,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func rootCmd(cmd *cobra.Command, args []string) {
	if versionFlag := getFlagBoolPtr(cmd, "version"); versionFlag != nil {
		fmt.Println("GitlabBot v1.0.0")
	} else {
		gitlabbot(dbHost, dbPort, dbName, dbUser, dbPassword, gitlabBase, gitlabToken, lgtmTreashold, gitlabBot)
	}
}

func getFlagBoolPtr(cmd *cobra.Command, flag string) *bool {
	f := cmd.Flags().Lookup(flag)
	if f == nil {
		log.Printf("Flag accessed but not defined for command %s: %s", cmd.Name(), flag)
	}
	// Check if flag was not set at all.
	if !f.Changed && f.DefValue == f.Value.String() {
		return nil
	}
	var ret bool
	// Caseless compare.
	if strings.ToLower(f.Value.String()) == "true" {
		ret = true
	} else {
		ret = false
	}
	return &ret
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	RootCmd.Flags().BoolP("version", "v", false, "Prints version info")
	RootCmd.Flags().StringVarP(&dbHost, "dbhost", "H", "localhost", "Postgres database Hostname/IPAddress")
	RootCmd.Flags().IntVarP(&dbPort, "dbport", "P", 5432, "Postgres database port number")
	RootCmd.Flags().StringVarP(&dbName, "dbname", "n", "gitlabhq_production", "Gitlab database name")
	RootCmd.Flags().StringVarP(&dbUser, "dbuser", "u", "gitlab", "Gitlab database username")
	RootCmd.Flags().StringVarP(&dbPassword, "dbpassword", "p", `Aa111111`, "Gitlab database password")
	RootCmd.Flags().StringVarP(&gitlabBase, "gitlabbase", "b", "http://localhost:10080", "Gitlab base url")
	RootCmd.Flags().StringVarP(&gitlabToken, "gitlabtoken", "t", "K8F8SZEHyq4Dm9osdTT3", "Gitlab user token for API access")
	RootCmd.Flags().IntVarP(&lgtmTreashold, "lgtmtreashold", "l", 2, "Number of LGTMs required to merge the request")
	RootCmd.Flags().StringVarP(&gitlabBot, "gitlabbot", "g", "gitlabbot", "Gitlab username for the bot, Case sensetive")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".gitlabbot" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gitlabbot")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
