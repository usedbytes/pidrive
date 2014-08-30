package main

import (
    "fmt"
    "log"
    "github.com/usedbytes/ui"
    "github.com/usedbytes/input"
    "github.com/usedbytes/leds"
	/* TODO: merge changes and use upstream gompd */
    "github.com/usedbytes/gompd/mpd"
    "image/draw"
    "image"
    "strings"
)


type ledMenu struct {
    *leds.Leds
    list *ui.List
    redBar, greenBar, blueBar, speedBar *ui.ProgressBar
    redLbl, greenLbl, blueLbl, speedLbl *ui.Label
    view *ui.View
    focus interface{}
}

func CreateLedMenu(p *Pidrive, basename string, values [4]int) *ledMenu {
    this := new(ledMenu)

    paths := [4]string{ strings.Join([]string{basename, "red"}, "/"),
                        strings.Join([]string{basename, "green"}, "/"),
                        strings.Join([]string{basename, "blue"}, "/"),
                        strings.Join([]string{basename, "speed"}, "/") }
    this.Leds = leds.NewLeds(paths, values)

    font := p.Fonts["Tiny Font"]
    if (font == nil) {
        log.Fatalln("Couldn't get Tiny Font")
    }

    iconFont := p.Fonts["Icon Font"]
    if (font == nil) {
        log.Fatalln("Couldn't get Icon Font")
    }

    this.view = ui.NewView(nil, "LED Menu");
    this.view.SetWidth(WIDTH)
    this.view.SetHeight(HEIGHT)
    this.view.Visible = false

    this.list = ui.NewList(this.view.Widget, font, iconFont);
    this.list.SetWidth(25)
    this.list.SetHeight(HEIGHT - 8)
    this.list.AutoHeight = false
    this.list.SetPos(image.Point{0, 8})
    //this.list.selected = 0
    
    this.list.AddItem("R:", 0, ledListAction, this)
    this.list.AddItem("G:", 0, ledListAction, this)
    this.list.AddItem("B:", 0, ledListAction, this)
    this.list.AddItem("S:", 0, ledListAction, this)
    this.view.AddChild(this.list)

    this.redBar = ui.NewProgressBar(this.view.Widget)
    this.redBar.SetWidth(WIDTH - 25)
    this.redBar.SetHeight(8)
    this.redBar.SetPos(image.Point{26, 8})
    this.redBar.Max = 255
    this.redBar.Progress = values[leds.RED]
    this.view.AddChild(this.redBar)

    this.greenBar = ui.NewProgressBar(this.view.Widget)
    this.greenBar.SetWidth(WIDTH - 25)
    this.greenBar.SetHeight(8)
    this.greenBar.SetPos(image.Point{26, 16})
    this.greenBar.Max = 255
    this.greenBar.Progress = values[leds.GREEN]
    this.view.AddChild(this.greenBar)

    this.blueBar = ui.NewProgressBar(this.view.Widget)
    this.blueBar.SetWidth(WIDTH - 25)
    this.blueBar.SetHeight(8)
    this.blueBar.SetPos(image.Point{26, 24})
    this.blueBar.Max = 255
    this.blueBar.Progress = values[leds.BLUE]
    this.view.AddChild(this.blueBar)

    this.speedBar = ui.NewProgressBar(this.view.Widget)
    this.speedBar.SetWidth(WIDTH - 25)
    this.speedBar.SetHeight(8)
    this.speedBar.SetPos(image.Point{26, 32})
    this.speedBar.Max = 255
    this.speedBar.Progress = values[leds.SPEED]
    this.view.AddChild(this.speedBar)

    this.focus = this.list

    return this
}

func (l *ledMenu) Open(attrs mpd.Attrs) {
    l.view.Visible = true
    fmt.Printf("%+v\n", attrs)
}

func (l *ledMenu) Hide(attrs mpd.Attrs) {
    l.view.Visible = false
    fmt.Printf("%+v\n", attrs)
}

func (l *ledMenu) End(attrs mpd.Attrs) {
    fmt.Printf("%+v\n", attrs)
}

func (l *ledMenu) Update(attrs []string) {
    fmt.Printf("%+v\n", attrs)
}

func (l *ledMenu) HandleInput(key rune) bool {
    if (l.focus == l.list) {
        return l.list.HandleInput(key)
    } else {
        var led int
        bar := l.focus.(*ui.ProgressBar)
        switch (l.focus) {
        case l.redBar:
            led = leds.RED
        case l.greenBar:
            led = leds.GREEN
        case l.blueBar:
            led = leds.BLUE
        case l.speedBar:
            led = leds.SPEED
        }
        switch key {
            case input.KEY_DOWN, input.KEY_SCROLLDOWN:
                if (bar.Progress >= 5) {
                    bar.Progress -= 5
                    l.SetLed(led, bar.Progress)
                }
                return true
            case input.KEY_UP, input.KEY_SCROLLUP:
                if (bar.Progress <= 250) {
                    bar.Progress += 5
                    l.SetLed(led, bar.Progress)
                }
                return true
            case input.KEY_BACK, input.KEY_ENTER:
                l.focus = l.list
                l.list.Item(l.list.Selected).IconIndex = int(ui.ICON_BLANK[0])
                return true
        }
        return false
    }
}

func (l *ledMenu) Draw(to draw.Image) image.Rectangle {
    return l.view.Draw(to)
}

func (l *ledMenu) ModuleName() string {
    return "LEDs"
}

func (l *ledMenu) Name() string {
    return "leds"
}

func ledListAction(v ...interface{}) {
    if li, ok := v[0].(*ui.ListItem); ok {
        if l, ok := li.Tag.(*ledMenu); ok {
            li.IconIndex = int(ui.ICON_SELECTED[0])
            switch li.Text {
            case "R:":
                fmt.Println("Red selected")
                l.focus = l.redBar
            case "G:":
                fmt.Println("Green selected")
                l.focus = l.greenBar
            case "B:":
                fmt.Println("Blue selected")
                l.focus = l.blueBar
            case "S:":
                fmt.Println("Speed selected")
                l.focus = l.speedBar
            }
            return
        }
    }
    fmt.Println("Doesn't seem to be a valid list item")
}
