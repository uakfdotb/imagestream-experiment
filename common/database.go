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

package common

import (
	_ "github.com/go-sql-driver/mysql"

	"database/sql"
)

type Database struct {
	db *sql.DB
}

func GetDatabaseString(username string, password string, host string, name string) string {
	s := username + ":" + password + "@"
	if host != "localhost" {
		s += host
	}
	s += "/" + name + "?charset=utf8&parseTime=true"
	return s
}

func NewDatabase(dbString string) *Database {
	dbw := new(Database)
	db, err := sql.Open("mysql", dbString)
	if err != nil {
		panic(err)
	}
	dbw.db = db
	return dbw
}

func (dbw *Database) Query(q string, args ...interface{}) Rows {
	rows, err := dbw.db.Query(q, args...)
	if err != nil {
		panic(err)
	}
	return Rows{rows}
}

func (dbw *Database) QueryRow(q string, args ...interface{}) Row {
	row := dbw.db.QueryRow(q, args...)
	return Row{row}
}

func (dbw *Database) Exec(q string, args ...interface{}) Result {
	result, err := dbw.db.Exec(q, args...)
	if err != nil {
		panic(err)
	}
	return Result{result}
}

type Rows struct {
	rows *sql.Rows
}

func (r Rows) Close() {
	err := r.rows.Close()
	if err != nil {
		panic(err)
	}
}

func (r Rows) Next() bool {
	return r.rows.Next()
}

func (r Rows) Scan(dest ...interface{}) {
	err := r.rows.Scan(dest...)
	if err != nil {
		panic(err)
	}
}

type Row struct {
	row *sql.Row
}

func (r Row) Scan(dest ...interface{}) {
	err := r.row.Scan(dest...)
	if err != nil {
		panic(err)
	}
}

type Result struct {
	result sql.Result
}

func (r Result) LastInsertId() int {
	id, err := r.result.LastInsertId()
	if err != nil {
		panic(err)
	}
	return int(id)
}

func (r Result) RowsAffected() int {
	count, err := r.result.RowsAffected()
	if err != nil {
		panic(err)
	}
	return int(count)
}
