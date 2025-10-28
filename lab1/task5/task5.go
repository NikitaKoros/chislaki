// Системы_уравнений_6.pdf
package main

import (
	"fmt"
	"math"
)

func main() {
	// A := [][]float64{
	// 	{-5, 13, -2, -4},
	// 	{3, -2, 4, -13},
	// 	{12, -4, -1, -5},
	// 	{-2, 5, -14, 3},
	// }

	// b := []float64{25, 71, -23, 40}

	A := [][]float64{
		{12, -4, -1, -5},
		{-5, 13, -2, -4},
		{-2, 5, -14, 3},
		{3, -2, 4, -13},
	}
	
	b := []float64{-23, 25, 40, 71}

	n := len(A)

	// Проверка условия сходимости (по строкам или по столбцам)
	convergesLine := true
	convergesColumn := true
	for i := 0; i < n; i++ {
		// строка
		rowSum := 0.0
		for j := 0; j < n; j++ {
			if i != j {
				rowSum += math.Abs(A[i][j])
			}
		}
		if math.Abs(A[i][i]) <= rowSum {
			fmt.Printf("Условие сходимости по строке %d не выполнено\n", i+1)
			convergesLine = false
		}

		// столбец
		colSum := 0.0
		for j := 0; j < n; j++ {
			if i != j {
				colSum += math.Abs(A[j][i])
			}
		}
		if math.Abs(A[i][i]) <= colSum {
			fmt.Printf("Условие сходимости по столбцу %d не выполнено\n", i+1)
			convergesColumn = false
		}
	}

	if !(convergesLine || convergesColumn) {
		fmt.Println("Остановка: достаточное условие сходимости не выполнено — метод Зейделя применять нельзя.")
		return
	}

	x := make([]float64, n)
	newX := make([]float64, n)
	eps := 1e-6

	iterations := 0
	for {
		iterations++
		for i := 0; i < n; i++ {
			sum := b[i]
			for j := 0; j < n; j++ {
				if j != i {
					// Используем новые значения, если они уже вычислены
					if j < i {
						sum -= A[i][j] * newX[j]
					} else {
						sum -= A[i][j] * x[j]
					}
				}
			}
			newX[i] = sum / A[i][i]
		}

		converged := true
		for i := 0; i < n; i++ {
			if math.Abs(newX[i]-x[i]) >= eps {
				converged = false
				break
			}
		}

		copy(x, newX)

		if converged {
			break
		}
	}

	fmt.Println("Решение системы:")
	for i := 0; i < n; i++ {
		fmt.Printf("x%d = %.6f\n", i+1, x[i])
	}
	fmt.Println("Количество итераций:", iterations)
}
