package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Результат решения в точке
type SolutionPoint struct {
	x         float64
	y         float64
	yAnalytic float64
	error     float64
}

// Исходное уравнение: xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0
// Приводим к виду: y'' + p(x)y' + q(x)y = f(x)
// y'' - (2(3x + 2)/x)y' + (3(3x + 4)/x)y = 0

// Коэффициент p(x)
func p(x float64) float64 {
	if math.Abs(x) < 1e-10 {
		return 0
	}
	return -2 * (3*x + 2) / x
}

// Коэффициент q(x)
func q(x float64) float64 {
	if math.Abs(x) < 1e-10 {
		return 0
	}
	return 3 * (3*x + 4) / x
}

// Правая часть f(x)
func f(x float64) float64 {
	return 0
}

// Аналитическое решение: y(x) = (1+x^5)e^(3x)
func analyticalSolution(x float64) float64 {
	return (1 + math.Pow(x, 5)) * math.Exp(3*x)
}

// Функция для построения матрицы системы 1-го порядка
func buildFirstOrderMatrix(a, b, h, ya, yb float64) ([][]float64, []float64, []float64) {
	n := int((b-a)/h) + 1
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = a + float64(i)*h
	}

	// Создаем полную матрицу для визуализации
	A := make([][]float64, n)
	for i := 0; i < n; i++ {
		A[i] = make([]float64, n)
	}
	d := make([]float64, n)

	// Первое уравнение: (h-1)y[0] + y[1] = ya*h
	A[0][0] = h - 1
	A[0][1] = 1
	d[0] = ya * h

	// Внутренние точки
	for i := 1; i < n-1; i++ {
		px := p(x[i])
		qx := q(x[i])
		fx := f(x[i])

		A[i][i-1] = 1 - px*h/2
		A[i][i] = -2 + h*h*qx
		A[i][i+1] = 1 + px*h/2
		d[i] = h * h * fx
	}

	// Последнее уравнение: y[n-1] = yb
	A[n-1][n-1] = 1
	d[n-1] = yb

	return A, d, x
}

// Функция для построения матрицы системы 2-го порядка
func buildSecondOrderMatrix(a, b, h, ya, yb float64) ([][]float64, []float64, []float64) {
	n := int((b-a)/h) + 1
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = a + float64(i)*h
	}

	A := make([][]float64, n)
	for i := 0; i < n; i++ {
		A[i] = make([]float64, n)
	}
	d := make([]float64, n)

	// Первое уравнение: (2h - 3)y[0] + 4y[1] - y[2] = 2h*ya
	A[0][0] = 2*h - 3
	A[0][1] = 4
	if n > 2 {
		A[0][2] = -1
	}
	d[0] = 2 * h * ya

	// Внутренние точки
	for i := 1; i < n-1; i++ {
		px := p(x[i])
		qx := q(x[i])
		fx := f(x[i])

		A[i][i-1] = 1 - px*h/2
		A[i][i] = -2 + h*h*qx
		A[i][i+1] = 1 + px*h/2
		d[i] = h * h * fx
	}

	// Последнее уравнение: y[n-1] = yb
	A[n-1][n-1] = 1
	d[n-1] = yb

	return A, d, x
}

// Функция для форматирования матрицы в строку
func formatMatrix(A [][]float64, d []float64, x []float64, title string) string {
	n := len(A)
	var text string

	text += title + "\n"
	text += strings.Repeat("=", 120) + "\n\n"
	text += fmt.Sprintf("Размер системы: %d × %d\n\n", n, n)

	// Заголовок столбцов
	text += "    | "
	for j := 0; j < n; j++ {
		text += fmt.Sprintf("y[%-2d]      ", j)
	}
	text += "|  Правая часть\n"
	text += strings.Repeat("-", 120) + "\n"

	// Строки матрицы
	for i := 0; i < n; i++ {
		text += fmt.Sprintf("%-3d | ", i)
		for j := 0; j < n; j++ {
			if math.Abs(A[i][j]) < 1e-10 {
				text += "    .       "
			} else {
				text += fmt.Sprintf("%10.5f ", A[i][j])
			}
		}
		text += fmt.Sprintf("| %10.5f\n", d[i])
	}

	text += "\n"
	return text
}

