package main

import (
    "fmt"
    "log"
    "github.com/usedbytes/ui"
    "github.com/usedbytes/input"
    //"github.com/usedbytes/pidrive/pidrive-global"
	/* TODO: merge changes and use upstream gompd */
    "github.com/usedbytes/gompd/mpd"
    "image"
    "image/draw"
    "strconv"
    "path/filepath"
    "strings"
    "time"
)

type nowPlaying struct {
    *ui.View
    conn *mpd.Client
    s *mpd.Attrs
    i chan Intent
    
    playState int
    title, artist string
    elapsed, length int
    volume, volCounter int
    volCounting bool

    file string

    statusBar *ui.StatusBar
    titleLbl, artistLbl, elapsedLbl, lengthLbl, volumeLbl *ui.Label
    elapsedBar, volumeBar *ui.ProgressBar
}

func CreateNowPlaying(p *Pidrive, i chan Intent) *nowPlaying {
    var err error
    this := new(nowPlaying)
    this.i = i
    
    this.conn, err = mpd.Dial("tcp", p.Server)
    if err != nil {
        fmt.Println("Error connecting")
        log.Fatalln(err)
    }

    font := p.Fonts["Tiny Font"]
    if (font == nil) {
        log.Fatalln("Couldn't get Tiny Font")
    }

    this.View = ui.NewView(nil, "Now Playing");
    this.SetWidth(WIDTH)
    this.SetHeight(HEIGHT)

    this.statusBar = ui.NewStatusBar(this.View.Widget)
    fmt.Println("Statusbar bounds", this.statusBar.Bounds())
    this.AddChild(this.statusBar)

    this.titleLbl = ui.NewLabel(this.View.Widget, font)
    this.titleLbl.AutoWidth = false
    this.titleLbl.AutoHeight = true
    this.titleLbl.SetWidth(WIDTH)
    this.titleLbl.VAlign = ui.Middle
    this.titleLbl.HAlign = ui.Centre
    this.titleLbl.Scroll = true
    this.titleLbl.SetPos(image.Point{0, 8})
    this.AddChild(this.titleLbl)

    this.artistLbl = ui.NewLabel(this.View.Widget, font)
    this.artistLbl.AutoWidth = false
    this.artistLbl.AutoHeight = true
    this.artistLbl.SetWidth(WIDTH)
    this.artistLbl.VAlign = ui.Middle
    this.artistLbl.HAlign = ui.Centre
    this.artistLbl.Scroll = true
    this.artistLbl.SetPos(image.Point{0, 16})
    this.AddChild(this.artistLbl)

    this.elapsedBar = ui.NewProgressBar(this.View.Widget)
    this.elapsedBar.SetWidth(WIDTH)
    this.elapsedBar.SetHeight(8)
    this.elapsedBar.SetPos(image.Point{0, 24})
    this.elapsedBar.Progress = 0
    this.AddChild(this.elapsedBar)

    this.elapsedLbl = ui.NewLabel(this.View.Widget, font)
    this.elapsedLbl.AutoWidth = false
    this.elapsedLbl.AutoHeight = true
    this.elapsedLbl.SetWidth(WIDTH / 2)
    this.elapsedLbl.SetHeight(8)
    this.elapsedLbl.VAlign = ui.Middle
    this.elapsedLbl.HAlign = ui.Left
    this.elapsedLbl.SetPos(image.Point{0, 32})
    this.AddChild(this.elapsedLbl)

    this.lengthLbl = ui.NewLabel(this.View.Widget, font)
    this.lengthLbl.AutoWidth = false
    this.lengthLbl.AutoHeight = true
    this.lengthLbl.SetWidth(WIDTH / 2)
    this.lengthLbl.SetHeight(8)
    this.lengthLbl.VAlign = ui.Middle
    this.lengthLbl.HAlign = ui.Right
    this.lengthLbl.SetPos(image.Point{WIDTH / 2, 32})
    this.AddChild(this.lengthLbl)

    this.volumeLbl = ui.NewLabel(this.View.Widget, font)
    this.volumeLbl.AutoWidth = false
    this.volumeLbl.AutoHeight = true
    this.volumeLbl.SetWidth(WIDTH)
    this.volumeLbl.SetHeight(8)
    this.volumeLbl.VAlign = ui.Middle
    this.volumeLbl.HAlign = ui.Centre
    this.volumeLbl.SetPos(image.Point{0, 32})
    this.volumeLbl.Visible = false
    this.volumeLbl.Text = "Volume"
    this.AddChild(this.volumeLbl)

    this.volumeBar = ui.NewProgressBar(this.View.Widget)
    this.volumeBar.SetWidth(WIDTH)
    this.volumeBar.SetHeight(8)
    this.volumeBar.SetPos(image.Point{0, 24})
    this.volumeBar.Max = 100
    this.volumeBar.Progress = 0
    this.volumeBar.Visible = false
    this.AddChild(this.volumeBar)

    return this
}

