package test

import (
	"log"
	"qqbot/server"
	"qqbot/utils"
	"testing"
)

func TestServer(t *testing.T) {
	t.Run(
		"test idiom_solitaire", func(t *testing.T) {
			server.NewIdiomMap()
			idiomTestExamples := []string{"锦上添花", "153锦上天花", "圆润", "   "}
			for _, testExample := range idiomTestExamples {
				nextRecover, _ := server.ChengYvInterlocking(testExample)
				log.Println(nextRecover)
			}
		},
	)
	t.Run("test dialogue_gpt", func(t *testing.T) {
		utils.NewConfig()
		messageRecover := server.SendMessage("番茄炒蛋怎么做", utils.ConfigInfo.DashScopeAPIKey)
		log.Println(messageRecover)
	})
}
