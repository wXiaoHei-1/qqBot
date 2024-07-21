package clients

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"qqbot/utils"
	"sync"
)

var (
	GlobalConn *gorm.DB
	once       sync.Once
)

func NewDBClient(config *utils.Config) *gorm.DB {
	once.Do(func() {
		conn, err := gorm.Open(mysql.Open(config.Mysql), &gorm.Config{})
		if err != nil {
			fmt.Printf("err:%s\n", err)
			panic(err)
		}
		GlobalConn = conn
	})
	return GlobalConn
}
