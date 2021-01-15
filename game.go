package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pmuston/pong_websocket/websocket"
)

// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

type GameStatus struct {
	A       int     `json:"a"`
	B       int     `json:"b"`
	X       float32 `json:"x"`
	Y       float32 `json:"y"`
	Dx      float32 `json:"dx"`
	Dy      float32 `json:"dy"`
	Running bool    `json:"running"`
}

type KeyPress struct {
	Key string `json:"key"`
}

func (gs *GameStatus) Update() {
	if gs.Running {
		gs.X += gs.Dx
		gs.Y += gs.Dy
		if gs.X == 20 {
			if gs.Y >= float32(gs.A) && gs.Y <= float32(gs.A+100) {
				gs.Dx = -gs.Dx
			}
		}
		if gs.X == 570 {
			if gs.Y >= float32(gs.B) && gs.Y <= float32(gs.B+100) {
				gs.Dx = -gs.Dx
			}
		}
		if gs.X > 600 || gs.X < 0 {
			gs.Running = false
		}
		if gs.Y >= 390 || gs.Y <= 0 {
			gs.Dy = -gs.Dy
		}
	}
}

type Game struct {
	Pool   *websocket.Pool
	Status *GameStatus
	mu     sync.Mutex
}

func NewGame(pool *websocket.Pool) *Game {
	fmt.Println("NewGame")
	status := &GameStatus{
		A:       0,
		B:       0,
		X:       300,
		Y:       200,
		Dx:      10,
		Dy:      10,
		Running: false,
	}
	return &Game{
		Pool:   pool,
		Status: status,
	}
}

func (game *Game) Update() {
	game.Status.Update()
	if game.Status.Running {
		bytes, _ := json.Marshal(game.Status)
		message := websocket.Message{Type: 2, Body: string(bytes)}
		game.Pool.Broadcast <- message
	}
}

func (game *Game) Start() {
	fmt.Println("Game Start")
	ticker := time.NewTicker(100 * time.Millisecond)
	done := make(chan bool)
	for {
		select {
		case <-done:
			fmt.Println("done")
			return
		case <-ticker.C:
			// fmt.Println("Tick at", t)
			game.mu.Lock()
			game.Update()
			game.mu.Unlock()
		case message := <-game.Pool.Game:
			fmt.Println("keystroke", message)
			if message.Type == 1 {
				strokeBytes := []byte(message.Body)
				var stroke KeyPress
				err := json.Unmarshal(strokeBytes, &stroke)
				if err != nil {
					fmt.Println(err)
				}
				game.mu.Lock()
				if stroke.Key == " " && !game.Status.Running {
					game.Status.Running = true
					game.Status.X = 300
					game.Status.Y = 200
					game.Status.Dx = 10
					game.Status.Dy = 10
				}
				if stroke.Key == "q" {
					game.Status.A -= 10
				}
				if stroke.Key == "a" {
					game.Status.A += 10
				}
				if stroke.Key == "o" {
					game.Status.B -= 10
				}
				if stroke.Key == "l" {
					game.Status.B += 10
				}
				game.mu.Unlock()
			}
		}
	}
}
