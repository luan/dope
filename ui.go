package main

import (
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/gizak/termui"
	"github.com/luan/idope/fetcher"
)

type UI struct {
	selectedIndex int
	listContent   []string
	listOffset    int

	listWidget *termui.List
}

func NewUI() *UI {
	return &UI{}
}

func (ui *UI) Render() {
	ui.listWidget.Items = ui.selectItem()
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
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(9, 0, ui.listWidget),
		),
	)
	ui.listWidget.Height = termui.TermHeight()
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
		return fmt.Sprintf("[%-10s](fg-white)", state)
	case "CLAIMED":
		return fmt.Sprintf("[%-10s](fg-yellow)", state)
	case "RUNNING":
		return fmt.Sprintf("[%-10s](fg-green)", state)
	case "CRASHED":
		return fmt.Sprintf("[%-10s](fg-red)", state)
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

func lrpToStrings(lrp *fetcher.LRP) []string {
	ret := []string{}
	ret = append(ret,
		fmt.Sprintf(
			"guid: [%s](fg-bold)\t[instances:](fg-white) [%d](fg-white,fg-bold) ",
			lrp.Desired.ProcessGuid[:8], lrp.Desired.Instances,
		),
	)
	ret = append(ret,
		fmt.Sprintf(
			"    %s %s %s %s %s %s",
			"[index](fg-white,bg-reverse)",
			"[cell  ](fg-yellow,bg-reverse)",
			"[state     ](fg-white,bg-reverse)",
			"[cpu   ](fg-magenta,bg-reverse)",
			"[memory](fg-cyan,bg-reverse)[/total    ](fg-cyan,bg-reverse)",
			"[disk](fg-red,bg-reverse)[/total       ](fg-red,bg-reverse)",
		),
	)
	for _, actual := range lrp.ActualLRPsByCPU(true) {
		state := colorizeState(actual.ActualLRP.State)
		ret = append(ret,
			fmt.Sprintf(
				"    [%5d](fg-white) %-6s %s [%5.1f%%](fg-magenta) [%8s](fg-cyan)[/%-8s](fg-cyan,fg-bold) [%8s](fg-red)[/%-8s](fg-red,fg-bold)",
				actual.ActualLRP.Index, fmtCell(actual.ActualLRP.CellId), state,
				actual.Metrics.CPU*100,
				fmtBytes(actual.Metrics.Memory), fmtBytes(uint64(lrp.Desired.MemoryMb*1000*1000)),
				fmtBytes(actual.Metrics.Disk), fmtBytes(uint64(lrp.Desired.DiskMb*1000*1000)),
			))
	}
	return ret
}

func (ui *UI) SetState(state *fetcher.Data) {
	ui.listContent = []string{}
	lrps := state.LRPs.SortedByProcessGuid()
	for _, lrp := range lrps {
		content := lrpToStrings(lrp)
		ui.listContent = append(ui.listContent, content...)
	}
	ui.Render()
}

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

	termui.Handle("/sys/wnd/resize", func(termui.Event) {
		ui.listWidget.Height = termui.TermHeight()
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

func (ui *UI) selectItem() []string {
	if ui.listWidget.InnerHeight() == 0 {
		return []string{}
	}
	index := ui.selectedIndex
	visibleContent := ui.listContent[ui.listOffset:]

	ret := make([]string, len(visibleContent))
	for i, item := range visibleContent {
		if i == index {
			ret[i] = fmt.Sprintf(" [➤](fg-cyan,fg-bold) %s", item)
		} else {
			ret[i] = fmt.Sprintf("   %s", item)
		}
	}
	return ret
}
