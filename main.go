package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
    "time"
)

func main() {
    if len(os.Args) == 1 {
        fmt.Println("Missing sudoku file name. \nUsage: SudokuGo filename")
        return
    }

    g := loadGrid(os.Args[1])
    fmt.Println(g.String())
    g.solve()
    fmt.Println(g.String())
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    fmt.Printf("%s took %s\n", name, elapsed)
}

type grid struct {
    n uint8
    g [][]uint8
}

func loadGrid(filename string) grid {
    file, err := ioutil.ReadFile(filename)
    check(err)
    s := string(file)
    lines := strings.Split(s, "\n")
    for i := range lines {
        lines[i] = strings.TrimSpace(lines[i])
    }
    n, _ := strconv.Atoi(lines[0])
    lines = lines[2:]

    g := make([][]uint8, n*n)
    for i := range g {
        g[i] = make([]uint8, n*n)
    }

    realRow := 0
    for i := 0; i < n*n+(n-1); i++ {
        if i%(n+1) == n {
            continue
        }
        realCol := 0
        line := strings.Split(lines[i], " ")

        for j := 0; j < n*n+(n-1); j++ {
            if j%(n+1) == n {
                continue
            }
            if line[j] == "x" {
                g[realRow][realCol] = uint8(0)
            } else {
                tmp, _ := strconv.Atoi(line[j])
                g[realRow][realCol] = uint8(tmp)
            }
            realCol++
        }
        realRow++
    }
    return grid{uint8(n), g}
}

func (gr *grid) solve() {
    defer timeTrack(time.Now(), "Solver")

    //Solve the sudoku
}

func (gr *grid) String() string {
    s := fmt.Sprintf("Sudoku of size %d\n", gr.n)
    digitSize := 1
    if gr.n > 3 {
        digitSize = 2
    }
    for i, row := range gr.g {
        for j, v := range row {
            if gr.n>3 && v <= 9 {
                s += " "
            }
            s += fmt.Sprintf("%d ", v)
            if j%int(gr.n) == int(gr.n-1) && j != len(row)-1 {
                s += "| "
            }
        }
        if i%int(gr.n) == int(gr.n-1) && i != len(gr.g)-1 {
            s += "\n"
            for j := 0; j < int(gr.n*gr.n)*(digitSize+1) + (int(gr.n)-1)*2 - 1; j++ {
                s += "-"
            }
        }
        s += "\n"
    }
    return s
}
