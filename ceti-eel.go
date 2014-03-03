package main

import (
    "os"
    "fmt"
    "strings"
    "bufio"
    "code.google.com/p/gcfg"
    "github.com/thoj/go-ircevent"
)

var server string
var channel = "#botsex"
var myNick = "gonbot"
var con *irc.Connection
var goTime = false
var admin = false
var op = false
var config = struct {
    Connection struct {
        Server string
        Port string
        Channel string
        Nick string
    }
}{}

func main() {
    args := os.Args
    if len(args) != 2 {
        fmt.Println("Usage: ceti-eel <config-file.gcfg>")
        os.Exit(1)
    }
    if args[1] == "-h" {
        fmt.Println("Usage: ceti-eel <config-file.gcfg>")
        os.Exit(0)
    }

    file, fiError := os.Open(args[1])
    if fiError != nil {
        fmt.Println("Error opening config file")
        os.Exit(1)
    }
    defer file.Close()

    var lines string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines += scanner.Text() + "\n"
    }
    if scanner.Err() != nil {
        fmt.Println("Scanner error")
        os.Exit(1)
    }

    err := gcfg.ReadStringInto(&config, lines)
    if err != nil {
        fmt.Println("Error parsing config")
        os.Exit(1)
    }

    server = config.Connection.Server + ":" + config.Connection.Port
    channel = "#" + config.Connection.Channel
    myNick = config.Connection.Nick

    fmt.Println("Connecting to " + server)
    fmt.Println("And joining " + channel)
    fmt.Println("As " + myNick)

    //Make a connection, "nick", "user"
    con = irc.IRC(myNick, myNick)
    conErr := con.Connect(server)
    if conErr != nil {
        fmt.Println("Failed connecting")
        return
    }

    con.AddCallback("001", connectionMade)
    con.AddCallback("PRIVMSG", newPrivmsg)
    con.AddCallback("MODE", modeChanged)
    con.AddCallback("353", gotNames)
    con.Loop()
}

//Send a message to the server requesting the list of everyone in the channel
func checkNames() {
    con.SendRaw("NAMES " + channel)
}

func changeUsersMode(nick string, mode string) {
    con.SendRaw("MODE " + channel + " " + mode + " " + nick)
}

//Connection to the server is successful, so let's join the channel
func connectionMade(e *irc.Event) {
    con.Join(channel)
}

//Parrot back a message if it begins with <name>: 
func newPrivmsg(e *irc.Event) {
    msg := e.Message()
    if len(msg) > len(myNick) && msg[0:len(myNick)] == myNick {
        con.Privmsg(channel, msg[len(myNick)+2:])
    }
}

func modeChanged(e *irc.Event) {
    //Check if our op/admin priveleges changed
    if len(e.Arguments) >= 3 && e.Arguments[0] == channel {
        adding := true
        gotOp := false
        gotAdmin := false
        wereInTheList := false
        //Note if modes are being added or removed
        if e.Arguments[1][0] == '-' {
            adding = false
        }
        //Check if op was changed
        for _, letter := range e.Arguments[1][1:] {
            if letter == 'o' {
                gotOp = true
            }
        }
        //Check if admin was changed
        for _, letter := range e.Arguments[1][1:] {
            if letter == 'a' {
                gotAdmin = true
            }
        }
        //Check if we're in the list
        for _, name := range e.Arguments[2:] {
            if name == myNick {
                wereInTheList = true
            }
        }
        if gotOp && wereInTheList {
            //We got op, let's take note of this
            op = adding
            fmt.Print("op:")
            fmt.Println(adding)
        }
        if gotAdmin && wereInTheList {
            //We got admin, let's take note of this
            admin = adding
            fmt.Print("admin:")
            fmt.Println(adding)
        }
        checkNames()
    }
}

func gotNames(e *irc.Event) {
    names := strings.Fields(e.Message())
    banish := true
    //If we're admin and it's not go time, then let's make it go time
    if admin {
        if !goTime {
            goTime, banish = true, false
        }
    } else if op {
        //Else if we're op and there's not an admin, and it's not go time, 
        //then let's make it go time
        theresAnAdmin := false
        for _, name := range names {
            if name[0] == '!' {
                theresAnAdmin = true
            }
        }

        if !theresAnAdmin && !goTime {
            goTime, banish = true, false
        } else if theresAnAdmin && goTime {
            //If there's an admin and we have op and it's go time, 
            //it's not actually go time :(
            goTime = false
        }
    } else {
        //We're not op or admin. It's not go time.
        goTime = false
    }
    if goTime {
        takeControl(names, banish)
    }
}

//De-op and de-admin everyone
func takeControl(names []string, banish bool) {
    for _, name := range names {
        if name[1:] != myNick  && name[1:] != "dgonyeo" {
            if name[0] == '!' {
                changeUsersMode(name[1:], "-a")
                changeUsersMode(name[1:], "-o")
                if banish {
                    changeUsersMode(name[1:], "+b")
                    con.SendRaw("KICK " + channel + " " + name[1:])
                }
            }
            if name[0] == '@' {
                changeUsersMode(name[1:], "-o")
                if banish {
                    changeUsersMode(name[1:], "+b")
                    con.SendRaw("KICK " + channel + " " + name[1:])
                }
            }
        }
    }
}
