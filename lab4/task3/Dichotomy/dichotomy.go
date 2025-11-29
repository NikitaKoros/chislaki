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

// Исходное уравнение: xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0
// Граничные условия: y(0.1) + y'(0.1) = 5.399, y(1) = 40.171
// Аналитическое решение: y(x) = (1 + x^5)e^(3x)

// Аналитическое решение
func analyticalSolution(x float64) float64 {
	return (1 + math.Pow(x, 5)) * math.Exp(3*x)
}

// Производная аналитического решения
func analyticalDerivative(x float64) float64 {
	// y = (1 + 2x^5)e^(3x)
	// y' = 10x^4*e^(3x) + 3(1 + 2x^5)e^(3x) = (3 + 10x^4 + 6x^5)e^(3x)
	return (3 + 10*math.Pow(x, 4) + 6*math.Pow(x, 5)) * math.Exp(3*x)
}

// Система уравнений первого порядка:
// y' = z
// z' = (2(3x + 2)z - 3(3x + 4)y) / x

// Функция f(x, y, z) = z
func f(x, y, z float64) float64 {
	return z
}

// Функция g(x, y, z) = (2(3x + 2)z - 3(3x + 4)y) / x
func g(x, y, z float64) float64 {
	if math.Abs(x) < 1e-10 {
		return 0
	}
	return (2*(3*x+2)*z - 3*(3*x+4)*y) / x
}

// Точка решения
type RK4Point struct {
	x         float64
	y         float64
	z         float64
	yAnalytic float64
	error     float64
	K1y       float64
	K2y       float64
	K3y       float64
	K4y       float64
	K1z       float64
	K2z       float64
	K3z       float64
	K4z       float64
	deltaY    float64
	deltaZ    float64
}

// Итерация стрельбы
type ShootingIteration struct {
	iteration int
	eta       float64
	z0        float64
	yEnd      float64
	error     float64
	comment   string
}

// Траектория для конкретного alpha
type Trajectory struct {
	points []RK4Point
	alpha  float64
}

// Глобальное хранилище траекторий
var globalTrajectoryHistory []Trajectory

// Решение задачи Коши методом Рунге-Кутты 4-го порядка
func solveRK4Cauchy(a, b, h, y0, z0 float64) []RK4Point {
	n := int((b-a)/h) + 1
	points := make([]RK4Point, n)

	y := y0
	z := z0
	x := a

	for i := 0; i < n; i++ {
		// Сохраняем текущее состояние
		points[i] = RK4Point{
			x:         x,
			y:         y,
			z:         z,
			yAnalytic: analyticalSolution(x),
			error:     math.Abs(y - analyticalSolution(x)),
		}

		// Делаем шаг Рунге-Кутты 4-го порядка (если не последняя точка)
		if i < n-1 {
			// K1 = h*f(x, y, z)
			K1y := h * f(x, y, z)
			K1z := h * g(x, y, z)

			// K2 = h*f(x + h/2, y + K1y/2, z + K1z/2)
			K2y := h * f(x+h/2, y+K1y/2, z+K1z/2)
			K2z := h * g(x+h/2, y+K1y/2, z+K1z/2)

			// K3 = h*f(x + h/2, y + K2y/2, z + K2z/2)
			K3y := h * f(x+h/2, y+K2y/2, z+K2z/2)
			K3z := h * g(x+h/2, y+K2y/2, z+K2z/2)

			// K4 = h*f(x + h, y + K3y, z + K3z)
			K4y := h * f(x+h, y+K3y, z+K3z)
			K4z := h * g(x+h, y+K3y, z+K3z)

			// Δy = (K1y + 2*K2y + 2*K3y + K4y) / 6
			deltaY := (K1y + 2*K2y + 2*K3y + K4y) / 6
			deltaZ := (K1z + 2*K2z + 2*K3z + K4z) / 6

			// Сохраняем коэффициенты K для отображения
			points[i].K1y = K1y
			points[i].K2y = K2y
			points[i].K3y = K3y
			points[i].K4y = K4y
			points[i].K1z = K1z
			points[i].K2z = K2z
			points[i].K3z = K3z
			points[i].K4z = K4z
			points[i].deltaY = deltaY
			points[i].deltaZ = deltaZ

			// Обновляем значения
			y = y + deltaY
			z = z + deltaZ
			x = x + h
		}
	}

	return points
}

