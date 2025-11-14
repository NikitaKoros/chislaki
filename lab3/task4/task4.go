package main

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var xi = []float64{-3.18, -2.72, -2.26, -1.80, -1.34, -0.88, -0.42, 0.04, 0.50, 0.96}
var yi = []float64{7.2581, 6.7618, 5.8412, 4.2463, 1.7815, 0.9568, 0.5649, 3.0194, 4.1121, 4.3743}

const xStar = -0.537

func solveSLAU(A [][]float64, b []float64) []float64 {
	n := len(b)
	augmented := make([][]float64, n)
	for i := range augmented {
		augmented[i] = make([]float64, n+1)
		copy(augmented[i], A[i])
		augmented[i][n] = b[i]
	}

	for i := 0; i < n; i++ {
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(augmented[k][i]) > math.Abs(augmented[maxRow][i]) {
				maxRow = k
			}
		}
		augmented[i], augmented[maxRow] = augmented[maxRow], augmented[i]

		for k := i + 1; k < n; k++ {
			factor := augmented[k][i] / augmented[i][i]
			for j := i; j <= n; j++ {
				augmented[k][j] -= factor * augmented[i][j]
			}
		}
	}

	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		x[i] = augmented[i][n]
		for j := i + 1; j < n; j++ {
			x[i] -= augmented[i][j] * x[j]
		}
		x[i] /= augmented[i][i]
	}

	return x
}

func leastSquares(X, Y []float64, degree int) []float64 {
	n := degree + 1
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
	}
	rightSide := make([]float64, n)

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			power := i + j
			sum := 0.0
			for _, x := range X {
				sum += math.Pow(x, float64(power))
			}
			matrix[i][j] = sum
		}
	}

	for i := 0; i < n; i++ {
		sum := 0.0
		for idx, x := range X {
			sum += Y[idx] * math.Pow(x, float64(i))
		}
		rightSide[i] = sum
	}

	return solveSLAU(matrix, rightSide)
}

func evaluatePolynomial(coeffs []float64, x float64) float64 {
	result := 0.0
	for i, c := range coeffs {
		result += c * math.Pow(x, float64(i))
	}
	return result
}

func calculateSquaredError(X, Y []float64, coeffs []float64) float64 {
	sum := 0.0
	for i := range X {
		predicted := evaluatePolynomial(coeffs, X[i])
		error := predicted - Y[i]
		sum += error * error
	}
	return sum
}

func polynomialToString(coeffs []float64) string {
	result := fmt.Sprintf("%.6f", coeffs[0])
	for i := 1; i < len(coeffs); i++ {
		sign := "+"
		val := coeffs[i]
		if val < 0 {
			sign = "-"
			val = -val
		}
		if i == 1 {
			result += fmt.Sprintf(" %s %.6f·x", sign, val)
		} else {
			result += fmt.Sprintf(" %s %.6f·x^%d", sign, val, i)
		}
	}
	return result
}

func createPlot(coeffs1, coeffs2, coeffs3 []float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(800, 500))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	xMin, xMax := -3.5, 1.5
	yMin, yMax := -1.0, 8.0
	marginLeft, marginRight := 80.0, 40.0
	marginTop, marginBottom := 40.0, 80.0
	plotWidth := 800.0 - marginLeft - marginRight
	plotHeight := 500.0 - marginTop - marginBottom

	scale := math.Min(plotWidth/(xMax-xMin), plotHeight/(yMax-yMin))

	xToPixel := func(x float64) float32 {
		return float32(marginLeft + (x-xMin)*scale)
	}
	yToPixel := func(y float64) float32 {
		return float32(marginTop + (yMax-y)*scale)
	}

	gridColor := color.RGBA{220, 220, 220, 255}
	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(x), yToPixel(yMax))
		gridLine.Position2 = fyne.NewPos(xToPixel(x), yToPixel(yMin))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}
	for y := math.Ceil(yMin); y <= math.Floor(yMax); y++ {
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(xMin), yToPixel(y))
		gridLine.Position2 = fyne.NewPos(xToPixel(xMax), yToPixel(y))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}

	xAxis := canvas.NewLine(color.Black)
	xAxis.Position1 = fyne.NewPos(xToPixel(xMin), yToPixel(0))
	xAxis.Position2 = fyne.NewPos(xToPixel(xMax), yToPixel(0))
	xAxis.StrokeWidth = 2
	objects = append(objects, xAxis)

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
		if x != 0 {
			label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(xToPixel(x)-8, yToPixel(0)+10))
			objects = append(objects, label)
		}
	}
	for y := math.Ceil(yMin); y <= math.Floor(yMax); y++ {
		if y != 0 {
			label := canvas.NewText(fmt.Sprintf("%.0f", y), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(xToPixel(0)-30, yToPixel(y)-5))
			objects = append(objects, label)
		}
	}

	steps := 200
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 150, 0, 255},
		color.RGBA{0, 0, 255, 255},
	}
	allCoeffs := [][]float64{coeffs1, coeffs2, coeffs3}

	for idx, coeffs := range allCoeffs {
		for i := 0; i < steps-1; i++ {
			x1 := xMin + float64(i)*(xMax-xMin)/float64(steps)
			x2 := xMin + float64(i+1)*(xMax-xMin)/float64(steps)
			y1 := evaluatePolynomial(coeffs, x1)
			y2 := evaluatePolynomial(coeffs, x2)

			if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
				line := canvas.NewLine(colors[idx])
				line.Position1 = fyne.NewPos(xToPixel(x1), yToPixel(y1))
				line.Position2 = fyne.NewPos(xToPixel(x2), yToPixel(y2))
				line.StrokeWidth = 2
				objects = append(objects, line)
			}
		}
	}

	for i := 0; i < len(xi); i++ {
		if xi[i] >= xMin && xi[i] <= xMax && yi[i] >= yMin && yi[i] <= yMax {
			circle := canvas.NewCircle(color.Black)
			circle.StrokeColor = color.Black
			circle.StrokeWidth = 2
			circle.FillColor = color.White
			circle.Resize(fyne.NewSize(8, 8))
			circle.Move(fyne.NewPos(xToPixel(xi[i])-4, yToPixel(yi[i])-4))
			objects = append(objects, circle)
		}
	}

	return container.NewWithoutLayout(objects...)
}

