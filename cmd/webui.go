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
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	
	"github.com/ziotom78/stdb/db"
)

const (
	maxNumOfTestsToDisplay = 15
)

type dbEntry struct {
	ID int
	Test db.Test
}

var (
	dbConn db.Connection
	username string
)

// Show the main web page (template: mainpage.html)
func mainPage(c *gin.Context) {
	ids, _ := dbConn.GetListOfTestIDs(username, -1)
	overallNumOfTests := len(ids)

	var numOfTests = overallNumOfTests
	if overallNumOfTests > maxNumOfTestsToDisplay {
		numOfTests = maxNumOfTestsToDisplay
	}

	entries := make([]dbEntry, numOfTests)
	ids, _ = dbConn.GetListOfTestIDs(username, numOfTests)

	for idx, curID := range ids {
		entries[idx].ID = curID
		dbConn.GetTest(curID, username, &entries[idx].Test)
	}

	c.HTML(http.StatusOK, "mainpage.html", gin.H{
		"databaseSchemaVersion": db.DatabaseSchemaVersion,
		"overallNumOfTests": overallNumOfTests,
		"entries": entries,
	})
}

// Show a page containing information for a test (template: testinfo.html)
func testInformation(c *gin.Context) {
	testID, err := strconv.Atoi(c.Param("testID"))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"errorMessage": fmt.Sprintf("%v", err),
		})
		return
	}

	var test db.Test
	dbConn.GetTest(testID, username, &test)

	c.HTML(http.StatusOK, "testinfo.html", gin.H{
		"testID": testID,
		"test": test,
	})
}

// webuiCmd represents the webui command
var webuiCmd = &cobra.Command{
	Use:   "webui",
	Short: "Start the web interface to the database",
	Long: `Start a web server on the specified port of
the current host.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, err := strconv.Atoi(cmd.Flag("port").Value.String())
		dbpath := cmd.Flag("dbpath").Value.String()

		log.Printf("webui called, connecting to database at \"%s\"", dbpath)

		if err != nil {
			log.Fatalf("wrong port number %v", cmd.Flag("port").Value)
		}

		if err := dbConn.Connect(dbpath); err != nil {
			log.Fatal(err)
		}

		router := gin.Default()
		router.LoadHTMLGlob("templates/*.html")

		router.GET("/", mainPage)
		router.GET("/tests/:testID", testInformation)

		router.Run(fmt.Sprintf(":%d", port))

		dbConn.Disconnect()
	},
}

func init() {
	RootCmd.AddCommand(webuiCmd)

	webuiCmd.PersistentFlags().String("port", "8080", "Number of the port to use")
}
