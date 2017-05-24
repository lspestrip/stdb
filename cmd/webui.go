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
	"golang.org/x/crypto/bcrypt"
	"github.com/satori/go.uuid"

	"github.com/ziotom78/stdb/db"
	"github.com/ziotom78/stdb/web"
)

const (
	maxNumOfTestsToDisplay = 15
	sessionCookieName = "session_cookie"
)

type dbEntry struct {
	ID int
	Test db.Test
}

var (
	dbConn db.Connection
	username string
)

func authenticate(c *gin.Context) {
	formUsername := c.PostForm("username")
	formPassword := []byte(c.PostForm("password"))

	password, err := dbConn.GetUserPassword(formUsername)
	if err != nil || bcrypt.CompareHashAndPassword(password, formPassword) != nil {
		dbConn.Log("failed authentication", formUsername)
		c.HTML(http.StatusUnauthorized, "error.html", gin.H{
			"errorMessage": fmt.Sprintf("access denied (%v)", err),
		})
		return
	}
	
	session, _ := web.CreateSession(formUsername)
	c.SetCookie(sessionCookieName, session.UUID.String(), 0, "", "", false, true)
	c.Redirect(http.StatusMovedPermanently, "/")
}

// isCookieValid checks if a session cookie represents an authenticated session.
// If it is so, it returns "true" and the name of the logged user; otherwise,
// it returns an error code.
func isCookieValid(c *gin.Context) (bool, web.Session, error) {
	idStr, err := c.Cookie(sessionCookieName)
	if err != nil {
		return false, web.Session{}, err
	}
	cookieUUID, err := uuid.FromString(idStr)
	if err != nil {
		return false, web.Session{}, err
	}

	session, err := web.FindSessionByUUID(cookieUUID)
	if err != nil {
		return false, web.Session{}, err
	}

	return true, session, nil
}

// This is a wrapper for HTTP methods that prevents them from
// being called when the user has not logged in yet.
func protect(h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		valid, session, err := isCookieValid(c)
		if valid {
			dbConn.Log(fmt.Sprintf("granting permission to page"), session.Username)
			h(c) // Chain
		} else {
			dbConn.Log(fmt.Sprintf("in protect: error %v", err), "")
			c.HTML(http.StatusServiceUnavailable, "error.html", gin.H{
				"errorMessage": "Access denied",
			})
		}
	}
}

// Disconnect the user
func logout(c *gin.Context) {
	valid, session, err := isCookieValid(c)
	if ! valid {
		c.HTML(http.StatusServiceUnavailable, "error.html", gin.H{
			"errorMessage": fmt.Sprintf("unable to disconnect properly: %v", err),
		})
		return
	}

	// Delete this cookie
	c.SetCookie(sessionCookieName, "", -1, "", "", false, true)
	c.Redirect(http.StatusMovedPermanently, "/")
	dbConn.Log(fmt.Sprintf("user has logged out"), session.Username)
}

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

	loggedIn, session, _ := isCookieValid(c)
	c.HTML(http.StatusOK, "mainpage.html", gin.H{
		"databaseSchemaVersion": db.DatabaseSchemaVersion,
		"overallNumOfTests": overallNumOfTests,
		"entries": entries,
		"loggedIn": loggedIn,
		"username": session.Username,
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
		router.GET("/tests/:testID", protect(testInformation))
		router.POST("/authenticate", authenticate)
		router.GET("/logout", protect(logout))

		router.Run(fmt.Sprintf(":%d", port))

		dbConn.Disconnect()
	},
}

func init() {
	RootCmd.AddCommand(webuiCmd)

	webuiCmd.PersistentFlags().String("port", "8080", "Number of the port to use")
}
