package main

import (
	"database/sql"
	"fmt"
	"github.com/gosimple/slug"
	_ "github.com/lib/pq"
	"io"
	"net/http"
	"os"
	"strconv"
)

const (
	DB_USER   = "postgres"
	DB_PASS   = "123456"
	DB_NAME   = "slugger"
	FILE_NAME = "final.csv"
)

func check(e error) {
	if e != nil {
		panic(e)
	}

}

func delete(f string) {
	if _, err := os.Stat(f); err == nil {
		var err = os.Remove(f)
		check(err)
	}
}

func main() {

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASS, DB_NAME)

	db, err := sql.Open("postgres", dbinfo)
	check(err)

	defer db.Close()

	rows, err := db.Query("Select id,description,image_link,code,color from imageslug Where done=false order by id")
	check(err)

	delete(FILE_NAME)
	csv, err := os.Create(FILE_NAME)
	check(err)

	defer csv.Close()

	for rows.Next() {

		var id int
		var description string
		var slug_url string
		var image_link sql.NullString
		var code string
		var color string

		err = rows.Scan(&id, &description, &image_link, &code, &color)
		check(err)

		slug_url = slug.Make(description)

		var img_path = ""

		if image_link.Valid {
			response, err := http.Get(image_link.String)
			check(err)
			defer response.Body.Close()

			img_path = "images/" + strconv.Itoa(id) + "-" + slug.Make(code) + "-" + slug.Make(color) + ".jpg"
			image, err := os.Create(img_path)
			check(err)

			_, err = io.Copy(image, response.Body)
			check(err)
			image.Close()
		}

		n, err := csv.WriteString(slug_url + ";" + img_path + ";" + "\n")
		fmt.Printf("wrote %d bytes\n", n)
		check(err)
		csv.Sync()

		_, err = db.Exec("Update imageslug Set done=true Where id=$1", id)
		check(err)

	}

}
