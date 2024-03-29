package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	file, err := os.Open("test.tr")

	if err != nil {
		log.Fatalf("failed to open file")
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	c := NewLeCaR(1000, 0.45, 0.005, 0.5)

	for scanner.Scan() {
		s := scanner.Text()

		ints := strings.Split(s, " ")

		id := ints[1]
		size, _ := strconv.Atoi(ints[2])

		val := make([]byte, size)
		_, ok := c.Get(id)

		if !ok {
			c.Set(id, val)
		}
	}

	file.Close()

	fmt.Println(c.stats.toString())
	fmt.Println(c.WeightsToString())
}