// Функция для печати матрицы в терминал (сокращенная версия для больших матриц)
func printMatrixToTerminal(A [][]float64, d []float64, x []float64, title string) {
	n := len(A)

	fmt.Println(strings.Repeat("=", 120))
	fmt.Println(title)
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("\nРазмер системы: %d × %d\n\n", n, n)

	// Для больших матриц показываем только первые и последние строки
	showRows := 5
	if n <= 2*showRows {
		showRows = n
	}

	// Заголовок
	fmt.Print("    | ")
	for j := 0; j < n; j++ {
		fmt.Printf("y[%-2d]      ", j)
	}
	fmt.Println("|  Правая часть")
	fmt.Println(strings.Repeat("-", 120))

	// Первые строки
	for i := 0; i < showRows && i < n; i++ {
		fmt.Printf("%-3d | ", i)
		for j := 0; j < n; j++ {
			if math.Abs(A[i][j]) < 1e-10 {
				fmt.Print("    .       ")
			} else {
				fmt.Printf("%10.5f ", A[i][j])
			}
		}
		fmt.Printf("| %10.5f", d[i])
		fmt.Printf("  (x = %.2f)\n", x[i])
	}

	// Средние строки (если матрица большая)
	if n > 2*showRows {
		fmt.Println("... (средние строки пропущены) ...")

		// Последние строки
		for i := n - showRows; i < n; i++ {
			fmt.Printf("%-3d | ", i)
			for j := 0; j < n; j++ {
				if math.Abs(A[i][j]) < 1e-10 {
					fmt.Print("    .       ")
				} else {
					fmt.Printf("%10.5f ", A[i][j])
				}
			}
			fmt.Printf("| %10.5f", d[i])
			fmt.Printf("  (x = %.2f)\n", x[i])
		}
	}

	fmt.Println()
}

// Метод прогонки для решения трехдиагональной системы
// a[i]*y[i-1] + b[i]*y[i] + c[i]*y[i+1] = d[i]
func tridiagonalSolve(a, b, c, d []float64, n int) []float64 {
	// Прямой ход
	alpha := make([]float64, n)
	beta := make([]float64, n)

	alpha[0] = -c[0] / b[0]
	beta[0] = d[0] / b[0]

	for i := 1; i < n; i++ {
		denom := b[i] + a[i]*alpha[i-1]
		if math.Abs(denom) < 1e-15 {
			denom = 1e-15
		}
		alpha[i] = -c[i] / denom
		beta[i] = (d[i] - a[i]*beta[i-1]) / denom
	}

	// Обратный ход
	y := make([]float64, n)
	y[n-1] = beta[n-1]
	for i := n - 2; i >= 0; i-- {
		y[i] = alpha[i]*y[i+1] + beta[i]
	}

	return y
}

// Решение краевой задачи с 1-м порядком точности
func SolveFirstOrder(a, b, h, ya, yb float64) []SolutionPoint {
	// Количество точек
	n := int((b-a)/h) + 1
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = a + float64(i)*h
	}

	// Строим полную матрицу системы
	A := make([][]float64, n)
	for i := 0; i < n; i++ {
		A[i] = make([]float64, n)
	}
	d := make([]float64, n)

	// Первое уравнение: y[0] + (y[1] - y[0])/h = ya
	// (h-1)y[0] + y[1] = ya*h
	A[0][0] = h - 1
	A[0][1] = 1
	d[0] = ya * h

	// Внутренние точки (2-й порядок для внутренних точек)
	// y''[i] + p(x[i])y'[i] + q(x[i])y[i] = f(x[i])
	// (y[i+1] - 2y[i] + y[i-1])/h² + p(x[i])(y[i+1] - y[i-1])/(2h) + q(x[i])y[i] = f(x[i])
	// (1 - p(x[i])h/2)y[i-1] + (-2 + h²q(x[i]))y[i] + (1 + p(x[i])h/2)y[i+1] = h²f(x[i])
	for i := 1; i < n-1; i++ {
		px := p(x[i])
		qx := q(x[i])
		fx := f(x[i])

		A[i][i-1] = 1 - px*h/2
		A[i][i] = -2 + h*h*qx
		A[i][i+1] = 1 + px*h/2
		d[i] = h * h * fx
	}

	// Последнее уравнение: y[n-1] = yb
	A[n-1][n-1] = 1
	d[n-1] = yb

	// Решаем систему методом Гаусса
	y := gaussSolve(A, d, n)

	// Формируем результаты
	var results []SolutionPoint
	for i := 0; i < n; i++ {
		yAnalytic := analyticalSolution(x[i])
		results = append(results, SolutionPoint{
			x:         x[i],
			y:         y[i],
			yAnalytic: yAnalytic,
			error:     math.Abs(y[i] - yAnalytic),
		})
	}

	return results
}

