package main

import (
	"flag"
	"github.com/xen0n/go-workwx"
	"log"
)

func main() {
	task := ""
	flag.StringVar(&task, "t", "", "任务名：默认默示用户列表")
	flag.StringVar(&corpID, "i", "ww23e68632206c98e7", "corpID")
	flag.StringVar(&corpSecret, "s", "G2gcJ8ui5xjCD0QTiKPhiW3jXXNxNWxfh29VWC1NSJY", "corpSecret")
	flag.Int64Var(&agentID, "a", 1000002, "agentID")
	flag.Parse()

	log.Printf("corpID %s,corpSecret %s,agentID %d",corpID,corpSecret,agentID)

	to:=workwx.Recipient{UserIDs:[]string{"@all"}} //"WuXiaoFei","HaiKuoTianKong"
	err:=app.SendTextMessage(&to,"test msg12",false)
	log.Println(err)

}

var corpID = "ww23e68632206c98e7"
var corpSecret = "G2gcJ8ui5xjCD0QTiKPhiW3jXXNxNWxfh29VWC1NSJY"
var  agentID = int64(1000002)
var app *workwx.WorkwxApp
func init(){
	client := workwx.New(corpID)
	app = client.WithApp(corpSecret, agentID)
	// preferably do this at app initialization
	app.SpawnAccessTokenRefresher()

}