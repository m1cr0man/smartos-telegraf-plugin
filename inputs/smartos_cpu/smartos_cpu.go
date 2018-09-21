package smartos_cpu

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/m1cr0man/smartos-telegraf-plugins/helpers"
	"github.com/siebenmann/go-kstat"
)

type SmartOSCPU struct {
	Global      bool
	UUIDToAlias bool
}

func (s *SmartOSCPU) Description() string {
	return "Collect global and zone cpu usage"
}

func (s *SmartOSCPU) SampleConfig() string {
	return `
  ## Collect global zone CPU stats?
  global = true

  ## Translate UUIDs into zone aliases?
  uuid_to_alias = true
`
}

func (s *SmartOSCPU) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		log.Fatal("cannot get kstat token")
		return err
	}

	stats, _ := token.Lookup("zones", -1, "3ff6c31f-8863-6805-dc0d-d35adf")
	name, val := helpers.ReadStat(stats, "nsec_user")

	acc.AddFields("cpu.nsec_user", map[string]interface{}{
		name: val
	}, map[string]string {
		"zone": "3ff6c31f-8863-6805-dc0d-d35adf"
	})

	token.Close()

	return nil
}

func init() {
	inputs.Add("SmartOSCPU", func() telegraf.Input { return &SmartOSCPU{} })
}
