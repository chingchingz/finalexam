package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/chingchingz/finalexam/todo"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB
var customers = []todo.Customers{}

func createTodosHandler(c *gin.Context) {
	var item todo.Customers
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(item)

	row := db.QueryRow("INSERT INTO customers (name, email, status) values ($1, $2, $3) RETURNING id", item.Name, item.Email, item.Status)
	var id int
	err := row.Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Status": "Error"})
		return
	}
	item.ID = id
	c.JSON(http.StatusCreated, item)
}

func getTodosByIDHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers WHERE id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	row := stmt.QueryRow(id)
	t := todo.Customers{}
	err = row.Scan(&t.ID, &t.Name, &t.Email, &t.Status)
	if err != nil {
		log.Fatal("Cannot Scan row into variables", err)
	}
	c.JSON(http.StatusOK, t)
	return
}

func getTodosHandler(c *gin.Context) {
	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers")
	if err != nil {
		log.Fatal("Cannot prepare query all customers statement", err)
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Cannot query all todos", err)
	}
	for rows.Next() {
		t := todo.Customers{}
		err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Status)

		if err != nil {
			log.Fatal("Cannot scan row into variable", err)
		}
		customers = append(customers, t)
	}
	c.JSON(http.StatusOK, customers)
}

func updateTodosHandler(c *gin.Context) {
	item := todo.Customers{}
	err := c.ShouldBindJSON(&item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stmt, err := db.Prepare("UPDATE customers SET name=$2, email=$3, status=$4 WHERE id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))

	if _, err := stmt.Exec(id, item.Name, item.Email, item.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	item.ID = id
	c.JSON(http.StatusOK, item)
}

func deleteTodosHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	stmt, err := db.Prepare("DELETE FROM customers WHERE id = $1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if _, err := stmt.Exec(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}

func loginMiddleware(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	if authKey != "token2019" {
		c.JSON(http.StatusUnauthorized, "Unauthorization")
		c.Abort()
		return
	}
	c.Next()
	log.Println("ending middleware")
}

func setUp() *gin.Engine {
	r := gin.Default()
	r.Use(loginMiddleware)
	r.GET("/customers", getTodosHandler)
	r.GET("/customers/:id", getTodosByIDHandler)
	r.POST("/customers", createTodosHandler)
	r.PUT("/customers/:id", updateTodosHandler)
	r.DELETE("/customers/:id", deleteTodosHandler)
	return r
}

func main() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Connect to Database error", err)
	}
	defer db.Close()

	createTb := `
	CREATE TABLE IF NOT EXISTS customers(
	 id SERIAL PRIMARY KEY,
	 name TEXT,
	 email TEXT,
	 status TEXT
	);
	`
	_, err = db.Exec(createTb)

	if err != nil {
		log.Fatal("Cannot create table", err)
	}

	fmt.Println("Create Table success")
	r := setUp()
	r.Run(":2019")
}
