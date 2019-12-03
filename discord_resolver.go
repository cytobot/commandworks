package commandworks

import (
	"encoding/json"
	"strconv"
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

	isNSFW, _ := strconv.ParseBool(res["NSFW"])

	return &discordgo.Channel{
		ID:            res["ID"],
		GuildID:       res["GuildID"],
		Name:          res["Name"],
		Topic:         res["Topic"],
		LastMessageID: res["LastMessageId"],
		NSFW:          isNSFW,
	}, nil
}

func (r *discordResolver) ResolveGuild(sourceID string, guildID string) (*discordgo.Guild, error) {
	msg, err := r.sendRequest(sourceID, "guild", map[string]string{
		"guildID": guildID,
	})
	if err != nil {
		return nil, err
	}

	res := r.getResult(msg)

	memberCount, _ := strconv.ParseInt(res["MemberCount"], 10, 64)
	embedEnabled, _ := strconv.ParseBool(res["EmbedEnabled"])

	return &discordgo.Guild{
		ID:           res["ID"],
		Name:         res["Name"],
		Icon:         res["Icon"],
		OwnerID:      res["OwnerID"],
		JoinedAt:     discordgo.Timestamp(res["JoinedAt"]),
		MemberCount:  int(memberCount),
		EmbedEnabled: embedEnabled,
		Description:  res["Description"],
	}, nil
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

	return &discordgo.User{
		ID:       res["ID"],
		Username: res["Username"],
		Avatar:   res["Avatar"],
	}, nil
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
