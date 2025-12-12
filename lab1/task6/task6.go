// Системы_уравнений_QR.pdf
package main

import (
	"fmt"
	"math"
)

type Matrix [][]float64

func main() {

	A := Matrix{
		{2, -7, -4, 3, 2},
		{4, 0, 1, 9, -2},
		{5, -3, 2, -3, 7},
		{2, -1, 4, 11, 3},
		{3, 3, -5, 1, -2},
	}

	fmt.Println("Исходная матрица:")
	printMatrix(A)

	eigenvalues, finalA := QRAlgorithm(A, 1000, 1e-10)

	fmt.Println("\nИтоговая матрица A (после QR-алгоритма):")
	printMatrix(finalA)

	fmt.Println("\nСобственные значения:")
	for i, lambda := range eigenvalues {
		if imag(lambda) == 0 {
			fmt.Printf("λ%d = %.6f\n", i+1, real(lambda))
		} else {
			fmt.Printf("λ%d = %.6f + %.6fi\n", i+1, real(lambda), imag(lambda))
		}
	}
}

// QRAlgorithm - основной алгоритм QR-разложения для нахождения собственных значений
func QRAlgorithm(A Matrix, maxIterations int, epsilon float64) ([]complex128, Matrix) {
	currentA := copyMatrix(A)
	var lastQ Matrix

	for iter := 0; iter < maxIterations; iter++ {
		// QR-разложение Хаусхолдером (ТУТ ВНУТРЕННИЙ ЦИКЛ!!!)
		Q, R := QRDecomposition(currentA, epsilon)
		lastQ = Q

		// Вывод матрицы Q на первой итерации
		if iter == 0 {
			fmt.Println("\nМатрица Q (первая итерация):")
			printMatrix(Q)
		}

		// Обновление матрицы: A_new = R * Q
		newA := multiplyMatrices(R, Q)

		// Проверка сходимости (матрица становится верхней квазитреугольной?)
		if isQuasiUpperTriangular(newA, epsilon) {
			fmt.Printf("Сходимость достигнута за %d итераций\n", iter+1)
			fmt.Println("\nМатрица Q (последняя итерация):")
			printMatrix(lastQ)
			currentA = newA
			break
		}

		currentA = newA

		if iter == maxIterations-1 {
			fmt.Printf("Достигнуто максимальное число итераций: %d\n", maxIterations)
			fmt.Println("\nМатрица Q (последняя итерация):")
			printMatrix(lastQ)
		}
	}

	// Извлечение собственных значений из квазитреугольной матрицы
	return extractEigenvalues(currentA, epsilon), currentA
}

// QRDecomposition - QR-разложение методом Хаусхолдера
func QRDecomposition(A Matrix, epsilon float64) (Matrix, Matrix) {
	n := len(A)
	Q := identityMatrix(n)
	R := copyMatrix(A)

	for k := 0; k < n-1; k++ {
		// Вектор x из k-го столбца, начиная с k-й строки
		x := make([]float64, n-k)
		for i := k; i < n; i++ {
			x[i-k] = R[i][k]
		}

		// Норма вектора x
		normX := norm(x)

		sign := 1.0
		if x[0] < 0 {
			sign = -1.0
		}

		// Вектор v
		v := make([]float64, n-k)
		v[0] = x[0] + sign*normX
		for i := 1; i < len(x); i++ {
			v[i] = x[i]
		}

		// Норма вектора v
		normV := norm(v)
		if normV < epsilon {
			continue
		}

		// Нормализация v
		for i := range v {
			v[i] /= normV
		}

		// Матрица Хаусхолдера H = I - 2*v*v^T
		H := identityMatrix(n)
		for i := k; i < n; i++ {
			for j := k; j < n; j++ {
				H[i][j] -= 2 * v[i-k] * v[j-k]
			}
		}

		// Применение преобразования: R = H * R
		R = multiplyMatrices(H, R)

		// Накопление Q: Q = Q * H^T (H симметрична)
		Q = multiplyMatrices(Q, H)
	}

	return Q, R
}

// extractEigenvalues - извлечение собственных значений из квазитреугольной матрицы
func extractEigenvalues(A Matrix, epsilon float64) []complex128 {
	n := len(A)
	eigenvalues := make([]complex128, 0, n)

	i := 0
	for i < n {
		if i == n-1 || math.Abs(A[i+1][i]) < epsilon {
			// Вещественное собственное значение
			eigenvalues = append(eigenvalues, complex(A[i][i], 0))
			i++
		} else {
			// Блок 2x2 для комплексных собственных значений
			a := A[i][i]
			b := A[i][i+1]
			c := A[i+1][i]
			d := A[i+1][i+1]

			// Вычисляем собственные значения блока 2x2
			trace := a + d
			det := a*d - b*c

			discriminant := trace*trace - 4*det
			if discriminant >= 0 {
				// Вещественные собственные значения
				lambda1 := (trace + math.Sqrt(discriminant)) / 2
				lambda2 := (trace - math.Sqrt(discriminant)) / 2
				eigenvalues = append(eigenvalues, complex(lambda1, 0), complex(lambda2, 0))
			} else {
				// Комплексные собственные значения
				realPart := trace / 2
				imagPart := math.Sqrt(-discriminant) / 2
				eigenvalues = append(eigenvalues, complex(realPart, imagPart), complex(realPart, -imagPart))
			}
			i += 2
		}
	}

	return eigenvalues
}