// Решение краевой задачи с 2-м порядком точности
func SolveSecondOrder(a, b, h, ya, yb float64) []SolutionPoint {
	// Количество точек
	n := int((b-a)/h) + 1
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = a + float64(i)*h
	}

	// Система с 2-м порядком точности для граничных условий
	// Для первого граничного условия: y[0] + y'[0] = ya
	// y'[0] ≈ (-3y[0] + 4y[1] - y[2])/(2h) (2-й порядок)
	// y[0] + (-3y[0] + 4y[1] - y[2])/(2h) = ya
	// (2h - 3)y[0] + 4y[1] - y[2] = 2h*ya

	// Создаем расширенную матрицу для первой строки (участвуют y[0], y[1], y[2])
	// Для последней строки: y[n-1] = yb (просто)

	// Используем метод Гаусса для общего случая
	// Построим полную матрицу системы
	A := make([][]float64, n)
	for i := 0; i < n; i++ {
		A[i] = make([]float64, n)
	}
	d := make([]float64, n)

	// Первое уравнение: (2h - 3)y[0] + 4y[1] - y[2] = 2h*ya
	A[0][0] = 2*h - 3
	A[0][1] = 4
	if n > 2 {
		A[0][2] = -1
	}
	d[0] = 2 * h * ya

	// Внутренние точки (2-й порядок)
	// (1 - p(x[i])h/2)y[i-1] + (-2 + h²q(x[i]))y[i] + (1 + p(x[i])h/2)y[i+1] = h²f(x[i])
	for i := 1; i < n-1; i++ {
		px := p(x[i])
		qx := q(x[i])
		fx := f(x[i])

		A[i][i-1] = 1 - px*h/2
		A[i][i] = -2 + h*h*qx
		A[i][i+1] = 1 + px*h/2
		d[i] = h * h * fx
	}

	// Последнее уравнение: y[n-1] = yb
	A[n-1][n-1] = 1
	d[n-1] = yb

	// Решаем систему методом Гаусса
	y := gaussSolve(A, d, n)

	// Формируем результаты
	var results []SolutionPoint
	for i := 0; i < n; i++ {
		yAnalytic := analyticalSolution(x[i])
		results = append(results, SolutionPoint{
			x:         x[i],
			y:         y[i],
			yAnalytic: yAnalytic,
			error:     math.Abs(y[i] - yAnalytic),
		})
	}

	return results
}

// Метод Гаусса
func gaussSolve(A [][]float64, b []float64, n int) []float64 {
	// Создаем расширенную матрицу [A | b]
	aug := make([][]float64, n)
	for i := 0; i < n; i++ {
		aug[i] = make([]float64, n+1)
		for j := 0; j < n; j++ {
			aug[i][j] = A[i][j]
		}
		aug[i][n] = b[i]
	}

	// Прямой ход
	for k := 0; k < n-1; k++ {
		pivot := aug[k][k]

		// Нормализация k-й строки (делим на pivot)
		for j := k; j <= n; j++ {
			aug[k][j] = aug[k][j] / pivot
		}

		// Исключение: вычитаем из каждой строки k-ю строку
		for i := k + 1; i < n; i++ {
			factor := aug[i][k]
			for j := k; j <= n; j++ {
				aug[i][j] = aug[i][j] - factor*aug[k][j]
			}
		}
	}

	// Нормализация последней строки
	aug[n-1][n] = aug[n-1][n] / aug[n-1][n-1]
	aug[n-1][n-1] = 1.0

	// Обратный ход
	x := make([]float64, n)
	x[n-1] = aug[n-1][n]

	for i := n - 2; i >= 0; i-- {
		sum := 0.0
		for j := i + 1; j < n; j++ {
			sum += aug[i][j] * x[j]
		}
		x[i] = aug[i][n] - sum
	}

	return x
}

