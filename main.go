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
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file")
	}

	// Подключаемся к SQL Server через master для создания базы данных
	masterDSN := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=master",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	sqlDB, err := sql.Open("sqlserver", masterDSN)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to master database: %v", err))
	}
	defer sqlDB.Close()

	// Проверяем или создаём базу данных
	dbName := os.Getenv("DB_NAME")
	if err := createDatabaseIfNotExists(sqlDB, dbName); err != nil {
		panic(fmt.Sprintf("failed to create database '%s': %v", dbName, err))
	}

	// Подключаемся к созданной базе данных
	appDSN := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		dbName,
	)

	db, err = gorm.Open(sqlserver.Open(appDSN), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to the database '%s': %v", dbName, err))
	}

	// Миграции
	db.AutoMigrate(&User{})

	// Запуск сервера
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

// createDatabaseIfNotExists проверяет существование базы данных и создает её, если она не существует
func createDatabaseIfNotExists(db *sql.DB, dbName string) error {
	var dbExists int
	query := fmt.Sprintf("SELECT COUNT(*) FROM sys.databases WHERE name = '%s'", dbName)
	err := db.QueryRow(query).Scan(&dbExists)
	if err != nil {
		return fmt.Errorf("error checking database existence: %w", err)
	}

	if dbExists == 0 {
		_, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("error creating database: %w", err)
		}
		fmt.Printf("Database '%s' created successfully.\n", dbName)
	} else {
		fmt.Printf("Database '%s' already exists.\n", dbName)
	}

	return nil
}
