package helpers

import (
	"log"

	"github.com/siebenmann/go-kstat"
)

func ReadStat(stats *kstat.KStat, name string) (string, interface{}) {
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
