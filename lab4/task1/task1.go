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

// ButcherTableau представляет таблицу Бутчера для метода Рунге-Кутты
type ButcherTableau struct {
	c []float64   // узлы (левая колонка)
	a [][]float64 // матрица коэффициентов
	b []float64   // весовые коэффициенты (нижняя строка)
	s int         // количество стадий
}

// Исходное ОДУ: x²y' + y² - xy + x² = 0
// Преобразуем к виду y' = f(x, y):
// x²y' = -y² + xy - x²
// y' = (-y² + xy - x²) / x² = -y²/x² + y/x - 1
func f(x, y float64) float64 {
	if math.Abs(x) < 1e-10 {
		return 0
	}
	return -y*y/(x*x) + y/x - 1
}

// Аналитическое решение: y = x * tg(1 - ln x)
func analyticalSolution(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return x * math.Tan(1-math.Log(x))
}

// Результат численного решения в точке
type SolutionPoint struct {
	x         float64
	y         float64
	yAnalytic float64
	error     float64
}

// SolveExplicitRK решает ОДУ явным методом Рунге-Кутты
func SolveExplicitRK(bt ButcherTableau, x0, y0, xEnd, h float64) []SolutionPoint {
	var results []SolutionPoint
	x, y := x0, y0

	// Добавляем начальную точку
	results = append(results, SolutionPoint{
		x:         x,
		y:         y,
		yAnalytic: analyticalSolution(x),
		error:     math.Abs(y - analyticalSolution(x)),
	})

	for x < xEnd-1e-10 {
		// Вычисляем коэффициенты K_i
		k := make([]float64, bt.s)
		for i := 0; i < bt.s; i++ {
			xi := x + bt.c[i]*h
			yi := y
			for j := 0; j < i; j++ {
				yi += bt.a[i][j] * k[j]
			}
			k[i] = h * f(xi, yi)
		}

		// Вычисляем приращение
		dy := 0.0
		for i := 0; i < bt.s; i++ {
			dy += bt.b[i] * k[i]
		}

		// Обновляем значения
		x += h
		y += dy

		// Сохраняем результат
		yAnalytic := analyticalSolution(x)
		results = append(results, SolutionPoint{
			x:         x,
			y:         y,
			yAnalytic: yAnalytic,
			error:     math.Abs(y - yAnalytic),
		})
	}

	return results
}

// SolveImplicitRK решает ОДУ неявным методом Рунге-Кутты
// используя метод простых итераций для решения системы нелинейных уравнений
func SolveImplicitRK(bt ButcherTableau, x0, y0, xEnd, h float64) []SolutionPoint {
	var results []SolutionPoint
	x, y := x0, y0

	// Добавляем начальную точку
	results = append(results, SolutionPoint{
		x:         x,
		y:         y,
		yAnalytic: analyticalSolution(x),
		error:     math.Abs(y - analyticalSolution(x)),
	})

	maxIter := 100
	tol := 1e-10

	for x < xEnd-1e-10 {
		// Инициализация k
		k := make([]float64, bt.s)
		for i := 0; i < bt.s; i++ {
			k[i] = h * f(x+bt.c[i]*h, y)
		}

		// Метод простых итераций для решения неявной системы
		for iter := 0; iter < maxIter; iter++ {
			kNew := make([]float64, bt.s)
			converged := true

			for i := 0; i < bt.s; i++ {
				xi := x + bt.c[i]*h
				yi := y
				for j := 0; j < bt.s; j++ {
					yi += bt.a[i][j] * k[j]
				}
				kNew[i] = h * f(xi, yi)

				if math.Abs(kNew[i]-k[i]) > tol {
					converged = false
				}
			}

			k = kNew
			if converged {
				break
			}
		}

		// Вычисляем приращение
		dy := 0.0
		for i := 0; i < bt.s; i++ {
			dy += bt.b[i] * k[i]
		}

		// Обновляем значения
		x += h
		y += dy

		// Сохраняем результат
		yAnalytic := analyticalSolution(x)
		results = append(results, SolutionPoint{
			x:         x,
			y:         y,
			yAnalytic: yAnalytic,
			error:     math.Abs(y - yAnalytic),
		})
	}

	return results
}

// Создание таблиц Бутчера для различных методов

// Метод Эйлера (1 порядок)
func NewEulerTableau() ButcherTableau {
	return ButcherTableau{
		s: 1,
		c: []float64{0},
		a: [][]float64{{0}},
		b: []float64{1},
	}
}

// Метод Хойна (2 порядок)
func NewHeunTableau() ButcherTableau {
	return ButcherTableau{
		s: 2,
		c: []float64{0, 0.5},
		a: [][]float64{
			{0, 0},
			{0.5, 0},
		},
		b: []float64{0, 1},
	}
}

// Метод 3 порядка (4 способ)
func NewRK3AltTableau() ButcherTableau {
	return ButcherTableau{
		s: 3,
		c: []float64{0, 0.5, 0.75},
		a: [][]float64{
			{0, 0, 0},
			{0.5, 0, 0},
			{0, 0.75, 0},
		},
		b: []float64{2.0 / 9.0, 1.0 / 3.0, 4.0 / 9.0},
	}
}

