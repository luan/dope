package fetcher

import (
	"sort"

	"github.com/cloudfoundry-incubator/bbs/models"
)

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

func (l *LRP) ActualLRPsByIndex() []*Actual {
	actuals := make([]*Actual, len(l.Actuals))
	copy(actuals, l.Actuals)

	sort.Sort(byIndex(actuals))
	return actuals
}

func (l *LRP) ActualLRPsByCPU() []*Actual {
	actuals := make([]*Actual, len(l.Actuals))
	copy(actuals, l.Actuals)

	sort.Sort(byCPU(actuals))
	return actuals
}

func (l *LRP) ActualLRPsByMemory() []*Actual {
	actuals := make([]*Actual, len(l.Actuals))
	copy(actuals, l.Actuals)

	sort.Sort(byMemory(actuals))
	return actuals
}

func (l *LRP) ActualLRPsByDisk() []*Actual {
	actuals := make([]*Actual, len(l.Actuals))
	copy(actuals, l.Actuals)

	sort.Sort(byDisk(actuals))
	return actuals
}

func byIndex(actuals []*Actual) ActualsByIndex {
	return actuals
}

type ActualsByIndex []*Actual

func (l ActualsByIndex) Len() int      { return len(l) }
func (l ActualsByIndex) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ActualsByIndex) Less(i, j int) bool {
	return l[i].ActualLRP.Index < l[j].ActualLRP.Index
}

func byCPU(actuals []*Actual) ActualsByCPU {
	return actuals
}

type ActualsByCPU []*Actual

func (l ActualsByCPU) Len() int      { return len(l) }
func (l ActualsByCPU) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ActualsByCPU) Less(i, j int) bool {
	return l[i].Metrics.CPU < l[j].Metrics.CPU
}

func byMemory(actuals []*Actual) ActualsByMemory {
	return actuals
}

type ActualsByMemory []*Actual

func (l ActualsByMemory) Len() int      { return len(l) }
func (l ActualsByMemory) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ActualsByMemory) Less(i, j int) bool {
	return l[i].Metrics.Memory < l[j].Metrics.Memory
}

func byDisk(actuals []*Actual) ActualsByDisk {
	return actuals
}

type ActualsByDisk []*Actual

func (l ActualsByDisk) Len() int      { return len(l) }
func (l ActualsByDisk) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ActualsByDisk) Less(i, j int) bool {
	return l[i].Metrics.Disk < l[j].Metrics.Disk
}
