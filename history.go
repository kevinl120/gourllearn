package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	_ "github.com/mattn/go-sqlite3"
)

type history struct {
	Url   string
	Score uint8
}

var c *mgo.Collection
var session *mgo.Session

func copyFile(src string, dst string) {
	// Read all content of src to data
	data, err := ioutil.ReadFile(src)
	checkErr(err)
	// Write data to dst
	err = ioutil.WriteFile(dst, data, 0644)
	checkErr(err)
}

// remove un-needed parts from URL
func trimURL(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "//")
	url = strings.TrimPrefix(url, "www.")
	url = strings.TrimSuffix(url, ":443")
	return url
}

func readChromeHistory() {
	usr, err := user.Current()
	checkErr(err)

	historyDb := usr.HomeDir + "/Library/Application Support/Google/Chrome/Default/history"
	tmpHistory := usr.HomeDir + "/history"
	copyFile(historyDb, tmpHistory)

	// Open Chrome history database
	db, err := sql.Open("sqlite3", tmpHistory)
	checkErr(err)
	rows, err := db.Query("SELECT url FROM urls")
	checkErr(err)

	// Open mongoDB to write prediction result
	session, err = mgo.Dial("localhost")
	checkErr(err)
	c = session.DB("test_database").C("test")
	c.RemoveAll(nil)

	// Make url unique index
	index := mgo.Index{
		Key:        []string{"url"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)
	checkErr(err)

	var url string
	// Read each browser history entry and write prediction result to mongoDB
	for rows.Next() {
		err = rows.Scan(&url)
		checkErr(err)
		url = trimURL(url)

		score := predict(url)
		err = c.Insert(&history{Url: url, Score: score})
		if !mgo.IsDup(err) {
			checkErr(err)
		}
	}

	rows.Close()
	db.Close()
	os.Remove(tmpHistory)
}

// Convert int score to string "good"/"bad" for print
func scoreToString(score uint8) string {
	if score == 1 {
		return "good"
	}
	return "bad"
}

// Query mongoDB to determine whether URL is good or bad
func isBadURL(url string) bool {
	url = trimURL(url)
	result := history{}
	err := c.Find(bson.M{"url": url}).One(&result)
	if err != nil && err.Error() == "not found" {
		// if we have not seen this URL before, predict result and save to
		// mongoDB
		score := predict(url)
		err = c.Insert(&history{Url: url, Score: score})
		fmt.Println("---", scoreToString(score), " URL---", url)
		return (score != 1)
	}
	checkErr(err)

	fmt.Println("---", scoreToString(result.Score), "URL---", result.Url)
	return (result.Score != 1)
}