// Метод дихотомии для метода стрельбы
func shootingMethodDichotomy(a, b, h, yaBoundary, yb float64, maxIter int, epsilon float64) ([]RK4Point, []ShootingIteration) {
	iterations := make([]ShootingIteration, 0)
	globalTrajectoryHistory = make([]Trajectory, 0) // Очищаем историю траекторий

	// Граничное условие: y(a) + y'(a) = yaBoundary
	// Параметризация через параметр стрельбы: y'(a) = tg(α)
	// Тогда y(a) = yaBoundary - tg(α)

	// Начальное приближение по формуле из текста:
	// tg(α₀) = (y_b - y_a) / (b - a)
	// Поскольку y_a неизвестен, выбираем разумное начальное значение
	// На основе граничных условий: y(a) ≈ 1.0, y'(a) ≈ 4.4
	eta0 := 1.0
	tgAlpha0 := yaBoundary - eta0 // tg(α₀)
	z0 := tgAlpha0

	points0 := solveRK4Cauchy(a, b, h, eta0, z0)
	yEnd0 := points0[len(points0)-1].y
	error0 := yEnd0 - yb

	// Сохраняем траекторию
	globalTrajectoryHistory = append(globalTrajectoryHistory, Trajectory{
		points: points0,
		alpha:  z0,
	})

	iterations = append(iterations, ShootingIteration{
		iteration: 0,
		eta:       eta0,
		z0:        z0,
		yEnd:      yEnd0,
		error:     error0,
		comment:   fmt.Sprintf("Начальное приближение: tg(α₀) = %.3f", tgAlpha0),
	})

	// Второе приближение - выбираем так, чтобы получить разные знаки ошибки
	// Корректируем угол наклона по методу дихотомии
	var eta1, z1, tgAlpha1 float64
	if yEnd0 < yb {
		// Нужно увеличить y(b), увеличиваем tg(α)
		tgAlpha1 = tgAlpha0 * 1.5
	} else {
		// Нужно уменьшить y(b), уменьшаем tg(α)
		tgAlpha1 = tgAlpha0 * 0.5
	}
	eta1 = yaBoundary - tgAlpha1
	z1 = tgAlpha1

	points1 := solveRK4Cauchy(a, b, h, eta1, z1)
	yEnd1 := points1[len(points1)-1].y
	error1 := yEnd1 - yb

	// Сохраняем траекторию
	globalTrajectoryHistory = append(globalTrajectoryHistory, Trajectory{
		points: points1,
		alpha:  z1,
	})

	var comment1 string
	if yEnd0 < yb {
		comment1 = fmt.Sprintf("y(b) < yb, увеличен tg(α₁) = %.3f", tgAlpha1)
	} else {
		comment1 = fmt.Sprintf("y(b) > yb, уменьшен tg(α₁) = %.3f", tgAlpha1)
	}

	iterations = append(iterations, ShootingIteration{
		iteration: 1,
		eta:       eta1,
		z0:        z1,
		yEnd:      yEnd1,
		error:     error1,
		comment:   comment1,
	})

	// Метод дихотомии
	etaLeft, etaRight := eta0, eta1
	errorLeft := error0

	if eta1 < eta0 {
		etaLeft, etaRight = eta1, eta0
		errorLeft = error1
	}

	var finalPoints []RK4Point
	for iter := 2; iter < maxIter; iter++ {
		etaMid := (etaLeft + etaRight) / 2
		zMid := yaBoundary - etaMid

		pointsMid := solveRK4Cauchy(a, b, h, etaMid, zMid)
		yEndMid := pointsMid[len(pointsMid)-1].y
		errorMid := yEndMid - yb

		// Сохраняем траекторию
		globalTrajectoryHistory = append(globalTrajectoryHistory, Trajectory{
			points: pointsMid,
			alpha:  zMid,
		})

		iterations = append(iterations, ShootingIteration{
			iteration: iter,
			eta:       etaMid,
			z0:        zMid,
			yEnd:      yEndMid,
			error:     errorMid,
			comment:   "η = (ηLeft + ηRight)/2",
		})

		if math.Abs(errorMid) <= epsilon {
			finalPoints = pointsMid
			break
		}

		// Обновляем интервал: выбираем половину, где знаки ошибок разные
		if (errorMid > 0 && errorLeft < 0) || (errorMid < 0 && errorLeft > 0) {
			// Корень между left и mid
			etaRight = etaMid
		} else {
			// Корень между mid и right
			etaLeft = etaMid
			errorLeft = errorMid
		}

		finalPoints = pointsMid
	}

	return finalPoints, iterations
}

