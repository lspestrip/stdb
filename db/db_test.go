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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"time"
)

// This is a temporary directory that will contain the database created for the tests
var targetPath string

func TestIntegratedDatabase(t *testing.T) {
	dbPath := path.Join(targetPath, "db")
	if err := CreateEmptyDatabase(dbPath, DoNotOverwrite); err != nil {
		t.Logf("unable to create an empty database in \"%s\": %v", dbPath, err)
		t.FailNow()
	}

	var conn Connection
	if err := conn.Connect(dbPath); err != nil {
		t.Logf("unable to connect to \"%s\": %v", dbPath, err)
		t.FailNow()
	}
	defer conn.Disconnect()

	if err := conn.CreateUser("testuser", []byte("testpass"),
	                          "Mr. Test User", "user@test.inc", true); err != nil {
		t.Errorf("unable to create a new user: %v", err)
	}

	inputFilePath := path.Join("..", "testdata", "keithley_file.xls")
	refTest := Test{
		ShortName: "short",
		Description: "long description",
		CreationDate: time.Now().UTC(),
		TestType: "sweep",
		CryogenicFlag: true,
		Polarimeter: 49,
		NumOfSamples: 12345,
	}
	testID, err := conn.AddTest(&refTest, "testuser", inputFilePath);
	if err != nil || testID < 0 {
		t.Errorf("unable to add a new test to the database: %v", err)
	}

	var test Test
	if err := conn.GetTest(testID, "dummy", &test); err != nil {
		t.Errorf("Error while calling GetTest with ID=%d: %v", testID, err)
	}

	if !reflect.DeepEqual(refTest, test) {
		t.Errorf("GetTest returned the wrong test: %v instead of %v", test, refTest)
	}
}

func TestMain(m *testing.M) {
	var err error
	targetPath, err = ioutil.TempDir("", "stdb_test")
	if err != nil {
		fmt.Println("Unable to create a temporary directory for the tests, quitting")
		os.Exit(1)
	}
	defer os.RemoveAll(targetPath)

	os.Exit(m.Run())
}