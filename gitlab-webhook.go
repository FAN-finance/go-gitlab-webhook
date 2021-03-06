package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xen0n/go-workwx"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)
type Webhook struct {
	After   string `json:"after"`
	Before  string `json:"before"`
	Commits []struct {
		Added  []string `json:"added"`
		Author struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"author"`
		ID        string        `json:"id"`
		Message   string        `json:"message"`
		Modified  []string      `json:"modified"`
		Removed   []interface{} `json:"removed"`
		Timestamp string        `json:"timestamp"`
		URL       string        `json:"url"`
	} `json:"commits"`
	ObjectKind string `json:"object_kind"`
	ProjectID  int    `json:"project_id"`
	Ref        string `json:"ref"`
	Repository struct {
		Description     string `json:"description"`
		GitHTTPURL      string `json:"git_http_url"`
		GitSSHURL       string `json:"git_ssh_url"`
		Homepage        string `json:"homepage"`
		Name            string `json:"name"`
		URL             string `json:"url"`
		VisibilityLevel int    `json:"visibility_level"`
	} `json:"repository"`
	TotalCommitsCount int    `json:"total_commits_count"`
	UserEmail         string `json:"user_email"`
	UserID            int    `json:"user_id"`
	UserName          string `json:"user_name"`
}
const (
	OK_PUSH="push"
	OK_TAGPUSH="tag_push"
)


//ConfigRepository represents a repository from the config file
type ConfigRepository struct {
	Name     string
	//Branch string
	//Commands []string
	ConfigBranchs []map[string][]string  `json:"config_branchs"`;
}

//Config represents the config file
type Config struct {
	Logfile      string
	Address      string
	Port         int64
	Repositories []ConfigRepository
}

func PanicIf(err error, what ...string) {
	if err != nil {
		if len(what) == 0 {
			panic(err)
		}

		panic(errors.New(err.Error() + what[0]))
	}
}

var config Config
var configFile ="config.json"
func main() {

	//sigc := make(chan os.Signal, 1)
	//signal.Notify(sigc, syscall.SIGHUP)
	//
	//go func() {
	//	<-sigc
	//	var err error
	//	config, err = loadConfig(configFile)
	//	if err != nil {
	//		log.Fatalf("Failed to read config: %s", err)
	//	}
	//	log.Println("config reloaded")
	//}()

	//if we have a "real" argument we take this as conf path to the config file

	args := os.Args
	if len(args) > 1 {
		configFile = args[1]
	} else {
		configFile = "config.json"
	}

	//load config
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to read config: %s", err)
	}

	//open log file
	writer, err := os.OpenFile(config.Logfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %s", err)
		os.Exit(1)
	}

	//close logfile on exit
	defer func() {
		writer.Close()
	}()

	//setting logging output
	log.SetOutput(writer)

	//setting handler
	http.HandleFunc("/", hookHandler)

	address := config.Address + ":" + strconv.FormatInt(config.Port, 10)

	log.Println("Listening on " + address)

	//starting server
	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.Println(err)
	}
}

func loadConfig(configFile string) (Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	count := 0

	count, err = file.Read(buffer)
	if err != nil {
		return Config{}, err
	}

	err = json.Unmarshal(buffer[:count], &config)
	if err != nil {
		log.Println("loadConfig err",err)
		return Config{}, err
	}
	return config, nil
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	var hook Webhook

	//read request body
	var data, err= ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request: %s", err)
		w.WriteHeader(500)
		return
	}

	//unmarshal request body
	err = json.Unmarshal(data, &hook)
	if err != nil {
		log.Printf("Failed to parse request: %s", err)
		w.WriteHeader(500)
		return
	}
	log.Println("get branch: ", hook.Repository.Name+"/"+hook.Ref)
	SendHookMsg(&hook) //send msg for every Repository and branch

	config, _ = loadConfig(configFile)
	//find matching config for repository name
	for _, repo := range config.Repositories {
		if repo.Name != hook.Repository.Name {
			continue
		}
		//execute commands for repository
		for _, configBranch := range repo.ConfigBranchs {
			for branchName, cmds := range configBranch {
				if "refs/heads/"+branchName == hook.Ref {
					//SendHookMsg(&hook) //only send message for  matched brandh

					var branchId= repo.Name + "/" + branchName
					log.Println("process branch: ", branchId)
					for _, cmd := range cmds {

						var exeId = branchId + ":" + cmd
						log.Println("trigger branch: ", exeId)
						var command= exec.Command(cmd)
						out, err1 := command.Output()
						err=err1
						log.Println("Output: " + string(out))
						if err != nil {
							log.Printf("Failed to execute command: %s", err)
							SendMsg(fmt.Sprintf("%s ??????: \n%s",exeId, string(out)))
							break;
						}else{
							SendMsg(fmt.Sprintf("%s ??????: \n%s",exeId, string(out)))
						}
					}
					if err == nil {
						log.Println("finish process branch: " + branchId)

					}
				}
			}
		}
	}
}


func SendHookMsg(hook *Webhook) {
	commits:=""
	for idx,commit:=range hook.Commits{
		ctime,_:=time.Parse("2006-01-02T15:04:05Z",commit.Timestamp)
		ctime=ctime.Local()
		ctimeStr:=ctime.Format("01-02 15:04")
		commits+= fmt.Sprintf("%d) %s %s %s \n",idx,ctimeStr,commit.Author.Name, commit.Message)
	}
	branchName:=strings.TrimPrefix(hook.Ref,"refs/heads/")
	msg:=fmt.Sprintf("%s/%s git ????????????,???????????????????????????\n%s",hook.Repository.Name,branchName,commits)
	SendMsg(msg)
	return
}

func SendMsg(msg string) {

	log.Println("send msg:\n",msg)
	to5 := workwx.Recipient{
		UserIDs: []string{"@all"},
	}

	msg=msg+"\n date:"+time.Now().Format("2006-01-02 15:04:05")
	err:= app.SendTextMessage(&to5, msg, false)
	log.Println(err)
	return
}

var corpID = "ww23e68632206c98e7"
var corpSecret = "G2gcJ8ui5xjCD0QTiKPhiW3jXXNxNWxfh29VWC1NSJY"
var agentID = int64(1000002)
var app *workwx.WorkwxApp
func init(){
	client := workwx.New(corpID)
	app = client.WithApp(corpSecret, agentID)
	// preferably do this at app initialization
	app.SpawnAccessTokenRefresher()
}