// Форматирование таблицы RK4
func formatRK4Table(points []RK4Point, title string) string {
	var text string

	text += title + "\n"
	text += strings.Repeat("=", 140) + "\n\n"

	h := points[1].x - points[0].x
	text += fmt.Sprintf("h = %.2f\n", h)
	text += fmt.Sprintf("xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0  →  y' = z, z' = (2(3x+2)z - 3(3x+4)y)/x\n")
	text += fmt.Sprintf("y(%.1f) + y'(%.1f) = %.3f, y(%.1f) = %.3f\n\n",
		points[0].x, points[0].x, points[0].y+points[0].z, points[len(points)-1].x, points[len(points)-1].y)

	// Заголовок
	text += fmt.Sprintf("%-4s | %-8s | %-15s | %-10s | %-10s | %-12s | %-12s\n",
		"№", "Xi", "Точное решение", "ΔY", "ΔZ", "Y", "Z")
	text += strings.Repeat("-", 100) + "\n"

	// Строки
	for i, p := range points {
		text += fmt.Sprintf("%-4d | %-8.4f | %-15.6f | ",
			i, p.x, p.yAnalytic)

		if i < len(points)-1 {
			text += fmt.Sprintf("%-10.6f | %-10.6f | ", p.deltaY, p.deltaZ)
		} else {
			text += fmt.Sprintf("%-10s | %-10s | ", "-", "-")
		}

		text += fmt.Sprintf("%-12.6f | %-12.6f\n", p.y, p.z)
	}

	text += "\n"
	return text
}

// Форматирование таблицы итераций стрельбы
func formatShootingIterations(iterations []ShootingIteration, title string, yb float64) string {
	var text string

	text += title + "\n"
	text += strings.Repeat("=", 130) + "\n\n"

	// Заголовок
	text += fmt.Sprintf("%-10s | %-12s | %-15s | %-15s | %-15s | %-15s | %-40s\n",
		"Итерация", "tg(α)", "Угол наклона", "y(b) задано", "y(b) вычисл.", "Ошибка", "Комментарий")
	text += strings.Repeat("-", 130) + "\n"

	// Строки
	for _, it := range iterations {
		angleRad := math.Atan(it.z0)
		angleDeg := angleRad * 180.0 / math.Pi

		text += fmt.Sprintf("%-10d | %-12.6f | %-15.6f | %-15.6f | %-15.6f | %-15.6f | %-40s\n",
			it.iteration, it.z0, angleDeg, yb, it.yEnd, it.error, it.comment)
	}

	text += "\n"
	return text
}

