package TCP

import (
	"fmt"
	"net"
)

type Client struct {
	conn net.Conn
	username string
	addr string
}

var (
	selfName string
	msg = make(chan string)
	clientsMap map[string] *Client = make(map[string] *Client)
)

//read chan of msg and broadcast
func broadcastMsg()  {
	for {
		//block if msg is nil
		msgByte := []byte(<-msg + "\n")

		for _, client := range clientsMap {
			_, err := client.conn.Write(msgByte)
			if err != nil {
				fmt.Println("broadcastMsg err", err)
				return
			}
		}
	}
}

//format ByteMessage to string
func formatByte(username string, byteMsg []byte) string {
	return username + " : " + string(byteMsg)
}

//tcp logout
func Logout(client *Client)  {
	delete(clientsMap, client.username)
	msg <- client.username + " is logout "
	err := client.conn.Close()
	if err != nil {
		fmt.Println("logout conn err: ", err)
		return
	}
}

//print online user list
func PrintList(conn net.Conn) {
	conn.Write([]byte("online users as follows:\nlist begin:\n"))
	for username, _ := range clientsMap {
		conn.Write([]byte(username + "\n"))
	}
	conn.Write([]byte("list end.\n"))
}

//read message of chat
func solveConn(client *Client, selfName string)  {
	buff := make([]byte, 1024)

	for {
		n, err := client.conn.Read(buff)
		n--
		if err != nil {
			fmt.Println("solveConn err:", err)
			return
		}

		if string(buff[:n]) == "/logout" {
			fmt.Println(client.username, " is logout")
			Logout(client)
		} else if string(buff[:n]) == "/list" {
			fmt.Println(client.username, " request user list")
			PrintList(client.conn)
		} else if n > 0 {
			msg <- formatByte(selfName, buff[:n])
		}
	}
	
}

//start listening and access client
func Start() {
	//start listening
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("listen err: ", err)
		return
	}
	defer listener.Close()

	go broadcastMsg()

	//access client
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept err: ", err)
			continue
		}

		var clientName string
		buff := make([]byte, 1024)

		//read username
		for {
			n, err := conn.Read(buff)
			if err != nil {
				fmt.Println("read username err: ", err)
				continue
			}

			// \n
			n--

			if n > 0 {
				clientName = string(buff[:n])
				fmt.Println(clientName, "is login")
				msgString := clientName + " is login"
				msg <- msgString
				selfName = clientName
				break
			}
		}

		//put related struct in map
		client := Client{conn: conn, username: selfName, addr: conn.RemoteAddr().String()}
		clientsMap[clientName] = &client

		//start chat
		go solveConn(&client, selfName)
	}
}
