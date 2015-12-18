package fetcher

import (
	"fmt"
	"sort"
	"strings"

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

func (l *LRP) StartCommand() string {
	return printAction(l.Desired.Action)
}

func printAction(action *models.Action) string {
	unwrappedAction := models.UnwrapAction(action)
	switch unwrappedAction.ActionType() {
	case models.ActionTypeDownload:
		download := unwrappedAction.(*models.DownloadAction)
		return fmt.Sprintf("Download: %s to %s as %s", download.From, download.To, download.User)
	case models.ActionTypeEmitProgress:
		emit := unwrappedAction.(*models.EmitProgressAction)
		return printAction(emit.Action)
	case models.ActionTypeRun:
		run := unwrappedAction.(*models.RunAction)
		return fmt.Sprintf("Run: %s %s", run.Path, strings.Join(run.Args, " "))
	case models.ActionTypeUpload:
		upload := unwrappedAction.(*models.UploadAction)
		return fmt.Sprintf("Upload: %s to %s", upload.From, upload.To)
	case models.ActionTypeTimeout:
		timeout := unwrappedAction.(*models.TimeoutAction)
		return printAction(timeout.Action)
	case models.ActionTypeTry:
		try := unwrappedAction.(*models.TryAction)
		return printAction(try.Action)
	case models.ActionTypeParallel:
		paralell := unwrappedAction.(*models.ParallelAction)
		actionText := "Parallel:\n"
		for _, action := range paralell.Actions {
			actionText += fmt.Sprintf("\t%s\n", printAction(action))
		}
		return actionText
	case models.ActionTypeSerial:
		serial := unwrappedAction.(*models.SerialAction)
		actionText := "Serial:\n"
		for _, action := range serial.Actions {
			actionText += fmt.Sprintf("\t%s\n", printAction(action))
		}
		return actionText
	case models.ActionTypeCodependent:
		codependent := unwrappedAction.(*models.CodependentAction)
		actionText := "Codependent:\n"
		for _, action := range codependent.Actions {
			actionText += fmt.Sprintf("\t%s\n", printAction(action))
		}
		return actionText
	default:
		return "Error: Unknown start command"
	}
}
