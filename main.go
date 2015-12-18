package main

import (
	"fmt"
	"runtime"

	ui "github.com/gizak/termui"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var err error
	// bbsClient, err := config_finder.NewBBS()
	// if err != nil {
	// 	panic(err)
	// }

	err = ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	// lrps, err := bbsClient.DesiredLRPs(models.DesiredLRPFilter{})

	strs := []string{}
	strs = []string{
		"string 01",
		"string 02",
		"string 03",
		"string 04",
		"string 05",
		"string 06",
		"string 07",
		"string 08",
		"string 09",
		"string 10",
		"string 11",
		"string 12",
		"string 13",
	}
	// for _, lrp := range lrps {
	// 	strs = append(strs, lrp.ProcessGuid)
	// }

	currentIndex := 0

	ls := ui.NewList()
	ls.Items = selectItem(strs, currentIndex)
	ls.ItemFgColor = ui.ColorYellow
	ls.BorderLabel = "List"
	ls.Y = 0
	ls.Height = ui.TermHeight()

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(6, 0, ls),
		),
	)

	ui.Body.Align()
	ui.Render(ui.Body)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Handle("/sys/kbd/j", func(ui.Event) {
		if currentIndex < len(strs)-1 {
			currentIndex++
		}
		ls.Items = selectItem(strs, currentIndex)
		ui.Render(ls)
	})

	ui.Handle("/sys/kbd/k", func(ui.Event) {
		if currentIndex > 0 {
			currentIndex--
		}
		ls.Items = selectItem(strs, currentIndex)
		ui.Render(ls)
	})

	ui.Handle("/sys/wnd/resize", func(ui.Event) {
		ls.Height = ui.TermHeight()
		ui.Render(ls)
	})

	ui.Loop()
}

func selectItem(strs []string, index int) []string {
	ret := make([]string, len(strs))
	copy(ret[:], strs)
	ret[index] = fmt.Sprintf("[%s](fg-black,bg-white)", ret[index])
	return ret
}
