package fetcher

import (
	"sort"

	"github.com/cloudfoundry-incubator/bbs/models"
)

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
