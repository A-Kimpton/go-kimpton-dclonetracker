package main

import (
	"flag"

	_ "kimpton.io/dclonetracker/d2ioscraper"
	"kimpton.io/dclonetracker/discord"
)

var(
	discordToken = flag.String("discordtoken", "", "Access token for discord bot")
)

func main(){

	flag.Parse()

	discord.ConnectToDiscord(*discordToken)
	
}