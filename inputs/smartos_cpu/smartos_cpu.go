package smartos_cpu

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/m1cr0man/smartos-telegraf-plugins/helpers"
	"github.com/siebenmann/go-kstat"
)

type SmartOSCPU struct {
	Global   bool
	TagAlias bool

	last map[string]map[string]uint64
}

func (s *SmartOSCPU) Description() string {
	return "Collect global and zone cpu usage"
}

func (s *SmartOSCPU) SampleConfig() string {
	return `
  ## Tag metrics with zone alias?
  tag_alias = false
`
}

func (s *SmartOSCPU) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		return err
	}

	defer token.Close()
	stats := token.All()

	// Idle CPU will be the same across all zones
	idle := helpers.SumUint(stats, "cpu", -1, "sys", "cpu_nsec_idle") - s.last["global"]["idle"]

	for _, zone_sample := range helpers.FilterStats(stats, "zones", -1, "") {
		zone_name := helpers.ReadString(zone_sample, "zonename")

		metrics := map[string]uint64{
			"nsec_idle": idle,
		}
		fields := map[string]interface{}{
			"nsec_idle": idle - s.last[zone_name]["nsec_idle"],
		}

		for _, field := range []string{"nsec_user", "nsec_sys", "nsec_waitrq"} {
			val := helpers.ReadUint(zone_sample, field)
			metrics[field] = val
			fields[field] = val - s.last[zone_name][field]
		}

		acc.AddFields("cpu", fields, map[string]string{
			"zone": zone_name,
		})

		if s.last == nil {
			s.last = map[string]map[string]uint64{}
		}

		s.last[zone_name] = metrics

	}

	return nil
}

func init() {
	inputs.Add("smartos_cpu", func() telegraf.Input { return &SmartOSCPU{} })
}