// Печать таблицы RK4 в терминал
func printRK4TableToTerminal(points []RK4Point, title string) {
	fmt.Println(strings.Repeat("=", 140))
	fmt.Println(title)
	fmt.Println(strings.Repeat("=", 140))

	h := points[1].x - points[0].x
	fmt.Printf("\nh = %.2f\n", h)
	fmt.Println("xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0  →  y' = z, z' = (2(3x+2)z - 3(3x+4)y)/x")
	fmt.Printf("y(%.1f) + y'(%.1f) = %.3f, y(%.1f) = %.3f\n\n",
		points[0].x, points[0].x, points[0].y+points[0].z, points[len(points)-1].x, points[len(points)-1].y)

	// Заголовок
	fmt.Printf("%-4s | %-8s | %-15s | %-10s | %-10s | %-12s | %-12s\n",
		"№", "Xi", "Точное решение", "ΔY", "ΔZ", "Y", "Z")
	fmt.Println(strings.Repeat("-", 100))

	// Строки
	for i, p := range points {
		fmt.Printf("%-4d | %-8.4f | %-15.6f | ",
			i, p.x, p.yAnalytic)

		if i < len(points)-1 {
			fmt.Printf("%-10.6f | %-10.6f | ", p.deltaY, p.deltaZ)
		} else {
			fmt.Printf("%-10s | %-10s | ", "-", "-")
		}

		fmt.Printf("%-12.6f | %-12.6f\n", p.y, p.z)
	}

	fmt.Println()
}

// Печать таблицы итераций в терминал
func printShootingIterationsToTerminal(iterations []ShootingIteration, title string, yb float64) {
	fmt.Println(strings.Repeat("=", 130))
	fmt.Println(title)
	fmt.Println(strings.Repeat("=", 130))
	fmt.Println()

	// Заголовок
	fmt.Printf("%-10s | %-12s | %-15s | %-15s | %-15s | %-15s | %-40s\n",
		"Итерация", "tg(α)", "Угол наклона", "y(b) задано", "y(b) вычисл.", "Ошибка", "Комментарий")
	fmt.Println(strings.Repeat("-", 130))

	// Строки
	for _, it := range iterations {
		angleRad := math.Atan(it.z0)
		angleDeg := angleRad * 180.0 / math.Pi

		fmt.Printf("%-10d | %-12.6f | %-15.6f | %-15.6f | %-15.6f | %-15.6f | %-40s\n",
			it.iteration, it.z0, angleDeg, yb, it.yEnd, it.error, it.comment)
	}

	fmt.Println()
}

