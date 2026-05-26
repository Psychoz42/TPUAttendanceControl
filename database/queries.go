package database

import (
	"backend/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func CreateEvent(name string, owner string, info string, code string, expirationTime time.Time, eventTime time.Time) error {
	if pool == nil {
		return fmt.Errorf("database is not connected to the server")
	}

	query := `
			INSERT INTO events (event_name, event_owner, event_info, event_code, expiration_time, event_time, update_need)
			VALUES ($1, $2, $3, $4, $5, $6, false)
			`

	_, err := pool.Exec(context.Background(), query, name, owner, info, code, expirationTime, eventTime)

	if err != nil {
		return fmt.Errorf("event haven't created: %s", err)
	}

	fmt.Println("Event created: $1, $2, $3, $4, $5")

	return err
}

func DeleteEventById(id int) error {
	if pool == nil {
		return fmt.Errorf("database is not connected to the server")
	}

	query := `
			DELETE FROM events
			WHERE event_id = $1
			`

	_, err := pool.Exec(context.Background(), query, id)

	if err != nil {
		return fmt.Errorf("%s", err)
	}

	fmt.Printf("event with id=%d been deleted", id)

	return err
}

func GetEventCode(id int) (string, error) {
	if pool == nil {
		return "", fmt.Errorf("database is not connected to the server")
	}

	query := `
			SELECT event_code FROM events WHERE event_id=$1
			`

	var code string
	err := pool.QueryRow(context.Background(), query, id).Scan(&code)

	if err != nil {
		return "", fmt.Errorf("%s", err)
	}

	return code, err
}

func UpdateEventCode(id int, newCode string) error {
	if pool == nil {
		return fmt.Errorf("database is not connected to the server")
	}

	query := `
			UPDATE events SET event_code = $2 where event_id=$1
			`

	_, err := pool.Exec(context.Background(), query, id, newCode)

	if err != nil {
		return fmt.Errorf("%s", err)
	}

	fmt.Printf("event id=%d code been changed on %s", id, newCode)

	return err
}

func GetExpirationTime(id int) (time.Time, error) {
	if pool == nil {
		return time.Time{}, fmt.Errorf("database is not connected to the server")
	}

	query := `
			SELECT expiration_time FROM events WHERE event_id=$1
			`

	var expirationTime time.Time
	err := pool.QueryRow(context.Background(), query, id).Scan(&expirationTime)

	if err != nil {
		return time.Time{}, fmt.Errorf("%s", err)
	}

	return expirationTime, err
}

func UpdateExpirationTime(id int, newExpirationTime time.Time) error {
	if pool == nil {
		return fmt.Errorf("database is not connected to the server")
	}

	query := `
			UPDATE events SET expiration_time = $2 where event_id=$1
			`

	_, err := pool.Exec(context.Background(), query, id, newExpirationTime)

	if err != nil {
		return fmt.Errorf("%s", err)
	}

	fmt.Printf("event id=%d expiration time been changed on %s", id, newExpirationTime)

	return err
}

func GetNotedStudents(id int) ([]utils.Student, error) {
	if pool == nil {
		return nil, fmt.Errorf("database is not connected to the server")
	}

	query := `
			SELECT noted_students FROM events WHERE event_id=$1
			`

	var rawJSON []byte
	err := pool.QueryRow(context.Background(), query, id).Scan(&rawJSON)

	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	var students []utils.Student

	if err = json.Unmarshal(rawJSON, &students); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %s", err)
	}

	return students, err
}

func UpdateNotedStudents(id int, students []utils.Student) error {
	if pool == nil {
		return fmt.Errorf("database is not connected to the server")
	}

	jsonData, err := json.Marshal(students)

	if err != nil {
		return fmt.Errorf("failed to marshal students to json: %s", err)
	}

	query := `
			UPDATE events SET noted_students=$2 WHERE event_id=$1
			`

	_, err = pool.Exec(context.Background(), query, id, jsonData)

	if err != nil {
		return fmt.Errorf("%s", err)
	}

	fmt.Printf("noted students been updated")

	return err
}

func GetExpiredEventIds() ([]int, error) {
	if pool == nil {
		return nil, fmt.Errorf("database is not connected to the server")
	}

	query := `
			SELECT event_id FROM events where CURRENT_TIME > expiration_time
			`

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan event_id: %s", err)
		}
		ids = append(ids, id)
	}

	return ids, err
}

func GetIventDataById(id int) (utils.EventData, error) {
	if pool == nil {
		return utils.EventData{}, fmt.Errorf("database is not connected to the server")
	}

	query := `
			SELECT event_name, event_owner, event_info FROM events where event_id=$1
			`

	var eventData utils.EventData
	err := pool.QueryRow(context.Background(), query, id).Scan(&eventData.EventName, &eventData.EventOwner, &eventData.EventInfo)
	if err != nil {
		return utils.EventData{}, fmt.Errorf("%s", err)
	}

	return eventData, err
}

// может ли быть два однинаковых набора (event_ownew, event_code) на два разных event_id?
func GetEventIdByOwnewAndCode(owner string, code string) (utils.EventData, error) {
	if pool == nil {
		return utils.EventData{}, fmt.Errorf("database is not connected to the server")
	}

	query := `
			SELECT event_id FROM events where event_ownew=$1 and event_code=$2
			`

	var eventData utils.EventData
	_, err := pool.Exec(context.Background(), query, owner, code)
	if err != nil {
		return utils.EventData{}, fmt.Errorf("%s", err)
	}

	return eventData, err
}
