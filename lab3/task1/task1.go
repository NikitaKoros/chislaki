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

var xi = []float64{-1.96, -1.43, -0.90, -0.37, 0.16, 0.69, 1.22, 1.75, 2.28}
var yi = []float64{1.3285, 0.4115, 0.9257, 3.1650, 2.9814, 3.7017, 3.3645, 3.0563, 1.4286}

const xStar = -0.718

// Выбор
var quadraticVariant = 2 // 1=LeftInterval, 2=RightInterval
var cubicVariant = 3     // 1=LeftAndTwoRight, 2=Middle, 3=RightAndTwoSides

func lagrangeBasis(i int, x float64, xNodes []float64) float64 {
	result := 1.0
	for j := 0; j < len(xNodes); j++ {
		if j != i {
			result *= (x - xNodes[j]) / (xNodes[i] - xNodes[j])
		}
	}
	return result
}

func lagrangePolynomial(x float64, xNodes, yNodes []float64) float64 {
	result := 0.0
	for i := 0; i < len(xNodes); i++ {
		result += yNodes[i] * lagrangeBasis(i, x, xNodes)
	}
	return result
}

func lagrangeBasisDerivative(i int, x float64, xNodes []float64) float64 {
	result := 0.0
	for k := 0; k < len(xNodes); k++ {
		if k != i {
			prod := 1.0
			for j := 0; j < len(xNodes); j++ {
				if j != i && j != k {
					prod *= (x - xNodes[j]) / (xNodes[i] - xNodes[j])
				}
			}
			result += prod / (xNodes[i] - xNodes[k])
		}
	}
	return result
}

func lagrangeDerivative(x float64, xNodes, yNodes []float64) float64 {
	result := 0.0
	for i := 0; i < len(xNodes); i++ {
		result += yNodes[i] * lagrangeBasisDerivative(i, x, xNodes)
	}
	return result
}

func buildPolynomial2LeftInterval(xStar float64) ([]float64, []float64, float64) {
	leftIdx := 0
	for i := 0; i < len(xi)-1; i++ {
		if xi[i] <= xStar && xi[i+1] >= xStar {
			leftIdx = i
			break
		}
	}

	xNodes := []float64{xi[leftIdx-1], xi[leftIdx], xi[leftIdx+1]}
	yNodes := []float64{yi[leftIdx-1], yi[leftIdx], yi[leftIdx+1]}
	value := lagrangePolynomial(xStar, xNodes, yNodes)
	return xNodes, yNodes, value
}

func buildPolynomial2RightInterval(xStar float64) ([]float64, []float64, float64) {
	leftIdx := 0
	for i := 0; i < len(xi)-1; i++ {
		if xi[i] <= xStar && xi[i+1] >= xStar {
			leftIdx = i
			break
		}
	}

	xNodes := []float64{xi[leftIdx], xi[leftIdx+1], xi[leftIdx+2]}
	yNodes := []float64{yi[leftIdx], yi[leftIdx+1], yi[leftIdx+2]}
	value := lagrangePolynomial(xStar, xNodes, yNodes)
	return xNodes, yNodes, value
}

func buildPolynomial3LeftAndTwoRight(xStar float64) ([]float64, []float64, float64) {
	leftIdx := 0
	for i := 0; i < len(xi)-1; i++ {
		if xi[i] <= xStar && xi[i+1] >= xStar {
			leftIdx = i
			break
		}
	}

	xNodes := []float64{xi[leftIdx], xi[leftIdx+1], xi[leftIdx+2], xi[leftIdx+3]}
	yNodes := []float64{yi[leftIdx], yi[leftIdx+1], yi[leftIdx+2], yi[leftIdx+3]}
	value := lagrangePolynomial(xStar, xNodes, yNodes)
	return xNodes, yNodes, value
}

func buildPolynomial3Middle(xStar float64) ([]float64, []float64, float64) {
	leftIdx := 0
	for i := 0; i < len(xi)-1; i++ {
		if xi[i] <= xStar && xi[i+1] >= xStar {
			leftIdx = i
			break
		}
	}

	xNodes := []float64{xi[leftIdx-1], xi[leftIdx], xi[leftIdx+1], xi[leftIdx+2]}
	yNodes := []float64{yi[leftIdx-1], yi[leftIdx], yi[leftIdx+1], yi[leftIdx+2]}
	value := lagrangePolynomial(xStar, xNodes, yNodes)
	return xNodes, yNodes, value
}

