package main

import (
	"bufio"
	"fmt"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

var inputFile = []string {
	"a_example.txt",
	"b_lovely_landscapes.txt",
	"c_memorable_moments.txt",
	"d_pet_pictures.txt",
	"e_shiny_selfies.txt",
}

type Orientation uint8
type PhotoId uint32

const (
	Horizontal Orientation = iota
	Vertical
	TwoVertical
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

	allTags := make(map[string]int)
	tagIndex := 0

	pos := 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		tags := make([]int, len(parts) - 2)
		for i := 2; i < len(parts); i++ {
			idx, already := allTags[parts[i]]
			if already {
				tags[i - 2] = idx
			} else {
				allTags[parts[i]] = tagIndex
				tags[i - 2] = tagIndex
				tagIndex++
			}
		}
		sort.Ints(tags)
		lines[pos] = Photo{
			Id: PhotoId(pos),
			Orientation: OrientationFromString(parts[0]),
			Tags: tags,
		}
		pos++
	}

	for i, photo := range photos {
		if int(photo.Id) != i {
			panic(i)
		}
	}

	return lines
}

type Photo struct {
	Tags []int
	Id PhotoId
	SecondId PhotoId
	Orientation Orientation
}

func (p *Photo) Slide() string {
	if p.Orientation == TwoVertical {
		return fmt.Sprint(p.Id, p.SecondId)
	} else {
		return fmt.Sprint(p.Id)
	}
}

var photos []Photo

func removePhoto(i int) Photo {
	photo := photos[i]
	photos[i], photos[len(photos) - 1] = photos[len(photos) - 1], photos[i]
	photos = photos[:len(photos) - 1]
	return photo
}

func tagsInBoth(a, b *Photo) (left, right, common int) {
	i, j := 0, 0
	ret := 0
	for i != len(a.Tags) && j != len(b.Tags) {
		x, y := a.Tags[i], b.Tags[j]
		if x < y {
			i++
		} else if y < x {
			j++
		} else {
			ret++
			i++
			j++
		}
	}
	return len(a.Tags) - ret, len(b.Tags) - ret, ret
}

func calcScore(from, to *Photo) int {
	a, b, common := tagsInBoth(from, to)

	if common < a {
		a = common
	}
	if b < a {
		a = b
	}
	return a
}

type Work struct {
	from *Photo
	a, b int
}
type WorkResult struct {
	next, score int
}
var work chan Work
var results chan WorkResult
var numWorkers = 8

func scoreWorker() {
	runtime.LockOSThread()
	for w := range work {
		bestIndex := -1
		bestScore := math.MinInt32
		for i := w.a; i <= w.b; i++ {
			score := calcScore(w.from, &photos[i])
			if score > bestScore {
				bestIndex = i
				bestScore = score
			}
		}
		results <- WorkResult{bestIndex, bestScore}
	}
}

var work2 chan Work
var results2 chan WorkResult

func score2Worker() {
	runtime.LockOSThread()
	for w := range work2 {
		bestIndex := -1
		bestScore := math.MinInt32
		for i := w.a; i <= w.b; i++ {
			if photos[i].Orientation != Vertical {
				continue
			}
			score := verticalScore(w.from, &photos[i])
			if score > bestScore {
				bestIndex = i
				bestScore = score
			}
		}
		results2 <- WorkResult{bestIndex, bestScore}
	}
}

func findBestScore(from *Photo) (next int, score int) {
	chunk := int(math.Ceil(float64(len(photos)) / float64(numWorkers)))
	for i := 0; i < numWorkers; i++ {
		start := chunk * i
		end := start + chunk - 1
		if end >= len(photos) {
			end = len(photos) - 1
		}
		work <- Work{from, start, end}
	}
	bestIndex := -1
	bestScore := math.MinInt32
	for i := 0; i < numWorkers; i++ {
		result := <- results
		if result.score > bestScore {
			bestIndex = result.next
			bestScore = result.score
		}
	}
	return bestIndex, bestScore
}

func chooseNextNode(from *Photo) (next int, score int) {
	bestIndex := -1
	bestScore := math.MinInt32
	for i, photo := range photos {
		score := calcScore(from, &photo)
		if score > bestScore {
			bestIndex = i
			bestScore = score
		}
	}
	return bestIndex, bestScore
}

