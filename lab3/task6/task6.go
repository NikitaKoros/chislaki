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

// Подынтегральная функция: F = (2sin2x - cos(x/4))^2 / sqrt(2x^2 + 5x + 6)
func integrandF(x float64) float64 {
	numerator := math.Pow(2*math.Sin(2*x)-math.Cos(x/4), 2)
	denominator := math.Sqrt(2*math.Pow(x, 2) + 5*x + 6)
	return numerator / denominator
}

// Метод средних прямоугольников
func midpointMethod(fn func(float64) float64, a, b, h float64) float64 {
	n := int(math.Ceil((b - a) / h))
	h = (b - a) / float64(n)
	sum := 0.0

	for i := 0; i < n; i++ {
		xMid := a + (float64(i)+0.5)*h
		sum += fn(xMid)
	}

	return sum * h
}

// Метод трапеций
func trapezoidMethod(fn func(float64) float64, a, b, h float64) float64 {
	n := int(math.Ceil((b - a) / h))
	h = (b - a) / float64(n)
	sum := (fn(a) + fn(b)) / 2.0

	for i := 1; i < n; i++ {
		x := a + float64(i)*h
		sum += fn(x)
	}

	return sum * h
}

// Метод Симпсона
func simpsonMethod(fn func(float64) float64, a, b, h float64) float64 {
	n := int(math.Ceil((b - a) / h))

	// Для метода Симпсона нужно четное число интервалов
	if n%2 != 0 {
		n++
	}

	h = (b - a) / float64(n)
	sum := fn(a) + fn(b)

	for i := 1; i < n; i++ {
		x := a + float64(i)*h
		if i%2 == 0 {
			sum += 2 * fn(x)
		} else {
			sum += 4 * fn(x)
		}
	}

	return (sum * h) / 3.0
}

// Производная  (аналитическая)
func derivative(x float64) float64 {
	// Исх. функция: (2*sin(2x) - cos(x/4))^2 / sqrt(2x^2 + 5x + 6)

	// u(x) = 2*sin(2x) - cos(x/4)
	// v(x) = sqrt(2x^2 + 5x + 6)
	// f(x) = u(x)^2 / v(x)

	// u(x) и её производная
	u := 2*math.Sin(2*x) - math.Cos(x/4)
	uPrime := 4*math.Cos(2*x) + 0.25*math.Sin(x/4)

	// v(x) и её производная
	v := math.Sqrt(2*x*x + 5*x + 6)
	vPrime := (4*x + 5) / (2 * v)

	// Производная f(x) = (u^2/v):
	// f'(x) = (2*u*uPrime*v - u^2*vPrime) / v^2

	numerator := 2*u*uPrime*v - u*u*vPrime
	denominator := v * v

	return numerator / denominator
}

// Метод Эйлера (Маклорена)
func eulerMethod(fn func(float64) float64, a, b, h float64) float64 {
	n := int(math.Ceil((b - a) / h))
	h = (b - a) / float64(n)

	fPrimeA := derivative(a)
	fPrimeB := derivative(b)

	// Сначала формула трапеций: h * [(f(a) + f(b))/2 + Sumf(x_i)]
	sum := (fn(a) + fn(b)) / 2.0
	for i := 1; i < n; i++ {
		x := a + float64(i)*h
		sum += fn(x)
	}
	trapezoidPart := sum * h

	// Поправка Эйлера: h^2/12 * [f'(a) - f'(b)]
	correction := (h * h / 12.0) * (fPrimeA - fPrimeB)

	return trapezoidPart + correction
}

// Метод Рунге-Ромберга для уточнения результата
func rungeRomberg(Ih, Ih2 float64, k int, p int) float64 {
	return Ih2 + (Ih2-Ih)/(math.Pow(float64(k), float64(p))-1)
}

// Эталонное значение (метод Симпсона с очень малым шагом)
func exactIntegral(fn func(float64) float64, a, b float64) float64 {
	return simpsonMethod(fn, a, b, 0.00001)
}

// Структура для хранения результатов
type IntegrationResults struct {
	h           float64
	midpoint    float64
	trapezoid   float64
	simpson     float64
	euler       float64
	midpointRR  float64
	trapezoidRR float64
	simpsonRR   float64
	eulerRR     float64
	exact       float64
}

// Вычисление всех методов для заданного шага
func computeAllMethods(fn func(float64) float64, a, b, h float64) IntegrationResults {
	return IntegrationResults{
		h:         h,
		midpoint:  midpointMethod(fn, a, b, h),
		trapezoid: trapezoidMethod(fn, a, b, h),
		simpson:   simpsonMethod(fn, a, b, h),
		euler:     eulerMethod(fn, a, b, h),
	}
}

// Полный анализ с двумя шагами и методом Рунге-Ромберга
func fullAnalysis(fn func(float64) float64, a, b, h1 float64) (IntegrationResults, IntegrationResults, IntegrationResults, float64) {
	h2 := h1 / 2.0

	exact := exactIntegral(fn, a, b)
	resultsH1 := computeAllMethods(fn, a, b, h1)
	resultsH2 := computeAllMethods(fn, a, b, h2)

	// Метод Рунге-Ромберга
	resultsRR := IntegrationResults{
		h:           h2,
		midpointRR:  rungeRomberg(resultsH1.midpoint, resultsH2.midpoint, 2, 2),
		trapezoidRR: rungeRomberg(resultsH1.trapezoid, resultsH2.trapezoid, 2, 2),
		simpsonRR:   rungeRomberg(resultsH1.simpson, resultsH2.simpson, 2, 4),
		eulerRR:     rungeRomberg(resultsH1.euler, resultsH2.euler, 2, 2),
		exact:       exact,
	}

	return resultsH1, resultsH2, resultsRR, exact
}

