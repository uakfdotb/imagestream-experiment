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

package main

import (
	"./common"
	"./util"
	"./web"

	"flag"
	"fmt"
)

var (
	mode = flag.String("mode", "", "one of init, serve, eval")
	experimentID = flag.String("expid", "", "experiment ID")
	dbUsername = flag.String("dbuser", "root", "database username")
	dbPassword = flag.String("dbpass", "", "database password")
	dbHost = flag.String("dbhost", "localhost", "database host")
	dbName = flag.String("dbname", "imagestream", "database name")
	threshold = flag.Float64("threshold", 1.0, "rapid model threshold")
)

var db *common.Database

func main() {
	flag.Parse()
	db = common.NewDatabase(common.GetDatabaseString(*dbUsername, *dbPassword, *dbHost, *dbName))
	if *mode == "init" {
		if err := util.InitializeExperiment(db, *experimentID); err != nil {
			panic(err)
		}
	} else if *mode == "serve" {
		web.DB = db
		web.ExperimentID = *experimentID
		web.Serve()
	} else if *mode == "eval" {
		conventionalResult := util.RunConventionalModel(db, *experimentID)
		rapidResult := util.RunRapidModel(db, *threshold, *experimentID)
		fmt.Printf("conventional: precision=%v, recall=%v, time=%v\n", conventionalResult.Precision, conventionalResult.Recall, conventionalResult.Time)
		fmt.Printf("rapid: precision=%v, recall=%v, time=%v\n", rapidResult.Precision, rapidResult.Recall, rapidResult.Time)
	} else {
		fmt.Println("invalid mode specified")
		flag.PrintDefaults()
	}
}
