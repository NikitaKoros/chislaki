package main

import (
	"fmt"
	"image/color"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Система уравнений:
// F1: x² - 7xy + 2y² + 5x - 9y - 7 = 0
// F2: 5x² - 4xy - 3y² - 2x + 8y + 6 = 0

func F1(x, y float64) float64 {
	return x*x - 7*x*y + 2*y*y + 5*x - 9*y - 7
}

func F2(x, y float64) float64 {
	return 5*x*x - 4*x*y - 3*y*y - 2*x + 8*y + 6
}

// Частные производные для матрицы Якоби
func dF1_dx(x, y float64) float64 {
	return 2*x - 7*y + 5
}

func dF1_dy(x, y float64) float64 {
	return -7*x + 4*y - 9
}

func dF2_dx(x, y float64) float64 {
	return 10*x - 4*y - 2
}

func dF2_dy(x, y float64) float64 {
	return -4*x - 6*y + 8
}

// Матрица Якоби 2x2
type Matrix2x2 [2][2]float64
type Vector2 [2]float64

func jacobian(x, y float64) Matrix2x2 {
	return Matrix2x2{
		{dF1_dx(x, y), dF1_dy(x, y)},
		{dF2_dx(x, y), dF2_dy(x, y)},
	}
}

func det2x2(m Matrix2x2) float64 {
	return m[0][0]*m[1][1] - m[0][1]*m[1][0]
}

func inverse2x2(m Matrix2x2) (Matrix2x2, bool) {
	d := det2x2(m)
	if math.Abs(d) < 1e-10 {
		return Matrix2x2{}, false
	}
	return Matrix2x2{
		{m[1][1] / d, -m[0][1] / d},
		{-m[1][0] / d, m[0][0] / d},
	}, true
}

func mulMatVec(m Matrix2x2, v Vector2) Vector2 {
	return Vector2{
		m[0][0]*v[0] + m[0][1]*v[1],
		m[1][0]*v[0] + m[1][1]*v[1],
	}
}

func mulMatMat(a, b Matrix2x2) Matrix2x2 {
	return Matrix2x2{
		{a[0][0]*b[0][0] + a[0][1]*b[1][0], a[0][0]*b[0][1] + a[0][1]*b[1][1]},
		{a[1][0]*b[0][0] + a[1][1]*b[1][0], a[1][0]*b[0][1] + a[1][1]*b[1][1]},
	}
}

func vecNorm(v Vector2) float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1])
}

type SystemResult struct {
	X          float64
	Y          float64
	Iterations int
	Converged  bool
	Message    string
}

// Модифицированный метод Ньютона (Якобиан вычисляется однажды в x0)
func newtonSystem(x0, y0, eps float64, maxIter int) SystemResult {
	x, y := x0, y0

	// Вычисляем матрицу Якоби только один раз в начальной точке
	J := jacobian(x0, y0)
	Jinv, ok := inverse2x2(J)
	if !ok {
		return SystemResult{x0, y0, 0, false, "❌ Матрица Якоби вырожденная"}
	}

	for i := 0; i < maxIter; i++ {
		f1, f2 := F1(x, y), F2(x, y)
		//F_vector := Vector2{f1, f2}
		//F_norm := vecNorm(F_vector)

		// Вычисляем приращение: Δx = -J⁻¹ * F
		delta := mulMatVec(Jinv, Vector2{-f1, -f2})
		error := vecNorm(delta)

		x += delta[0]
		y += delta[1]

		if error < eps {
			return SystemResult{x, y, i + 1, true, "✅ Метод сходится - условие |Δx| < ε выполнено"}
		}
	}
	return SystemResult{x, y, maxIter, false, "⚠️ Достигнуто максимальное количество итераций"}
}

var lambda = 0.05

// Итерационные функции φ
func phi1(x, y float64) float64 {
	// Из F1 = 0: 2y² - (7x+9)y + (x²+5x-7) = 0
	// Решаем относительно y и выражаем x через итерацию
	// Используем x = x - λ*F1(x,y)
	return x - lambda*F1(x, y)
}