// наивная реализация
// isQuasiUpperTriangular - проверка, является ли матрица верхней квазитреугольной
// (с блоками 2x2 на диагонали, которые не разлагаются далее)
// func isQuasiUpperTriangular(A Matrix, epsilon float64) bool {
// 	n := len(A)
// 	for i := 1; i < n; i++ {
// 		// пропускаем элементы непосредственно под диагональю, которые могут быть ненулевыми в блоках 2x2
// 		if i > 0 && math.Abs(A[i][i-1]) > epsilon {
// 			// проверяем, что это действительно блок 2x2, а не элемент, который должен быть нулевым
// 			if i == 1 || math.Abs(A[i-1][i-2]) < epsilon {
// 				// так как это может быть блок 2x2, продолжаем проверку
// 				continue
// 			}
// 		}

// 		// Для всех остальных элементов под диагональю проверяем, что они близки к нулю
// 		for j := 0; j < i-1; j++ {
// 			if math.Abs(A[i][j]) > epsilon {
// 				return false
// 			}
// 		}
// 	}
// 	return true
// }

//проверка через сумму квадратов
// isQuasiUpperTriangular - проверка сходимости по критерию суммы квадратов поддиагональных элементов
// func isQuasiUpperTriangular(A Matrix, epsilon float64) bool {
// 	n := len(A)

// 	// Проходим по столбцам от 0 до n-1
// 	for m := 0; m < n-1; m++ {
// 		// Вычисляем сумму квадратов элементов ниже диагонали в m-м столбце
// 		sumSquares := 0.0
// 		for l := m + 1; l < n; l++ {
// 			sumSquares += A[l][m] * A[l][m]
// 		}

// 		// Проверяем критерий сходимости для этого столбца
// 		if math.Sqrt(sumSquares) > epsilon {
// 			// Проверяем, не является ли это частью блока 2x2
// 			// Блок 2x2 возможен только если m+1 < n и элемент A[m+1][m] большой
// 			if m+1 < n && math.Abs(A[m+1][m]) > epsilon {
// 				// Это может быть блок 2x2, проверяем что остальные элементы малы
// 				sumSquaresBelow := 0.0
// 				for l := m + 2; l < n; l++ {
// 					sumSquaresBelow += A[l][m] * A[l][m]
// 				}
// 				if math.Sqrt(sumSquaresBelow) > epsilon {
// 					return false
// 				}
// 			} else {
// 				return false
// 			}
// 		}
// 	}

// 	return true
// }

// проверка через пересчет нулевых элементов
// isQuasiUpperTriangular - проверка с подсчетом количества нулей под диагональю
func isQuasiUpperTriangular(A Matrix, epsilon float64) bool {
	n := len(A)
	totalSubdiagonalElements := 0
	zeroElements := 0

	// Подсчитываем общее количество элементов под главной диагональю
	// и количество "нулевых" элементов
	for i := 1; i < n; i++ {
		for j := 0; j < i; j++ {
			totalSubdiagonalElements++
			if math.Abs(A[i][j]) <= epsilon {
				zeroElements++
			}
		}
	}

	// Учитываем возможные блоки 2x2 на диагонали
	// (элементы на поддиагонали могут быть ненулевыми)
	blocksCount := 0
	for i := 1; i < n; i++ {
		if math.Abs(A[i][i-1]) > epsilon {
			blocksCount++
		}
	}

	// Требуем, чтобы все элементы кроме возможных поддиагональных в блоках 2x2 были нулевыми
	requiredZeros := totalSubdiagonalElements - blocksCount

	return zeroElements >= requiredZeros
}

func copyMatrix(A Matrix) Matrix {
	n := len(A)
	copyMat := make(Matrix, n)
	for i := range A {
		copyMat[i] = make([]float64, len(A[i]))
		copy(copyMat[i], A[i])
	}
	return copyMat
}

// identityMatrix - создание единичной матрицы
func identityMatrix(n int) Matrix {
	I := make(Matrix, n)
	for i := range I {
		I[i] = make([]float64, n)
		I[i][i] = 1.0
	}
	return I
}

func multiplyMatrices(A, B Matrix) Matrix {
	n := len(A)
	m := len(B[0])
	p := len(B)

	result := make(Matrix, n)
	for i := range result {
		result[i] = make([]float64, m)
		for j := 0; j < m; j++ {
			for k := 0; k < p; k++ {
				result[i][j] += A[i][k] * B[k][j]
			}
		}
	}
	return result
}

// norm - вычисление евклидовой нормы вектора
func norm(v []float64) float64 {
	sum := 0.0
	for _, x := range v {
		sum += x * x
	}
	return math.Sqrt(sum)
}

func printMatrix(A Matrix) {
	for i := range A {
		for j := range A[i] {
			fmt.Printf("%8.4f ", A[i][j])
		}
		fmt.Println()
	}
}
