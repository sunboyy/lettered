package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB contains a set of database accessing functionality. It wraps the gorm
// backend to make this package a central location for accessing the database.
// Database accessing functionality should not be placed outside this package.
type DB struct {
	backend *gorm.DB
}

// New creates a new instance of DB. It initializes database backend according
// to the configuration, migrates the database schema and wraps within a DB
// struct.
func New(config Config) (*DB, error) {
	backend, err := gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	backend.AutoMigrate(&FriendRequest{})
	backend.AutoMigrate(&Friend{})

	return &DB{backend: backend}, nil
}
