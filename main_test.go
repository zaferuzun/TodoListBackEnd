// main_test.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a App

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize("sql7362589", "7WqYrReLNn", "sql7.freemysqlhosting.net", "sql7362589")

	ensureTableExists()
	BeforeRow()
	code := m.Run()

	clearTable()
	AfterRow()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM task")
	a.DB.Exec("ALTER TABLE task AUTO_INCREMENT = 1")
}

const tableCreationQuery = `
CREATE TABLE IF NOT EXISTS task
(
    TaskId INT AUTO_INCREMENT PRIMARY KEY,
    TaskDefination VARCHAR(50) NOT NULL,
    TaskMarked BOOL NOT NULL
)`

func BeforeRow() {
	a.DB.Exec("INSERT INTO task_yedek SELECT * FROM task")
}
func AfterRow() {
	a.DB.Exec("INSERT INTO task SELECT * FROM task_yedek")
	a.DB.Exec("DELETE FROM task_yedek ")
}
func TestEmptyTable(t *testing.T) {
	BeforeRow()
	clearTable()

	req, _ := http.NewRequest("GET", "/tasks", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
func TestGetNonExistentTask(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/task/5", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Task not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Task not found'. Got '%s'", m["error"])
	}
}
func TestCreateTask(t *testing.T) {
	clearTable()

	payload := []byte(`{"TaskDefination":"test task","TaskMarked":true}`)

	req, _ := http.NewRequest("POST", "/task", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["TaskDefination"] != "test task" {
		t.Errorf("Expected Task name to be 'test Task'. Got '%v'", m["TaskDefination"])
	}

	if m["TaskMarked"] != true {
		t.Errorf("Expected Task age to be 'true'. Got '%v'", m["TaskMarked"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["TaskId"] != 0.0 {
		t.Errorf("Expected task ID to be '1'. Got '%v'", m["TaskId"])
	}
}
func addTasks(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		statement := fmt.Sprintf("INSERT INTO task(TaskDefination, TaskMarked) VALUES('%s', %v)", ("task " + strconv.Itoa(i+1)), true)
		a.DB.Exec(statement)
	}
}
func TestGetTask(t *testing.T) {
	clearTable()
	addTasks(1)

	req, _ := http.NewRequest("GET", "/task/0", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateTask(t *testing.T) {
	clearTable()
	addTasks(1)

	req, _ := http.NewRequest("GET", "/task/0", nil)
	response := executeRequest(req)
	var originalTask map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalTask)

	payload := []byte(`{"TaskDefination":"test Task - updated name","TestMarked":false}`)

	req, _ = http.NewRequest("PUT", "/task/0", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["TaskId"] != originalTask["TaskId"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalTask["TaskId"], m["TaskId"])
	}

	if m["TaskDefination"] == originalTask["TaskDefination"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalTask["TaskDefination"], m["TaskDefination"], m["TaskDefination"])
	}

	if m["TaskMarked"] == originalTask["TaskMarked"] {
		t.Errorf("Expected the age to change from '%v' to '%v'. Got '%v'", originalTask["TaskMarked"], m["TaskMarked"], m["TaskMarked"])
	}
}

func TestDeleteTask(t *testing.T) {
	clearTable()
	addTasks(1)

	req, _ := http.NewRequest("GET", "/task/0", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/task/0", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/task/0", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
