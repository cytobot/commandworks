package commandworks

import (
	"bytes"
	"fmt"
	"text/tabwriter"
	"time"

	pbs "github.com/cytobot/messaging/transport/shared"
	"github.com/dustin/go-humanize"
	"github.com/lampjaw/discordclient"
)

type statsCommandProcessor struct {
	managerClient *managerClient
}

func newStatsCommandProcessor(managerClient *managerClient) *statsCommandProcessor {
	return &statsCommandProcessor{
		managerClient: managerClient,
	}
}

func (p *statsCommandProcessor) getCommandID() string {
	return "stats"
}

func (p *statsCommandProcessor) process(client *discordclient.DiscordClient, req *pbs.DiscordWorkRequest) error {
	w := &tabwriter.Writer{}
	buf := &bytes.Buffer{}

	w.Init(buf, 0, 4, 0, ' ', 0)

	fmt.Fprintf(w, "\nListeners:\n")
	p.prepareListenerStats(w)

	fmt.Fprintf(w, "\nWorkers:\n")
	p.prepareWorkerStats(w)

	fmt.Fprintf(w, "\n```")

	w.Flush()
	out := buf.String()

	end := ""

	if end != "" {
		out += "\n" + end
	}

	client.SendMessage(req.ChannelID, out)

	return nil
}

func (p *statsCommandProcessor) prepareWorkerStats(w *tabwriter.Writer) {
	statuses, _ := p.managerClient.getWorkerStatus()

	sumDuration := 0
	var sumAllocMem uint64 = 0
	var sumSysMem uint64 = 0
	var sumTotalAllocMem uint64 = 0
	sumTasks := 0
	for _, s := range statuses {
		sumDuration += int(s.Uptime)
		sumAllocMem += uint64(s.MemAllocated)
		sumSysMem += uint64(s.MemSystem)
		sumTotalAllocMem += uint64(s.MemCumulative)
		sumTasks += int(s.TaskCount)
	}

	var avgDuration int
	if len(statuses) > 0 {
		avgDuration = sumDuration / len(statuses)
	}

	fmt.Fprintf(w, "```\n")
	fmt.Fprintf(w, "Uptime: \t%s\n", getDurationString(avgDuration))
	fmt.Fprintf(w, "Memory used: \t%s / %s (%s garbage collected)\n", humanize.Bytes(sumAllocMem), humanize.Bytes(sumSysMem), humanize.Bytes(sumTotalAllocMem))
	fmt.Fprintf(w, "Concurrent tasks: \t%d\n", sumTasks)
	fmt.Fprintf(w, "```\n")
}

func (p *statsCommandProcessor) prepareListenerStats(w *tabwriter.Writer) {
	statuses, _ := p.managerClient.getListenerStatus()

	totalServers := 0
	totalUsers := 0
	sumDuration := 0
	var sumAllocMem uint64 = 0
	var sumSysMem uint64 = 0
	var sumTotalAllocMem uint64 = 0
	sumTasks := 0
	for _, s := range statuses {
		totalServers += int(s.ConnectedServers)
		totalUsers += int(s.ConnectedUsers)
		sumDuration += int(s.Uptime)
		sumAllocMem += uint64(s.MemAllocated)
		sumSysMem += uint64(s.MemSystem)
		sumTotalAllocMem += uint64(s.MemCumulative)
		sumTasks += int(s.TaskCount)
	}

	var avgDuration int
	if len(statuses) > 0 {
		avgDuration = sumDuration / len(statuses)
	}

	fmt.Fprintf(w, "```\n")
	fmt.Fprintf(w, "Uptime: \t%s\n", getDurationString(avgDuration))
	fmt.Fprintf(w, "Memory used: \t%s / %s (%s garbage collected)\n", humanize.Bytes(sumAllocMem), humanize.Bytes(sumSysMem), humanize.Bytes(sumTotalAllocMem))
	fmt.Fprintf(w, "Concurrent tasks: \t%d\n", sumTasks)
	fmt.Fprintf(w, "Connected servers: \t%d\n", totalServers)
	fmt.Fprintf(w, "Connected users: \t%d\n", totalUsers)
	fmt.Fprintf(w, "```\n")
}

func getDurationString(val int) string {
	duration, _ := time.ParseDuration(fmt.Sprint(val))
	return fmt.Sprintf(
		"%0.2d:%02d:%02d",
		int(duration.Hours()),
		int(duration.Minutes())%60,
		int(duration.Seconds())%60,
	)
}
