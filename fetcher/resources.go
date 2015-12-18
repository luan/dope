package fetcher

import (
	"sort"

	"github.com/cloudfoundry-incubator/bbs/models"
)

type Data struct {
	Domains []string
	Tasks   []*models.Task
	LRPs    LRPs
}

type LRPs map[string]*LRP

func (l LRPs) SortedByProcessGuid() []*LRP {
	var lrps []*LRP

	for _, lrp := range l {
		lrps = append(lrps, lrp)
	}

	sort.Sort(ByProcessGuid(lrps))
	return lrps
}

func ByProcessGuid(lrps []*LRP) LRPsByProcessGuid {
	return lrps
}

type LRPsByProcessGuid []*LRP

func (l LRPsByProcessGuid) Len() int      { return len(l) }
func (l LRPsByProcessGuid) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l LRPsByProcessGuid) Less(i, j int) bool {
	return l[i].Desired.ProcessGuid < l[j].Desired.ProcessGuid
}

type LRP struct {
	Desired *models.DesiredLRP
	Actuals []*Actual
}

func (l *LRP) ActualLRPsByIndex() []*Actual {
	actuals := make([]*Actual, len(l.Actuals))
	copy(actuals, l.Actuals)

	sort.Sort(ByIndex(actuals))
	return actuals
}

func ByIndex(actuals []*Actual) ActualsByIndex {
	return actuals
}

type ActualsByIndex []*Actual

func (l ActualsByIndex) Len() int      { return len(l) }
func (l ActualsByIndex) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ActualsByIndex) Less(i, j int) bool {
	return l[i].ActualLRP.Index < l[j].ActualLRP.Index
}

type Actual struct {
	Evacuating bool
	ActualLRP  *models.ActualLRP
	Metrics    ContainerMetrics
}

type ContainerMetrics struct {
	CPU    float64
	Memory uint64
	Disk   uint64
}
