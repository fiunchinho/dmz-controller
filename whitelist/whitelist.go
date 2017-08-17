package whitelist

import (
	"strings"
	"net"
	"github.com/golang/glog"
)

// Whitelist contains a list of ips to allow connections from
type Whitelist struct {
	Ips []string
}

func NewWhitelistFromString(ipsAsString string) (*Whitelist) {
	whitelist := new(Whitelist)
	whitelist.Add(strings.Split(ipsAsString, ","))

	return whitelist
}

func NewWhitelistFromArray(ips []string) (*Whitelist) {
	whitelist := new(Whitelist)
	whitelist.Add(ips)

	return whitelist
}

func NewEmptyWhitelist() (*Whitelist) {
	return new(Whitelist)
}

// Merge merges two lists together
func (whitelist *Whitelist) Merge(anotherWhitelist *Whitelist) {
	whitelist.Add(anotherWhitelist.Ips)
}

// Add adds ips to the current Whitelist
func (whitelist *Whitelist) Add(sourceWhitelist []string) {
	whitelist.Ips = RemoveDuplicates(append(whitelist.Ips, validateIPs(sourceWhitelist)...))
}

// Minus removes IP's from given Whitelist from current Whitelist
func (whitelist *Whitelist) Minus(substractedWhitelist *Whitelist) {
	for i := len(substractedWhitelist.Ips) - 1; i >= 0; i-- {
		index := findValue(whitelist.Ips, substractedWhitelist.Ips[i])
		if index > -1 {
			whitelist.Ips = append(whitelist.Ips[:index], whitelist.Ips[index+1:]...)
		}
	}
}

func (whitelist *Whitelist) ToString() string {
	return strings.Join(whitelist.Ips, ",")
}

// validateIPs makes sure all the IP's that we want to add are valid IPs or valid CIDR
func validateIPs(sourceWhitelist []string) []string {
	for i := len(sourceWhitelist) - 1; i >= 0; i-- {
		_, _, err := net.ParseCIDR(sourceWhitelist[i])
		if err != nil {
			sourceWhitelist[i] += "/32"
			_, _, err = net.ParseCIDR(sourceWhitelist[i])
		}

		if err != nil {
			glog.Info(err)
			glog.Infof("The following IP won't be added to the whitelist: %s", sourceWhitelist[i])
			sourceWhitelist = append(sourceWhitelist[:i], sourceWhitelist[i+1:]...)
		}
	}
	return sourceWhitelist
}

// RemoveDuplicates removes duplicate elements from array
func RemoveDuplicates(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

// findValue returns the slice key where the value is stored
func findValue(slice []string, valueToFind string) int {
	for key, v := range slice {
		if v == valueToFind {
			return key
		}
	}
	return -1
}
