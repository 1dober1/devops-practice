package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `form:"name"`
	Email string `form:"email"`
}

var db *gorm.DB

func main() {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	// Формируем DSN для подключения к SQL Server
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// Подключаемся к SQL Server
	var err error
	db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to the database")
	}

	// Получаем экземпляр базы данных
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get db instance")
	}

	// Проверяем существование базы данных
	if err := createDatabaseIfNotExists(sqlDB, "testdb"); err != nil {
		panic(fmt.Sprintf("failed to create database: %v", err))
	}

	// Переподключаемся к новой базе данных
	dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=testdb",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to the database 'testdb': %v", err))
	}

	db.AutoMigrate(&User{})

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.POST("/submit", func(c *gin.Context) {
		var user User
		if err := c.ShouldBind(&user); err == nil {
			db.Create(&user)
			c.JSON(http.StatusOK, gin.H{"status": "User created!"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	})

	router.GET("/users", func(c *gin.Context) {
		var users []User
		result := db.Find(&users)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	})

	router.LoadHTMLFiles("index.html")
	router.Run(":8080")
}

// createDatabaseIfNotExists проверяет существование базы данных и создает ее, если она не существует
func createDatabaseIfNotExists(db *sql.DB, dbName string) error { // Изменили аргумент на *sql.DB
	var dbExists int
	query := fmt.Sprintf("SELECT COUNT(*) FROM sys.databases WHERE name = '%s'", dbName)
	err := db.QueryRow(query).Scan(&dbExists)
	if err != nil {
		return err
	}

	if dbExists == 0 {
		_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		return err
	}

	return nil
}