// Создание графика решения с траекториями для разных alpha
func createSolutionPlotWithTrajectories(points []RK4Point, width, height float64, title string) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	// Определяем границы с учетом всех траекторий
	xMin := points[0].x
	xMax := points[len(points)-1].x

	// Фиксированные границы по оси Y
	yMin := -1500.0
	yMax := 1500.0

	marginLeft, marginRight := 80.0, 40.0
	marginTop, marginBottom := 60.0, 100.0
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
		label.Move(fyne.NewPos(float32(marginLeft)-60, yToPixel(y)-5))
		objects = append(objects, label)
	}

	// Заголовок
	titleLabel := canvas.NewText(title, color.Black)
	titleLabel.TextSize = 14
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Move(fyne.NewPos(float32(width/2)-150, 10))
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
	xLabel.Move(fyne.NewPos(float32(width)-50, float32(height)-70))
	objects = append(objects, xLabel)

	// Цвета для траекторий (каждая третья)
	trajectoryColors := []color.RGBA{
		{255, 100, 100, 200}, // Светло-красный
		{100, 255, 100, 200}, // Светло-зеленый
		{255, 150, 0, 200},   // Оранжевый
		{150, 100, 255, 200}, // Фиолетовый
		{255, 200, 100, 200}, // Желтоватый
		{100, 200, 255, 200}, // Голубой
	}

	// Рисуем каждую третью траекторию (как в Python коде)
	colorIdx := 0
	for i := 0; i < len(globalTrajectoryHistory); i += 3 {
		traj := globalTrajectoryHistory[i]
		trajColor := trajectoryColors[colorIdx%len(trajectoryColors)]
		colorIdx++

		// Рисуем линии траектории
		for j := 0; j < len(traj.points)-1; j++ {
			line := canvas.NewLine(trajColor)
			line.Position1 = fyne.NewPos(xToPixel(traj.points[j].x), yToPixel(traj.points[j].y))
			line.Position2 = fyne.NewPos(xToPixel(traj.points[j+1].x), yToPixel(traj.points[j+1].y))
			line.StrokeWidth = 1
			objects = append(objects, line)
		}
	}

	// Аналитическое решение (черная линия)
	for i := 0; i < len(points)-1; i++ {
		line := canvas.NewLine(color.Black)
		line.Position1 = fyne.NewPos(xToPixel(points[i].x), yToPixel(points[i].yAnalytic))
		line.Position2 = fyne.NewPos(xToPixel(points[i+1].x), yToPixel(points[i+1].yAnalytic))
		line.StrokeWidth = 3
		objects = append(objects, line)
	}

	// Численное решение (синяя линия)
	blueColor := color.RGBA{0, 0, 255, 255}
	for i := 0; i < len(points)-1; i++ {
		line := canvas.NewLine(blueColor)
		line.Position1 = fyne.NewPos(xToPixel(points[i].x), yToPixel(points[i].y))
		line.Position2 = fyne.NewPos(xToPixel(points[i+1].x), yToPixel(points[i+1].y))
		line.StrokeWidth = 2
		objects = append(objects, line)
	}
	for _, p := range points {
		circle := canvas.NewCircle(blueColor)
		circle.FillColor = blueColor
		circle.Resize(fyne.NewSize(6, 6))
		circle.Move(fyne.NewPos(xToPixel(p.x)-3, yToPixel(p.y)-3))
		objects = append(objects, circle)
	}

	// Легенда
	legendX := float32(width) - 250
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

	blueLine := canvas.NewLine(blueColor)
	blueLine.Position1 = fyne.NewPos(legendX, legendY+25)
	blueLine.Position2 = fyne.NewPos(legendX+30, legendY+25)
	blueLine.StrokeWidth = 2
	objects = append(objects, blueLine)
	blueLabel := canvas.NewText("Численное (финальное)", color.Black)
	blueLabel.TextSize = 9
	blueLabel.Move(fyne.NewPos(legendX+35, legendY+20))
	objects = append(objects, blueLabel)

	// Легенда для траекторий
	colorIdx = 0
	for i := 0; i < len(globalTrajectoryHistory) && colorIdx < 5; i += 3 {
		traj := globalTrajectoryHistory[i]
		trajColor := trajectoryColors[colorIdx%len(trajectoryColors)]

		trajLine := canvas.NewLine(trajColor)
		trajLine.Position1 = fyne.NewPos(legendX, legendY+45+float32(colorIdx)*20)
		trajLine.Position2 = fyne.NewPos(legendX+30, legendY+45+float32(colorIdx)*20)
		trajLine.StrokeWidth = 1
		objects = append(objects, trajLine)

		trajLabel := canvas.NewText(fmt.Sprintf("α = %.4f", traj.alpha), color.Black)
		trajLabel.TextSize = 8
		trajLabel.Move(fyne.NewPos(legendX+35, legendY+40+float32(colorIdx)*20))
		objects = append(objects, trajLabel)

		colorIdx++
	}

	return container.NewWithoutLayout(objects...)
}

