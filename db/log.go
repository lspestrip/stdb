package db

import (
	"fmt"
	"time"
)

// Log writes a message in the "log" table of the database
func (conn *Connection) Log(message string, username string) error {
	if !conn.Active {
		return fmt.Errorf(MsgInactiveConnection)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err := conn.Connection.Exec(`
insert into log (user_id, date, message) values (?, ?, ?)`,
		username, now, message)
	return err
}