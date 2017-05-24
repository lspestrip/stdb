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

package web

import (
	"reflect"
	"testing"
)

func TestCreateSession(t *testing.T) {
	sess1, isNew := CreateSession("foo")
	if ! isNew {
		t.Error("isNew should be true")
	}
	if sess1.Username != "foo" {
		t.Errorf("session for user foo has user %s", sess1.Username)
	}

	if _, isNew := CreateSession("foo"); isNew {
		t.Error("session clash went undetected")
	}

	sess2, isNew := CreateSession("bar")
	if ! isNew {
		t.Error("isNew should be true")
	}
	if sess2.Username != "bar" {
		t.Errorf("session for user bar has user %s", sess2.Username)
	}

	if sess2.CreatedAt.Unix() < sess1.CreatedAt.Unix() {
		t.Errorf("creation time for session 2 (%v) is before time for session 1 (%v)",
		         sess2.CreatedAt, sess1.CreatedAt)
	}

	if matchSession, _ := FindSessionByUsername("foo"); ! reflect.DeepEqual(matchSession, sess1) {
		t.Errorf("wrong return value for FindSessionByUsername: %v instead of %v",
		         matchSession, sess1)
	}

	if matchSession, _ := FindSessionByUUID(sess2.UUID); ! reflect.DeepEqual(matchSession, sess2) {
		t.Errorf("wrong return value for FindSessionByUUID: %v instead of %v",
		         matchSession, sess2)
	}

	DeleteSession(sess2)

	if _, err := FindSessionByUsername("bar"); err == nil {
		t.Errorf("DeleteSession does not seem to work")
	}
}