// Создание графика решения
func createSolutionPlot(points []RK4Point, width, height float64, title string) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	// Границы
	xMin := points[0].x
	xMax := points[len(points)-1].x
	yMin, yMax := math.Inf(1), math.Inf(-1)

	for _, p := range points {
		if p.yAnalytic < yMin {
			yMin = p.yAnalytic
		}
		if p.yAnalytic > yMax {
			yMax = p.yAnalytic
		}
		if p.y < yMin {
			yMin = p.y
		}
		if p.y > yMax {
			yMax = p.y
		}
	}
	yRange := yMax - yMin
	yMin -= yRange * 0.1
	yMax += yRange * 0.1

	marginLeft, marginRight := 80.0, 40.0
	marginTop, marginBottom := 60.0, 80.0
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
		label.Move(fyne.NewPos(float32(marginLeft)-60, yToPixel(y)-5))
		objects = append(objects, label)
	}

	// Заголовок
	titleLabel := canvas.NewText(title, color.Black)
	titleLabel.TextSize = 14
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Move(fyne.NewPos(float32(width/2)-150, 10))
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

	// Аналитическое решение (черная линия)
	for i := 0; i < len(points)-1; i++ {
		line := canvas.NewLine(color.Black)
		line.Position1 = fyne.NewPos(xToPixel(points[i].x), yToPixel(points[i].yAnalytic))
		line.Position2 = fyne.NewPos(xToPixel(points[i+1].x), yToPixel(points[i+1].yAnalytic))
		line.StrokeWidth = 3
		objects = append(objects, line)
	}

	// Численное решение (синяя линия)
	blueColor := color.RGBA{0, 0, 255, 255}
	for i := 0; i < len(points)-1; i++ {
		line := canvas.NewLine(blueColor)
		line.Position1 = fyne.NewPos(xToPixel(points[i].x), yToPixel(points[i].y))
		line.Position2 = fyne.NewPos(xToPixel(points[i+1].x), yToPixel(points[i+1].y))
		line.StrokeWidth = 2
		objects = append(objects, line)
	}
	for _, p := range points {
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

	blueLine := canvas.NewLine(blueColor)
	blueLine.Position1 = fyne.NewPos(legendX, legendY+25)
	blueLine.Position2 = fyne.NewPos(legendX+30, legendY+25)
	blueLine.StrokeWidth = 2
	objects = append(objects, blueLine)
	blueLabel := canvas.NewText("Численное", color.Black)
	blueLabel.TextSize = 9
	blueLabel.Move(fyne.NewPos(legendX+35, legendY+20))
	objects = append(objects, blueLabel)

	return container.NewWithoutLayout(objects...)
}

