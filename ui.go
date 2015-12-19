package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry-incubator/bbs/models"
	humanize "github.com/dustin/go-humanize"
	"github.com/gizak/termui"
	"github.com/luan/idope/fetcher"
)

type content struct {
	String string

	lrp     *fetcher.LRP
	desired *models.DesiredLRP
	actual  *fetcher.Actual
}

type UI struct {
	selectedIndex int
	listContent   []content
	listOffset    int
	sort          int
	sortReverse   bool
	state         *fetcher.Data

	listWidget    *termui.List
	detailWidget  *termui.Par
	summaryWidget *termui.Par
	cellsWidget   *termui.Par
}

func NewUI() *UI {
	return &UI{}
}

func (ui *UI) Render() {
	var selected content
	ui.listWidget.Items, selected = ui.selectItem()
	text := ""
	if selected.desired != nil {
		routes, _ := json.Marshal((*selected.desired.Routes)["cf-router"])
		ui.detailWidget.BorderLabel = "Desired LRP"
		text = fmt.Sprintf(
			`[guid:](fg-bold) %s
[start command:](fg-bold)
%s
[routes:](fg-bold) %s
`,
			selected.desired.ProcessGuid,
			selected.lrp.StartCommand(),
			string(routes),
		)
	}
	if selected.actual != nil {
		ui.detailWidget.BorderLabel = "Actual LRP"
		ports, _ := json.Marshal(selected.actual.ActualLRP.Ports)
		text = fmt.Sprintf(
			`[guid:](fg-bold) %s
[cell:](fg-bold) %s
[instance index:](fg-bold) %d
[state:](fg-bold) %s
[address:](fg-bold) %s
[ports:](fg-bold) %s
[cpu load:](fg-bold) %.1f
[memory usage:](fg-bold) %s/%s
[disk usage:](fg-bold) %s/%s
[crash reason:](fg-bold) %s
`,
			selected.actual.ActualLRP.ProcessGuid,
			fmtCell(selected.actual.ActualLRP.CellId),
			selected.actual.ActualLRP.Index,
			colorizeState(selected.actual.ActualLRP.State),
			selected.actual.ActualLRP.Address,
			ports,
			selected.actual.Metrics.CPU*100,
			fmtBytes(selected.actual.Metrics.Memory), fmtBytes(uint64(selected.lrp.Desired.MemoryMb*1000*1000)),
			fmtBytes(selected.actual.Metrics.Disk), fmtBytes(uint64(selected.lrp.Desired.DiskMb*1000*1000)),
			selected.actual.ActualLRP.CrashReason,
		)
	}
	dat, _ := ioutil.ReadFile("gopher.txt")
	gopher := string(dat)
	ui.detailWidget.Text = text + "\n\n\n\n\n" + gopher
	var totalCPU float64
	var totalMemoryUsed uint64
	var totalMemoryReserved uint64
	var totalDiskUsed uint64
	var totalDiskReserved uint64
	var totalLRPs uint64
	var totalTasks uint64
	ui.cellsWidget.Text = ""

	if ui.state != nil && len(ui.state.LRPs) > 0 {
		cells := ui.state.GetCellState().SortedByCellId()
		totalCells := len(cells)

		for _, cell := range cells {
			totalCPU += cell.CPUPercentage
			totalMemoryUsed += cell.MemoryUsed
			totalMemoryReserved += cell.MemoryReserved
			totalDiskUsed += cell.DiskUsed
			totalDiskReserved += cell.DiskReserved
			totalLRPs += cell.NumLRPs
			totalTasks += cell.NumTasks

			ui.cellsWidget.Text += fmt.Sprintf(
				`[%8s:](fg-white,fg-bold) | [LRPs:](fg-white,fg-bold) %3d | [Tasks:](fg-white,fg-bold) %3d | [Average CPU:](fg-white,fg-bold) %8.1f%% | [Total Memory:](fg-white,fg-bold) %8s/%-8s | [Total Disk:](fg-white,fg-bold) %8s/%-8s
`, cell.CellId, cell.NumLRPs, cell.NumTasks,
				float64(100)*cell.CPUPercentage,
				fmtBytes(cell.MemoryUsed), fmtBytes(cell.MemoryReserved),
				fmtBytes(cell.DiskUsed), fmtBytes(cell.DiskReserved),
			)
		}

		ui.summaryWidget.Text = fmt.Sprintf(
			`[Cells:](fg-white,fg-bold) %d
[LRPs:](fg-white,fg-bold) %d
[Tasks:](fg-white,fg-bold) %d
[Average CPU:](fg-white,fg-bold) %.1f%%
[Total Memory:](fg-white,fg-bold) %s/%s
[Total Disk:](fg-white,fg-bold) %s/%s
`, totalCells, totalLRPs, totalTasks,
			float64(100)*totalCPU/float64(totalCells),
			fmtBytes(totalMemoryUsed), fmtBytes(totalMemoryReserved),
			fmtBytes(totalDiskUsed), fmtBytes(totalDiskReserved),
		)

	}
	termui.Render(termui.Body)
}

