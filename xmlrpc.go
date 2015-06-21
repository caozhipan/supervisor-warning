package main

import (
	"encoding/json"
	"fmt"
	"github.com/kolo/xmlrpc"
	"gopkg.in/gomail.v1"
	"io/ioutil"
	"strconv"
	"time"
)

type ServiceRes struct {
	Group          string //Name of the process’ group
	Pid            int    //UNIX process ID (PID) of the process, or 0 if the process is not running.
	Exitstatus     int    //Exit status (errorlevel) of process, or 0 if the process is still running.
	Spawnerr       string //Description of error that occurred during spawn, or empty string if none.
	Name           string //Name of the process
	Stderr_logfile string //Absolute path and filename to the STDOUT logfile
	Now            int64  //UNIX timestamp of the current time, which can be used to calculate process up-time.
	Start          int64  //UNIX timestamp of when the process was started
	State          int    //State code: STOPPED(0),STARTING (10),RUNNING (20),BACKOFF (30),STOPPING (40),EXITED (100),FATAL (200),UNKNOWN (1000)
	Stdout_logfile string //Absolute path and filename to the STDOUT logfile
	Stop           int64  //UNIX timestamp of when the process last ended, or 0 if the process has never been stopped.
	Statename      string //tring description of state,
}
type Config struct {
	Servers    []Server
	Admin      []User
	MailServer MailServer
}
type Server struct {
	Host     string
	Port     int
	Services []Service
}
type Service struct {
	Name         string
	Alias        string
	Subscription []User
}
type User struct {
	Name  string
	Email string
}
type MailServer struct {
	User     string
	Password string
	Smtp     string
	Port     int
}

func main() {
	var content string = ""
	var title string = ""
	for {
		config := getConfig("./config.json")
		for _, sv := range config.Servers {
			client, _ := xmlrpc.NewClient("http://"+sv.Host+":"+strconv.Itoa(sv.Port)+"/RPC2", nil)
			var res1, res2 interface{}
			var res3 []ServiceRes
			client.Call("supervisor.getState", nil, &res1)
			result1 := res1.(map[string]interface{})
			if result1["statecode"].(int64) == 1 {
				client.Call("supervisor.getAllProcessInfo", nil, &res2)
				tmp, _ := json.Marshal(res2)
				err := json.Unmarshal(tmp, &res3)
				if err == nil {
					if len(res3) > 0 {
						for _, s := range res3 {
							if s.State != 20 {
								//向订阅该服务的用户发邮件
								content = ""
								title = ""
								for _, us := range sv.Services {
									if us.Name == s.Name {
										title = "服务异常:" + us.Alias + "(" + us.Name + ")"
										content += "服务异常:" + us.Alias + "(" + us.Name + ")<br />"
										content += "运行状态:" + s.Statename + "<br />"
										content += "错误信息:" + s.Spawnerr + "<br />"
										content += "address:" + sv.Host + ":" + strconv.Itoa(sv.Port) + "<br />"
										content += "stdout_logfilg:" + s.Stdout_logfile + "<br />"
										content += "stderr_logfile:" + s.Stderr_logfile + "<br />"
										sendMail(config.MailServer, us.Subscription, title, content)
									}
								}
							}
						}
					}
				} else {
					panic(err)
				}
			} else {
				//向管理员发邮件
				sendMail(config.MailServer, config.Admin, "supervisor 异常", sv.Host+":"+strconv.Itoa(sv.Port)+"   statename:"+result1["statename"].(string))
			}

		}
		time.Sleep(5 * time.Minute)
	}
}
func getConfig(filename string) (config Config) {
	jsonStr, err1 := ioutil.ReadFile(filename)
	if err1 != nil {
		panic(err1)
	}
	err2 := json.Unmarshal(jsonStr, &config)
	if err2 != nil {
		panic(err2)
	}
	return
}

//发邮件
func sendMail(ms MailServer, us []User, title, content string) {
	fmt.Println("发邮件")
	msg := gomail.NewMessage()
	msg.SetHeader("From", ms.User)
	tmp := make([]string, 0)
	for _, v := range us {
		tmp = append(tmp, v.Email)
	}
	msg.SetHeader("To", tmp...)
	//msg.SetAddressHeader("Cc", "dan@example.com")
	msg.SetHeader("Subject", title)
	msg.SetBody("text/html", content)

	// Send the email to Bob, Cora and Dan
	mailer := gomail.NewMailer(ms.Smtp, ms.User, ms.Password, ms.Port)
	if err := mailer.Send(msg); err != nil {
		fmt.Println(err)
	}
}
