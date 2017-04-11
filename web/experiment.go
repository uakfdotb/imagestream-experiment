/*
Copyright 2017 Favyen Bastani <fbastani@perennate.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package web

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

type Task struct {
	ID int `json:"id"`
	WorkerID int `json:"worker_id"`
	Type string `json:"type"`
	Images []TaskImage `json:"images"`
}

type TaskImage struct {
	ID int `json:"id"`
	ImageID int `json:"image_id"`
}

func startExperiment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}
	
	// create a new worker
	workerID := DB.Exec("INSERT INTO workers (experiment_id) VALUES (?)", ExperimentID).LastInsertId()
	
	// output tasks for this experiment that are not currently locked
	// if there are not enough available tasks, then return an error
	var conventionalTasks, rapidTasks int
	DB.QueryRow("SELECT conventional_tasks_per_worker, rapid_tasks_per_worker FROM experiments WHERE id = ?", ExperimentID).Scan(&conventionalTasks, &rapidTasks)
	getTasks := func(t string, l int) []Task {
		var tasks []Task
		rows := DB.Query("SELECT id, type FROM tasks WHERE experiment_id = ? AND type = ? AND (locked_until IS NULL OR locked_until < NOW()) ORDER BY RAND() LIMIT ?", ExperimentID, t, l)
		for rows.Next() {
			task := Task{WorkerID: workerID}
			rows.Scan(&task.ID, &task.Type)
			DB.Exec("UPDATE tasks SET locked_until = DATE_ADD(NOW(), INTERVAL 1 HOUR) WHERE id = ?", task.ID)
			imageRows := DB.Query("SELECT id, image_id FROM task_images WHERE task_id = ? ORDER BY id", task.ID)
			for imageRows.Next() {
				var image TaskImage
				imageRows.Scan(&image.ID, &image.ImageID)
				task.Images = append(task.Images, image)
			}
			tasks = append(tasks, task)
		}
		return tasks
	}
	var tasks []Task
	if rand.Intn(2) == 0 {
		tasks = append(getTasks("conventional", conventionalTasks), getTasks("rapid", rapidTasks)...)
	} else {
		tasks = append(getTasks("rapid", rapidTasks), getTasks("conventional", conventionalTasks)...)
	}
	if len(tasks) != conventionalTasks + rapidTasks {
		log.Printf("found only %d available tasks, but need %d conventional, %d rapid", conventionalTasks, rapidTasks)
		for _, task := range tasks {
			DB.Exec("UPDATE tasks SET locked_until = NOW() WHERE id = ?", task.ID)
		}
		w.WriteHeader(501)
		return
	}
	
	var taskIDs []int
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}
	log.Printf("handing out tasks to new worker %d from %s: %v", workerID, r.RemoteAddr, taskIDs)

	bytes, err := json.Marshal(tasks)
	if err != nil {
		panic(err)
	}
	w.Write(bytes)
}

type ConventionalLabel struct {
	TaskImageID int `json:"task_image_id"`
	IsInClass bool `json:"is_in_class"`
}

type RapidLog struct {
	DisplayJSON string `json:"display_json"`
	ClickJSON string `json:"click_json"`
	Mean float64 `json:"mean"`
	Sigma float64 `json:"sigma"`
}

type TaskResult struct {
	TaskID int `json:"task_id"`
	WorkerID int `json:"worker_id"`
	Duration int `json:"duration"`
	ConventionalLabels []ConventionalLabel `json:"conventional_labels"`
	RapidLogs []RapidLog `json:"rapid_logs"`
}

func endExperiment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}
	
	defer r.Body.Close()
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	var result TaskResult
	if err := json.Unmarshal(bytes, &result); err != nil {
		w.WriteHeader(400)
		log.Printf("error on postExperiment: %v", err)
		return
	}
	DB.Exec("UPDATE tasks SET completed = 1 WHERE id = ?", result.TaskID)
	DB.Exec("INSERT INTO task_results (worker_id, task_id, duration_ms) VALUES (?, ?, ?)", result.WorkerID, result.TaskID, result.Duration)
	for _, label := range result.ConventionalLabels {
		DB.Exec("INSERT INTO conventional_labels (worker_id, task_id, task_image_id, is_in_class) VALUES (?, ?, ?, ?)", result.WorkerID, result.TaskID, label.TaskImageID, label.IsInClass)
	}
	for _, log := range result.RapidLogs {
		DB.Exec("INSERT INTO rapid_logs (worker_id, task_id, display_json, click_json, mean, sigma) VALUES (?, ?, ?, ?, ?, ?)", result.WorkerID, result.TaskID, log.DisplayJSON, log.ClickJSON, log.Mean, log.Sigma)
	}
}