// Создание графика погрешности
func createErrorPlot(points []RK4Point, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	xMin := points[0].x
	xMax := points[len(points)-1].x
	yMin, yMax := 1e-15, 1.0

	for _, p := range points {
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

	// Рисуем линии ошибок
	redColor := color.RGBA{255, 0, 0, 255}
	for i := 0; i < len(points)-1; i++ {
		if points[i].error > 0 && points[i+1].error > 0 {
			line := canvas.NewLine(redColor)
			line.Position1 = fyne.NewPos(xToPixel(points[i].x), yToPixel(points[i].error))
			line.Position2 = fyne.NewPos(xToPixel(points[i+1].x), yToPixel(points[i+1].error))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}
	for _, p := range points {
		if p.error > 0 {
			circle := canvas.NewCircle(redColor)
			circle.FillColor = redColor
			circle.Resize(fyne.NewSize(5, 5))
			circle.Move(fyne.NewPos(xToPixel(p.x)-2.5, yToPixel(p.error)-2.5))
			objects = append(objects, circle)
		}
	}

	return container.NewWithoutLayout(objects...)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Метод стрельбы - Метод дихотомии")
	myWindow.Resize(fyne.NewSize(1200, 900))

	// Параметры задачи
	a := 0.1
	b := 1.0
	h := 0.09
	yaBoundary := 5.399 // y(a) + z(a) = 5.399
	yb := 40.171        // y(b) = 40.171
	maxIter := 20
	epsilon := 0.01

	// Решаем методом дихотомии
	points, iterations := shootingMethodDichotomy(a, b, h, yaBoundary, yb, maxIter, epsilon)

	// Выводим результаты в терминал
	fmt.Println()
	fmt.Println(strings.Repeat("=", 140))
	fmt.Println("МЕТОД СТРЕЛЬБЫ ДЛЯ КРАЕВОЙ ЗАДАЧИ - МЕТОД ДИХОТОМИИ")
	fmt.Println(strings.Repeat("=", 140))
	fmt.Println()
	fmt.Println("Уравнение: xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0")
	fmt.Printf("Граничные условия: y(%.1f) + y'(%.1f) = %.3f, y(%.1f) = %.3f\n", a, a, yaBoundary, b, yb)
	fmt.Printf("Интервал: [%.1f; %.1f], h = %.2f\n", a, b, h)
	fmt.Printf("Аналитическое решение: y(x) = (1 + x⁵)e³ˣ\n")
	fmt.Println()

	printShootingIterationsToTerminal(iterations, "ИТЕРАЦИИ МЕТОДА ДИХОТОМИИ", yb)
	printRK4TableToTerminal(points, "РЕШЕНИЕ ЗАДАЧИ КОШИ МЕТОДОМ РУНГЕ-КУТТЫ 4-ГО ПОРЯДКА")

	// Статистика
	maxError, avgError := 0.0, 0.0
	for _, p := range points {
		avgError += p.error
		if p.error > maxError {
			maxError = p.error
		}
	}
	avgError /= float64(len(points))
	fmt.Printf("Средняя ошибка = %.6e, Максимальная ошибка = %.6e\n", avgError, maxError)
	fmt.Printf("Количество итераций: %d\n", len(iterations))
	fmt.Printf("Финальное значение η = %.6f, z(a) = %.6f\n", iterations[len(iterations)-1].eta, iterations[len(iterations)-1].z0)
	fmt.Printf("y(b) вычисленное = %.6f, y(b) заданное = %.6f, погрешность = %.6e\n",
		iterations[len(iterations)-1].yEnd, yb, iterations[len(iterations)-1].error)
	fmt.Println(strings.Repeat("=", 140))
	fmt.Println()

	// Создаем вкладки
	var tabs []*container.TabItem

	// Информация
	infoText := `МЕТОД СТРЕЛЬБЫ - МЕТОД ДИХОТОМИИ

Исходная задача:
  xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0

Граничные условия:
  y(0.1) + y'(0.1) = 5.399
  y(1) = 40.171

Аналитическое решение:
  y(x) = (1 + x⁵)e³ˣ

Параметры:
  Интервал: [0.1; 1]
  Шаг интегрирования: h = 0.09
  Количество точек: 11

МЕТОД РЕШЕНИЯ:

1. Сведение к системе уравнений первого порядка:
   y' = z
   z' = (2(3x + 2)z - 3(3x + 4)y) / x

2. Преобразование граничного условия:
   Условие y(0.1) + y'(0.1) = 5.399 преобразуется в:
   y(0.1) + z(0.1) = 5.399

   Вводим параметр η = y(0.1), тогда z(0.1) = 5.399 - η

3. Итерационный процесс подбора параметра η (метод дихотомии):
   - Выбираем начальное η₀ и вычисляем z₀ = 5.399 - η₀
   - Решаем задачу Коши методом Рунге-Кутты 4-го порядка
   - Если |y(b) - 40.171| > ε, корректируем η методом половинного деления:
     η_{k+1} = (η_{k-1} + η_k) / 2
   - Повторяем до достижения точности

4. Решение задачи Коши методом Рунге-Кутты 4-го порядка:
   Используется классический метод RK4 с постоянным шагом h = 0.09

   K₁ʸ = h·f(xᵢ, yᵢ, zᵢ)
   K₁ᶻ = h·g(xᵢ, yᵢ, zᵢ)

   K₂ʸ = h·f(xᵢ + h/2, yᵢ + K₁ʸ/2, zᵢ + K₁ᶻ/2)
   K₂ᶻ = h·g(xᵢ + h/2, yᵢ + K₁ʸ/2, zᵢ + K₁ᶻ/2)

   K₃ʸ = h·f(xᵢ + h/2, yᵢ + K₂ʸ/2, zᵢ + K₂ᶻ/2)
   K₃ᶻ = h·g(xᵢ + h/2, yᵢ + K₂ʸ/2, zᵢ + K₂ᶻ/2)

   K₄ʸ = h·f(xᵢ + h, yᵢ + K₃ʸ, zᵢ + K₃ᶻ)
   K₄ᶻ = h·g(xᵢ + h, yᵢ + K₃ʸ, zᵢ + K₃ᶻ)

   Δyᵢ = (K₁ʸ + 2K₂ʸ + 2K₃ʸ + K₄ʸ) / 6
   Δzᵢ = (K₁ᶻ + 2K₂ᶻ + 2K₃ᶻ + K₄ᶻ) / 6

   yᵢ₊₁ = yᵢ + Δyᵢ
   zᵢ₊₁ = zᵢ + Δzᵢ`

	infoContainer := container.NewVScroll(widget.NewLabel(infoText))
	tabs = append(tabs, container.NewTabItem("Информация", infoContainer))

	// Таблица итераций
	iterText := formatShootingIterations(iterations, "ИТЕРАЦИИ МЕТОДА ДИХОТОМИИ", yb)
	iterText += fmt.Sprintf("\nКоличество итераций: %d\n", len(iterations))
	iterText += fmt.Sprintf("Финальное значение η = %.6f\n", iterations[len(iterations)-1].eta)
	iterText += fmt.Sprintf("Финальное значение z(a) = %.6f\n", iterations[len(iterations)-1].z0)
	iterText += fmt.Sprintf("y(b) вычисленное = %.6f\n", iterations[len(iterations)-1].yEnd)
	iterText += fmt.Sprintf("y(b) заданное = %.6f\n", yb)
	iterText += fmt.Sprintf("Погрешность на правой границе = %.6e\n", iterations[len(iterations)-1].error)

	iterContainer := container.NewVScroll(widget.NewLabel(iterText))
	tabs = append(tabs, container.NewTabItem("Итерации", iterContainer))

	// Таблица RK4
	rk4Text := formatRK4Table(points, "РЕШЕНИЕ ЗАДАЧИ КОШИ МЕТОДОМ РУНГЕ-КУТТЫ 4-ГО ПОРЯДКА")
	rk4Text += fmt.Sprintf("\nСтатистика:\n")
	rk4Text += fmt.Sprintf("Средняя ошибка = %.6e\n", avgError)
	rk4Text += fmt.Sprintf("Максимальная ошибка = %.6e\n", maxError)

	rk4Container := container.NewVScroll(widget.NewLabel(rk4Text))
	tabs = append(tabs, container.NewTabItem("Таблица RK4", rk4Container))

	// График решения
	plot := createSolutionPlot(points, 1100, 600, "Решение методом стрельбы (Дихотомия)")
	plotContent := container.NewVBox(
		widget.NewLabel("ГРАФИК РЕШЕНИЯ"),
		widget.NewLabel("Черная линия — аналитическое решение"),
		widget.NewLabel("Синяя линия — численное решение"),
		widget.NewSeparator(),
		plot,
	)
	tabs = append(tabs, container.NewTabItem("График решения", container.NewVScroll(plotContent)))

	// График с траекториями для разных alpha
	plotWithTrajectories := createSolutionPlotWithTrajectories(points, 1100, 600, "Траектории для разных значений α")
	plotTrajContent := container.NewVBox(
		widget.NewLabel("ГРАФИК С ТРАЕКТОРИЯМИ"),
		widget.NewLabel("Показаны траектории для каждого третьего значения α в процессе итераций"),
		widget.NewLabel("Черная линия — аналитическое решение"),
		widget.NewLabel("Синяя линия — финальное численное решение"),
		widget.NewLabel("Цветные линии — промежуточные траектории для разных α"),
		widget.NewSeparator(),
		plotWithTrajectories,
	)
	tabs = append(tabs, container.NewTabItem("Траектории α", container.NewVScroll(plotTrajContent)))

	// График погрешности
	errorPlot := createErrorPlot(points, 1100, 600)
	errorPlotContent := container.NewVBox(
		widget.NewLabel("ПОГРЕШНОСТЬ"),
		widget.NewLabel("График показывает абсолютную погрешность |y_численное - y_аналитическое|"),
		widget.NewSeparator(),
		errorPlot,
	)
	tabs = append(tabs, container.NewTabItem("Погрешность", container.NewVScroll(errorPlotContent)))

	tabContainer := container.NewAppTabs(tabs...)

	myWindow.SetContent(tabContainer)
	myWindow.ShowAndRun()
}
