package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Mireu-Lab/cloud-pricing-mcp/internal/catalog"
	"github.com/Mireu-Lab/cloud-pricing-mcp/internal/models"
	"github.com/Mireu-Lab/cloud-pricing-mcp/pkg/calculator"
	"github.com/Mireu-Lab/cloud-pricing-mcp/pkg/mcp"
)

func main() {
	if len(os.Args) < 2 {
		// Default to MCP stdio server mode if no args given
		runServer()
		return
	}

	command := os.Args[1]
	switch command {
	case "server":
		runServer()

	case "compare":
		runCompare(os.Args[2:])

	case "aws":
		runAWS(os.Args[2:])

	case "azure":
		runAzure(os.Args[2:])

	case "gcp":
		runGCP(os.Args[2:])

	case "catalog", "pricing":
		runCatalog(os.Args[2:])

	case "help", "-h", "--help":
		printHelp()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printHelp()
		os.Exit(1)
	}
}

func runServer() {
	srv := mcp.NewServer(os.Stdin, os.Stdout)
	if err := srv.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server exited with error: %v\n", err)
		os.Exit(1)
	}
}

func runCompare(args []string) {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)
	vcpu := fs.Int("vcpu", 2, "vCPU count")
	ram := fs.Float64("ram", 4, "Memory in GB")
	storage := fs.Float64("storage", 50, "Storage in GB")
	storageType := fs.String("storage-type", "ssd", "Storage type (ssd/hdd)")
	hours := fs.Float64("hours", 730, "Monthly hours (default 730)")
	egress := fs.Float64("egress", 100, "Outbound egress GB")
	spot := fs.Bool("spot", false, "Use Spot/Preemptible pricing")
	fs.Parse(args)

	res, err := calculator.CompareBig3(models.CompareRequest{
		VCPU:          *vcpu,
		MemoryGB:      *ram,
		StorageGB:     *storage,
		StorageType:   *storageType,
		HoursPerMonth: *hours,
		OutboundGB:    *egress,
		IsSpot:        *spot,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Comparison error: %v\n", err)
		os.Exit(1)
	}
	printJSON(res)
}

func runAWS(args []string) {
	fs := flag.NewFlagSet("aws", flag.ExitOnError)
	inst := fs.String("instance", "t4g.medium", "t4g.small, t4g.medium, t3.medium, m6i.xlarge, g4dn.xlarge")
	count := fs.Int("count", 1, "Instance count")
	hours := fs.Float64("hours", 730, "Running hours")
	ebs := fs.Float64("ebs", 50, "EBS gp3 GB")
	egress := fs.Float64("egress", 0, "Outbound egress GB")
	spot := fs.Bool("spot", false, "Use Spot instance")
	fs.Parse(args)

	res, err := calculator.CalculateAWS(*inst, *count, *hours, *ebs, *egress, *spot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AWS calculation error: %v\n", err)
		os.Exit(1)
	}
	printJSON(res)
}

func runAzure(args []string) {
	fs := flag.NewFlagSet("azure", flag.ExitOnError)
	vm := fs.String("vm", "Standard_B2s", "Standard_B2s, Standard_D4s_v5, Standard_NC4as_T4_v3")
	count := fs.Int("count", 1, "Instance count")
	hours := fs.Float64("hours", 730, "Running hours")
	disk := fs.Float64("disk", 50, "Managed disk GB")
	egress := fs.Float64("egress", 0, "Outbound egress GB")
	spot := fs.Bool("spot", false, "Use Spot VM")
	fs.Parse(args)

	res, err := calculator.CalculateAzure(*vm, *count, *hours, *disk, *egress, *spot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Azure calculation error: %v\n", err)
		os.Exit(1)
	}
	printJSON(res)
}

func runGCP(args []string) {
	fs := flag.NewFlagSet("gcp", flag.ExitOnError)
	machine := fs.String("machine", "e2-standard-2", "e2-micro, e2-standard-2, e2-standard-4, g2-standard-4")
	count := fs.Int("count", 1, "Instance count")
	hours := fs.Float64("hours", 730, "Running hours")
	disk := fs.Float64("disk", 50, "pd-balanced disk GB")
	egress := fs.Float64("egress", 0, "Outbound egress GB")
	spot := fs.Bool("spot", false, "Use Spot VM")
	fs.Parse(args)

	res, err := calculator.CalculateGCP(*machine, *count, *hours, *disk, *egress, *spot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GCP calculation error: %v\n", err)
		os.Exit(1)
	}
	printJSON(res)
}

func runCatalog(args []string) {
	fs := flag.NewFlagSet("catalog", flag.ExitOnError)
	prov := fs.String("provider", "", "aws, azure, gcp")
	fs.Parse(args)

	if *prov != "" {
		printJSON(catalog.GetProductsByProvider(models.Provider(*prov)))
		return
	}
	printJSON(catalog.GetAllProducts())
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printHelp() {
	helpText := `Big 3 Cloud Pricing Calculator MCP & CLI (Mireu-Lab)
Supported Providers: AWS, Azure, GCP

Usage:
  cloud-pricing-mcp [command] [flags]

Commands:
  server              Run as Stdio MCP server (Default)
  compare             Compare costs across Big 3 cloud providers (AWS, Azure, GCP)
  aws                 Calculate AWS EC2, EBS, and Data Transfer costs
  azure               Calculate Azure VM, Disk, and Bandwidth costs
  gcp                 Calculate Google Cloud Compute Engine & Storage costs
  catalog (pricing)   View embedded pricing catalog by provider

Examples:
  cloud-pricing-mcp compare -vcpu 2 -ram 4 -storage 50 -egress 100
  cloud-pricing-mcp aws -instance t4g.medium -ebs 100 -spot
  cloud-pricing-mcp azure -vm Standard_B2s -disk 100
  cloud-pricing-mcp gcp -machine e2-standard-2 -disk 100 -spot
  cloud-pricing-mcp catalog -provider gcp
`
	fmt.Println(helpText)
}
