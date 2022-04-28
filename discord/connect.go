package discord

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"kimpton.io/dclonetracker/d2ioscraper"
	"kimpton.io/dclonetracker/utils"
)

const(
	ALERT_THRESHOLD = 4
)


func ConnectToDiscord(token string) {

	// Establish connection using token
	dg, err := discordgo.New("Bot " + token)
	defer dg.Close()
	if err != nil {
		panic(err)
	}

	// Register Handlers
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		panic(err)
	}

	// Set status
	setStatus(dg)

	// Auto update some messages
	go autoUpdateMessages(dg)

	Printf("Bot is running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc




}

func setStatus(s *discordgo.Session) {

	s.UpdateListeningStatus("!dct-help for help")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Check if the message is from bot, if so, ignore
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!dct-help" {
		help(s, m)
	}

	if m.Content == "!dct-live" {
		live(s, m)
	}

	if m.Content == "!dct-setup" {
		setup(s, m)
	}

	if m.Content == "!dct-alert" {
		alert(s, m)
	}

	if m.Content == "!dct-clear" {
		clear(s, m)
	}
}

func help(s *discordgo.Session, m *discordgo.MessageCreate) {
	message := ""
	message = message + "Welcome to DClone Tracker - BETA\n"
	message = message + "DClone Tracker was built by Kimpton using diablo2.io API\n"
	message = message + "Make a role called `Diablo` and those users will recieve alerts\n"
	message = message + "\n"
	message = message + "The following commands are valid...\n"
	message = message + "`!dct-help` - Displays this help message\n"
	message = message + "`!dct-live` - Displays the current dclone status\n"
	message = message + "`!dct-setup` - Uses this channel for publishing dclone spawn progress\n"
	message = message + "`!dct-alert` - Uses this channel for alerting when dclone spawns\n"	
	message = message + "`!dct-clear` - Stops updating this channel (setup and alert)\n"

	s.ChannelMessageSend(m.ChannelID, message)
}

func live(s *discordgo.Session, m *discordgo.MessageCreate) (*discordgo.Message, error) {

	message := GetDcloneInfo()
	return s.ChannelMessageSend(m.ChannelID, message)

}

func setup(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Printing....
	dgmsg, _ := live(s, m)
	message := dgmsg.Content
	message = message + "\nThis message will update every 5 seconds!"
	s.ChannelMessageEdit(m.ChannelID, dgmsg.ID, message)
	AddMessage(DiscordMessage{
		ChannelID: m.ChannelID, 
		MessageID: dgmsg.ID,
	})
}

func alert(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Printing....
	message := ""
	message = message + "Welcome to DClone Tracker\n"
	message = message + "DClone Tracker was built by Kimpton using diablo2.io API\n"
	message = message + "\n"
	message = message + "This channel will now be used for alerts when DClone is at 4/6 or more progress"
	s.ChannelMessageSend(m.ChannelID, message)
	AddMessage(DiscordMessage{
		ChannelID: m.ChannelID, 
		MessageID: "0",
	})
}

func clear(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Printing....
	message := ""
	message = message + "Welcome to DClone Tracker\n"
	message = message + "DClone Tracker was built by Kimpton using diablo2.io API\n"
	message = message + "\n"
	message = message + "This channel will now no longer recieve alerts or updates"
	s.ChannelMessageSend(m.ChannelID, message)
	RemoveAllMessage(DiscordMessage{
		ChannelID: m.ChannelID, 
		MessageID: "0",
	})
}

// Better Printer
func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Printf("[discord-bot] " + format, a...)
}

func autoUpdateMessages(s *discordgo.Session) {
	state := d2ioscraper.GetGlobalState()
	// While there is no records, do nothing
	for len(state) == 0 {
		time.Sleep(1 * time.Second)
		state = d2ioscraper.GetGlobalState()
	}

	// Container for previously alerted on
	prevAlerts := make([]d2ioscraper.D2CloneState,0)

	// Infinate loop
	for {
		
		// Get a handle on all messages that need updating
		allMessages := GetAllMessages()
		
		// Get latest info
		state = d2ioscraper.GetGlobalState()
		message := GetDcloneInfo()
		message = message + "\nThis message will update every 5 seconds!"

		// Check if any dclones are ready to spawn
		toAlertOn := make([]d2ioscraper.D2CloneState,0)
		doNotAlertOn := make([]d2ioscraper.D2CloneState,0)
		for _, serverState := range(state) {
			if serverState.Progress >= ALERT_THRESHOLD { // all servers to alert on
				// Loop over all previous alerted servers
				wasInTheList := false
				for _, prevServer := range(prevAlerts) {
					// Check if its in the list
					if serverState.IsSameServer(prevServer) {
						wasInTheList = true
						// Check if the state has changed
						changed, _ := serverState.HasUpdated(prevServer)
						if changed {
							// State has changed, lets alert on it
							toAlertOn = append(toAlertOn, serverState)
						} else {
							// It was already alerted on, but state hasn't
							// Until state changes, do not alert
							doNotAlertOn = append(doNotAlertOn, serverState)
						}
					}
				}
				if !wasInTheList {
					toAlertOn = append(toAlertOn, serverState)
				}
				
			}
		}

		// Update each one
		for _, dm := range(allMessages) {
			if dm.MessageID != "0"{
				// Update Post
				s.ChannelMessageEdit(dm.ChannelID, dm.MessageID, message)
			} else {
				// Post if any DClones are getting close				
				for _, serverState := range(toAlertOn) {
					message = serverState.PrettyString() + "\nPrepare, diablo may walk sanctuary!"
					extra := " @everyone"
					ch, _ := s.Channel(dm.ChannelID)
					gu, _ := s.Guild(ch.GuildID)
					for _, r := range(gu.Roles) {
						if r.Name == "Diablo" || r.Name == "diablo" {
							extra = fmt.Sprintf(" <@&%s>", r.ID)
						}
					}
					message = message + extra
					message = message + "\nData courtesy of diablo2.io"
					s.ChannelMessageSend(dm.ChannelID, message)
				}				
			}			
		}

		// Update Previous Alerted Servers
		prevAlerts = toAlertOn
		prevAlerts = append(prevAlerts, doNotAlertOn...)

		// Wait 5 secs
		time.Sleep(5 * time.Second)
		
	}
	
}

func GetDcloneInfo() string {
	// To print
	message := ""
	message = message + "Welcome to DClone Tracker\n"
	message = message + "DClone Tracker was built by Kimpton using diablo2.io API\n"
	message = message + "\n"

	// Get current state
	state := d2ioscraper.GetGlobalState()

	// If there is nothing to say, return
	if len(state) == 0 {
		
		message = message + "The bot is still starting up - Please try again in 1 minute\n"
		return message

	}

	// Build a table
	usT := utils.Table{} // Table

	usT.AddHeaders("Region", "Ladder", "Hardcore", "Progress")
	for _, d := range(state) {
		usT.AddString(d.Region.String()) // Region
		if d.IsLadder { // Ladder
			usT.AddString("Yes")
		} else {
			usT.AddString("No")
		}
		if d.IsHardcore { // Hardcore
			usT.AddString("Yes")
		} else {
			usT.AddString("No")
		}
		usT.AddString(fmt.Sprintf("%d/%d", d.Progress, d2ioscraper.MAX_CLONE_LEVELS))
		usT.NewRow()
	}

	return message + "```" + usT.ToPrettyString() + "```"
	
}