// Создание графика для метода средних прямоугольников
func createMidpointPlot(fn func(float64) float64, a, b, h float64, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	// Найдем min и max значения функции
	yMin, yMax := math.MaxFloat64, -math.MaxFloat64
	for x := a; x <= b; x += 0.01 {
		y := fn(x)
		if y < yMin {
			yMin = y
		}
		if y > yMax {
			yMax = y
		}
	}
	yMin = math.Floor(yMin) - 1
	yMax = math.Ceil(yMax) + 1

	marginLeft, marginRight := 40.0, 20.0
	marginTop, marginBottom := 40.0, 80.0
	plotWidth := width - marginLeft - marginRight
	plotHeight := height - marginTop - marginBottom

	xToPixel := func(x float64) float32 {
		return float32(marginLeft + (x-a)/(b-a)*plotWidth)
	}
	yToPixel := func(y float64) float32 {
		return float32(marginTop + (yMax-y)/(yMax-yMin)*plotHeight)
	}

	// Сетка
	gridColor := color.RGBA{220, 220, 220, 255}
	for x := math.Ceil(a); x <= math.Floor(b); x++ {
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(x), yToPixel(yMax))
		gridLine.Position2 = fyne.NewPos(xToPixel(x), yToPixel(yMin))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}

	// Оси
	xAxis := canvas.NewLine(color.Black)
	xAxis.Position1 = fyne.NewPos(xToPixel(a), yToPixel(0))
	xAxis.Position2 = fyne.NewPos(xToPixel(b), yToPixel(0))
	xAxis.StrokeWidth = 2
	objects = append(objects, xAxis)

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	// Метки осей
	for x := math.Ceil(a); x <= math.Floor(b); x++ {
		label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
		label.TextSize = 10
		label.Move(fyne.NewPos(xToPixel(x)-8, yToPixel(0)+10))
		objects = append(objects, label)
	}

	// Прямоугольники
	n := int(math.Ceil((b - a) / h))
	h = (b - a) / float64(n)
	rectColor := color.RGBA{100, 150, 255, 100}

	for i := 0; i < n; i++ {
		xLeft := a + float64(i)*h
		xRight := a + float64(i+1)*h
		xMid := xLeft + h/2
		yVal := fn(xMid)

		rect := canvas.NewRectangle(rectColor)
		x1 := xToPixel(xLeft)
		x2 := xToPixel(xRight)
		y1 := yToPixel(0)
		y2 := yToPixel(yVal)

		if y1 > y2 {
			y1, y2 = y2, y1
		}

		rect.Move(fyne.NewPos(x1, y2))
		rect.Resize(fyne.NewSize(x2-x1, y1-y2))
		objects = append(objects, rect)

		// Границы прямоугольника
		borderColor := color.RGBA{50, 100, 200, 255}
		line1 := canvas.NewLine(borderColor)
		line1.Position1 = fyne.NewPos(x1, yToPixel(0))
		line1.Position2 = fyne.NewPos(x1, yToPixel(yVal))
		line1.StrokeWidth = 1
		objects = append(objects, line1)

		line2 := canvas.NewLine(borderColor)
		line2.Position1 = fyne.NewPos(x1, yToPixel(yVal))
		line2.Position2 = fyne.NewPos(x2, yToPixel(yVal))
		line2.StrokeWidth = 1
		objects = append(objects, line2)

		line3 := canvas.NewLine(borderColor)
		line3.Position1 = fyne.NewPos(x2, yToPixel(0))
		line3.Position2 = fyne.NewPos(x2, yToPixel(yVal))
		line3.StrokeWidth = 1
		objects = append(objects, line3)
	}

	// График функции
	funcColor := color.RGBA{255, 0, 0, 255}
	for x := a; x <= b-0.01; x += 0.01 {
		y1 := fn(x)
		y2 := fn(x + 0.01)
		if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
			line := canvas.NewLine(funcColor)
			line.Position1 = fyne.NewPos(xToPixel(x), yToPixel(y1))
			line.Position2 = fyne.NewPos(xToPixel(x+0.01), yToPixel(y2))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}

	return container.NewWithoutLayout(objects...)
}

// Создание графика для метода трапеций
func createTrapezoidPlot(fn func(float64) float64, a, b, h float64, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	yMin, yMax := math.MaxFloat64, -math.MaxFloat64
	for x := a; x <= b; x += 0.01 {
		y := fn(x)
		if y < yMin {
			yMin = y
		}
		if y > yMax {
			yMax = y
		}
	}
	yMin = math.Floor(yMin) - 1
	yMax = math.Ceil(yMax) + 1

	marginLeft, marginRight := 80.0, 20.0
	marginTop, marginBottom := 40.0, 80.0
	plotWidth := width - marginLeft - marginRight
	plotHeight := height - marginTop - marginBottom

	xToPixel := func(x float64) float32 {
		return float32(marginLeft + (x-a)/(b-a)*plotWidth)
	}
	yToPixel := func(y float64) float32 {
		return float32(marginTop + (yMax-y)/(yMax-yMin)*plotHeight)
	}

	// Сетка
	gridColor := color.RGBA{220, 220, 220, 255}
	for x := math.Ceil(a); x <= math.Floor(b); x++ {
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(x), yToPixel(yMax))
		gridLine.Position2 = fyne.NewPos(xToPixel(x), yToPixel(yMin))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}

	// Оси
	xAxis := canvas.NewLine(color.Black)
	xAxis.Position1 = fyne.NewPos(xToPixel(a), yToPixel(0))
	xAxis.Position2 = fyne.NewPos(xToPixel(b), yToPixel(0))
	xAxis.StrokeWidth = 2
	objects = append(objects, xAxis)

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	// Метки осей
	for x := math.Ceil(a); x <= math.Floor(b); x++ {
		label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
		label.TextSize = 10
		label.Move(fyne.NewPos(xToPixel(x)-8, yToPixel(0)+10))
		objects = append(objects, label)
	}

	// Трапеции
	n := int(math.Ceil((b - a) / h))
	h = (b - a) / float64(n)
	trapColor := color.RGBA{100, 255, 150, 100}

	for i := 0; i < n; i++ {
		xLeft := a + float64(i)*h
		xRight := a + float64(i+1)*h
		yLeft := fn(xLeft)
		yRight := fn(xRight)

		// Заполнение трапеции
		steps := 50
		for j := 0; j < steps; j++ {
			t1 := float64(j) / float64(steps)
			t2 := float64(j+1) / float64(steps)
			x1 := xLeft + t1*h
			x2 := xLeft + t2*h
			y1 := yLeft + t1*(yRight-yLeft)
			y2 := yLeft + t2*(yRight-yLeft)

			rect := canvas.NewRectangle(trapColor)
			px1 := xToPixel(x1)
			px2 := xToPixel(x2)
			py1 := yToPixel(0)
			py2_1 := yToPixel(y1)
			py2_2 := yToPixel(y2)
			py2 := (py2_1 + py2_2) / 2

			if py1 > py2 {
				rect.Move(fyne.NewPos(px1, py2))
				rect.Resize(fyne.NewSize(px2-px1, py1-py2))
			} else {
				rect.Move(fyne.NewPos(px1, py1))
				rect.Resize(fyne.NewSize(px2-px1, py2-py1))
			}
			objects = append(objects, rect)
		}

		// Границы трапеции
		borderColor := color.RGBA{50, 200, 100, 255}
		line1 := canvas.NewLine(borderColor)
		line1.Position1 = fyne.NewPos(xToPixel(xLeft), yToPixel(0))
		line1.Position2 = fyne.NewPos(xToPixel(xLeft), yToPixel(yLeft))
		line1.StrokeWidth = 1
		objects = append(objects, line1)

		line2 := canvas.NewLine(borderColor)
		line2.Position1 = fyne.NewPos(xToPixel(xLeft), yToPixel(yLeft))
		line2.Position2 = fyne.NewPos(xToPixel(xRight), yToPixel(yRight))
		line2.StrokeWidth = 2
		objects = append(objects, line2)

		line3 := canvas.NewLine(borderColor)
		line3.Position1 = fyne.NewPos(xToPixel(xRight), yToPixel(0))
		line3.Position2 = fyne.NewPos(xToPixel(xRight), yToPixel(yRight))
		line3.StrokeWidth = 1
		objects = append(objects, line3)
	}

	// График функции
	funcColor := color.RGBA{255, 0, 0, 255}
	for x := a; x <= b-0.01; x += 0.01 {
		y1 := fn(x)
		y2 := fn(x + 0.01)
		if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
			line := canvas.NewLine(funcColor)
			line.Position1 = fyne.NewPos(xToPixel(x), yToPixel(y1))
			line.Position2 = fyne.NewPos(xToPixel(x+0.01), yToPixel(y2))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}

	return container.NewWithoutLayout(objects...)
}