func (ui *UI) Setup() {
	err := termui.Init()
	if err != nil {
		panic(err)
	}

	list := termui.NewList()
	list.BorderLabel = "LRPs"
	list.ItemFgColor = termui.ColorYellow
	ui.listWidget = list
	ui.detailWidget = termui.NewPar("")
	ui.summaryWidget = termui.NewPar("")
	ui.summaryWidget.BorderLabel = "Summary"
	ui.summaryWidget.Height = 12
	ui.cellsWidget = termui.NewPar("")
	ui.cellsWidget.BorderLabel = "Cells"
	ui.cellsWidget.Height = 12

	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(3, 0, ui.summaryWidget),
			termui.NewCol(9, 0, ui.cellsWidget),
		),
		termui.NewRow(
			termui.NewCol(6, 0, ui.listWidget),
			termui.NewCol(6, 0, ui.detailWidget),
		),
	)
	ui.listWidget.Height = termui.TermHeight() - 12
	ui.detailWidget.Height = termui.TermHeight() - 12
	termui.Body.Align()

	ui.Render()
	ui.bindEvents()
}

func (ui *UI) Close() {
	termui.Close()
}

func (ui *UI) Loop() {
	termui.Loop()
}

func colorizeState(state string) string {
	switch state {
	case "UNCLAIMED":
		return fmt.Sprintf("[%-11s](fg-white)", state)
	case "CLAIMED":
		return fmt.Sprintf("[%-11s](fg-yellow)", state)
	case "RUNNING":
		return fmt.Sprintf("[%-11s](fg-green)", state)
	case "CRASHED":
		return fmt.Sprintf("[%-11s](fg-red)", state)
	default:
		return state
	}
}

func fmtBytes(s uint64) string {
	return strings.Replace(humanize.Bytes(s), " ", "", -1)
}

func fmtCell(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) != 2 {
		return "none"
	}
	parts = strings.Split(parts[1], "-")
	return fmt.Sprintf("%s/%s", parts[0], parts[1])
}

func fmtSort(s string, sort int, reverse bool) string {
	if sortOptions[sort] != s {
		return " "
	}
	if reverse {
		return "▾"
	}
	return "▴"
}

func (ui *UI) lrpToContents(lrp *fetcher.LRP) []content {
	ret := []content{}
	ret = append(ret,
		content{
			String: fmt.Sprintf(
				"guid: [%s](fg-bold)\t[instances:](fg-white) [%d](fg-white,fg-bold) ",
				lrp.Desired.ProcessGuid[:8], lrp.Desired.Instances,
			),
			lrp:     lrp,
			desired: lrp.Desired,
		},
	)
	ret = append(ret,
		content{
			String: fmt.Sprintf(
				"    %s %s %s %s %s %s",
				fmt.Sprintf("[ index %s ](fg-white,bg-reverse)", fmtSort("index", ui.sort, ui.sortReverse)),
				fmt.Sprintf("[ cell %s ](fg-yellow,bg-reverse)", fmtSort("cell", ui.sort, ui.sortReverse)),
				fmt.Sprintf("[ state   %s ](fg-white,bg-reverse)", fmtSort("state", ui.sort, ui.sortReverse)),
				fmt.Sprintf("[ cpu %s ](fg-magenta,bg-reverse)", fmtSort("cpu", ui.sort, ui.sortReverse)),
				fmt.Sprintf("[ memory](fg-cyan,bg-reverse)[/total  %s ](fg-cyan,bg-reverse)", fmtSort("memory", ui.sort, ui.sortReverse)),
				fmt.Sprintf("[ disk](fg-red,bg-reverse)[/total     %s ](fg-red,bg-reverse)", fmtSort("disk", ui.sort, ui.sortReverse)),
			),
		},
	)
	var lrps []*fetcher.Actual
	switch ui.sort {
	case 0:
		lrps = lrp.ActualLRPsByIndex(ui.sortReverse)
	case 1:
		lrps = lrp.ActualLRPsByCPU(ui.sortReverse)
	case 2:
		lrps = lrp.ActualLRPsByMemory(ui.sortReverse)
	case 3:
		lrps = lrp.ActualLRPsByDisk(ui.sortReverse)
	}
	for _, actual := range lrps {
		state := colorizeState(actual.ActualLRP.State)
		ret = append(ret,
			content{
				String: fmt.Sprintf(
					"    [%9d](fg-white) %-8s %s [%6.1f%%](fg-magenta) [%9s](fg-cyan)[/%-8s](fg-cyan,fg-bold) [%9s](fg-red)[/%-8s](fg-red,fg-bold)",
					actual.ActualLRP.Index, fmtCell(actual.ActualLRP.CellId), state,
					actual.Metrics.CPU*100,
					fmtBytes(actual.Metrics.Memory), fmtBytes(uint64(lrp.Desired.MemoryMb*1000*1000)),
					fmtBytes(actual.Metrics.Disk), fmtBytes(uint64(lrp.Desired.DiskMb*1000*1000)),
				),
				lrp:    lrp,
				actual: actual,
			},
		)
	}
	return ret
}

