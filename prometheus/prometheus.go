package prometheus

import (
	"context"
	"net/http"
	"time"

	prometheusClient "github.com/prometheus/client_golang/api"
	prometheusApi "github.com/prometheus/client_golang/api/prometheus/v1"
)

const ListHostsQuery = "vsphere_host_cpu_usage_average{cpu=\"instance-total\",owner=\"platform-rke-prod-env\",clustername=\"%s\"}"
const ListHostsQueryUtre = "vsphere_host_cpu_usage_average{cpu=\"instance-total\", owner=\"gtc-ops-prod-rke-utre-env\", clustername=\"%s\"}"
const GetCpuQuery = "max(vsphere_host_cpu_usage_average{cpu=\"instance-total\",owner=\"platform-rke-prod-env\",esxhostname=~\"%s\"})"
const GetCpuQueryUtre = "max(vsphere_host_cpu_usage_average{cpu=\"instance-total\", clustername=\"%s\", esxhostname=~\"%s\"})"

func NewPromAPI(url string) (prometheusApi.API, error) {
	cfg := prometheusClient.Config{
		Address: url,
		Client:  http.DefaultClient,
	}
	client, err := prometheusClient.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return prometheusApi.NewAPI(client), nil
}

func PromBasicQuery(query string, api prometheusApi.API) (string, error) {
	res, _, err := api.Query(context.TODO(), query, time.Now())
	if err != nil {
		return "", err
	}
	return res.String(), nil
}

func PromRangeQuery(query string, api prometheusApi.API, promRange prometheusApi.Range) (string, error) {
	res, _, err := api.QueryRange(context.TODO(), query, promRange)
	if err != nil {
		return "", err
	}
	return res.String(), nil
}
