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
type EulerPoint struct {
	x         float64
	y         float64
	z         float64
	yAnalytic float64
	error     float64
	deltaZ    float64
	deltaY    float64
	zPrime    float64
	yPrime    float64
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

// Решение задачи Коши методом Эйлера
func solveEulerCauchy(a, b, h, y0, z0 float64) []EulerPoint {
	n := int((b-a)/h) + 1
	points := make([]EulerPoint, n)

	y := y0
	z := z0
	x := a

	for i := 0; i < n; i++ {
		// Вычисляем значения производных
		zPrime := g(x, y, z)
		yPrime := f(x, y, z)

		// Сохраняем текущее состояние
		points[i] = EulerPoint{
			x:         x,
			y:         y,
			z:         z,
			yAnalytic: analyticalSolution(x),
			error:     math.Abs(y - analyticalSolution(x)),
			zPrime:    zPrime,
			yPrime:    yPrime,
		}

		// Делаем шаг Эйлера (если не последняя точка)
		if i < n-1 {
			deltaZ := h * zPrime
			deltaY := h * yPrime

			points[i].deltaZ = deltaZ
			points[i].deltaY = deltaY

			z = z + deltaZ
			y = y + deltaY
			x = x + h
		}
	}

	return points
}

// Метод дихотомии для метода стрельбы
func shootingMethodDichotomy(a, b, h, yaBoundary, yb float64, maxIter int, epsilon float64) ([]EulerPoint, []ShootingIteration) {
	iterations := make([]ShootingIteration, 0)

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

	points0 := solveEulerCauchy(a, b, h, eta0, z0)
	yEnd0 := points0[len(points0)-1].y
	error0 := yEnd0 - yb

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

	points1 := solveEulerCauchy(a, b, h, eta1, z1)
	yEnd1 := points1[len(points1)-1].y
	error1 := yEnd1 - yb

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

	var finalPoints []EulerPoint
	for iter := 2; iter < maxIter; iter++ {
		etaMid := (etaLeft + etaRight) / 2
		zMid := yaBoundary - etaMid

		pointsMid := solveEulerCauchy(a, b, h, etaMid, zMid)
		yEndMid := pointsMid[len(pointsMid)-1].y
		errorMid := yEndMid - yb

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

func shootingMethodNewton(a, b, h, yaBoundary, ybTarget float64, maxIter int, epsilon float64) ([]EulerPoint, []ShootingIteration) {
	iterations := make([]ShootingIteration, 0)

	// α = y'(a) - тангенс угла наклона (как в тексте)
	alphaPrev := 2.0                   // начальное приближение для α₀
	y_a_prev := yaBoundary - alphaPrev // y(a) = 5.399 - y'(a)

	pointsPrev := solveEulerCauchy(a, b, h, y_a_prev, alphaPrev)
	ybPrev := pointsPrev[len(pointsPrev)-1].y

	iterations = append(iterations, ShootingIteration{
		iteration: 0,
		eta:       y_a_prev,
		z0:        alphaPrev,
		yEnd:      ybPrev,
		error:     ybPrev - ybTarget,
		comment:   "Начальное приближение: α₀ = " + fmt.Sprintf("%.3f", alphaPrev),
	})

	// Второе приближение α₁
	alphaCurr := alphaPrev * 10.15
	y_a_curr := yaBoundary - alphaCurr

	pointsCurr := solveEulerCauchy(a, b, h, y_a_curr, alphaCurr)
	ybCurr := pointsCurr[len(pointsCurr)-1].y

	iterations = append(iterations, ShootingIteration{
		iteration: 1,
		eta:       y_a_curr,
		z0:        alphaCurr,
		yEnd:      ybCurr,
		error:     ybCurr - ybTarget,
		comment:   "Второе приближение: α₁ = " + fmt.Sprintf("%.3f", alphaCurr),
	})

	var finalPoints []EulerPoint

	for iter := 2; iter < maxIter; iter++ {
		// Формула из текста: αk+1 = αk + (αk - αk-1) * (yb - y(b,αk)) / (y(b,αk) - y(b,αk-1))
		deltaAlpha := alphaCurr - alphaPrev
		deltaY := ybCurr - ybPrev

		if math.Abs(deltaY) < 1e-10 {
			deltaY = 1e-10
		}

		// Правильная формула Ньютона (секущих) для α
		alphaNext := alphaCurr + deltaAlpha*(ybTarget-ybCurr)/deltaY
		y_a_next := yaBoundary - alphaNext

		pointsNext := solveEulerCauchy(a, b, h, y_a_next, alphaNext)
		ybNext := pointsNext[len(pointsNext)-1].y
		errorNext := ybNext - ybTarget

		iterations = append(iterations, ShootingIteration{
			iteration: iter,
			eta:       y_a_next,
			z0:        alphaNext,
			yEnd:      ybNext,
			error:     errorNext,
			comment:   "Итерация метода Ньютона",
		})

		if math.Abs(errorNext) <= epsilon {
			finalPoints = pointsNext
			break
		}

		// Обновляем для следующей итерации
		alphaPrev, alphaCurr = alphaCurr, alphaNext
		ybPrev, ybCurr = ybCurr, ybNext
		finalPoints = pointsNext
	}

	return finalPoints, iterations
}

// Форматирование таблицы Эйлера
func formatEulerTable(points []EulerPoint, title string) string {
	var text string

	text += title + "\n"
	text += strings.Repeat("=", 140) + "\n\n"

	h := points[1].x - points[0].x
	text += fmt.Sprintf("h = %.2f\n", h)
	text += fmt.Sprintf("xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0  →  y' = z, z' = (2(3x+2)z - 3(3x+4)y)/x\n")
	text += fmt.Sprintf("y(%.1f) + y'(%.1f) = %.3f, y(%.1f) = %.3f\n\n",
		points[0].x, points[0].x, points[0].y+points[0].z, points[len(points)-1].x, points[len(points)-1].y)

	// Заголовок
	text += fmt.Sprintf("%-4s | %-8s | %-15s | %-12s | %-12s | %-10s | %-10s | %-12s | %-12s\n",
		"№", "Xi", "Точное решение", "Z'=g(x,y,z)", "Y'=f(x,y,z)=Z", "ΔZ", "ΔY", "Z", "Y")
	text += strings.Repeat("-", 140) + "\n"

	// Строки
	for i, p := range points {
		text += fmt.Sprintf("%-4d | %-8.4f | %-15.6f | %-12.3f | %-12.3f | ",
			i, p.x, p.yAnalytic, p.zPrime, p.yPrime)

		if i < len(points)-1 {
			text += fmt.Sprintf("%-10.3f | %-10.3f | ", p.deltaZ, p.deltaY)
		} else {
			text += fmt.Sprintf("%-10s | %-10s | ", "-", "-")
		}

		text += fmt.Sprintf("%-12.3f | %-12.6f\n", p.z, p.y)
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

// Печать таблицы Эйлера в терминал
func printEulerTableToTerminal(points []EulerPoint, title string) {
	fmt.Println(strings.Repeat("=", 140))
	fmt.Println(title)
	fmt.Println(strings.Repeat("=", 140))

	h := points[1].x - points[0].x
	fmt.Printf("\nh = %.2f\n", h)
	fmt.Println("xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0  →  y' = z, z' = (2(3x+2)z - 3(3x+4)y)/x")
	fmt.Printf("y(%.1f) + y'(%.1f) = %.3f, y(%.1f) = %.3f\n\n",
		points[0].x, points[0].x, points[0].y+points[0].z, points[len(points)-1].x, points[len(points)-1].y)

	// Заголовок
	fmt.Printf("%-4s | %-8s | %-15s | %-12s | %-12s | %-10s | %-10s | %-12s | %-12s\n",
		"№", "Xi", "Точное решение", "Z'=g(x,y,z)", "Y'=f(x,y,z)=Z", "ΔZ", "ΔY", "Z", "Y")
	fmt.Println(strings.Repeat("-", 140))

	// Строки
	for i, p := range points {
		fmt.Printf("%-4d | %-8.4f | %-15.6f | %-12.3f | %-12.3f | ",
			i, p.x, p.yAnalytic, p.zPrime, p.yPrime)

		if i < len(points)-1 {
			fmt.Printf("%-10.3f | %-10.3f | ", p.deltaZ, p.deltaY)
		} else {
			fmt.Printf("%-10s | %-10s | ", "-", "-")
		}

		fmt.Printf("%-12.3f | %-12.6f\n", p.z, p.y)
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

// Создание графика решения
func createSolutionPlot(points []EulerPoint, width, height float64, title string) *fyne.Container {
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
func createErrorPlot(points []EulerPoint, width, height float64) *fyne.Container {
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
	myWindow := myApp.NewWindow("Метод стрельбы для краевых задач")
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
	pointsDich, iterDich := shootingMethodDichotomy(a, b, h, yaBoundary, yb, maxIter, epsilon)

	// Решаем методом Ньютона
	pointsNewton, iterNewton := shootingMethodNewton(a, b, h, yaBoundary, yb, maxIter, epsilon)

	// Выводим результаты в терминал
	fmt.Println()
	fmt.Println(strings.Repeat("=", 140))
	fmt.Println("МЕТОД СТРЕЛЬБЫ ДЛЯ КРАЕВОЙ ЗАДАЧИ")
	fmt.Println(strings.Repeat("=", 140))
	fmt.Println()
	fmt.Println("Уравнение: xy'' - 2(3x + 2)y' + 3(3x + 4)y = 0")
	fmt.Printf("Граничные условия: y(%.1f) + y'(%.1f) = %.3f, y(%.1f) = %.3f\n", a, a, yaBoundary, b, yb)
	fmt.Printf("Интервал: [%.1f; %.1f], h = %.2f\n", a, b, h)
	fmt.Printf("Аналитическое решение: y(x) = (1 + x⁵)e³ˣ\n")
	fmt.Println()

	// Метод дихотомии
	printShootingIterationsToTerminal(iterDich, "ИТЕРАЦИИ МЕТОДА ДИХОТОМИИ", yb)
	printEulerTableToTerminal(pointsDich, "РЕШЕНИЕ ЗАДАЧИ КОШИ МЕТОДОМ ЭЙЛЕРА (МЕТОД ДИХОТОМИИ)")

	// Статистика для дихотомии
	maxErrorDich, avgErrorDich := 0.0, 0.0
	for _, p := range pointsDich {
		avgErrorDich += p.error
		if p.error > maxErrorDich {
			maxErrorDich = p.error
		}
	}
	avgErrorDich /= float64(len(pointsDich))
	fmt.Printf("Метод дихотомии: Средняя ошибка = %.6e, Максимальная ошибка = %.6e\n", avgErrorDich, maxErrorDich)
	fmt.Printf("Количество итераций: %d\n", len(iterDich))
	fmt.Printf("Финальное значение η = %.6f, z(a) = %.6f\n", iterDich[len(iterDich)-1].eta, iterDich[len(iterDich)-1].z0)
	fmt.Printf("y(b) вычисленное = %.6f, y(b) заданное = %.6f, погрешность = %.6e\n\n",
		iterDich[len(iterDich)-1].yEnd, yb, iterDich[len(iterDich)-1].error)

	// Метод Ньютона
	printShootingIterationsToTerminal(iterNewton, "ИТЕРАЦИИ МЕТОДА НЬЮТОНА", yb)
	printEulerTableToTerminal(pointsNewton, "РЕШЕНИЕ ЗАДАЧИ КОШИ МЕТОДОМ ЭЙЛЕРА (МЕТОД НЬЮТОНА)")

	// Статистика для Ньютона
	maxErrorNewton, avgErrorNewton := 0.0, 0.0
	for _, p := range pointsNewton {
		avgErrorNewton += p.error
		if p.error > maxErrorNewton {
			maxErrorNewton = p.error
		}
	}
	avgErrorNewton /= float64(len(pointsNewton))
	fmt.Printf("Метод Ньютона: Средняя ошибка = %.6e, Максимальная ошибка = %.6e\n", avgErrorNewton, maxErrorNewton)
	fmt.Printf("Количество итераций: %d\n", len(iterNewton))
	fmt.Printf("Финальное значение η = %.6f, z(a) = %.6f\n", iterNewton[len(iterNewton)-1].eta, iterNewton[len(iterNewton)-1].z0)
	fmt.Printf("y(b) вычисленное = %.6f, y(b) заданное = %.6f, погрешность = %.6e\n",
		iterNewton[len(iterNewton)-1].yEnd, yb, iterNewton[len(iterNewton)-1].error)
	fmt.Println(strings.Repeat("=", 140))
	fmt.Println()

	// Создаем вкладки
	var tabs []*container.TabItem

	// Информация
	infoText := `МЕТОД СТРЕЛЬБЫ ДЛЯ КРАЕВЫХ ЗАДАЧ ПЕРВОГО РОДА

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

3. Итерационный процесс подбора параметра η:

   МЕТОД ДИХОТОМИИ (половинного деления):
   - Выбираем начальное η₀ и вычисляем z₀ = 5.399 - η₀
   - Решаем задачу Коши методом Эйлера
   - Если |y(b) - 40.171| > ε, корректируем η методом половинного деления:
     η_{k+1} = (η_{k-1} + η_k) / 2
   - Повторяем до достижения точности

   МЕТОД НЬЮТОНА (итерационная процедура):
   - Более быстрая сходимость
   - Использует производную функции y(b, η) по η
   - Формула: η_{k+1} = η_k + (y_b - y(b,η_k)) / ((y(b,η_k) - y(b,η_{k-1})) / (η_k - η_{k-1}))

4. Решение задачи Коши:
   Используется метод Эйлера с постоянным шагом h = 0.09

   y_{i+1} = y_i + h·f(x_i, y_i, z_i)
   z_{i+1} = z_i + h·g(x_i, y_i, z_i)

ОЖИДАЕМЫЕ РЕЗУЛЬТАТЫ:
  - Метод Ньютона сходится быстрее (меньше итераций)
  - Обе процедуры дают близкие результаты при достижении точности
  - Погрешность зависит от шага интегрирования h`

	infoContainer := container.NewVScroll(widget.NewLabel(infoText))
	tabs = append(tabs, container.NewTabItem("Информация", infoContainer))

	// Таблица итераций дихотомии
	iterDichText := formatShootingIterations(iterDich, "ИТЕРАЦИИ МЕТОДА ДИХОТОМИИ", yb)
	iterDichText += fmt.Sprintf("\nКоличество итераций: %d\n", len(iterDich))
	iterDichText += fmt.Sprintf("Финальное значение η = %.6f\n", iterDich[len(iterDich)-1].eta)
	iterDichText += fmt.Sprintf("Финальное значение z(a) = %.6f\n", iterDich[len(iterDich)-1].z0)
	iterDichText += fmt.Sprintf("y(b) вычисленное = %.6f\n", iterDich[len(iterDich)-1].yEnd)
	iterDichText += fmt.Sprintf("y(b) заданное = %.6f\n", yb)
	iterDichText += fmt.Sprintf("Погрешность на правой границе = %.6e\n", iterDich[len(iterDich)-1].error)

	iterDichContainer := container.NewVScroll(widget.NewLabel(iterDichText))
	tabs = append(tabs, container.NewTabItem("Итерации (Дихотомия)", iterDichContainer))

	// Таблица Эйлера для дихотомии
	eulerDichText := formatEulerTable(pointsDich, "РЕШЕНИЕ ЗАДАЧИ КОШИ МЕТОДОМ ЭЙЛЕРА (МЕТОД ДИХОТОМИИ)")
	eulerDichText += fmt.Sprintf("\nСтатистика:\n")
	eulerDichText += fmt.Sprintf("Средняя ошибка = %.6e\n", avgErrorDich)
	eulerDichText += fmt.Sprintf("Максимальная ошибка = %.6e\n", maxErrorDich)

	eulerDichContainer := container.NewVScroll(widget.NewLabel(eulerDichText))
	tabs = append(tabs, container.NewTabItem("Таблица Эйлера (Дихотомия)", eulerDichContainer))

	// График решения (дихотомия)
	plotDich := createSolutionPlot(pointsDich, 1100, 600, "Решение методом стрельбы (Дихотомия)")
	plotDichContent := container.NewVBox(
		widget.NewLabel("ГРАФИК РЕШЕНИЯ (МЕТОД ДИХОТОМИИ)"),
		widget.NewLabel("Черная линия — аналитическое решение"),
		widget.NewLabel("Синяя линия — численное решение"),
		widget.NewSeparator(),
		plotDich,
	)
	tabs = append(tabs, container.NewTabItem("График (Дихотомия)", container.NewVScroll(plotDichContent)))

	// График погрешности (дихотомия)
	errorPlotDich := createErrorPlot(pointsDich, 1100, 600)
	errorPlotDichContent := container.NewVBox(
		widget.NewLabel("ПОГРЕШНОСТЬ (МЕТОД ДИХОТОМИИ)"),
		widget.NewLabel("График показывает абсолютную погрешность |y_численное - y_аналитическое|"),
		widget.NewSeparator(),
		errorPlotDich,
	)
	tabs = append(tabs, container.NewTabItem("Погрешность (Дихотомия)", container.NewVScroll(errorPlotDichContent)))

	// Таблица итераций Ньютона
	iterNewtonText := formatShootingIterations(iterNewton, "ИТЕРАЦИИ МЕТОДА НЬЮТОНА", yb)
	iterNewtonText += fmt.Sprintf("\nКоличество итераций: %d\n", len(iterNewton))
	iterNewtonText += fmt.Sprintf("Финальное значение η = %.6f\n", iterNewton[len(iterNewton)-1].eta)
	iterNewtonText += fmt.Sprintf("Финальное значение z(a) = %.6f\n", iterNewton[len(iterNewton)-1].z0)
	iterNewtonText += fmt.Sprintf("y(b) вычисленное = %.6f\n", iterNewton[len(iterNewton)-1].yEnd)
	iterNewtonText += fmt.Sprintf("y(b) заданное = %.6f\n", yb)
	iterNewtonText += fmt.Sprintf("Погрешность на правой границе = %.6e\n", iterNewton[len(iterNewton)-1].error)

	iterNewtonContainer := container.NewVScroll(widget.NewLabel(iterNewtonText))
	tabs = append(tabs, container.NewTabItem("Итерации (Ньютон)", iterNewtonContainer))

	// Таблица Эйлера для Ньютона
	eulerNewtonText := formatEulerTable(pointsNewton, "РЕШЕНИЕ ЗАДАЧИ КОШИ МЕТОДОМ ЭЙЛЕРА (МЕТОД НЬЮТОНА)")
	eulerNewtonText += fmt.Sprintf("\nСтатистика:\n")
	eulerNewtonText += fmt.Sprintf("Средняя ошибка = %.6e\n", avgErrorNewton)
	eulerNewtonText += fmt.Sprintf("Максимальная ошибка = %.6e\n", maxErrorNewton)

	eulerNewtonContainer := container.NewVScroll(widget.NewLabel(eulerNewtonText))
	tabs = append(tabs, container.NewTabItem("Таблица Эйлера (Ньютон)", eulerNewtonContainer))

	// График решения (Ньютон)
	plotNewton := createSolutionPlot(pointsNewton, 1100, 600, "Решение методом стрельбы (Ньютон)")
	plotNewtonContent := container.NewVBox(
		widget.NewLabel("ГРАФИК РЕШЕНИЯ (МЕТОД НЬЮТОНА)"),
		widget.NewLabel("Черная линия — аналитическое решение"),
		widget.NewLabel("Синяя линия — численное решение"),
		widget.NewSeparator(),
		plotNewton,
	)
	tabs = append(tabs, container.NewTabItem("График (Ньютон)", container.NewVScroll(plotNewtonContent)))

	// График погрешности (Ньютон)
	errorPlotNewton := createErrorPlot(pointsNewton, 1100, 600)
	errorPlotNewtonContent := container.NewVBox(
		widget.NewLabel("ПОГРЕШНОСТЬ (МЕТОД НЬЮТОНА)"),
		widget.NewLabel("График показывает абсолютную погрешность |y_численное - y_аналитическое|"),
		widget.NewSeparator(),
		errorPlotNewton,
	)
	tabs = append(tabs, container.NewTabItem("Погрешность (Ньютон)", container.NewVScroll(errorPlotNewtonContent)))

	tabContainer := container.NewAppTabs(tabs...)

	myWindow.SetContent(tabContainer)
	myWindow.ShowAndRun()
}
