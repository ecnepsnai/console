# Deprecated: please use github.com/ecnepsnai/logtic or something else.

---

# console
A go console implementation

# Installation

```
go get github.com/ecnepsnai/console
```

# Usage

```golang
package main

import "github.com/ecnepsnai/console"

func main() {
    Console, err := console.New(logPath, console.LevelDebug)
    if err != nil {
        panic(err.Error())
    }

    Console.Debug("Hopefully this helps")
    Console.Info("Hey, listen!")
    Console.Warn("Uh oh, what did you do?")
    Console.Error("You've ruined EVERYTHING!")
}
```
