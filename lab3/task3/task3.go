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

var xi = []float64{-1.98, -1.507, -0.991, -0.647, -0.303, 0.002, 0.514, 0.815, 1.103, 1.632, 2.32}
var yi = []float64{0.109, 0.437, 1.435, 1.913, 1.718, 1.168, 1.097, 0.249, 0.023, 0.212, 0.841}

const xStar = -0.519

type SplineCoeffs struct {
	a, b, c, d float64
}

func solveTridiagonal(a, b, c, d []float64) []float64 {
	n := len(d)
	cp := make([]float64, n)
	dp := make([]float64, n)
	x := make([]float64, n)

	cp[0] = c[0] / b[0]
	dp[0] = d[0] / b[0]

	for i := 1; i < n; i++ {
		denom := b[i] - a[i]*cp[i-1]
		cp[i] = c[i] / denom
		dp[i] = (d[i] - a[i]*dp[i-1]) / denom
	}

	x[n-1] = dp[n-1]
	for i := n - 2; i >= 0; i-- {
		x[i] = dp[i] - cp[i]*x[i+1]
	}

	return x
}

func buildNaturalCubicSpline(X, Y []float64) []SplineCoeffs {
	n := len(X) - 1
	h := make([]float64, n)

	for i := 0; i < n; i++ {
		h[i] = X[i+1] - X[i]
	}

	alpha := make([]float64, n)
	for i := 1; i < n; i++ {
		alpha[i] = (3/h[i])*(Y[i+1]-Y[i]) - (3/h[i-1])*(Y[i]-Y[i-1])
	}

	size := n - 1
	a := make([]float64, size)
	b := make([]float64, size)
	c := make([]float64, size)
	d := make([]float64, size)

	b[0] = 2 * (h[0] + h[1])
	c[0] = h[1]
	d[0] = alpha[1]

	for i := 1; i < size-1; i++ {
		a[i] = h[i]
		b[i] = 2 * (h[i] + h[i+1])
		c[i] = h[i+1]
		d[i] = alpha[i+1]
	}

	if size > 1 {
		a[size-1] = h[n-2]
		b[size-1] = 2 * (h[n-2] + h[n-1])
		d[size-1] = alpha[n-1]
	}

	cInner := solveTridiagonal(a, b, c, d)

	cCoeffs := make([]float64, n+1)
	cCoeffs[0] = 0
	for i := 0; i < len(cInner); i++ {
		cCoeffs[i+1] = cInner[i]
	}
	cCoeffs[n] = 0

	splines := make([]SplineCoeffs, n)
	for i := 0; i < n; i++ {
		ai := Y[i]
		bi := (Y[i+1]-Y[i])/h[i] - h[i]*(2*cCoeffs[i]+cCoeffs[i+1])/3
		ci := cCoeffs[i]
		di := (cCoeffs[i+1] - cCoeffs[i]) / (3 * h[i])

		splines[i] = SplineCoeffs{a: ai, b: bi, c: ci, d: di}
	}

	return splines
}

func evaluateSpline(xNodes []float64, coeffs []SplineCoeffs, xVal float64) float64 {
	idx := 0
	for i := 0; i < len(xNodes)-1; i++ {
		if xVal >= xNodes[i] && xVal <= xNodes[i+1] {
			idx = i
			break
		}
	}

	dx := xVal - xNodes[idx]
	return coeffs[idx].a + coeffs[idx].b*dx + coeffs[idx].c*dx*dx + coeffs[idx].d*dx*dx*dx
}

func findInterval(x []float64, xVal float64) int {
	for i := 0; i < len(x)-1; i++ {
		if xVal >= x[i] && xVal <= x[i+1] {
			return i
		}
	}
	return 0
}

