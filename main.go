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

    avail := make([][]uint64, gr.n*gr.n)
    for i := range gr.g {
        avail[i] = make([]uint64, gr.n*gr.n)
        for j := range avail[i] {
            for k := 0; k < gr.n*gr.n; k++ {
                avail[i][j] |= 1 << uint(k)
            }
        }
    }

    for i := range gr.g {
        for j, val := range gr.g[i] {
            if val != 0 {
                blockI := i/gr.n * gr.n + j/gr.n
                for k := range gr.g {
                    if k != j {
                        avail[i][k] &= ^(1 << uint(val - 1))
                    }

                    if k != i {
                        avail[k][j] &= ^(1 << uint(val - 1))
                    }

                    blockY := blockI/gr.n * gr.n + k/gr.n
                    blockX := blockI%gr.n * gr.n + k%gr.n
                    if blockY != i || blockX != j {
                        avail[blockY][blockX] &= ^(1 << uint(val - 1))
                    }
                }
            }
        }
    }

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

                curAvail := avail[i][j]
                availRow, availCol, availBlock := backupAvail(avail, i, j, gr.n)

                if gr.g[i][j] != 0 {
                    continue
                }
                for v := 0; v < gr.n*gr.n; v++ {
                    mask := uint64(1 << uint(v))
                    if (avail[i][j] & mask) == 0 ||
                       (curRow & mask) != 0 ||
                       (curCol & mask) != 0 ||
                       (curBlock & mask) != 0 {
                       continue
                    }
                    //fmt.Println("v ", v)
                    gr.g[i][j] = v + 1
                    rows[i] |= mask
                    columns[j] |= mask
                    blocks[blockI] |= mask
                    for k := 0; k < gr.n*gr.n; k++ {
                        avail[i][k] &= ^mask
                        avail[k][j] &= ^mask

                        blockY := blockI/gr.n * gr.n + k/gr.n
                        blockX := blockI%gr.n * gr.n + k%gr.n
                        avail[blockY][blockX] &= ^mask
                    }

                    if recurse(i, j+1) {
                        return true
                    } else {
                        rows[i] = curRow
                        columns[j] = curCol
                        blocks[blockI] = curBlock
                        gr.g[i][j] = 0
                        restoreAvail(avail, availRow, availCol, availBlock, i, j, gr.n)
                    }
                }
                avail[i][j] = curAvail
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

func backupAvail(avail [][]uint64, i, j, n int) (row, col, block []uint64) {
    row = make([]uint64, n*n)
    col = make([]uint64, n*n)
    block = make([]uint64, n*n)

    for k := range avail {
        row[k] = avail[i][k]
        col[k] = avail[k][j]
        blockI := i/n * n + j/n
        block[k] = avail[blockI/n * n + k/n][blockI%n * n + k%n]
    }
    return
}

func restoreAvail(avail [][]uint64, row, col, block []uint64, i, j, n int) {
    for k := range avail {
        if k != j {
            avail[i][k] = row[k]
        }

        if k != i {
            avail[k][j] = col[k]
        }

        blockI := i/n * n + j/n
        blockY := blockI/n * n + k/n
        blockX := blockI%n * n + k%n
        if blockY != i || blockX != j {
            avail[blockY][blockX] = block[k]
        }
    }
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
