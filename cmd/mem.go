package cmd

import (
	"bufio"
	"os"
	"regexp"
	"runtime"
	"strconv"
)

var fallbackAvailableMemory int64 = 1 * 1024 * 1024 * 1024 // 1GB
var memAvailableRE = regexp.MustCompile("MemAvailable:\\s*([[:digit:]]+)\\s*kB")

// GetAvailableMemory returns an estimation how much memory is available to
// applications without swapping. Currently only Linux is supported. For all
// other operating systems the fallbackAvailableMemory value is returned and no
// error. In case of errors fallbackAvailableMemory is also returned with the
// error.
func GetAvailableMemory() (int64, error) {
	if runtime.GOOS == "linux" {
		file, err := os.Open("/proc/meminfo")
		if err != nil {
			return fallbackAvailableMemory, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if match := memAvailableRE.FindStringSubmatch(line); match != nil {
				retVal, err := strconv.ParseInt(match[1], 10, 64)
				if err != nil {
					return fallbackAvailableMemory, err
				}
				return retVal * 1024, nil
			}
		}

		// no answer yet, fallback
	}

	return fallbackAvailableMemory, nil
}