// Создание графика для метода Симпсона (с параболами)
// Создание графика для метода Симпсона (с параболами)
func createSimpsonPlot(fn func(float64) float64, a, b, h float64, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	yMin, yMax := math.MaxFloat64, -math.MaxFloat64
	for x := a; x <= b; x += 0.01 {
		y := fn(x)
		if y < yMin {
			yMin = y
		}
		if y > yMax {
			yMax = y
		}
	}
	yMin = math.Floor(yMin) - 1
	yMax = math.Ceil(yMax) + 1

	marginLeft, marginRight := 80.0, 20.0
	marginTop, marginBottom := 40.0, 80.0
	plotWidth := width - marginLeft - marginRight
	plotHeight := height - marginTop - marginBottom

	xToPixel := func(x float64) float32 {
		return float32(marginLeft + (x-a)/(b-a)*plotWidth)
	}
	yToPixel := func(y float64) float32 {
		return float32(marginTop + (yMax-y)/(yMax-yMin)*plotHeight)
	}

	// Сетка и оси (оставляем без изменений)
	gridColor := color.RGBA{220, 220, 220, 255}
	for x := math.Ceil(a); x <= math.Floor(b); x++ {
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(x), yToPixel(yMax))
		gridLine.Position2 = fyne.NewPos(xToPixel(x), yToPixel(yMin))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}

	xAxis := canvas.NewLine(color.Black)
	xAxis.Position1 = fyne.NewPos(xToPixel(a), yToPixel(0))
	xAxis.Position2 = fyne.NewPos(xToPixel(b), yToPixel(0))
	xAxis.StrokeWidth = 2
	objects = append(objects, xAxis)

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	for x := math.Ceil(a); x <= math.Floor(b); x++ {
		label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
		label.TextSize = 10
		label.Move(fyne.NewPos(xToPixel(x)-8, yToPixel(0)+10))
		objects = append(objects, label)
	}

	// ИСПРАВЛЕННАЯ ЧАСТЬ: Параболы для метода Симпсона
	n := int(math.Ceil((b - a) / h))
	// Для метода Симпсона нужно четное число интервалов
	if n%2 != 0 {
		n++
	}
	h = (b - a) / float64(n)

	parabolaColor := color.RGBA{255, 150, 0, 255}
	fillColor := color.RGBA{255, 200, 100, 80} // Цвет заливки

	// Строим параболы на каждом двойном интервале
	for i := 0; i < n; i += 2 {
		x0 := a + float64(i)*h
		x1 := a + float64(i+1)*h
		x2 := a + float64(i+2)*h

		y0 := fn(x0)
		y1 := fn(x1)
		y2 := fn(x2)

		// Заливка под параболой
		for t := 0.0; t <= 1.0; t += 0.02 {
			xStart := x0 + t*2*h
			xEnd := x0 + (t+0.02)*2*h

			// Вычисляем y на параболе через интерполяцию Лагранжа
			L0 := ((xStart - x1) * (xStart - x2)) / ((x0 - x1) * (x0 - x2))
			L1 := ((xStart - x0) * (xStart - x2)) / ((x1 - x0) * (x1 - x2))
			L2 := ((xStart - x0) * (xStart - x1)) / ((x2 - x0) * (x2 - x1))
			yStart := y0*L0 + y1*L1 + y2*L2

			L0 = ((xEnd - x1) * (xEnd - x2)) / ((x0 - x1) * (x0 - x2))
			L1 = ((xEnd - x0) * (xEnd - x2)) / ((x1 - x0) * (x1 - x2))
			L2 = ((xEnd - x0) * (xEnd - x1)) / ((x2 - x0) * (x2 - x1))
			yEnd := y0*L0 + y1*L1 + y2*L2

			// Создаем прямоугольник для заливки
			if yStart >= yMin && yStart <= yMax && yEnd >= yMin && yEnd <= yMax {
				rect := canvas.NewRectangle(fillColor)
				px1 := xToPixel(xStart)
				px2 := xToPixel(xEnd)
				py1 := yToPixel(0)
				py2 := yToPixel(math.Max(yStart, yEnd))

				if py1 > py2 {
					rect.Move(fyne.NewPos(px1, py2))
					rect.Resize(fyne.NewSize(px2-px1, py1-py2))
				} else {
					rect.Move(fyne.NewPos(px1, py1))
					rect.Resize(fyne.NewSize(px2-px1, py2-py1))
				}
				objects = append(objects, rect)
			}
		}

		// График параболы
		for t := 0.0; t <= 1.0; t += 0.01 {
			x := x0 + t*2*h
			L0 := ((x - x1) * (x - x2)) / ((x0 - x1) * (x0 - x2))
			L1 := ((x - x0) * (x - x2)) / ((x1 - x0) * (x1 - x2))
			L2 := ((x - x0) * (x - x1)) / ((x2 - x0) * (x2 - x1))
			y := y0*L0 + y1*L1 + y2*L2

			if t < 1.0 {
				xNext := x0 + (t+0.01)*2*h
				L0Next := ((xNext - x1) * (xNext - x2)) / ((x0 - x1) * (x0 - x2))
				L1Next := ((xNext - x0) * (xNext - x2)) / ((x1 - x0) * (x1 - x2))
				L2Next := ((xNext - x0) * (xNext - x1)) / ((x2 - x0) * (x2 - x1))
				yNext := y0*L0Next + y1*L1Next + y2*L2Next

				if y >= yMin && y <= yMax && yNext >= yMin && yNext <= yMax {
					line := canvas.NewLine(parabolaColor)
					line.Position1 = fyne.NewPos(xToPixel(x), yToPixel(y))
					line.Position2 = fyne.NewPos(xToPixel(xNext), yToPixel(yNext))
					line.StrokeWidth = 2
					objects = append(objects, line)
				}
			}
		}

		// Вертикальные линии на границах интервалов
		borderColor := color.RGBA{200, 120, 0, 255}
		line1 := canvas.NewLine(borderColor)
		line1.Position1 = fyne.NewPos(xToPixel(x0), yToPixel(0))
		line1.Position2 = fyne.NewPos(xToPixel(x0), yToPixel(y0))
		line1.StrokeWidth = 1
		objects = append(objects, line1)

		line2 := canvas.NewLine(borderColor)
		line2.Position1 = fyne.NewPos(xToPixel(x2), yToPixel(0))
		line2.Position2 = fyne.NewPos(xToPixel(x2), yToPixel(y2))
		line2.StrokeWidth = 1
		objects = append(objects, line2)
	}

	// График исходной функции
	funcColor := color.RGBA{255, 0, 0, 255}
	for x := a; x <= b-0.01; x += 0.01 {
		y1 := fn(x)
		y2 := fn(x + 0.01)
		if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
			line := canvas.NewLine(funcColor)
			line.Position1 = fyne.NewPos(xToPixel(x), yToPixel(y1))
			line.Position2 = fyne.NewPos(xToPixel(x+0.01), yToPixel(y2))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}

	return container.NewWithoutLayout(objects...)
}

