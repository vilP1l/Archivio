package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "strings"
    "sync"
    "github.com/bwmarrin/discordgo"
    "github.com/logrusorgru/aurora"
)

var config = readConfig()

func main() {
	client, err := discordgo.New("Bot " + config.Token)

	if err != nil {
		log.Fatalln("Failed to login to discord:", err)
	}

	err = client.Open()

	if err != nil {
		log.Fatalln("Failed to open connection to discord:", err)
    }
    
    client.AddHandler(messageCreate)
    client.AddHandler(ready)

    fmt.Println("-------------")
    fmt.Println(aurora.Cyan("Archivio v0.1"))
    fmt.Println(aurora.Green("Logged In!"))
    fmt.Println("-------------")

    if config.UserToken != "" {
        connectUser(config.UserToken)
    }

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	client.Close()
} 

func ready(s *discordgo.Session, event *discordgo.Ready) {
    s.UpdateStatus(0, ">help")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if len(m.Content) < len(config.Prefix) + 1 {
        return
    }

    if m.Author.Bot || m.Content[0:len(config.Prefix)] != config.Prefix || len(m.Content) == len(config.Prefix) {
        return
    }

    m.Content = m.Content[len(config.Prefix):]

    args := strings.Split(m.Content, " ")
    command := args[0]

    args = args[1:]

    if command == "archive" {
        perms, err := s.State.UserChannelPermissions(m.Author.ID, m.ChannelID)

        if err != nil {
            fmt.Println("Failed to get perms")
            return
        }

        if (perms & 0x8) != 0x8 {
            s.ChannelMessageSend(m.ChannelID, "You need to be a server administrator to use this command.")
            return
        }

        s.ChannelMessageSend(m.ChannelID, "Fetching channels...")

        var channels []string

        if len(args) == 0 {
            channels = []string{m.ChannelID}
        }

        if len(args) != 0 && args[0] == "all" {
            GuildChannels, err := s.GuildChannels(m.GuildID)

            if err != nil {
                s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Couldn't get guild channels: %s", err))
                return
            }
        
            arr := make([]string, len(GuildChannels))

            for i := 0; i < len(arr); i++ {
                arr[i] = GuildChannels[i].ID
            }

            for i := 0; i < len(arr); i++ {
                c, err := s.Channel(arr[i])

                if err == nil {
                    if c.Type != 0 {
                        arr[i] = ""
                    }
                } else {
                    arr[i] = ""
                }
            }

            selector := func(s string) bool { return s != "" }
            res := filter(arr, selector)

            channels = res
        } else {
            if len(args) != 0 {
                args[0] = strings.Replace(args[0], "<", "", -1)
                args[0] = strings.Replace(args[0], ">", "", -1)
                args[0] = strings.Replace(args[0], "#", "", -1)
    
                channels = strings.Split(args[0], ",")
            }
        }

        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Starting archive for %d channels, this may take a while...", len(channels)))

        var wg sync.WaitGroup

        wg.Add(len(channels))
        
        for i := 0; i < len(channels); i++ {
            go fetch(s, channels[i], m.ChannelID, &wg)
        }

        wg.Wait()

        createZip(m.GuildID)

        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Finished fetching messages from %d channels!**", len(channels)))

        file, err := os.Open(fmt.Sprintf("./archives/%s.zip", m.GuildID))

        if err != nil {
            s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to get zip."))
            return
        }

        s.ChannelFileSend(m.ChannelID, "Archive.zip", file)

        if !config.SaveArchives {
            err = os.Remove(fmt.Sprintf("./archives/%s.zip", m.GuildID))

            if err != nil {
                fmt.Println(err)
                return
            }
    
            err = os.RemoveAll(fmt.Sprintf("./archives/%s", m.GuildID))
    
            if err != nil {
                fmt.Println(err)
                return
            }
        }
    }

    if command == "extarchive" {
        s.ChannelMessageSend(m.ChannelID, "Fetching channels...")

        var channels []string

        if len(args) == 0 {
            channels = []string{m.ChannelID}
        }

        if len(args) != 0 && args[1] == "all" {
            GuildChannels, err := userSession.GuildChannels(args[0])

            if err != nil {
                s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Couldn't get guild channels: %s", err))
                return
            }
        
            arr := make([]string, len(GuildChannels))

            for i := 0; i < len(arr); i++ {
                arr[i] = GuildChannels[i].ID
            }

            for i := 0; i < len(arr); i++ {
                c, err := s.Channel(arr[i])

                if err == nil {
                    if c.Type != 0 {
                        arr[i] = ""
                    }
                } else {
                    arr[i] = ""
                }
            }

            selector := func(s string) bool { return s != "" }
            res := filter(arr, selector)

            channels = res
        } else {
            if len(args) != 0 {
                args[0] = strings.Replace(args[0], "<", "", -1)
                args[0] = strings.Replace(args[0], ">", "", -1)
                args[0] = strings.Replace(args[0], "#", "", -1)
    
                channels = strings.Split(args[0], ",")
            }
        }

        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Starting archive for %d channels, this may take a while...", len(channels)))

        var wg sync.WaitGroup

        wg.Add(len(channels))
        
        for i := 0; i < len(channels); i++ {
            go fetch(userSession, channels[i], m.ChannelID, &wg)
        }

        wg.Wait()

        createZip(args[0])

        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Finished fetching messages from %d channels!**", len(channels)))

        file, err := os.Open(fmt.Sprintf("./archives/%s.zip", args[0]))

        if err != nil {
            s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to get zip."))
            return
        }

        s.ChannelFileSend(m.ChannelID, "Archive.zip", file)

        if !config.SaveArchives {
            err = os.Remove(fmt.Sprintf("./archives/%s.zip", args[0]))

            if err != nil {
                fmt.Println(err)
                return
            }
    
            err = os.RemoveAll(fmt.Sprintf("./archives/%s", args[0]))
    
            if err != nil {
                fmt.Println(err)
                return
            }
        }
    }

    if command == "emote" {
        if len(args) < 2 {
            s.ChannelMessageSend(m.ChannelID, "Not enough args provided")
            return
        }   
        _, err := s.GuildEmojiCreate(m.GuildID, args[0], args[1], nil)

        if err != nil {
            s.ChannelMessageSend(m.ChannelID, err.Error())
            return
        }

        s.ChannelMessageSend(m.ChannelID, "Successfully uploaded emote.")
    }

    // if command == "purgeall" {
    //     rawMessages := fetchall(s, m.ChannelID)

    //     type Messages []struct {
    //         ID string `json:"id"`
    //         Author struct {
    //             ID string `json:"id"`
    //         }
    //     }

    //     var messages Messages

    //     json.Unmarshal([]byte(rawMessages), &messages)

    //     for _, message := range messages {
    //         s.ChannelMessagesBulkDelete(m.ChannelID, messages)
    //     }
    // }
}