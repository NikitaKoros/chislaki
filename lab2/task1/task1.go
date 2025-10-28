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

// График функции с сеткой и стрелками
func createPlot() *fyne.Container {
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

	centerX := marginLeft + plotWidth/2
	centerY := marginTop + plotHeight/2

	xToPixel := func(x float64) float32 {
		return float32(centerX + x*scale)
	}
	yToPixel := func(y float64) float32 {
		return float32(centerY - y*scale)
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

	xArrow1 := canvas.NewLine(color.Black)
	xArrow1.Position1 = fyne.NewPos(xToPixel(xMax), yToPixel(0))
	xArrow1.Position2 = fyne.NewPos(xToPixel(xMax)-8, yToPixel(0)-4)
	xArrow1.StrokeWidth = 2
	objects = append(objects, xArrow1)

	xArrow2 := canvas.NewLine(color.Black)
	xArrow2.Position1 = fyne.NewPos(xToPixel(xMax), yToPixel(0))
	xArrow2.Position2 = fyne.NewPos(xToPixel(xMax)-8, yToPixel(0)+4)
	xArrow2.StrokeWidth = 2
	objects = append(objects, xArrow2)

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	yArrow1 := canvas.NewLine(color.Black)
	yArrow1.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yArrow1.Position2 = fyne.NewPos(xToPixel(0)-4, yToPixel(yMax)+8)
	yArrow1.StrokeWidth = 2
	objects = append(objects, yArrow1)

	yArrow2 := canvas.NewLine(color.Black)
	yArrow2.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yArrow2.Position2 = fyne.NewPos(xToPixel(0)+4, yToPixel(yMax)+8)
	yArrow2.StrokeWidth = 2
	objects = append(objects, yArrow2)

	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
		if x != 0 {
			tick := canvas.NewLine(color.Black)
			tick.Position1 = fyne.NewPos(xToPixel(x), yToPixel(0)-5)
			tick.Position2 = fyne.NewPos(xToPixel(x), yToPixel(0)+5)
			tick.StrokeWidth = 1
			objects = append(objects, tick)

			label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(xToPixel(x)-5, yToPixel(0)+10))
			objects = append(objects, label)
		}
	}

	for y := math.Ceil(yMin); y <= math.Floor(yMax); y++ {
		if y != 0 {
			tick := canvas.NewLine(color.Black)
			tick.Position1 = fyne.NewPos(xToPixel(0)-5, yToPixel(y))
			tick.Position2 = fyne.NewPos(xToPixel(0)+5, yToPixel(y))
			tick.StrokeWidth = 1
			objects = append(objects, tick)

			label := canvas.NewText(fmt.Sprintf("%.0f", y), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(xToPixel(0)-30, yToPixel(y)-5))
			objects = append(objects, label)
		}
	}

	xLabel := canvas.NewText("x", color.Black)
	xLabel.TextSize = 14
	xLabel.Move(fyne.NewPos(xToPixel(xMax)+5, yToPixel(0)-10))
	objects = append(objects, xLabel)

	yLabel := canvas.NewText("f(x)", color.Black)
	yLabel.TextSize = 14
	yLabel.Move(fyne.NewPos(xToPixel(0)+5, yToPixel(yMax)-5))
	objects = append(objects, yLabel)

	zeroLabel := canvas.NewText("0", color.Black)
	zeroLabel.TextSize = 10
	zeroLabel.Move(fyne.NewPos(xToPixel(0)-15, yToPixel(0)+10))
	objects = append(objects, zeroLabel)

	steps := 1000
	dx := (xMax - xMin) / float64(steps)

	for i := 0; i < steps-1; i++ {
		x1 := xMin + float64(i)*dx
		x2 := xMin + float64(i+1)*dx
		y1 := f(x1)
		y2 := f(x2)

		if !math.IsNaN(y1) && !math.IsNaN(y2) {
			if (y1 >= yMin || y2 >= yMin) && (y1 <= yMax || y2 <= yMax) {
				clipY1 := math.Max(yMin, math.Min(yMax, y1))
				clipY2 := math.Max(yMin, math.Min(yMax, y2))

				line := canvas.NewLine(color.RGBA{0, 0, 255, 255})
				line.Position1 = fyne.NewPos(xToPixel(x1), yToPixel(clipY1))
				line.Position2 = fyne.NewPos(xToPixel(x2), yToPixel(clipY2))
				line.StrokeWidth = 2
				objects = append(objects, line)
			}
		}
	}

	return container.NewWithoutLayout(objects...)
}

////////////////////////////////////////////////////////////////////////////////////

