package chart

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	log "github.com/sirupsen/logrus"
)

type MetricValue struct {
	Timestap time.Time
	Value    string
}

type PrometheusResultMetric struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string                   `json:"resultType"`
		Result     []PrometheusResultMetric `json:"result"`
	} `json:"data"`
}

type readMetricInput struct {
	filename string
	label    string
	title    string
	subtitle string
}

// type LineExamples struct{}

type MustGatherMetric struct {
	Path               string
	OriginalQuery      string
	PlotLabel          string
	PlotTitle          string
	PlotSubTitle       string
	CreateChart        func() *charts.Line
	CollectorAvailable bool
	MetricData         *PrometheusResponse
}

var ChartsAvailable map[string]*MustGatherMetric

func init() {
	ChartsAvailable = make(map[string]*MustGatherMetric, 0)
	ChartsAvailable["query_range-etcd-disk-fsync-db-duration-p99.json.gz"] = &MustGatherMetric{
		Path:          "query_range-etcd-disk-fsync-db-duration-p99.json.gz",
		OriginalQuery: "",
		PlotLabel:     "instance",
		PlotTitle:     "etcd fsync DB p99",
		PlotSubTitle:  "",
		// CreateChart:        plotETCDp99,
		CollectorAvailable: true,
	}
	ChartsAvailable["query_range-api-kas-request-duration-p99.json.gz"] = &MustGatherMetric{
		Path:          "query_range-api-kas-request-duration-p99.json.gz",
		OriginalQuery: "",
		PlotLabel:     "verb",
		PlotTitle:     "Kube API request p99",
		PlotSubTitle:  "",
		// CreateChart:        plotKubeAPI,
		CollectorAvailable: true,
	}
	ChartsAvailable["query_range-etcd-disk-fsync-wal-duration-p99.json.gz"] = &MustGatherMetric{
		Path:               "query_range-etcd-disk-fsync-wal-duration-p99.json.gz",
		OriginalQuery:      "",
		PlotLabel:          "instance",
		PlotTitle:          "etcd fsync WAL p99",
		PlotSubTitle:       "",
		CollectorAvailable: true,
	}
	ChartsAvailable["query_range-etcd-peer-round-trip-time.json.gz"] = &MustGatherMetric{
		Path:               "query_range-etcd-peer-round-trip-time.json.gz",
		OriginalQuery:      "",
		PlotLabel:          "instance",
		PlotTitle:          "etcd peer round trip",
		PlotSubTitle:       "",
		CollectorAvailable: true,
	}

	ChartsAvailable["query_range-etcd-total-leader-elections-day.json.gz"] = &MustGatherMetric{
		Path:               "query_range-etcd-total-leader-elections-day.json.gz",
		OriginalQuery:      "",
		PlotLabel:          "instance",
		PlotTitle:          "etcd peer total leader election",
		PlotSubTitle:       "",
		CollectorAvailable: true,
	}
	ChartsAvailable["query_range-etcd-request-duration-p99.json.gz"] = &MustGatherMetric{
		Path:               "query_range-etcd-request-duration-p99.json.gz",
		OriginalQuery:      "",
		PlotLabel:          "operation",
		PlotTitle:          "etcd req duration p99",
		PlotSubTitle:       "",
		CollectorAvailable: false,
	}

	ChartsAvailable["query_range-cluster-storage-iops.json.gz"] = &MustGatherMetric{
		Path:               "query_range-cluster-storage-iops.json.gz",
		OriginalQuery:      "",
		PlotLabel:          "namespace",
		PlotTitle:          "Cluster storage IOPS",
		PlotSubTitle:       "",
		CollectorAvailable: false,
	}
	ChartsAvailable["query_range-cluster-storage-throughput.json.gz"] = &MustGatherMetric{
		Path:               "query_range-cluster-storage-throughput.json.gz",
		OriginalQuery:      "",
		PlotLabel:          "namespace",
		PlotTitle:          "Cluster storage throughput",
		PlotSubTitle:       "",
		CollectorAvailable: false,
	}
	ChartsAvailable["query_range-cluster-cpu-usage.json.gz"] = &MustGatherMetric{
		Path:               "query_range-cluster-cpu-usage.json.gz",
		OriginalQuery:      "",
		PlotLabel:          "namespace",
		PlotTitle:          "Cluster CPU",
		PlotSubTitle:       "",
		CollectorAvailable: false,
	}
}

func NewMetricsPage() *components.Page {
	page := components.NewPage()
	page.PageTitle = "OPCT Report Metrics"
	return page
}

func SaveMetricsPageReport(page *components.Page, path string) error {
	// Create HTML metrics file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	page.Render(io.MultiWriter(f))
	return nil
}

func (mmm *MustGatherMetric) NewChart() *charts.Line {
	return mmm.processMetric(&readMetricInput{
		filename: mmm.Path,
		label:    mmm.PlotLabel,
		title:    mmm.PlotTitle,
		subtitle: mmm.PlotSubTitle,
	})
}

func (mmm *MustGatherMetric) LoadData(payload []byte) error {
	mmm.MetricData = &PrometheusResponse{}

	err := json.Unmarshal(payload, &mmm.MetricData)
	if err != nil {
		log.Errorf("error parsing metric data: %v", err)
		return err
	}
	fmt.Printf("Metric Loaded with type: %s\n", mmm.MetricData.Status)

	return nil
}

func (mmm *MustGatherMetric) processMetric(in *readMetricInput) *charts.Line {

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    in.title,
			Subtitle: in.subtitle,
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true, Trigger: "axis"}),
	)

	allTimestamps := []string{}

	type ChartData struct {
		Label      string
		DataPoints []opts.LineData
	}

	chartData := []ChartData{}
	for _, res := range mmm.MetricData.Data.Result {
		chart := ChartData{
			Label:      res.Metric[in.label],
			DataPoints: make([]opts.LineData, 0),
		}
		for _, datapoints := range res.Values {
			// instance1Items = append(instance1Items, datapoints[0].(string))
			value := datapoints[1].(string)
			if value == "" {
				fmt.Printf("Empty value [%s], ignoring...", value)
				continue
			}
			tm := time.Unix(int64(datapoints[0].(float64)), 0)
			strTimestamp := fmt.Sprintf("%d-%d-%d %d:%d:%d", tm.Year(), tm.Month(), tm.Day(),
				tm.Hour(), tm.Minute(), tm.Second())
			allTimestamps = append(allTimestamps, strTimestamp)
			// fmt.Println(tm.String(), value)
			chart.DataPoints = append(chart.DataPoints, opts.LineData{Value: value})
		}
		chartData = append(chartData, chart)
	}
	sort.Strings(allTimestamps)
	line.SetXAxis(allTimestamps).
		SetSeriesOptions(charts.WithLineChartOpts(
			opts.LineChart{Smooth: false, ShowSymbol: true, SymbolSize: 15, Symbol: "diamond"},
		))
	for _, chart := range chartData {
		line.AddSeries(chart.Label, chart.DataPoints)
	}

	return line
}