// Создание общего графика с обоими методами
func createCombinedPlot(results1, results2 []SolutionPoint, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	// Определяем границы по данным
	xMin := results1[0].x
	xMax := results1[len(results1)-1].x
	yMin, yMax := math.Inf(1), math.Inf(-1)

	for _, p := range results1 {
		if p.yAnalytic < yMin {
			yMin = p.yAnalytic
		}
		if p.yAnalytic > yMax {
			yMax = p.yAnalytic
		}
	}
	yRange := yMax - yMin
	yMin -= yRange * 0.1
	yMax += yRange * 0.1

	marginLeft, marginRight := 80.0, 40.0
	marginTop, marginBottom := 40.0, 80.0
	plotWidth := width - marginLeft - marginRight
	plotHeight := height - marginTop - marginBottom

	xToPixel := func(x float64) float32 {
		return float32(marginLeft + (x-xMin)/(xMax-xMin)*plotWidth)
	}
	yToPixel := func(y float64) float32 {
		return float32(marginTop + (yMax-y)/(yMax-yMin)*plotHeight)
	}

	// Сетка
	gridColor := color.RGBA{220, 220, 220, 255}
	xStep := (xMax - xMin) / 10
	for i := 0; i <= 10; i++ {
		x := xMin + float64(i)*xStep
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(x), yToPixel(yMax))
		gridLine.Position2 = fyne.NewPos(xToPixel(x), yToPixel(yMin))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}

	yStep := (yMax - yMin) / 10
	for i := 0; i <= 10; i++ {
		y := yMin + float64(i)*yStep
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(xMin), yToPixel(y))
		gridLine.Position2 = fyne.NewPos(xToPixel(xMax), yToPixel(y))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}

	// Оси
	xAxis := canvas.NewLine(color.Black)
	xAxis.Position1 = fyne.NewPos(xToPixel(xMin), yToPixel(0))
	xAxis.Position2 = fyne.NewPos(xToPixel(xMax), yToPixel(0))
	xAxis.StrokeWidth = 2
	objects = append(objects, xAxis)

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(float32(marginLeft), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(float32(marginLeft), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	// Метки осей X
	for i := 0; i <= 10; i++ {
		x := xMin + float64(i)*xStep
		label := canvas.NewText(fmt.Sprintf("%.2f", x), color.Black)
		label.TextSize = 9
		label.Move(fyne.NewPos(xToPixel(x)-12, yToPixel(0)+10))
		objects = append(objects, label)
	}

	// Метки осей Y
	for i := 0; i <= 10; i++ {
		y := yMin + float64(i)*yStep
		label := canvas.NewText(fmt.Sprintf("%.1f", y), color.Black)
		label.TextSize = 9
		label.Move(fyne.NewPos(float32(marginLeft)-50, yToPixel(y)-5))
		objects = append(objects, label)
	}

	// Заголовок
	titleLabel := canvas.NewText("Сравнение методов 1-го и 2-го порядка точности", color.Black)
	titleLabel.TextSize = 14
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Move(fyne.NewPos(float32(width/2)-200, 10))
	objects = append(objects, titleLabel)

	// Подписи осей
	yLabel := canvas.NewText("y", color.Black)
	yLabel.TextSize = 12
	yLabel.TextStyle = fyne.TextStyle{Bold: true}
	yLabel.Move(fyne.NewPos(float32(marginLeft)-70, float32(marginTop)-10))
	objects = append(objects, yLabel)

	xLabel := canvas.NewText("x", color.Black)
	xLabel.TextSize = 12
	xLabel.TextStyle = fyne.TextStyle{Bold: true}
	xLabel.Move(fyne.NewPos(float32(width)-50, float32(height)-50))
	objects = append(objects, xLabel)

	// Аналитическое решение (черная толстая линия)
	for i := 0; i < len(results1)-1; i++ {
		line := canvas.NewLine(color.Black)
		line.Position1 = fyne.NewPos(xToPixel(results1[i].x), yToPixel(results1[i].yAnalytic))
		line.Position2 = fyne.NewPos(xToPixel(results1[i+1].x), yToPixel(results1[i+1].yAnalytic))
		line.StrokeWidth = 3
		objects = append(objects, line)
	}

	// Метод 1-го порядка (красная линия)
	redColor := color.RGBA{255, 0, 0, 255}
	for i := 0; i < len(results1)-1; i++ {
		line := canvas.NewLine(redColor)
		line.Position1 = fyne.NewPos(xToPixel(results1[i].x), yToPixel(results1[i].y))
		line.Position2 = fyne.NewPos(xToPixel(results1[i+1].x), yToPixel(results1[i+1].y))
		line.StrokeWidth = 2
		objects = append(objects, line)
	}
	for _, p := range results1 {
		circle := canvas.NewCircle(redColor)
		circle.FillColor = redColor
		circle.Resize(fyne.NewSize(6, 6))
		circle.Move(fyne.NewPos(xToPixel(p.x)-3, yToPixel(p.y)-3))
		objects = append(objects, circle)
	}

	// Метод 2-го порядка (синяя линия)
	blueColor := color.RGBA{0, 0, 255, 255}
	for i := 0; i < len(results2)-1; i++ {
		line := canvas.NewLine(blueColor)
		line.Position1 = fyne.NewPos(xToPixel(results2[i].x), yToPixel(results2[i].y))
		line.Position2 = fyne.NewPos(xToPixel(results2[i+1].x), yToPixel(results2[i+1].y))
		line.StrokeWidth = 2
		objects = append(objects, line)
	}
	for _, p := range results2 {
		circle := canvas.NewCircle(blueColor)
		circle.FillColor = blueColor
		circle.Resize(fyne.NewSize(6, 6))
		circle.Move(fyne.NewPos(xToPixel(p.x)-3, yToPixel(p.y)-3))
		objects = append(objects, circle)
	}

	// Легенда
	legendX := float32(width) - 200
	legendY := float32(marginTop) + 20

	blackLine := canvas.NewLine(color.Black)
	blackLine.Position1 = fyne.NewPos(legendX, legendY+5)
	blackLine.Position2 = fyne.NewPos(legendX+30, legendY+5)
	blackLine.StrokeWidth = 3
	objects = append(objects, blackLine)
	blackLabel := canvas.NewText("Аналитическое", color.Black)
	blackLabel.TextSize = 9
	blackLabel.Move(fyne.NewPos(legendX+35, legendY))
	objects = append(objects, blackLabel)

	redLine := canvas.NewLine(redColor)
	redLine.Position1 = fyne.NewPos(legendX, legendY+25)
	redLine.Position2 = fyne.NewPos(legendX+30, legendY+25)
	redLine.StrokeWidth = 2
	objects = append(objects, redLine)
	redLabel := canvas.NewText("1-й порядок", color.Black)
	redLabel.TextSize = 9
	redLabel.Move(fyne.NewPos(legendX+35, legendY+20))
	objects = append(objects, redLabel)

	blueLine := canvas.NewLine(blueColor)
	blueLine.Position1 = fyne.NewPos(legendX, legendY+45)
	blueLine.Position2 = fyne.NewPos(legendX+30, legendY+45)
	blueLine.StrokeWidth = 2
	objects = append(objects, blueLine)
	blueLabel := canvas.NewText("2-й порядок", color.Black)
	blueLabel.TextSize = 9
	blueLabel.Move(fyne.NewPos(legendX+35, legendY+40))
	objects = append(objects, blueLabel)

	return container.NewWithoutLayout(objects...)
}

// Создание графика погрешностей
func createErrorPlot(results1, results2 []SolutionPoint, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	xMin := results1[0].x
	xMax := results1[len(results1)-1].x
	yMin, yMax := 1e-15, 1.0

	// Находим максимальную ошибку
	for _, p := range results1 {
		if p.error > 0 && p.error < yMin {
			yMin = p.error
		}
		if p.error > yMax {
			yMax = p.error
		}
	}
	for _, p := range results2 {
		if p.error > 0 && p.error < yMin {
			yMin = p.error
		}
		if p.error > yMax {
			yMax = p.error
		}
	}
	yMin = yMin * 0.1
	yMax = yMax * 10

	marginLeft, marginRight := 100.0, 50.0
	marginTop, marginBottom := 60.0, 80.0
	plotWidth := width - marginLeft - marginRight
	plotHeight := height - marginTop - marginBottom

	xToPixel := func(x float64) float32 {
		return float32(marginLeft + (x-xMin)/(xMax-xMin)*plotWidth)
	}
	yToPixel := func(y float64) float32 {
		if y <= 0 {
			return float32(marginTop + plotHeight)
		}
		logY := math.Log10(y)
		logYMin := math.Log10(yMin)
		logYMax := math.Log10(yMax)
		return float32(marginTop + plotHeight*(logYMax-logY)/(logYMax-logYMin))
	}

	// Сетка
	gridColor := color.RGBA{220, 220, 220, 255}
	xStep := (xMax - xMin) / 10
	for i := 0; i <= 10; i++ {
		x := xMin + float64(i)*xStep
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(x), yToPixel(yMax))
		gridLine.Position2 = fyne.NewPos(xToPixel(x), yToPixel(yMin))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}

	logYMin := math.Log10(yMin)
	logYMax := math.Log10(yMax)
	for i := int(math.Floor(logYMin)); i <= int(math.Ceil(logYMax)); i++ {
		y := math.Pow(10, float64(i))
		if y >= yMin && y <= yMax {
			gridLine := canvas.NewLine(gridColor)
			gridLine.Position1 = fyne.NewPos(xToPixel(xMin), yToPixel(y))
			gridLine.Position2 = fyne.NewPos(xToPixel(xMax), yToPixel(y))
			gridLine.StrokeWidth = 0.5
			objects = append(objects, gridLine)

			label := canvas.NewText(fmt.Sprintf("%.0e", y), color.Black)
			label.TextSize = 9
			label.Move(fyne.NewPos(float32(marginLeft)-70, yToPixel(y)-5))
			objects = append(objects, label)
		}
	}

	// Оси
	xAxis := canvas.NewLine(color.Black)
	xAxis.Position1 = fyne.NewPos(xToPixel(xMin), yToPixel(yMin))
	xAxis.Position2 = fyne.NewPos(xToPixel(xMax), yToPixel(yMin))
	xAxis.StrokeWidth = 2
	objects = append(objects, xAxis)

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(xToPixel(xMin), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(xMin), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	// Метки оси X
	for i := 0; i <= 10; i++ {
		x := xMin + float64(i)*xStep
		label := canvas.NewText(fmt.Sprintf("%.2f", x), color.Black)
		label.TextSize = 9
		label.Move(fyne.NewPos(xToPixel(x)-12, yToPixel(yMin)+10))
		objects = append(objects, label)
	}

	// Заголовки
	titleLabel := canvas.NewText("Абсолютная погрешность решения", color.Black)
	titleLabel.TextSize = 14
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Move(fyne.NewPos(float32(width/2)-120, 10))
	objects = append(objects, titleLabel)

	yLabel := canvas.NewText("Погрешность (log)", color.Black)
	yLabel.TextSize = 11
	yLabel.Move(fyne.NewPos(10, float32(height/2)-50))
	objects = append(objects, yLabel)

	xLabel := canvas.NewText("x", color.Black)
	xLabel.TextSize = 11
	xLabel.Move(fyne.NewPos(float32(width/2), float32(height)-30))
	objects = append(objects, xLabel)

	// Цвета для методов
	redColor := color.RGBA{255, 0, 0, 255}
	blueColor := color.RGBA{0, 0, 255, 255}

	// Рисуем линии ошибок для 1-го порядка
	for i := 0; i < len(results1)-1; i++ {
		if results1[i].error > 0 && results1[i+1].error > 0 {
			line := canvas.NewLine(redColor)
			line.Position1 = fyne.NewPos(xToPixel(results1[i].x), yToPixel(results1[i].error))
			line.Position2 = fyne.NewPos(xToPixel(results1[i+1].x), yToPixel(results1[i+1].error))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}
	for _, p := range results1 {
		if p.error > 0 {
			circle := canvas.NewCircle(redColor)
			circle.FillColor = redColor
			circle.Resize(fyne.NewSize(5, 5))
			circle.Move(fyne.NewPos(xToPixel(p.x)-2.5, yToPixel(p.error)-2.5))
			objects = append(objects, circle)
		}
	}

	// Рисуем линии ошибок для 2-го порядка
	for i := 0; i < len(results2)-1; i++ {
		if results2[i].error > 0 && results2[i+1].error > 0 {
			line := canvas.NewLine(blueColor)
			line.Position1 = fyne.NewPos(xToPixel(results2[i].x), yToPixel(results2[i].error))
			line.Position2 = fyne.NewPos(xToPixel(results2[i+1].x), yToPixel(results2[i+1].error))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}
	for _, p := range results2 {
		if p.error > 0 {
			circle := canvas.NewCircle(blueColor)
			circle.FillColor = blueColor
			circle.Resize(fyne.NewSize(5, 5))
			circle.Move(fyne.NewPos(xToPixel(p.x)-2.5, yToPixel(p.error)-2.5))
			objects = append(objects, circle)
		}
	}

	// Легенда
	legendX := float32(width) - 180
	legendY := float32(marginTop) + 20

	redSquare := canvas.NewRectangle(redColor)
	redSquare.Move(fyne.NewPos(legendX, legendY))
	redSquare.Resize(fyne.NewSize(15, 15))
	objects = append(objects, redSquare)
	redLabel := canvas.NewText("1-й порядок", color.Black)
	redLabel.TextSize = 9
	redLabel.Move(fyne.NewPos(legendX+20, legendY))
	objects = append(objects, redLabel)

	blueSquare := canvas.NewRectangle(blueColor)
	blueSquare.Move(fyne.NewPos(legendX, legendY+20))
	blueSquare.Resize(fyne.NewSize(15, 15))
	objects = append(objects, blueSquare)
	blueLabel := canvas.NewText("2-й порядок", color.Black)
	blueLabel.TextSize = 9
	blueLabel.Move(fyne.NewPos(legendX+20, legendY+20))
	objects = append(objects, blueLabel)

	return container.NewWithoutLayout(objects...)
}

// Создание таблицы результатов
func createResultsTable(results1, results2 []SolutionPoint) string {
	var text string
	text += "РЕЗУЛЬТАТЫ ЧИСЛЕННОГО РЕШЕНИЯ КРАЕВОЙ ЗАДАЧИ\n\n"
	text += "Уравнение: xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0\n"
	text += "Граничные условия: y(0.1) + y'(0.1) = 5.399, y(1) = 40.171\n"
	text += "Интервал: [0.1; 1], h = 0.09\n\n"

	text += fmt.Sprintf("%-8s | %-15s | %-15s | %-15s\n",
		"x", "Аналит.", "1-й порядок", "2-й порядок")
	text += strings.Repeat("-", 70) + "\n"

	for i := 0; i < len(results1); i++ {
		text += fmt.Sprintf("%-8.2f | %15.6f | %15.6f | %15.6f\n",
			results1[i].x,
			results1[i].yAnalytic,
			results1[i].y,
			results2[i].y)
	}

	return text
}

// Создание таблицы погрешностей
func createErrorTable(results1, results2 []SolutionPoint) string {
	var text string
	text += "ПОГРЕШНОСТИ ЧИСЛЕННЫХ МЕТОДОВ\n\n"

	text += fmt.Sprintf("%-8s | %-15s | %-15s\n",
		"x", "1-й порядок", "2-й порядок")
	text += strings.Repeat("-", 50) + "\n"

	for i := 0; i < len(results1); i++ {
		text += fmt.Sprintf("%-8.2f | %15.2e | %15.2e\n",
			results1[i].x,
			results1[i].error,
			results2[i].error)
	}

	// Статистика
	text += "\n\nСТАТИСТИКА:\n"
	text += strings.Repeat("-", 50) + "\n"

	// 1-й порядок
	maxError1 := 0.0
	avgError1 := 0.0
	for _, p := range results1 {
		avgError1 += p.error
		if p.error > maxError1 {
			maxError1 = p.error
		}
	}
	avgError1 /= float64(len(results1))
	text += fmt.Sprintf("1-й порядок: Средняя ошибка = %12.2e, Максимальная ошибка = %12.2e\n",
		avgError1, maxError1)

	// 2-й порядок
	maxError2 := 0.0
	avgError2 := 0.0
	for _, p := range results2 {
		avgError2 += p.error
		if p.error > maxError2 {
			maxError2 = p.error
		}
	}
	avgError2 /= float64(len(results2))
	text += fmt.Sprintf("2-й порядок: Средняя ошибка = %12.2e, Максимальная ошибка = %12.2e\n",
		avgError2, maxError2)

	return text
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Решение краевой задачи конечно-разностным методом")
	myWindow.Resize(fyne.NewSize(1200, 900))

	// Параметры задачи
	a := 0.1
	b := 1.0
	h := 0.09
	ya := 5.399  // y(0.1) + y'(0.1) = 5.399
	yb := 40.171 // y(1) = 40.171

	// Строим матрицы систем
	A1, d1, x1 := buildFirstOrderMatrix(a, b, h, ya, yb)
	A2, d2, x2 := buildSecondOrderMatrix(a, b, h, ya, yb)

	// Выводим матрицы в терминал
	fmt.Println()
	printMatrixToTerminal(A1, d1, x1, "МАТРИЦА СИСТЕМЫ 1-ГО ПОРЯДКА ТОЧНОСТИ (метод Гаусса)")
	printMatrixToTerminal(A2, d2, x2, "МАТРИЦА СИСТЕМЫ 2-ГО ПОРЯДКА ТОЧНОСТИ (метод Гаусса)")

	// Вычисляем оба метода
	results1 := SolveFirstOrder(a, b, h, ya, yb)
	results2 := SolveSecondOrder(a, b, h, ya, yb)

	// Выводим результаты в терминал
	fmt.Println("=" + strings.Repeat("=", 100))
	fmt.Println("РЕШЕНИЕ КРАЕВОЙ ЗАДАЧИ: xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0")
	fmt.Println("=" + strings.Repeat("=", 100))
	fmt.Printf("\n%-8s | %-15s | %-15s | %-15s\n",
		"x", "Аналит.", "1-й порядок", "2-й порядок")
	fmt.Println(strings.Repeat("-", 70))

	for i := 0; i < len(results1); i++ {
		fmt.Printf("%-8.2f | %15.6f | %15.6f | %15.6f\n",
			results1[i].x,
			results1[i].yAnalytic,
			results1[i].y,
			results2[i].y)
	}

	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("ПОГРЕШНОСТИ ЧИСЛЕННЫХ МЕТОДОВ (АБСОЛЮТНЫЕ)")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("\n%-8s | %-15s | %-15s\n",
		"x", "1-й порядок", "2-й порядок")
	fmt.Println(strings.Repeat("-", 50))

	for i := 0; i < len(results1); i++ {
		fmt.Printf("%-8.2f | %15.2e | %15.2e\n",
			results1[i].x,
			results1[i].error,
			results2[i].error)
	}

	// Статистика
	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("СТАТИСТИКА ПОГРЕШНОСТЕЙ")
	fmt.Println(strings.Repeat("=", 100))

	maxError1, avgError1 := 0.0, 0.0
	for _, p := range results1 {
		avgError1 += p.error
		if p.error > maxError1 {
			maxError1 = p.error
		}
	}
	avgError1 /= float64(len(results1))
	fmt.Printf("1-й порядок: Средняя = %12.2e, Максимальная = %12.2e\n",
		avgError1, maxError1)

	maxError2, avgError2 := 0.0, 0.0
	for _, p := range results2 {
		avgError2 += p.error
		if p.error > maxError2 {
			maxError2 = p.error
		}
	}
	avgError2 /= float64(len(results2))
	fmt.Printf("2-й порядок: Средняя = %12.2e, Максимальная = %12.2e\n",
		avgError2, maxError2)
	fmt.Println(strings.Repeat("=", 100) + "\n")

	// Создаем вкладки
	var tabs []*container.TabItem

	// Информация
	infoText := `РЕШЕНИЕ КРАЕВОЙ ЗАДАЧИ ДЛЯ ЛИНЕЙНОГО ДИФФЕРЕНЦИАЛЬНОГО УРАВНЕНИЯ ВТОРОГО ПОРЯДКА

Уравнение: xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0

Приведенный вид: y'' + p(x)y' + q(x)y = f(x), где:
  p(x) = -2(3x + 2)/x
  q(x) = 3(3x + 4)/x
  f(x) = 0

Граничные условия:
  y(0.1) + y'(0.1) = 5.399
  y(1) = 40.171

Аналитическое решение: y(x) = (1 + x⁵)e³ˣ

Интервал: [0.1; 1]
Шаг: h = 0.09
Количество точек: 11

МЕТОД 1-ГО ПОРЯДКА ТОЧНОСТИ:

Граничные условия аппроксимируются с 1-м порядком:
  y'(a) ≈ (y₁ - y₀)/h + O(h)

Система уравнений:
  1) (h-1)y₀ + y₁ = y_a·h
  2) (1 - p(xᵢ)h/2)yᵢ₋₁ + (-2 + h²q(xᵢ))yᵢ + (1 + p(xᵢ)h/2)yᵢ₊₁ = h²f(xᵢ)
  3) y_N = y_b

МЕТОД 2-ГО ПОРЯДКА ТОЧНОСТИ:

Граничные условия аппроксимируются с 2-м порядком:
  y'(a) ≈ (-3y₀ + 4y₁ - y₂)/(2h) + O(h²)

Система уравнений:
  1) (2h - 3)y₀ + 4y₁ - y₂ = 2h·y_a
  2) (1 - p(xᵢ)h/2)yᵢ₋₁ + (-2 + h²q(xᵢ))yᵢ + (1 + p(xᵢ)h/2)yᵢ₊₁ = h²f(xᵢ)
  3) y_N = y_b

Метод решения:
  - Для обоих методов используется метод Гаусса

Ожидаемые результаты:
  - Метод 2-го порядка должен давать значительно меньшую погрешность
  - Погрешность 2-го порядка должна убывать как O(h²)
  - Погрешность 1-го порядка должна убывать как O(h)`

	infoContainer := container.NewVScroll(widget.NewLabel(infoText))
	tabs = append(tabs, container.NewTabItem("Информация", infoContainer))

	// Матрицы систем
	matricesText := formatMatrix(A1, d1, x1, "МАТРИЦА СИСТЕМЫ 1-ГО ПОРЯДКА ТОЧНОСТИ")
	matricesText += "\n\nОсобенности:\n"
	matricesText += "- Трехдиагональная матрица (каждая строка содержит максимум 3 ненулевых элемента)\n"
	matricesText += "- Решается методом Гаусса\n"
	matricesText += "- Сложность: O(n³)\n"
	matricesText += "- Граничное условие y'(0.1) аппроксимировано с 1-м порядком: (y₁ - y₀)/h\n\n"
	matricesText += strings.Repeat("=", 120) + "\n\n"

	matricesText += formatMatrix(A2, d2, x2, "МАТРИЦА СИСТЕМЫ 2-ГО ПОРЯДКА ТОЧНОСТИ")
	matricesText += "\n\nОсобенности:\n"
	matricesText += "- Первая строка содержит 3 элемента: A[0][0], A[0][1], A[0][2]\n"
	matricesText += "- Не является строго трехдиагональной из-за элемента A[0][2] = -1\n"
	matricesText += "- Решается методом Гаусса\n"
	matricesText += "- Сложность: O(n³)\n"
	matricesText += "- Граничное условие y'(0.1) аппроксимировано с 2-м порядком: (-3y₀ + 4y₁ - y₂)/(2h)\n"

	matricesContainer := container.NewVScroll(widget.NewLabel(matricesText))
	tabs = append(tabs, container.NewTabItem("Матрицы систем", matricesContainer))

	// Общий график
	combinedPlot := createCombinedPlot(results1, results2, 1100, 600)
	combinedContent := container.NewVBox(
		widget.NewLabel("СРАВНЕНИЕ МЕТОДОВ"),
		widget.NewLabel("Черная линия — аналитическое решение"),
		widget.NewLabel("Красная линия — метод 1-го порядка точности"),
		widget.NewLabel("Синяя линия — метод 2-го порядка точности"),
		widget.NewSeparator(),
		combinedPlot,
	)
	tabs = append(tabs, container.NewTabItem("Графики решений", container.NewVScroll(combinedContent)))

	// График погрешностей
	errorPlot := createErrorPlot(results1, results2, 1100, 600)
	errorPlotContent := container.NewVBox(
		widget.NewLabel("СРАВНЕНИЕ ПОГРЕШНОСТЕЙ"),
		widget.NewLabel("График показывает абсолютную погрешность |y_численное - y_аналитическое|"),
		widget.NewLabel("Ось Y в логарифмической шкале"),
		widget.NewSeparator(),
		errorPlot,
	)
	tabs = append(tabs, container.NewTabItem("Погрешности (график)", container.NewVScroll(errorPlotContent)))

	// Таблица результатов
	resultsText := createResultsTable(results1, results2)
	resultsContainer := container.NewVScroll(widget.NewLabel(resultsText))
	tabs = append(tabs, container.NewTabItem("Результаты (таблица)", resultsContainer))

	// Таблица погрешностей
	errorText := createErrorTable(results1, results2)
	errorContainer := container.NewVScroll(widget.NewLabel(errorText))
	tabs = append(tabs, container.NewTabItem("Погрешности (таблица)", errorContainer))

	tabContainer := container.NewAppTabs(tabs...)

	myWindow.SetContent(tabContainer)
	myWindow.ShowAndRun()
}