// Функция f(x) = x * log10(x + 3) + (x + 1)^2 - 5
func f(x float64) float64 {
	if x+3 <= 0 {
		return math.NaN()
	}
	return x*math.Log10(x+3) + math.Pow(x+1, 2) - 5
}

// Производная f'(x)
func df(x float64) float64 {
	if x+3 <= 0 {
		return math.NaN()
	}
	return math.Log10(x+3) + x/(math.Ln10*(x+3)) + 2*(x+1)
}

// Метод простой итерации: x = φ(x)
func phi(x float64) float64 {
	if x+3 <= 0 {
		return math.NaN()
	}
	val := 5 - x*math.Log10(x+3)
	if val < 0 {
		return math.NaN()
	}
	return -1 + math.Sqrt(val)
}

func dphi(x float64) float64 {
	if x+3 <= 0 {
		return math.NaN()
	}
	val := 5 - x*math.Log10(x+3)
	if val <= 0 {
		return math.NaN()
	}
	return (-math.Log10(x+3) - x/(math.Ln10*(x+3))) / (2 * math.Sqrt(val))
}

type MethodResult struct {
	Root       float64
	Iterations int
	Converged  bool
	Message    string
}

// Вторая производная для условий сходимости
func ddf(x float64) float64 {
	if x+3 <= 0 {
		return math.NaN()
	}
	// f(x) = x*log10(x+3) + (x+1)^2 - 5
	// f'(x) = log10(x+3) + x/(ln(10)*(x+3)) + 2*(x+1)
	// f''(x) = 1/(ln(10)*(x+3)) - x/(ln(10)*(x+3)^2) + 2
	return 1/(math.Ln10*(x+3)) - x/(math.Ln10*math.Pow(x+3, 2)) + 2
}

// Метод простой итерации
func simpleIterationInterval(a, b, eps float64, maxIter int) MethodResult {
	fa, fb := f(a), f(b)

	// Проверка смены знака
	if fa*fb >= 0 {
		return MethodResult{0, 0, false, fmt.Sprintf("❌ f(a)*f(b)=%.6f >= 0 — функция не меняет знак на [a,b]", fa*fb)}
	}

	// Проверка монотонности
	dfa, dfb := df(a), df(b)
	msg := ""
	if dfa*dfb > 0 {
		msg += "f'(a)*f'(b)>0 — функция монотонна\n"
	} else {
		msg += "f'(a)*f'(b)<=0 — функция немонотонна\n"
	}

	// Проверка знакопостоянства второй производной
	d2fa, d2fb := ddf(a), ddf(b)
	if d2fa*d2fb > 0 {
		msg += "f''(a)*f''(b)>0 — функция сохраняет выпуклость/вогнутость\n"
	} else {
		msg += "f''(a)*f''(b)<=0 — вторая производная меняет знак\n"
	}

	// Выбор x0 по условию f(x0)*f''(x0) > 0
	var x0 float64
	condA := fa * d2fa
	condB := fb * d2fb
	if condA > 0 && condB > 0 {
		if math.Abs(fa) > math.Abs(fb) {
			x0 = a
		} else {
			x0 = b
		}
		msg += fmt.Sprintf("Оба конца удовлетворяют f(x)*f''(x)>0. Выбрана x₀=%.6f\n", x0)
	} else if condA > 0 {
		x0 = a
		msg += fmt.Sprintf("f(a)*f''(a)>0. Выбрана x₀=a=%.6f\n", x0)
	} else if condB > 0 {
		x0 = b
		msg += fmt.Sprintf("f(b)*f''(b)>0. Выбрана x₀=b=%.6f\n", x0)
	} else {
		x0 = (a + b) / 2
		msg += fmt.Sprintf("Ни один конец не подходит. Выбрана середина x₀=%.6f\n", x0)
	}

	// Проверка условия сходимости |φ'(x0)| < 1
	derivative := math.Abs(dphi(x0))
	if derivative >= 1 || math.IsNaN(derivative) {
		msg += fmt.Sprintf("❌ |φ'(x₀)|=%.6f >= 1 — метод может не сходиться", derivative)
		return MethodResult{0, 0, false, msg}
	}

	// Итерации
	x := x0
	for i := 0; i < maxIter; i++ {
		xNew := phi(x)
		if math.IsNaN(xNew) {
			return MethodResult{0, i, false, msg + fmt.Sprintf("φ(x) неопределена при x=%.6f", x)}
		}
		if math.Abs(xNew-x) < eps {
			msg += fmt.Sprintf("|φ'(x)|=%.6f<1 — метод сошёлся", math.Abs(dphi(xNew)))
			return MethodResult{xNew, i + 1, true, msg}
		}
		x = xNew
	}
	return MethodResult{x, maxIter, false, msg + "Не сошёлся за maxIter"}
}

