package cmd

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// quotaCmd represents the quota command
var quotaCmd = &cobra.Command{
	Use:   "quota [namespace]",
	Short: "Display resource quota for a namespace in human readable format",
	Long: `Display resource quota for a namespace in human readable format.
If no namespace is provided, it will use the current namespace from context.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := ""
		if len(args) > 0 {
			namespace = args[0]
		}

		// Get the current kubeconfig
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

		// Get the namespace if not provided
		if namespace == "" {
			var err error
			namespace, _, err = kubeConfig.Namespace()
			if err != nil {
				log.Fatalf("failed to get namespace: %v", err)
			}
		}

		// Get the REST config
		restConfig, err := kubeConfig.ClientConfig()
		if err != nil {
			log.Fatalf("failed to get REST config: %v", err)
		}

		// Create the clientset
		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			log.Fatalf("failed to create clientset: %v", err)
		}

		// Get resource quotas
		quotas, err := clientset.CoreV1().ResourceQuotas(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatalf("failed to get resource quotas: %v", err)
		}

		if len(quotas.Items) == 0 {
			fmt.Printf("No resource quotas found in namespace '%s'\n", namespace)
			return
		}

		// Display quotas in human readable format
		for _, quota := range quotas.Items {
			// Collect all resource data first to determine column widths
			type resourceInfo struct {
				name          string
				usedFormatted string
				hardFormatted string
				usage         string
			}

			var resources []resourceInfo

			// Process requests
			requestResources := []string{"requests.cpu", "requests.memory", "requests.storage"}
			for _, resourceName := range requestResources {
				if used, hard := getResourceValues(quota.Status.Used, quota.Spec.Hard, resourceName); used != "" && hard != "" {
					usage := calculateUsage(used, hard)
					usedFormatted := formatResourceValue(resourceName, used)
					hardFormatted := formatResourceValue(resourceName, hard)
					resources = append(resources, resourceInfo{
						name:          resourceName,
						usedFormatted: usedFormatted,
						hardFormatted: hardFormatted,
						usage:         usage,
					})
				}
			}

			// Process limits
			limitResources := []string{"limits.cpu", "limits.memory"}
			for _, resourceName := range limitResources {
				if used, hard := getResourceValues(quota.Status.Used, quota.Spec.Hard, resourceName); used != "" && hard != "" {
					usage := calculateUsage(used, hard)
					usedFormatted := formatResourceValue(resourceName, used)
					hardFormatted := formatResourceValue(resourceName, hard)
					resources = append(resources, resourceInfo{
						name:          resourceName,
						usedFormatted: usedFormatted,
						hardFormatted: hardFormatted,
						usage:         usage,
					})
				}
			}

			// Calculate column widths
			nameWidth := len("NAME")
			usedWidth := len("USED")
			hardWidth := len("HARD")
			usageWidth := len("USAGE")

			for _, res := range resources {
				if len(res.name) > nameWidth {
					nameWidth = len(res.name)
				}
				if len(res.usedFormatted) > usedWidth {
					usedWidth = len(res.usedFormatted)
				}
				if len(res.hardFormatted) > hardWidth {
					hardWidth = len(res.hardFormatted)
				}
				if len(res.usage) > usageWidth {
					usageWidth = len(res.usage)
				}
			}

			// Add some padding
			nameWidth += 2
			usedWidth += 2
			hardWidth += 2
			usageWidth += 2

			// Print header
			fmt.Printf("%-*s%-*s%-*s%-*s\n", nameWidth, "NAME", usedWidth, "USED", hardWidth, "HARD", usageWidth, "USAGE")
			fmt.Printf("%s\n", strings.Repeat("-", nameWidth+usedWidth+hardWidth+usageWidth))

			// Print resource data
			for _, res := range resources {
				fmt.Printf("%-*s%-*s%-*s%-*s\n", nameWidth, res.name, usedWidth, res.usedFormatted, hardWidth, res.hardFormatted, usageWidth, res.usage)
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(quotaCmd)
}

// getResourceValues gets the used and hard values for a resource
func getResourceValues(used, hard v1.ResourceList, resourceName string) (string, string) {
	usedValue, usedExists := used[v1.ResourceName(resourceName)]
	hardValue, hardExists := hard[v1.ResourceName(resourceName)]

	if !usedExists || !hardExists {
		return "", ""
	}

	return usedValue.String(), hardValue.String()
}

// calculateUsage calculates the usage percentage
func calculateUsage(usedStr, hardStr string) string {
	usedValue, usedUnit := parseResourceValue(usedStr)
	hardValue, hardUnit := parseResourceValue(hardStr)

	if usedUnit != hardUnit {
		// Convert to the same unit if needed
		usedValue = convertToUnit(usedValue, usedUnit, hardUnit)
	}

	if hardValue == 0 {
		return "N/A"
	}

	percentage := (usedValue / hardValue) * 100
	return fmt.Sprintf("%.1f%%", percentage)
}

// parseResourceValue parses a resource value and returns the numeric value and unit
func parseResourceValue(value string) (float64, string) {
	value = strings.TrimSpace(value)

	// Handle CPU millicores (e.g., "106550m")
	if strings.HasSuffix(value, "m") {
		numStr := strings.TrimSuffix(value, "m")
		num, _ := parseFloat(numStr)
		return num, "m"
	}

	// Handle memory/storage units (Ki, Mi, Gi, Ti, etc.)
	units := []string{"Ki", "Mi", "Gi", "Ti", "Pi", "Ei"}
	for _, unit := range units {
		if strings.HasSuffix(value, unit) {
			numStr := strings.TrimSuffix(value, unit)
			num, _ := parseFloat(numStr)
			return num, unit
		}
	}

	// Handle plain numbers (CPU cores)
	num, _ := parseFloat(value)
	return num, ""
}

// parseFloat safely parses a float64 from a string
func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// convertToUnit converts a value from one unit to another
func convertToUnit(value float64, fromUnit, toUnit string) float64 {
	// Conversion factors for units
	// For CPU: "" = cores, "m" = milli-cores (1 core = 1000 milli-cores)
	// For memory/storage: binary prefixes
	unitFactors := map[string]float64{
		"":   1,     // cores or bytes
		"m":  0.001, // milli-cores or milli-bytes
		"Ki": 1024,
		"Mi": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
		"Ti": 1024 * 1024 * 1024 * 1024,
		"Pi": 1024 * 1024 * 1024 * 1024 * 1024,
		"Ei": 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
	}

	fromFactor := unitFactors[fromUnit]
	toFactor := unitFactors[toUnit]

	// Special handling for converting to/from milli-units
	if fromUnit == "m" && toUnit == "" {
		// Convert milli-units to base units
		return value * fromFactor
	} else if fromUnit == "" && toUnit == "m" {
		// Convert base units to milli-units
		return value / toFactor
	}

	// For memory/storage units, convert through bytes
	fromBytes := fromFactor
	toBytes := toFactor

	return value * fromBytes / toBytes
}

// formatResourceValue formats a resource value in a human readable way
func formatResourceValue(resourceName, value string) string {
	if strings.Contains(resourceName, "cpu") {
		return formatCPUValue(value)
	} else if strings.Contains(resourceName, "memory") || strings.Contains(resourceName, "storage") {
		return formatMemoryValue(value)
	}
	return value
}

// formatCPUValue formats CPU values
func formatCPUValue(value string) string {
	if strings.HasSuffix(value, "m") {
		// Already in millicores
		return value
	}

	// Convert cores to millicores for consistency
	num, err := parseFloat(strings.TrimSpace(value))
	if err != nil {
		return value
	}

	// If it's a whole number of cores, show as cores
	if math.Mod(num, 1) == 0 {
		return fmt.Sprintf("%.0f", num)
	}

	// Otherwise show as millicores
	return fmt.Sprintf("%.0fm", num*1000)
}

// formatMemoryValue formats memory values with unified units
func formatMemoryValue(value string) string {
	num, unit := parseResourceValue(value)

	// Convert to the most appropriate unit
	if unit == "" {
		// Plain number, assume bytes
		return convertAndFormatBytes(num)
	}

	// Already has a unit, convert to the most appropriate one
	return convertAndFormatWithUnit(num, unit)
}

// convertAndFormatBytes converts bytes to the most appropriate unit
func convertAndFormatBytes(bytes float64) string {
	units := []string{"B", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei"}
	unitIndex := 0

	for bytes >= 1024 && unitIndex < len(units)-1 {
		bytes /= 1024
		unitIndex++
	}

	if math.Mod(bytes, 1) == 0 {
		return fmt.Sprintf("%.0f%s", bytes, units[unitIndex])
	}

	return fmt.Sprintf("%.1f%s", bytes, units[unitIndex])
}

// convertAndFormatWithUnit converts a value with unit to the most appropriate unit
func convertAndFormatWithUnit(value float64, unit string) string {
	// Convert to bytes first
	bytes := convertToUnit(value, unit, "")

	// Then convert to the most appropriate unit
	return convertAndFormatBytes(bytes)
}
