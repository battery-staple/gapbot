package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// process commands that can only be run in dms
func dmCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string) {
	switch command {
	case "help":
		dmHelpCommand(s, m)
	case "ping":
		pingCommand(s, m)
	case "register":
		registerUserCommand(s, m)
	case "lastplayed":
		lastPlayingCommand(s, m)
	case "lastloved":
		lastLovedCommand(s, m)
	case "lastregister":
		registerUserLastFMCommand(s, m)
	case "bigletters":
		makeBigLettersCommand(s, m)
	default:
		defaultHelpCommand(s, m)
	}
}

// process commands as normal user
func userCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string) {
	switch command {
	case "help":
		serverHelpCommand(s, m)
	case "user":
		userInfoCommand(s, m)
	case "server":
		serverInfoCommand(s, m)
	case "addrole":
		addRoleCommand(s, m)
	case "delrole":
		delRoleCommand(s, m)
	case "roles":
		listAvailableRolesCommand(s, m)
	case "myroles":
		listMyRolesCommand(s, m)
	default:
		dmCommand(s, m, command)
	}
}

// process commands as admin
func adminCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string) {
	switch command {
	case "purge":
		purgeCommand(s, m)
		logCommand(s, m)
	case "addlog":
		addLoggingChannelCommand(s, m)
		logCommand(s, m)
	case "removelog":
		removeLoggingChannelCommand(s, m)
		logCommand(s, m)
	case "kick":
		kickUserCommand(s, m)
		logCommand(s, m)
	case "ban":
		banUserCommand(s, m)
		logCommand(s, m)
	case "help":
		adminHelpCommand(s, m)
	case "deregister":
		deregisterUserCommand(s, m)
		logCommand(s, m)
	case "mute":
		muteCommand(s, m)
		logCommand(s, m)
	case "unmute":
		unmuteCommand(s, m)
		logCommand(s, m)
	default:
		userCommand(s, m, command)
	}
}

// process commands for ops
func opsCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string) {
	switch command {
	case "massregister":
		tempMassRegisterCommand(s, m)
		logCommand(s, m)
	case "repush-names":
		pushNamesCommand(s, m)
		logCommand(s, m)
	case "push-user":
		pushUserCommand(s, m)
		logCommand(s, m)
	case "reload-bot":
		reloadBotCommand(s, m)
		logCommand(s, m)
	case "help":
		opsHelpCommand(s, m)
	default:
		adminCommand(s, m, command)
	}
}

func dmHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSendEmbed(m.ChannelID, getDMHelpEmbed())
}

func serverHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSendEmbed(m.ChannelID, getServerHelpEmbed())
}

func adminHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSendEmbed(m.ChannelID, getAdminHelpEmbed())
}

func opsHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSendEmbed(m.ChannelID, getOpsHelpEmbed())
}

func defaultHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// this is evaluation to check and see if it contains a number at the start
	// if it does then it's probably someone typing something like $20
	command := strings.TrimPrefix(m.Content, loadedConfigData.Prefix)
	firstChar := string([]rune(command)[0])
	if !strings.ContainsAny(firstChar, "0123456789 ") {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s isn't a valid command. Use %shelp to learn more", strings.TrimPrefix(m.Content, loadedConfigData.Prefix), loadedConfigData.Prefix))
	}
}

func pingCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "Pong!")
}

func avatarCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Mentions) > 0 {
		if len(m.Mentions) > 4 {
			s.ChannelMessageSend(m.ChannelID, "Make sure to mention less than 5 users")
			return
		}
		for _, u := range m.Mentions {
			s.ChannelMessageSendEmbed(m.ChannelID, getAvatarEmbed(u))
		}
		return
	}
	s.ChannelMessageSendEmbed(m.ChannelID, getAvatarEmbed(m.Author))
}

func purgeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	fields := strings.Fields(m.Content)
	n, err := strconv.Atoi(fields[len(fields)-1])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Make sure a number of messages to delete is specified at the end of the command")
		return
	}
	if n > 99 || n < 1 {
		s.ChannelMessageSend(m.ChannelID, "Please enter a number between 1 and 99 (inclusive)")
		return
	}
	var messageIDs []string
	messages, err := s.ChannelMessages(m.ChannelID, n+1, "", "", "")
	if err != nil {
		fmt.Printf("Error getting messages: %s", err)
		return
	}
	for _, element := range messages {
		messageIDs = append(messageIDs, element.ID)
	}
	err = s.ChannelMessagesBulkDelete(m.ChannelID, messageIDs)
	if err != nil {
		fmt.Printf("Error deleting messages in channel %s: %s", m.ChannelID, err)
		s.ChannelMessageSend(m.ChannelID, "Unable to delete messages. Please check permissions and try again")
		return
	}
}

func userInfoCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	g, err := s.Guild(m.GuildID)
	if err != nil {
		fmt.Printf("Error getting guild: %s", err)
		return
	}
	if len(m.Mentions) > 0 {
		if len(m.Mentions) > 4 {
			s.ChannelMessageSend(m.ChannelID, "Make sure to mention less than 5 users")
			return
		}
		for _, u := range m.Mentions {
			s.ChannelMessageSendEmbed(m.ChannelID, getUserEmbed(u, s, g))
		}
		return
	}
	s.ChannelMessageSendEmbed(m.ChannelID, getUserEmbed(m.Author, s, g))
}

func removeUserCommand(s *discordgo.Session, m *discordgo.MessageCreate, ban bool) {
	method := "kick"
	if ban {
		method = "ban"
	}

	fields := strings.SplitN(strings.TrimPrefix(m.Content, loadedConfigData.Prefix), " ", 3)
	if len(fields) < 3 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Please make sure to specify a user and reason when giving the %s", method))
		return
	}
	reason := fields[2]
	if len(m.Mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Make sure to specify a user to %s", method))
		return
	}
	user := m.Mentions[0]

	dmUser(s, *user, fmt.Sprintf("You have been given the %s because %s", method, reason))
	if ban {
		dmUser(s, *user, "next time, follow the goddamn rules :)\nhttps://www.youtube.com/watch?v=FXPKJUE86d0")
	} else {
		dmUser(s, *user, "get dabbed on\nhttps://cdn.discordapp.com/attachments/593650772227653672/631587659604688897/dabremy.png")
	}

	var err error
	if ban {
		err = s.GuildBanCreateWithReason(m.GuildID, user.ID, reason, 1)
	} else {
		err = s.GuildMemberDeleteWithReason(m.GuildID, user.ID, reason)
	}
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to %s %s for reason %s", method, user.Mention(), reason))
		fmt.Printf("Error when giving user %s the %s , %s", user.Mention(), method, err)
		return
	}

	prevMessage, err := s.ChannelMessages(m.ChannelID, 1, "", "", "")
	if err != nil || len(prevMessage) < 1 {
		fmt.Printf("Error retrieving previous message: %s", err)
	}
	err = s.ChannelMessageDelete(m.ChannelID, prevMessage[0].ID)
	if err != nil {
		fmt.Printf("Error deleting previous message: %s", err)
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s was given the %s because of reason: %s", user.Mention(), method, reason))
}

func kickUserCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	removeUserCommand(s, m, false)
}

func banUserCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	removeUserCommand(s, m, true)
}

func serverInfoCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	g, err := s.Guild(m.GuildID)
	if err != nil {
		fmt.Printf("Error getting guild: %s", err)
		return
	}
	guildOwner, err := s.User(g.OwnerID)
	if err != nil {
		fmt.Printf("Error getting guild owner: %s", err)
		return
	}
	s.ChannelMessageSendEmbed(m.ChannelID, getServerEmbed(s, g, guildOwner))
}

func addLoggingChannelCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	err := addLoggingChannel(m.ChannelID)
	if err == nil {
		s.ChannelMessageSend(m.ChannelID, "This channel will now be used for logging")
	} else if err == errChannelRegistered {
		s.ChannelMessageSend(m.ChannelID, "This channel is already set up for logging")
	} else {
		s.ChannelMessageSend(m.ChannelID, "There was an error while setting up this channel for logging")
		fmt.Printf("Error while configuring %s for logging: %s", m.ChannelID, err)
	}
}

func removeLoggingChannelCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	err := removeLoggingChannel(m.ChannelID)
	if err == nil {
		s.ChannelMessageSend(m.ChannelID, "This channel will no longer be used for logging")
	} else if err == errChannelNotRegistered {
		s.ChannelMessageSend(m.ChannelID, "This channel has not yet been configured for logging")
	} else {
		s.ChannelMessageSend(m.ChannelID, "There was an error while removing this channel's logging status")
		fmt.Printf("Error while removing %s from logging: %s", m.ChannelID, err)
	}
}

func tempMassRegisterCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		fmt.Printf("Error getting guild: %s", err)
		return
	}
	for _, mem := range guild.Members {
		if !mem.User.Bot {
			c, err := s.UserChannelCreate(mem.User.ID)
			if err != nil {
				fmt.Printf("Error creating channel: %s", err)
			} else if _, ok := loadedUserData.Users[mem.User.ID]; !ok {
				_, err := s.ChannelMessageSend(c.ID, fmt.Sprintf("Please send me `%sregister {your first and last name} {grade as a number}` (e.g. `%sregister Jono Jenkens 12`)", loadedConfigData.Prefix, loadedConfigData.Prefix))
				if err != nil {
					fmt.Printf("Error sending message to user: %s", err)
				}
			}
		}
	}
}

func listAvailableRolesCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	roles, err := getAvailableRoles(s, m, m.Author)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Unable to query roles.")
		fmt.Printf("Unable to query roles: %s", err)
		return
	}
	s.ChannelMessageSendEmbed(m.ChannelID, getRolesEmbed(roles, "Available Roles"))
}

func listMyRolesCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	mem, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		fmt.Printf("Error getting member: %s", err)
		return
	}
	roles := mem.Roles
	sortedroles := make([]*discordgo.Role, len(roles))
	for i, role := range roles {
		r, err := s.State.Role(m.GuildID, role)
		if err != nil {
			fmt.Printf("Error finding role: %s", err)
			return
		}
		sortedroles[i] = r
	}
	sort.SliceStable(sortedroles, func(i, j int) bool {
		return sortedroles[i].Position > sortedroles[j].Position
	})
	s.ChannelMessageSendEmbed(m.ChannelID, getRolesEmbed(sortedroles, "Your Roles"))
}

func addRoleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	parseUpdateRole(s, m, false)
}

func delRoleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	parseUpdateRole(s, m, true)
}

func pushNamesCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// iterate through all stored users and push the user joined string to the names channel
	for _, userID := range loadedUserData.Users {
		discordUser, err := s.User(userID.DiscordID)
		if err != nil {
			// press f in chat
			fmt.Printf("Unable to make user object for %v: %s", userID, err)
		} else {
			pushNewUser(discordUser, s)
		}
	}
}

func muteCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// automatically add a configured muted role to a user
	if len(m.Mentions) <= 0 {
		s.ChannelMessageSend(m.ChannelID, "Make sure to mention a user to mute")
		return
	}
	err := setMuted(s, m, m.Mentions[0], true)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Unable to mute user")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "User muted.")
}

func unmuteCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// automatically remove a muted role from a user
	if len(m.Mentions) <= 0 {
		s.ChannelMessageSend(m.ChannelID, "Make sure to mention a user to unmute")
		return
	}
	err := setMuted(s, m, m.Mentions[0], false)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Unable to unmute user")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "User unmuted.")
}

func lastPlayingCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	userStruct, err := getUserStruct(m.Author)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Unable to query user. Are you registered?")
		return
	}
	lastListened, err := getUserLastListened(userStruct)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to get last.fm data for user `%s`. Please make sure you've linked accounts.", userStruct.LastFmAccount))
		return
	}
	lastPlayingEmbed := getLastFMTrackEmbed(lastListened)
	s.ChannelMessageSend(m.ChannelID, "Most recently scrobbled song:")
	s.ChannelMessageSendEmbed(m.ChannelID, lastPlayingEmbed)
}

func lastLovedCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	userStruct, err := getUserStruct(m.Author)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Unable to query user. Are you registered?")
		return
	}
	lastListened, err := getUserLastLoved(userStruct)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to get last.fm data for user `%s`. Please make sure you've linked accounts.", userStruct.LastFmAccount))
		return
	}
	lastLovedEmbed := getLastFMTrackEmbed(lastListened)
	s.ChannelMessageSend(m.ChannelID, "Most recently loved song:")
	s.ChannelMessageSendEmbed(m.ChannelID, lastLovedEmbed)
}

func registerUserLastFMCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	fields := strings.SplitN(strings.TrimPrefix(m.Content, loadedConfigData.Prefix), " ", 2)
	if len(fields) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Ensure that your username is included")
		return
	}
	user := m.Author
	lastFMUserName := fields[1]
	err := registerUserLastFM(user, lastFMUserName)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Unable to link accounts.")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Linked accounts successfully!")
}

func makeBigLettersCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	fields := strings.SplitN(strings.ToLower(strings.TrimPrefix(m.Content, loadedConfigData.Prefix)), " ", 2)
	if len(fields) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Please make sure you enter text to be transformed")
		return
	}
	initial := fields[1]
	message := ""
	for _, char := range initial {
		if char != ' ' {
			if char == 'b' {
				message += ":b:"
			} else if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
				message += fmt.Sprintf(":regional_indicator_%c:", char)
				continue
			} else {
				message += fmt.Sprintf("%c", char)
			}
		}
		message += " "
	}
	s.ChannelMessageSend(m.ChannelID, message)
}

func pushUserCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Mentions) <= 0 {
		s.ChannelMessageSend(m.ChannelID, "Make sure to mention a user to push the name for")
		return
	}
	pushNewUser(m.Mentions[0], s)
	s.ChannelMessageSend(m.ChannelID, "User pushed")
}

func reloadBotCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// reload bot resources in place
	if err := loadUsers(); err != nil {
		s.ChannelMessageSend(m.ChannelID, "error loading user file")
	}

	if err := loadConfig(); err != nil {
		s.ChannelMessageSend(m.ChannelID, "error loading config file")
	} else {
		startBotConfig(s)
	}
}
