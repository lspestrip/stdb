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
	"log"
	"os"
	"path"
	"time"

	"database/sql"
	// Driver for accessing the Sqlite3 database file
	_ "github.com/mattn/go-sqlite3"
)

const (
	IndexFileName = "index.db"
	DatabaseSchemaVersion = "0.1.0"
)

func savePropertiesToDb(db *sql.DB, props map[string]string) error {
	for key, value := range props {
		if _, err := db.Exec(`insert or replace into properties (key, value) values (?, ?)`,
							 key, value); err != nil {
			return err
		}
	}

	return nil
}

// CreationMode specifies how a new database should be created
type CreationMode uint8
const (
	// DoNotOverwrite asks not to overwrite index.db
	DoNotOverwrite CreationMode = iota

	// Overwrite any database that happens to be in the destination path
	Overwrite
)

// CreateEmptyDatabase creates a new folder on the local disk
// and populates it with the minimum number of components needed
// for the folder to be a valid STDB database.
func CreateEmptyDatabase(dbpath string, mode CreationMode) error {
	indexFileName := path.Join(dbpath, IndexFileName)
	if mode == Overwrite {
		// Ignore the return value
		os.RemoveAll(dbpath)
	}

	log.Printf("creating directory \"%s\"", dbpath)
	// Note that os.Mkdir fails if the directory already exist.
	// This is exactly what we want!
	if err := os.Mkdir(dbpath, os.ModeDir); err != nil {
		return err
	}

	log.Printf("creating a new database file \"%s\"", indexFileName)
	db, err := sql.Open("sqlite3", indexFileName)
	if err != nil {
		return err
	}
	defer db.Close()

	tableCreationStmt := `
create table tests (
-- List of all the tests saved in the database

	test_id integer not null primary key,   -- Unique ID for this test
	short_name text,                        -- Short, easy to remember name
	description text,                       -- Full description of the test
	creation_date text not null,            -- Time when the acquisition stopped (YYYY-MM-DDTHH:MM:SS.SSS)
	user_id text not null,                  -- ID of the user which uploaded the test
	fits_checksum text,                     -- Checksum of the FITS file
	type text not null,                     -- "dc", "noise", "bandpass", ...
	time_span_sec number,                   -- Length of the acquisition, in seconds
	is_cryogenic integer not null,          -- Was the test done in cryogenic conditions? (0/1)
	polarimeter integer not null,           -- Number of the polarimeter being tested
	num_of_samples integer not null         -- Number of samples acquired during the test
);

create table users (
-- List of all the users allowed to log into the database

	user_id text not null primary key,     -- Unique ID for this user
	full_name text,                        -- Full name of the user
	creation_date text,                    -- Date of creation for this user (YYYY-MM-DDTHH:MM:SS.SSS)
	email text,                            -- Email address of the user
	password_hash text,                    -- Encrypted password
	is_enabled integer not null            -- Can this user connect to the database? (0/1)
);

create table log (
-- Log messages

	msg_id integer not null primary key, -- Unique ID for this log message
	user_id text not null,               -- ID of the user being connected at the moment
	date text,                           -- Time of the message (YYYY-MM-DDTHH:MM:SS.SSS)
	message text
);

create table attachments (
-- Attachments

	attachment_id integer not null primary key, -- Unique ID for this attachment
	test_id integer not null,                   -- Test associated with this attachment
	file_name text not null,                    -- Real name of the file
	mime_type text,                             -- MIME type of the file
	checksum text                               -- Checksum of the file
);

create table test_attachment_assoc (
-- This is used to create a N-to-M association between the "tests" and the "attachments" tables

	test_id integer not null,       -- ID of the test
	attachment_id integer not null  -- ID of the attachment
);

create table connections (
-- List of active connections

	conn_hash text not null primary key,  -- Cookie
	user_id text not null,                -- User ID
	start_time text,                      -- When this connection was established (YYYY-MM-DDTHH:MM:SS.SSS)
	is_active integer not null            -- Is this connection active? (0/1)
);

create table properties (
-- General properties of the database

	key text not null primary key,  -- Name of the key
	value text                      -- Value of the key
);
`
	if _, err := db.Exec(tableCreationStmt); err != nil {
		return err
	}

	// Save a few generic information about this database
	props := make(map[string]string)
	props["creation_date"] = time.Now().UTC().Format(time.RFC3339)
	props["stdb_version"] = DatabaseSchemaVersion
	return savePropertiesToDb(db, props)
}
