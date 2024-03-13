package trends

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	common "github.com/Fiscoon/trends/common"
	prometheus "github.com/Fiscoon/trends/prometheus"
	prometheusApi "github.com/prometheus/client_golang/api/prometheus/v1"
)

const (
	queryRangeDays = 7

	TrendsTextGreen           = ":large_green_circle: *%s*: This cluster is considered to be in good status regarding CPU consumption. "
	TrendsTextYellow          = ":large_yellow_circle: *%s*: This cluster is considered to be in average status regarding CPU consumption. "
	TrendsTextRed             = ":red_circle: *%s*: This cluster is considered to be in bad/critical status regarding CPU consumption. "
	TrendsText70PercSurpass   = "There were spikes surpassing the 70%% threshold from (%s). "
	TrendsText90PercSurpass   = "There were spikes surpassing the 90%% threshold from (%s). "
	TrendsText99_5PercSurpass = "There were spikes reaching 100%% CPU usage from (%s). "
)

var AllowedClusters = []string{
	"nl", "nl-mt", "nl-dta", "nl-utre", "ld", "ld7",
	"ld7-dta", "sg", "sg3", "jb", "mi", "hk", "vsan-01",
}

type Host struct {
	name          string
	cpuUsageSteps *[]float64
}

type Cluster struct {
	name                  string
	api                   *prometheusApi.API
	hosts                 []Host
	hostnamesOver70Perc   []string
	hostnamesOver90Perc   []string
	hostnamesOver99_5Perc []string
	score                 uint
}

func NewCluster(name string) (*Cluster, error) {
	promApi, err := prometheus.NewPromAPI("https://thanos.prod.env")
	if err != nil {
		return nil, err
	}
	return &Cluster{
		name: name,
		api:  &promApi,
	}, nil
}

func CreateClusterObjects(clusterNames []string) ([]*Cluster, error) {
	clusters := make([]*Cluster, 0)
	for _, clusterName := range clusterNames {
		cluster, err := NewCluster(clusterName)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func (c *Cluster) GetHosts() error {
	promQuery := fmt.Sprintf(prometheus.ListHostsQuery, c.name)
	if strings.Contains(c.name, "utre") {
		promQuery = fmt.Sprintf(prometheus.ListHostsQueryUtre, c.name)
	}

	res, err := prometheus.PromBasicQuery(promQuery, *c.api)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(res, "\n")
	for _, line := range lines {
		host := common.FindHostRegex(line)
		if host == "" {
			continue
		}
		c.hosts = append(c.hosts, Host{name: host, cpuUsageSteps: new([]float64)})
	}

	if len(c.hosts) == 0 {
		return fmt.Errorf("no hosts found for %s cluster, exiting", c.name)
	}
	return nil
}

func (c *Cluster) GetCpuUsageSteps() {
	timeRange := prometheusApi.Range{
		Start: time.Now().UTC().AddDate(0, 0, -queryRangeDays),
		End:   time.Now().UTC(),
		Step:  time.Minute,
	}
	for _, host := range c.hosts {
		var hostCpuUsageByStep []float64
		promQuery := fmt.Sprintf(prometheus.GetCpuQuery, host.name)

		if strings.Contains(c.name, "utre") {
			promQuery = fmt.Sprintf(prometheus.GetCpuQueryUtre, c.name, host.name)
		}

		res, err := prometheus.PromRangeQuery(promQuery, *c.api, timeRange)
		if err != nil {
			panic(err)
		}

		lines := strings.Split(res, "\n")
		for _, line := range lines[2:] {
			line, _, _ = strings.Cut(line, " @")
			cpuUsageValue, err := strconv.ParseFloat(line, 64)
			if err != nil {
				panic(err)
			}
			hostCpuUsageByStep = append(hostCpuUsageByStep, cpuUsageValue)
		}
		*host.cpuUsageSteps = hostCpuUsageByStep
	}
}

func (c *Cluster) CalculateTrendsScore() {
	for _, host := range c.hosts {
		cpuAvg := common.Average(*host.cpuUsageSteps)
		countOver70 := common.CountOverThreshold(*host.cpuUsageSteps, 70.0)
		countOver90 := common.CountOverThreshold(*host.cpuUsageSteps, 90.0)
		countOver99_5 := common.CountOverThreshold(*host.cpuUsageSteps, 99.5)

		if countOver70 == 0 {
			continue // This host is completely fine
		}

		if countOver99_5 != 0 {
			if countOver99_5 == 1 {
				c.score += 38
			} else {
				c.score += 55
			}
			c.hostnamesOver99_5Perc = append(c.hostnamesOver99_5Perc, host.name)
			continue
		}
		if countOver90 != 0 {
			if countOver90 == 1 {
				c.score += 20
			} else {
				c.score += 27
			}
			c.hostnamesOver90Perc = append(c.hostnamesOver90Perc, host.name)
			continue
		}
		if countOver70 != 0 {
			if countOver70 == 1 {
				c.score += 8
			} else {
				c.score += 12
			}
			c.hostnamesOver70Perc = append(c.hostnamesOver70Perc, host.name)
		}
		if cpuAvg > 70 {
			c.score += 40
		}
	}

	c.score = c.score / uint(len(c.hosts)) //average!!
}

func (c *Cluster) DefineTrendsMessage() string {
	var message string
	if c.score < 20 {
		message = fmt.Sprintf(TrendsTextGreen, strings.ToUpper(c.name))
	} else if c.score < 40 {
		message = fmt.Sprintf(TrendsTextYellow, strings.ToUpper(c.name))
	} else {
		message = fmt.Sprintf(TrendsTextRed, strings.ToUpper(c.name))
	}
	//if (c.hostnamesOver70Perc) != nil {
	//	message += fmt.Sprintf(trendsText70PercSurpass, strings.Join(c.hostnamesOver70Perc, ", "))
	//}
	if (c.hostnamesOver90Perc) != nil {
		message += fmt.Sprintf(TrendsText90PercSurpass, strings.Join(c.hostnamesOver90Perc, ", "))
	}
	if (c.hostnamesOver99_5Perc) != nil {
		message += fmt.Sprintf(TrendsText99_5PercSurpass, strings.Join(c.hostnamesOver99_5Perc, ", "))
	}

	return message
}
