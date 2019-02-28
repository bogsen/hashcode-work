package main

import "fmt"

type Photo struct {
	Tags []string
	//Id uint16
	Orientation Orientation
}

var photos []Photo

func main() {
	photos = parseInput(0)

	vert := 0
	horiz := 0
	for _, photo := range photos {
		if photo.Orientation == Horizontal {
			horiz++
		} else {
			vert++
		}
	}
	fmt.Println(vert, horiz)
}
