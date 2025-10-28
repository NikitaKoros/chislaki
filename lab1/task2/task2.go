// Система_уравнений_4.pdf
package main

import "fmt"

const N = 5

func main() {
    // Матрица коэффициентов A (5x5)
    A := [][]float64{
        {6, -5, -3, 2, 3},
        {4, 2, 1, -5, 2},
        {-6, 3, 5, 2, 7},
        {2, 3, -4, 1, 3},
        {3, 5, -5, -6, 2},
    }

    // Вектор правых частей b
    b := []float64{73, -20, 69, 59, 10}
    
    // Копия для проверки
    A_original := make([][]float64, N)
    for i := 0; i < 5; i++ {
        A_original[i] = make([]float64, N)
        copy(A_original[i], A[i])
    }

    // Выполняем LU-разложение внутри A
    luInPlace(A)

    // Решение системы Ax = b
    x := solveLUInPlace(A, b)
    printVector(x, "Решение системы:")

    // Определитель матрицы A (произведение диагоналей U)
    det := determinantInPlace(A)
    fmt.Printf("\nОпределитель матрицы A: %.2f\n", det)

    // Обратная матрица
    invA := inverseMatrixInPlace(A)
    printMatrix(invA, "Обратная матрица A^(-1):")
    
    fmt.Println("\nПроверка: A * A^-1 (должно быть ~E):")
    result := make([][]float64, N)
    for i := 0; i < N; i++ {
        result[i] = make([]float64, N)
        for j := 0; j < N; j++ {
            for k := 0; k < N; k++ {
                result[i][j] += A_original[i][k] * invA[k][j]
            }
        }
    }
    printMatrix(result, "A * A^-1")
}

// luInPlace - выполняет LU-разложение внутри матрицы A
func luInPlace(A [][]float64) {
    n := len(A)
    for k := 0; k < n; k++ {
        // Шаг 1: обновление верхней части (U)
        for j := k; j < n; j++ {
            sum := 0.0
            for m := 0; m < k; m++ {
                sum += A[k][m] * A[m][j]
            }
            A[k][j] -= sum
        }

        // Шаг 2: обновление нижней части (L)
        for i := k + 1; i < n; i++ {
            sum := 0.0
            for m := 0; m < k; m++ {
                sum += A[i][m] * A[m][k]
            }
            A[i][k] = (A[i][k] - sum) / A[k][k]
        }
    }
}

// forwardSubstitutionInPlace - решает Ly = b, используя нижнюю часть A
func forwardSubstitutionInPlace(A [][]float64, b []float64) []float64 {
    n := len(b)
    y := make([]float64, n)
    for i := 0; i < n; i++ {
        sum := 0.0
        for j := 0; j < i; j++ {
            sum += A[i][j] * y[j]
        }
        y[i] = b[i] - sum
    }
    return y
}

// backwardSubstitutionInPlace - решает Ux = y, используя верхнюю часть A
func backwardSubstitutionInPlace(A [][]float64, y []float64) []float64 {
    n := len(y)
    x := make([]float64, n)
    for i := n - 1; i >= 0; i-- {
        sum := 0.0
        for j := i + 1; j < n; j++ {
            sum += A[i][j] * x[j]
        }
        x[i] = (y[i] - sum) / A[i][i]
    }
    return x
}

// solveLUInPlace - решает Ax = b с использованием LU-разложения внутри A
func solveLUInPlace(A [][]float64, b []float64) []float64 {
    y := forwardSubstitutionInPlace(A, b)
    x := backwardSubstitutionInPlace(A, y)
    return x
}

// determinantInPlace - определитель через произведение диагоналей U
func determinantInPlace(A [][]float64) float64 {
    det := 1.0
    for i := 0; i < len(A); i++ {
        det *= A[i][i]
    }
    return det
}

// inverseMatrixInPlace - строит обратную матрицу, используя LU-разложение
func inverseMatrixInPlace(A [][]float64) [][]float64 {
    n := len(A)
    inv := make([][]float64, n)
    for i := 0; i < n; i++ {
        inv[i] = make([]float64, n)
    }

    // Для каждого столбца единичной матрицы e_j
    for j := 0; j < n; j++ {
        ej := make([]float64, n)
        ej[j] = 1.0

        y := forwardSubstitutionInPlace(A, ej)
        x := backwardSubstitutionInPlace(A, y)

        // Сохраняем столбец x в обратной матрице
        for i := 0; i < n; i++ {
            inv[i][j] = x[i]
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