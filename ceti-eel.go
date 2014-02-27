package main

import (
    "fmt"
    "github.com/thoj/go-ircevent"
)

var roomName = "#botsex"
var myNick = "gonbot"
var con *irc.Connection
var goTime = false

func main() {
    //Make a connection, "nick", "user"
    con = irc.IRC("gonbot", "gonbot")
    err := con.Connect("skynet.csh.rit.edu:6667")
    if err != nil {
        fmt.Println("Failed connecting")
        return
    }

    con.AddCallback("001", connectionMade)
    con.AddCallback("JOIN", joinedChannel)
    con.AddCallback("PRIVMSG", newPrivmsg)
    con.AddCallback("MODE", modeChanged)
    con.AddCallback("CTCP_ACTION", actionReceived)
    con.Loop()
}

func connectionMade(e *irc.Event) {
    con.Join(roomName)
}

func joinedChannel(e *irc.Event) {
    con.Privmsg(roomName, "Sup bitches?")
}

func newPrivmsg(e *irc.Event) {
    msg := e.Message()
    if len(msg) > len(myNick) && msg[0:len(myNick)] == myNick {
        con.Privmsg(roomName, msg[len(myNick)+2:])
    }
}

func modeChanged(e *irc.Event) {
    //Check if we got op
    if len(e.Arguments) >= 3 && e.Arguments[0] == roomName {
        //If modes were given,
        if e.Arguments[1][0] == '+' {
            gotOp := false
            gotAdmin := false
            wereInTheList := false
            //Check if op was given out
            for _, letter := range e.Arguments[1][1:] {
                if letter == 'o' {
                    gotOp = true
                }
            }
            //Check if admin was given out
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
                con.Privmsg(roomName, "I got op")
            }
            if gotAdmin && wereInTheList {
                //We got admin, let's take note of this
                con.Privmsg(roomName, "I got admin")
            }
        }
    }
}

func actionReceived(e *irc.Event) {
    fmt.Println(e.Message())
}
