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

// Данные таблицы
var xi = []float64{-1.96, -1.43, -0.90, -0.37, 0.16, 0.69, 1.22, 1.75, 2.28}
var yi = []float64{1.3285, 0.4115, 0.9257, 3.1650, 2.9814, 3.7017, 3.3645, 3.0563, 1.4286}

const xStar = -0.718

// Базисный полином Лагранжа li(x)
func lagrangeBasis(i int, x float64, xNodes []float64) float64 {
	result := 1.0
	for j := 0; j < len(xNodes); j++ {
		if j != i {
			result *= (x - xNodes[j]) / (xNodes[i] - xNodes[j])
		}
	}
	return result
}

// Интерполяционный многочлен Лагранжа
func lagrangePolynomial(x float64, xNodes, yNodes []float64) float64 {
	result := 0.0
	for i := 0; i < len(xNodes); i++ {
		result += yNodes[i] * lagrangeBasis(i, x, xNodes)
	}
	return result
}

// Построение многочлена степени n по узлам около точки x*
func buildPolynomial(degree int, xStar float64) ([]float64, []float64, float64) {
	// Находим ближайшие узлы к x*
	n := degree + 1
	distances := make([]struct {
		idx  int
		dist float64
	}, len(xi))

	for i := 0; i < len(xi); i++ {
		distances[i].idx = i
		distances[i].dist = math.Abs(xi[i] - xStar)
	}

	// Сортируем по расстоянию
	for i := 0; i < len(distances)-1; i++ {
		for j := i + 1; j < len(distances); j++ {
			if distances[j].dist < distances[i].dist {
				distances[i], distances[j] = distances[j], distances[i]
			}
		}
	}

	// Выбираем n ближайших узлов
	xNodes := make([]float64, n)
	yNodes := make([]float64, n)
	for i := 0; i < n; i++ {
		idx := distances[i].idx
		xNodes[i] = xi[idx]
		yNodes[i] = yi[idx]
	}

	// Сортируем узлы по x
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if xNodes[j] < xNodes[i] {
				xNodes[i], xNodes[j] = xNodes[j], xNodes[i]
				yNodes[i], yNodes[j] = yNodes[j], yNodes[i]
			}
		}
	}

	// Вычисляем значение в точке x*
	value := lagrangePolynomial(xStar, xNodes, yNodes)

	return xNodes, yNodes, value
}

// Создание графика
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

	// Сетка
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

	// Оси
	xAxis := canvas.NewLine(color.Black)
	xAxis.Position1 = fyne.NewPos(xToPixel(xMin), yToPixel(0))
	xAxis.Position2 = fyne.NewPos(xToPixel(xMax), yToPixel(0))
	xAxis.StrokeWidth = 2
	objects = append(objects, xAxis)

	// Стрелка оси X
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

	// Стрелка оси Y
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

	// Метки осей
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

	// Подписи осей
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

	// Многочлен 2-й степени (синий)
	xNodes2, yNodes2, _ := buildPolynomial(2, xStar)
	//xNodes2, yNodes2 := xi, yi

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

	// Многочлен 3-й степени (красный)
	xNodes3, yNodes3, _ := buildPolynomial(3, xStar)
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

	// Точки данных (черные кружки)
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

	// Точка интерполяции x* (зеленый крест)
	_, _, y2 := buildPolynomial(2, xStar)
	_, _, y3 := buildPolynomial(3, xStar)
	yAvg := (y2 + y3) / 2

	if xStar >= xMin && xStar <= xMax && yAvg >= yMin && yAvg <= yMax {
		// Вертикальная линия креста
		vLine := canvas.NewLine(color.RGBA{0, 200, 0, 255})
		vLine.Position1 = fyne.NewPos(xToPixel(xStar), yToPixel(yAvg)-8)
		vLine.Position2 = fyne.NewPos(xToPixel(xStar), yToPixel(yAvg)+8)
		vLine.StrokeWidth = 3
		objects = append(objects, vLine)

		// Горизонтальная линия креста
		hLine := canvas.NewLine(color.RGBA{0, 200, 0, 255})
		hLine.Position1 = fyne.NewPos(xToPixel(xStar)-8, yToPixel(yAvg))
		hLine.Position2 = fyne.NewPos(xToPixel(xStar)+8, yToPixel(yAvg))
		hLine.StrokeWidth = 3
		objects = append(objects, hLine)
	}

	return container.NewWithoutLayout(objects...)
}

// Создание таблицы данных
func createDataTable() *fyne.Container {
	var rows []fyne.CanvasObject

	// Строка i
	iRow := []fyne.CanvasObject{
		widget.NewLabelWithStyle("i:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for i := 0; i < len(xi); i++ {
		iRow = append(iRow, widget.NewLabel(fmt.Sprintf("%d", i)))
	}
	rows = append(rows, container.NewHBox(iRow...))

	// Строка xi
	xiRow := []fyne.CanvasObject{
		widget.NewLabelWithStyle("xi:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	for i := 0; i < len(xi); i++ {
		xiRow = append(xiRow, widget.NewLabel(fmt.Sprintf("%.2f", xi[i])))
	}
	rows = append(rows, container.NewHBox(xiRow...))

	// Строка yi
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

	// Вычисляем результаты
	xNodes2, _, value2 := buildPolynomial(2, xStar)
	xNodes3, _, value3 := buildPolynomial(3, xStar)

	// Проверка в узловой точке (берем первый узел из многочлена 2-й степени)
	xNodes2Check, yNodes2Check, _ := buildPolynomial(2, xStar)
	checkX := xNodes2Check[0]
	checkValue := lagrangePolynomial(checkX, xNodes2Check, yNodes2Check)

	// Находим исходное значение
	var originalY float64
	for i := 0; i < len(xi); i++ {
		if math.Abs(xi[i]-checkX) < 1e-10 {
			originalY = yi[i]
			break
		}
	}

	// Результаты
	result2Text := fmt.Sprintf("Многочлен 2-й степени:\nУзлы: ")
	for i, x := range xNodes2 {
		result2Text += fmt.Sprintf("x%d=%.2f ", i, x)
	}
	result2Text += fmt.Sprintf("\nL₂(%.3f) = %.6f", xStar, value2)

	result3Text := fmt.Sprintf("Многочлен 3-й степени:\nУзлы: ")
	for i, x := range xNodes3 {
		result3Text += fmt.Sprintf("x%d=%.2f ", i, x)
	}
	result3Text += fmt.Sprintf("\nL₃(%.3f) = %.6f", xStar, value3)

	checkText := fmt.Sprintf("Проверка в узловой точке x=%.2f:\nL₂(%.2f) = %.6f\nИсходное значение: %.4f\nПогрешность: %.2e",
		checkX, checkX, checkValue, originalY, math.Abs(checkValue-originalY))

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
		widget.NewLabel(fmt.Sprintf("Точка интерполяции: x* = %.3f", xStar)),
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
