package whitelist

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestWhiteListFromString(t *testing.T) {
	whitelist := NewWhitelistFromString("1.2.3.4,8.8.8.0/28")
	assert := assert.New(t)
	assert.Contains(whitelist.Ips, "1.2.3.4/32", "Initial ip's must appear")
	assert.Contains(whitelist.Ips, "8.8.8.0/28", "Initial ip's must appear")
}

func TestWhiteListFromArray(t *testing.T) {
	whitelist := NewWhitelistFromArray([]string{"8.8.8.8", "1.2.3.4"})
	assert := assert.New(t)
	assert.Contains(whitelist.Ips, "1.2.3.4/32", "Initial ip's must appear")
	assert.Contains(whitelist.Ips, "8.8.8.8/32", "Initial ip's must appear")
}

func TestEmptyWhiteList(t *testing.T) {
	whitelist := NewEmptyWhitelist()
	assert := assert.New(t)
	assert.Len(whitelist.Ips, 0, "Empty whitelist must contain 0 elements")
}

func TestWhiteListRemoveDuplicates(t *testing.T) {
	whitelist := NewWhitelistFromString("1.2.3.0/28,8.8.8.0/28")
	whitelist.Add([]string{"5.5.5.5", "8.8.8.0/28", "1.2.3.4"})
	assert := assert.New(t)
	assert.Len(whitelist.Ips, 4, "It must contain no duplicates")
	assert.Contains(whitelist.Ips, "1.2.3.4/32", "Added ip's must appear")
	assert.Contains(whitelist.Ips, "1.2.3.0/28", "Added ip's must appear")
	assert.Contains(whitelist.Ips, "8.8.8.0/28", "Added ip's must appear")
	assert.Contains(whitelist.Ips, "5.5.5.5/32", "Added ip's must appear")
}

func TestMergingWhitelists(t *testing.T) {
	whitelist1 := NewWhitelistFromString("1.2.3.4")
	whitelist2 := NewWhitelistFromString("8.8.8.0/28")
	whitelist1.Merge(whitelist2)

	assert := assert.New(t)
	assert.Contains(whitelist1.Ips, "1.2.3.4/32", "There is one missing IP from merged Whitelists")
	assert.Contains(whitelist1.Ips, "8.8.8.0/28", "There is one missing IP from merged Whitelists")
}

func TestThatItIgnoresAnInvalidIp(t *testing.T) {
	whitelist := NewWhitelistFromString("1.2.3.4")
	whitelist.Add([]string{"5.5.5", "8.8.8.8", "non-valid.3.4"})
	assert := assert.New(t)
	assert.Contains(whitelist.Ips, "1.2.3.4/32", "There is one missing IP from merged Whitelists")
	assert.Contains(whitelist.Ips, "8.8.8.8/32", "There is one missing IP from merged Whitelists")
	assert.Len(whitelist.Ips, 2, "It must ignore non valid IP addresses")
}

func TestThatItRemovesItemsFromSecondWhitelist(t *testing.T) {
	whitelist := NewWhitelistFromString("1.2.3.4/32,4.4.4.4,8.8.8.8")
	whitelist.Minus(NewWhitelistFromString("4.4.4.4/32,3.3.3.3/28"))
	assert := assert.New(t)
	assert.Contains(whitelist.Ips, "1.2.3.4/32", "It should only substract IPs from second whitelist")
	assert.Contains(whitelist.Ips, "8.8.8.8/32", "It should only substract IPs from second whitelist")
	assert.Len(whitelist.Ips, 2, "It must remove IP from the substracted whitelist")
}

func TestThatTransformWhitelistToString(t *testing.T) {
	whitelist := NewWhitelistFromString("1.2.3.4/32,4.4.4.4,8.8.8.8")
	assert := assert.New(t)
	assert.Equal("1.2.3.4/32,4.4.4.4/32,8.8.8.8/32", whitelist.ToString(), "String representation is wrong")
}
