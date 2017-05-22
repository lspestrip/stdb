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
	"log"
	"github.com/ziotom78/stdb/db"

	"github.com/spf13/cobra"
)

var overwriteFlag = false

// createDbCmd represents the create command
var createDbCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new empty database",
	Long: `Initialize an empty database
in the current directory. (Use --dbpath to specify
another one.)`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			log.Fatalf("unexpected arguments in the command line: %v", args)
		}
		
		dbpath := cmd.Flag("dbpath").Value.String()

		var overwriteMode db.CreationMode
		if overwriteFlag {
			overwriteMode = db.Overwrite
		} else {
			overwriteMode = db.DoNotOverwrite
		}

		err := db.CreateEmptyDatabase(dbpath, overwriteMode)
		if err != nil {
			log.Printf("unable to create database \"%s\": %v", dbpath, err)
		} else {
			log.Printf("database \"%s\" created", dbpath)
		}
	},
}

func init() {
	RootCmd.AddCommand(createDbCmd)
	createDbCmd.Flags().BoolVar(&overwriteFlag, "overwrite", false, 
                              "Overwrite any existing database")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createDbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createDbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