/* This should be called when main thread gets an update */
func (n * nowPlaying) Update(updates []string) {
    s, _ := n.conn.Status()

    swapped_attrs := make(map[string]bool)
    for _, v := range updates {
        swapped_attrs[v] = true
    }

    if (swapped_attrs["playlist"]) {
        n.statusBar.Tracks, _ = strconv.Atoi(s["playlistlength"])
    }

    if (swapped_attrs["player"]) {
        if (s["state"] == "play") {
            n.statusBar.State = ui.STATE_PLAYING
            n.titleLbl.Scroll = true
            n.artistLbl.Scroll = true
        } else if (s["state"] == "pause") {
            n.statusBar.State = ui.STATE_PAUSED
            n.titleLbl.Scroll = false
            n.artistLbl.Scroll = false
            n.titleLbl.ResetScroll()
            n.artistLbl.ResetScroll()
        } else {
            n.statusBar.State = ui.STATE_NONE
        }

        n.statusBar.TrackNum, _ = strconv.Atoi(s["song"])

        song, _ := n.conn.CurrentSong()
        if (song != nil) {
            file := song["file"]
            if (file != n.file) {
                fmt.Println("File changed")
                /* Update all the shiz */
                n.file = file
                /* Set the title/artist labels */
                if (len(song["Title"]) > 0) {
                    n.artistLbl.Text = song["Artist"]
                    n.titleLbl.Text = song["Title"]
                } else {
                    n.titleLbl.Text = filepath.Base(song["file"])
                    n.artistLbl.Text = ""
                }
                /* Reset the scroll position for a tick */
                n.titleLbl.ResetScroll()
                n.artistLbl.ResetScroll()

                /* Set the align so that we do "the right thing" */
                if (n.titleLbl.Font().Width(n.titleLbl.Text) > n.titleLbl.Bounds().Dx()) {
                    n.titleLbl.HAlign = ui.Left
                } else {
                    n.titleLbl.HAlign = ui.Centre
                }
                if (n.artistLbl.Font().Width(n.artistLbl.Text) > n.artistLbl.Bounds().Dx()) {
                    n.artistLbl.HAlign = ui.Left
                } else {
                    n.artistLbl.HAlign = ui.Centre
                }

            }
            timstrs := strings.Split(s["time"], ":")
            if len(timstrs) == 2 {
                elapsed := timstrs[0] + "s"
                songlen := timstrs[1] + "s"
                elapsedDur, _ := time.ParseDuration(elapsed)
                songlenDur, _ := time.ParseDuration(songlen)

                h := int(elapsedDur.Hours())
                m := int(elapsedDur.Minutes()) - (60 * h)
                s := int(elapsedDur.Seconds()) - (60 * m) - (3600 * h)
                elapsed = ""
                if (h > 0) {
                    elapsed = fmt.Sprintf("%d:", h)
                }
                elapsed = fmt.Sprintf("%s%2d:%02d", elapsed, m, s)
                n.elapsedLbl.Text = elapsed

                h = int(songlenDur.Hours())
                m = int(songlenDur.Minutes()) - (60 * h)
                s = int(songlenDur.Seconds()) - (60 * m)
                songlen = ""
                if (h > 0) {
                    songlen = fmt.Sprintf("%d:", h)
                }
                songlen = fmt.Sprintf("%s%2d:%02d", songlen, m, s)
                n.lengthLbl.Text = songlen

                n.elapsedBar.Min = 0
                n.elapsedBar.Max = int(songlenDur.Seconds())
                n.elapsedBar.Progress = int(elapsedDur.Seconds())
            }
        }
    }

    if (swapped_attrs["options"]) {
        if (s["random"] == "1") {
            n.statusBar.Shuffle = true
        } else {
            n.statusBar.Shuffle = false
        }
        if (s["repeat"] == "1") {
            n.statusBar.Repeat = true
        } else {
            n.statusBar.Repeat = false
        }
    }

    if (swapped_attrs["mixer"]) {
        oldvol := n.volume
        n.volume, _ = strconv.Atoi(s["volume"])
        if ( oldvol != n.volume ) {
            n.volumeBar.Progress = n.volume
            n.volCounter = 8
            if (!n.volCounting) {
                n.volCounting = true
                n.volumeLbl.Visible = true
                n.volumeBar.Visible = true
                n.elapsedLbl.Visible = false
                n.elapsedBar.Visible = false
                n.lengthLbl.Visible = false
                n.Damage = n.volumeBar.Bounds().Union(
                    n.volumeLbl.Bounds())
                go func() {
                    for n.volCounter > 0 {
                        fmt.Println("Decrementing")
                        time.Sleep(100 * time.Millisecond)
                        n.volCounter--
                    }
                    n.volumeLbl.Visible = false
                    n.volumeBar.Visible = false
                    n.volCounting = false
                    n.elapsedLbl.Visible = true
                    n.elapsedBar.Visible = true
                    n.lengthLbl.Visible = true
                    n.Damage = n.volumeBar.Bounds().Union(
                        n.volumeLbl.Bounds())
                }()
            }
        }
    }
}

