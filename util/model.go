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

package util

import (
	"../common"

	"encoding/json"
	"fmt"
	"math"
	"time"
)

type Result struct {
	Precision float64
	Recall float64
	Time time.Duration
}

func getTime(db *common.Database, t string, experimentID string) time.Duration {
	var durationSum int
	row := db.QueryRow("SELECT SUM(task_results.duration_ms) FROM task_results, tasks WHERE task_results.task_id = tasks.id AND tasks.experiment_id = ? AND tasks.type = ?", experimentID, t)
	row.Scan(&durationSum)
	return time.Duration(durationSum) * time.Millisecond
}

func getPrecisionRecall(db *common.Database, predictions map[int]bool) (float64, float64) {
	var numMatch, numActualTrue, numPredictTrue int
	for imageID, prediction := range predictions {
		var actual bool
		db.QueryRow("SELECT is_in_class FROM images WHERE id = ?", imageID).Scan(&actual)
		if actual {
			numActualTrue++
		}
		if prediction {
			numPredictTrue++
		}
		if actual && prediction {
			numMatch++
		}
	}
	return float64(numMatch) / float64(numPredictTrue), float64(numMatch) / float64(numActualTrue)
}

func RunConventionalModel(db *common.Database, experimentID string) Result {
	// apply majority voting on conventional_labels table
	type ImagePrediction struct {
		ImageID int
		Prediction bool
	}
	imageVotes := make(map[ImagePrediction]int)
	imageIDs := make(map[int]bool)
	rows := db.Query("SELECT task_images.image_id, conventional_labels.is_in_class FROM task_images, conventional_labels, tasks WHERE conventional_labels.task_image_id = task_images.id AND task_images.task_id = tasks.id AND tasks.experiment_id = ? AND tasks.type = 'conventional'", experimentID)
	for rows.Next() {
		var imageID int
		var prediction bool
		rows.Scan(&imageID, &prediction)
		imageVotes[ImagePrediction{imageID, prediction}]++
		imageIDs[imageID] = true
	}

	predictions := make(map[int]bool)
	for imageID, _ := range imageIDs {
		predictions[imageID] = imageVotes[ImagePrediction{imageID, true}] > imageVotes[ImagePrediction{imageID, false}]
	}

	var result Result
	result.Precision, result.Recall = getPrecisionRecall(db, predictions)
	result.Time = getTime(db, "conventional", experimentID)
	return result
}

func RunRapidModel(db *common.Database, threshold float64, experimentID string) Result {
	// compute probability sum for each image
	// the set of keypresses for each worker is modeled as a sum of Gaussian distributions
	// (the paper actually says they use product of Gaussians, but then their figures show them using sum, and sum makes more sense)
	imageProbs := make(map[int]float64)

	type DisplayEntry struct {
		ImageID int `json:"image_id"`
		Time int `json:"time"`
	}

	type ClickEntry struct {
		Time int `json:"time"`
	}

	rows := db.Query("SELECT display_json, click_json, mean, sigma FROM rapid_logs, tasks WHERE rapid_logs.task_id = tasks.id AND tasks.experiment_id = ?", experimentID)
	for rows.Next() {
		var displayJSON, clickJSON string
		var mean, sigma float64
		rows.Scan(&displayJSON, &clickJSON, &mean, &sigma)

		normalPDF := func(timeDif float64) float64 {
			return 1 / math.Sqrt(2 * math.Pi * sigma * sigma) * math.Exp(-(timeDif - mean) * (timeDif - mean) / 2 / sigma / sigma)
		}

		var displayData []DisplayEntry
		var clickData []ClickEntry
		if err := json.Unmarshal([]byte(displayJSON), &displayData); err != nil {
			panic(err)
		}
		if err := json.Unmarshal([]byte(clickJSON), &clickData); err != nil {
			panic(err)
		}

		for _, click := range clickData {
			for _, display := range displayData {
				fmt.Printf("add %v for %d\n", normalPDF(float64(click.Time - display.Time)), display.ImageID)
				imageProbs[display.ImageID] += normalPDF(float64(click.Time - display.Time))
			}
		}
	}

	// apply threshold to obtain predictions
	predictions := make(map[int]bool)
	for imageID, score := range imageProbs {
		predictions[imageID] = score >= threshold
	}

	var result Result
	result.Precision, result.Recall = getPrecisionRecall(db, predictions)
	result.Time = getTime(db, "rapid", experimentID)
	return result
}