// Классический метод Рунге-Кутты (4 порядок)
func NewRK4Tableau() ButcherTableau {
	return ButcherTableau{
		s: 4,
		c: []float64{0, 0.5, 0.5, 1},
		a: [][]float64{
			{0, 0, 0, 0},
			{0.5, 0, 0, 0},
			{0, 0.5, 0, 0},
			{0, 0, 1, 0},
		},
		b: []float64{1.0 / 6.0, 1.0 / 3.0, 1.0 / 3.0, 1.0 / 6.0},
	}
}

// Неявные

// Метод Гаусса 2 порядка (Правило средней точки)
func NewGauss2Tableau() ButcherTableau {
	return ButcherTableau{
		s: 1,
		c: []float64{0.5},
		a: [][]float64{{0.5}},
		b: []float64{1},
	}
}

// Метод Радо IIA 3 порядка
func NewRadauIIA3Tableau() ButcherTableau {
	return ButcherTableau{
		s: 2,
		c: []float64{1.0 / 3.0, 1},
		a: [][]float64{
			{5.0 / 12.0, -1.0 / 12.0},
			{3.0 / 4.0, 1.0 / 4.0},
		},
		b: []float64{3.0 / 4.0, 1.0 / 4.0},
	}
}

// Структура для хранения результатов всех методов
type AllResults struct {
	euler     []SolutionPoint
	heun      []SolutionPoint
	rk3Alt    []SolutionPoint
	rk4       []SolutionPoint
	gauss2    []SolutionPoint
	radauIIA3 []SolutionPoint
}

// Вычисление всех методов
func computeAllMethods(x0, y0, xEnd, h float64) AllResults {
	return AllResults{
		euler:     SolveExplicitRK(NewEulerTableau(), x0, y0, xEnd, h),
		heun:      SolveExplicitRK(NewHeunTableau(), x0, y0, xEnd, h),
		rk3Alt:    SolveExplicitRK(NewRK3AltTableau(), x0, y0, xEnd, h),
		rk4:       SolveExplicitRK(NewRK4Tableau(), x0, y0, xEnd, h),
		gauss2:    SolveImplicitRK(NewGauss2Tableau(), x0, y0, xEnd, h),
		radauIIA3: SolveImplicitRK(NewRadauIIA3Tableau(), x0, y0, xEnd, h),
	}
}

