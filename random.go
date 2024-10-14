package main

import (
	"crypto/rand"
	"math/big"
)

func generateRandomNumbersWithStats(min, max, count int, unique bool) ([]int, map[int]int, error) {
	numbers := make([]int, 0, count)
	stats := make(map[int]int)
	rangeSize := max - min + 1

	if unique {
		numSet := make(map[int]struct{})
		for len(numSet) < count {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(rangeSize)))
			if err != nil {
				return nil, nil, err
			}
			value := min + int(num.Int64())
			if _, exists := numSet[value]; !exists {
				numSet[value] = struct{}{}
				numbers = append(numbers, value)
				stats[value]++
			}
		}
	} else {
		for i := 0; i < count; i++ {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(rangeSize)))
			if err != nil {
				return nil, nil, err
			}
			value := min + int(num.Int64())
			numbers = append(numbers, value)
			stats[value]++
		}
	}

	return numbers, stats, nil
}

func generateRandomFloats(min, max float64, count int) ([]float64, error) {
	numbers := make([]float64, count)
	rangeSize := max - min

	for i := 0; i < count; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(1e9))
		if err != nil {
			return nil, err
		}
		fraction := float64(num.Int64()) / 1e9

		numbers[i] = min + fraction*rangeSize
	}

	return numbers, nil
}
