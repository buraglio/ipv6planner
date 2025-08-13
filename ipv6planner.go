package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net"
	"os"
	"strconv"
	"strings"
)

type IPv6Plan struct {
	BaseSubnet     string        `json:"base_subnet"`
	POPCount       int           `json:"pop_count"`
	PreferredSize  int           `json:"preferred_size"`
	SubnetLevels   []int         `json:"subnet_levels"`
	POPAllocations []POPAlloc    `json:"pop_allocations"`
	SubnetCounts   []SubnetCount `json:"subnet_counts"`
}

type POPAlloc struct {
	POPNumber  int            `json:"pop_number"`
	POPSubnet  string         `json:"pop_subnet"`
	Subnets    []SubnetDetail `json:"subnets"`
	LevelNames []string       `json:"level_names"`
}

type SubnetDetail struct {
	CIDR      string `json:"cidr"`
	Count     int64  `json:"count"`
	Available int64  `json:"available"`
}

type SubnetCount struct {
	PrefixSize int   `json:"prefix_size"`
	Count      int64 `json:"count"`
	Available  int64 `json:"available"`
}

func main() {
	// Default values
	subnet := "3fff::/20"
	popCount := 5
	preferredSize := 36
	subnetLevelsStr := "44,48,64"
	outputFormat := "text"
	interactive := false
	showHelp := false

	// Parse flags
	flag.StringVar(&subnet, "s", subnet, "Base IPv6 subnet (e.g., 3fff::/20)")
	flag.IntVar(&popCount, "n", popCount, "Number of POPs")
	flag.IntVar(&preferredSize, "p", preferredSize, "Preferred subnet size per POP")
	flag.StringVar(&subnetLevelsStr, "l", subnetLevelsStr, "Comma-separated list of subnet levels")
	flag.BoolVar(&interactive, "i", interactive, "Interactive mode")
	flag.BoolVar(&showHelp, "h", showHelp, "Show help information")

	// Output format flags
	jsonFlag := flag.Bool("j", false, "JSON output format")
	htmlFlag := flag.Bool("k", false, "HTML output format")
	textFlag := flag.Bool("t", false, "Text output format (default)")

	flag.Parse()

	// Handle output format
	if *jsonFlag {
		outputFormat = "json"
	} else if *htmlFlag {
		outputFormat = "html"
	} else if *textFlag {
		outputFormat = "text"
	}

	// Handle -h flag
	if showHelp {
		printHelp()
		return
	}

	// Parse subnet levels
	subnetLevels := parseSubnetLevels(subnetLevelsStr)

	if interactive {
		subnet, popCount, preferredSize, subnetLevels = getInteractiveInput()
	}

	plan := generateIPv6Plan(subnet, popCount, preferredSize, subnetLevels)

	switch outputFormat {
	case "json":
		outputJSON(plan)
	case "html":
		outputHTML(plan)
	default:
		outputText(plan)
	}
}

func parseSubnetLevels(levelsStr string) []int {
	levels := strings.Split(levelsStr, ",")
	subnetLevels := make([]int, len(levels))
	for i, l := range levels {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "/") {
			l = l[1:]
		}
		subnetLevels[i], _ = strconv.Atoi(l)
	}
	return subnetLevels
}

func printHelp() {
	fmt.Println(`IPv6 Address Planner - Help
Usage: ipv6planner [flags]

Flags:
  -s string    Base IPv6 subnet (default "3fff::/20")
  -n int       Number of POPs (default 5)
  -p int       Preferred subnet size per POP (default 36)
  -l string    Comma-separated list of subnet levels (default "44,48,64")
  -t           Text output format (default)
  -j           JSON output format
  -k           HTML output format
  -i           Interactive mode
  -h           Show this help message

Examples:
  Basic usage with defaults:
    ipv6planner

  Custom parameters with JSON output:
    ipv6planner -s 2001:db8::/32 -n 10 -p 40 -l 48,52,56,64 -j

  Interactive mode:
    ipv6planner -i

  HTML output:
    ipv6planner -k`)
}

