package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"archive/zip"
    "encoding/json"
    "sync"
    "errors"
    "net/http"
    "github.com/bwmarrin/discordgo"
)

func filter(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func saveFile(name string, guild string, data string) {
    if _, err := os.Stat("./archives"); os.IsNotExist(err) {
        err := os.Mkdir("./archives", 0777)

        if err != nil {
            fmt.Println(err)
            return
        }
    }

    if _, err := os.Stat(fmt.Sprintf("./archives/%s", guild)); os.IsNotExist(err) {
        err := os.Mkdir(fmt.Sprintf("./archives/%s", guild), 0777)

        if err != nil {
            fmt.Println(err)
            return
        }
    }

	f, err := os.Create(fmt.Sprintf("./archives/%s/%s.json", guild, name))

	if err != nil {
        fmt.Println(err)
        return
	}

	_, err = f.WriteString(data)

	if err != nil {
        fmt.Println(err)
        f.Close()
        return
	}

	err = f.Close()

	if err != nil {
        fmt.Println(err)
        return
    }
    
    return
}

func createZip(guild string) {
	files, err := ioutil.ReadDir(fmt.Sprintf("./archives/%s", guild))

	out, err := os.Create(fmt.Sprintf("./archives/%s.zip", guild))

	if err != nil {
		 fmt.Println(err)
		 return
	}

	defer out.Close()

	w := zip.NewWriter(out)

	for i := 0; i < len(files); i++ {
			file := files[i]

			f, err := w.Create(file.Name())
			if err != nil {
				 fmt.Println(err)
				 return
			}

			data, err := os.Open(fmt.Sprintf("./archives/%s/%s", guild, file.Name()))
			if err != nil {
					fmt.Println(err)
					return
			}

			defer data.Close()

			body, err := ioutil.ReadAll(data)
			if err != nil {
					fmt.Println(err)
					return
			}

			_, err = f.Write([]byte(body))
			if err != nil {
				 fmt.Println(err)
				 return
			}
	}

	err = w.Close()
	if err != nil {
		 fmt.Println(err)
		 return
	}
}

func getChannelMessages(s *discordgo.Session, channelID string, limit int, before string, after string) ([]*discordgo.Message, error) {
    self, err := s.User("@me")

    if err != nil {
        return nil, errors.New("Failed to get bot info")
    }

    if !self.Bot {
        url := fmt.Sprintf("https://discordapp.com/api/v6/channels/%s/messages?limit=%d", channelID, limit)

        if before != "" {
            url += "&before=" + before
        }
        if after != "" {
            url += "&after=" + after
        }

        req, _ := http.NewRequest("GET", url, nil)
        req.Header.Add("authorization", s.Token)

        res, err := s.Client.Do(req)

        if err != nil {
            return nil, err
        }

        defer res.Body.Close()
        body, _ := ioutil.ReadAll(res.Body)

        var messages []*discordgo.Message

        err = json.Unmarshal([]byte(body), &messages)
        if err != nil {
            return nil, err
        }

        fmt.Println(messages)

        return messages, nil
    }

    messages, err := s.ChannelMessages(channelID, limit, before, after, "")
    return messages, err
}

func fetchall(s *discordgo.Session, channelID string) string {
	var done = false
	var lastMessageID = "0"
	var msgs = map[int]*discordgo.Message{}

	for !done {
        var messages []*discordgo.Message
        var err error

        if lastMessageID == "0" {
            messages, err = getChannelMessages(s, channelID, 100, "", "")
        } else {
            messages, err = getChannelMessages(s, channelID, 100, lastMessageID, "")
        }

        if err != nil {
            fmt.Println(err)

            return "[]"
        }

        for i := 0; i < len(messages); i++ {
            msgs[len(msgs)] = messages[i]

            if (i == len(messages) - 1) {
                lastMessageID = messages[i].ID
            }
        }

        if len(messages) == 0 {
            done = true
        }
    }

	arr := make([]*discordgo.Message, len(msgs))

	for i := 0; i < len(msgs); i++ {
		arr[i] = msgs[i]
	}

	res, err := json.Marshal(arr)

	if err != nil {
        fmt.Println(err)
        return "[]"
	}
	
	return string(res)
}

func fetch(s *discordgo.Session, ChannelID string, mChannelID string, wg *sync.WaitGroup) {
    Channel, err := s.Channel(ChannelID)

    if (err != nil) {
        s.ChannelMessageSend(mChannelID, fmt.Sprintf("Failed to fetch messages: %s", err))
        return
    }

    msgs := fetchall(s, Channel.ID)

    saveFile(fmt.Sprintf("%s|%s", Channel.Name, Channel.ID), Channel.GuildID, msgs)

    fmt.Println("Fetch done for", Channel.Name)

    defer wg.Done()

    s.ChannelMessageSend(mChannelID, fmt.Sprintf("Finished fetching messages from channel #%s", Channel.Name))
}

// Server struct, used for saving roles, channels, etc.
type Server struct {

}

// func getServerStructure(s *discordgo.Session, guildID string) Server {
//     g, err := s.Guild(guildID)

//     g.Roles
// }

// Config struct for bot config.json file
type Config struct {
	Token        string `json:"token"`
	Prefix       string `json:"prefix"`
	UserToken    string `json:"userToken"`
	SaveArchives bool   `json:"saveArchives"`
}

func readConfig() Config {
    file, err := os.Open("./config.json")

    if err != nil {
        fmt.Println("Failed to read config file:", err)
        os.Exit(3)
    }

    defer file.Close()
    bytes, _ := ioutil.ReadAll(file)

    var config Config
    json.Unmarshal(bytes, &config)

    return config
}