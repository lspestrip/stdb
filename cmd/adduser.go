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
	"bytes"
	"log"

	"github.com/spf13/cobra"
	"github.com/chzyer/readline"

	"github.com/lspestrip/stdb/db"
)

// AskPassword gets a password to the user and asks to repeat
// it twice to verify its correctness. The string "prompt1"
// provides the prompt used the first time, while "prompt2"
// is the prompt to be used when the user is asked to re-enter
// the password
func AskPassword(rl *readline.Instance, prompt1 string, prompt2 string) ([]byte, error) {
	
	var password []byte
	var err error
	for {
		password, err = rl.ReadPassword(prompt1)
		if err != nil {
			return []byte{}, err
		}

		password2, err := rl.ReadPassword(prompt2)
		if err != nil {
			return []byte{}, err
		}

		if bytes.Compare(password, password2) != 0 {
			log.Print("passwords do not match, re-enter them")
		} else if len(password) == 0 {
			log.Print("empty passwords are not allowed, pick a valid password")
		} else {
			break
		}
	}

	return password, nil
}


// adduserCmd represents the adduser command
var adduserCmd = &cobra.Command{
	Use:   "adduser",
	Short: "Add an user to the database",
	Long: `Create an entry for a new user entitled to access the database.

This command requires a working tty terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			log.Printf("unexpected arguments %v", args)
		}

		dbpath := cmd.Flag("dbpath").Value.String()
		username := cmd.Flag("username").Value.String()
		email := cmd.Flag("email").Value.String()
		fullname := cmd.Flag("fullname").Value.String()
		disabled := cmd.Flag("disabled").Value.String()

		rl, err := readline.New("")
		if err != nil {
			log.Fatalf("error initializing input from console: %v", err)
		}
		defer rl.Close()

		if fullname == "" {
			rl.SetPrompt("Full name: ")
			fullnameFromPrompt, err := rl.Readline()
			if err != nil {
				log.Fatal(err)
			}
			fullname = fullnameFromPrompt
		}

		if email == "" {
			rl.SetPrompt("Email: ")
			emailFromPrompt, err := rl.Readline()
			if err != nil {
				log.Fatal(err)
			}
			email = emailFromPrompt
		}

		if username == "" {
			rl.SetPrompt("Username (must be unique in the database): ")
			usernameFromPrompt, err := rl.Readline()
			if err != nil {
				log.Fatal(err)
			}
			username = usernameFromPrompt
		}
		if username == "" {
			log.Fatal("you cannot specify an empty username")
		}

		password, err := AskPassword(rl, "Password: ", "Enter again the password: ")
		if err != nil {
			log.Fatal(err)
		}
		if username == "" {
			log.Fatal("blank passwords are not allowed")
		}

		conn := db.Connection{}
		if err := conn.Connect(dbpath); err != nil {
			log.Fatal(err)
		}
		defer conn.Disconnect()

		if err := conn.CreateUser(username, password, fullname, email, true); err != nil {
			log.Fatalf("unable to create user \"%s\": %v", username, err)
		}

		log.Printf("user \"%s\" created successfully in database \"%s\"", username, dbpath)

		if disabled == "true" {
			if err := conn.DisableUser(username); err != nil {
				log.Fatalf("unable to disable user \"%s\": %v", username, err)
			}
			log.Printf("user \"%s\" has been disabled", username)
		}
	},
}

func init() {
	RootCmd.AddCommand(adduserCmd)

	adduserCmd.PersistentFlags().String("username", "", "Username (must be unique)")
	adduserCmd.PersistentFlags().String("email", "", "Email of the user")
	adduserCmd.PersistentFlags().String("fullname", "", "Full name of the user")
	adduserCmd.PersistentFlags().Bool("disabled", false, "Should the user be prevented from logging in?")
}
