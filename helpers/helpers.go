package helpers

import (
	"log"
	"os/exec"
	s "strings"

	"github.com/siebenmann/go-kstat"
)

func ReadUint(stats *kstat.KStat, name string) uint64 {
	val, err := stats.GetNamed(name)

	if err != nil {
		log.Fatal("Cannot read ", name, " from stats: ", err)
		return 0
	}

	return val.UintVal
}

func ReadInt(stats *kstat.KStat, name string) int64 {
	val, err := stats.GetNamed(name)

	if err != nil {
		log.Fatal("Cannot read ", name, " from stats: ", err)
		return 0
	}

	return val.IntVal
}

func ReadString(stats *kstat.KStat, name string) string {
	val, err := stats.GetNamed(name)

	if err != nil {
		log.Fatal("Cannot read ", name, " from stats: ", err)
		return ""
	}

	return val.StringVal
}

func match(stat *kstat.KStat, module string, instance int, name string) bool {
	return (module == "" || stat.Module == module) &&
		(instance == -1 || stat.Instance == instance) &&
		(name == "" || stat.Name == name)
}

func FilterStats(stats []*kstat.KStat, module string, instance int, name string) (filtered []*kstat.KStat) {
	for _, stat := range stats {
		if match(stat, module, instance, name) {
			filtered = append(filtered, stat)
		}
	}
	return
}

func FindStat(stats []*kstat.KStat, module string, instance int, name string) *kstat.KStat {
	for _, stat := range stats {
		if match(stat, module, instance, name) {
			return stat
		}
	}
	return nil
}

func SumUint(stats []*kstat.KStat, module string, instance int, name string, val_name string) (val uint64) {
	for _, stat := range stats {
		if match(stat, module, instance, name) {
			val += ReadUint(stat, val_name)
		}
	}
	return
}

func SumInt(stats []*kstat.KStat, module string, instance int, name string, val_name string) (val int64) {
	for _, stat := range stats {
		if match(stat, module, instance, name) {
			val += ReadInt(stat, val_name)
		}
	}
	return
}

func GetZones() ([]string, error) {
	out, err := exec.Command("zoneadm", "list").Output()
	if err != nil {
		return nil, err
	}
	return s.Split(s.Trim(string(out), "\n"), "\n"), nil
}
