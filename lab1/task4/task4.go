package main

import (
	"fmt"
	"math"
)

func main() {
	// A := [][]float64{
	// 	{-5, 13, -2, -4},
	// 	{ 3, -2,  4, -13},
	// 	{12, -4, -1, -5},
	// 	{-2,  5, -14, 3},
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
	eps := 0.0001

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
		fmt.Println("Остановка: достаточное условие сходимости не выполнено — метод простых итераций применять нельзя.")
		return
	}

	// Вычисляем alpha и beta
	alpha := make([][]float64, n)
	beta := make([]float64, n)
	for i := 0; i < n; i++ {
		alpha[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if i == j {
				alpha[i][j] = 0
			} else {
				alpha[i][j] = -A[i][j] / A[i][i]
			}
		}
		beta[i] = b[i] / A[i][i]
	}

	// Начальное приближение
	xOld := make([]float64, n)
	copy(xOld, beta)

	xNew := make([]float64, n)

	iterations := 0
	for {
		iterations++
		// шаг итерации
		for i := 0; i < n; i++ {
			sum := beta[i]
			for j := 0; j < n; j++ {
				sum += alpha[i][j] * xOld[j]
			}
			xNew[i] = sum
		}

		// проверка сходимости
		converged := true
		for i := 0; i < n; i++ {
			if math.Abs(xNew[i]-xOld[i]) >= eps {
				converged = false
				break
			}
		}
		if converged {
			break
		}
		copy(xOld, xNew)
	}

	fmt.Println("Решение:")
	for i := 0; i < n; i++ {
		fmt.Printf("x%d = %.6f\n", i+1, xNew[i])
	}
	fmt.Printf("Количество итераций: %d\n", iterations)
}
