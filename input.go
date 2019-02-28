package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

var inputFile = []string {
	"input/a_example.txt",
	"input/b_lovely_landscapes.txt",
	"input/c_memorable_moments.txt",
	"input/d_pet_pictures.txt",
	"input/e_shiny_selfies.txt",
}

type Orientation uint8

const (
	Horizontal Orientation = iota
	Vertical
)

func OrientationFromString(ch string) Orientation {
	switch ch {
	case "H":
		return Horizontal
	case "V":
		return Vertical
	default:
		panic(ch)
	}
}

func parseInput(index int) []Photo {
	f, _ := os.Open(inputFile[index])
	scanner := bufio.NewScanner(f)

	scanner.Scan()
	numLines, _ := strconv.Atoi(scanner.Text())
	lines := make([]Photo, numLines)

	pos := uint16(0)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		lines[pos] = Photo{
			//Id: pos,
			Orientation: OrientationFromString(parts[0]),
			Tags: parts[2:],
		}
		pos++
	}

	return lines
}
