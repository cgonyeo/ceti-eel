package ceti-eel

import (
    "fmt"
    "strings"
    "github.com/thoj/go-ircevent"
)

var roomName = "#botsex"
var myNick = "gonbot"
var con *irc.Connection
var goTime = false
var admin = false
var op = false

func main() {
    //Make a connection, "nick", "user"
    con = irc.IRC("gonbot", "gonbot")
    err := con.Connect("skynet.csh.rit.edu:6667")
    if err != nil {
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
    con.SendRaw("NAMES " + roomName)
}

func connectionMade(e *irc.Event) {
    con.Join(roomName)
}

//Parrot back a message if it begins with <name>: 
func newPrivmsg(e *irc.Event) {
    msg := e.Message()
    if len(msg) > len(myNick) && msg[0:len(myNick)] == myNick {
        con.Privmsg(roomName, msg[len(myNick)+2:])
    }
}

func modeChanged(e *irc.Event) {
    //Check if we got op
    if len(e.Arguments) >= 3 && e.Arguments[0] == roomName {
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
        }
        if gotAdmin && wereInTheList {
            //We got admin, let's take note of this
            admin = adding
        }
        checkNames()
    }
}

func gotNames(e *irc.Event) {
    names := strings.Fields(e.Message())
    //If we're admin and it's not go time, then let's make it go time
    if admin && !goTime {
        fmt.Println("Go time!")
        goTime = true
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
            fmt.Println("Go time!")
            goTime = true
        } else if goTime {
            //If there's an admin and we have op and it's go time, 
            //it's not actually go time
            goTime = false
        }
    } else {
        //We're not op or admin. It's not go time.
        goTime = false
    }
    if goTime {
        takeControl(names)
    }
}

//De-op and de-admin everyone
func takeControl(names []string) {
    for _, name := range names {
        if name[1:] != myNick  && name[1:] != "dgonyeo" {
            if name[0] == '!' {
                con.SendRaw("MODE " + roomName + " -ao " + name[1:])
            }
            if name[0] == '@' {
                con.SendRaw("MODE " + roomName + " -o " + name[1:])
            }
        }
    }
}
