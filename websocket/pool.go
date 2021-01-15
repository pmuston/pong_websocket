package websocket

import "fmt"

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
	Game       chan Message
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message, 2),
		Game:       make(chan Message),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Register: Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				fmt.Println("sending New User Joined...")
				client.Conn.WriteJSON(Message{Type: 3, Body: "New User Joined..."})
			}
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("UnRegister: Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				fmt.Println("sending User Disconnected...")
				client.Conn.WriteJSON(Message{Type: 3, Body: "User Disconnected..."})
			}
			break
		case message := <-pool.Broadcast:
			if message.Type == 1 {
				fmt.Println("Sending message to Game")
				pool.Game <- message
			} else {
				fmt.Println("Sending message to all clients in Pool", message.Body)
				for c := range pool.Clients {
					if err := c.Conn.WriteJSON(message); err != nil {
						fmt.Println(err)
						// return
					}
				}
			}
		}
	}
}
