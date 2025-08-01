
# IPv6 Address Planner

A command-line tool for generating hierarchical IPv6 address plans with subnet counts at each level.

## Features

- Generates IPv6 address plans from a base subnet
- Supports multiple Points of Presence (POPs)
- Creates hierarchical subnet levels within each POP
- Calculates and displays available subnet counts at each level
- Multiple output formats (text, JSON, HTML)
- Interactive mode for guided input
- Comprehensive help system

## Installation

### Prerequisites
- Go 1.16 or higher
- Git (optional)

### Installation Steps

1. **Download the tool**:
```
git clone https://github.com/yourusername/ipv6-planner.git
cd ipv6-planner
```

Build the executable:

```
go build ipv6planner.go
```


(Optional) Install system-wide:
```
sudo mv ipv6planner /usr/local/bin/
```

### Usage

Basic Command

```
./ipv6planner [flags]
Command Line Options
Flag	Description	Default Value	Example
-s	Base IPv6 subnet	3fff::/20	-s 3fff:db8::/32
-n	Number of POPs	5	-n 10
-p	Preferred subnet size per POP	36	-p 40
-l	Comma-separated subnet levels	44,48,64	-l 48,52,56,64
-t	Text output (default)	N/A	N/A
-j	JSON output	N/A	-j
-k	HTML output	N/A	-k
-i	Interactive mode	N/A	-i
-h	Show help	N/A	-h
```


### Examples

#### Basic Usage

```
./ipv6planner
```

#### Custom Configuration with JSON Output

```
./ipv6planner -s 3fff:db8::/32 -n 10 -p 40 -l 48,52,56,64 -j plan.json
```

#### Interactive Mode

```
./ipv6planner -i
```

#### HTML Output

```
./ipv6planner -s 3fff:db8::/32 -n 10 -p 40 -l 48,52,56,64 -k  plan.html
```

#### Output Formats

Text Output (Default)

```
Text Only
IPv6 Address Plan
Base Subnet: 3fff:db8::/32
Number of POPs: 5
Preferred POP subnet size: /40
Subnet levels: /48 /52 /56 /64 

Global Subnet Counts:
  /48: 65536 available subnets
  /52: 1048576 available subnets
  /56: 16777216 available subnets
  /64: 1099511627776 available subnets

POP Allocations:

POP 1: 3fff:db8::/40
  Level 1 (/48): 3fff:db8::/48 (Available: 256)
  Level 2 (/52): 3fff:db8::/52 (Available: 4096)
  Level 3 (/56): 3fff:db8::/56 (Available: 65536)
  Level 4 (/64): 3fff:db8::/64 (Available: 4294967296)
...

```

JSON Output

```
{
  base_subnet: 3fff:db8::/32,
  pop_count: 5,
  preferred_size: 40,
  subnet_levels: [48,52,56,64],
  subnet_counts: [
    {prefix_size:48,count:65536,available:65536},
    {prefix_size:52,count:1048576,available:1048576},
    {prefix_size:56,count:16777216,available:16777216},
    {prefix_size:64,count:1099511627776,available:1099511627776}
  ],
  pop_allocations: [
    {
      pop_number:1,
      pop_subnet:3fff:db8::/40,
      subnets:[
        {cidr:3fff:db8::/48,count:256,available:256},
        {cidr:3fff:db8::/52,count:4096,available:4096},
        {cidr:3fff:db8::/56,count:65536,available:65536},
        {cidr:3fff:db8::/64,count:4294967296,available:4294967296}
      ],
      level_names:[
        Level 1 (/48),
        Level 2 (/52),
        Level 3 (/56),
        Level 4 (/64)
      ]
    }
  ]
}

```

HTML Output

```
coming soon
```

Subnet Calculation Methodology
The tool calculates available subnets using the formula:

Text Only

Available Subnets = 2^(child_prefix - parent_prefix)
For example:
- From /40 to /48: 2^(48-40) = 256 subnets
- From /40 to /64: 2^(64-40) = 16,777,216 subnets

Frequently Asked Questions
Q: Can I use this for IPv4 planning?
A: No, this tool is specifically designed for IPv6 address planning. IPv4 is legacy, embrace the today.

Q: How are POP allocations determined?
A: POPs are allocated sequentially from the base subnet, using the minimum number of bits required for the POP count. This could probably be smarter, but alas, I am not that smart.

Q: What if my preferred POP size conflicts with the base subnet?
A: The tool will display a warning and adjust the allocation accordingly.

Contributing
We welcome contributions! Please follow these steps:

Fork the repository
Create a feature branch (git checkout -b feature/your-feature)
Commit your changes (git commit -am 'Add some feature')
Push to the branch (git push origin feature/your-feature)
Open a Pull Request