func (n *nowPlaying) Open(attrs mpd.Attrs) {
    fmt.Printf("%+v\n", attrs)
    s, _ := n.conn.Status()
    n.volume, _ = strconv.Atoi(s["volume"])
    n.titleLbl.Active = true
    n.artistLbl.Active = true
    n.Visible = true
    n.Update([]string{"playlist", "player", "options"})
}

func (n *nowPlaying) Hide(attrs mpd.Attrs) {
    fmt.Printf("%+v\n", attrs)
    n.titleLbl.Active = false
    n.artistLbl.Active = false
    n.Visible = false
}

func (n *nowPlaying) End(attrs mpd.Attrs) {
    fmt.Printf("%+v\n", attrs)
    n.conn.Close()
}

func (n *nowPlaying) HandleInput(key rune) bool {
    switch key {
    case input.KEY_VOLUMEUP, input.KEY_UP, input.KEY_SCROLLUP, '+':
        n.conn.SetVolume(n.volume + 3)
    case input.KEY_VOLUMEDOWN, input.KEY_DOWN, input.KEY_SCROLLDOWN, '-':
        n.conn.SetVolume(n.volume - 3)
    case input.KEY_NEXTSONG, 'N':
        n.conn.Next()
    case input.KEY_PREVIOUSSONG, 'P':
        n.conn.Previous()
    case input.KEY_FASTFORWARD, input.KEY_RIGHT, 'F':
        n.conn.Seekrel(5)
    case input.KEY_REWIND, input.KEY_LEFT, 'B':
        n.conn.Seekrel(-5)
    case input.KEY_PLAYPAUSE, ' ':
        if (n.statusBar.State != ui.STATE_PLAYING) {
            n.conn.Pause(false)
        } else {
            n.conn.Pause(true)
        }
        return true
    case input.KEY_ENTER:
        /* Launch current playlist */
        /* (Intent) */
        return false
    case 'R':
        n.conn.Repeat(!n.statusBar.Repeat)
        return true
    case 'S':
        n.conn.Random(!n.statusBar.Shuffle)
        return true

    }


    return false
}

/* Everything should be good-to-go in here */
func (n *nowPlaying) Draw(to draw.Image) image.Rectangle {
    if (n.IsVisible()) {
        n.Update([]string{"player"})
        return n.View.Draw(to)
    }
    return image.ZR
}

func (n *nowPlaying) ModuleName() string {
    return "Now Playing"
}

func (n *nowPlaying) Name() string {
    return "nowplaying"
}
