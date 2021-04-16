package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//online user list
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//broadcast msg
	Message chan string
}

//create a server API
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func CreateAccount(newaccount *Account) bool {
	filePtr, err := os.OpenFile(accoutn_name, os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("open json file failed，err=", err)
		return false
	}
	defer filePtr.Close()
	encoded := json.NewEncoder(filePtr)
	fmt.Println(newaccount)
	err = encoded.Encode(newaccount)
	if err != nil {
		fmt.Println("creat account failed，err=", err)
		return false
	} else {
		fmt.Println("creat account succeed")
		return true
	}

}

//Record user behavior
func UserBehavior(user *User, behavior string) {
	log_lock.Lock()
	f, err1 = os.OpenFile(log_name, os.O_APPEND, 0666)
	if err1 != nil {
		fmt.Println("OpenFile error")
	} else {
		_, err1 = io.WriteString(f, time.Now().Format("15:04:05")+": "+behavior)
		f.Close()
	}
	log_lock.Unlock()
}

//listen message ,once get msg,boradcast to all online users
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//boradcast
func (this *Server) BoradCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
	// fmt.Println(sendMsg)
}

func (this *Server) Handle(conn net.Conn) {
	//fmt.Println("connect success")

	user := NewUser(conn, this)

	//online map
	user.Online()
	log_lock.Lock()
	f, err1 = os.OpenFile(log_name, os.O_APPEND, 0666)
	if err1 != nil {
		fmt.Println("OpenFile error")
	} else {
		_, err1 = io.WriteString(f, time.Now().Format("2006-01-02 15:04:05")+":"+user.Name+" is online\n")
		f.Close()
	}
	log_lock.Unlock()

	//user online state
	islive := make(chan bool)

	//read client msg
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				log_lock.Lock()
				f, err1 = os.OpenFile(log_name, os.O_APPEND, 0666)
				if err1 != nil {
					fmt.Println("OpenFile error")
				} else {
					_, err1 = io.WriteString(f, time.Now().Format("2006-01-02 15:04:05")+":"+user.Name+" is offline\n")
					f.Close()
				}
				log_lock.Unlock()
				return
			}
			//server read error log
			if err != nil && err != io.EOF {
				log_lock.Lock()
				f, err1 = os.OpenFile(log_name, os.O_APPEND, 0666)
				if err1 != nil {
					fmt.Println("OpenFile error")
				} else {
					_, err1 = io.WriteString(f, "conn read err:"+fmt.Sprintf("%s", err)+"\n")
					f.Close()
				}
				log_lock.Unlock()
				return
			}

			//extarct user's msg delete \n
			if n == 4096 {
				n = n - 1
			}
			msg := string(buf[:n])
			user.DoMessage(msg)
			islive <- true

		}

	}()

	for {
		select {
		case <-islive:
		case <-time.After(time.Second * 600):
			user.SendMsg("you have been offline")
			close(user.C)
			conn.Close()
			return
		}
	}
}

//server log
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func init_log() {
	if checkFileIsExist(log_name) {
		//如果文件存在
		f, err1 = os.OpenFile(log_name, os.O_APPEND, 0666)
		//打开文件
		fmt.Println("log is exist")
	} else {
		f, err1 = os.Create(log_name)
		//创建文件
		fmt.Println("log has been created")
	}
	_, err1 = io.WriteString(f, "server open on:"+time.Now().Format("2006-01-02 15:04:05\n"))
	if err1 != nil {
		fmt.Println("server log init error")
	}
	f.Close()
}

func init_account() {
	if checkFileIsExist(accoutn_name) {
		f, err1 = os.OpenFile(accoutn_name, os.O_APPEND, 0666)
		fmt.Println("accoutn_list is exist")
	} else {
		f, err1 = os.Create(accoutn_name)
		fmt.Println("accoutn_list has been created")
	}
	f.Close()
}

func (this *Server) Start() {

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port)) //改成udp试试？
	if err != nil {
		fmt.Println("connection error ：", err)
		return
	}
	defer listener.Close()

	go this.ListenMessager()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener err : ", err)
			continue
		}

		go this.Handle(conn)
	}

}

func init() {
	init_log()
	init_account()
}

var log_name = "./server_log.txt"
var accoutn_name = "./account_log.json"
var f *os.File
var log_lock sync.RWMutex
var err1 error

// 附加功能：
// 1.服务器日志  complete
// 2.账号系统    complete
// 3.好友系统    complete
// 4.加密传输
// 5.文件传输
