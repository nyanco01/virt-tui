package main

import (


    "github.com/nyanco01/virt-tui/src/tui"
    "github.com/nyanco01/virt-tui/src/operate"

)

func main() {
    operate.Initialize()

    app := tui.MakeApp()

    if err := app.EnableMouse(true).Run(); err != nil {
        panic(err)
    }
}
