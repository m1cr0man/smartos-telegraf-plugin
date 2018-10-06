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

	// Note: Idle CPU in global zone accounts for all idle CPU

	globalMetrics := map[string]uint64{}
	globalFields := map[string]interface{}{"cpu_nsec_total": 0}
	var total uint64

	cpuSamples := helpers.FilterStats(stats, "cpu", -1, "sys")
	for _, field := range []string{"cpu_nsec_idle", "cpu_nsec_intr", "cpu_nsec_kernel", "cpu_nsec_user"} {
		val := helpers.SumUint(cpuSamples, "cpu", -1, "sys", field)
		globalMetrics[field] = val
		globalFields[field] = val - s.last["global"][field]
		total += val - s.last["global"][field]
	}
	globalFields["cpu_nsec_total"] = total

	for _, zoneSample := range helpers.FilterStats(stats, "zones", -1, "") {
		zoneName := helpers.ReadString(zoneSample, "zonename")

		metrics := map[string]uint64{}
		fields := map[string]interface{}{}

		if zoneName == "global" {
			metrics = globalMetrics
			fields = globalFields
		}

		for _, field := range []string{"nsec_user", "nsec_sys", "nsec_waitrq"} {
			val := helpers.ReadUint(zoneSample, field)
			metrics[field] = val
			fields[field] = val - s.last[zoneName][field]
		}

		acc.AddFields("cpu", fields, map[string]string{
			"zone": zoneName,
		})

		s.last[zoneName] = metrics
	}

	return nil
}

func init() {
	inputs.Add("smartos_cpu", func() telegraf.Input { return &SmartOSCPU{last: map[string]map[string]uint64{}} })
}
