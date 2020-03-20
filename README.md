# Archivio

## Archivio is not done, and is currently being worked on  

A simple discord server archive bot  
**[Invite link](https://discordapp.com/api/oauth2/authorize?client_id=664992430738898944&permissions=8&scope=bot)**

## Commands

**\>archive [channel(s)/all]** - Creates an archive of channel/server messages. Channel(s) parameter can be left blank, a comma seperated list of channels (channel1,channel2,channel3), or `all`. The command creates a zip file and uploads it to the channel in which the command was run in. This command can only be run by administrators.
  
**\>help** - Shows command help

## Building

`go build src/*.go`  
Then run `./archivio`

## Configuration

All configuration is stored in config.json  

`token`(string) - Discord bot authentication token  
`prefix`(string) - Discord bot prefix  
`userToken`(string) - Discord selfbot token used for external server backups, leave blank to disable `>extarchive`  
`saveArchives`(bool) - Default: `false`, if set to true archive zips will be kept in archives folder
