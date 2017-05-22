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

package db

import (
	"path"

	"database/sql"
	// Driver for accessing the Sqlite3 database file
	_ "github.com/mattn/go-sqlite3"
)

// Connection is a connection to some existing database
type Connection struct {
	Active bool
	BasePath string
	Connection *sql.DB
}

const MsgInactiveConnection = "connection to the database has not been established yet"

// Connect establishes a connection to some local database.
// After having called this function successfully, you should
// defer the execution of "Disconnect".
func (conn *Connection) Connect(basepath string) error {
	conn.BasePath = basepath

	indexFileName := path.Join(basepath, IndexFileName)
	var err error
	conn.Connection, err = sql.Open("sqlite3", indexFileName)
	if err != nil {
		return err
	}

	conn.Active = true
	return nil
}

// Disconnect closes the connection with the database
func (conn *Connection) Disconnect() error {
	if conn.Active {
		result := conn.Connection.Close()
		conn.Active = false
		return result
	}

	return nil
}
