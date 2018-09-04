/*

Ask for address to connect to
Connect to the server, and add it to a list of connections
Start listening for incoming connections


*/

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	connections = append(connections, conn)

	otherEnd := conn.RemoteAddr().String()

	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')

		if err != nil {
			fmt.Println("Ending session with " + otherEnd)
			addArrow()

			// Finding the index of conn in connections
			index := -1

			for connIndex, tempConn := range connections {
				if conn == tempConn {
					index = connIndex
					break
				}
			}

			// Remove the connection from the array
			if index != -1 {
				connections = append(connections[:index], connections[index+1:]...)
			}

			return
		} else {
			outbound <- msg
		}
	}
}

func listenForConnections() {

	fmt.Println("Listening for connections...")
	ln, _ := net.Listen("tcp", ":")
	defer ln.Close()

	// Printing the port
	host, port, _ := net.SplitHostPort(ln.Addr().String())
	fmt.Println("Host: " + host)
	fmt.Println("Port: " + port)

	addArrow()

	for {
		conn, _ := ln.Accept()
		fmt.Println("Got a connection ", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func broadcast() {
	for {
		msg := <-outbound

		if messagesSent[msg] == true {
			continue
		} else {
			messagesSent[msg] = true
		}

		t := time.Now()
		fmt.Print("[", t.Format("15:04:05"), "] ", msg)

		for i := 0; i < len(connections); i++ {
			connections[i].Write([]byte(msg))
		}

		addArrow()
	}
}

func addArrow() {
	fmt.Print("> ")
}

var outbound = make(chan string)
var connections = []net.Conn{}
var messagesSent = make(map[string]bool)

func main() {

	reader := bufio.NewReader(os.Stdin)

	// Get IP
	fmt.Print("Enter IP address> ")
	ip, _ := reader.ReadString('\n')

	if ip == "" {
		ip = "127.0.0.1"
	}

	// Get port
	fmt.Print("Enter port> ")
	port, _ := reader.ReadString('\n')

	// Establish connection
	fullAddress := strings.Replace(ip+":"+port, "\n", "", -1)
	conn, err := net.Dial("tcp", fullAddress)
	if err != nil {
		fmt.Println("No peer found or invalid IP or Port")
	} else {
		go handleConnection(conn)
	}

	// Start broadcasting messages
	go broadcast()

	// Start listening for new connections
	go listenForConnections()

	// Start listening for input
	addArrow()
	for {
		reader := bufio.NewReader(os.Stdin)

		// Begin chat loop
		for {
			text, _ := reader.ReadString('\n')

			if text == "quit\n" {
				return
			}

			outbound <- text
		}
	}
}