func createDataTable() *fyne.Container {
	var rows []fyne.CanvasObject

	iRow := []fyne.CanvasObject{
		widget.NewLabelWithStyle("i: ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for i := 0; i < len(xi); i++ {
		iRow = append(iRow, widget.NewLabel(fmt.Sprintf("%8d", i)))
	}
	rows = append(rows, container.NewHBox(iRow...))

	xiRow := []fyne.CanvasObject{
		widget.NewLabelWithStyle("xi:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for i := 0; i < len(xi); i++ {
		xiRow = append(xiRow, widget.NewLabel(fmt.Sprintf("%8.2f", xi[i])))
	}
	rows = append(rows, container.NewHBox(xiRow...))

	yiRow := []fyne.CanvasObject{
		widget.NewLabelWithStyle("yi:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for i := 0; i < len(yi); i++ {
		yiRow = append(yiRow, widget.NewLabel(fmt.Sprintf("%8.4f", yi[i])))
	}
	rows = append(rows, container.NewHBox(yiRow...))

	return container.NewVBox(rows...)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Метод наименьших квадратов (МНК)")
	myWindow.Resize(fyne.NewSize(900, 800))

	coeffs1 := leastSquares(xi, yi, 1)
	coeffs2 := leastSquares(xi, yi, 2)
	coeffs3 := leastSquares(xi, yi, 3)

	error1 := calculateSquaredError(xi, yi, coeffs1)
	error2 := calculateSquaredError(xi, yi, coeffs2)
	error3 := calculateSquaredError(xi, yi, coeffs3)

	value1 := evaluatePolynomial(coeffs1, xStar)
	value2 := evaluatePolynomial(coeffs2, xStar)
	value3 := evaluatePolynomial(coeffs3, xStar)

	plot := createPlot(coeffs1, coeffs2, coeffs3)

	resultText := fmt.Sprintf("Метод наименьших квадратов (МНК)\nВариант 34\n\n")
	resultText += fmt.Sprintf("Точка вычисления: x* = %.3f\n\n", xStar)
	resultText += "=== МНОГОЧЛЕН 1-Й СТЕПЕНИ ===\n"
	resultText += fmt.Sprintf("F₁(x) = %s\n", polynomialToString(coeffs1))
	resultText += fmt.Sprintf("Сумма квадратов ошибок: Φ₁ = %.6f\n", error1)
	resultText += fmt.Sprintf("F₁(%.3f) = %.6f\n\n", xStar, value1)

	resultText += "=== МНОГОЧЛЕН 2-Й СТЕПЕНИ ===\n"
	resultText += fmt.Sprintf("F₂(x) = %s\n", polynomialToString(coeffs2))
	resultText += fmt.Sprintf("Сумма квадратов ошибок: Φ₂ = %.6f\n", error2)
	resultText += fmt.Sprintf("F₂(%.3f) = %.6f\n\n", xStar, value2)

	resultText += "=== МНОГОЧЛЕН 3-Й СТЕПЕНИ ===\n"
	resultText += fmt.Sprintf("F₃(x) = %s\n", polynomialToString(coeffs3))
	resultText += fmt.Sprintf("Сумма квадратов ошибок: Φ₃ = %.6f\n", error3)
	resultText += fmt.Sprintf("F₃(%.3f) = %.6f\n\n", xStar, value3)

	resultText += "=== СРАВНЕНИЕ РЕЗУЛЬТАТОВ ===\n"
	resultText += fmt.Sprintf("┌──────────┬─────────────────┬────────────────┐\n")
	resultText += fmt.Sprintf("│ Степень  │ Φ (сумма кв.)   │ F(x*)          │\n")
	resultText += fmt.Sprintf("├──────────┼─────────────────┼────────────────┤\n")
	resultText += fmt.Sprintf("│    1     │ %15.8f │ %14.10f │\n", error1, value1)
	resultText += fmt.Sprintf("│    2     │ %15.8f │ %14.10f │\n", error2, value2)
	resultText += fmt.Sprintf("│    3     │ %15.8f │ %14.10f │\n", error3, value3)
	resultText += fmt.Sprintf("└──────────┴─────────────────┴────────────────┘\n\n")

	minError := math.Min(error1, math.Min(error2, error3))
	bestDegree := 1
	if error2 == minError {
		bestDegree = 2
	}
	if error3 == minError {
		bestDegree = 3
	}
	resultText += fmt.Sprintf("Наименьшая сумма квадратов ошибок у многочлена %d-й степени: Φ = %.6f\n", bestDegree, minError)

	coeffsText := "=== КОЭФФИЦИЕНТЫ МНОГОЧЛЕНОВ ===\n\n"
	coeffsText += "Многочлен 1-й степени: F₁(x) = a₀ + a₁·x\n"
	for i, c := range coeffs1 {
		coeffsText += fmt.Sprintf("  a%d = %.10f\n", i, c)
	}
	coeffsText += "\nМногочлен 2-й степени: F₂(x) = a₀ + a₁·x + a₂·x²\n"
	for i, c := range coeffs2 {
		coeffsText += fmt.Sprintf("  a%d = %.10f\n", i, c)
	}
	coeffsText += "\nМногочлен 3-й степени: F₃(x) = a₀ + a₁·x + a₂·x² + a₃·x³\n"
	for i, c := range coeffs3 {
		coeffsText += fmt.Sprintf("  a%d = %.10f\n", i, c)
	}

	checkText := "=== ПРОВЕРКА: ЗНАЧЕНИЯ МНОГОЧЛЕНОВ В УЗЛОВЫХ ТОЧКАХ ===\n\n"
	checkText += fmt.Sprintf("┌────────┬──────────┬──────────┬──────────┬──────────┐\n")
	checkText += fmt.Sprintf("│   x    │    y     │   F₁(x)  │   F₂(x)  │   F₃(x)  │\n")
	checkText += fmt.Sprintf("├────────┼──────────┼──────────┼──────────┼──────────┤\n")
	for i := 0; i < len(xi); i++ {
		f1 := evaluatePolynomial(coeffs1, xi[i])
		f2 := evaluatePolynomial(coeffs2, xi[i])
		f3 := evaluatePolynomial(coeffs3, xi[i])
		checkText += fmt.Sprintf("│ %6.2f │ %8.4f │ %8.4f │ %8.4f │ %8.4f │\n", xi[i], yi[i], f1, f2, f3)
	}
	checkText += fmt.Sprintf("└────────┴──────────┴──────────┴──────────┴──────────┘\n")

	infoText := `Метод наименьших квадратов (МНК):

Метод позволяет найти приближающий многочлен заданной степени,
минимизирующий сумму квадратов отклонений от исходных данных.

Для многочлена степени n:
F(x) = a₀ + a₁·x + a₂·x² + ... + aₙ·xⁿ

Коэффициенты находятся из решения нормальной системы МНК.

Сумма квадратов ошибок:
Φ = Σ(F(xᵢ) - yᵢ)²

Чем меньше Φ, тем лучше приближение.`

	plotContainer := container.NewVBox(
		widget.NewLabel("График приближающих многочленов"),
		widget.NewLabel("Красная линия — многочлен 1-й степени"),
		widget.NewLabel("Зеленая линия — многочлен 2-й степени"),
		widget.NewLabel("Синяя линия — многочлен 3-й степени"),
		widget.NewLabel("Черные точки — данные таблицы"),
		plot,
	)

	resultsContainer := container.NewVBox(
		widget.NewLabel("Результаты аппроксимации"),
		widget.NewSeparator(),
		widget.NewLabel(resultText),
		widget.NewSeparator(),
		widget.NewLabel(infoText),
	)

	coeffsContainer := container.NewVBox(
		widget.NewLabel("Коэффициенты многочленов"),
		widget.NewSeparator(),
		widget.NewLabel(coeffsText),
	)

	checkContainer := container.NewVBox(
		widget.NewLabel("Проверка в узловых точках"),
		widget.NewSeparator(),
		widget.NewLabel(checkText),
	)

	dataTable := createDataTable()
	tableContainer := container.NewVBox(
		widget.NewLabel("Исходные данные"),
		widget.NewSeparator(),
		dataTable,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("График", plotContainer),
		container.NewTabItem("Результаты", container.NewVScroll(resultsContainer)),
		container.NewTabItem("Коэффициенты", container.NewVScroll(coeffsContainer)),
		container.NewTabItem("Проверка", container.NewVScroll(checkContainer)),
		container.NewTabItem("Таблица данных", container.NewVScroll(tableContainer)),
	)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