// Создание графика для метода Эйлера (с касательными)
func createEulerPlot(fn func(float64) float64, a, b, h float64, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	yMin, yMax := math.MaxFloat64, -math.MaxFloat64
	for x := a; x <= b; x += 0.01 {
		y := fn(x)
		if y < yMin {
			yMin = y
		}
		if y > yMax {
			yMax = y
		}
	}
	yMin = math.Floor(yMin) - 1
	yMax = math.Ceil(yMax) + 1

	marginLeft, marginRight := 80.0, 20.0
	marginTop, marginBottom := 40.0, 80.0
	plotWidth := width - marginLeft - marginRight
	plotHeight := height - marginTop - marginBottom

	xToPixel := func(x float64) float32 {
		return float32(marginLeft + (x-a)/(b-a)*plotWidth)
	}
	yToPixel := func(y float64) float32 {
		return float32(marginTop + (yMax-y)/(yMax-yMin)*plotHeight)
	}

	// Сетка
	gridColor := color.RGBA{220, 220, 220, 255}
	for x := math.Ceil(a); x <= math.Floor(b); x++ {
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(x), yToPixel(yMax))
		gridLine.Position2 = fyne.NewPos(xToPixel(x), yToPixel(yMin))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)
	}

	// Оси
	xAxis := canvas.NewLine(color.Black)
	xAxis.Position1 = fyne.NewPos(xToPixel(a), yToPixel(0))
	xAxis.Position2 = fyne.NewPos(xToPixel(b), yToPixel(0))
	xAxis.StrokeWidth = 2
	objects = append(objects, xAxis)

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(yMin))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	// Метки осей
	for x := math.Ceil(a); x <= math.Floor(b); x++ {
		label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
		label.TextSize = 10
		label.Move(fyne.NewPos(xToPixel(x)-8, yToPixel(0)+10))
		objects = append(objects, label)
	}

	// Касательные в точках сетки
	n := int(math.Ceil((b - a) / h))
	h = (b - a) / float64(n)
	tangentColor := color.RGBA{128, 0, 255, 255}

	for i := 0; i <= n; i++ {
		x0 := a + float64(i)*h
		y0 := fn(x0)

		// Вычисляем производную используя отдельную функцию
		deriv := derivative(x0)

		// Рисуем касательную линию
		tangentLen := h * 0.5
		xStart := x0 - tangentLen
		xEnd := x0 + tangentLen
		yStart := y0 - deriv*tangentLen
		yEnd := y0 + deriv*tangentLen

		if yStart >= yMin && yStart <= yMax && yEnd >= yMin && yEnd <= yMax {
			line := canvas.NewLine(tangentColor)
			line.Position1 = fyne.NewPos(xToPixel(xStart), yToPixel(yStart))
			line.Position2 = fyne.NewPos(xToPixel(xEnd), yToPixel(yEnd))
			line.StrokeWidth = 1.5
			objects = append(objects, line)
		}

		// Точка на графике
		circle := canvas.NewCircle(tangentColor)
		circle.FillColor = tangentColor
		circle.Resize(fyne.NewSize(5, 5))
		circle.Move(fyne.NewPos(xToPixel(x0)-2.5, yToPixel(y0)-2.5))
		objects = append(objects, circle)
	}

	// График функции
	funcColor := color.RGBA{255, 0, 0, 255}
	for x := a; x <= b-0.01; x += 0.01 {
		y1 := fn(x)
		y2 := fn(x + 0.01)
		if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
			line := canvas.NewLine(funcColor)
			line.Position1 = fyne.NewPos(xToPixel(x), yToPixel(y1))
			line.Position2 = fyne.NewPos(xToPixel(x+0.01), yToPixel(y2))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}

	return container.NewWithoutLayout(objects...)
}

