package discord

import (
	"errors"
	"io"
	"log"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

// MessageType is the type of Message
type MessageType string

const (
	// MessageTypeCreate is the message type for message creation.
	MessageTypeCreate MessageType = "create"
	// MessageTypeUpdate is the message type for message updates.
	MessageTypeUpdate = "update"
	// MessageTypeDelete is the message type for message deletion.
	MessageTypeDelete = "delete"
)

var errAlreadyJoined = errors.New("already joined")

// DiscordClient handles discord sessions and client configurations
type DiscordClient struct {
	args        []interface{}
	messageChan chan Message

	Session     *discordgo.Session
	Sessions    []*discordgo.Session
	OwnerUserID string
}

var channelIDRegex = regexp.MustCompile("<#[0-9]*>")

func (d *DiscordClient) replaceChannelNames(message *discordgo.Message, content string) string {
	return channelIDRegex.ReplaceAllStringFunc(content, func(str string) string {
		c, err := d.Channel(str[2 : len(str)-1])
		if err != nil {
			return str
		}

		return "#" + c.Name
	})
}

var roleIDRegex = regexp.MustCompile("<@&[0-9]*>")

func (d *DiscordClient) replaceRoleNames(message *discordgo.Message, content string) string {
	return roleIDRegex.ReplaceAllStringFunc(content, func(str string) string {
		roleID := str[3 : len(str)-1]

		c, err := d.Channel(message.ChannelID)
		if err != nil {
			return str
		}

		g, err := d.Guild(c.GuildID)
		if err != nil {
			return str
		}

		for _, r := range g.Roles {
			if r.ID == roleID {
				return "@" + r.Name
			}
		}

		return str
	})
}

func (d *DiscordClient) onMessageCreate(s *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Content == "" || (message.Author != nil && message.Author.Bot) {
		return
	}

	d.messageChan <- &DiscordMessage{
		Discord:          d,
		DiscordgoMessage: message.Message,
		MessageType:      MessageTypeCreate,
	}
}

func (d *DiscordClient) onMessageUpdate(s *discordgo.Session, message *discordgo.MessageUpdate) {
	if message.Content == "" || (message.Author != nil && message.Author.Bot) {
		return
	}

	d.messageChan <- &DiscordMessage{
		Discord:          d,
		DiscordgoMessage: message.Message,
		MessageType:      MessageTypeUpdate,
	}
}

func (d *DiscordClient) onMessageDelete(s *discordgo.Session, message *discordgo.MessageDelete) {
	d.messageChan <- &DiscordMessage{
		Discord:          d,
		DiscordgoMessage: message.Message,
		MessageType:      MessageTypeDelete,
	}
}

// UserName gets the username of the client
func (d *DiscordClient) UserName() string {
	if d.Session.State.User == nil {
		return ""
	}
	return d.Session.State.User.Username
}

// UserID the UserID of the client
func (d *DiscordClient) UserID() string {
	if d.Session.State.User == nil {
		return ""
	}
	return d.Session.State.User.ID
}

// Open a connection and start listening
func (d *DiscordClient) Open() (<-chan Message, error) {
	gateway, err := discordgo.New(d.args...)
	if err != nil {
		return nil, err
	}

	s, err := gateway.GatewayBot()
	if err != nil {
		return nil, err
	}

	d.Sessions = make([]*discordgo.Session, s.Shards)

	for i := 0; i < s.Shards; i++ {
		session, err := discordgo.New(d.args...)
		if err != nil {
			return nil, err
		}
		session.ShardCount = s.Shards
		session.ShardID = i
		session.AddHandler(d.onMessageCreate)
		session.AddHandler(d.onMessageUpdate)
		session.AddHandler(d.onMessageDelete)
		session.State.TrackPresences = false

		d.Sessions[i] = session
	}

	d.Session = d.Sessions[0]

	for i := 0; i < len(d.Sessions); i++ {
		d.Sessions[i].Open()
	}

	return d.messageChan, nil
}

// IsMe checks if the message has the same id as this session
func (d *DiscordClient) IsMe(message Message) bool {
	if d.Session.State.User == nil {
		return false
	}
	return message.UserID() == d.Session.State.User.ID
}

// SendMessage sends a discord message
func (d *DiscordClient) SendMessage(channel string, message string) error {
	if channel == "" {
		log.Println("Empty channel could not send message", message)
		return nil
	}

	if _, err := d.Session.ChannelMessageSend(channel, message); err != nil {
		log.Println("Error sending discord message: ", err)
		return err
	}

	return nil
}

// SendEmbedMessage sends an embed discord message
func (d *DiscordClient) SendEmbedMessage(channel string, message *discordgo.MessageEmbed) error {
	if channel == "" {
		log.Println("Empty channel could not send message", message)
		return nil
	}

	if _, err := d.Session.ChannelMessageSendEmbed(channel, message); err != nil {
		log.Println("Error sending discord embed message: ", err)
		return err
	}

	return nil
}

// DeleteMessage deletes a discord message
func (d *DiscordClient) DeleteMessage(channel, messageID string) error {
	return d.Session.ChannelMessageDelete(channel, messageID)
}

// SendFile sends a file to a discord channel
func (d *DiscordClient) SendFile(channel, name string, r io.Reader) error {
	if _, err := d.Session.ChannelFileSend(channel, name, r); err != nil {
		log.Println("Error sending discord message: ", err)
		return err
	}
	return nil
}

// BanUser bans a user for a specified duration
func (d *DiscordClient) BanUser(channel, userID string, duration int) error {
	return d.Session.GuildBanCreate(channel, userID, 0)
}

// UnbanUser unbans a user
func (d *DiscordClient) UnbanUser(channel, userID string) error {
	return d.Session.GuildBanDelete(channel, userID)
}

// Join forces the client to join a guild based on an invite ID
func (d *DiscordClient) Join(join string) error {
	if i, err := d.Session.Invite(join); err == nil {
		if _, err := d.Guild(i.Guild.ID); err == nil {
			return errAlreadyJoined
		}
	}

	if _, err := d.Session.InviteAccept(join); err != nil {
		return err
	}
	return nil
}

// Typing makes makes it appear that the client is typing in a specified Channel
func (d *DiscordClient) Typing(channel string) error {
	return d.Session.ChannelTyping(channel)
}

// PrivateMessage sends a private message to a user
func (d *DiscordClient) PrivateMessage(userID string, message string) error {
	c, err := d.Session.UserChannelCreate(userID)
	if err != nil {
		return err
	}
	return d.SendMessage(c.ID, message)
}

// IsBotOwner checks if a message is from the configured bot owner
func (d *DiscordClient) IsBotOwner(message Message) bool {
	return message.UserID() == d.OwnerUserID
}

// IsPrivate checks if a message is being sent from a direct message
func (d *DiscordClient) IsPrivate(message Message) bool {
	c, err := d.Channel(message.Channel())
	return err == nil && c.Type == discordgo.ChannelTypeDM
}

// IsChannelOwner checks if the message is from the Guild admin
func (d *DiscordClient) IsChannelOwner(message Message) bool {
	c, err := d.Channel(message.Channel())
	if err != nil {
		return false
	}
	g, err := d.Guild(c.GuildID)
	if err != nil {
		return false
	}
	return g.OwnerID == message.UserID() || d.IsBotOwner(message)
}

// IsModerator checks if the message is from a Guild moderator
func (d *DiscordClient) IsModerator(message Message) bool {
	p, err := d.UserChannelPermissions(message.UserID(), message.Channel())
	if err == nil {
		if p&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator || p&discordgo.PermissionManageChannels == discordgo.PermissionManageChannels || p&discordgo.PermissionManageServer == discordgo.PermissionManageServer {
			return true
		}
	}

	return d.IsChannelOwner(message)
}

// ChannelCount returns the number of channels the client is connected to
func (d *DiscordClient) ChannelCount() int {
	return len(d.Guilds())
}

// UserCount returns the number of users the client is connected to
func (d *DiscordClient) UserCount() int {
	totalCount := 0
	for _, guild := range d.Guilds() {
		totalCount += guild.MemberCount
	}
	return totalCount
}

// MessageHistory returns the last x number of messages from a channel
func (d *DiscordClient) MessageHistory(channel string) []Message {
	c, err := d.Channel(channel)
	if err != nil {
		return nil
	}

	messages := make([]Message, len(c.Messages))
	for i := 0; i < len(c.Messages); i++ {
		messages[i] = &DiscordMessage{
			Discord:          d,
			DiscordgoMessage: c.Messages[i],
			MessageType:      MessageTypeCreate,
		}
	}

	return messages
}

// GetMessages returns a list of messages for a channel
func (d *DiscordClient) GetMessages(channelID string, limit int, beforeID string) ([]Message, error) {
	channelMessages, err := d.Session.ChannelMessages(channelID, limit, beforeID, "", "")
	if err != nil {
		return nil, err
	}

	messages := make([]Message, len(channelMessages))
	for i := 0; i < len(channelMessages); i++ {
		messages[i] = &DiscordMessage{
			Discord:          d,
			DiscordgoMessage: channelMessages[i],
			MessageType:      MessageTypeCreate,
		}
	}

	return messages, err
}

// Channel resolves a Channel object from a channelID
func (d *DiscordClient) Channel(channelID string) (channel *discordgo.Channel, err error) {
	for _, s := range d.Sessions {
		channel, err = s.State.Channel(channelID)
		if err == nil {
			return channel, nil
		}
	}
	return
}

// Guild resolves a Guild object from a guildID
func (d *DiscordClient) Guild(guildID string) (guild *discordgo.Guild, err error) {
	for _, s := range d.Sessions {
		guild, err = s.State.Guild(guildID)
		if err == nil {
			return guild, nil
		}
	}
	return
}

// Guilds returns an arrray of Guilds the client is connected to
func (d *DiscordClient) Guilds() []*discordgo.Guild {
	guilds := []*discordgo.Guild{}
	for _, s := range d.Sessions {
		guilds = append(guilds, s.State.Guilds...)
	}
	return guilds
}

// UserChannelPermissions returns the permissions for a user in a channel
func (d *DiscordClient) UserChannelPermissions(userID, channelID string) (apermissions int, err error) {
	for _, s := range d.Sessions {
		apermissions, err = s.State.UserChannelPermissions(userID, channelID)
		if err == nil {
			return apermissions, nil
		}
	}
	return
}

// UserColor returns the color code for a user in a channel
func (d *DiscordClient) UserColor(userID, channelID string) int {
	for _, s := range d.Sessions {
		color := s.State.UserColor(userID, channelID)
		if color != 0 {
			return color
		}
	}
	return 0
}

// Nickname resolves a users nickname based on a message
func (d *DiscordClient) Nickname(message Message) string {
	return d.NicknameForID(message.UserID(), message.UserName(), message.Channel())
}

// NicknameForID resolves a users nickname based on a userId with a fallback username
func (d *DiscordClient) NicknameForID(userID, userName, channelID string) string {
	c, err := d.Channel(channelID)
	if err == nil {
		g, err := d.Guild(c.GuildID)
		if err == nil {
			for _, m := range g.Members {
				if m.User.ID == userID {
					if m.Nick != "" {
						return m.Nick
					}
					break
				}
			}
		}
	}
	return userName
}
