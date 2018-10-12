package smartos_disk

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/m1cr0man/smartos-telegraf-plugins/helpers"
	"github.com/siebenmann/go-kstat"
)

type SmartOSDisk struct {
	Global   bool
	TagAlias bool
	pgsize   uint64
	swapsize uint64
}

func (s *SmartOSDisk) Description() string {
	return "Collect global and zone disk usage"
}

func (s *SmartOSDisk) SampleConfig() string {
	return `
  ## Tag metrics with zone alias?
  tag_alias = false
`
}

func (s *SmartOSDisk) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		return err
	}

	defer token.Close()
	stats := token.All()

	// Arcstats are all awesome, just upload the whole lot :)
	arcstats := helpers.FindStat(stats, "zfs", -1, "arcstats")
	arcfields := map[string]interface{}{}
	arcvals, _ := arcstats.AllNamed()
	for _, stat := range arcvals {
		arcfields[stat.Name] = stat.UintVal
	}
	acc.AddFields("arcstats", arcfields, map[string]string{
		"zone": "global",
	})

	// TODO figure out syntax for merging lists
	fsstats := append(helpers.FilterStats(stats, "zone_vfs", -1, ""), helpers.FilterStats(stats, "zone_zfs", -1, "")...)
	for _, stat := range fsstats {
		zoneName := helpers.ReadString(stat, "zonename")

		fields := map[string]interface{}{}
		vals, _ := stat.AllNamed()
		for _, stat := range vals {
			fields[stat.Name] = stat.UintVal
		}

		acc.AddFields("vfsstats", fields, map[string]string{
			"zone": zoneName,
		})
	}

	return nil
}

func init() {
	inputs.Add("smartos_disk", func() telegraf.Input {
		return &SmartOSDisk{}
	})
}