func phi2(x, y float64) float64 {
	// y = y - λ*F2(x,y)
	return y - lambda*F2(x, y)
}

// Частные производные φ
func dphi1_dx(x, y float64) float64 {
	return 1 - lambda*dF1_dx(x, y)
}

func dphi1_dy(x, y float64) float64 {
	return -lambda * dF1_dy(x, y)
}

func dphi2_dx(x, y float64) float64 {
	return -lambda * dF2_dx(x, y)
}

func dphi2_dy(x, y float64) float64 {
	return 1 - lambda*dF2_dy(x, y)
}

// Метод простой итерации
func simpleIterationSystem(x0, y0, eps float64, maxIter int) SystemResult {
	x, y := x0, y0
	k := 0
	converged := false
	msg := ""

	// Шаг 1: Оценка условия сходимости в окрестности начальной точки
	supNorm := 0.0
	samples := 5
	radius := 0.2

	for i := 0; i < samples; i++ {
		for j := 0; j < samples; j++ {
			xi := x0 + (-radius + (2*radius)*float64(i)/float64(samples-1))
			yi := y0 + (-radius + (2*radius)*float64(j)/float64(samples-1))

			// Вычисляем матрицу Якоби в точке (xi, yi)
			J := Matrix2x2{
				{dphi1_dx(xi, yi), dphi1_dy(xi, yi)},
				{dphi2_dx(xi, yi), dphi2_dy(xi, yi)},
			}

			// Вычисляем норму матрицы (супремум по строкам)
			row1Norm := math.Abs(J[0][0]) + math.Abs(J[0][1])
			row2Norm := math.Abs(J[1][0]) + math.Abs(J[1][1])
			maxRowNorm := math.Max(row1Norm, row2Norm)

			if maxRowNorm > supNorm {
				supNorm = maxRowNorm
			}
		}
	}

	// Формируем сообщение о сходимости
	if supNorm < 1.0 {
		msg = fmt.Sprintf("✅ sup||Φ'|| ~= %.3f < 1 — условие сходимости выполнено", supNorm)
	} else {
		msg = fmt.Sprintf("⚠️ sup||Φ'|| ~= %.3f >= 1 — сходимость не гарантирована", supNorm)
	}

	// Шаг 2: Итерационный процесс
	for k < maxIter {
		// Вычисляем следующее приближение
		xNew := phi1(x, y)
		yNew := phi2(x, y)

		// Проверяем сходимость по разнице между итерациями
		errorX := math.Abs(xNew - x)
		errorY := math.Abs(yNew - y)
		maxError := math.Max(errorX, errorY)

		if maxError <= eps {
			converged = true
			k++
			msg += fmt.Sprintf("\n✅ Решение найдено за %d итераций", k)
			break
		}

		// Переход к следующей итерации
		x, y = xNew, yNew
		k++
	}

	if !converged {
		msg += fmt.Sprintf("\n⚠️ Не сошёлся за %d итераций", maxIter)
	}

	return SystemResult{
		X:          x,
		Y:          y,
		Iterations: k,
		Converged:  converged,
		Message:    msg,
	}
}

