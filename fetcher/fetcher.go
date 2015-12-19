package fetcher

import (
	"os"
	"sort"
	"sync"

	"github.com/cloudfoundry-incubator/bbs"
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry/noaa"
)

type Fetcher interface {
	Fetch() (Data, error)
}

type fetcher struct {
	bbsClient  bbs.Client
	noaaClient *noaa.Consumer
}

type Data struct {
	Domains []string
	Tasks   []*models.Task
	LRPs    LRPs
}

func NewFetcher(bbsClient bbs.Client, noaaClient *noaa.Consumer) Fetcher {
	return &fetcher{
		bbsClient:  bbsClient,
		noaaClient: noaaClient,
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

	wg := sync.WaitGroup{}
	authToken := os.Getenv("OAUTH_TOKEN")
	for _, lrp := range lrps {
		lrp := lrp
		wg.Add(1)
		go func() {
			defer wg.Done()
			containerMetrics, err := f.noaaClient.ContainerMetrics(lrp.Desired.LogGuid, authToken)
			if err != nil {
				return
			}

			for _, metrics := range containerMetrics {
				for _, actual := range lrp.Actuals {
					if actual.ActualLRP.Index == metrics.GetInstanceIndex() {
						containerMetrics := ContainerMetrics{
							CPU:    metrics.GetCpuPercentage(),
							Memory: metrics.GetMemoryBytes(),
							Disk:   metrics.GetDiskBytes(),
						}

						actual.Metrics = containerMetrics
					}
				}
			}
		}()
	}
	wg.Wait()

	return lrps, nil
}

type CellState struct {
	MemoryUsed     uint64
	MemoryReserved uint64

	DiskUsed     uint64
	DiskReserved uint64

	CPUPercentage float64

	NumLRPs  uint64
	NumTasks uint64

	CellId string
}

func (d *Data) GetCellState() CellStates {
	cellStates := map[string]*CellState{}

	for _, lrp := range d.LRPs {
		for _, actual := range lrp.Actuals {
			if actual.ActualLRP.CellId != "" {
				cellState, ok := cellStates[actual.ActualLRP.CellId]
				if !ok {
					cellState = &CellState{CellId: actual.ActualLRP.CellId}
					cellStates[cellState.CellId] = cellState
				}

				cellState.NumLRPs++
				cellState.CPUPercentage += actual.Metrics.CPU
				cellState.MemoryUsed += actual.Metrics.Memory
				cellState.MemoryReserved += uint64(lrp.Desired.MemoryMb * 1024 * 1024)
				cellState.DiskUsed += actual.Metrics.Disk
				cellState.DiskReserved += uint64(lrp.Desired.DiskMb * 1024 * 1024)
			}
		}
	}

	for _, task := range d.Tasks {
		if task.CellId != "" {
			cellState, ok := cellStates[task.CellId]
			if !ok {
				cellState = &CellState{CellId: task.CellId}
				cellStates[cellState.CellId] = cellState
			}

			cellState.NumTasks++
		}
	}

	return cellStates
}

type CellStates map[string]*CellState

func (l CellStates) SortedByCellId() []*CellState {
	var cellStates []*CellState

	for _, state := range l {
		cellStates = append(cellStates, state)
	}

	sort.Sort(ByCellId(cellStates))
	return cellStates
}

func ByCellId(cellStates []*CellState) CellStatesByCellId {
	return cellStates
}

type CellStatesByCellId []*CellState

func (l CellStatesByCellId) Len() int      { return len(l) }
func (l CellStatesByCellId) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l CellStatesByCellId) Less(i, j int) bool {
	return l[i].CellId < l[j].CellId
}