// Создание таблицы результатов
func createResultsTable(resultsH1, resultsH2, resultsRR IntegrationResults, exact float64) string {
	var text string
	separator := "═════════════════════════════════════════════════════════════════\n"
	line := "─────────────────────────────────────────────────────────────────\n"

	text += separator
	text += "                     ЧИСЛЕННОЕ ИНТЕГРИРОВАНИЕ\n"
	text += separator
	text += "\n"
	text += "Функция: F = (2sin2x - cos(x/4))² / √(2x² + 5x + 6)\n"
	text += fmt.Sprintf("Интервал: [%.1f, %.1f]\n", -2.0, 2.0)
	text += fmt.Sprintf("Эталонное значение (Симпсон с h=0.00001): %.10f\n\n", exact)

	text += line
	text += fmt.Sprintf("РЕЗУЛЬТАТЫ С ШАГОМ h = %.2f\n", resultsH1.h)
	text += line
	text += fmt.Sprintf("Метод средних прямоугольников: %16.10f\n", resultsH1.midpoint)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n", math.Abs(resultsH1.midpoint-exact))
	text += fmt.Sprintf("\nМетод трапеций:                %16.10f\n", resultsH1.trapezoid)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n", math.Abs(resultsH1.trapezoid-exact))
	text += fmt.Sprintf("\nМетод Симпсона:                %16.10f\n", resultsH1.simpson)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n", math.Abs(resultsH1.simpson-exact))
	text += fmt.Sprintf("\nМетод Эйлера:                  %16.10f\n", resultsH1.euler)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n\n", math.Abs(resultsH1.euler-exact))

	text += line
	text += fmt.Sprintf("РЕЗУЛЬТАТЫ С ШАГОМ h/2 = %.2f\n", resultsH2.h)
	text += line
	text += fmt.Sprintf("Метод средних прямоугольников: %16.10f\n", resultsH2.midpoint)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n", math.Abs(resultsH2.midpoint-exact))
	text += fmt.Sprintf("\nМетод трапеций:                %16.10f\n", resultsH2.trapezoid)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n", math.Abs(resultsH2.trapezoid-exact))
	text += fmt.Sprintf("\nМетод Симпсона:                %16.10f\n", resultsH2.simpson)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n", math.Abs(resultsH2.simpson-exact))
	text += fmt.Sprintf("\nМетод Эйлера:                  %16.10f\n", resultsH2.euler)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n\n", math.Abs(resultsH2.euler-exact))

	text += line
	text += "УТОЧНЕНИЕ ПО МЕТОДУ РУНГЕ-РОМБЕРГА\n"
	text += line
	text += fmt.Sprintf("Средние прямоугольники (p=2): %16.10f\n", resultsRR.midpointRR)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n", math.Abs(resultsRR.midpointRR-exact))
	if math.Abs(resultsRR.midpointRR-exact) > 0 {
		text += fmt.Sprintf("  Улучшение:                   %16.2fx\n", math.Abs(resultsH2.midpoint-exact)/math.Abs(resultsRR.midpointRR-exact))
	}

	text += fmt.Sprintf("\nТрапеции (p=2):                %16.10f\n", resultsRR.trapezoidRR)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n", math.Abs(resultsRR.trapezoidRR-exact))
	if math.Abs(resultsRR.trapezoidRR-exact) > 0 {
		text += fmt.Sprintf("  Улучшение:                   %16.2fx\n", math.Abs(resultsH2.trapezoid-exact)/math.Abs(resultsRR.trapezoidRR-exact))
	}

	text += fmt.Sprintf("\nСимпсон (p=4):                 %16.10f\n", resultsRR.simpsonRR)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n", math.Abs(resultsRR.simpsonRR-exact))
	if math.Abs(resultsRR.simpsonRR-exact) > 0 {
		text += fmt.Sprintf("  Улучшение:                   %16.2fx\n", math.Abs(resultsH2.simpson-exact)/math.Abs(resultsRR.simpsonRR-exact))
	}

	text += fmt.Sprintf("\nЭйлер (p=2):                   %16.10f\n", resultsRR.eulerRR)
	text += fmt.Sprintf("  Погрешность:                 %16.6e\n", math.Abs(resultsRR.eulerRR-exact))
	if math.Abs(resultsRR.eulerRR-exact) > 0 {
		text += fmt.Sprintf("  Улучшение:                   %16.2fx\n", math.Abs(resultsH2.euler-exact)/math.Abs(resultsRR.eulerRR-exact))
	}

	text += "\n" + separator

	return text
}

// Создание сравнительной таблицы
func createComparisonTable(resultsH1, resultsH2, resultsRR IntegrationResults, exact float64) string {
	var text string
	separator := "═════════════════════════════════════════════════════════════════\n"
	line := "─────────────────────────────────────────────────────────────────\n"

	text += separator
	text += "                  СРАВНИТЕЛЬНАЯ ТАБЛИЦА МЕТОДОВ\n"
	text += separator
	text += "\n"

	text += fmt.Sprintf("%-25s | %15s | %15s | %15s\n", "Метод", "h = "+fmt.Sprintf("%.2f", resultsH1.h), "h/2 = "+fmt.Sprintf("%.2f", resultsH2.h), "Рунге-Ромберг")
	text += line

	text += fmt.Sprintf("%-25s | %15.10f | %15.10f | %15.10f\n", "Средние прямоугольники", resultsH1.midpoint, resultsH2.midpoint, resultsRR.midpointRR)
	text += fmt.Sprintf("%-25s | %15.6e | %15.6e | %15.6e\n\n", "  Погрешность", math.Abs(resultsH1.midpoint-exact), math.Abs(resultsH2.midpoint-exact), math.Abs(resultsRR.midpointRR-exact))

	text += fmt.Sprintf("%-25s | %15.10f | %15.10f | %15.10f\n", "Трапеции", resultsH1.trapezoid, resultsH2.trapezoid, resultsRR.trapezoidRR)
	text += fmt.Sprintf("%-25s | %15.6e | %15.6e | %15.6e\n\n", "  Погрешность", math.Abs(resultsH1.trapezoid-exact), math.Abs(resultsH2.trapezoid-exact), math.Abs(resultsRR.trapezoidRR-exact))

	text += fmt.Sprintf("%-25s | %15.10f | %15.10f | %15.10f\n", "Симпсон", resultsH1.simpson, resultsH2.simpson, resultsRR.simpsonRR)
	text += fmt.Sprintf("%-25s | %15.6e | %15.6e | %15.6e\n\n", "  Погрешность", math.Abs(resultsH1.simpson-exact), math.Abs(resultsH2.simpson-exact), math.Abs(resultsRR.simpsonRR-exact))

	text += fmt.Sprintf("%-25s | %15.10f | %15.10f | %15.10f\n", "Эйлер", resultsH1.euler, resultsH2.euler, resultsRR.eulerRR)
	text += fmt.Sprintf("%-25s | %15.6e | %15.6e | %15.6e\n\n", "  Погрешность", math.Abs(resultsH1.euler-exact), math.Abs(resultsH2.euler-exact), math.Abs(resultsRR.eulerRR-exact))

	text += separator

	return text
}