func createPlot(coeffs []SplineCoeffs) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(800, 500))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	xMin, xMax := -2.5, 2.5
	yMin, yMax := -0.5, 2.5
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

	arrowX1 := canvas.NewLine(color.Black)
	arrowX1.Position1 = fyne.NewPos(xToPixel(xMax), yToPixel(0))
	arrowX1.Position2 = fyne.NewPos(xToPixel(xMax)-8, yToPixel(0)+4)
	arrowX1.StrokeWidth = 2
	objects = append(objects, arrowX1)
	arrowX2 := canvas.NewLine(color.Black)
	arrowX2.Position1 = fyne.NewPos(xToPixel(xMax), yToPixel(0))
	arrowX2.Position2 = fyne.NewPos(xToPixel(xMax)-8, yToPixel(0)-4)
	arrowX2.StrokeWidth = 2
	objects = append(objects, arrowX2)

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	arrowY1 := canvas.NewLine(color.Black)
	arrowY1.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	arrowY1.Position2 = fyne.NewPos(xToPixel(0)-4, yToPixel(yMax)+8)
	arrowY1.StrokeWidth = 2
	objects = append(objects, arrowY1)
	arrowY2 := canvas.NewLine(color.Black)
	arrowY2.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	arrowY2.Position2 = fyne.NewPos(xToPixel(0)+4, yToPixel(yMax)+8)
	arrowY2.StrokeWidth = 2
	objects = append(objects, arrowY2)

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

	xLabel := canvas.NewText("x", color.Black)
	xLabel.TextSize = 14
	xLabel.Move(fyne.NewPos(xToPixel(xMax)+10, yToPixel(0)-10))
	objects = append(objects, xLabel)

	yLabel := canvas.NewText("y", color.Black)
	yLabel.TextSize = 14
	yLabel.Move(fyne.NewPos(xToPixel(0)+10, yToPixel(yMax)-10))
	objects = append(objects, yLabel)

	zeroLabel := canvas.NewText("0", color.Black)
	zeroLabel.TextSize = 10
	zeroLabel.Move(fyne.NewPos(xToPixel(0)-15, yToPixel(0)+10))
	objects = append(objects, zeroLabel)

	steps := 100
	for seg := 0; seg < len(xi)-1; seg++ {
		for i := 0; i < steps-1; i++ {
			x1 := xi[seg] + float64(i)*(xi[seg+1]-xi[seg])/float64(steps)
			x2 := xi[seg] + float64(i+1)*(xi[seg+1]-xi[seg])/float64(steps)
			y1 := evaluateSpline(xi, coeffs, x1)
			y2 := evaluateSpline(xi, coeffs, x2)

			line := canvas.NewLine(color.RGBA{0, 0, 255, 255})
			line.Position1 = fyne.NewPos(xToPixel(x1), yToPixel(y1))
			line.Position2 = fyne.NewPos(xToPixel(x2), yToPixel(y2))
			line.StrokeWidth = 2
			objects = append(objects, line)
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

	yVal := evaluateSpline(xi, coeffs, xStar)
	if xStar >= xMin && xStar <= xMax && yVal >= yMin && yVal <= yMax {
		vLine := canvas.NewLine(color.RGBA{255, 0, 0, 255})
		vLine.Position1 = fyne.NewPos(xToPixel(xStar), yToPixel(yVal)-8)
		vLine.Position2 = fyne.NewPos(xToPixel(xStar), yToPixel(yVal)+8)
		vLine.StrokeWidth = 3
		objects = append(objects, vLine)

		hLine := canvas.NewLine(color.RGBA{255, 0, 0, 255})
		hLine.Position1 = fyne.NewPos(xToPixel(xStar)-8, yToPixel(yVal))
		hLine.Position2 = fyne.NewPos(xToPixel(xStar)+8, yToPixel(yVal))
		hLine.StrokeWidth = 3
		objects = append(objects, hLine)
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
		xiRow = append(xiRow, widget.NewLabel(fmt.Sprintf("%8.3f", xi[i])))
	}
	rows = append(rows, container.NewHBox(xiRow...))

	yiRow := []fyne.CanvasObject{
		widget.NewLabelWithStyle("yi:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for i := 0; i < len(yi); i++ {
		yiRow = append(yiRow, widget.NewLabel(fmt.Sprintf("%8.3f", yi[i])))
	}
	rows = append(rows, container.NewHBox(yiRow...))

	return container.NewVBox(rows...)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Естественный кубический сплайн")
	myWindow.Resize(fyne.NewSize(900, 800))

	coeffs := buildNaturalCubicSpline(xi, yi)
	plot := createPlot(coeffs)

	yVal := evaluateSpline(xi, coeffs, xStar)
	intervalIdx := findInterval(xi, xStar)

	resultText := fmt.Sprintf("Естественный кубический сплайн дефекта 1\n\n")
	resultText += fmt.Sprintf("Точка интерполяции: x* = %.3f\n", xStar)
	resultText += fmt.Sprintf("Значение сплайна: S(%.3f) = %.6f\n\n", xStar, yVal)
	resultText += fmt.Sprintf("Точка x* находится в интервале [x%d, x%d] = [%.3f, %.3f]\n\n",
		intervalIdx, intervalIdx+1, xi[intervalIdx], xi[intervalIdx+1])

	coeffsText := "Коэффициенты сплайна на всех отрезках:\n\n"
	for i := 0; i < len(coeffs); i++ {
		coeffsText += fmt.Sprintf("Отрезок [x%d, x%d] = [%.3f, %.3f]:\n", i, i+1, xi[i], xi[i+1])
		coeffsText += fmt.Sprintf("  S%d(x) = %.6f + %.6f(x-%.3f) + %.6f(x-%.3f)² + %.6f(x-%.3f)³\n",
			i, coeffs[i].a, coeffs[i].b, xi[i], coeffs[i].c, xi[i], coeffs[i].d, xi[i])
		coeffsText += fmt.Sprintf("  a%d = %.6f\n", i, coeffs[i].a)
		coeffsText += fmt.Sprintf("  b%d = %.6f\n", i, coeffs[i].b)
		coeffsText += fmt.Sprintf("  c%d = %.6f\n", i, coeffs[i].c)
		coeffsText += fmt.Sprintf("  d%d = %.6f\n\n", i, coeffs[i].d)
	}

	intervalCoeffsText := fmt.Sprintf("Коэффициенты сплайна на отрезке, содержащем x* = %.3f:\n\n", xStar)
	intervalCoeffsText += fmt.Sprintf("Отрезок [x%d, x%d] = [%.3f, %.3f]:\n\n", intervalIdx, intervalIdx+1, xi[intervalIdx], xi[intervalIdx+1])
	intervalCoeffsText += fmt.Sprintf("S%d(x) = %.6f + %.6f(x-%.3f) + %.6f(x-%.3f)² + %.6f(x-%.3f)³\n\n",
		intervalIdx, coeffs[intervalIdx].a, coeffs[intervalIdx].b, xi[intervalIdx],
		coeffs[intervalIdx].c, xi[intervalIdx], coeffs[intervalIdx].d, xi[intervalIdx])
	intervalCoeffsText += fmt.Sprintf("a%d = %.6f\n", intervalIdx, coeffs[intervalIdx].a)
	intervalCoeffsText += fmt.Sprintf("b%d = %.6f\n", intervalIdx, coeffs[intervalIdx].b)
	intervalCoeffsText += fmt.Sprintf("c%d = %.6f\n", intervalIdx, coeffs[intervalIdx].c)
	intervalCoeffsText += fmt.Sprintf("d%d = %.6f\n\n", intervalIdx, coeffs[intervalIdx].d)

	dx := xStar - xi[intervalIdx]
	intervalCoeffsText += fmt.Sprintf("Вычисление S%d(%.3f):\n", intervalIdx, xStar)
	intervalCoeffsText += fmt.Sprintf("Δx = %.3f - %.3f = %.6f\n", xStar, xi[intervalIdx], dx)
	intervalCoeffsText += fmt.Sprintf("S%d(%.3f) = %.6f + %.6f·%.6f + %.6f·%.6f² + %.6f·%.6f³\n",
		intervalIdx, xStar, coeffs[intervalIdx].a, coeffs[intervalIdx].b, dx,
		coeffs[intervalIdx].c, dx, coeffs[intervalIdx].d, dx)
	intervalCoeffsText += fmt.Sprintf("         = %.6f\n", yVal)

	checkText := "Проверка в узловых точках:\n\n"
	for i := 0; i < len(xi); i++ {
		sVal := evaluateSpline(xi, coeffs, xi[i])
		checkText += fmt.Sprintf("x%d = %.3f: S(x) = %.6f, y%d = %.3f, |S-y| = %.3e\n",
			i, xi[i], sVal, i, yi[i], math.Abs(sVal-yi[i]))
	}

	infoText := `Естественный кубический сплайн дефекта 1:

Свойства:
• Сплайн проходит через все узловые точки
• Непрерывны функция и её первые две производные
• На концах отрезка вторая производная равна нулю (естественные граничные условия)
• Дефект 1 означает, что третья производная может иметь разрывы в узлах

Формула сплайна на отрезке [xi, xi+1]:
Si(x) = ai + bi(x-xi) + ci(x-xi)² + di(x-xi)³

где ai, bi, ci, di — коэффициенты сплайна на i-м отрезке`

	plotContainer := container.NewVBox(
		widget.NewLabel("График кубического сплайна"),
		widget.NewLabel("Синяя линия — кубический сплайн"),
		widget.NewLabel("Черные точки — данные таблицы, Красный крест — точка интерполяции x*"),
		plot,
	)

	resultsContainer := container.NewVBox(
		widget.NewLabel("Результаты интерполяции"),
		widget.NewSeparator(),
		widget.NewLabel(resultText),
		widget.NewSeparator(),
		widget.NewLabel(intervalCoeffsText),
		widget.NewSeparator(),
		widget.NewLabel(checkText),
		widget.NewSeparator(),
		widget.NewLabel(infoText),
	)

	coeffsContainer := container.NewVBox(
		widget.NewLabel("Коэффициенты сплайна"),
		widget.NewSeparator(),
		widget.NewLabel(coeffsText),
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
		container.NewTabItem("Таблица данных", container.NewVScroll(tableContainer)),
	)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
