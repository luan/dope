package main

import (
	"fmt"

	"github.com/gizak/termui"
	"github.com/luan/idope/fetcher"
)

type UI struct {
	selectedIndex  int
	listContent    []string
	visibleContent []string

	listWidget *termui.List
}

func NewUI() *UI {
	return &UI{}
}

func (ui *UI) Render() {
	ui.visibleContent = ui.listContent
	ui.listWidget.Items = ui.selectItem()
	ui.listWidget.Height = termui.TermHeight()
	termui.Body.Align()
	termui.Render(termui.Body)
}

func (ui *UI) Setup() {
	err := termui.Init()
	if err != nil {
		panic(err)
	}

	list := termui.NewList()
	// list.Border = false
	list.ItemFgColor = termui.ColorYellow
	ui.listWidget = list
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(9, 0, ui.listWidget),
		),
	)

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
		return "[UNCLAIMED](fg-white)"
	case "CLAIMED":
		return "[CLAIMED](fg-yellow)"
	case "RUNNING":
		return "[RUNNING](fg-green)"
	case "CRASHED":
		return "[CRASHED](fg-red)"
	default:
		return state
	}
}

func lrpToStrings(lrp *fetcher.LRP) []string {
	ret := []string{}
	ret = append(ret, fmt.Sprintf("guid: [%s](fg-bold)\t[instances:](fg-white) [%d](fg-white,fg-bold) ", lrp.Desired.ProcessGuid[:8], lrp.Desired.Instances))
	for _, actual := range lrp.ActualLRPsByIndex() {
		state := colorizeState(actual.ActualLRP.State)
		ret = append(ret, fmt.Sprintf("\t[%d](fg-white) %s %s ", actual.ActualLRP.Index, actual.ActualLRP.CellId, state))
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

	termui.Handle("/sys/wnd/resize", func(termui.Event) {
		ui.Render()
	})
}

func (ui *UI) handleDown(_ termui.Event) {
	if ui.selectedIndex < len(ui.listContent)-1 {
		ui.selectedIndex++
	}
	ui.Render()
}

func (ui *UI) handleUp(_ termui.Event) {
	if ui.selectedIndex > 0 {
		ui.selectedIndex--
	}
	ui.Render()
}

func (ui *UI) selectItem() []string {
	ret := make([]string, len(ui.visibleContent))
	for i, item := range ui.visibleContent {
		if i == ui.selectedIndex {
			ret[i] = fmt.Sprintf(" [➤](fg-cyan,fg-bold) %s", item)
		} else {
			ret[i] = fmt.Sprintf("   %s", item)
		}
	}
	return ret
}
