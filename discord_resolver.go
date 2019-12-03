package commandworks

import (
	"encoding/json"
	"time"

	"github.com/bwmarrin/discordgo"
	cytonats "github.com/cytobot/messaging/nats"
	pbd "github.com/cytobot/messaging/transport/discord"
	pbw "github.com/cytobot/messaging/transport/worker"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	nats "github.com/nats-io/nats.go"
)

type discordResolver struct {
	nats *cytonats.NatsClient
}

func newDiscordResolver(natsClient *cytonats.NatsClient) *discordResolver {
	return &discordResolver{
		nats: natsClient,
	}
}

func (r *discordResolver) ResolveChannel(sourceID string, channelID string) (*discordgo.Channel, error) {
	msg, err := r.sendRequest(sourceID, "channel", map[string]string{
		"channelID": channelID,
	})
	if err != nil {
		return nil, err
	}

	res := r.getResult(msg)

	rawIn := json.RawMessage(res["channel"])
	bytes, err := rawIn.MarshalJSON()

	channel := &discordgo.Channel{}
	json.Unmarshal(bytes, &channel)

	return channel, nil
}

func (r *discordResolver) ResolveGuild(sourceID string, guildID string) (*discordgo.Guild, error) {
	msg, err := r.sendRequest(sourceID, "guild", map[string]string{
		"guildID": guildID,
	})
	if err != nil {
		return nil, err
	}

	res := r.getResult(msg)

	rawIn := json.RawMessage(res["guild"])
	bytes, err := rawIn.MarshalJSON()

	guild := &discordgo.Guild{}
	json.Unmarshal(bytes, &guild)

	return guild, nil
}

func (r *discordResolver) ResolveUser(sourceID string, userID string, channelID string) (*discordgo.User, error) {
	msg, err := r.sendRequest(sourceID, "user", map[string]string{
		"userID":    userID,
		"channelID": channelID,
	})
	if err != nil {
		return nil, err
	}

	res := r.getResult(msg)

	rawIn := json.RawMessage(res["user"])
	bytes, err := rawIn.MarshalJSON()

	user := &discordgo.User{}
	json.Unmarshal(bytes, &user)

	return user, nil
}

func (r *discordResolver) getResult(msg *nats.Msg) map[string]string {
	response := &pbd.DiscordInformationResponse{}
	json.Unmarshal(msg.Data, response)
	return response.Payload
}

func (r *discordResolver) sendRequest(sourceID string, requestType string, payload map[string]string) (*nats.Msg, error) {
	return r.nats.Request(sourceID, &pbw.DiscordInformationRequest{
		Timestamp: mapToProtoTimestamp(time.Now().UTC()),
		Type:      requestType,
		Payload:   payload,
	})
}

func mapToProtoTimestamp(timeValue time.Time) *timestamp.Timestamp {
	protoTimestamp, _ := ptypes.TimestampProto(timeValue)
	return protoTimestamp
}
