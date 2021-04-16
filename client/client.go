package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

var Sendfile_flag bool = false
var transfer_flag bool = false

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(ip string, port int) *Client {
	newClient := &Client{
		ServerIp:   ip,
		ServerPort: port,
		flag:       999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	newClient.conn = conn

	return newClient
}

//init ip port
var serverIp string
var serverPort int
var file_path string
var receive_flage bool = false

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "set server ip")
	flag.IntVar(&serverPort, "port", 8888, "set server port")
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.pubilc")
	fmt.Println("2.private")
	fmt.Println("3.rename user")
	fmt.Println("0.exit")

	fmt.Scanln(&flag)
	client.judgement(string(flag))

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("input error")
		return false
	}
}

func (client *Client) menu1() bool {
	var flag int
	fmt.Println("1.pubilc")
	fmt.Println("2.private")
	fmt.Println("3.rename user")
	fmt.Println("4.friends")
	fmt.Println("0.exit")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 4 {
		client.flag = flag
		return true
	} else {
		fmt.Println("input error")
		return false
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println("input new name")
	var newname string
	fmt.Scanln(&newname)
	client.judgement(newname)
	sendmsg := "rename:" + newname + "\n"
	_, err := client.conn.Write([]byte(sendmsg))
	if err != nil {
		fmt.Println("conn.Write err", err)
		return false
	}
	client.Name = newname
	return true
}

func (client *Client) Friends() bool {
	fmt.Println("add friend")
	var newfriend string
	fmt.Scanln(&newfriend)
	client.judgement(newfriend)
	sendmsg := "add:" + newfriend
	_, err := client.conn.Write([]byte(sendmsg))
	if err != nil {
		fmt.Println("conn.Write err", err)
		return false
	}
	return true
}

func (client *Client) Transferfile() {
	f, err := os.Open(file_path)
	if err != nil {
		fmt.Println("os.Open err = ", err)
		return
	}
	defer f.Close()
	buf := make([]byte, 1024*4)
	for {
		n, err := f.Read(buf) //从文件读取内容
		if err != nil {
			if err == io.EOF {
				fmt.Println("文件发送完毕")
			} else {
				fmt.Println("f.Read err = ", err)
			}
			return
		}
		client.conn.Write(buf[:n]) //给服务器发送内容
	}
}

func (client *Client) DealResponse() {

	for {
		buf1 := make([]byte, 1024)
		n, err := client.conn.Read(buf1)
		buf := string(buf1[:n])
		if err != nil {
			fmt.Println("conn.Read err = ", err)
			return
		} else if strings.Contains(buf, "sendfile:") {
			//加上是否登录的判断
			if strings.Split(strings.Split(buf, "]")[1], ":")[0] == client.Name {
				receive_flage = true
			} else {
				fmt.Println("Whether or not to receive files:", strings.Split(buf, "sendfile:")[1], "intput:\n1. yes\n2. no")
				Sendfile_flag = true
				file_path = strings.Split(buf, "sendfile:")[1]
				return
			}
		} else {
			fmt.Print(buf)
		}
	}
}

func (client *Client) Selectmode() bool {
	var chatmsg string
	fmt.Println("Choice model\n1.login\n2.vistor\n,input '0' to quit")
	fmt.Scanln(&chatmsg)
	client.judgement(chatmsg)
	switch chatmsg {
	case "0":
		return false
	case "1":
		fmt.Println("input [accoutn:password]")
		chatmsg = ""
		fmt.Scanln(&chatmsg)
		client.judgement(chatmsg)
		_, err := client.conn.Write([]byte("login:" + chatmsg))
		if err != nil {
			fmt.Println("conn.Write err", err)
		}
		client.Run1()
		return true
	case "2":
		client.Run()
		return true
	}
	return false
}

func (client *Client) Sendfile(msg int) {
	Sendfile_flag = false
	fmt.Println("into Sendfile!!!", msg)
	switch msg {
	case 1:
		fmt.Println("receive...")
		client.conn.Write([]byte("1"))
		client.receivefile()
		break
	case 2:
		fmt.Println("refuse the request")
		client.conn.Write([]byte("2"))
		go client.DealResponse()
		break
	default:
		fmt.Println("input error")
		go client.DealResponse()
		break
	}
}

func (client *Client) receivefile() {
	f, err := os.Create("./new_receive.txt")
	if err != nil {
		fmt.Println("os.Create err = ", err)
		return
	}
	buf := make([]byte, 1024*4)
	for i := 0; i <= 2; i++ {
		if i == 0 {
			continue
		}
		n, err := client.conn.Read(buf) //接收对方发过来的文件内容
		if err != nil {
			if err == io.EOF {
				fmt.Println("文件接收完毕")
			} else {
				fmt.Println("conn.Read err = ", err)
			}
			return
		}
		f.Write(buf[:n]) //往文件写入内容
	}
	go client.DealResponse()
}

func (client *Client) judgement(msg string) {
	if Sendfile_flag == true {
		if msg == "1" {
			client.Sendfile(1)
		} else {
			client.Sendfile(2)
			go client.DealResponse()
		}

	}
	if transfer_flag == true {
		file_path = msg
	}
	if strings.Contains(msg, "sendfile:") && receive_flage == true {
		transfer_flag = true
		file_path = strings.Split(msg, ":")[1]
		client.Transferfile()
		go client.DealResponse()
		receive_flage = false
	}
}

func (client *Client) PublicChat() {
	var chatmsg string
	fmt.Println("input msg,input 'exit' to quit")

	for chatmsg != "exit" {
		if len(chatmsg) != 0 {
			sendmsg := chatmsg
			_, err := client.conn.Write([]byte(sendmsg))
			if err != nil {
				fmt.Println("conn.Write err", err)
				break
			}
		}
		chatmsg = ""
		fmt.Scanln(&chatmsg)
		client.judgement(chatmsg)
	}
}

func (client *Client) SelectUser() {
	_, err := client.conn.Write([]byte("who"))
	if err != nil {
		fmt.Println("conn.Write err", err)
		return
	}

}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUser()
	fmt.Println("select user : ,input 'exit' to quit")
	fmt.Scanln(&remoteName)
	client.judgement(remoteName)
	fmt.Println("you having a private chat with " + remoteName + "...")
	for chatMsg != "exit" {
		fmt.Scanln(&chatMsg)
		client.judgement(chatMsg)

		if len(chatMsg) != 0 {
			sendmsg := "to:" + remoteName + ":" + chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendmsg))
			if err != nil {
				fmt.Println("conn.Write err", err)
				break
			}
		}

	}

}

func (client *Client) Run() {
	for client.flag != 0 {
		//loop until user make right select
		for client.menu() != true {
		}

		switch client.flag {
		case 0:
			return
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break

		}
	}
}

func (client *Client) Run1() {
	for client.flag != 0 {
		//loop until user make right select
		for client.menu1() != true {
		}

		switch client.flag {
		case 0:
			return
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break
		case 4:
			client.Friends()
			break

		}
	}
}

func main() {

	//命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("link error")
		return
	}

	fmt.Println("link success")

	go client.DealResponse()

	res := client.Selectmode()
	if res == false {
		return
	}

}
