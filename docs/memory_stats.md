# Solaris Memory stats

Solaris does not run the paging daemon all the time. There are a couple of tunables that control it.

`lotsfree` is the upper threshold for free memory which determines when the daemon does not need to run.
Scanning begins at `slowscan`/second

`desfree` determines when the OS will scan for unused pages to be moved to swap.
Scanning begins at `desscan`/second

`minfree` determines when the system should get into a wee panic about available memory
Scanning begins at `fastscan`/second

Scan speed scales linearly between these values. The current scan speed is available as `nscan`

`memory_cap` KStat module describes memory usage for each zone

`prstat -Z` shows a percentage for memory usage per zone. This is `rss/total_memory`, which isn't much use
