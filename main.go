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
    if g.solve() {
        fmt.Println(g.String())
    } else {
        fmt.Println("Could not find a solution for this sudoku.")
    }
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
    n int
    g [][]int
    choices [][]uint64
}

type action struct {
    x, y int
    val int
    logic bool
    row, col, block []uint64
}

func (gr *grid) apply(a action) {
    a.row = make([]uint64, gr.n*gr.n)
    a.col = make([]uint64, gr.n*gr.n)
    a.block = make([]uint64, gr.n*gr.n)
    blockI := a.y/gr.n * gr.n + a.x/gr.n
    
    for k := range gr.choices {
        a.row[k] = gr.choices[a.y][k]
        a.col[k] = gr.choices[k][a.x]
        blockI := a.y/gr.n * gr.n + a.x/gr.n
        a.block[k] = gr.choices[blockI/gr.n * gr.n + k/gr.n][blockI%gr.n * gr.n + k%gr.n]
    }
    
    mask := uint64(1 << uint(a.val - 1))
    gr.g[a.y][a.x] = a.val
    for k := 0; k < gr.n*gr.n; k++ {
        gr.choices[a.y][k] &= ^mask
        gr.choices[k][a.x] &= ^mask

        blockY := blockI/gr.n * gr.n + k/gr.n
        blockX := blockI%gr.n * gr.n + k%gr.n
        gr.choices[blockY][blockX] &= ^mask
    }
}

func (gr *grid) undo(a action) {
    gr.g[a.y][a.x] = 0
    blockI := a.y/gr.n * gr.n + a.x/gr.n
    
    for k := range gr.choices {
        if k != a.y {
            gr.choices[a.y][k] = a.row[k]
        }

        if k != a.y {
            gr.choices[k][a.x] = a.col[k]
        }

        blockY := blockI/gr.n * gr.n + k/gr.n
        blockX := blockI%gr.n * gr.n + k%gr.n
        if blockY != a.y || blockX != a.x {
            gr.choices[blockY][blockX] = a.block[k]
        }
    }
}

//Solve the sudoku
func (gr *grid) solve() bool {
    defer timeTrack(time.Now(), "Solver")
    
    var actions []action
    
    gr.choices = make([][]uint64, gr.n*gr.n)
    for i := range gr.g {
        gr.choices[i] = make([]uint64, gr.n*gr.n)
        for j := range gr.choices[i] {
            for k := 0; k < gr.n*gr.n; k++ {
                gr.choices[i][j] |= 1 << uint(k)
            }
        }
    }

    for i := range gr.g {
        for j, val := range gr.g[i] {
            if val != 0 {
                blockI := i/gr.n * gr.n + j/gr.n
                for k := range gr.g {
                    if k != j {
                        gr.choices[i][k] &= ^(1 << uint(val - 1))
                    }

                    if k != i {
                        gr.choices[k][j] &= ^(1 << uint(val - 1))
                    }

                    blockY := blockI/gr.n * gr.n + k/gr.n
                    blockX := blockI%gr.n * gr.n + k%gr.n
                    if blockY != i || blockX != j {
                        gr.choices[blockY][blockX] &= ^(1 << uint(val - 1))
                    }
                }
            }
        }
    }

    var recurse func(i, j int) bool
    recurse = func(i, j int) bool {
        for ; i < len(gr.g); i++ {
            for ; j < len(gr.g[i]); j++ {
                curChoices := gr.choices[i][j]

                if gr.g[i][j] != 0 {
                    continue
                }
                for v := 0; v < gr.n*gr.n; v++ {
                    mask := uint64(1 << uint(v))
                    if (gr.choices[i][j] & mask) == 0 {
                       continue
                    }
                    a := action{x:j, y:i, val:v+1, logic:false}
                    gr.apply(a)
                    actions = append(actions, a)

                    if recurse(i, j+1) {
                        return true
                    } else {
                        gr.undo(actions[len(actions)-1])
                        actions = actions[:len(actions)-1]
                    }
                }
                gr.choices[i][j] = curChoices
                return false
            }
            j = 0
        }
        return true
    }

    return recurse(0, 0)
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

    g := make([][]int, n*n)
    for i := range g {
        g[i] = make([]int, n*n)
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
                g[realRow][realCol] = 0
            } else {
                g[realRow][realCol], _ = strconv.Atoi(line[j])
            }
            realCol++
        }
        realRow++
    }
    return grid{n, g, nil}
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
            if j%gr.n == gr.n-1 && j != len(row)-1 {
                s += "| "
            }
        }
        if i%gr.n == gr.n-1 && i != len(gr.g)-1 {
            s += "\n"
            for j := 0; j < gr.n*gr.n*(digitSize+1) + (gr.n-1)*2 - 1; j++ {
                s += "-"
            }
        }
        s += "\n"
    }
    return s
}