// Создание общего графика со всеми методами
func createCombinedPlot(results AllResults, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	// Определяем границы по данным
	xMin, xMax := 1.0, 8.0
	yMin, yMax := -10.0, 10.0

	// Автоматическое определение yMin и yMax
	allPoints := [][]SolutionPoint{
		results.euler, results.heun, results.rk3Alt,
		results.rk4, results.gauss2, results.radauIIA3,
	}
	for _, points := range allPoints {
		for _, p := range points {
			if !math.IsNaN(p.yAnalytic) && !math.IsInf(p.yAnalytic, 0) {
				if p.yAnalytic < yMin {
					yMin = p.yAnalytic
				}
				if p.yAnalytic > yMax {
					yMax = p.yAnalytic
				}
			}
		}
	}
	yMin = math.Max(yMin*1.2, -15)
	yMax = math.Min(yMax*1.2, 15)

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
	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(x), yToPixel(yMax))
		gridLine.Position2 = fyne.NewPos(xToPixel(x), yToPixel(yMin))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}
	for y := math.Ceil(yMin); y <= math.Floor(yMax); y += 5 {
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
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	// Метки осей
	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
		label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
		label.TextSize = 10
		label.Move(fyne.NewPos(xToPixel(x)-8, yToPixel(0)+10))
		objects = append(objects, label)
	}
	for y := math.Ceil(yMin); y <= math.Floor(yMax); y += 5 {
		if y != 0 {
			label := canvas.NewText(fmt.Sprintf("%.0f", y), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(xToPixel(0)-35, yToPixel(y)-5))
			objects = append(objects, label)
		}
	}

	// Заголовок
	titleLabel := canvas.NewText("Сравнение всех методов", color.Black)
	titleLabel.TextSize = 14
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Move(fyne.NewPos(float32(width/2)-100, 10))
	objects = append(objects, titleLabel)

	// Подписи осей
	yLabel := canvas.NewText("y", color.Black)
	yLabel.TextSize = 12
	yLabel.TextStyle = fyne.TextStyle{Bold: true}
	yLabel.Move(fyne.NewPos(float32(marginLeft)-60, float32(marginTop)-10))
	objects = append(objects, yLabel)

	xLabel := canvas.NewText("x", color.Black)
	xLabel.TextSize = 12
	xLabel.TextStyle = fyne.TextStyle{Bold: true}
	xLabel.Move(fyne.NewPos(float32(width)-50, float32(height)-50))
	objects = append(objects, xLabel)

	// Дополнительная ось Y слева с делениями
	leftYAxis := canvas.NewLine(color.Black)
	leftYAxis.Position1 = fyne.NewPos(float32(marginLeft)-15, yToPixel(yMin))
	leftYAxis.Position2 = fyne.NewPos(float32(marginLeft)-15, yToPixel(yMax))
	leftYAxis.StrokeWidth = 2
	objects = append(objects, leftYAxis)

	// Деления и метки на левой оси Y
	yStep := 1.0
	if yMax-yMin > 20 {
		yStep = 5.0
	} else if yMax-yMin > 10 {
		yStep = 2.0
	}

	for y := math.Ceil(yMin/yStep) * yStep; y <= yMax; y += yStep {
		if y >= yMin && y <= yMax {
			// Деление
			tick := canvas.NewLine(color.Black)
			tick.Position1 = fyne.NewPos(float32(marginLeft)-20, yToPixel(y))
			tick.Position2 = fyne.NewPos(float32(marginLeft)-10, yToPixel(y))
			tick.StrokeWidth = 2
			objects = append(objects, tick)

			// Метка
			label := canvas.NewText(fmt.Sprintf("%.0f", y), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(float32(marginLeft)-55, yToPixel(y)-5))
			objects = append(objects, label)
		}
	}

	// Единичный отрезок (от 0 до 1 и от 0 до -1)
	if 0 >= yMin && 0 <= yMax {
		// Подпись 0
		zeroLabel := canvas.NewText("0", color.Black)
		zeroLabel.TextSize = 10
		zeroLabel.TextStyle = fyne.TextStyle{Bold: true}
		zeroLabel.Move(fyne.NewPos(float32(marginLeft)-55, yToPixel(0)-5))
		objects = append(objects, zeroLabel)

		// Единичный отрезок от 0 до 1
		if 1 >= yMin && 1 <= yMax {
			unitSegment1 := canvas.NewLine(color.RGBA{255, 0, 0, 255})
			unitSegment1.Position1 = fyne.NewPos(float32(marginLeft)-15, yToPixel(0))
			unitSegment1.Position2 = fyne.NewPos(float32(marginLeft)-15, yToPixel(1))
			unitSegment1.StrokeWidth = 3
			objects = append(objects, unitSegment1)

			oneLabel := canvas.NewText("1", color.RGBA{255, 0, 0, 255})
			oneLabel.TextSize = 10
			oneLabel.TextStyle = fyne.TextStyle{Bold: true}
			oneLabel.Move(fyne.NewPos(float32(marginLeft)-55, yToPixel(1)-5))
			objects = append(objects, oneLabel)
		}

		// Единичный отрезок от 0 до -1
		if -1 >= yMin && -1 <= yMax {
			unitSegment2 := canvas.NewLine(color.RGBA{255, 0, 0, 255})
			unitSegment2.Position1 = fyne.NewPos(float32(marginLeft)-15, yToPixel(0))
			unitSegment2.Position2 = fyne.NewPos(float32(marginLeft)-15, yToPixel(-1))
			unitSegment2.StrokeWidth = 3
			objects = append(objects, unitSegment2)

			minusOneLabel := canvas.NewText("-1", color.RGBA{255, 0, 0, 255})
			minusOneLabel.TextSize = 10
			minusOneLabel.TextStyle = fyne.TextStyle{Bold: true}
			minusOneLabel.Move(fyne.NewPos(float32(marginLeft)-55, yToPixel(-1)-5))
			objects = append(objects, minusOneLabel)
		}
	}

	// Аналитическое решение (черная толстая линия)
	for i := 0; i < len(results.euler)-1; i++ {
		y1 := results.euler[i].yAnalytic
		y2 := results.euler[i+1].yAnalytic
		if !math.IsNaN(y1) && !math.IsInf(y1, 0) && !math.IsNaN(y2) && !math.IsInf(y2, 0) {
			if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
				line := canvas.NewLine(color.Black)
				line.Position1 = fyne.NewPos(xToPixel(results.euler[i].x), yToPixel(y1))
				line.Position2 = fyne.NewPos(xToPixel(results.euler[i+1].x), yToPixel(y2))
				line.StrokeWidth = 3
				objects = append(objects, line)
			}
		}
	}

	// Цвета для методов
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Эйлер - красный
		color.RGBA{0, 150, 0, 255},   // Хойн - зеленый
		color.RGBA{255, 165, 0, 255}, // RK3Alt - оранжевый
		color.RGBA{0, 0, 255, 255},   // RK4 - синий
		color.RGBA{128, 0, 128, 255}, // Gauss2 - фиолетовый
		color.RGBA{0, 200, 200, 255}, // RadauIIA3 - голубой
	}

	methodNames := []string{"Эйлер", "Хойн", "RK3 (4 способ)", "RK4", "Гаусс-2", "Радо IIA-3"}

	// Рисуем линии всех методов
	for idx, points := range allPoints {
		for i := 0; i < len(points)-1; i++ {
			y1 := points[i].y
			y2 := points[i+1].y
			if !math.IsNaN(y1) && !math.IsInf(y1, 0) && !math.IsNaN(y2) && !math.IsInf(y2, 0) {
				if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
					line := canvas.NewLine(colors[idx])
					line.Position1 = fyne.NewPos(xToPixel(points[i].x), yToPixel(y1))
					line.Position2 = fyne.NewPos(xToPixel(points[i+1].x), yToPixel(y2))
					line.StrokeWidth = 2
					objects = append(objects, line)
				}
			}
		}

		// Точки
		for _, p := range points {
			if !math.IsNaN(p.y) && !math.IsInf(p.y, 0) && p.y >= yMin && p.y <= yMax {
				circle := canvas.NewCircle(colors[idx])
				circle.FillColor = colors[idx]
				circle.Resize(fyne.NewSize(5, 5))
				circle.Move(fyne.NewPos(xToPixel(p.x)-2.5, yToPixel(p.y)-2.5))
				objects = append(objects, circle)
			}
		}
	}

	// Легенда
	legendX := float32(width) - 180
	legendY := float32(marginTop) + 20
	legendTitle := canvas.NewText("Методы:", color.Black)
	legendTitle.TextSize = 10
	legendTitle.TextStyle = fyne.TextStyle{Bold: true}
	legendTitle.Move(fyne.NewPos(legendX, legendY))
	objects = append(objects, legendTitle)

	// Аналитическое решение в легенде
	blackLine := canvas.NewLine(color.Black)
	blackLine.Position1 = fyne.NewPos(legendX, legendY+25)
	blackLine.Position2 = fyne.NewPos(legendX+30, legendY+25)
	blackLine.StrokeWidth = 3
	objects = append(objects, blackLine)

	blackLabel := canvas.NewText("Аналитическое", color.Black)
	blackLabel.TextSize = 9
	blackLabel.Move(fyne.NewPos(legendX+35, legendY+20))
	objects = append(objects, blackLabel)

	// Методы в легенде
	for i, name := range methodNames {
		yPos := legendY + 45 + float32(i)*20

		methodLine := canvas.NewLine(colors[i])
		methodLine.Position1 = fyne.NewPos(legendX, yPos+5)
		methodLine.Position2 = fyne.NewPos(legendX+30, yPos+5)
		methodLine.StrokeWidth = 2
		objects = append(objects, methodLine)

		text := canvas.NewText(name, color.Black)
		text.TextSize = 9
		text.Move(fyne.NewPos(legendX+35, yPos))
		objects = append(objects, text)
	}

	return container.NewWithoutLayout(objects...)
}

