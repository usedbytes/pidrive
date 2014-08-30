package main

import (
    "fmt"
    "log"
    "github.com/usedbytes/ui"
	/* TODO: merge changes and use upstream gompd */
    "github.com/usedbytes/gompd/mpd"

    "image/draw"
    "image"
)


type mainMenu struct {
    i chan Intent
    tasks []Task
    
    list *ui.List
    view *ui.View
}

func CreateMainMenu(p *Pidrive, i chan Intent, names []string, tasks []Task) *mainMenu {
    this := new(mainMenu)
    this.i = i
    
    font := p.Fonts["Tiny Font"]
    if (font == nil) {
        log.Fatalln("Couldn't get Tiny Font")
    }
    
    iconFont := p.Fonts["Icon Font"]
    if (font == nil) {
        log.Fatalln("Couldn't get Icon Font")
    }

    if (len(names) != len(tasks)) {
        log.Fatalln("names and tasks must be same length")
    }

    this.view = ui.NewView(nil, "Main Menu");
    this.view.SetWidth(WIDTH)
    this.view.SetHeight(HEIGHT)
    this.view.Visible = true

    this.list = ui.NewList(this.view.Widget, font, iconFont);
    this.list.Title = "Main Menu"
    this.list.SetWidth(WIDTH)
    this.list.SetHeight(HEIGHT)
    this.list.AutoHeight = false
    //this.list.selected = 0
    
    this.tasks = make([]Task, 0, len(tasks))
    for i, n := range names {
        this.list.AddItem(n, 0, mainMenuAction, tasks[i].Name())
        this.tasks = append(this.tasks, tasks[i])
    }
    
    this.view.AddChild(this.list)
    
    return this
}

func (m *mainMenu) Open(attrs mpd.Attrs) {
    fmt.Printf("%+v\n", attrs)
}

func (m *mainMenu) Hide(attrs mpd.Attrs) {
    fmt.Printf("%+v\n", attrs)
}

func (m *mainMenu) End(attrs mpd.Attrs) {
    fmt.Printf("%+v\n", attrs)
}

func (m *mainMenu) Update(attrs []string) {
    fmt.Printf("%+v\n", attrs)
}

func (m *mainMenu) HandleInput(key rune) bool {
    /*
    switch key {
        case input.KEY_UP, input.KEY_SCROLLUP:
            if (m.list.Selected > 0) {
                m.list.Selected--
            }
            return true
        case input.KEY_DOWN, input.KEY_SCROLLDOWN:
            if (m.list.Selected < m.list.NumItems() - 1) {
                m.list.Selected++
            }
            return true
        case input.KEY_ENTER:
            li := m.list.Item(m.list.Selected)
            li.Action(li)
            return true
    }
    */
    return m.list.HandleInput(key)
}

func (m *mainMenu) Draw(to draw.Image) image.Rectangle {
    return m.view.Draw(to)
}

func mainMenuAction(v ...interface{}) {
    if item, ok := v[0].(*ui.ListItem); ok {
        if name, ok := item.Tag.(string); ok {
            fmt.Println("Pressed enter on item", name)
            i := Intent{ INTENT_OPEN, name, 
                         map[string]string{"red": "255", "blue": "255"}}
            Intents <- i
            fmt.Println("Sent intent")
            return
        }
    }
    fmt.Println("Seems like an invalid menu item")
}

func (m *mainMenu) ModuleName() string {
    return "Main Menu"
}

func (m *mainMenu) Name() string {
    return "mainmenu"
}
