package main

import (
	"fmt"
	"sync"

	"github.com/Fiscoon/trends/trends"
)

func main() {
	var wg sync.WaitGroup

	clusters, err := trends.CreateClusterObjects(trends.AllowedClusters)
	if err != nil {
		panic(err)
	}

	for _, cluster := range clusters {
		wg.Add(1)
		go func(cluster *trends.Cluster) {
			defer wg.Done()
			err = cluster.GetHosts()
			if err != nil {
				panic(err)
			}
			cluster.GetCpuUsageSteps()
			cluster.CalculateTrendsScore()
		}(cluster)
	}
	wg.Wait()

	for _, cluster := range clusters {
		fmt.Println(cluster.DefineTrendsMessage())
		fmt.Println()
	}
}
