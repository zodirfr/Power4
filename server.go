package main

import (
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Game struct {
	Grid       [][]string
	Player     string
	Winner     string
	Difficulty string
	VsAI       bool
}

var (
	game  Game
	funcs = template.FuncMap{"title": titleCase, "seq": seq}
)

func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

func seq(start, end int) []int {
	s := make([]int, end-start+1)
	for i := range s {
		s[i] = start + i
	}
	return s
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", startPage)
	http.HandleFunc("/start", startGame)
	http.HandleFunc("/play", playMove)
	http.HandleFunc("/reset", resetGame)

	http.ListenAndServe(":8080", nil)
}

func startPage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/start.html")
	t.Execute(w, nil)
}

func startGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	diff := r.FormValue("difficulty")
	mode := r.FormValue("mode")

	game.Difficulty = diff
	game.Player = "red"
	game.Winner = ""
	game.VsAI = (mode == "ai")

	switch diff {
	case "easy":
		game.Grid = makeGrid(6, 7)
	case "normal":
		game.Grid = makeGrid(6, 9)
	case "hard":
		game.Grid = makeGrid(7, 8)
	default:
		game.Grid = makeGrid(6, 7)
	}

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func makeGrid(rows, cols int) [][]string {
	grid := make([][]string, rows)
	for i := range grid {
		grid[i] = make([]string, cols)
	}
	return grid
}

func renderGame(w http.ResponseWriter) {
	t := template.Must(template.New("index.html").Funcs(funcs).ParseFiles("templates/index.html"))
	t.Execute(w, game)
}

func playMove(w http.ResponseWriter, r *http.Request) {
	col, _ := strconv.Atoi(r.URL.Query().Get("col"))
	dropToken(col, game.Player)

	if checkWinner() {
		game.Winner = game.Player
		renderGame(w)
		return
	}

	switchTurn()

	if game.VsAI && game.Winner == "" {
		time.Sleep(700 * time.Millisecond)
		aiMove()
		checkWinner()
	}

	renderGame(w)
}

func dropToken(col int, player string) bool {
	for i := len(game.Grid) - 1; i >= 0; i-- {
		if game.Grid[i][col] == "" {
			game.Grid[i][col] = player
			return true
		}
	}
	return false
}

func aiMove() {
	cols := len(game.Grid[0])
	var validCols []int
	for c := 0; c < cols; c++ {
		if game.Grid[0][c] == "" {
			validCols = append(validCols, c)
		}
	}
	if len(validCols) == 0 {
		return
	}
	move := validCols[rand.Intn(len(validCols))]
	dropToken(move, game.Player)
	switchTurn()
}

func switchTurn() {
	if game.Player == "red" {
		game.Player = "yellow"
	} else {
		game.Player = "red"
	}
}

func resetGame(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkWinner() bool {
	rows := len(game.Grid)
	cols := len(game.Grid[0])

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			player := game.Grid[i][j]
			if player == "" {
				continue
			}

			if j+3 < cols && player == game.Grid[i][j+1] && player == game.Grid[i][j+2] && player == game.Grid[i][j+3] {
				return true
			}
			if i+3 < rows && player == game.Grid[i+1][j] && player == game.Grid[i+2][j] && player == game.Grid[i+3][j] {
				return true
			}
			if i+3 < rows && j+3 < cols && player == game.Grid[i+1][j+1] && player == game.Grid[i+2][j+2] && player == game.Grid[i+3][j+3] {
				return true
			}
			if i-3 >= 0 && j+3 < cols && player == game.Grid[i-1][j+1] && player == game.Grid[i-2][j+2] && player == game.Grid[i-3][j+3] {
				return true
			}
		}
	}
	return false
}
