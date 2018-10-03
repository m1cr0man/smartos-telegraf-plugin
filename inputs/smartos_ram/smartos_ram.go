package smartos_ram

import (
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/m1cr0man/smartos-telegraf-plugins/helpers"
	"github.com/siebenmann/go-kstat"
)

type SmartOSRAM struct {
	Global   bool
	TagAlias bool
	pgsize   uint64

	last map[string]uint64
}

func (s *SmartOSRAM) Description() string {
	return "Collect global and zone cpu usage"
}

func (s *SmartOSRAM) SampleConfig() string {
	return `
  ## Tag metrics with zone alias?
  tag_alias = false
`
}

func (s *SmartOSRAM) Gather(acc telegraf.Accumulator) error {
	token, err := kstat.Open()
	if err != nil {
		return err
	}

	defer token.Close()
	stats := token.All()

	pagestats := helpers.FindStat(stats, "unix", -1, "system_pages")
	vmstats := helpers.FilterStats(stats, "cpu", -1, "vm")

	pagesFree := helpers.ReadUint(pagestats, "pagesfree") * s.pgsize
	globalFields := map[string]interface{}{
		// Free page scan rate
		"nscan": helpers.ReadUint(pagestats, "nscan"),
		// Pages paged by kernel
		"pp_kernel":     helpers.ReadUint(pagestats, "pp_kernel"),
		"pagesfree":     pagesFree,
		"pagesfree_pct": 100.0 * pagesFree / (helpers.ReadUint(pagestats, "pagestotal")),
		// Pages paged in/out
		"pgpgin":  helpers.SumUint(vmstats, "", -1, "", "pgpgin"),
		"pgpgout": helpers.SumUint(vmstats, "", -1, "", "pgpgout"),
		// Pages swapped in/out
		"pgswapin":  helpers.SumUint(vmstats, "", -1, "", "pgswapin"),
		"pgswapout": helpers.SumUint(vmstats, "", -1, "", "pgswapout"),
	}

	// Get diff for fault metrics
	// Fault types: Major, Copy on write, Kernel anon, Protection fault, !(cow|prot) AKA anon
	for _, field := range []string{"maj_fault", "cow_fault", "kernel_asflt", "prot_fault", "as_fault"} {
		val := helpers.SumUint(vmstats, "", -1, "", field)
		globalFields[field] = val - s.last["as_fault"]
		s.last["as_fault"] = val
	}

	for _, zone_sample := range helpers.FilterStats(stats, "memory_cap", -1, "") {
		zoneName := helpers.ReadString(zone_sample, "zonename")
		fields := map[string]interface{}{}

		if zoneName == "global" {
			fields = globalFields
		}

		rss := helpers.ReadUint(zone_sample, "rss")
		total := helpers.ReadUint(zone_sample, "physcap")
		available := total - rss

		// Number of times over allocated memory
		fields["nover"] = helpers.ReadUint(zone_sample, "nover")
		fields["rss"] = rss
		fields["total"] = total
		fields["available"] = available
		fields["available_pct"] = 100.0 * available / total
		fields["swap_percent"] = 100.0 * helpers.ReadUint(zone_sample, "swap") / helpers.ReadUint(zone_sample, "swapcap")

		// TODO arcstats
		acc.AddFields("ram", fields, map[string]string{
			"zone": zoneName,
		})
	}

	return nil
}

func init() {

	// Get pagesize now instead of every time we read stats
	// Realistically, it'll never change
	out, err := exec.Command("/usr/bin/pagesize").Output()
	if err != nil {
		log.Fatal("Failed to get RAM page size: ", err)
	}
	newval, _ := strconv.Atoi(strings.Trim(string(out), "\n"))
	pgsize := uint64(newval)

	inputs.Add("smartos_ram", func() telegraf.Input { return &SmartOSRAM{pgsize: pgsize, last: map[string]uint64{}} })
}
