package fetcher

import (
	"sort"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
)

type Fetcher interface {
	Fetch() (Data, error)
}

type fetcher struct {
	bbsClient bbs.Client
}

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
}

func NewFetcher(bbsClient bbs.Client) Fetcher {
	return &fetcher{
		bbsClient: bbsClient,
	}
}

func (f *fetcher) Fetch() (Data, error) {
	domains, err := f.bbsClient.Domains()
	if err != nil {
		return Data{}, err
	}

	tasks, err := f.bbsClient.Tasks()
	if err != nil {
		return Data{}, err
	}

	lrps, err := f.fetchLRPs()
	if err != nil {
		return Data{}, err
	}

	return Data{
		Domains: domains,
		Tasks:   tasks,
		LRPs:    lrps,
	}, nil
}

func (f *fetcher) fetchLRPs() (map[string]*LRP, error) {
	lrps := map[string]*LRP{}

	desiredLRPs, err := f.bbsClient.DesiredLRPs(models.DesiredLRPFilter{})
	if err != nil {
		return nil, err
	}

	for _, desiredLRP := range desiredLRPs {
		lrps[desiredLRP.ProcessGuid] = &LRP{Desired: desiredLRP}
	}

	actualLRPGroups, err := f.bbsClient.ActualLRPGroups(models.ActualLRPFilter{})
	if err != nil {
		return nil, err
	}

	for _, actualLRPGroup := range actualLRPGroups {
		actualLRP, evacuating := actualLRPGroup.Resolve()

		lrp, ok := lrps[actualLRP.ProcessGuid]
		if !ok {
			continue
		}

		actual := Actual{ActualLRP: actualLRP, Evacuating: evacuating}
		lrp.Actuals = append(lrp.Actuals, &actual)
	}

	return lrps, nil
}
