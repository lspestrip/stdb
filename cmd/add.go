// Copyright © 2017 Maurizio Tomasi <maurizio.tomasi@unimi.it>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/chzyer/readline"

	"github.com/ziotom78/stdb/db"
)

var (
	testShortName string // Provided by --shortname
	testDescription string // Provided by --description
	testUsername string // Provided by --username
	testType string // Provided by --type
	testCryogenicFlag bool // Provided by --cryogenic
	testPolarimeter int // Provided by --polarimeter
)

// testInfoInteractive fills the variables named "test*" (see above)
// with data provided by the user through the command line
func testInfoInteractive() error {
	rl, err := readline.New("")
	if err != nil {
		log.Fatalf("error initializing input from console: %v", err)
	}
	defer rl.Close()

	if testShortName == "" {
		rl.SetPrompt("Short name of the test: ")
		shortNamePrompt, err := rl.Readline()
		if err != nil {
			return err
		}
		testShortName = shortNamePrompt
	}

	if testDescription == "" {
		rl.SetPrompt("Description: ")
		descriptionPrompt, err := rl.Readline()
		if err != nil {
			return err
		}
		testDescription = descriptionPrompt
	}

	if testUsername == "" {
		rl.SetPrompt("Your username: ")
		usernamePrompt, err := rl.Readline()
		if err != nil {
			return err
		}
		testUsername = usernamePrompt
	}

	if testType == "" {
		rl.SetPrompt("Test type: ")
		testTypePrompt, err := rl.Readline()
		if err != nil {
			return err
		}
		testType = testTypePrompt
	}

	if testPolarimeter == 0 {
		rl.SetPrompt("Polarimeter number: ")
		testPolarimeterPrompt, err := rl.Readline()
		if err != nil {
			return err
		}

		val, err := strconv.Atoi(testPolarimeterPrompt)
		if err != nil {
			return err
		}
		testPolarimeter = val
	}

	return nil
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add one or more tests to the database",
	Long: `Insert one or more tests into the database. Each test can either
be an Excel file created by the Keithley apparatus, or a text file
saved by the application used to talk with the bias board used in
the Bicocca labs.

Information about the test that is not provided through the flags
needs to be inserted using the command line. In this case, a
working tty is required.

The first argument is the file containing the test data; the second
argument must be one of the following words:

   * warm: the test was done at room temperature;
   * cryo: the test was done at cryogenic temperatures.

Any other argument is assumed to specify attachments to be associated
with the test.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Fatal("you must specify the full path of the file containing the data" +
					  "acquired during the test, and the word \"warm\" or \"cryo\"," +
					  "depending on the environment used during the test")
		}

		testFile := args[0]
		switch strings.ToLower(args[1]) {
			case "w", "warm", "room", "rt":
			    testCryogenicFlag = false
			case "c", "cryo", "cryogenic", "ct":
			    testCryogenicFlag = true
			default:
				log.Fatalf("unrecognized string \"%s\"", args[1])
		}
		log.Printf("importing file \"%s\"", testFile)

		if err := testInfoInteractive(); err != nil {
			log.Fatal(err)
		}

		dbpath := cmd.Flag("dbpath").Value.String()
		username := cmd.Flag("username").Value.String()

		conn := db.Connection{}
		if err := conn.Connect(dbpath); err != nil {
			log.Fatal(err)
		}
		defer conn.Disconnect()

		newTest := db.Test{
			ShortName: testShortName,
			Description: testDescription,
			Username: testUsername,
			TestType: testType,
			CryogenicFlag: testCryogenicFlag,
			Polarimeter: testPolarimeter,
		}
		testID, err := conn.AddTest(&newTest, username, testFile)
		if err != nil {
			log.Fatalf("unable to add file \"%s\": %v", testFile, err)
		}

		log.Printf("new test with ID %d has been created", testID)
	},
}

func init() {
	RootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVar(&testShortName, "shortname", "", "Short name of the test")
	addCmd.Flags().StringVar(&testDescription, "description", "", "Long description of the test")
	addCmd.Flags().StringVar(&testUsername, "username", "", "Name of the user which is uploading the test")
	addCmd.Flags().StringVar(&testType, "type", "", "Type of the test (refer to the test plan report)")
	addCmd.Flags().IntVar(&testPolarimeter, "polarimeter", 0, "Number of the polarimeter being tested")
}
