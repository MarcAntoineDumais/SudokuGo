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
    return grid{n, g}
}

func (gr *grid) solve() bool {
    defer timeTrack(time.Now(), "Solver")

    //Solve the sudoku
    rows := make([]uint64, gr.n*gr.n)
    columns := make([]uint64, gr.n*gr.n)
    blocks := make([]uint64, gr.n*gr.n)
    
    for i := range gr.g {
        for j := range gr.g[i] {
            val := gr.g[i][j]
            if val != 0 {
                rows[i] |= 1 << uint(val-1)
                columns[j] |= 1 << uint(val-1)
                blocks[i/gr.n * gr.n + j/gr.n] |= 1 << uint(val-1)
            }
        }
    }
    
    var recurse func(i, j int) bool
    recurse = func(i, j int) bool {
        //fmt.Println(i, j)
        for ; i < len(gr.g); i++ {
            for ; j < len(gr.g[i]); j++ {
                blockI := i/gr.n * gr.n + j/gr.n
                curRow := rows[i]
                curCol := columns[j]
                curBlock := blocks[blockI]
                
                if gr.g[i][j] != 0 {
                    continue
                }
                for v := 0; v < gr.n*gr.n; v++ {
                    if (curRow & (1 << uint(v))) != 0 ||
                       (curCol & (1 << uint(v))) != 0 ||
                       (curBlock & (1 << uint(v))) != 0 {
                       continue
                    }
                    //fmt.Println("v ", v)
                    gr.g[i][j] = v + 1
                    rows[i] |= 1 << uint(v)
                    columns[j] |= 1 << uint(v)
                    blocks[blockI] |= 1 << uint(v)
                    
                    if recurse(i, j+1) {
                        return true
                    } else {
                        rows[i] = curRow
                        columns[j] = curCol
                        blocks[blockI] = curBlock
                        gr.g[i][j] = 0
                    }
                }
                return false
            }
            j = 0
        }
        return true
    }
    
    return recurse(0, 0)
    
    /*fmt.Println("rows")
    for i, k := range rows {
        fmt.Printf("%d: %s\n", i, strconv.FormatInt(int64(k), 2))
    }
    fmt.Println("columns")
    for i, k := range columns {
        fmt.Printf("%d: %s\n", i, strconv.FormatInt(int64(k), 2))
    }
    fmt.Println("blocks")
    for i, k := range blocks {
        fmt.Printf("%d: %s\n", i, strconv.FormatInt(int64(k), 2))
    }
    */
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
