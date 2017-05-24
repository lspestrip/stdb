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
	"fmt"
	"time"
	"github.com/satori/go.uuid"
)

var (
	sessions []Session
)

// Session contains the state of an active session
type Session struct {
	Username string
	UUID uuid.UUID
	CreatedAt time.Time
}

// CreateSession creates a session associated with the given username.
// If the username was already connected, return the session and a
// bool flag set to "false"; otherwise, create a new session and
// return a bool flag set to "true".
func CreateSession(username string) (Session, bool) {
	existingSession, err := FindSessionByUsername(username)
	if err != nil {
		newSession := Session{
			Username: username,
			UUID: uuid.NewV4(),
			CreatedAt: time.Now().UTC(),
		}

		sessions = append(sessions, newSession)
		return newSession, true
	}

	return existingSession, false
}


// DeleteSession removes an element from the list of active sessions.
// If "session" is not in the list of active sessions, nothing happens.
func DeleteSession(session Session) {
	for idx, curSession := range sessions {
		if curSession.UUID == session.UUID {
			// Delete from "sessions" without preserving the order
			sessions[idx] = sessions[len(sessions) - 1]
			sessions = sessions[:len(sessions) - 1]

			// Quit immediately, as we cannot keep iterating
			// on the slice now that we have modified it
			break
		}
	}
}

// FindSessionByUsername finds an active session matching the username.
// If no session is found, an error is returned. This is the only
// way this function can fail.
func FindSessionByUsername(username string) (Session, error) {
	for _, curSession := range sessions {
		if curSession.Username == username {
			return curSession, nil
		}
	}
	return Session{}, fmt.Errorf("no session with username %s is active", username)
}

// FindSessionByUUID finds an active session matching the UUID.
// If no session is found, an error is returned. This is the only
// way this function can fail.
func FindSessionByUUID(UUID uuid.UUID) (Session, error) {
	for _, curSession := range sessions {
		if curSession.UUID == UUID {
			return curSession, nil
		}
	}

	return Session{}, fmt.Errorf("no session with UUID %v is active", UUID)
}
