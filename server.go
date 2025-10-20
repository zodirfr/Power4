package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Game struct {
	Grid       [][]string
	Player1    string
	Player2    string
	Current    string
	Winner     string
	Rows, Cols int
	VsAI       bool
	Mutex      sync.Mutex
}

var game Game

var templates = template.Must(template.New("").Funcs(template.FuncMap{
	"title": func(s string) string {
		if len(s) == 0 {
			return s
		}
		return string(s[0]-32) + s[1:]
	},
}).ParseGlob("templates/*.html"))

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handleStart)
	http.HandleFunc("/start", handleStartGame)
	http.HandleFunc("/game", handleGame)
	http.HandleFunc("/play", handlePlay)
	http.HandleFunc("/reset", handleReset)

	log.Println("Serveur lancé sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "start.html", nil)
}

func handleStartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	player1 := r.FormValue("player1")
	player2 := r.FormValue("player2")
	mode := r.FormValue("mode")
	difficulty := r.FormValue("difficulty")

	rows, cols := 6, 7
	switch difficulty {
	case "normal":
		cols = 9
	case "hard":
		rows = 7
	}

	game = Game{
		Grid:    makeGrid(rows, cols),
		Player1: player1,
		Player2: player2,
		Current: "red",
		Rows:    rows,
		Cols:    cols,
		VsAI:    (mode == "ai"),
	}

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func handleGame(w http.ResponseWriter, r *http.Request) {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	templates.ExecuteTemplate(w, "index.html", game)
}

func handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	colStr := r.FormValue("col")
	col, err := strconv.Atoi(colStr)
	if err != nil {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	if game.Winner != "" {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	dropToken(col, game.Current)

	if checkWinner() {
		game.Winner = game.Current
	} else if isFull() {
		game.Winner = "draw"
	} else {
		if game.Current == "red" {
			game.Current = "yellow"
		} else {
			game.Current = "red"
		}

		if game.VsAI && game.Current == "yellow" {
			game.Mutex.Unlock()
			time.Sleep(1 * time.Second)
			game.Mutex.Lock()

			colAI := rand.Intn(game.Cols)
			for !canPlay(colAI) {
				colAI = rand.Intn(game.Cols)
			}
			dropToken(colAI, game.Current)

			if checkWinner() {
				game.Winner = game.Current
			} else if isFull() {
				game.Winner = "draw"
			} else {
				game.Current = "red"
			}
		}
	}

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ---- Fonctions de jeu ----

func makeGrid(rows, cols int) [][]string {
	grid := make([][]string, rows)
	for i := range grid {
		grid[i] = make([]string, cols)
	}
	return grid
}

func canPlay(col int) bool {
	return game.Grid[0][col] == ""
}

func dropToken(col int, player string) {
	for i := game.Rows - 1; i >= 0; i-- {
		if game.Grid[i][col] == "" {
			game.Grid[i][col] = player
			break
		}
	}
}

func isFull() bool {
	for i := 0; i < game.Cols; i++ {
		if game.Grid[0][i] == "" {
			return false
		}
	}
	return true
}

func checkWinner() bool {
	g := game.Grid
	r, c := game.Rows, game.Cols
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if g[i][j] == "" {
				continue
			}
			if j+3 < c && g[i][j] == g[i][j+1] && g[i][j] == g[i][j+2] && g[i][j] == g[i][j+3] {
				return true
			}
			if i+3 < r && g[i][j] == g[i+1][j] && g[i][j] == g[i+2][j] && g[i][j] == g[i+3][j] {
				return true
			}
			if i+3 < r && j+3 < c && g[i][j] == g[i+1][j+1] && g[i][j] == g[i+2][j+2] && g[i][j] == g[i+3][j+3] {
				return true
			}
			if i-3 >= 0 && j+3 < c && g[i][j] == g[i-1][j+1] && g[i][j] == g[i-2][j+2] && g[i][j] == g[i-3][j+3] {
				return true
			}
		}
	}
	return false
}
