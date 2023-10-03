package mustgathermetrics

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/redhat-openshift-ecosystem/provider-certification-tool/internal/opct/chart"
	log "github.com/sirupsen/logrus"
	"github.com/ulikunitz/xz"
)

type MustGatherMetrics struct {
	fileName        string
	data            *bytes.Buffer
	ReportPath      string
	ReportChartFile string
}

func NewMustGatherMetrics(report, file string, data *bytes.Buffer) (*MustGatherMetrics, error) {

	return &MustGatherMetrics{
		fileName:        filepath.Base(file),
		data:            data,
		ReportPath:      report,
		ReportChartFile: "/metrics.html",
	}, nil
}

func (mg *MustGatherMetrics) Process() error {
	log.Debugf("Processing results/Populating/Populating Summary/Processing/MustGather/Reading")
	tar, err := mg.read(mg.data)
	if err != nil {
		return err
	}
	log.Debugf("Processing results/Populating/Populating Summary/Processing/MustGather/Processing")
	err = mg.extract(tar)
	if err != nil {
		return err
	}
	return nil
}

func (mg *MustGatherMetrics) read(buf *bytes.Buffer) (*tar.Reader, error) {
	file, err := xz.NewReader(buf)
	if err != nil {
		return nil, err
	}
	return tar.NewReader(file), nil
}

// extract dispatch to process must-gather items.
func (mg *MustGatherMetrics) extract(tarball *tar.Reader) error {

	// Walk through files in must-gather tarball file.
	keepReading := true
	metricsPage := chart.NewMetricsPage()
	// charts := make(map[string]*chart.MustGatherMetric, 0)
	for keepReading {
		header, err := tarball.Next()

		switch {
		// no more files
		case err == io.EOF:
			// log.Debugf("Must-gather processor queued, queue size: %d", procQueueSize)
			// waiterProcNS.Wait()
			keepReading = false
			// log.Debugf("Must-gather processor finished, queue size: %d", procQueueSize)
			reportPath := mg.ReportPath + mg.ReportChartFile
			err := chart.SaveMetricsPageReport(metricsPage, reportPath)
			if err != nil {
				log.Errorf("error saving metrics to: %s\n", reportPath)
				return err
			}
			log.Infof("metrics saved at: %s\n", reportPath)
			return nil

		// return on error
		case err != nil:
			return errors.Wrapf(err, "error reading tarball")
			// return err

		// skip it when the headr isn't set (not sure how this happens)
		case header == nil:
			continue
		}

		isMetricData := strings.HasPrefix(header.Name, "monitoring/prometheus/metrics") && strings.HasSuffix(header.Name, ".json.gz")
		if !isMetricData {
			continue
		}

		metricFileName := filepath.Base(header.Name)

		chart, ok := chart.ChartsAvailable[metricFileName]
		if !ok {
			log.Warnf("Unable to find metric data definition for: %s\n", header.Name)
			continue
		}
		if !chart.CollectorAvailable {
			log.Warnf("Ignoring processor for metric data: %s\n", header.Name)
			continue
		}
		fmt.Printf("Processing: %s\n", header.Name)

		gr, err := gzip.NewReader(tarball)
		if err != nil {
			log.Errorf("error unziping the metric: %v", err)
			return err
		}
		defer gr.Close()
		metricPayload, err := ioutil.ReadAll(gr)
		if err != nil {
			log.Errorf("error loading metric data: %v", err)
			return err
		}

		chart.LoadData(metricPayload)
		metricsPage.AddCharts(chart.NewChart())
		fmt.Printf("Done: %s\n", header.Name)
	}

	return nil
}