func (ui *UI) refreshState() {
	ui.listContent = []content{}
	lrps := ui.state.LRPs.SortedByProcessGuid()
	for _, lrp := range lrps {
		content := ui.lrpToContents(lrp)
		ui.listContent = append(ui.listContent, content...)
	}
}

func (ui *UI) SetState(state *fetcher.Data) {
	ui.state = state
	ui.refreshState()
	ui.Render()
}

func (ui *UI) reverseSort() {
	ui.sortReverse = !ui.sortReverse
}

func (ui *UI) setSort(delta int) {
	ui.sort += delta
	if ui.sort < 0 {
		ui.sort = 0
	} else if ui.sort > 3 {
		ui.sort = 3
	}
}

var sortOptions = []string{"index", "cpu", "memory", "disk"}

func (ui *UI) bindEvents() {
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("/sys/kbd/j", ui.handleDown)
	termui.Handle("/sys/kbd/<down>", ui.handleDown)
	termui.Handle("/sys/kbd/k", ui.handleUp)
	termui.Handle("/sys/kbd/<up>", ui.handleUp)
	termui.Handle("/sys/kbd/g", ui.handleTop)
	termui.Handle("/sys/kbd/<home>", ui.handleTop)
	termui.Handle("/sys/kbd/G", ui.handleBottom)
	termui.Handle("/sys/kbd/<end>", ui.handleBottom)

	termui.Handle("/sys/kbd/<right>", func(termui.Event) {
		ui.setSort(1)
		ui.refreshState()
		ui.Render()
	})

	termui.Handle("/sys/kbd/l", func(termui.Event) {
		ui.setSort(1)
		ui.refreshState()
		ui.Render()
	})

	termui.Handle("/sys/kbd/<left>", func(termui.Event) {
		ui.setSort(-1)
		ui.refreshState()
		ui.Render()
	})

	termui.Handle("/sys/kbd/h", func(termui.Event) {
		ui.setSort(-1)
		ui.refreshState()
		ui.Render()
	})

	termui.Handle("/sys/kbd/s", func(termui.Event) {
		ui.reverseSort()
		ui.refreshState()
		ui.Render()
	})

	termui.Handle("/sys/wnd/resize", func(termui.Event) {
		ui.listWidget.Height = termui.TermHeight() - 12
		ui.detailWidget.Height = termui.TermHeight() - 12
		termui.Body.Align()
		ui.Render()
	})
}

func (ui *UI) handleBottom(_ termui.Event) {
	visibleHeight := ui.listWidget.InnerHeight()
	totalHeight := len(ui.listContent)
	if totalHeight < visibleHeight {
		visibleHeight = totalHeight
	}
	ui.selectedIndex = visibleHeight - 1
	ui.listOffset = totalHeight - visibleHeight
	ui.Render()
}

func (ui *UI) handleTop(_ termui.Event) {
	ui.selectedIndex = 0
	ui.listOffset = 0
	ui.Render()
}

func (ui *UI) handleDown(_ termui.Event) {
	visibleHeight := ui.listWidget.InnerHeight()
	totalHeight := len(ui.listContent)
	if totalHeight < visibleHeight {
		visibleHeight = totalHeight
	}

	if ui.selectedIndex < visibleHeight-1 {
		ui.selectedIndex++
	} else if ui.listOffset < totalHeight-visibleHeight-1 {
		ui.listOffset++
	}
	ui.Render()
}

func (ui *UI) handleUp(_ termui.Event) {
	if ui.selectedIndex > 0 {
		ui.selectedIndex--
	} else if ui.listOffset > 0 {
		ui.listOffset--
	}
	ui.Render()
}

func (ui *UI) selectItem() ([]string, content) {
	if ui.listWidget.InnerHeight() == 0 {
		return []string{}, content{}
	}
	index := ui.selectedIndex
	visibleContent := ui.listContent[ui.listOffset:]

	ret := make([]string, len(visibleContent))
	for i, item := range visibleContent {
		if i == index {
			ret[i] = fmt.Sprintf(" [❯](fg-cyan,fg-bold) %s", item.String)
		} else {
			ret[i] = fmt.Sprintf("   %s", item.String)
		}
	}
	return ret, visibleContent[index]
}
