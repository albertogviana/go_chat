package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Client struct {
	connection net.Conn
	channel chan string
	username string
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
	bufioConnection := bufio.NewReader(connection)

	client := Client{
		connection: connection,
		username: getUsername(connection, bufioConnection),
		channel: make(chan string),
	}

	// Allow a user to quit the chat
	defer func() {
		removeChannel <- client
	}()

	if strings.TrimSpace(client.username) == "" {
		io.WriteString(connection, "Invalid username\n")
		connection.Close()
		return
	}

	addChannel <- client

	io.WriteString(connection, fmt.Sprintf("Welcome, %s!\n\n", client.username))
	messageChannel <- fmt.Sprintf("New user %s has joined the chat room.\n", client.username)

	// I/O
	go client.ReadLines(messageChannel)
	client.WriteLinesFrom(client.channel)
}

func handleMessages(messageChannel <- chan string, addChannel <- chan Client, removeChannel <- chan Client) {
	clients := make(map[net.Conn]chan <- string)

	for {
		select {
		case message := <- messageChannel:
			fmt.Printf("%v", message)
			for _, ch := range clients {
				go func(mch chan <- string) { mch <- "\033[1;33;40m" + message + "\033[m\r\n" } (ch)
			}
		case client := <- addChannel:
			fmt.Printf("New client: %v \n", client.connection.RemoteAddr())
			clients[client.connection] = client.channel
		case client := <- removeChannel:
			fmt.Printf("Connection from %v closed.\n", client.connection.RemoteAddr())
			delete(clients, client.connection)
		}
	}
}

func getUsername(connection net.Conn, bufioConnection *bufio.Reader) string {
	io.WriteString(connection, "Welcome to go chat!\n")
	io.WriteString(connection, "Please, provide your username: ")
	username, _, _ := bufioConnection.ReadLine()
	return string(username)
}

func (client Client) ReadLines(channel chan<- string) {
	bufioConnection := bufio.NewReader(client.connection)
	for {
		line, err := bufioConnection.ReadString('\n')
		if err != nil {
			break
		}
		channel <- fmt.Sprintf("%s: %s", client.username, line)
	}
}

func (client Client) WriteLinesFrom(channel <-chan string) {
	for message := range channel {
		_, err := io.WriteString(client.connection, message)
		if err != nil {
			return
		}
	}
}