// Создание графика с одним методом
func createMethodPlot(points []SolutionPoint, methodName string, methodColor color.Color, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	xMin, xMax := 1.0, 8.0
	yMin, yMax := -10.0, 10.0

	// Автоматическое определение yMin и yMax
	for _, p := range points {
		if !math.IsNaN(p.yAnalytic) && !math.IsInf(p.yAnalytic, 0) {
			if p.yAnalytic < yMin {
				yMin = p.yAnalytic
			}
			if p.yAnalytic > yMax {
				yMax = p.yAnalytic
			}
		}
		if !math.IsNaN(p.y) && !math.IsInf(p.y, 0) {
			if p.y < yMin {
				yMin = p.y
			}
			if p.y > yMax {
				yMax = p.y
			}
		}
	}
	yMin = math.Max(yMin*1.2, -15)
	yMax = math.Min(yMax*1.2, 15)

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
	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(x), yToPixel(yMax))
		gridLine.Position2 = fyne.NewPos(xToPixel(x), yToPixel(yMin))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}
	for y := math.Ceil(yMin); y <= math.Floor(yMax); y += 5 {
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
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	// Метки осей
	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
		label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
		label.TextSize = 10
		label.Move(fyne.NewPos(xToPixel(x)-8, yToPixel(0)+10))
		objects = append(objects, label)
	}
	for y := math.Ceil(yMin); y <= math.Floor(yMax); y += 5 {
		if y != 0 {
			label := canvas.NewText(fmt.Sprintf("%.0f", y), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(xToPixel(0)-35, yToPixel(y)-5))
			objects = append(objects, label)
		}
	}

	// Заголовок
	titleLabel := canvas.NewText(methodName, color.Black)
	titleLabel.TextSize = 12
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Move(fyne.NewPos(float32(width/2)-100, 10))
	objects = append(objects, titleLabel)

	// Подписи осей
	yAxisLabel := canvas.NewText("y", color.Black)
	yAxisLabel.TextSize = 12
	yAxisLabel.TextStyle = fyne.TextStyle{Bold: true}
	yAxisLabel.Move(fyne.NewPos(float32(marginLeft)-60, float32(marginTop)-10))
	objects = append(objects, yAxisLabel)

	xAxisLabel := canvas.NewText("x", color.Black)
	xAxisLabel.TextSize = 12
	xAxisLabel.TextStyle = fyne.TextStyle{Bold: true}
	xAxisLabel.Move(fyne.NewPos(float32(width)-50, float32(height)-50))
	objects = append(objects, xAxisLabel)

	// Дополнительная ось Y слева с делениями
	leftYAxis := canvas.NewLine(color.Black)
	leftYAxis.Position1 = fyne.NewPos(float32(marginLeft)-15, yToPixel(yMin))
	leftYAxis.Position2 = fyne.NewPos(float32(marginLeft)-15, yToPixel(yMax))
	leftYAxis.StrokeWidth = 2
	objects = append(objects, leftYAxis)

	// Деления и метки на левой оси Y
	yStep := 1.0
	if yMax-yMin > 20 {
		yStep = 5.0
	} else if yMax-yMin > 10 {
		yStep = 2.0
	}

	for y := math.Ceil(yMin/yStep) * yStep; y <= yMax; y += yStep {
		if y >= yMin && y <= yMax {
			// Деление
			tick := canvas.NewLine(color.Black)
			tick.Position1 = fyne.NewPos(float32(marginLeft)-20, yToPixel(y))
			tick.Position2 = fyne.NewPos(float32(marginLeft)-10, yToPixel(y))
			tick.StrokeWidth = 2
			objects = append(objects, tick)

			// Метка
			label := canvas.NewText(fmt.Sprintf("%.0f", y), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(float32(marginLeft)-55, yToPixel(y)-5))
			objects = append(objects, label)
		}
	}

	// Единичный отрезок (от 0 до 1 и от 0 до -1)
	if 0 >= yMin && 0 <= yMax {
		// Подпись 0
		zeroLabel := canvas.NewText("0", color.Black)
		zeroLabel.TextSize = 10
		zeroLabel.TextStyle = fyne.TextStyle{Bold: true}
		zeroLabel.Move(fyne.NewPos(float32(marginLeft)-55, yToPixel(0)-5))
		objects = append(objects, zeroLabel)

		// Единичный отрезок от 0 до 1
		if 1 >= yMin && 1 <= yMax {
			unitSegment1 := canvas.NewLine(color.RGBA{255, 0, 0, 255})
			unitSegment1.Position1 = fyne.NewPos(float32(marginLeft)-15, yToPixel(0))
			unitSegment1.Position2 = fyne.NewPos(float32(marginLeft)-15, yToPixel(1))
			unitSegment1.StrokeWidth = 3
			objects = append(objects, unitSegment1)

			oneLabel := canvas.NewText("1", color.RGBA{255, 0, 0, 255})
			oneLabel.TextSize = 10
			oneLabel.TextStyle = fyne.TextStyle{Bold: true}
			oneLabel.Move(fyne.NewPos(float32(marginLeft)-55, yToPixel(1)-5))
			objects = append(objects, oneLabel)
		}

		// Единичный отрезок от 0 до -1
		if -1 >= yMin && -1 <= yMax {
			unitSegment2 := canvas.NewLine(color.RGBA{255, 0, 0, 255})
			unitSegment2.Position1 = fyne.NewPos(float32(marginLeft)-15, yToPixel(0))
			unitSegment2.Position2 = fyne.NewPos(float32(marginLeft)-15, yToPixel(-1))
			unitSegment2.StrokeWidth = 3
			objects = append(objects, unitSegment2)

			minusOneLabel := canvas.NewText("-1", color.RGBA{255, 0, 0, 255})
			minusOneLabel.TextSize = 10
			minusOneLabel.TextStyle = fyne.TextStyle{Bold: true}
			minusOneLabel.Move(fyne.NewPos(float32(marginLeft)-55, yToPixel(-1)-5))
			objects = append(objects, minusOneLabel)
		}
	}

	// Аналитическое решение (черная линия)
	for i := 0; i < len(points)-1; i++ {
		y1 := points[i].yAnalytic
		y2 := points[i+1].yAnalytic
		if !math.IsNaN(y1) && !math.IsInf(y1, 0) && !math.IsNaN(y2) && !math.IsInf(y2, 0) {
			if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
				line := canvas.NewLine(color.Black)
				line.Position1 = fyne.NewPos(xToPixel(points[i].x), yToPixel(y1))
				line.Position2 = fyne.NewPos(xToPixel(points[i+1].x), yToPixel(y2))
				line.StrokeWidth = 2
				objects = append(objects, line)
			}
		}
	}

	// Численное решение (цветная линия с точками)
	for i := 0; i < len(points)-1; i++ {
		y1 := points[i].y
		y2 := points[i+1].y
		if !math.IsNaN(y1) && !math.IsInf(y1, 0) && !math.IsNaN(y2) && !math.IsInf(y2, 0) {
			if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
				line := canvas.NewLine(methodColor)
				line.Position1 = fyne.NewPos(xToPixel(points[i].x), yToPixel(y1))
				line.Position2 = fyne.NewPos(xToPixel(points[i+1].x), yToPixel(y2))
				line.StrokeWidth = 2
				objects = append(objects, line)
			}
		}
	}

	// Точки
	for _, p := range points {
		if !math.IsNaN(p.y) && !math.IsInf(p.y, 0) && p.y >= yMin && p.y <= yMax {
			circle := canvas.NewCircle(methodColor)
			circle.FillColor = methodColor
			circle.Resize(fyne.NewSize(6, 6))
			circle.Move(fyne.NewPos(xToPixel(p.x)-3, yToPixel(p.y)-3))
			objects = append(objects, circle)
		}
	}

	// Легенда
	legendX := float32(width) - 180
	legendY := float32(marginTop) + 20

	blackLine := canvas.NewLine(color.Black)
	blackLine.Position1 = fyne.NewPos(legendX, legendY+5)
	blackLine.Position2 = fyne.NewPos(legendX+30, legendY+5)
	blackLine.StrokeWidth = 2
	objects = append(objects, blackLine)

	blackLabel := canvas.NewText("Аналитическое", color.Black)
	blackLabel.TextSize = 9
	blackLabel.Move(fyne.NewPos(legendX+35, legendY))
	objects = append(objects, blackLabel)

	colorLine := canvas.NewLine(methodColor)
	colorLine.Position1 = fyne.NewPos(legendX, legendY+25)
	colorLine.Position2 = fyne.NewPos(legendX+30, legendY+25)
	colorLine.StrokeWidth = 2
	objects = append(objects, colorLine)

	colorLabel := canvas.NewText("Численное", color.Black)
	colorLabel.TextSize = 9
	colorLabel.Move(fyne.NewPos(legendX+35, legendY+20))
	objects = append(objects, colorLabel)

	return container.NewWithoutLayout(objects...)
}