// Информационная вкладка
func createInfoText() string {
	return `ЧИСЛЕННОЕ ИНТЕГРИРОВАНИЕ

Функция: F = (2sin2x - cos(x/4))² / √(2x² + 5x + 6)
Интервал: [-2, 2]
Начальный шаг: h = 0.4

МЕТОДЫ ИНТЕГРИРОВАНИЯ:

1. Метод средних прямоугольников (порядок точности O(h²))
   ∫f(x)dx ≈ h·Σf(x_i + h/2)

   Визуализация: прямоугольники, высота которых равна значению
   функции в середине каждого интервала.

2. Метод трапеций (порядок точности O(h²))
   ∫f(x)dx ≈ h·[(f(a) + f(b))/2 + Σf(x_i)]

   Визуализация: трапеции, соединяющие соседние точки функции
   прямыми линиями.

3. Метод Симпсона (порядок точности O(h⁴))
   ∫f(x)dx ≈ (h/3)·[f(a) + 4·Σf(x_{2i-1}) + 2·Σf(x_{2i}) + f(b)]
   Требует четное число интервалов

   Визуализация: параболы, проходящие через каждые три
   последовательные точки функции.

4. Метод Эйлера (уточненный метод трапеций)
   ∫f(x)dx ≈ T_h + (h²/12)·[f'(a) - f'(b)]
   где T_h - результат метода трапеций

   Визуализация: касательные к функции в точках сетки,
   показывающие использование производных для уточнения.

МЕТОД РУНГЕ-РОМБЕРГА:

Уточнение результата по формуле:
I_refined = I_h/2 + (I_h/2 - I_h)/(k^p - 1)

где:
  k = 2 (коэффициент уменьшения шага)
  p - порядок точности метода:
    • p = 2 для методов прямоугольников, трапеций, Эйлера
    • p = 4 для метода Симпсона

АНАЛИЗ ПОГРЕШНОСТИ:

Точное значение вычисляется методом Симпсона с очень малым шагом (h=0.00001).
Для каждого метода вычисляется абсолютная погрешность:
ε = |I_вычисленное - I_точное|

При уменьшении шага в 2 раза:
• Методы O(h²): погрешность уменьшается в ~4 раза
• Метод O(h⁴): погрешность уменьшается в ~16 раз

Метод Рунге-Ромберга дополнительно уточняет результат,
используя разность значений при двух шагах.

ЦВЕТОВАЯ СХЕМА ГРАФИКОВ:

• Красный - график исходной функции
• Синий - прямоугольники (метод средних прямоугольников)
• Зеленый - трапеции (метод трапеций)
• Оранжевый - параболы (метод Симпсона)
• Фиолетовый - касательные (метод Эйлера)`
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Численное интегрирование")
	myWindow.Resize(fyne.NewSize(1200, 900))

	a, b := -2.0, 2.0
	h1 := 0.5

	resultsH1, resultsH2, resultsRR, exact := fullAnalysis(integrandF, a, b, h1)

	// Вывод результатов в терминал
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("                     ЧИСЛЕННОЕ ИНТЕГРИРОВАНИЕ")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Функция: F = (2sin2x - cos(x/4))² / √(2x² + 5x + 6)")
	fmt.Printf("Интервал: [%.1f, %.1f]\n", a, b)
	fmt.Printf("Начальный шаг: h = %.2f\n", h1)
	fmt.Printf("Эталонное значение (Симпсон с h=0.00001): %.10f\n\n", exact)

	fmt.Println("───────────────────────────────────────────────────────────────────")
	fmt.Printf("РЕЗУЛЬТАТЫ С ШАГОМ h = %.2f\n", resultsH1.h)
	fmt.Println("───────────────────────────────────────────────────────────────────")
	fmt.Printf("Метод средних прямоугольников: %16.10f\n", resultsH1.midpoint)
	fmt.Printf("  Погрешность:                 %16.6e\n", math.Abs(resultsH1.midpoint-exact))
	fmt.Printf("\nМетод трапеций:                %16.10f\n", resultsH1.trapezoid)
	fmt.Printf("  Погрешность:                 %16.6e\n", math.Abs(resultsH1.trapezoid-exact))
	fmt.Printf("\nМетод Симпсона:                %16.10f\n", resultsH1.simpson)
	fmt.Printf("  Погрешность:                 %16.6e\n", math.Abs(resultsH1.simpson-exact))
	fmt.Printf("\nМетод Эйлера:                  %16.10f\n", resultsH1.euler)
	fmt.Printf("  Погрешность:                 %16.6e\n\n", math.Abs(resultsH1.euler-exact))

	fmt.Println("───────────────────────────────────────────────────────────────────")
	fmt.Printf("РЕЗУЛЬТАТЫ С ШАГОМ h/2 = %.2f\n", resultsH2.h)
	fmt.Println("───────────────────────────────────────────────────────────────────")
	fmt.Printf("Метод средних прямоугольников: %16.10f\n", resultsH2.midpoint)
	fmt.Printf("  Погрешность:                 %16.6e\n", math.Abs(resultsH2.midpoint-exact))
	fmt.Printf("\nМетод трапеций:                %16.10f\n", resultsH2.trapezoid)
	fmt.Printf("  Погрешность:                 %16.6e\n", math.Abs(resultsH2.trapezoid-exact))
	fmt.Printf("\nМетод Симпсона:                %16.10f\n", resultsH2.simpson)
	fmt.Printf("  Погрешность:                 %16.6e\n", math.Abs(resultsH2.simpson-exact))
	fmt.Printf("\nМетод Эйлера:                  %16.10f\n", resultsH2.euler)
	fmt.Printf("  Погрешность:                 %16.6e\n\n", math.Abs(resultsH2.euler-exact))

	fmt.Println("───────────────────────────────────────────────────────────────────")
	fmt.Println("УТОЧНЕНИЕ ПО МЕТОДУ РУНГЕ-РОМБЕРГА")
	fmt.Println("───────────────────────────────────────────────────────────────────")
	fmt.Printf("Средние прямоугольники (p=2): %16.10f\n", resultsRR.midpointRR)
	fmt.Printf("  Погрешность:                 %16.6e\n", math.Abs(resultsRR.midpointRR-exact))
	if math.Abs(resultsRR.midpointRR-exact) > 0 {
		fmt.Printf("  Улучшение:                   %16.2fx\n", math.Abs(resultsH2.midpoint-exact)/math.Abs(resultsRR.midpointRR-exact))
	}

	fmt.Printf("\nТрапеции (p=2):                %16.10f\n", resultsRR.trapezoidRR)
	fmt.Printf("  Погрешность:                 %16.6e\n", math.Abs(resultsRR.trapezoidRR-exact))
	if math.Abs(resultsRR.trapezoidRR-exact) > 0 {
		fmt.Printf("  Улучшение:                   %16.2fx\n", math.Abs(resultsH2.trapezoid-exact)/math.Abs(resultsRR.trapezoidRR-exact))
	}

	fmt.Printf("\nСимпсон (p=4):                 %16.10f\n", resultsRR.simpsonRR)
	fmt.Printf("  Погрешность:                 %16.6e\n", math.Abs(resultsRR.simpsonRR-exact))
	if math.Abs(resultsRR.simpsonRR-exact) > 0 {
		fmt.Printf("  Улучшение:                   %16.2fx\n", math.Abs(resultsH2.simpson-exact)/math.Abs(resultsRR.simpsonRR-exact))
	}

	fmt.Printf("\nЭйлер (p=2):                   %16.10f\n", resultsRR.eulerRR)
	fmt.Printf("  Погрешность:                 %16.6e\n", math.Abs(resultsRR.eulerRR-exact))
	if math.Abs(resultsRR.eulerRR-exact) > 0 {
		fmt.Printf("  Улучшение:                   %16.2fx\n", math.Abs(resultsH2.euler-exact)/math.Abs(resultsRR.eulerRR-exact))
	}

	fmt.Println("\n═══════════════════════════════════════════════════════════════════")
	fmt.Println("                  СРАВНИТЕЛЬНАЯ ТАБЛИЦА МЕТОДОВ")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Printf("%-25s | %15s | %15s | %15s\n", "Метод", fmt.Sprintf("h = %.2f", resultsH1.h), fmt.Sprintf("h/2 = %.2f", resultsH2.h), "Рунге-Ромберг")
	fmt.Println("───────────────────────────────────────────────────────────────────")

	fmt.Printf("%-25s | %15.10f | %15.10f | %15.10f\n", "Средние прямоугольники", resultsH1.midpoint, resultsH2.midpoint, resultsRR.midpointRR)
	fmt.Printf("%-25s | %15.6e | %15.6e | %15.6e\n\n", "  Погрешность", math.Abs(resultsH1.midpoint-exact), math.Abs(resultsH2.midpoint-exact), math.Abs(resultsRR.midpointRR-exact))

	fmt.Printf("%-25s | %15.10f | %15.10f | %15.10f\n", "Трапеции", resultsH1.trapezoid, resultsH2.trapezoid, resultsRR.trapezoidRR)
	fmt.Printf("%-25s | %15.6e | %15.6e | %15.6e\n\n", "  Погрешность", math.Abs(resultsH1.trapezoid-exact), math.Abs(resultsH2.trapezoid-exact), math.Abs(resultsRR.trapezoidRR-exact))

	fmt.Printf("%-25s | %15.10f | %15.10f | %15.10f\n", "Симпсон", resultsH1.simpson, resultsH2.simpson, resultsRR.simpsonRR)
	fmt.Printf("%-25s | %15.6e | %15.6e | %15.6e\n\n", "  Погрешность", math.Abs(resultsH1.simpson-exact), math.Abs(resultsH2.simpson-exact), math.Abs(resultsRR.simpsonRR-exact))

	fmt.Printf("%-25s | %15.10f | %15.10f | %15.10f\n", "Эйлер", resultsH1.euler, resultsH2.euler, resultsRR.eulerRR)
	fmt.Printf("%-25s | %15.6e | %15.6e | %15.6e\n\n", "  Погрешность", math.Abs(resultsH1.euler-exact), math.Abs(resultsH2.euler-exact), math.Abs(resultsRR.eulerRR-exact))

	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()

	// Создание вкладок с результатами и сравнением
	resultsText := createResultsTable(resultsH1, resultsH2, resultsRR, exact)
	resultsLabel := widget.NewLabel(resultsText)
	resultsLabel.Wrapping = fyne.TextWrapOff
	resultsContainer := container.NewVScroll(resultsLabel)

	comparisonText := createComparisonTable(resultsH1, resultsH2, resultsRR, exact)
	comparisonLabel := widget.NewLabel(comparisonText)
	comparisonLabel.Wrapping = fyne.TextWrapOff
	comparisonContainer := container.NewVScroll(comparisonLabel)

	infoText := createInfoText()
	infoLabel := widget.NewLabel(infoText)
	infoLabel.Wrapping = fyne.TextWrapWord
	infoContainer := container.NewVScroll(infoLabel)

	// Создание вкладок для каждого метода с графиком и текстом отдельно
	midpointDetails := createMethodDetailsTable("Метод средних прямоугольников", resultsH1.midpoint, resultsH2.midpoint, resultsRR.midpointRR, exact, 2, h1)
	midpointPlot := createMidpointPlot(integrandF, a, b, h1, 1100, 400)
	midpointPlotContainer := createPlotOnlyTabContent("МЕТОД СРЕДНИХ ПРЯМОУГОЛЬНИКОВ", midpointPlot)
	midpointDetailsContainer := createDetailsOnlyTabContent(midpointDetails)

	trapezoidDetails := createMethodDetailsTable("Метод трапеций", resultsH1.trapezoid, resultsH2.trapezoid, resultsRR.trapezoidRR, exact, 2, h1)
	trapezoidPlot := createTrapezoidPlot(integrandF, a, b, h1, 1100, 400)
	trapezoidPlotContainer := createPlotOnlyTabContent("МЕТОД ТРАПЕЦИЙ", trapezoidPlot)
	trapezoidDetailsContainer := createDetailsOnlyTabContent(trapezoidDetails)

	simpsonDetails := createMethodDetailsTable("Метод Симпсона", resultsH1.simpson, resultsH2.simpson, resultsRR.simpsonRR, exact, 4, h1)
	simpsonPlot := createSimpsonPlot(integrandF, a, b, h1, 1100, 400)
	simpsonPlotContainer := createPlotOnlyTabContent("МЕТОД СИМПСОНА", simpsonPlot)
	simpsonDetailsContainer := createDetailsOnlyTabContent(simpsonDetails)

	eulerDetails := createMethodDetailsTable("Метод Эйлера", resultsH1.euler, resultsH2.euler, resultsRR.eulerRR, exact, 2, h1)
	eulerPlot := createEulerPlot(integrandF, a, b, h1, 1100, 400)
	eulerPlotContainer := createPlotOnlyTabContent("МЕТОД ЭЙЛЕРА (УТОЧНЕННЫЙ МЕТОД ТРАПЕЦИЙ)", eulerPlot)
	eulerDetailsContainer := createDetailsOnlyTabContent(eulerDetails)

	tabs := container.NewAppTabs(
		container.NewTabItem("Результаты", resultsContainer),
		container.NewTabItem("Сравнение", comparisonContainer),
		container.NewTabItem("График: Прямоугольники", midpointPlotContainer),
		container.NewTabItem("Данные: Прямоугольники", midpointDetailsContainer),
		container.NewTabItem("График: Трапеции", trapezoidPlotContainer),
		container.NewTabItem("Данные: Трапеции", trapezoidDetailsContainer),
		container.NewTabItem("График: Симпсон", simpsonPlotContainer),
		container.NewTabItem("Данные: Симпсон", simpsonDetailsContainer),
		container.NewTabItem("График: Эйлер", eulerPlotContainer),
		container.NewTabItem("Данные: Эйлер", eulerDetailsContainer),
		container.NewTabItem("Информация", infoContainer),
	)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}

