package smartos_ram

import (
	"log"
	"os/exec"
	"regexp"
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
	swapsize uint64

	last map[string]uint64
}

func (s *SmartOSRAM) Description() string {
	return "Collect global and zone ram usage"
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
	arcstats := helpers.FindStat(stats, "zfs", -1, "arcstats")
	vmstats := helpers.FilterStats(stats, "cpu", -1, "vm")

	totalPhysical := helpers.ReadUint(pagestats, "physmem") * s.pgsize
	pagesFree := helpers.ReadUint(pagestats, "pagesfree")
	globalFields := map[string]interface{}{
		// Free page scan rate
		"nscan": helpers.ReadUint(pagestats, "nscan"),
		// Pages paged by kernel
		"pp_kernel":     helpers.ReadUint(pagestats, "pp_kernel"),
		"pagesfree":     pagesFree,
		"pagesfree_pct": 100.0 * pagesFree / helpers.ReadUint(pagestats, "pagestotal"),
		// Pages paged in/out
		"pgpgin":  helpers.SumUint(vmstats, "", -1, "", "pgpgin"),
		"pgpgout": helpers.SumUint(vmstats, "", -1, "", "pgpgout"),
		// Pages swapped in/out
		"pgswapin":  helpers.SumUint(vmstats, "", -1, "", "pgswapin"),
		"pgswapout": helpers.SumUint(vmstats, "", -1, "", "pgswapout"),
		// Zone overcommit scan activity
		"zone_cap_scan": helpers.ReadUint(pagestats, "zone_cap_scan"),
		// ARC size
		"arcsize":            helpers.ReadUint(arcstats, "size"),
		"arcsize_compressed": helpers.ReadUint(arcstats, "compressed_size"),
	}

	// Get diff for fault metrics
	// Fault types: Major, Copy on write, Kernel anon, Protection fault, !(cow|prot) AKA anon
	for _, field := range []string{"maj_fault", "cow_fault", "kernel_asflt", "prot_fault", "as_fault"} {
		val := helpers.SumUint(vmstats, "", -1, "", field)
		globalFields[field] = val - s.last[field]
		s.last[field] = val
	}

	for _, zone_sample := range helpers.FilterStats(stats, "memory_cap", -1, "") {
		zoneName := helpers.ReadString(zone_sample, "zonename")
		fields := map[string]interface{}{}
		rss := helpers.ReadUint(zone_sample, "rss")
		swap := helpers.ReadUint(zone_sample, "swap")
		swapcap := helpers.ReadUint(zone_sample, "swapcap")
		total := helpers.ReadUint(zone_sample, "physcap")
		available := total - rss

		if zoneName == "global" {
			fields = globalFields
			swapcap = s.swapsize
			total = totalPhysical
		}

		// Number of times over allocated memory
		fields["nover"] = helpers.ReadUint(zone_sample, "nover")
		fields["rss"] = rss
		fields["total"] = total
		fields["available"] = available
		fields["available_pct"] = 100.0 * available / total
		fields["swap"] = swap
		fields["swap_total"] = swapcap
		fields["swap_percent"] = 100.0 * swap / swapcap

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
	pgsizeRaw, _ := strconv.Atoi(strings.Trim(string(out), "\n"))
	pgsize := uint64(pgsizeRaw)

	// Swap size isn't going to change much either
	out, err = exec.Command("/usr/sbin/swap", "-l").Output()
	if err != nil {
		log.Fatal("Failed to get swap size: ", err)
	}
	re := regexp.MustCompile(`(\d{4,})\s+\d+[^\d]*$`)
	match := re.FindAllStringSubmatch(string(out), -1)[0]
	swapsizeRaw, _ := strconv.Atoi(match[1])
	swapsize := uint64(swapsizeRaw) * 512

	inputs.Add("smartos_ram", func() telegraf.Input {
		return &SmartOSRAM{pgsize: pgsize, swapsize: swapsize, last: map[string]uint64{}}
	})
}
