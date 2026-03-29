package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

func formatJSON(data json.RawMessage) string {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		return string(data)
	}
	return pretty.String()
}

func formatRate(v float64, metric string) string {
	if strings.Contains(metric, "bytes") || metric == "bytes" {
		switch {
		case v >= 1e9:
			return fmt.Sprintf("%.2f Gbps", v/1e9)
		case v >= 1e6:
			return fmt.Sprintf("%.2f Mbps", v/1e6)
		case v >= 1e3:
			return fmt.Sprintf("%.2f Kbps", v/1e3)
		default:
			return fmt.Sprintf("%.2f bps", v)
		}
	}
	switch {
	case v >= 1e6:
		return fmt.Sprintf("%.2fM", v/1e6)
	case v >= 1e3:
		return fmt.Sprintf("%.2fK", v/1e3)
	default:
		return fmt.Sprintf("%.2f", v)
	}
}