// Создание детальной таблицы для конкретного метода
func createMethodDetailsTable(methodName string, valH1, valH2, valRR, exact float64, order int, h1 float64) string {
	var text string
	separator := "═══════════════════════════════════════════════════════════\n"
	line := "────────────────────────────────────────────────────────────\n"

	text += separator
	text += fmt.Sprintf("                     %s\n", methodName)
	text += separator
	text += "\n"

	text += fmt.Sprintf("Порядок точности: O(h^%d)\n", order)
	text += fmt.Sprintf("Эталонное значение: %.10f\n\n", exact)

	text += line
	text += fmt.Sprintf("Шаг h = %.2f:\n", h1)
	text += line
	text += fmt.Sprintf("  Значение:     %20.10f\n", valH1)
	text += fmt.Sprintf("  Погрешность:  %20.6e\n\n", math.Abs(valH1-exact))

	text += line
	text += fmt.Sprintf("Шаг h/2 = %.2f:\n", h1/2)
	text += line
	text += fmt.Sprintf("  Значение:     %20.10f\n", valH2)
	text += fmt.Sprintf("  Погрешность:  %20.6e\n", math.Abs(valH2-exact))

	errorH1 := math.Abs(valH1 - exact)
	errorH2 := math.Abs(valH2 - exact)
	if errorH2 > 0 {
		text += fmt.Sprintf("  Отношение погрешностей (h)/(h/2): %.2f\n", errorH1/errorH2)
		text += fmt.Sprintf("  Теоретически ожидается: %.2f (так как O(h^%d))\n\n", math.Pow(2, float64(order)), order)
	}

	text += line
	text += "Уточнение по методу Рунге-Ромберга:\n"
	text += line
	text += fmt.Sprintf("  Значение:     %20.10f\n", valRR)
	text += fmt.Sprintf("  Погрешность:  %20.6e\n", math.Abs(valRR-exact))

	if math.Abs(valRR-exact) > 0 {
		improvement := errorH2 / math.Abs(valRR-exact)
		text += fmt.Sprintf("  Улучшение по сравнению с h/2: %.2fx\n", improvement)
	}

	text += "\n" + separator

	return text
}

