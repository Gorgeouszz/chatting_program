package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server

	login bool

	friends []string
}

type Account struct {
	User_account     string
	User_password    string
	User_Accountname string
	User_friends     []string
}

func NewAccount(account string, password string, name string) *Account {
	newaccount := &Account{
		User_account:     account,
		User_password:    password,
		User_Accountname: name,
	}
	return newaccount
}

func (this *User) AddFriend(name string, account_log string) {
	if !this.login {
		this.conn.Write([]byte("you have not logged in\n"))
	}
	var tmp Account
	f, err := os.OpenFile(account_log, os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("open file error:", err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	for decoder.More() {
		err = decoder.Decode(&tmp)
		fmt.Println("More!!!!!!!!!!!!!!!!!!!!", name, tmp.User_Accountname, len(name), len(tmp.User_Accountname))
		if err != nil {
			fmt.Println("decode err，err=", err)
		} else if tmp.User_Accountname == name {
			fmt.Println("find!!!!!!!!!!!!!!!!!!!!")
			this.friends = append(this.friends, name)
			tmp.User_friends = append(tmp.User_friends, name)
			encoder := json.NewEncoder(f)
			fmt.Println(tmp)
			err = encoder.Encode(tmp)
			if err != nil {
				this.conn.Write([]byte("add friends err: \n"))
			} else {
				this.conn.Write([]byte("add friends success: \n"))
			}
			return
		}
	}
}

func (this *User) User_Login(useraccount string, userpassword string, account_log string) {
	var tmp Account
	f, err := os.OpenFile(account_log, os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("open file error:", err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)

	for decoder.More() {
		err := decoder.Decode(&tmp)
		if err != nil {
			fmt.Println("decode err，err=", err)
		} else if tmp.User_account == useraccount && tmp.User_password == userpassword {
			this.Link_user(tmp)
			return

		}
	}
}

// link user-account give user permissions
func (this *User) Link_user(a Account) {
	this.Log_Behavior(this.Name + " login " + a.User_Accountname + "\n")
	this.Name = a.User_Accountname
	this.login = true
	this.conn.Write([]byte("you have logged in " + this.Name + "\n"))

}

//create user's API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,

		login: false,
	}

	//start the user channel goroutine
	go user.ListenMessage()

	return user

}

func (this *User) Log_Behavior(behave string) {
	UserBehavior(this, behave)
}

func (this *User) Online() {

	//user online
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//boradcast msg
	this.server.BoradCast(this, "user online ")

}

func (this *User) Offline() {

	//user offline
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//boradcast msg
	this.server.BoradCast(this, "user offline ")
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *User) DoMessage(msg string) {
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "online...\n"
			this.SendMsg(onlineMsg)
			user.Log_Behavior(this.Name + "Query list" + "\n")
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 9 && msg[:9] == "register:" {
		newaccount := NewAccount(strings.Split(msg, ":")[1], strings.Split(msg, ":")[2], strings.Split(msg, ":")[3])
		res := CreateAccount(newaccount)
		if res {
			this.Log_Behavior(this.Name + "register:" + newaccount.User_Accountname + "\n")
		}
	} else if len(msg) > 4 && msg[:4] == "add:" {
		this.AddFriend(strings.Split(msg, ":")[1], accoutn_name)
	} else if len(msg) > 6 && msg[:6] == "login:" {
		this.User_Login(strings.Split(msg, ":")[1], strings.Split(msg, ":")[2], accoutn_name)
	} else if len(msg) > 7 && msg[:7] == "rename:" {
		newname := strings.Split(msg, ":")[1]
		_, ok := this.server.OnlineMap[newname]
		if ok {
			this.SendMsg("this name are used \n")
		} else {
			//delete \n
			newname = strings.Replace(newname, "\n", "", -1)
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.Log_Behavior(this.Name + " rename as " + strings.Replace(newname, "\n", "", -1) + "\n")
			this.Name = newname
			this.server.OnlineMap[this.Name] = this
			this.server.mapLock.Unlock()
			this.SendMsg("you have been renamed: " + newname + "\n")

		}
	} else if len(msg) > 3 && msg[:3] == "to:" {
		remotename := strings.Split(msg, ":")[1]
		if remotename == "" {
			this.SendMsg("format error ,use \"to:user:content\"")
			return
		}
		remoteuser, ok := this.server.OnlineMap[remotename]
		if !ok {
			this.SendMsg("user not exist")
		} else {
			remoteuser.SendMsg(this.Name + ": " + strings.Split(msg, ":")[2])
			this.Log_Behavior(this.Name + " Whisper to " + strings.Replace(remotename, "\n", "", -1) + strings.Split(msg, ":")[2])
		}

	} else {
		this.Log_Behavior(this.Name + ":" + msg + "\n")
		this.server.BoradCast(this, msg)
	}

}

//listen User channel's func ,send message to client
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))

	}
}