// Метод Ньютона
func newtonInterval(a, b, eps float64, maxIter int) MethodResult {
	fa, fb := f(a), f(b)

	// 1. Проверка смены знака
	if fa*fb >= 0 {
		return MethodResult{0, 0, false, fmt.Sprintf("❌ f(a)*f(b)=%.6f >= 0 — функция не меняет знак на [a,b]", fa*fb)}
	}

	// 2. Монотонность
	dfa, dfb := df(a), df(b)
	msg := ""
	if dfa*dfb > 0 {
		msg += "f'(a)*f'(b)>0 — функция монотонна\n"
	} else {
		msg += "f'(a)*f'(b)<=0 — производная меняет знак\n"
	}

	// 3. Проверка второй производной
	d2fa, d2fb := ddf(a), ddf(b)
	if d2fa*d2fb > 0 {
		msg += "f''(a)*f''(b)>0 — функция сохраняет выпуклость/вогнутость\n"
	} else {
		msg += "f''(a)*f''(b)<=0 — вторая производная меняет знак\n"
	}

	// 4. Выбор x0
	var x0 float64
	condA := fa * d2fa
	condB := fb * d2fb
	if condA > 0 && condB > 0 {
		if math.Abs(fa) > math.Abs(fb) {
			x0 = a
		} else {
			x0 = b
		}
		msg += fmt.Sprintf("Оба конца удовлетворяют f(x)*f''(x)>0. Выбрана x₀=%.6f\n", x0)
	} else if condA > 0 {
		x0 = a
		msg += fmt.Sprintf("f(a)*f''(a)>0. Выбрана x₀=a=%.6f\n", x0)
	} else if condB > 0 {
		x0 = b
		msg += fmt.Sprintf("f(b)*f''(b)>0. Выбрана x₀=b=%.6f\n", x0)
	} else {
		x0 = (a + b) / 2
		msg += fmt.Sprintf("Ни один конец не подходит. Выбрана середина x₀=%.6f\n", x0)
	}

	// Итерации Ньютона
	x := x0
	for i := 0; i < maxIter; i++ {
		fx := f(x)
		dfx := df(x)
		if math.Abs(dfx) < 1e-10 {
			return MethodResult{x, i, false, msg + "f'(x)≈0 — деление невозможно"}
		}
		xNew := x - fx/dfx
		if math.Abs(xNew-x) < eps {
			msg += fmt.Sprintf("|f'(x)|=%.6f != 0 — метод сошёлся", math.Abs(df(xNew)))
			return MethodResult{xNew, i + 1, true, msg}
		}
		x = xNew
	}
	return MethodResult{x, maxIter, false, msg + "Не сошёлся за maxIter"}
}

// Метод секущих
func secant(x0, x1, eps float64, maxIter int) MethodResult {
	for i := 0; i < maxIter; i++ {
		fx0, fx1 := f(x0), f(x1)
		if math.Abs(fx1-fx0) < 1e-10 {
			return MethodResult{Root: x1, Iterations: i, Converged: false, Message: "f(x₁) ≈ f(x₀)"}
		}
		xNew := x1 - fx1*(x1-x0)/(fx1-fx0)
		if math.Abs(xNew-x1) < eps {
			return MethodResult{
				Root:       xNew,
				Iterations: i + 1,
				Converged:  true,
				Message:    fmt.Sprintf("|x₁ - x₀| = %.6f < ε", math.Abs(xNew-x1)),
			}
		}
		x0, x1 = x1, xNew
	}
	return MethodResult{Root: x1, Iterations: maxIter, Converged: false, Message: "Не сошелся"}
}

// Метод хорд
func chord(a, b, eps float64, maxIter int) MethodResult {
	fa, fb := f(a), f(b)
	if fa*fb >= 0 {
		return MethodResult{Root: 0, Iterations: 0, Converged: false,
			Message: fmt.Sprintf("f(a)*f(b) = %.6f >= 0 — нет гарантии наличия корня на [a,b]", fa*fb),
		}
	}

	f2a := ddf(a)
	if fa*f2a <= 0 {
		return MethodResult{
			Root:       0,
			Iterations: 0,
			Converged:  false,
			Message:    fmt.Sprintf("Условие сходимости: f(a)*f''(a) = %.6f <= 0 — метод хорд может не сойтись", fa*f2a),
		}
	}

	x := a
	for i := 0; i < maxIter; i++ {
		fx := f(x)
		xNew := x - fx*(x-b)/(fx-fb)
		if math.Abs(xNew-x) < eps {
			return MethodResult{
				Root:       xNew,
				Iterations: i + 1,
				Converged:  true,
				Message:    fmt.Sprintf("Условие сходимости: f(a)*f''(a) = %.6f > 0 — метод хорд сходится", fa*f2a),
			}
		}
		x = xNew
	}
	return MethodResult{Root: x, Iterations: maxIter, Converged: false, Message: "Не сошелся за maxIter"}
}

