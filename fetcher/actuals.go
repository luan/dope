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

func (l *LRP) actualsSorted(sortOrder func([]*Actual) sort.Interface, reversed bool) []*Actual {
	actuals := make([]*Actual, len(l.Actuals))
	copy(actuals, l.Actuals)

	if reversed {
		sort.Sort(sort.Reverse(sortOrder(actuals)))
	} else {
		sort.Sort(sortOrder(actuals))
	}
	return actuals
}

func (l *LRP) ActualLRPsByIndex(reversed bool) []*Actual {
	return l.actualsSorted(byIndex, reversed)
}

func (l *LRP) ActualLRPsByCPU(reversed bool) []*Actual {
	return l.actualsSorted(byCPU, reversed)
}

func (l *LRP) ActualLRPsByMemory(reversed bool) []*Actual {
	return l.actualsSorted(byMemory, reversed)
}

func (l *LRP) ActualLRPsByDisk(reversed bool) []*Actual {
	return l.actualsSorted(byDisk, reversed)
}

func byIndex(actuals []*Actual) sort.Interface {
	return ActualsByIndex(actuals)
}

type ActualsByIndex []*Actual

func (l ActualsByIndex) Len() int      { return len(l) }
func (l ActualsByIndex) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ActualsByIndex) Less(i, j int) bool {
	return l[i].ActualLRP.Index < l[j].ActualLRP.Index
}

func byCPU(actuals []*Actual) sort.Interface {
	return ActualsByCPU(actuals)
}

type ActualsByCPU []*Actual

func (l ActualsByCPU) Len() int      { return len(l) }
func (l ActualsByCPU) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ActualsByCPU) Less(i, j int) bool {
	return l[i].Metrics.CPU < l[j].Metrics.CPU
}

func byMemory(actuals []*Actual) sort.Interface {
	return ActualsByMemory(actuals)
}

type ActualsByMemory []*Actual

func (l ActualsByMemory) Len() int      { return len(l) }
func (l ActualsByMemory) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ActualsByMemory) Less(i, j int) bool {
	return l[i].Metrics.Memory < l[j].Metrics.Memory
}

func byDisk(actuals []*Actual) sort.Interface {
	return ActualsByDisk(actuals)
}

type ActualsByDisk []*Actual

func (l ActualsByDisk) Len() int      { return len(l) }
func (l ActualsByDisk) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l ActualsByDisk) Less(i, j int) bool {
	return l[i].Metrics.Disk < l[j].Metrics.Disk
}
