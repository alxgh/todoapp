package main

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

func renderPrompt(s tcell.Screen, prompt string) {
}

func main() {
	todos := []Todo{
		{
			Text: "Complete this app",
			Done: true,
		},
	}
	for i := 0; i < 200; i++ {
		t := todos[0]
		t.Text += fmt.Sprintf("%d", i)
		t.Done = i%2 == 0
		todos = append(todos, t)
	}

	_ = todos

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	s.SetStyle(defStyle)
	s.EnablePaste()
	s.SetCursorStyle(tcell.CursorStyleSteadyBlock)
	s.Clear()

	defer func() {
		a := recover()
		s.Fini()
		if a != nil {
			panic(a)
		}
	}()

	app := New(s, todos)
	app.Run()

}
