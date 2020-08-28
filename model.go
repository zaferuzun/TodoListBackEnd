// model.go

package main

import (
	"database/sql"
	"fmt"
)

type task struct {
	TaskID         int    `json:"TaskId"`
	TaskDefination string `json:"TaskDefination"`
	TaskMarked     bool   `json:"TaskMarked"`
}

func (t *task) updateTask(db *sql.DB) error {
	statement := fmt.Sprintf("UPDATE task SET TaskDefination='%s', TaskMarked=%t WHERE TaskId=%d", t.TaskDefination, t.TaskMarked, t.TaskID)
	_, err := db.Exec(statement)
	return err
}
func (t *task) getTask(db *sql.DB) error {
	statement := fmt.Sprintf("SELECT TaskDefination,TaskMarked FROM task WHERE TaskId=%d", t.TaskID)
	return db.QueryRow(statement).Scan(&t.TaskDefination, &t.TaskMarked)
}

func (t *task) deleteTask(db *sql.DB) error {
	statement := fmt.Sprintf("DELETE FROM task WHERE TaskId=%d", t.TaskID)
	_, err := db.Exec(statement)
	return err
}

func (t *task) createTask(db *sql.DB) error {
	statement := fmt.Sprintf("INSERT INTO task(TaskDefination, TaskMarked) VALUES('%s', %t)", t.TaskDefination, t.TaskMarked)
	_, err := db.Exec(statement)

	if err != nil {
		return err
	}

	err = db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&t.TaskID)

	if err != nil {
		return err
	}

	return nil
}

func getTasks(db *sql.DB) ([]task, error) {
	statement := fmt.Sprintf("SELECT TaskId, TaskDefination, TaskMarked FROM task ")
	rows, err := db.Query(statement)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tasks := []task{}

	for rows.Next() {
		var t task
		if err := rows.Scan(&t.TaskID, &t.TaskDefination, &t.TaskMarked); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
		//fmt.Println(tasks)
	}

	return tasks, nil
}
