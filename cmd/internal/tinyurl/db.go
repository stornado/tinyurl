package tinyurl

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type Tinyurl struct {
	ID        int64
	Short     string
	Origin    string
	CreatedAt time.Time
	ExpiredAt time.Time
}

type TinyurlStats struct {
	ID    int64
	Short string
	PV    int64
	UV    int64
}

func init() {

}

func (tiny *Tinyurl) Add() (*Tinyurl, error) {
	host := viper.GetString("db.host")
	port := viper.GetInt("db.port")
	username := viper.GetString("db.username")
	password := viper.GetString("db.password")
	dbname := viper.GetString("db.dbname")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=true", username, password, host, port, dbname))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO `tinyurl`.`tinyurl` (`short`,`origin`,`created_At`,`expired_at`) VALUES (?,?,?,?);") // ? = placeholder
	if err != nil {
		return nil, err
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates
	tiny.CreatedAt = time.Now()
	tiny.ExpiredAt = time.Now().Add(24 * time.Hour)

	// Insert square numbers for 0-24 in the database
	result, err := stmtIns.Exec(tiny.Short, tiny.Origin, tiny.CreatedAt, tiny.ExpiredAt) // Insert tuples (i, i^2)
	if err != nil {
		return nil, err
	}

	shortID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	tiny.ID = shortID

	return tiny, nil
}

func PersistGet(short string) (*Tinyurl, error) {
	host := viper.GetString("db.host")
	port := viper.GetInt("db.port")
	username := viper.GetString("db.username")
	password := viper.GetString("db.password")
	dbname := viper.GetString("db.dbname")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=true", username, password, host, port, dbname))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Prepare statement for reading data
	stmtOut, err := db.Prepare("SELECT id,short,origin,created_at,expired_at FROM tinyurl WHERE short = ?")
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()

	tiny := &Tinyurl{Short: short}

	row := stmtOut.QueryRow(tiny.Short)

	err = row.Scan(&tiny.ID, &tiny.Short, &tiny.Origin, &tiny.CreatedAt, &tiny.ExpiredAt)
	if err != nil {
		return nil, err
	}

	return tiny, nil
}

func ListAll() ([]Tinyurl, error) {
	host := viper.GetString("db.host")
	port := viper.GetInt("db.port")
	username := viper.GetString("db.username")
	password := viper.GetString("db.password")
	dbname := viper.GetString("db.dbname")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=true", username, password, host, port, dbname))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Prepare statement for reading data
	stmtOut, err := db.Prepare("SELECT id,short,origin,created_at,expired_at FROM tinyurl WHERE expired_at < ?")
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(time.Now())
	if err != nil {
		return nil, err
	}
	var shorts []Tinyurl
	for rows.Next() {
		var tiny Tinyurl
		err = rows.Scan(&tiny.ID, &tiny.Short, &tiny.Origin, &tiny.CreatedAt, &tiny.ExpiredAt)
		if err != nil {
			return nil, err
		}
		shorts = append(shorts, tiny)
	}

	return shorts, nil
}
