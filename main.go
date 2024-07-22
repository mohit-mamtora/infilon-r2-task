package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	router := gin.Default()

	username := "root"
	password := "password"
	hostname := "localhost:32768"
	dbname := "cetec"

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbname)

	// Initialize the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Ping to db connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Register Routes
	router.GET("/person/:id/info", func(c *gin.Context) {
		getPerson(db, c)
	})

	router.POST("/person/create", func(c *gin.Context) {
		createPerson(db, c)
	})

	// Run the server
	router.Run(":8080")
}

func getPerson(db *sql.DB, c *gin.Context) {
	personId := c.Param("id")

	row := db.QueryRow(`
			SELECT p.name, a.city, ph.number, a.state, a.street1, a.street2, a.zip_code 
				FROM person p join phone ph on ph.person_id = p.id 
				join address_join aj on aj.person_id = p.id 
				join address a on a.id = aj.address_id 
				where p.id = ?`, personId)

	var personInfoResponse PersonInfoResponse

	if err := row.Scan(&personInfoResponse.Name,
		&personInfoResponse.City,
		&personInfoResponse.PhoneNumber,
		&personInfoResponse.State,
		&personInfoResponse.Street1,
		&personInfoResponse.Street2,
		&personInfoResponse.ZipCode,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, personInfoResponse)
}

func createPerson(db *sql.DB, c *gin.Context) {
	var personCreateRequest PersonCreateRequest
	if err := c.ShouldBindJSON(&personCreateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ADD IN PERSON TABLE
	result, err := db.Exec("INSERT INTO person (name) VALUES (?)", personCreateRequest.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	personId, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ADD IN PHONE TABLE
	_, err = db.Exec("INSERT INTO phone (number, person_id) VALUES (?, ?)", personCreateRequest.PhoneNumber, personId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ADD IN ADDRESS TABLE
	result, err = db.Exec("INSERT INTO address (city, state, street1, street2, zip_code) VALUES (?, ?, ?, ?, ?)",
		personCreateRequest.PhoneNumber,
		personCreateRequest.State,
		personCreateRequest.Street1,
		personCreateRequest.Street2,
		personCreateRequest.ZipCode)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	addressId, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ADD IN ADDRESS_JOIN TABLE
	_, err = db.Exec("INSERT INTO address_join (person_id, address_id) VALUES (?, ?)", personId, addressId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