func findBestScore2(from *Photo) (next int, score int) {
	chunk := int(math.Ceil(float64(len(photos)) / float64(numWorkers)))
	for i := 0; i < numWorkers; i++ {
		start := chunk * i
		end := start + chunk - 1
		if end >= len(photos) {
			end = len(photos) - 1
		}
		work2 <- Work{from, start, end}
	}
	bestIndex := -1
	bestScore := math.MinInt32
	for i := 0; i < numWorkers; i++ {
		result := <- results2
		if result.score > bestScore {
			bestIndex = result.next
			bestScore = result.score
		}
	}
	return bestIndex, bestScore
}

func chooseNextNode2(from *Photo) (next int, score int) {
	bestIndex := -1
	bestScore := math.MinInt32
	for i, photo := range photos {
		if photo.Orientation != Vertical {
			continue
		}
		score := verticalScore(from, &photo)
		if score > bestScore {
			bestIndex = i
			bestScore = score
		}
	}
	return bestIndex, bestScore
}

func inputStats() {
	vert := 0
	horiz := 0
	for _, photo := range photos {
		if photo.Orientation == Horizontal {
			horiz++
		} else {
			vert++
		}
	}
	fmt.Printf("H=%v   V=%v\n", horiz, vert)

	/*allTags := make(map[string]bool)
	for _, photo := range photos {
		for tag := range photo.Tags {
			allTags[tag] = true
		}
	}
	fmt.Println("Total tags:", len(allTags))*/
}

func profiling() {
	http.ListenAndServe(":1313", nil)
}

func findAVertical() int {
	for i, photo := range photos {
		if photo.Orientation == Vertical {
			return i
		}
	}
	return -1
}

func removeAVertical() *Photo {
	i := findAVertical()
	if i == -1 {
		return nil
	}
	p := removePhoto(i)
	return &p
}

func mergeTags(a, b *Photo) []int {
	tags := append(b.Tags, a.Tags...)
	sort.Ints(tags)
	for j := 0; j < len(tags) - 1; j++ {
		if tags[j] == tags[j + 1] {
			tags = append(tags[:j], tags[j+1:]...)
			j--
		}
	}
	return tags
}

func verticalScore(a, b *Photo) int {
	x, y, common := tagsInBoth(a, b)
	return x + y + common
}

func makeANiceVerticalSlide() bool {
	a := removeAVertical()
	if a == nil {
		return false
	}

	var bestIndex int
	if len(photos) > numWorkers * 100 {
		bestIndex, _ = findBestScore2(a)
	} else {
		bestIndex, _ = chooseNextNode2(a)
	}

	if bestIndex == -1 {
		return false
	}

	photos[bestIndex].SecondId = a.Id
	photos[bestIndex].Orientation = TwoVertical
	photos[bestIndex].Tags = mergeTags(a, &photos[bestIndex])

	return true
}

func processVerticals() {
	cnt := 0

	for makeANiceVerticalSlide() {
		cnt++

		if cnt % 100 == 0 {
			fmt.Println(cnt)
		}
	}
}

func main() {
	go profiling()

	work = make(chan Work, 8)
	results = make(chan WorkResult, 8)
	work2 = make(chan Work, 8)
	results2 = make(chan WorkResult, 8)
	for i := 0; i < numWorkers; i++ {
		go scoreWorker()
		go score2Worker()
	}

	inputIndex := 3
	photos = parseInput(inputIndex)
	of, _ := os.OpenFile(inputFile[inputIndex] + ".out", os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0644)
	defer of.Close()
	output := bufio.NewWriter(of)
	defer output.Flush()

	inputStats()

	processVerticals()

	_, _ = fmt.Fprintln(output, len(photos))

	totalScore := 0
	currentPhoto := removePhoto(0)
	fmt.Println(currentPhoto)

	slides := []string{currentPhoto.Slide()}

	for len(photos) > 0 {
		if len(photos) % 100 == 0 {
			fmt.Println("Remaining:", len(photos))
		}

		var next, score int
		if len(photos) > numWorkers * 100 {
			next, score = findBestScore(&currentPhoto)
		} else {
			next, score = chooseNextNode(&currentPhoto)
		}
		totalScore += score

		slides = append(slides, photos[next].Slide())
		currentPhoto = removePhoto(next)
	}

	for _, slide := range slides {
		_, _ = fmt.Fprintln(output, slide)
	}

	fmt.Println("Total score =", totalScore)
}
