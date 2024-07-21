package test

import (
	"context"
	"log"
	"qqbot/common/service"
	"qqbot/common/types"
	"qqbot/utils"
	"testing"
	"time"
)

func TestGetReq(t *testing.T) {
	t.Run(
		"get websocket accessIp by gateway", func(t *testing.T) {
			ctx := context.Background()
			utils.NewConfig()
			httpClient := service.NewClient(utils.ConfigInfo.AppID, utils.ConfigInfo.Token, 3*time.Second)
			// 通过http获取webSocket连接地址
			ws, err := httpClient.GetWSS(ctx)
			if err != nil {
				log.Fatalln("websocket err :", err)
				return
			}
			log.Println("webSocket连接地址为:", ws)
		},
	)
	t.Run("post method", func(t *testing.T) {
		ctx := context.Background()
		utils.NewConfig()
		httpClient := service.NewClient(utils.ConfigInfo.AppID, utils.ConfigInfo.Token, 3*time.Second)
		_, err := httpClient.PostMessage(ctx, "656434002", &types.MessageToCreate{MsgID: "08b496a5b4c1e483b6840110d2c681b90238c2024889b9f1b406", Content: "测试成功。"})
		if err != nil {
			log.Fatalln("PostMessage err: ", err)
			return
		}
		log.Println("发送成功")
	})
}
