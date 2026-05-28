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

	fmt.Printf("Event created: %s, %s, %s, %s, %s, %s", name, owner, info, code, expirationTime, eventTime)

	return err
}

func GetEventById(id int) (utils.Event, error) {
	if pool == nil {
		return utils.Event{}, fmt.Errorf("database is not connected to the server")
	}

	query := `
			SELECT event_id, event_name, event_owner, event_info, event_code, expiration_time, event_time, update_need, noted_students 
			FROM events
			WHERE event_id = $1
			`

	var Event utils.Event
	err := pool.QueryRow(context.Background(), query, id).Scan(&Event.Id, &Event.Name, &Event.Owner, &Event.Info, &Event.Code, &Event.ExpirationTime, &Event.Time, &Event.UpdateNeed, &Event.NotedStudents)

	if err != nil {
		return utils.Event{}, fmt.Errorf("%s", err)
	}

	return Event, err
}

func DeleteEventById(id int) (utils.Event, error) {
	if pool == nil {
		return utils.Event{}, fmt.Errorf("database is not connected to the server")
	}

	query := `
			DELETE FROM events
			WHERE event_id = $1
			RETURNING event_id, event_name, event_owner, event_info, event_code, expiration_time, event_time, update_need, noted_students
			`
	var deletedEvent utils.Event
	err := pool.QueryRow(context.Background(), query, id).Scan(&deletedEvent.Id, &deletedEvent.Name, &deletedEvent.Owner, &deletedEvent.Info, &deletedEvent.Code, &deletedEvent.ExpirationTime, &deletedEvent.Time, &deletedEvent.UpdateNeed, &deletedEvent.NotedStudents)

	if err != nil {
		return utils.Event{}, fmt.Errorf("%s", err)
	}

	fmt.Printf("event with id=%d been deleted", id)

	return deletedEvent, err
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

func GetEventDataById(id int) (utils.EventData, error) {
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

func GetEventIdByCode(code string) (int, error) {
	if pool == nil {
		return -1, fmt.Errorf("database is not connected to the server")
	}

	query := `
			SELECT event_id FROM events where event_code=$1
			`

	var id int
	err := pool.QueryRow(context.Background(), query, code).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("%s", err)
	}

	return id, err
}

func OnCodeExistCheck(code string) (bool, error) {
	_, err := GetEventIdByCode(code)

	if err != nil {
		return false, err
	}

	return true, err
}
