package main

import (
	"fmt"
	"log"

	"github.com/siebenmann/go-kstat"
)

func readStat(stats *kstat.KStat, name string) (string, interface{}) {
	val, err := stats.GetNamed(name)

	if err != nil {
		log.Fatal("Cannot read", name, "from stats")
		return "", nil
	}

	switch val.Type {
	case kstat.Uint32:
		return val.Name, val.UintVal
	case kstat.Uint64:
		return val.Name, val.UintVal

	case kstat.Int32:
		return val.Name, val.IntVal
	case kstat.Int64:
		return val.Name, val.IntVal

	case kstat.String:
		return val.Name, val.StringVal
	}

	return "", nil
}

func main() {
	token, err := kstat.Open()
	if err != nil {
		log.Fatal("cannot get kstat token")
	}
	stats, _ := token.Lookup("zones", -1, "3ff6c31f-8863-6805-dc0d-d35adf")
	name, val := readStat(stats, "nsec_user")
	fmt.Printf("%s: %d\n", name, val)
}
