package db

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var dbModels = []interface{}{&Customer{}, &ChatHistory{}}

type DBService struct {
	Host     string
	User     string
	Password string
	DBname   string
	Port     *string
	SSLMode  bool
	TimeZone string
}

func NewDBService(host string, user string, password string, dbname string, port *string, sslmode bool, timezone string) DBService {
	return DBService{
		Host:     host,
		User:     user,
		Password: password,
		DBname:   dbname,
		Port:     port,
		SSLMode:  sslmode,
		TimeZone: timezone,
	}
}

func (s *DBService) Connect() error {
	sslModeStr := "disable"
	if s.SSLMode {
		sslModeStr = "require"
	}

	strBuilder := strings.Builder{}
	strBuilder.WriteString(fmt.Sprintf("host=%s ", s.Host))
	strBuilder.WriteString(fmt.Sprintf("user=%s ", s.User))
	strBuilder.WriteString(fmt.Sprintf("password=%s ", s.Password))
	strBuilder.WriteString(fmt.Sprintf("dbname=%s ", s.DBname))
	strBuilder.WriteString(fmt.Sprintf("sslmode=%s ", sslModeStr))
	strBuilder.WriteString(fmt.Sprintf("TimeZone=%s ", s.TimeZone))

	if s.Port != nil {
		strBuilder.WriteString(fmt.Sprintf("port=%s", *s.Port))
	}

	dsn := strBuilder.String()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	DB = db
	return nil
}

func (s *DBService) Migrate() error {
	return DB.AutoMigrate(dbModels...)
}

func Insert[T any](value T, modifiedDb *gorm.DB) error {
	ctx := context.Background()

	dbInstance := DB
	if modifiedDb != nil {
		dbInstance = modifiedDb
	}

	return dbInstance.WithContext(ctx).Create(&value).Error
}
