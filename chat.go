package main

import (
	"fmt"
	"net"
	"os"
)

type Client struct {
	connection net.Conn
	channel   chan<- string
}

func main() {
	listen, err := net.Listen("tcp", ":7000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	addChannel := make(chan Client)
	removeChannel := make(chan Client)
	messageChannel := make(chan string)
	go handleMessages(messageChannel, addChannel, removeChannel)

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnection(conn, messageChannel, addChannel, removeChannel)
	}
}

func handleConnection(connection net.Conn, messageChannel chan<- string, addChannel chan<- Client, removeChannel chan<- Client) {
	buffer := make([]byte, 4096)
	channel := make(chan string)


	addChannel <- Client{connection, channel}

	defer func() {
		removeChannel <- Client{connection, channel}
	}()

	for {
		n, err := connection.Read(buffer)
		if err != nil || n == 0 {
			connection.Close()
			break
		}

		messageChannel <-string(buffer[0:n])

		n, err = connection.Write(buffer[0:n])
		if err != nil {
			connection.Close()
			break
		}
	}

	fmt.Printf("Connection from %v closed.\n", connection.RemoteAddr())
}

func handleMessages(messageChannel <- chan string, addChannel <- chan Client, removeChannel <- chan Client) {
	clients := make(map[net.Conn]chan <- string)

	for {
		select {
		case message := <- messageChannel:
			for _, ch := range clients {
				go func(mch chan <- string) { mch <- "\033[1;33;40m" + message + "\033[m\r\n" } (ch)
			}
		case client := <- addChannel:
			clients[client.connection] = client.channel
		case conn := <-removeChannel:
			delete(clients, conn.channel)
		}
//		message := <- messageChannel
//		fmt.Printf("New message: %s\n", message)
	}
}