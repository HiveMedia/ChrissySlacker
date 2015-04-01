package main

import (
	"log"
    "fmt"
	"net/http"
    "net/url"
    "encoding/json"
    "time"
    "errors"
    "strings"
    "io"
    "os"
    "golang.org/x/net/websocket"
    "io/ioutil"
)

var config MyConfig
func main() {
    config = LoadConfig()
    connected := false
    var ws *websocket.Conn
    var wsURL string
    var err error
    for {
        if connected == false {
            fmt.Printf("Reconnecting :(  \r\n")
            wsURL, err = GetWS()
            if err != nil {
                fmt.Print(err)
                return
            }
            fmt.Printf(wsURL)
            ws, err = websocket.Dial(wsURL, "", wsURL)
            if err != nil {
                fmt.Println(err)
                time.Sleep(1*time.Second)
                fmt.Printf("Reconnecting FAILED\r\n")
                continue
            }
            connected = true
        }
        fmt.Printf("Process incomming text\r\n")
        msg,err := ProcessSlackMSG(ws)
        if err != nil {
            fmt.Print(err)
            if err == io.EOF {
                connected = false
            }
            fmt.Printf("Process incomming HAS FAILED!\r\n")
            continue
        }
        fmt.Printf("Process Commands\r\n")
        RunFunctions(msg)
        fmt.Printf("\r\n RAW Received: %s.\n\r", msg)
    }
}
func LoadConfig()(MyConfig){
    configFile, err := os.Open("config.json")
    if err != nil {
        fmt.Printf("opening config file %v", err.Error())
    }
    mybot := MyConfig{}
    jsonParser := json.NewDecoder(configFile)
    if err = jsonParser.Decode(&mybot); err != nil {
        fmt.Printf("parsing config file", err.Error())
    }
    return mybot
}
func RunFunctions(msg SlackMSG){
    if msg.Text != "" {
        mycmd := strings.Fields(msg.Text)
        fmt.Printf("Fields are: %q", mycmd)

        if strings.Contains(mycmd[0], "chrissy") || strings.Contains(mycmd[0], config.ID)  {
            var text string
            text = "Meow"

            if len(mycmd) >= 2  {
                fmt.Printf("MyCMD1: %s\r\n",mycmd[1])
                if mycmd[1] == "" {
                    text = "Meow"
                } else {
                    if strings.Contains(mycmd[1], "now") {
                        text = "Fucked if I know, you read https://hiveradio.net/icebreath/icecast/stats/"
                    }
                }
            }


            url := fmt.Sprintf("https://slack.com/api/chat.postMessage?token=%s&as_user=true&channel=%s&text=%s",config.Token,msg.Channel,url.QueryEscape(text))
            resp, err := http.Get(url)
            if err != nil {
                log.Fatal("API ERROR: ", err)
            }
            defer resp.Body.Close()
            body, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                log.Fatal("API ERROR: ", err,  body)
            }
        }
    }

}

func ProcessSlackMSG (ws *websocket.Conn)(SlackMSG, error){
    var err error
    message := SlackMSG{}

    if ws.IsClientConn == nil {
        err = errors.New("not connected")
        return message, err
    }
    var msg = make([]byte, 1024)
    var n int
    if n, err = ws.Read(msg); err != nil {
        return message, err
    }
    err = json.Unmarshal(msg[:n], &message)
    if err != nil {
        return message, err
    }
    return message, nil
}

func GetWS() (string, error){
    req := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s",config.Token)
    resp, err := http.Get(req)
    if err != nil {
        log.Fatal("API ERROR: ", err)
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal("API ERROR: ", err)
    }
    mybot := RTMstart{}
    err = json.Unmarshal(body, &mybot)
    if err != nil {
        fmt.Print(err)
    }
    return mybot.URL, nil
}

type MyConfig struct {
    Token string `json:"token"`
    ID string `json:"id"`
}

type RTMstart struct {
    URL string `json:"url"`

}

type SlackMSG struct {
    Type string `json:"type"`
    Text string `json:"text"`
    Channel string `json:"channel"`
}