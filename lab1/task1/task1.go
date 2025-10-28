package main

import "fmt"

const N = 5

func gaussSolve(A [][]float64, b []float64) ([]float64, float64) {
    // Создаем расширенную матрицу [A | b]
    aug := make([][]float64, N)
    for i := 0; i < N; i++ {
        aug[i] = make([]float64, N+1)
        for j := 0; j < N; j++ {
            aug[i][j] = A[i][j]
        }
        aug[i][N] = b[i]
    }
    
    det := 1.0 // инициализируем определитель

    // Прямой ход
    for k := 0; k < N-1; k++ {
        pivot := aug[k][k]
        det *= pivot
        
        for j := k; j <= N; j++ {
            aug[k][j] = aug[k][j] / pivot
        }

        for i := k + 1; i < N; i++ {
            factor := aug[i][k]
            for j := k; j <= N; j++ {
                aug[i][j] = aug[i][j] - factor * aug[k][j] // вычитаем из каждой строки k-ю строку
            }
        }
    }

    det *= aug[N-1][N-1]
    aug[N-1][N] = aug[N-1][N] / aug[N-1][N-1]
    aug[N-1][N-1] = 1.0

    // Обратный ход
    x := make([]float64, N)
    x[N-1] = aug[N-1][N]

    for i := N - 2; i >= 0; i-- {
        sum := 0.0
        for j := i + 1; j < N; j++ {
            sum += aug[i][j] * x[j]
        }
        x[i] = aug[i][N] - sum
    }

    return x, det
}

func gaussInverse(A [][]float64) [][]float64 {
    n := N
    aug := make([][]float64, n)
    for i := 0; i < n; i++ {
        aug[i] = make([]float64, 2*n)
        for j := 0; j < n; j++ {
            aug[i][j] = A[i][j]
        }
        aug[i][i+n] = 1.0
    }

    for k := 0; k < n; k++ {
        pivot := aug[k][k]
        for j := k; j < 2*n; j++ {
            aug[k][j] = aug[k][j] / pivot
        }

        for i := 0; i < n; i++ {
            if i != k {
                factor := aug[i][k]
                for j := k; j < 2*n; j++ {
                    aug[i][j] = aug[i][j] - factor * aug[k][j]
                }
            }
        }
    }

    // Извлечение результата
    inv := make([][]float64, n)
    for i := 0; i < n; i++ {
        inv[i] = make([]float64, n)
        for j := 0; j < n; j++ {
            inv[i][j] = aug[i][j+n]
        }
    }

    return inv
}

func printVector(v []float64, name string) {
    fmt.Printf("\n%s:\n", name)
    for i := 0; i < len(v); i++ {
        fmt.Printf("x%d = %.4f\n", i+1, v[i])
    }
}

func printMatrix(mat [][]float64, name string) {
    fmt.Printf("\n%s:\n", name)
    for i := 0; i < len(mat); i++ {
        for j := 0; j < len(mat[i]); j++ {
            fmt.Printf("%10.4f ", mat[i][j])
        }
        fmt.Println()
    }
}

func main() {
    A := [][]float64{
        {6, -5, -3, 2, 3},
        {4, 2, 1, -5, 2},
        {-6, 3, 5, 2, 7},
        {2, 3, -4, 1, 3},
        {3, 5, -5, -6, 2},
    }

    b := []float64{73, -20, 69, 59, 10}

    x, det := gaussSolve(A, b)
    fmt.Printf("Определитель матрицы: %10.4f\n", det)
    printVector(x, "Решение системы")

    inv := gaussInverse(A)
    printMatrix(inv, "Обратная матрица")

    fmt.Println("\nПроверка: A * A^-1 (должно быть ~E):")
    result := make([][]float64, N)
    for i := 0; i < N; i++ {
        result[i] = make([]float64, N)
        for j := 0; j < N; j++ {
            for k := 0; k < N; k++ {
                result[i][j] += A[i][k] * inv[k][j]
            }
        }
    }
    printMatrix(result, "A * A^-1")
}