func getInteractiveInput() (string, int, int, []int) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter base IPv6 subnet (default 3fff::/20): ")
	subnet, _ := reader.ReadString('\n')
	subnet = strings.TrimSpace(subnet)
	if subnet == "" {
		subnet = "3fff::/20"
	}

	fmt.Print("Enter number of POPs (default 5): ")
	popStr, _ := reader.ReadString('\n')
	popStr = strings.TrimSpace(popStr)
	popCount := 5
	if popStr != "" {
		popCount, _ = strconv.Atoi(popStr)
	}

	fmt.Print("Enter preferred subnet size per POP (default /36): ")
	sizeStr, _ := reader.ReadString('\n')
	sizeStr = strings.TrimSpace(sizeStr)
	preferredSize := 36
	if sizeStr != "" {
		if strings.HasPrefix(sizeStr, "/") {
			sizeStr = sizeStr[1:]
		}
		preferredSize, _ = strconv.Atoi(sizeStr)
	}

	fmt.Print("Enter subnet levels (comma separated, default 44,48,64): ")
	levelsStr, _ := reader.ReadString('\n')
	levelsStr = strings.TrimSpace(levelsStr)
	subnetLevels := []int{44, 48, 64}
	if levelsStr != "" {
		subnetLevels = parseSubnetLevels(levelsStr)
	}

	return subnet, popCount, preferredSize, subnetLevels
}

func calculateAvailableSubnets(parentSize, childSize int) int64 {
	if childSize <= parentSize {
		return 0
	}
	return int64(1) << uint(childSize-parentSize)
}

func generateIPv6Plan(subnet string, popCount, preferredSize int, subnetLevels []int) IPv6Plan {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		fmt.Printf("Error parsing subnet: %v\n", err)
		os.Exit(1)
	}

	ones, _ := ipNet.Mask.Size()

	plan := IPv6Plan{
		BaseSubnet:    subnet,
		POPCount:      popCount,
		PreferredSize: preferredSize,
		SubnetLevels:  subnetLevels,
	}

	// Calculate subnet counts for each level
	for _, level := range subnetLevels {
		if level <= ones {
			continue
		}
		count := calculateAvailableSubnets(ones, level)
		plan.SubnetCounts = append(plan.SubnetCounts, SubnetCount{
			PrefixSize: level,
			Count:      count,
			Available:  count,
		})
	}

	// Calculate how many bits we need for POP allocation
	bitsNeeded := 0
	for (1 << bitsNeeded) < popCount {
		bitsNeeded++
	}

	// Calculate the new prefix length for POP allocations
	newPrefixLen := ones + bitsNeeded
	if newPrefixLen > preferredSize {
		fmt.Printf("Warning: Required prefix length %d is larger than preferred size %d\n", newPrefixLen, preferredSize)
	}

	// Generate POP allocations
	for i := 0; i < popCount; i++ {
		popIP := make(net.IP, len(ipNet.IP))
		copy(popIP, ipNet.IP)

		// Set the POP bits
		for bit := 0; bit < bitsNeeded; bit++ {
			byteIndex := (ones + bit) / 8
			bitIndex := 7 - (ones+bit)%8
			if (i>>bit)&1 == 1 {
				popIP[byteIndex] |= 1 << bitIndex
			}
		}

		// Create the POP subnet
		popSubnet := &net.IPNet{
			IP:   popIP,
			Mask: net.CIDRMask(preferredSize, 128),
		}

		// Generate subnets for this POP
		var subnets []SubnetDetail
		levelNames := make([]string, len(subnetLevels))

		for j, level := range subnetLevels {
			if level <= preferredSize {
				fmt.Printf("Warning: Subnet level %d is not more specific than POP size %d\n", level, preferredSize)
				continue
			}

			// Calculate available subnets at this level
			available := calculateAvailableSubnets(preferredSize, level)

			// For demonstration, we'll just show the first subnet at each level
			subnetIP := make(net.IP, len(popIP))
			copy(subnetIP, popIP)
			subnet := &net.IPNet{IP: subnetIP, Mask: net.CIDRMask(level, 128)}

			subnets = append(subnets, SubnetDetail{
				CIDR:      subnet.String(),
				Count:     available,
				Available: available,
			})
			levelNames[j] = fmt.Sprintf("Level %d (/%d)", j+1, level)
		}

		plan.POPAllocations = append(plan.POPAllocations, POPAlloc{
			POPNumber:  i + 1,
			POPSubnet:  popSubnet.String(),
			Subnets:    subnets,
			LevelNames: levelNames,
		})
	}

	return plan
}

