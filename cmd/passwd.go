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

	"github.com/ziotom78/stdb/db"
)

// passwdCmd represents the passwd command
var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Change the password of an existing user",
	Long: `This command allows to change the password of an user
that was already added to the database using the "adduser" command.

This command requires a working tty terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			log.Printf("unexpected arguments %v", args)
		}

		dbpath := cmd.Flag("dbpath").Value.String()

		rl, err := readline.New("")
		if err != nil {
			log.Fatalf("error initializing input from console: %v", err)
		}
		defer rl.Close()

		rl.SetPrompt("Username: ")
		username, err := rl.Readline()
		if err != nil {
			log.Fatal(err)
		}

		oldpass, err := rl.ReadPassword("Current password: ")
		if err != nil {
			log.Fatal(err)
		}

		conn := db.Connection{}
		if err := conn.Connect(dbpath); err != nil {
			log.Fatal(err)
		}
		defer conn.Disconnect()

		oldpassFromDb, err := conn.GetUserPassword(username)
		if err != nil {
			log.Fatalf("unable to retrieve the password for user \"%s\": %v", username, err)
		}

		if bytes.Compare(db.PasswordHash(oldpass), oldpassFromDb) != 0 {
			log.Fatalf("wrong password for user \"%s\"", username)
		}

		newpass, err := AskPassword(rl, "New password: ", "Enter again the new password: ")
		if err != nil {
			log.Fatal(err)
		}

		if err := conn.ChangeUserPassword(username, db.PasswordHash(newpass)); err != nil {
			log.Fatal(err)
		}

		log.Printf("password for user \"%s\" has been changed successfully", username)
	},
}

func init() {
	RootCmd.AddCommand(passwdCmd)
}