// Создание графика погрешностей
func createErrorPlot(results AllResults, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	xMin, xMax := 1.0, 8.0
	yMin, yMax := 1e-15, 1.0

	// Находим максимальную ошибку
	allPoints := [][]SolutionPoint{
		results.euler, results.heun, results.rk3Alt,
		results.rk4, results.gauss2, results.radauIIA3,
	}
	for _, points := range allPoints {
		for _, p := range points {
			if p.error > 0 && p.error < yMin {
				yMin = p.error
			}
			if p.error > yMax {
				yMax = p.error
			}
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
	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
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
	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
		label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
		label.TextSize = 10
		label.Move(fyne.NewPos(xToPixel(x)-8, yToPixel(yMin)+10))
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
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Эйлер - красный
		color.RGBA{0, 150, 0, 255},   // Хойн - зеленый
		color.RGBA{255, 165, 0, 255}, // RK3Alt - оранжевый
		color.RGBA{0, 0, 255, 255},   // RK4 - синий
		color.RGBA{128, 0, 128, 255}, // Gauss2 - фиолетовый
		color.RGBA{0, 200, 200, 255}, // RadauIIA3 - голубой
	}

	methodNames := []string{"Эйлер", "Хойн", "RK3 (4 способ)", "RK4", "Гаусс-2", "Радо IIA-3"}

	// Рисуем линии ошибок
	for idx, points := range allPoints {
		for i := 0; i < len(points)-1; i++ {
			if points[i].error > 0 && points[i+1].error > 0 {
				line := canvas.NewLine(colors[idx])
				line.Position1 = fyne.NewPos(xToPixel(points[i].x), yToPixel(points[i].error))
				line.Position2 = fyne.NewPos(xToPixel(points[i+1].x), yToPixel(points[i+1].error))
				line.StrokeWidth = 2
				objects = append(objects, line)
			}
		}

		// Точки
		for _, p := range points {
			if p.error > 0 {
				circle := canvas.NewCircle(colors[idx])
				circle.FillColor = colors[idx]
				circle.Resize(fyne.NewSize(5, 5))
				circle.Move(fyne.NewPos(xToPixel(p.x)-2.5, yToPixel(p.error)-2.5))
				objects = append(objects, circle)
			}
		}
	}

	// Легенда справа по центру
	legendX := float32(width) - 180
	legendY := float32(marginTop) + float32(plotHeight)/2 - 60
	legendTitle := canvas.NewText("Методы:", color.Black)
	legendTitle.TextSize = 10
	legendTitle.TextStyle = fyne.TextStyle{Bold: true}
	legendTitle.Move(fyne.NewPos(legendX, legendY))
	objects = append(objects, legendTitle)

	for i, name := range methodNames {
		yPos := legendY + 20 + float32(i)*20

		square := canvas.NewRectangle(colors[i])
		square.Move(fyne.NewPos(legendX, yPos))
		square.Resize(fyne.NewSize(15, 15))
		objects = append(objects, square)

		text := canvas.NewText(name, color.Black)
		text.TextSize = 9
		text.Move(fyne.NewPos(legendX+20, yPos))
		objects = append(objects, text)
	}

	return container.NewWithoutLayout(objects...)
}

// Создание таблицы результатов
func createResultsTable(results AllResults) string {
	var text string
	text += "РЕЗУЛЬТАТЫ ЧИСЛЕННОГО РЕШЕНИЯ ОДУ\n\n"
	text += "Уравнение: x²y' + y² - xy + x² = 0\n"
	text += "Начальное условие: y(1) = tg(1)\n"
	text += "Интервал: [1; 8], h = 0.25\n\n"

	text += fmt.Sprintf("%-8s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "Аналит.", "Эйлер", "Хойн", "RK3 Alt", "RK4", "Гаусс-2", "Радо IIA-3")
	text += "--------------------------------------------------------------------------------------------------------------------\n"

	for i := 0; i < len(results.euler); i++ {
		text += fmt.Sprintf("%-8.2f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f\n",
			results.euler[i].x,
			results.euler[i].yAnalytic,
			results.euler[i].y,
			results.heun[i].y,
			results.rk3Alt[i].y,
			results.rk4[i].y,
			results.gauss2[i].y,
			results.radauIIA3[i].y)
	}

	return text
}

// Создание таблицы погрешностей
func createErrorTable(results AllResults) string {
	var text string
	text += "ПОГРЕШНОСТИ ЧИСЛЕННЫХ МЕТОДОВ\n\n"

	text += fmt.Sprintf("%-8s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "Эйлер", "Хойн", "RK3 Alt", "RK4", "Гаусс-2", "Радо IIA-3")
	text += "--------------------------------------------------------------------------------------------------------------------\n"

	for i := 0; i < len(results.euler); i++ {
		text += fmt.Sprintf("%-8.2f | %12.2e | %12.2e | %12.2e | %12.2e | %12.2e | %12.2e\n",
			results.euler[i].x,
			results.euler[i].error,
			results.heun[i].error,
			results.rk3Alt[i].error,
			results.rk4[i].error,
			results.gauss2[i].error,
			results.radauIIA3[i].error)
	}

	// Статистика
	text += "\n\nСТАТИСТИКА:\n"
	text += "--------------------------------------------------------------------------------------------------------------------\n"

	allMethods := []struct {
		name   string
		points []SolutionPoint
	}{
		{"Эйлер", results.euler},
		{"Хойн", results.heun},
		{"RK3 (4 способ)", results.rk3Alt},
		{"RK4", results.rk4},
		{"Гаусс-2", results.gauss2},
		{"Радо IIA-3", results.radauIIA3},
	}

	for _, method := range allMethods {
		maxError := 0.0
		avgError := 0.0
		for _, p := range method.points {
			avgError += p.error
			if p.error > maxError {
				maxError = p.error
			}
		}
		avgError /= float64(len(method.points))
		text += fmt.Sprintf("%-15s: Средняя ошибка = %12.2e, Максимальная ошибка = %12.2e\n",
			method.name, avgError, maxError)
	}

	return text
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Решение ОДУ методами Рунге-Кутты")
	myWindow.Resize(fyne.NewSize(1200, 900))

	// Параметры задачи
	x0 := 1.0
	y0 := math.Tan(1.0)
	xEnd := 8.0
	h := 0.25

	// Вычисляем все методы
	results := computeAllMethods(x0, y0, xEnd, h)

	// Выводим результаты в терминал для анализа
	fmt.Println("=" + strings.Repeat("=", 140))
	fmt.Println("РЕЗУЛЬТАТЫ ЧИСЛЕННОГО РЕШЕНИЯ ОДУ: x²y' + y² - xy + x² = 0")
	fmt.Println("=" + strings.Repeat("=", 140))
	fmt.Printf("\n%-8s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "Аналит.", "Эйлер", "Хойн", "RK3 Alt", "RK4", "Гаусс-2", "Радо IIA-3")
	fmt.Println(strings.Repeat("-", 141))

	for i := 0; i < len(results.euler); i++ {
		fmt.Printf("%-8.2f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f\n",
			results.euler[i].x,
			results.euler[i].yAnalytic,
			results.euler[i].y,
			results.heun[i].y,
			results.rk3Alt[i].y,
			results.rk4[i].y,
			results.gauss2[i].y,
			results.radauIIA3[i].y)
	}

	fmt.Println("\n" + strings.Repeat("=", 141))
	fmt.Println("ПОГРЕШНОСТИ ЧИСЛЕННЫХ МЕТОДОВ (АБСОЛЮТНЫЕ)")
	fmt.Println(strings.Repeat("=", 141))
	fmt.Printf("\n%-8s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "Эйлер", "Хойн", "RK3 Alt", "RK4", "Гаусс-2", "Радо IIA-3")
	fmt.Println(strings.Repeat("-", 141))

	for i := 0; i < len(results.euler); i++ {
		fmt.Printf("%-8.2f | %12.2e | %12.2e | %12.2e | %12.2e | %12.2e | %12.2e\n",
			results.euler[i].x,
			results.euler[i].error,
			results.heun[i].error,
			results.rk3Alt[i].error,
			results.rk4[i].error,
			results.gauss2[i].error,
			results.radauIIA3[i].error)
	}

	// Статистика
	fmt.Println("\n" + strings.Repeat("=", 141))
	fmt.Println("СТАТИСТИКА ПОГРЕШНОСТЕЙ")
	fmt.Println(strings.Repeat("=", 141))

	allMethods := []struct {
		name   string
		points []SolutionPoint
	}{
		{"Эйлер", results.euler},
		{"Хойн", results.heun},
		{"RK3 (4 способ)", results.rk3Alt},
		{"RK4", results.rk4},
		{"Гаусс-2", results.gauss2},
		{"Радо IIA-3", results.radauIIA3},
	}

	for _, method := range allMethods {
		maxError := 0.0
		avgError := 0.0
		minError := math.Inf(1)
		for _, p := range method.points {
			avgError += p.error
			if p.error > maxError {
				maxError = p.error
			}
			if p.error < minError {
				minError = p.error
			}
		}
		avgError /= float64(len(method.points))
		fmt.Printf("%-15s: Мин = %12.2e, Макс = %12.2e, Средняя = %12.2e\n",
			method.name, minError, maxError, avgError)
	}
	fmt.Println(strings.Repeat("=", 141) + "\n")

	// Создаем вкладки
	var tabs []*container.TabItem

	// Информация
	infoText := `РЕШЕНИЕ ЗАДАЧИ КОШИ ДЛЯ ОДУ ПЕРВОГО ПОРЯДКА

Уравнение: x²y' + y² - xy + x² = 0
Преобразовано к виду: y' = -y²/x² + y/x - 1

Начальное условие: y(1) = tg(1) ≈ 1.5574
Аналитическое решение: y = x·tg(1 - ln x)

Интервал интегрирования: [1; 8]
Шаг интегрирования: h = 0.25

ЯВНЫЕ МЕТОДЫ:

1. Метод Эйлера (1 порядок):
   y_{k+1} = y_k + hf(x_k, y_k)

2. Метод Хойна (2 порядок):
   K₁ = hf(x_k, y_k)
   K₂ = hf(x_k + h/2, y_k + K₁/2)
   y_{k+1} = y_k + K₂

3. Классический RK4 (4 порядок):
   K₁ = hf(x_k, y_k)
   K₂ = hf(x_k + h/2, y_k + K₁/2)
   K₃ = hf(x_k + h/2, y_k + K₂/2)
   K₄ = hf(x_k + h, y_k + K₃)
   y_{k+1} = y_k + (K₁ + 2K₂ + 2K₃ + K₄)/6

4. Метод 3 порядка (4 способ):
   K₁ = hf(x_k, y_k)
   K₂ = hf(x_k + h/2, y_k + K₁/2)
   K₃ = hf(x_k + 3h/4, y_k + 3K₂/4)
   y_{k+1} = y_k + (2K₁ + 3K₂ + 4K₃)/9

НЕЯВНЫЕ МЕТОДЫ:

5. Метод Гаусса 2 порядка (Правило средней точки):
   K₁ = f(x_k + h/2, y_k + hK₁/2)
   y_{k+1} = y_k + hK₁
   (решается итерационно)

6. Метод Радо IIA 3 порядка:
   K₁ = f(x_k + h/3, y_k + 5hK₁/12 - hK₂/12)
   K₂ = f(x_k + h, y_k + 3hK₁/4 + hK₂/4)
   y_{k+1} = y_k + 3hK₁/4 + hK₂/4
   (решается итерационно)

Неявные методы решаются методом простых итераций
с точностью 10⁻¹⁰ и максимум 100 итераций.`

	infoContainer := container.NewVScroll(widget.NewLabel(infoText))
	tabs = append(tabs, container.NewTabItem("Информация", infoContainer))

	// Общий график со всеми методами
	combinedPlot := createCombinedPlot(results, 1100, 600)
	combinedContent := container.NewVBox(
		widget.NewLabel("СРАВНЕНИЕ ВСЕХ МЕТОДОВ"),
		widget.NewLabel("Черная линия — аналитическое решение"),
		widget.NewLabel("Цветные линии с точками — численные методы"),
		widget.NewSeparator(),
		combinedPlot,
	)
	tabs = append(tabs, container.NewTabItem("Все методы", container.NewVScroll(combinedContent)))

	// Графики для каждого метода
	methodColors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Эйлер - красный
		color.RGBA{0, 150, 0, 255},   // Хойн - зеленый
		color.RGBA{255, 165, 0, 255}, // RK3Alt - оранжевый
		color.RGBA{0, 0, 255, 255},   // RK4 - синий
		color.RGBA{128, 0, 128, 255}, // Gauss2 - фиолетовый
		color.RGBA{0, 200, 200, 255}, // RadauIIA3 - голубой
	}

	methodData := []struct {
		name   string
		points []SolutionPoint
		color  color.Color
	}{
		{"Метод Эйлера (1 порядок)", results.euler, methodColors[0]},
		{"Метод Хойна (2 порядок)", results.heun, methodColors[1]},
		{"Метод 3 порядка (4 способ)", results.rk3Alt, methodColors[2]},
		{"Классический RK4 (4 порядок)", results.rk4, methodColors[3]},
		{"Метод Гаусса-2 (неявный)", results.gauss2, methodColors[4]},
		{"Метод Радо IIA-3 (неявный)", results.radauIIA3, methodColors[5]},
	}

	for _, method := range methodData {
		plot := createMethodPlot(method.points, method.name, method.color, 1100, 600)
		content := container.NewVBox(
			widget.NewLabel(method.name),
			widget.NewLabel("Черная линия — аналитическое решение"),
			widget.NewLabel("Цветная линия с точками — численное решение"),
			widget.NewSeparator(),
			plot,
		)
		tabs = append(tabs, container.NewTabItem(method.name, container.NewVScroll(content)))
	}

	// График погрешностей
	errorPlot := createErrorPlot(results, 1100, 600)
	errorPlotContent := container.NewVBox(
		widget.NewLabel("СРАВНЕНИЕ ПОГРЕШНОСТЕЙ ВСЕХ МЕТОДОВ"),
		widget.NewLabel("График показывает абсолютную погрешность |y_численное - y_аналитическое| для каждого метода"),
		widget.NewLabel("Ось Y в логарифмической шкале"),
		widget.NewSeparator(),
		errorPlot,
	)
	tabs = append(tabs, container.NewTabItem("Погрешности (график)", container.NewVScroll(errorPlotContent)))

	// Таблица результатов
	resultsText := createResultsTable(results)
	resultsContainer := container.NewVScroll(widget.NewLabel(resultsText))
	tabs = append(tabs, container.NewTabItem("Результаты (таблица)", resultsContainer))

	// Таблица погрешностей
	errorText := createErrorTable(results)
	errorContainer := container.NewVScroll(widget.NewLabel(errorText))
	tabs = append(tabs, container.NewTabItem("Погрешности (таблица)", errorContainer))

	tabContainer := container.NewAppTabs(tabs...)

	myWindow.SetContent(tabContainer)
	myWindow.ShowAndRun()
}
