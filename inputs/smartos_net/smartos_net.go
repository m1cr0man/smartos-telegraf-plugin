package smartos_net

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/m1cr0man/smartos-telegraf-plugins/helpers"
	"github.com/siebenmann/go-kstat"
)

type SmartOSNet struct {
	Global   bool
	TagAlias bool
	pgsize   uint64

	last map[string]map[string]uint64
}

func (s *SmartOSNet) Description() string {
	return "Collect global and zone net usage"
}

func (s *SmartOSNet) SampleConfig() string {
	return `
  ## Tag metrics with zone alias?
  tag_alias = false
`
}

func (s *SmartOSNet) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		return err
	}

	defer token.Close()
	stats := token.All()

	for _, linkstat := range helpers.FilterStats(stats, "link", -1, "") {
		zoneName := helpers.ReadString(linkstat, "zonename")

		metrics := map[string]uint64{}
		fields := map[string]interface{}{}

		for _, field := range []string{"obytes64", "rbytes64", "collisions", "ierrors", "oerrors"} {
			val := helpers.ReadUint(linkstat, field)
			metrics[field] = val
			fields[field] = val - s.last[linkstat.Name][field]
		}

		acc.AddFields("net", fields, map[string]string{
			"zone":      zoneName,
			"interface": linkstat.Name,
		})

		s.last[linkstat.Name] = metrics
	}

	return nil
}

func init() {
	inputs.Add("smartos_net", func() telegraf.Input { return &SmartOSNet{last: map[string]map[string]uint64{}} })
}
