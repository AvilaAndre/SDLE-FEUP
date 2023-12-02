package utils

import (
	"fmt"
	"log"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatal("Fatal error ", err)
	}
}

func Int64ToString(number int64) string {
	return fmt.Sprintf("%d", number)
}

func Float64ToString(number float64) string {
	return fmt.Sprintf("%f", number)
}

type Stack[T any] struct {
	values []T
}

func (s *Stack[T]) New() {
	s.values = make([]T, 0)
}

func (s *Stack[T]) Push(newValue T) {
	s.values = append(s.values, newValue)
}

func (s *Stack[T]) Pop() T {
	var value T = s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]

	return value
}

func (s *Stack[T]) Peek() T {
	return s.values[len(s.values)-1]
}

func (s *Stack[T]) Size() int {
	return len(s.values)
}
