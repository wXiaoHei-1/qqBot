package clients

import (
	"log"
	"qqbot/utils"
	"testing"
)

func TestCommon(t *testing.T) {
	t.Run("test DB connection", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic occurred when initializing database: %v", r)
				log.Println("Failed to initialize database connection. Please check your configuration.")
			}
		}()
		utils.NewConfig()
		NewDBClient(utils.ConfigInfo)
		log.Println("Database connection successful")
	})
}