func createMethodTabContent(title string, plot *fyne.Container, details string) *container.Scroll {
	// Заголовок
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Разделитель
	separator := widget.NewSeparator()

	// Детали метода
	detailsLabel := widget.NewLabel(details)
	detailsLabel.Wrapping = fyne.TextWrapWord

	// Создаем основной контент
	content := container.NewVBox(
		titleLabel,
		separator,
		// График
		container.NewPadded(plot),
		widget.NewLabel(""), // Дополнительный отступ
		separator,
		// Текст деталей
		detailsLabel,
	)

	// Возвращаем скроллируемый контейнер
	return container.NewVScroll(content)
}

// Создание вкладки только с графиком
func createPlotOnlyTabContent(title string, plot *fyne.Container) *container.Scroll {
	// Заголовок
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Разделитель
	separator := widget.NewSeparator()

	// Создаем основной контент
	content := container.NewVBox(
		titleLabel,
		separator,
		// График
		container.NewPadded(plot),
	)

	// Возвращаем скроллируемый контейнер
	return container.NewVScroll(content)
}

// Создание вкладки только с деталями
func createDetailsOnlyTabContent(details string) *container.Scroll {
	// Детали метода
	detailsLabel := widget.NewLabel(details)
	detailsLabel.Wrapping = fyne.TextWrapWord

	// Возвращаем скроллируемый контейнер
	return container.NewVScroll(detailsLabel)
}
