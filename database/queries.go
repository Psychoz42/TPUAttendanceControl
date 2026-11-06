package database

import (
	"backend/code"
	"backend/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("not found")

// сколько раз перегенерировать код, если он совпал с уже существующим
const codeGenAttempts = 3

const eventColumns = `event_id, event_name, event_owner, event_info, event_code, expiration_time, event_time, update_need, noted_students`

func scanEvent(row pgx.Row) (utils.Event, error) {
	var event utils.Event
	err := row.Scan(&event.Id, &event.Name, &event.Owner, &event.Info, &event.Code, &event.ExpirationTime, &event.Time, &event.UpdateNeed, &event.NotedStudents)
	return event, err
}

// нарушение UNIQUE constraint, у нас это может быть только event_code
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// CreateEvent сам генерирует код события и возвращает id и код
func CreateEvent(ctx context.Context, name string, owner string, info string, eventTime time.Time, ttl time.Duration) (int, string, error) {
	query := `
			INSERT INTO events (event_name, event_owner, event_info, event_code, expiration_time, event_time, update_need)
			VALUES ($1, $2, $3, $4, now() + $5::interval, $6, true)
			RETURNING event_id
			`

	var err error
	for i := 0; i < codeGenAttempts; i++ {
		newCode := code.GenerateNewCode()

		var id int
		err = pool.QueryRow(ctx, query, name, owner, info, newCode, ttl, eventTime).Scan(&id)
		if err == nil {
			fmt.Printf("event with id=%d been created\n", id)
			return id, newCode, nil
		}
		if !isUniqueViolation(err) {
			break
		}
		fmt.Println("regenerating code")
	}

	return -1, "", fmt.Errorf("create event: %w", err)
}

func GetEventById(ctx context.Context, id int) (utils.Event, error) {
	query := `
			SELECT ` + eventColumns + `
			FROM events
			WHERE event_id = $1
			`

	event, err := scanEvent(pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return utils.Event{}, ErrNotFound
	}
	if err != nil {
		return utils.Event{}, fmt.Errorf("get event id=%d: %w", id, err)
	}

	return event, nil
}

// GetActiveEventByCode ищет событие, по коду которого сейчас можно отметиться:
// ротация не выключена и срок действия кода не истёк. Подходит и предыдущий код,
// пока не кончился его короткий льготный срок
func GetActiveEventByCode(ctx context.Context, eventCode string) (utils.Event, error) {
	query := `
			SELECT ` + eventColumns + `
			FROM events
			WHERE update_need AND (
				(event_code = $1 AND expiration_time > now())
				OR (previous_event_code = $1 AND previous_code_expiration_time > now())
			)
			`

	event, err := scanEvent(pool.QueryRow(ctx, query, eventCode))
	if errors.Is(err, pgx.ErrNoRows) {
		return utils.Event{}, ErrNotFound
	}
	if err != nil {
		return utils.Event{}, fmt.Errorf("get active event by code: %w", err)
	}

	return event, nil
}

func GetEventsByOwner(ctx context.Context, owner string) ([]utils.Event, error) {
	query := `
			SELECT ` + eventColumns + `
			FROM events
			WHERE event_owner = $1
			`

	rows, err := pool.Query(ctx, query, owner)
	if err != nil {
		return nil, fmt.Errorf("get events by owner: %w", err)
	}
	defer rows.Close()

	var events []utils.Event
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, event)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get events by owner: %w", err)
	}

	return events, nil
}

// У DeleteEventById, UpdateUpdateNeedById и RotateEventCode параметр owner ограничивает
// действие событиями этого владельца. Чужое событие выглядит как ErrNotFound, чтобы нельзя
// было перебирать id. Пустой owner значит запрос без токена: проверки нет, пока AUTH_REQUIRED выключен
func DeleteEventById(ctx context.Context, id int, owner string) (utils.Event, error) {
	query := `
			DELETE FROM events
			WHERE event_id = $1 AND ($2::text = '' OR event_owner = $2)
			RETURNING ` + eventColumns + `
			`

	deletedEvent, err := scanEvent(pool.QueryRow(ctx, query, id, owner))
	if errors.Is(err, pgx.ErrNoRows) {
		return utils.Event{}, ErrNotFound
	}
	if err != nil {
		return utils.Event{}, fmt.Errorf("delete event id=%d: %w", id, err)
	}

	fmt.Printf("event with id=%d been deleted\n", id)

	return deletedEvent, nil
}

// DeleteExpiredEvents удаляет события, код которых истёк больше чем grace назад
func DeleteExpiredEvents(ctx context.Context, grace time.Duration) ([]int, error) {
	query := `
			DELETE FROM events
			WHERE expiration_time < now() - $1::interval
			RETURNING event_id
			`

	rows, err := pool.Query(ctx, query, grace)
	if err != nil {
		return nil, fmt.Errorf("delete expired events: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan event_id: %w", err)
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("delete expired events: %w", err)
	}

	return ids, nil
}

// RotateEventCode меняет код активного события и продлевает срок. Старый код ещё grace
// принимается для отметки, но не дольше собственного срока: истёкший код не оживает.
// Возвращает ErrNotFound, если код неизвестен или ротация выключена
func RotateEventCode(ctx context.Context, oldCode string, ttl time.Duration, grace time.Duration, owner string) (string, []utils.Student, error) {
	query := `
			UPDATE events
			SET previous_event_code = event_code,
				previous_code_expiration_time = LEAST(expiration_time, now() + $4::interval),
				event_code = $2,
				expiration_time = now() + $3::interval
			WHERE event_code = $1 AND update_need AND ($5::text = '' OR event_owner = $5)
			RETURNING event_id, noted_students
			`

	var err error
	for i := 0; i < codeGenAttempts; i++ {
		newCode := code.GenerateNewCode()

		var id int
		var students []utils.Student
		err = pool.QueryRow(ctx, query, oldCode, newCode, ttl, grace, owner).Scan(&id, &students)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, ErrNotFound
		}
		if err == nil {
			fmt.Printf("event id=%d code been changed\n", id)
			return newCode, students, nil
		}
		if !isUniqueViolation(err) {
			break
		}
		fmt.Println("regenerating code")
	}

	return "", nil, fmt.Errorf("rotate event code: %w", err)
}

// AddNotedStudent атомарно и идемпотентно добавляет студента в noted_students.
// Возвращает added=false, если студент уже был отмечен
func AddNotedStudent(ctx context.Context, eventId int, login string) (bool, error) {
	query := `
			UPDATE events
			SET noted_students = noted_students || jsonb_build_array(jsonb_build_object('login', $2::text))
			WHERE event_id = $1
			  AND NOT noted_students @> jsonb_build_array(jsonb_build_object('login', $2::text))
			`

	tag, err := pool.Exec(ctx, query, eventId, login)
	if err != nil {
		return false, fmt.Errorf("add noted student to event id=%d: %w", eventId, err)
	}

	return tag.RowsAffected() == 1, nil
}

func UpdateUpdateNeedById(ctx context.Context, id int, newState bool, owner string) error {
	query := `
			UPDATE events SET update_need=$2 where event_id=$1 AND ($3::text = '' OR event_owner = $3)
			`

	tag, err := pool.Exec(ctx, query, id, newState, owner)
	if err != nil {
		return fmt.Errorf("update update_need of event id=%d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