func outputText(plan IPv6Plan) {
	fmt.Printf("This tool is not intended to provide a comprehensive address plan.\n")
	fmt.Printf("It should be used to generate a top level hierarchy of IPv6 address plans.\n")
	fmt.Printf("IPv6 Address Plan\n")
	fmt.Printf("Base Subnet: %s\n", plan.BaseSubnet)
	fmt.Printf("Number of POPs: %d\n", plan.POPCount)
	fmt.Printf("Preferred POP subnet size: /%d\n", plan.PreferredSize)
	fmt.Printf("Subnet levels: /%v\n", plan.SubnetLevels)

	fmt.Println("\nGlobal Subnet Counts:")
	for _, count := range plan.SubnetCounts {
		fmt.Printf("  /%d: %d available subnets\n", count.PrefixSize, count.Available)
	}

	fmt.Println("\nPOP Allocations:")
	for _, pop := range plan.POPAllocations {
		fmt.Printf("\nPOP %d: %s\n", pop.POPNumber, pop.POPSubnet)
		for i, subnet := range pop.Subnets {
			fmt.Printf("  %s: %s (Available: %d)\n", pop.LevelNames[i], subnet.CIDR, subnet.Available)
		}
	}
}

func outputJSON(plan IPv6Plan) {
	jsonData, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		fmt.Printf("Error generating JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

func outputHTML(plan IPv6Plan) {
	const tpl = `
<!DOCTYPE html>
<html>
<head>
    <title>IPv6 Address Plan</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin-bottom: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .pop { margin-bottom: 30px; }
        .pop-header { background-color: #e6f7ff; padding: 10px; margin-bottom: 10px; }
        .count { color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <h1>IPv6 Address Plan</h1>
    <table>
        <tr><th>Base Subnet</th><td>{{.BaseSubnet}}</td></tr>
        <tr><th>Number of POPs</th><td>{{.POPCount}}</td></tr>
        <tr><th>Preferred POP subnet size</th><td>/{{.PreferredSize}}</td></tr>
        <tr><th>Subnet levels</th><td>{{range .SubnetLevels}}/{{.}} {{end}}</td></tr>
    </table>

    <h2>Global Subnet Counts</h2>
    <table>
        <tr>
            <th>Prefix Size</th>
            <th>Available Subnets</th>
        </tr>
        {{range .SubnetCounts}}
        <tr>
            <td>/{{.PrefixSize}}</td>
            <td>{{.Available}}</td>
        </tr>
        {{end}}
    </table>

    <h2>POP Allocations</h2>
    {{range .POPAllocations}}
    <div class="pop">
        <div class="pop-header">
            <strong>POP {{.POPNumber}}:</strong> {{.POPSubnet}}
        </div>
        <table>
            <tr>
                <th>Level</th>
                <th>Subnet</th>
                <th>Available</th>
            </tr>
            {{range $index, $subnet := .Subnets}}
            <tr>
                <td>{{index $.POPAllocations $index "LevelNames"}}</td>
                <td>{{$subnet.CIDR}}</td>
                <td>{{$subnet.Available}}</td>
            </tr>
            {{end}}
        </table>
    </div>
    {{end}}
</body>
</html>
`

	tmpl, err := template.New("plan").Parse(tpl)
	if err != nil {
		fmt.Printf("Error creating HTML template: %v\n", err)
		os.Exit(1)
	}

	err = tmpl.Execute(os.Stdout, plan)
	if err != nil {
		fmt.Printf("Error generating HTML: %v\n", err)
		os.Exit(1)
	}
}
