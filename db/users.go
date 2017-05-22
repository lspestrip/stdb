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
	"time"
	"crypto/sha1"
)

// PasswordHash computes the SHA1 hash of a password. The
// return value of this function is what is saved as the
// "encrypted_password" in the "users" table of the database.
func PasswordHash(password []byte) []byte {
	result := sha1.Sum(password)
	return result[:]
}

// CreateUser add a new user in the "users" table of the database
func (conn *Connection) CreateUser(
	user string, 
 	password []byte,
  	fullname string, 
  	email string,
	isEnabled bool) error {

	if !conn.Active {
		return fmt.Errorf(MsgInactiveConnection)
	}

	curDate := time.Now().UTC().Format(time.RFC3339)
	_, err := conn.Connection.Exec(`
insert into users (user_id, password_hash, full_name, creation_date, email, is_enabled)
values (?, ?, ?, ?, ?, ?)`,
		user, PasswordHash(password), fullname, curDate, email, isEnabled)

	conn.Log("new user has been created", user)

	return err
}

// EnableUser changes the status of an user to "active"
func (conn *Connection) EnableUser(user string) error {
	if !conn.Active {
		return fmt.Errorf(MsgInactiveConnection)
	}

	_, err := conn.Connection.Exec(`
update users set is_enabled = 1 where user_id = ?`,
		user)

	conn.Log("user has been enabled", user)

	return err
}

// DisableUser changes the status of an user to "active"
func (conn *Connection) DisableUser(user string) error {
	if !conn.Active {
		return fmt.Errorf(MsgInactiveConnection)
	}

	_, err := conn.Connection.Exec(`
update users set is_enabled = 0 where user_id = ?`,
		user)

	conn.Log("user has been disabled", user)

	return err
}

// GetUserPassword returns the *hashed* password for a specified user
func (conn *Connection) GetUserPassword(user string) ([]byte, error) {
	if !conn.Active {
		return []byte{}, fmt.Errorf(MsgInactiveConnection)
	}

	var password string
	err := conn.Connection.QueryRow(`
select password_hash from users where user_id = ?`,
		user).Scan(&password)
	if err != nil {
		return []byte{}, err
	}

	conn.Log("password has been requested", user)

	return []byte(password), nil
}

// ChangeUserPassword updates the password of a user in the database
func (conn *Connection) ChangeUserPassword(user string,	newPassword []byte) error {
	if !conn.Active {
		return fmt.Errorf(MsgInactiveConnection)
	}

	_, err := conn.Connection.Exec(`
update users set password_hash = ? where user_id = ?`,
		PasswordHash(newPassword), user)

	conn.Log("password has been changed", user)

	return err
}