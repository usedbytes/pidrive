package main

import (
    "fmt"
    "github.com/usedbytes/fonts"
    "github.com/usedbytes/input"
    "github.com/usedbytes/s4548"
    "github.com/gvalkov/golang-evdev"
	/* TODO: merge changes and use upstream gompd */
    "github.com/usedbytes/gompd/mpd"
    //"testing"
    "time"
    //"math/rand"
    "path"
    "path/filepath"
    "os"
    "log"
    "strings"
    "image/draw"
    "image"
)

/* This should be read only */
type Pidrive struct {
    Server string
    S mpd.Attrs
    Fonts map[string]*fonts.Font
}

type IntentAction int
const (
    INTENT_OPEN IntentAction = iota
    INTENT_HIDE
    INTENT_END
    INTENT_UPDATE
)

type Intent struct {
    Action IntentAction
    Target string
    Payload map[string]string
}

type Task interface {
    Open(mpd.Attrs)
    Hide(mpd.Attrs)
    End(mpd.Attrs) // Maybe a no-op?
    Update([]string)
    HandleInput(key rune) bool
    Draw(draw.Image) image.Rectangle

    ModuleName() string // The module name for the hashmap
    Name() string // Human-readable name
}

func LoadFonts(dir string) (map[string]*fonts.Font, error) {
    font_list := make(map[string]*fonts.Font)
    err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {  
        if (err != nil) {
            return err;
        }
        if strings.ToLower(path.Ext(p)) == ".fnt" {
            fnt := fonts.NewFontFromFile(p)
            if (fnt == nil) {
                return &PidriveError{fmt.Sprintf("Error reading: %s", p)}
            }
            fmt.Println("Read font:", fnt.Name())
            font_list[fnt.Name()] = fnt
        }
        return nil
    })

    return font_list, err;
}

func handleGlobalInput(key rune) bool {
    
    switch key {
    case input.KEY_POWER:
        fmt.Println("Got power key");
        return true
    case input.KEY_BACK, input.KEY_ESC:
        i := Intent{INTENT_HIDE, "special_current", nil}
        Intents <- i
    case input.KEY_HOME:
        fmt.Println("Got home key");
        i := Intent{INTENT_HIDE, "special_current",
                map[string]string{"clear": "yes"}}
        Intents <- i
        return true
    }
    
    return false
}

var Intents chan Intent

func main() {

    var err error
    /* Initialisation */
    /* TODO: Load all this from a config file */
    P := new(Pidrive)
    P.Server = "raspberrypi:6600"

    P.Fonts, err = LoadFonts("/home/kernelcode/Programming/go/src/github.com/usedbytes/fonts")
    if (err != nil) {
        log.Fatalln(err)
    }

    Screen := s4548.NewS4548(s4548.GetS4548EnvPath())
    fmt.Println(Screen.Width())

    /* MPD Handling */
    conn, err := mpd.Dial("tcp", P.Server)
    if err != nil {
        fmt.Println("Error connecting")
        return
    }
    defer conn.Close()
    mpdUpdates := make(chan []string)
    fmt.Println("Opened MPD")

    go func() {
        updates, err := conn.Idle()
        if (err == nil) {
            fmt.Println("Send updates")
            mpdUpdates <- updates
        } else {
            fmt.Println(err)
            return
        }
    }()

    /* Input Handling */
    events := make(chan *evdev.InputEvent)
    quitInput := make(chan int)
    keys := make(chan rune, 10)
    input.StartListening(events)
    go input.ProcessInputEvents(events, keys, quitInput)

    /* UI Handling */
    uiTicker := time.NewTicker(100 * time.Millisecond)
    Intents = make(chan Intent, 10)

    tasks := make(map[string]Task)

    tasks["nowplaying"] = CreateNowPlaying(P, Intents)
    tasks["nowplaying1"] = CreateNowPlaying(P, Intents)
    ledvals := [4]int{0, 255, 0, 5}
    tasks["leds"] = CreateLedMenu(P, "/sys/devices/platform/bcm2708_i2c.0/" +
                                     "i2c-0/0-003b/pidrio-leds", ledvals)

    names := []string{"Now Playing", "Settings", "LEDs"}
    tees := []Task{ tasks["nowplaying"],
                    tasks["nowplaying1"],
                    tasks["leds"] }
    tasks["main"] = CreateMainMenu(P, Intents, names, tees)
    
    crumbs := make([]Task, 0, 10)
    var currentTask Task
    
    currentTask = tasks["main"]
    currentTask.Open(nil)
    
    /* Main Loop */
    for {
        select {
        case k := <-keys:
            fmt.Println("Got an input event", k)
            if (!currentTask.HandleInput(k)) {
                fmt.Println("Passing on to global");
                handleGlobalInput(k)
            }
            // Check if global op (HOME, POWER...)
            // if !currentTask.Handle() 
            //      try and find a default (play, pause, volume...)
        case <-uiTicker.C:
            Screen.Damage(currentTask.Draw(Screen))
            Screen.Repair()
        case u := <-mpdUpdates:
            /* Update S
             * Notify current task first...
             * Then goroutine all the background tasks' notification
             */
            fmt.Printf("Updates! %+v\n", u)
            P.S, _ = conn.Status()
            currentTask.Update(u)
            go func() {
                updates, err := conn.Idle()
                if (err == nil) {
                    mpdUpdates <- updates
                } else {
                    fmt.Println(err)
                    return
                }
            }()

        case i := <-Intents:
            /* Open - Send open to the relevant task, along with attrs
             *        Make it the default input handler
             * Close - Send close to the relevant task (allowing it to
             *         save any state etc
             */
            fmt.Printf("Got an intent %+v\n", i);
            if (i.Target == "special_current") {
                switch (i.Action) {
                case INTENT_HIDE:
                    if (i.Payload["clear"] == "yes") {
                        crumbs = make([]Task, 0, 10)
                    }
                    if len(crumbs) > 0 {
                        /* Pop the last crumb */
                        var prev Task
                        prev, crumbs =
                            crumbs[len(crumbs)-1], crumbs[:len(crumbs)-1]
                        fmt.Println("Popped", prev.Name())
                        currentTask.Hide(nil)
                        currentTask = prev
                        currentTask.Open(nil)
                        Screen.Damage(image.Rect(0, 0, s4548.WIDTH, s4548.HEIGHT))
                    } else {
                        fmt.Println("Start Over - Homescreen")
                        currentTask.Hide(nil)
                        currentTask = tasks["main"]
                        currentTask.Open(nil)
                        Screen.Damage(image.Rect(0, 0, s4548.WIDTH, s4548.HEIGHT))
                    }
                case INTENT_UPDATE:
                    currentTask.Update(nil)
                }
            } else {
                task := tasks[i.Target]
                if (task != nil) {
                    switch (i.Action) {
                    case INTENT_OPEN:
                        crumbs = append(crumbs, currentTask)
                        fmt.Println("Pushed", currentTask.Name())
                        currentTask.Hide(nil)
                        currentTask = task
                        currentTask.Open(i.Payload)
                        Screen.Damage(image.Rect(0, 0, s4548.WIDTH, s4548.HEIGHT))
                    }
                }
            }
        }
    }
}


type PidriveError struct {
    What string
}

func (p *PidriveError) Error() string {
    return p.What
}
