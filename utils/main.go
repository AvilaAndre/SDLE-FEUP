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