// Метод Зейделя
func seidelSystem(x0, y0, eps float64, maxIter int) SystemResult {
	x, y := x0, y0
	k := 0
	converged := false

	J0 := jacobian(x0, y0)
	B, ok := inverse2x2(J0)
	if !ok {
		return SystemResult{x0, y0, 0, false, "❌ J(x0) вырожденная"}
	}

	omega := 0.05

	// Проверка условия сходимости
	supNorm := 0.0
	samples := 5
	radius := 0.2

	for i := 0; i < samples; i++ {
		for j := 0; j < samples; j++ {
			xi := x0 + (-radius + (2*radius)*float64(i)/float64(samples-1))
			yi := y0 + (-radius + (2*radius)*float64(j)/float64(samples-1))

			Jx := jacobian(xi, yi)
			BJ := mulMatMat(B, Jx)
			J := Matrix2x2{
				{1 - omega*BJ[0][0], -omega * BJ[0][1]},
				{-omega * BJ[1][0], 1 - omega*BJ[1][1]},
			}

			row1Norm := math.Abs(J[0][0]) + math.Abs(J[0][1])
			row2Norm := math.Abs(J[1][0]) + math.Abs(J[1][1])
			maxRowNorm := math.Max(row1Norm, row2Norm)

			if maxRowNorm > supNorm {
				supNorm = maxRowNorm
			}
		}
	}

	msg := ""
	if supNorm < 1.0 {
		msg = fmt.Sprintf("✅ sup||Φ'|| ~= %.3f < 1 — условие сходимости выполнено", supNorm)
	} else {
		msg = fmt.Sprintf("⚠️ sup||Φ'|| ~= %.3f >= 1 — сходимость не гарантирована", supNorm)
	}

	// Итерационный процесс
	for k < maxIter {
		// Обновляем x
		Fv := Vector2{F1(x, y), F2(x, y)}
		B_F := mulMatVec(B, Fv)
		xNew := x - omega*B_F[0]

		// Обновляем y с новым x
		Fv = Vector2{F1(xNew, y), F2(xNew, y)}
		B_F = mulMatVec(B, Fv)
		yNew := y - omega*B_F[1]

		// Проверяем сходимость по разнице между итерациями
		errorX := math.Abs(xNew - x)
		errorY := math.Abs(yNew - y)
		maxError := math.Max(errorX, errorY)

		if maxError <= eps {
			converged = true
			k++
			msg += fmt.Sprintf("\n✅ Решение найдено за %d итераций", k)
			break
		}

		// Переход к следующей итерации
		x, y = xNew, yNew
		k++
	}

	if !converged {
		msg += fmt.Sprintf("\n⚠️ Не сошёлся за %d итераций", maxIter)
	}

	return SystemResult{
		X:          x,
		Y:          y,
		Iterations: k,
		Converged:  converged,
		Message:    msg,
	}
}

