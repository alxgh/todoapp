package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

type stack[T any] struct {
	s []T
}

func (s *stack[T]) Len() int {
	return len(s.s)
}

func (s *stack[T]) Push(v T) {
	s.s = append(s.s, v)
}

func (s *stack[T]) Pop() (v T, ok bool) {
	if len(s.s) == 0 {
		return
	}
	ok = true
	l := len(s.s)
	v = s.s[l-1]
	s.s = s.s[0 : l-1]
	return
}

func (s *stack[T]) Last() (v T, ok bool) {
	if len(s.s) == 0 {
		return
	}
	ok = true
	l := len(s.s)
	v = s.s[l-1]
	return
}

type Pos struct {
	X int
	Y int
}

type page struct {
	name string
	fn   func()
}

type app struct {
	todos         []Todo
	startIdx      int
	cursorPos     Pos
	s             tcell.Screen
	prompt        string
	defaultPrompt string
	promptMode    bool
	insertMode    bool
	height        int
	width         int
	cursor        Pos
	prevCursor    Pos
	inp           input
	err           string
	exit          bool
	hideDoneTodos bool
	pages         stack[page]
	pageMap       map[string]page
}

func New(s tcell.Screen, todos []Todo) *app {
	w, h := s.Size()
	s.ShowCursor(0, 1)
	a := &app{
		todos:         todos,
		s:             s,
		defaultPrompt: "> Default",
		height:        h,
		width:         w,
		cursor:        Pos{X: 0, Y: 1},
	}

	a.pageMap = map[string]page{
		"main": {
			name: "main",
			fn:   a.renderTodos,
		},
		"help": {
			name: "help",
			fn:   a.renderHelp,
		},
	}

	a.pages.Push(a.pageMap["main"])
	return a
}

func (a *app) Run() {
	a.render()
	a.loop()
}

func (a *app) showCursor() {
	a.s.ShowCursor(a.cursor.X, a.cursor.Y)
}

func (a *app) exitPromptMode() {
	a.insertMode = false
	a.promptMode = false
	a.prompt = a.defaultPrompt
	a.renderPrompt()
	a.cursor = a.prevCursor
	a.showCursor()
	a.inp.Clear()
}

func (a *app) enterPromptMode() {
	a.insertMode = true
	a.promptMode = true
	a.prompt = ":"
	a.renderPrompt()
	a.prevCursor = a.cursor
	a.cursor.Y = a.height - 1
	a.cursor.X = a.inp.c + 1
	a.showCursor()
}

func (a *app) helpPage() {
	if lastPage, _ := a.pages.Last(); lastPage.name == "help" {
		a.goToLastPage()
	} else {
		a.pages.Push(a.pageMap["help"])
	}
	a.renderPage()
}

func (a *app) renderHelp() {
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	a.clear(0, -1, 0, -1, style)
	row := 0
	writeStr(a.s, 0, row, style, "Help")
	row++
	writeStr(a.s, 0, row, style, "? => Show this help")
	row++
	writeStr(a.s, 0, row, style, ":add => Add todo")
	row++
	writeStr(a.s, 0, row, style, ":edit [todoidx] => Edit todo")
	row++
	writeStr(a.s, 0, row, style, ":del [todoidx] => Delete todo")
	row++
	writeStr(a.s, 0, row, style, "c => (un)Mark todo")
	row++
	writeStr(a.s, 0, row, style, "c => (un)Mark todo")
	row++
	writeStr(a.s, 0, row, style, "e => Edit current todo")
	row++
	writeStr(a.s, 0, row, style, "d => Delete current todo")
	row++
	writeStr(a.s, 0, row, style, "a => (un)Hide completed todos")
	row++
}

func (a *app) writeTextToInp(s string) {
	a.inp.WriteText(s)
	a.cursor.X += len(s)
	a.showCursor()
	a.renderPrompt()
}

func (a *app) writeRuneToInp(r rune) {
	a.inp.Write(r)
	a.cursor.X += 1
	a.showCursor()
	a.renderPrompt()
}

func (a *app) promptModeHandle(ev *tcell.EventKey) bool {
	if !a.promptMode {
		return false
	}
	if ev.Key() == tcell.KeyCtrlC {
		a.exitPromptMode()
	}
	if a.promptMode && !a.insertMode && ev.Key() == tcell.KeyRune && ev.Rune() == 'i' {
		a.insertMode = true
		return true
	}
	if ev.Key() == tcell.KeyEscape {
		if a.promptMode && a.insertMode {
			a.insertMode = false
		} else if a.promptMode {
			a.exitPromptMode()
		}
	} else if ev.Key() == tcell.KeyEnter {
		a.err = ""
		// parse inp
		strinp := strings.TrimSpace(a.inp.Get())
		// exit
		if strinp == "q" {
			a.exit = true
			return true
		}
		if strings.HasPrefix(strinp, "add") {
			a.todos = append(a.todos, Todo{Text: strings.TrimSpace(strinp[3:])})
			a.render()
		} else if strings.HasPrefix(strinp, "edit") {
			idx := -1
			stridx := 0
			for i, c := range strinp[4:] {
				if c == ' ' {
					continue
				}
				if unicode.IsNumber(c) {
					if idx == -1 {
						idx = 0
					}
					n, _ := strconv.Atoi(string(c))
					idx = (idx * 10) + n
				} else {
					stridx = i
					break
				}
			}
			if idx == -1 || idx >= len(a.todos) {
				a.err = "Unknow todo item to edit!"
			} else {
				a.todos[idx].Text = strings.TrimSpace(strinp[stridx+4:])
				a.render()
			}
		} else {
			a.err = "Unknown command"
		}
		a.exitPromptMode()
	}
	if !a.insertMode {
		return false
	}
	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		if a.inp.Erase() {
			a.cursor.X--
			a.showCursor()
		}
		a.renderPrompt()
	} else if ev.Key() == tcell.KeyRune {
		a.writeRuneToInp(ev.Rune())
	}
	return true
}

