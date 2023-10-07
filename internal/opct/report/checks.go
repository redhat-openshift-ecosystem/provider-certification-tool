/*
Checks handles all acceptance criteria from data
collected and processed in summary package.

Existing Checks:
- OPCT-001: "Plugin Conformance Kubernetes [10-openshift-kube-conformance] must pass (after filters)"
- OPCT-002: "Plugin Conformance Upgrade [05-openshift-cluster-upgrade] must pass"
- OPCT-003: "Plugin Collector [99-openshift-artifacts-collector] must pass"
- ...TBD
*/
package report

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/redhat-openshift-ecosystem/provider-certification-tool/internal/opct/plugin"
	log "github.com/sirupsen/logrus"
)

const (
	docsRulesPath  = "/review/rules"
	defaultBaseURL = "https://redhat-openshift-ecosystem.github.io/provider-certification-tool"

	CheckResultNamePass = "pass"
	CheckResultNameFail = "fail"
	CheckResultNameSkip = "skip"
)

type CheckResultName string

type CheckResult struct {
	Name    CheckResultName `json:"result"`
	Message string          `json:"message"`
	Want    string          `json:"want"`
	Got     string          `json:"got"`
}

func (cr *CheckResult) String() string {
	return string(cr.Name)
}

type Check struct {
	// ID is the unique identifier for the check. It is used
	// to mount the documentation for each check.
	ID string `json:"id"`

	// Name is the unique name for the check to be reported.
	// It must have short and descriptive name identifying the
	// failure item.
	Name string `json:"name"`

	// Description describes shortly the check.
	Description string `json:"description"`

	// Reference must point to documentation URL to review the
	// item.
	Reference string `json:"reference"`

	// Accepted must report acceptance criteria, when true
	// the Check is accepted by the tool, otherwise it is
	// failed and must be reviewede.
	Result CheckResult `json:"result"`

	// ResultMessage string `json:"resultMessage"`

	Test func() CheckResult `json:"-"`

	// Priority is the priority to execute the check.
	// 0 is higher.
	Priority uint64
}

func ExampleAcceptanceCheckPass() CheckResultName {
	return CheckResultNamePass
}

func AcceptanceCheckFail() CheckResultName {
	return CheckResultNameFail
}

// func CheckRespCustomFail(custom string) CheckResult {
// 	resp := CheckResult(fmt.Sprintf("%s [%s]", CheckResultNameFail, custom))
// 	return resp
// }

// CheckSummary aggregates the checks.
type CheckSummary struct {
	baseURL string
	Checks  []*Check `json:"checks"`
}

