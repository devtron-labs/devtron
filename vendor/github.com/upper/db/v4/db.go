// Copyright (c) 2012-present The upper.io/db authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package db (or upper/db) provides an agnostic data access layer to work with
// different databases.
//
// Install upper/db:
//
//  go get github.com/upper/db
//
// Usage
//
//  package main
//
//  import (
//  	"log"
//
//  	"github.com/upper/db/v4/adapter/postgresql" // Imports the postgresql adapter.
//  )
//
//  var settings = postgresql.ConnectionURL{
//  	Database: `booktown`,
//  	Host:     `demo.upper.io`,
//  	User:     `demouser`,
//  	Password: `demop4ss`,
//  }
//
//  // Book represents a book.
//  type Book struct {
//  	ID        uint   `db:"id"`
//  	Title     string `db:"title"`
//  	AuthorID  uint   `db:"author_id"`
//  	SubjectID uint   `db:"subject_id"`
//  }
//
//  func main() {
//  	sess, err := postgresql.Open(settings)
//  	if err != nil {
//  		log.Fatal(err)
//  	}
//  	defer sess.Close()
//
//  	var books []Book
//  	if err := sess.Collection("books").Find().OrderBy("title").All(&books); err != nil {
//  		log.Fatal(err)
//  	}
//
//  	log.Println("Books:")
//  	for _, book := range books {
//  		log.Printf("%q (ID: %d)\n", book.Title, book.ID)
//  	}
//  }
package db
