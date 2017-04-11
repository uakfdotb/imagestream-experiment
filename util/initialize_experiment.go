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

	"fmt"
	"math/rand"
)

func InitializeExperiment(db *common.Database, experimentID string) error {
	var totalImages, conventionalRedundancy, rapidRedundancy, imagesPerTask int
	row := db.QueryRow("SELECT total_images, conventional_redundancy, rapid_redundancy, images_per_task FROM experiments WHERE id = ?", experimentID)
	row.Scan(&totalImages, &conventionalRedundancy, &rapidRedundancy, &imagesPerTask)

	if totalImages % imagesPerTask != 0 {
		return fmt.Errorf("totalImages must be a multiple of imagesPerTask, but got %d and %d", totalImages, imagesPerTask)
	}

	// pick images for this experiment
	var experimentImages []int
	rows := db.Query("SELECT id FROM images ORDER BY RAND() LIMIT ?", totalImages)
	for rows.Next() {
		var imageID int
		rows.Scan(&imageID)
		experimentImages = append(experimentImages, imageID)
	}

	// create labeling tasks
	// we use a different permutation of the images for each redundancy
	createTasks := func(t string, redundancy int) {
		for n := 0; n < redundancy; n++ {
			perm := rand.Perm(len(experimentImages))
			for taskIndex := 0; taskIndex < totalImages / imagesPerTask; taskIndex++ {
				taskID := db.Exec("INSERT INTO tasks (experiment_id, type) VALUES (?, ?)", experimentID, t).LastInsertId()
				for i := 0; i < imagesPerTask; i++ {
					imageID := experimentImages[perm[taskIndex * imagesPerTask + i]]
					db.Exec("INSERT INTO task_images (task_id, image_id) VALUES (?, ?)", taskID, imageID)
				}
			}
		}
	}
	createTasks("conventional", conventionalRedundancy)
	createTasks("rapid", rapidRedundancy)

	return nil
}