// График для визуализации пересечений
func createSystemPlot() *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(700, 500))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	xMin, xMax := -6.0, 6.0
	yMin, yMax := -6.0, 6.0
	marginLeft, marginRight := 80.0, 80.0
	marginTop, marginBottom := 80.0, 80.0
	plotWidth := 700.0 - marginLeft - marginRight
	plotHeight := 500.0 - marginTop - marginBottom

	scale := math.Min(plotWidth/(xMax-xMin), plotHeight/(yMax-yMin))
	scaleX := scale
	scaleY := scale

	centerX := marginLeft + plotWidth/2
	centerY := marginTop + plotHeight/2

	xToPixel := func(x float64) float32 {
		return float32(centerX + x*scaleX)
	}
	yToPixel := func(y float64) float32 {
		return float32(centerY - y*scaleY)
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

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	// Метки осей
	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
		if x != 0 {
			label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(xToPixel(x)-5, yToPixel(0)+10))
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

	// Рисуем кривые F1=0 и F2=0
	steps := 1000
	dx := (xMax - xMin) / float64(steps)

	// F1 = 0 (синий)
	for i := 0; i < steps-1; i++ {
		x1 := xMin + float64(i)*dx
		x2 := xMin + float64(i+1)*dx

		// Решаем квадратное уравнение 2y² - (7x+9)y + (x²+5x-7) = 0
		a := 2.0
		b1 := -(7*x1 + 9)
		c1 := x1*x1 + 5*x1 - 7
		disc1 := b1*b1 - 4*a*c1

		b2 := -(7*x2 + 9)
		c2 := x2*x2 + 5*x2 - 7
		disc2 := b2*b2 - 4*a*c2

		if disc1 >= 0 && disc2 >= 0 {
			y1_1 := (-b1 + math.Sqrt(disc1)) / (2 * a)
			y2_1 := (-b2 + math.Sqrt(disc2)) / (2 * a)
			if y1_1 >= yMin && y1_1 <= yMax && y2_1 >= yMin && y2_1 <= yMax {
				line := canvas.NewLine(color.RGBA{0, 0, 255, 255})
				line.Position1 = fyne.NewPos(xToPixel(x1), yToPixel(y1_1))
				line.Position2 = fyne.NewPos(xToPixel(x2), yToPixel(y2_1))
				line.StrokeWidth = 2
				objects = append(objects, line)
			}

			y1_2 := (-b1 - math.Sqrt(disc1)) / (2 * a)
			y2_2 := (-b2 - math.Sqrt(disc2)) / (2 * a)
			if y1_2 >= yMin && y1_2 <= yMax && y2_2 >= yMin && y2_2 <= yMax {
				line := canvas.NewLine(color.RGBA{0, 0, 255, 255})
				line.Position1 = fyne.NewPos(xToPixel(x1), yToPixel(y1_2))
				line.Position2 = fyne.NewPos(xToPixel(x2), yToPixel(y2_2))
				line.StrokeWidth = 2
				objects = append(objects, line)
			}
		}
	}

	// F2 = 0 (красный)
	for i := 0; i < steps-1; i++ {
		x1 := xMin + float64(i)*dx
		x2 := xMin + float64(i+1)*dx

		// Решаем квадратное уравнение -3y² + (8-4x)y + (5x²-2x+6) = 0
		a := -3.0
		b1 := 8 - 4*x1
		c1 := 5*x1*x1 - 2*x1 + 6
		disc1 := b1*b1 - 4*a*c1

		b2 := 8 - 4*x2
		c2 := 5*x2*x2 - 2*x2 + 6
		disc2 := b2*b2 - 4*a*c2

		if disc1 >= 0 && disc2 >= 0 {
			y1_1 := (-b1 + math.Sqrt(disc1)) / (2 * a)
			y2_1 := (-b2 + math.Sqrt(disc2)) / (2 * a)
			if y1_1 >= yMin && y1_1 <= yMax && y2_1 >= yMin && y2_1 <= yMax {
				line := canvas.NewLine(color.RGBA{255, 0, 0, 255})
				line.Position1 = fyne.NewPos(xToPixel(x1), yToPixel(y1_1))
				line.Position2 = fyne.NewPos(xToPixel(x2), yToPixel(y2_1))
				line.StrokeWidth = 2
				objects = append(objects, line)
			}

			y1_2 := (-b1 - math.Sqrt(disc1)) / (2 * a)
			y2_2 := (-b2 - math.Sqrt(disc2)) / (2 * a)
			if y1_2 >= yMin && y1_2 <= yMax && y2_2 >= yMin && y2_2 <= yMax {
				line := canvas.NewLine(color.RGBA{255, 0, 0, 255})
				line.Position1 = fyne.NewPos(xToPixel(x1), yToPixel(y1_2))
				line.Position2 = fyne.NewPos(xToPixel(x2), yToPixel(y2_2))
				line.StrokeWidth = 2
				objects = append(objects, line)
			}
		}
	}

	return container.NewWithoutLayout(objects...)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Система нелинейных уравнений")
	myWindow.Resize(fyne.NewSize(900, 700))

	plot := createSystemPlot()
	plotContainer := container.NewVBox(
		widget.NewLabel("Графики: F₁=0 (синий), F₂=0 (красный)"),
		widget.NewLabel("F₁: x² - 7xy + 2y² + 5x - 9y - 7 = 0"),
		widget.NewLabel("F₂: 5x² - 4xy - 3y² - 2x + 8y + 6 = 0"),
		widget.NewLabel("Точки пересечения — решения системы"),
		plot,
	)

	epsEntry := widget.NewEntry()
	epsEntry.SetText("0.0001")
	epsEntry.Resize(fyne.NewSize(200, epsEntry.MinSize().Height))

	var resultLabels []*widget.Label
	clearBtn := widget.NewButton("Очистить результаты", func() {
		for _, label := range resultLabels {
			label.SetText("")
		}
	})

	// Метод Ньютона
	x0Entry1 := widget.NewEntry()
	x0Entry1.SetText("-1.0")
	y0Entry1 := widget.NewEntry()
	y0Entry1.SetText("-1.0")
	result1 := widget.NewLabel("")
	resultLabels = append(resultLabels, result1)

	solve1Btn := widget.NewButton("Найти решение", func() {
		eps, _ := strconv.ParseFloat(epsEntry.Text, 64)
		x0, _ := strconv.ParseFloat(x0Entry1.Text, 64)
		y0, _ := strconv.ParseFloat(y0Entry1.Text, 64)
		res := newtonSystem(x0, y0, eps, 1000)
		if !res.Converged {
			result1.SetText(fmt.Sprintf("❌ Не сошёлся\nx = %.6f, y = %.6f\nИтераций: %d\n%s", res.X, res.Y, res.Iterations, res.Message))
		} else {
			result1.SetText(fmt.Sprintf("x = %.6f, y = %.6f\nИтераций: %d\n%s", res.X, res.Y, res.Iterations, res.Message))
		}
	})

	method1 := container.NewVBox(
		widget.NewLabel("Метод Ньютона"),
		container.NewHBox(widget.NewLabel("x₀:"), x0Entry1),
		container.NewHBox(widget.NewLabel("y₀:"), y0Entry1),
		solve1Btn,
		result1,
	)

	// Метод простой итерации
	x0Entry2 := widget.NewEntry()
	x0Entry2.SetText("-1.0")
	y0Entry2 := widget.NewEntry()
	y0Entry2.SetText("-1.0")
	result2 := widget.NewLabel("")
	resultLabels = append(resultLabels, result2)

	solve2Btn := widget.NewButton("Найти решение", func() {
		eps, _ := strconv.ParseFloat(epsEntry.Text, 64)
		x0, _ := strconv.ParseFloat(x0Entry2.Text, 64)
		y0, _ := strconv.ParseFloat(y0Entry2.Text, 64)
		res := simpleIterationSystem(x0, y0, eps, 1000)
		if !res.Converged {
			result2.SetText(fmt.Sprintf("❌ Не сошёлся\nx = %.6f, y = %.6f\nИтераций: %d\n%s", res.X, res.Y, res.Iterations, res.Message))
		} else {
			result2.SetText(fmt.Sprintf("x = %.6f, y = %.6f\nИтераций: %d\n%s", res.X, res.Y, res.Iterations, res.Message))
		}
	})

	method2 := container.NewVBox(
		widget.NewLabel("Метод простой итерации"),
		container.NewHBox(widget.NewLabel("x₀:"), x0Entry2),
		container.NewHBox(widget.NewLabel("y₀:"), y0Entry2),
		solve2Btn,
		result2,
	)

	// Метод Зейделя
	x0Entry3 := widget.NewEntry()
	x0Entry3.SetText("-0.5")
	y0Entry3 := widget.NewEntry()
	y0Entry3.SetText("4.0")
	result3 := widget.NewLabel("")
	resultLabels = append(resultLabels, result3)

	solve3Btn := widget.NewButton("Найти решение", func() {
		eps, _ := strconv.ParseFloat(epsEntry.Text, 64)
		x0, _ := strconv.ParseFloat(x0Entry3.Text, 64)
		y0, _ := strconv.ParseFloat(y0Entry3.Text, 64)
		res := seidelSystem(x0, y0, eps, 1000)
		if !res.Converged {
			result3.SetText(fmt.Sprintf("❌ Не сошёлся\nx = %.6f, y = %.6f\nИтераций: %d\n%s", res.X, res.Y, res.Iterations, res.Message))
		} else {
			result3.SetText(fmt.Sprintf("x = %.6f, y = %.6f\nИтераций: %d\n%s", res.X, res.Y, res.Iterations, res.Message))
		}
	})

	method3 := container.NewVBox(
		widget.NewLabel("Метод Зейделя"),
		container.NewHBox(widget.NewLabel("x₀:"), x0Entry3),
		container.NewHBox(widget.NewLabel("y₀:"), y0Entry3),
		solve3Btn,
		result3,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("График", plotContainer),
		container.NewTabItem("Методы", container.NewVScroll(container.NewVBox(
			container.NewHBox(widget.NewLabel("Точность ε:"), epsEntry),
			clearBtn,
			widget.NewSeparator(),
			container.NewGridWithColumns(3, method1, method2, method3),
		))),
	)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
