package main

type input struct {
	buf []rune
	c   int
}

func (i *input) Clear() {
	i.buf = nil
	i.c = 0
}

func (i *input) Erase() bool {
	if i.c == 0 {
		return false
	}
	i.buf = append(i.buf[0:i.c-1], i.buf[i.c:]...)
	i.c--
	return true
}

func (i *input) Get() string {
	buf := make([]rune, len(i.buf))
	copy(buf, i.buf)
	return string(buf)
}

func (i *input) Write(r rune) {
	i.buf = append(i.buf[0:i.c], append([]rune{r}, i.buf[i.c:]...)...)
	i.c++
}

func (i *input) WriteText(s string) {
	i.buf = append(i.buf[0:i.c], append([]rune(s), i.buf[i.c:]...)...)
	i.c += len(s)
}

func (i *input) Next() bool {
	if i.c >= len(i.buf) {
		return false
	}
	i.c++
	return true
}

func (i *input) Back() bool {
	if i.c <= 0 {
		return false
	}
	i.c--
	return true
}