func NewCheckSummary(re *Report) *CheckSummary {

	baseURL := defaultBaseURL
	msgDefaultNotMatch := "default value does not match the acceptance criteria"
	// Developer environment:
	// $ mkdocs serve
	// $ export OPCT_DEV_BASE_URL_DOC="http://127.0.0.1:8000/provider-certification-tool"
	localDevBaseURL := os.Getenv("OPCT_DEV_BASE_URL_DOC")
	if localDevBaseURL != "" {
		baseURL = localDevBaseURL
	}
	checkSum := &CheckSummary{
		Checks:  []*Check{},
		baseURL: fmt.Sprintf("%s%s", baseURL, docsRulesPath),
	}

	// OpenShift / Infrastructure Object Check
	checkSum.Checks = append(checkSum.Checks, &Check{
		Name: "Platform Type should be None or External",
		Test: func() CheckResult {
			prefix := "Check OPCT-TBD Failed"
			res := CheckResult{Name: CheckResultNameFail, Want: "None|External|AWS|Azure"}
			if re.Provider == nil || re.Provider.Infra == nil {
				res.Message = fmt.Sprintf("%s: unable to read the infrastructure object", prefix)
				log.Debug(res.Message)
				return res
			}
			// Acceptance Criteria
			res.Got = re.Provider.Infra.PlatformType
			switch res.Got {
			case "None":
			case "External":
			case "AWS":
			case "Azure":
				res.Message = CheckResultNamePass
				return res
			}
			log.Debugf("%s (Platform Type): %s: got=[%s]", prefix, msgDefaultNotMatch, re.Provider.Infra.PlatformType)
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		Name: "Cluster Version Operator must be Available",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "True"}
			if re.Provider == nil || re.Provider.Version == nil || re.Provider.Version.OpenShift == nil {
				res.Message = fmt.Sprintf("%s: unable to read provider version")
				return res
			}
			res.Got = re.Provider.Version.OpenShift.CondAvailable
			if res.Got != "True" {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		Name: "Cluster condition Failing must be False",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "False"}
			if re.Provider == nil || re.Provider.Version == nil || re.Provider.Version.OpenShift == nil {
				res.Message = fmt.Sprintf("%s: unable to read provider version")
				return res
			}
			res.Got = re.Provider.Version.OpenShift.CondFailing
			if res.Got != "False" {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		Name: "Cluster upgrade must not be Progressing",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "False"}
			if re.Provider == nil || re.Provider.Version == nil || re.Provider.Version.OpenShift == nil {
				return res
			}
			res.Got = re.Provider.Version.OpenShift.CondProgressing
			if res.Got != "False" {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		Name: "Cluster ReleaseAccepted must be True",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "True"}
			if re.Provider == nil || re.Provider.Version == nil || re.Provider.Version.OpenShift == nil {
				return res
			}
			res.Got = re.Provider.Version.OpenShift.CondReleaseAccepted
			if res.Got != "True" {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		Name: "Infrastructure status must have Topology=HighlyAvailable",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "HighlyAvailable"}
			if re.Provider == nil || re.Provider.Infra == nil {
				return res
			}
			res.Got = re.Provider.Infra.Topology
			if res.Got != "HighlyAvailable" {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		Name: "Infrastructure status must have ControlPlaneTopology=HighlyAvailable",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "HighlyAvailable"}
			if re.Provider == nil || re.Provider.Infra == nil {
				return res
			}
			res.Got = re.Provider.Infra.ControlPlaneTopology
			if re.Provider.Infra.ControlPlaneTopology != "HighlyAvailable" {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-008",
		Name: "All nodes must be healthy",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "100"}
			if re.Provider == nil || re.Provider.ClusterHealth == nil {
				log.Debugf("Check Failed: OPCT-008: unavailable results")
				return res
			}
			if re.Provider.ClusterHealth.NodeHealthPerc != 100 {
				log.Debugf("Check Failed: OPCT-008: want[!=100] got[%f]", re.Provider.ClusterHealth.NodeHealthPerc)
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-009",
		Name: "Pods Healthy must report higher than 98%",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: ">=98%"}
			if re.Provider == nil || re.Provider.ClusterHealth == nil {
				return res
			}
			if re.Provider.ClusterHealth.PodHealthPerc < 98.0 {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	// Plugins Checks
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-001",
		Name: "Plugin Conformance Kubernetes [10-openshift-kube-conformance] must pass (after filters)",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "Pass|0|Priority==0|Total!=Failed"}
			prefix := "Check OPCT-001 Failed"
			if _, ok := re.Provider.Plugins[plugin.PluginNameKubernetesConformance]; !ok {
				log.Debugf("%s Runtime: processed plugin data not found: %v", prefix, re.Provider.Plugins[plugin.PluginNameKubernetesConformance])
				return res
			}
			p := re.Provider.Plugins[plugin.PluginNameKubernetesConformance]
			if p.Stat.Total == p.Stat.Failed {
				res.Message = "Potential Runtime Failure. Check the Plugin logs."
				log.Debugf("%s Runtime: Total and Failed counters are equals indicating execution failure", prefix)
				return res
			}
			if len(p.TestsFailedPrio) > 0 {
				log.Debugf("%s Acceptance criteria: TestsFailedPrio counter are greater than 0: %v", prefix, len(p.TestsFailedPrio))
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-004",
		Name: "OpenShift Conformance [20-openshift-conformance-validated]: Failed tests must report less than 1.5%",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "<=1.5"}
			if _, ok := re.Provider.Plugins[plugin.PluginNameOpenShiftConformance]; !ok {
				return res
			}
			// "Acceptance" are relative, the baselines is observed to set
			// an "accepted" value considering a healthy cluster in known provider/installation.
			plugin := re.Provider.Plugins[plugin.PluginNameOpenShiftConformance]
			perc := (float64(plugin.Stat.Failed) / float64(plugin.Stat.Total)) * 100
			if perc > 1.5 {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-005",
		Name: "OpenShift Conformance [20-openshift-conformance-validated]: Priority must report less than 0.5%",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "<=0.5"}
			if _, ok := re.Provider.Plugins[plugin.PluginNameOpenShiftConformance]; !ok {
				return res
			}
			// "Acceptance" are relative, the baselines is observed to set
			// an "accepted" value considering a healthy cluster in known provider/installation.
			plugin := re.Provider.Plugins[plugin.PluginNameOpenShiftConformance]
			perc := (float64(plugin.Stat.FilterFailedPrio) / float64(plugin.Stat.Total)) * 100
			if perc > 0.5 {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-006",
		Name: "Suite Errors must report a lower number of log errors",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "<=150"}
			if re.Provider.ErrorCounters == nil {
				return res
			}
			cnt := *re.Provider.ErrorCounters
			if _, ok := cnt["total"]; !ok {
				res.Message = "Unable to load Total Counter"
				return res
			}
			// "Acceptance" are relative, the baselines is observed to set
			// an "accepted" value considering a healthy cluster in known provider/installation.
			total := cnt["total"]
			res.Got = fmt.Sprintf("%d", total)
			if total > 150 {
				return res
			}
			// 0? really? something went wrong!
			if total == 0 {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-007",
		Name: "Workloads must report a lower number of errors in the logs",
		Test: func() CheckResult {
			wantLimit := 30000
			res := CheckResult{Name: CheckResultNameFail, Want: fmt.Sprintf("<=%d", wantLimit)}
			prefix := "Check OPCT-007 Failed"
			if re.Provider.MustGatherInfo == nil {
				log.Debugf("%s: MustGatherInfo is not defined", prefix)
				return res
			}
			if _, ok := re.Provider.MustGatherInfo.ErrorCounters["total"]; !ok {
				log.Debugf("%s: OPCT-007: ErrorCounters[\"total\"]", prefix)
				return res
			}
			// "Acceptance" are relative, the baselines is observed to set
			// an "accepted" value considering a healthy cluster in known provider/installation.
			total := re.Provider.MustGatherInfo.ErrorCounters["total"]
			res.Got = fmt.Sprintf("%d", total)
			if total > wantLimit {
				log.Debugf("%s acceptance criteria: want[<=%d] got[%d]", prefix, wantLimit, total)
				return res
			}
			// 0? really? something went wrong!
			if total == 0 {
				log.Debugf("%s acceptance criteria: want[!=0] got[%d]", prefix, total)
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-003",
		Name: "Plugin Collector [99-openshift-artifacts-collector] must pass",
		Test: func() CheckResult {
			prefix := "Check OPCT-003 Failed"
			res := CheckResult{Name: CheckResultNameFail, Want: "passed"}
			if _, ok := re.Provider.Plugins[plugin.PluginNameArtifactsCollector]; !ok {
				return res
			}
			p := re.Provider.Plugins[plugin.PluginNameArtifactsCollector]
			if p.Stat.Total == p.Stat.Failed {
				log.Debugf("%s Runtime: Total and Failed counters are equals indicating execution failure", prefix)
				return res
			}
			// Acceptance check
			if re.Provider.Plugins[plugin.PluginNameArtifactsCollector].Stat.Status == "passed" {
				res.Message = CheckResultNamePass
				return res
			}
			log.Debugf("%s: %s", prefix, msgDefaultNotMatch)
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-002",
		Name: "Plugin Conformance Upgrade [05-openshift-cluster-upgrade] must pass",
		Test: func() CheckResult {
			res := CheckResult{Name: CheckResultNameFail, Want: "passed"}
			if _, ok := re.Provider.Plugins[plugin.PluginNameOpenShiftUpgrade]; !ok {
				return res
			}
			if re.Provider.Plugins[plugin.PluginNameOpenShiftUpgrade].Stat.Status != "passed" {
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	// TODO(etcd)
	/*
		checkSum.Checks = append(checkSum.Checks, &Check{
			Name: "[TODO] etcd fio must accept the tests (TODO)",
			Test: AcceptanceCheckFail,
		})
		checkSum.Checks = append(checkSum.Checks, &Check{
			Name: "[TODO] etcd slow requests: p99 must be lower than 900ms",
			Test: AcceptanceCheckFail,
		})
	*/
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-010",
		Name: "etcd logs: slow requests: average should be under 500ms",
		Test: func() CheckResult {
			prefix := "Check OPCT-010 Failed"
			wantLimit := 500.0
			res := CheckResult{Name: CheckResultNameFail, Want: fmt.Sprintf("<=%f ms", wantLimit)}
			if re.Provider.MustGatherInfo == nil {
				log.Debugf("%s: unable to read must-gather information.", prefix)
				return res
			}
			if re.Provider.MustGatherInfo.ErrorEtcdLogs.FilterRequestSlowAll["all"] == nil {
				log.Debugf("%s: unable to read statistics from parsed etcd logs.", prefix)
				return res
			}
			if re.Provider.MustGatherInfo.ErrorEtcdLogs.FilterRequestSlowAll["all"].StatMean == "" {
				log.Debugf("%s: unable to get p50/mean statistics from parsed data: %v", prefix, re.Provider.MustGatherInfo.ErrorEtcdLogs.FilterRequestSlowAll["all"])
				return res
			}
			values := strings.Split(re.Provider.MustGatherInfo.ErrorEtcdLogs.FilterRequestSlowAll["all"].StatMean, " ")
			if values[0] == "" {
				log.Debugf("%s: unable to get parse p50/mean: %v", prefix, values)
				return res
			}
			value, err := strconv.ParseFloat(values[0], 64)
			if err != nil {
				log.Debugf("%s: unable to convert p50/mean to float: %v", prefix, err)
				return res
			}
			if value >= wantLimit {
				log.Debugf("%s acceptance criteria: want=[<%.0f] got=[%v]", prefix, wantLimit, value)
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	checkSum.Checks = append(checkSum.Checks, &Check{
		ID:   "OPCT-011",
		Name: "etcd logs: slow requests: maximum should be under 1000ms",
		Test: func() CheckResult {
			prefix := "Check OPCT-011 Failed"
			wantLimit := 1000.0
			res := CheckResult{Name: CheckResultNameFail, Want: fmt.Sprintf("<=%f ms", wantLimit)}
			if re.Provider.MustGatherInfo == nil {
				log.Debugf("%s: unable to read must-gather information.", prefix)
				return res
			}
			if re.Provider.MustGatherInfo.ErrorEtcdLogs.FilterRequestSlowAll["all"] == nil {
				log.Debugf("%s: unable to read statistics from parsed etcd logs.", prefix)
				return res
			}
			if re.Provider.MustGatherInfo.ErrorEtcdLogs.FilterRequestSlowAll["all"].StatMax == "" {
				log.Debugf("%s: unable to get max statistics from parsed data: %v", prefix, re.Provider.MustGatherInfo.ErrorEtcdLogs.FilterRequestSlowAll["all"])
				return res
			}
			values := strings.Split(re.Provider.MustGatherInfo.ErrorEtcdLogs.FilterRequestSlowAll["all"].StatMax, " ")
			if values[0] == "" {
				log.Debugf("%s: unable to get parse max: %v", prefix, values)
				return res
			}
			value, err := strconv.ParseFloat(values[0], 64)
			if err != nil {
				log.Debugf("%s: unable to convert max to float: %v", prefix, err)
				return res
			}
			if value >= wantLimit {
				log.Debugf("%s acceptance criteria: want=[<%.0f] got=[%v]", prefix, wantLimit, value)
				return res
			}
			res.Message = CheckResultNamePass
			return res
		},
	})
	// TODO(network): podConnectivityChecks must not have outages

	// Create docs reference when ID is set
	for c := range checkSum.Checks {
		if checkSum.Checks[c].ID != "" {
			checkSum.Checks[c].Reference = fmt.Sprintf("%s/#%s", checkSum.baseURL, checkSum.Checks[c].ID)
		}
	}
	return checkSum
}

func (csum *CheckSummary) GetBaseURL() string {
	return csum.baseURL
}

func (csum *CheckSummary) GetChecksFailed() []*Check {
	failures := []*Check{}
	for _, check := range csum.Checks {
		if check.Result.Name == CheckResultNameFail {
			failures = append(failures, check)
		}
	}
	return failures
}

func (csum *CheckSummary) GetChecksPassed() []*Check {
	failures := []*Check{}
	for _, check := range csum.Checks {
		if check.Result.Name == CheckResultNamePass {
			failures = append(failures, check)
		}
	}
	return failures
}

func (csum *CheckSummary) Run() error {
	for _, check := range csum.Checks {
		check.Result = check.Test()
	}
	return nil
}