func (a *app) cursorHandle(ev *tcell.EventKey) {
	if a.promptMode && !a.insertMode {
		if ev.Rune() == 'l' {
			if a.inp.Next() {
				a.cursor.X++
			}
		} else if ev.Rune() == 'h' {
			if a.inp.Back() {
				a.cursor.X--
			}
		}
		goto show_cursor
	}
	if a.promptMode {
		return
	}
	if ev.Rune() == 'l' {
		if a.cursor.X < a.width-1 {
			a.cursor.X++
		}
	} else if ev.Rune() == 'h' {
		if a.cursor.X > 0 {
			a.cursor.X--
		}
	} else if ev.Rune() == 'k' {
		if a.cursor.Y > 1 {
			a.cursor.Y--
		} else if a.startIdx > 0 {
			a.startIdx -= 1
		}
	} else if ev.Rune() == 'j' {
		if a.cursor.Y < a.height-2 {
			a.cursor.Y++
		} else if a.startIdx+a.height-2 < len(a.todos) {
			a.startIdx += 1
		}
	}
show_cursor:
	a.renderPage()
	a.showCursor()
}

func (a *app) loop() {
	for {
		a.s.Show()
		ev := a.s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			a.s.Sync()
			a.width, a.height = a.s.Size()
			a.render()
		case *tcell.EventKey:
			a.promptModeHandle(ev)
			// Is this piece of code useful?
			/*if ev.Key() == tcell.KeyCtrlL {
				s.Sync()
			}*/
			a.cursorHandle(ev)
			if !a.promptMode {
				if ev.Rune() == '?' {
					a.helpPage()
				} else if ev.Rune() == ':' {
					a.enterPromptMode()
				} else if ev.Rune() == 'n' {
					a.enterPromptMode()
					a.writeTextToInp("add ")
				} else if ev.Rune() == 'c' {
					todoIdx := a.startIdx + a.cursor.Y - 1
					if todoIdx >= len(a.todos) {
						continue
					}
					a.todos[todoIdx].Done = !a.todos[todoIdx].Done
					a.renderPage()
				} else if ev.Rune() == 'd' {
					todoIdx := a.startIdx + a.cursor.Y - 1
					if todoIdx < 0 || todoIdx >= len(a.todos) {
						continue
					}
					a.todos = append(a.todos[0:todoIdx], a.todos[todoIdx+1:]...)
					if a.startIdx+a.height-2 > len(a.todos) && a.startIdx > 0 {
						a.startIdx -= 1
					}
					a.renderPage()
				} else if ev.Rune() == 'e' {
					todoIdx := a.startIdx + a.cursor.Y - 1
					if todoIdx < 0 || todoIdx >= len(a.todos) {
						continue
					}
					a.enterPromptMode()
					a.writeTextToInp(fmt.Sprintf("edit %d ", todoIdx))
				} else if ev.Rune() == 'a' {
					a.hideDoneTodos = !a.hideDoneTodos
					a.renderPage()
				}
			}
		}
		if a.exit {
			return
		}
	}
}

func (a *app) renderPrompt() {
	style := tcell.StyleDefault.Background(tcell.ColorWhite).
		Foreground(tcell.ColorReset)
	errstyle := tcell.StyleDefault.Background(tcell.ColorWhite).
		Foreground(tcell.ColorRed)
	for i := 0; i < a.width; i++ {
		a.s.SetContent(i, a.height-1, ' ', nil, style)
	}
	prompt := a.prompt
	if a.promptMode {
		prompt = a.prompt + a.inp.Get()
	}
	writeStr(a.s, 0, a.height-1, style, prompt)
	if !a.promptMode && a.err != "" {
		writeStr(a.s, len(prompt), a.height-1, errstyle, fmt.Sprintf(" | %s", a.err))
	}
}

func (a *app) renderPage() {
	p, _ := a.pages.Last()
	p.fn()
}

func (a *app) changePage(pages page) {
	a.pages.Push(pages)
}

func (a *app) goToLastPage() {
	if a.pages.Len() > 1 {
		a.pages.Pop()
	}
}

func (a *app) clear(fw, tw, fc, tc int, style tcell.Style) {
	if tw == -1 {
		tw = a.width
	}
	if tc == -1 {
		tc = a.height - 1
	}
	for x := fw; x < tw; x += 1 {
		for y := fc; y < tc; y++ {
			a.s.SetContent(x, y, ' ', nil, style)
		}
	}
}

func (a *app) renderTodos() {
	style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	tickStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorSkyblue)
	row := 0
	a.clear(0, -1, 0, -1, style)
	writeStr(a.s, 0, row, style, "TODO LIST :-)")
	for _, todo := range a.todos[a.startIdx:] {
		if todo.Done && a.hideDoneTodos {
			continue
		}
		row++
		if row >= a.height-1 {
			break
		}
		writeStr(a.s, 0, row, style, fmt.Sprintf("[ ] %s", todo.Text))
		if todo.Done {
			writeStr(a.s, 1, row, tickStyle, "✓")
		}
	}
}

func (a *app) render() {
	a.renderPage()
	a.renderPrompt()
}