func buildPolynomial3RightAndTwoSides(xStar float64) ([]float64, []float64, float64) {
	leftIdx := 0
	for i := 0; i < len(xi)-1; i++ {
		if xi[i] <= xStar && xi[i+1] >= xStar {
			leftIdx = i
			break
		}
	}

	xNodes := []float64{xi[leftIdx-2], xi[leftIdx-1], xi[leftIdx], xi[leftIdx+1]}
	yNodes := []float64{yi[leftIdx-2], yi[leftIdx-1], yi[leftIdx], yi[leftIdx+1]}
	value := lagrangePolynomial(xStar, xNodes, yNodes)
	return xNodes, yNodes, value
}

func buildPolynomial(degree int, xStar float64) ([]float64, []float64, float64) {
	if degree == 2 {
		if quadraticVariant == 2 {
			return buildPolynomial2RightInterval(xStar)
		}
		return buildPolynomial2LeftInterval(xStar)
	}
	switch cubicVariant {
	case 2:
		return buildPolynomial3Middle(xStar)
	case 3:
		return buildPolynomial3RightAndTwoSides(xStar)
	default:
		return buildPolynomial3LeftAndTwoRight(xStar)
	}
}

func createPlot() *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(800, 500))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	xMin, xMax := -3.0, 3.0
	yMin, yMax := 0.0, 4.0
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

	xNodes2, yNodes2, _ := buildPolynomial(2, xStar)

	steps := 500
	xMin2, xMax2 := xNodes2[0], xNodes2[len(xNodes2)-1]
	for i := 0; i < steps-1; i++ {
		x1 := xMin2 + float64(i)*(xMax2-xMin2)/float64(steps)
		x2 := xMin2 + float64(i+1)*(xMax2-xMin2)/float64(steps)
		y1 := lagrangePolynomial(x1, xNodes2, yNodes2)
		y2 := lagrangePolynomial(x2, xNodes2, yNodes2)

		line := canvas.NewLine(color.RGBA{0, 0, 255, 255})
		line.Position1 = fyne.NewPos(xToPixel(x1), yToPixel(y1))
		line.Position2 = fyne.NewPos(xToPixel(x2), yToPixel(y2))
		line.StrokeWidth = 2
		objects = append(objects, line)
	}

	xNodes3, yNodes3, _ := buildPolynomial(3, xStar)

	fmt.Println("Quadratic nodes:", xNodes2)
	fmt.Println("Cubic nodes:", xNodes3)

	xMin3, xMax3 := xNodes3[0], xNodes3[len(xNodes3)-1]
	for i := 0; i < steps-1; i++ {
		x1 := xMin3 + float64(i)*(xMax3-xMin3)/float64(steps)
		x2 := xMin3 + float64(i+1)*(xMax3-xMin3)/float64(steps)
		y1 := lagrangePolynomial(x1, xNodes3, yNodes3)
		y2 := lagrangePolynomial(x2, xNodes3, yNodes3)

		line := canvas.NewLine(color.RGBA{255, 0, 0, 255})
		line.Position1 = fyne.NewPos(xToPixel(x1), yToPixel(y1))
		line.Position2 = fyne.NewPos(xToPixel(x2), yToPixel(y2))
		line.StrokeWidth = 2
		objects = append(objects, line)
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

	_, _, y2 := buildPolynomial(2, xStar)
	_, _, y3 := buildPolynomial(3, xStar)
	yAvg := (y2 + y3) / 2

	if xStar >= xMin && xStar <= xMax && yAvg >= yMin && yAvg <= yMax {
		vLine := canvas.NewLine(color.RGBA{0, 200, 0, 255})
		vLine.Position1 = fyne.NewPos(xToPixel(xStar), yToPixel(yAvg)-8)
		vLine.Position2 = fyne.NewPos(xToPixel(xStar), yToPixel(yAvg)+8)
		vLine.StrokeWidth = 3
		objects = append(objects, vLine)

		hLine := canvas.NewLine(color.RGBA{0, 200, 0, 255})
		hLine.Position1 = fyne.NewPos(xToPixel(xStar)-8, yToPixel(yAvg))
		hLine.Position2 = fyne.NewPos(xToPixel(xStar)+8, yToPixel(yAvg))
		hLine.StrokeWidth = 3
		objects = append(objects, hLine)
	}

	return container.NewWithoutLayout(objects...)
}

