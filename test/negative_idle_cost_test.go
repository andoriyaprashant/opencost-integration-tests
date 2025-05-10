package tests

import (
    "encoding/json"
    "io"
    "net/http"
    "testing"
)

type NamespaceEntry struct {
    Name       string  `json:"name"`
    CPUCost    float64 `json:"cpuCost"`
    RAMCost    float64 `json:"ramCost"`
    GPUCost    float64 `json:"gpuCost"`
    PVCost     float64 `json:"pvCost"`
    TotalCost  float64 `json:"totalCost"`
    Efficiency float64 `json:"totalEfficiency"`
    Start      string  `json:"start"`
    End        string  `json:"end"`
}

type AllocationData map[string]NamespaceEntry

type AllocationResponse struct {
    Code   int              `json:"code"`
    Status string           `json:"status"`
    Data   []AllocationData `json:"data"`
}

func TestNegativeIdleValues(t *testing.T) {
    url := "https://demo.infra.opencost.io/model/allocation/compute?window=2025-05-10T00:00:00Z,2025-05-11T00:00:00Z&aggregate=namespace&includeIdle=true&step=1d&accumulate=false"

    resp, err := http.Get(url)
    if err != nil {
        t.Fatalf("Failed to fetch allocation data: %v", err)
    }
    defer resp.Body.Close()

    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        t.Fatalf("Failed to read response body: %v", err)
    }

    var result AllocationResponse
    if err := json.Unmarshal(bodyBytes, &result); err != nil {
        t.Fatalf("Failed to unmarshal JSON: %v\nRaw body:\n%s", err, bodyBytes)
    }

    foundNegative := false

    for _, allocation := range result.Data {
        if idleEntry, exists := allocation["__idle__"]; exists {
            t.Logf("Found __idle__ entry: CPU=$%.2f, GPU=$%.2f, RAM=$%.2f, PV=$%.2f, Total=$%.2f, Efficiency=%.2f",
                idleEntry.CPUCost, idleEntry.GPUCost, idleEntry.RAMCost, idleEntry.PVCost, idleEntry.TotalCost, idleEntry.Efficiency)

            t.Logf("Window start: %s | end: %s", idleEntry.Start, idleEntry.End)

            if idleEntry.TotalCost < 0 {
                t.Errorf("__idle__ entry has negative TotalCost: $%.2f", idleEntry.TotalCost)
                foundNegative = true
            }
            if idleEntry.CPUCost < 0 {
                t.Errorf("__idle__ entry has negative CPU cost: $%.2f", idleEntry.CPUCost)
                foundNegative = true
            }
            if idleEntry.RAMCost < 0 {
                t.Errorf("__idle__ entry has negative RAM cost: $%.2f", idleEntry.RAMCost)
                foundNegative = true
            }
            if idleEntry.GPUCost < 0 {
                t.Errorf("__idle__ entry has negative GPU cost: $%.2f", idleEntry.GPUCost)
                foundNegative = true
            }
        } else {
            t.Log("No __idle__ entry found in this allocation.")
        }
    }

    if !foundNegative {
        t.Log("No negative idle-related values found — test passed.")
    } else {
        t.Log("Some negative idle-related values found — test failed.")
    }
}