// Метод дихотомии
func bisection(a, b, eps float64, maxIter int) MethodResult {
	fa, fb := f(a), f(b)

	if math.IsNaN(fa) || math.IsNaN(fb) || math.IsInf(fa, 0) || math.IsInf(fb, 0) {
		return MethodResult{
			Root:       0,
			Iterations: 0,
			Converged:  false,
			Message:    fmt.Sprintf("Значения функции на концах должны быть конечными: f(a)=%.6f, f(b)=%.6f", fa, fb),
		}
	}

	if fa*fb >= 0 {
		return MethodResult{
			Root:       0,
			Iterations: 0,
			Converged:  false,
			Message:    fmt.Sprintf("f(a)*f(b) = %.6f >= 0 — нет гарантии корня на [a,b]", fa*fb),
		}
	}

	// Дихотомия всегда сходится, если f непрерывна и f(a)*f(b)<0
	for i := 0; i < maxIter; i++ {
		c := (a + b) / 2
		fc := f(c)
		if (b - a) < eps {
			return MethodResult{
				Root:       c,
				Iterations: i + 1,
				Converged:  true,
				Message:    fmt.Sprintf("Условие сходимости: f(a)*f(b) = %.6f < 0 — метод дихотомии гарантированно сходится", fa*fb),
			}
		}
		if fa*fc < 0 {
			b, fb = c, fc
		} else {
			a, fa = c, fc
		}
	}
	return MethodResult{Root: (a + b) / 2, Iterations: maxIter, Converged: false, Message: "Не сошелся за maxIter"}
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Численные методы решения уравнений")
	myWindow.Resize(fyne.NewSize(900, 700))

	plot := createPlot()
	plotContainer := container.NewVBox(
		widget.NewLabel("График функции f(x) = x·log₁₀(x+3) + (x+1)² - 5"),
		widget.NewLabel("Найдите пересечения с осью X для определения начальных приближений"),
		plot,
	)

	epsEntry := widget.NewEntry()
	epsEntry.SetText("0.0001")
	epsLabel := widget.NewLabel("Точность ε:")
	epsContainer := container.NewBorder(nil, nil, epsLabel, nil, epsEntry)

	var resultLabels []*widget.Label
	clearBtn := widget.NewButton("Очистить результаты", func() {
		for _, label := range resultLabels {
			label.SetText("")
		}
	})

	// Метод простой итерации
	aEntry1 := widget.NewEntry()
	aEntry1.SetText("-3.0")
	bEntry1 := widget.NewEntry()
	bEntry1.SetText("3.0")
	result1 := widget.NewLabel("")
	resultLabels = append(resultLabels, result1)

	solve1Btn := widget.NewButton("Найти решение", func() {
		eps, _ := strconv.ParseFloat(epsEntry.Text, 64)
		a, _ := strconv.ParseFloat(aEntry1.Text, 64)
		b, _ := strconv.ParseFloat(bEntry1.Text, 64)

		if a >= b {
			result1.SetText("Ошибка: a должно быть меньше b")
			return
		}

		res := simpleIterationInterval(a, b, eps, 1000)
		if res.Root < a || res.Root > b {
			result1.SetText(fmt.Sprintf("Корень вне заданных границ: x = %.6f", res.Root))
			return
		}
		result1.SetText(fmt.Sprintf("Корень: x = %.6f\nИтераций: %d\n%s", res.Root, res.Iterations, res.Message))
	})

	method1 := container.NewVBox(
		widget.NewLabel("Метод простой итерации"),
		container.NewHBox(widget.NewLabel("a:"), aEntry1),
		container.NewHBox(widget.NewLabel("b:"), bEntry1),
		solve1Btn,
		result1,
	)

	// Метод Ньютона
	aEntry2 := widget.NewEntry()
	aEntry2.SetText("-3.0")
	bEntry2 := widget.NewEntry()
	bEntry2.SetText("3.0")
	result2 := widget.NewLabel("")
	resultLabels = append(resultLabels, result2)

	solve2Btn := widget.NewButton("Найти решение", func() {
		eps, _ := strconv.ParseFloat(epsEntry.Text, 64)
		a, _ := strconv.ParseFloat(aEntry2.Text, 64)
		b, _ := strconv.ParseFloat(bEntry2.Text, 64)

		if a >= b {
			result2.SetText("Ошибка: a должно быть меньше b")
			return
		}

		res := newtonInterval(a, b, eps, 1000)
		if res.Root < a || res.Root > b {
			result2.SetText(fmt.Sprintf("Корень вне заданных границ: x = %.6f", res.Root))
			return
		}
		result2.SetText(fmt.Sprintf("Корень: x = %.6f\nИтераций: %d\n%s", res.Root, res.Iterations, res.Message))
	})

	method2 := container.NewVBox(
		widget.NewLabel("Метод Ньютона"),
		container.NewHBox(widget.NewLabel("a:"), aEntry2),
		container.NewHBox(widget.NewLabel("b:"), bEntry2),
		solve2Btn,
		result2,
	)

	// Метод секущих
	x0Entry3 := widget.NewEntry()
	x0Entry3.SetText("1.0")
	x1Entry3 := widget.NewEntry()
	x1Entry3.SetText("1.5")
	result3 := widget.NewLabel("")
	resultLabels = append(resultLabels, result3)

	solve3Btn := widget.NewButton("Найти решение", func() {
		eps, _ := strconv.ParseFloat(epsEntry.Text, 64)
		x0, _ := strconv.ParseFloat(x0Entry3.Text, 64)
		x1, _ := strconv.ParseFloat(x1Entry3.Text, 64)
		res := secant(x0, x1, eps, 1000)
		result3.SetText(fmt.Sprintf("Корень: x = %.6f\nИтераций: %d\n%s", res.Root, res.Iterations, res.Message))
	})

	method3 := container.NewVBox(
		widget.NewLabel("Метод секущих"),
		container.NewHBox(widget.NewLabel("x₀:"), x0Entry3),
		container.NewHBox(widget.NewLabel("x₁:"), x1Entry3),
		solve3Btn,
		result3,
	)

	// Метод хорд
	aEntry4 := widget.NewEntry()
	aEntry4.SetText("-2.9")
	bEntry4 := widget.NewEntry()
	bEntry4.SetText("-2.0")
	result4 := widget.NewLabel("")
	resultLabels = append(resultLabels, result4)

	solve4Btn := widget.NewButton("Найти решение", func() {
		eps, _ := strconv.ParseFloat(epsEntry.Text, 64)
		a, _ := strconv.ParseFloat(aEntry4.Text, 64)
		b, _ := strconv.ParseFloat(bEntry4.Text, 64)
		res := chord(a, b, eps, 1000)
		result4.SetText(fmt.Sprintf("Корень: x = %.6f\nИтераций: %d\n%s", res.Root, res.Iterations, res.Message))
	})

	method4 := container.NewVBox(
		widget.NewLabel("Метод хорд"),
		container.NewHBox(widget.NewLabel("a:"), aEntry4),
		container.NewHBox(widget.NewLabel("b:"), bEntry4),
		solve4Btn,
		result4,
	)

	// Метод дихотомии
	aEntry5 := widget.NewEntry()
	aEntry5.SetText("-2.9")
	bEntry5 := widget.NewEntry()
	bEntry5.SetText("-1.0")
	result5 := widget.NewLabel("")
	resultLabels = append(resultLabels, result5)

	solve5Btn := widget.NewButton("Найти решение", func() {
		eps, _ := strconv.ParseFloat(epsEntry.Text, 64)
		a, _ := strconv.ParseFloat(aEntry5.Text, 64)
		b, _ := strconv.ParseFloat(bEntry5.Text, 64)
		res := bisection(a, b, eps, 1000)
		result5.SetText(fmt.Sprintf("Корень: x = %.6f\nИтераций: %d\n%s", res.Root, res.Iterations, res.Message))
	})

	method5 := container.NewVBox(
		widget.NewLabel("Метод дихотомии"),
		container.NewHBox(widget.NewLabel("a:"), aEntry5),
		container.NewHBox(widget.NewLabel("b:"), bEntry5),
		solve5Btn,
		result5,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("График", plotContainer),
		container.NewTabItem("Методы", container.NewVScroll(container.NewVBox(
			epsContainer,
			clearBtn,
			widget.NewSeparator(),
			container.NewGridWithColumns(2,
				method1, method2,
				method3, method4,
			),
			method5,
		))),
	)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