func createDataTable() *fyne.Container {
	var rows []fyne.CanvasObject

	iRow := []fyne.CanvasObject{
		widget.NewLabelWithStyle("i:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for i := 0; i < len(xi); i++ {
		iRow = append(iRow, widget.NewLabel(fmt.Sprintf("%d", i)))
	}
	rows = append(rows, container.NewHBox(iRow...))

	xiRow := []fyne.CanvasObject{
		widget.NewLabelWithStyle("xi:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for i := 0; i < len(xi); i++ {
		xiRow = append(xiRow, widget.NewLabel(fmt.Sprintf("%.2f", xi[i])))
	}
	rows = append(rows, container.NewHBox(xiRow...))

	yiRow := []fyne.CanvasObject{
		widget.NewLabelWithStyle("yi:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for i := 0; i < len(yi); i++ {
		yiRow = append(yiRow, widget.NewLabel(fmt.Sprintf("%.4f", yi[i])))
	}
	rows = append(rows, container.NewHBox(yiRow...))

	return container.NewVBox(rows...)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Интерполяция Лагранжа")
	myWindow.Resize(fyne.NewSize(900, 800))

	plot := createPlot()

	xNodes2, yNodes2, value2 := buildPolynomial(2, xStar)
	xNodes3, yNodes3, value3 := buildPolynomial(3, xStar)

	deriv2 := lagrangeDerivative(xStar, xNodes2, yNodes2)
	deriv3 := lagrangeDerivative(xStar, xNodes3, yNodes3)

	boundaryX := xNodes2[len(xNodes2)-1]
	deriv2AtBoundary := lagrangeDerivative(boundaryX, xNodes2, yNodes2)
	deriv3AtBoundary := lagrangeDerivative(boundaryX, xNodes3, yNodes3)

	checkX := xi[2]
	xNodes2Check, yNodes2Check, _ := buildPolynomial(2, checkX)
	checkValue := lagrangePolynomial(checkX, xNodes2Check, yNodes2Check)
	originalY := yi[2]

	var quad2Name, cubic3Name string
	if quadraticVariant == 1 {
		quad2Name = "LeftInterval"
	} else {
		quad2Name = "RightInterval"
	}
	switch cubicVariant {
	case 1:
		cubic3Name = "LeftAndTwoRight"
	case 2:
		cubic3Name = "Middle"
	case 3:
		cubic3Name = "RightAndTwoSides"
	}

	result2Text := fmt.Sprintf("Многочлен 2-й степени (вариант: %s):\nУзлы: ", quad2Name)
	for i, x := range xNodes2 {
		result2Text += fmt.Sprintf("x%d=%.2f ", i, x)
	}
	result2Text += fmt.Sprintf("\nL₂(%.3f) = %.6f\nL₂'(%.3f) = %.6f", xStar, value2, xStar, deriv2)

	result3Text := fmt.Sprintf("Многочлен 3-й степени (вариант: %s):\nУзлы: ", cubic3Name)
	for i, x := range xNodes3 {
		result3Text += fmt.Sprintf("x%d=%.2f ", i, x)
	}
	result3Text += fmt.Sprintf("\nL₃(%.3f) = %.6f\nL₃'(%.3f) = %.6f", xStar, value3, xStar, deriv3)

	checkText := fmt.Sprintf("Проверка в узловой точке x=%.2f:\nL₂(%.2f) = %.6f\nИсходное значение: %.4f\nПогрешность: %.2e",
		checkX, checkX, checkValue, originalY, math.Abs(checkValue-originalY))

	continuityText := fmt.Sprintf("Проверка непрерывности производной:\nТочка: x=%.2f\nL₂'(%.2f) = %.6f\nL₃'(%.2f) = %.6f\nРазность: %.2e",
		boundaryX, boundaryX, deriv2AtBoundary, boundaryX, deriv3AtBoundary, math.Abs(deriv2AtBoundary-deriv3AtBoundary))

	variantsText := `Варианты выбора узлов (измените переменные в начале файла):

Квадратичный (quadraticVariant):
1 - LeftInterval: x* в левом интервале [i-1, i, i+1]
2 - RightInterval: x* в правом интервале [i, i+1, i+2]

Кубический (cubicVariant):
1 - LeftAndTwoRight: x* слева, 2 справа [i, i+1, i+2, i+3]
2 - Middle: x* в среднем [i-1, i, i+1, i+2]
3 - RightAndTwoSides: x* справа, 2 по бокам [i-1, i, i+1, i+2]`

	plotContainer := container.NewVBox(
		widget.NewLabel("График интерполяционных многочленов"),
		widget.NewLabel("Синий — L₂(x) (2-я степень), Красный — L₃(x) (3-я степень)"),
		widget.NewLabel("Черные точки — данные таблицы, Зеленый крест — точка интерполяции x*"),
		plot,
	)

	resultsContainer := container.NewVBox(
		widget.NewLabel("Результаты интерполяции"),
		widget.NewSeparator(),
		widget.NewLabel(result2Text),
		widget.NewSeparator(),
		widget.NewLabel(result3Text),
		widget.NewSeparator(),
		widget.NewLabel(checkText),
		widget.NewSeparator(),
		widget.NewLabel(continuityText),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Точка интерполяции: x* = %.3f", xStar)),
		widget.NewSeparator(),
		widget.NewLabel(variantsText),
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
		container.NewTabItem("Таблица данных", container.NewVScroll(tableContainer)),
	)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
