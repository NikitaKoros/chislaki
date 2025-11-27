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

// Исходная функция
func f(x float64) float64 {
	return math.Log((math.Pow(math.Cos(x), 2)+3)/(x+1)) * (math.Pow(x, 3) + 3*math.Cos(2*x))
}

// Аналитическая первая производная
func fPrimeAnalytical(x float64) float64 {
	cosx := math.Cos(x)
	cos2x := math.Cos(2 * x)
	sinx := math.Sin(x)
	sin2x := math.Sin(2 * x)

	term1 := math.Log((cosx*cosx+3)/(x+1)) * (3*math.Pow(x, 2) - 6*sin2x)

	numerator := -2*cosx*sinx*(x+1) - (cosx*cosx + 3)
	denominator := (x + 1) * (cosx*cosx + 3)
	term2 := (numerator / denominator) * (math.Pow(x, 3) + 3*cos2x)

	return term1 + term2
}

// Аналитическая вторая производная
func fDoublePrimeAnalytical(x float64) float64 {
	cosx := math.Cos(x)
	cos2x := math.Cos(2 * x)
	sinx := math.Sin(x)
	sin2x := math.Sin(2 * x)

	// Вычисление компонентов второй производной
	ln_part := math.Log((cosx*cosx + 3) / (x + 1))
	poly_part := math.Pow(x, 3) + 3*cos2x

	// d²/dx²[ln_part * poly_part]
	d2_poly := 6*x - 12*cos2x
	d_poly := 3*math.Pow(x, 2) - 6*sin2x

	numerator1 := -2*cosx*sinx*(x+1) - (cosx*cosx + 3)
	denominator1 := (x + 1) * (cosx*cosx + 3)
	d_ln := numerator1 / denominator1

	// Вторая производная логарифмической части
	d2_ln_num := -2*(cosx*cosx-sinx*sinx)*(x+1) - 2*(-2*cosx*sinx) + 4*cosx*sinx
	d2_ln_den := denominator1
	d2_ln_correction := -(numerator1 * numerator1) / (denominator1 * denominator1)
	d2_ln := d2_ln_num/d2_ln_den + d2_ln_correction

	result := ln_part*d2_poly + 2*d_ln*d_poly + d2_ln*poly_part

	return result
}

// Схемы численного дифференцирования первого порядка

// 1. Правая разностная производная (2-точечная, O(h))
func firstDerivRightDiff(fn func(float64) float64, x, h float64) float64 {
	return (fn(x+h) - fn(x)) / h
}

// 2. Левая разностная производная (2-точечная, O(h))
func firstDerivLeftDiff(fn func(float64) float64, x, h float64) float64 {
	return (fn(x) - fn(x-h)) / h
}

// 3. Центральная разностная производная (2-точечная, O(h²))
func firstDerivCentral(fn func(float64) float64, x, h float64) float64 {
	return (fn(x+h) - fn(x-h)) / (2 * h)
}

// 4. Трехточечная формула с O(h²) - смещенная вперед
func firstDerivThreePointForward(fn func(float64) float64, x, h float64) float64 {
	return (-3*fn(x) + 4*fn(x+h) - fn(x+2*h)) / (2 * h)
}

// 5. Трехточечная формула с O(h²) - смещенная назад
func firstDerivThreePointBackward(fn func(float64) float64, x, h float64) float64 {
	return (3*fn(x) - 4*fn(x-h) + fn(x-2*h)) / (2 * h)
}

// 6. Пятиточечная центральная формула с O(h⁴)
func firstDerivFivePoint(fn func(float64) float64, x, h float64) float64 {
	return (fn(x-2*h) - 8*fn(x-h) + 8*fn(x+h) - fn(x+2*h)) / (12 * h)
}

// Схемы численного дифференцирования второго порядка

// 1. Трехточечная центральная формула для второй производной O(h²)
func secondDerivThreePoint(fn func(float64) float64, x, h float64) float64 {
	return (fn(x-h) - 2*fn(x) + fn(x+h)) / (h * h)
}

// 2. Четырехточечная формула для второй производной (смещенная вперед) O(h²)
func secondDerivFourPointForward(fn func(float64) float64, x, h float64) float64 {
	return (2*fn(x) - 5*fn(x+h) + 4*fn(x+2*h) - fn(x+3*h)) / (h * h)
}

// 3. Четырехточечная формула для второй производной (смещенная назад) O(h²)
func secondDerivFourPointBackward(fn func(float64) float64, x, h float64) float64 {
	return (2*fn(x) - 5*fn(x-h) + 4*fn(x-2*h) - fn(x-3*h)) / (h * h)
}

// 4. Пятиточечная центральная формула для второй производной O(h⁴)
func secondDerivFivePoint(fn func(float64) float64, x, h float64) float64 {
	return (-fn(x-2*h) + 16*fn(x-h) - 30*fn(x) + 16*fn(x+h) - fn(x+2*h)) / (12 * h * h)
}

type DerivativeResults struct {
	x                  float64
	analytical1        float64
	analytical2        float64
	rightDiff          float64
	leftDiff           float64
	central            float64
	threePointForward  float64
	threePointBackward float64
	fivePoint          float64
	secondDeriv3Point  float64
	secondDeriv4FWD    float64
	secondDeriv4BWD    float64
	secondDeriv5Point  float64
}

func computeDerivatives(a, b, h float64) []DerivativeResults {
	var results []DerivativeResults

	for x := a; x <= b+1e-10; x += h {
		res := DerivativeResults{x: x}
		res.analytical1 = fPrimeAnalytical(x)
		res.analytical2 = fDoublePrimeAnalytical(x)

		// Проверка границ для различных схем
		if x+h <= b+1e-10 {
			res.rightDiff = firstDerivRightDiff(f, x, h)
		}
		if x-h >= a-1e-10 {
			res.leftDiff = firstDerivLeftDiff(f, x, h)
		}
		if x-h >= a-1e-10 && x+h <= b+1e-10 {
			res.central = firstDerivCentral(f, x, h)
			res.secondDeriv3Point = secondDerivThreePoint(f, x, h)
		}
		if x+2*h <= b+1e-10 {
			res.threePointForward = firstDerivThreePointForward(f, x, h)
		}
		if x-2*h >= a-1e-10 {
			res.threePointBackward = firstDerivThreePointBackward(f, x, h)
		}
		if x-2*h >= a-1e-10 && x+2*h <= b+1e-10 {
			res.fivePoint = firstDerivFivePoint(f, x, h)
			res.secondDeriv5Point = secondDerivFivePoint(f, x, h)
		}
		if x+3*h <= b+1e-10 {
			res.secondDeriv4FWD = secondDerivFourPointForward(f, x, h)
		}
		if x-3*h >= a-1e-10 {
			res.secondDeriv4BWD = secondDerivFourPointBackward(f, x, h)
		}

		results = append(results, res)
	}

	return results
}

// Создает один общий график со всеми схемами
func createCombinedPlot(results1, results2 []DerivativeResults, derivOrder int, h1, h2 float64, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	xMin, xMax := -0.5, 3.5
	var yMin, yMax float64

	if derivOrder == 1 {
		yMin, yMax = -30.0, 10.0
	} else {
		yMin, yMax = -50.0, 50.0
	}

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
	for y := math.Ceil(yMin); y <= math.Floor(yMax); y += 10 {
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
	for y := math.Ceil(yMin); y <= math.Floor(yMax); y += 10 {
		if y != 0 {
			label := canvas.NewText(fmt.Sprintf("%.0f", y), color.Black)
			label.TextSize = 10
			label.Move(fyne.NewPos(xToPixel(0)-35, yToPixel(y)-5))
			objects = append(objects, label)
		}
	}

	// Аналитическое решение (черная линия)
	for i := 0; i < len(results1)-1; i++ {
		var y1, y2 float64
		if derivOrder == 1 {
			y1 = results1[i].analytical1
			y2 = results1[i+1].analytical1
		} else {
			y1 = results1[i].analytical2
			y2 = results1[i+1].analytical2
		}

		if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
			line := canvas.NewLine(color.Black)
			line.Position1 = fyne.NewPos(xToPixel(results1[i].x), yToPixel(y1))
			line.Position2 = fyne.NewPos(xToPixel(results1[i+1].x), yToPixel(y2))
			line.StrokeWidth = 3
			objects = append(objects, line)
		}
	}

	// Цвета для различных схем
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Красный
		color.RGBA{0, 150, 0, 255},   // Зеленый
		color.RGBA{0, 0, 255, 255},   // Синий
		color.RGBA{255, 165, 0, 255}, // Оранжевый
		color.RGBA{128, 0, 128, 255}, // Фиолетовый
		color.RGBA{0, 200, 200, 255}, // Голубой
	}

	// Численные решения
	for _, res := range results1 {
		var values []float64
		if derivOrder == 1 {
			values = []float64{res.rightDiff, res.leftDiff, res.central,
				res.threePointForward, res.threePointBackward, res.fivePoint}
		} else {
			values = []float64{res.secondDeriv3Point, res.secondDeriv4FWD,
				res.secondDeriv4BWD, res.secondDeriv5Point}
		}

		for idx, val := range values {
			if val != 0 && val >= yMin && val <= yMax {
				circle := canvas.NewCircle(colors[idx%len(colors)])
				circle.FillColor = colors[idx%len(colors)]
				circle.Resize(fyne.NewSize(6, 6))
				circle.Move(fyne.NewPos(xToPixel(res.x)-3, yToPixel(val)-3))
				objects = append(objects, circle)
			}
		}
	}

	return container.NewWithoutLayout(objects...)
}

// Создает отдельный график сравнения одной схемы с аналитикой
func createComparisonPlot(results []DerivativeResults, derivOrder int, schemeIndex int, schemeName string, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	xMin, xMax := -0.5, 3.5
	var yMin, yMax float64

	if derivOrder == 1 {
		yMin, yMax = -30.0, 10.0
	} else {
		yMin, yMax = -50.0, 50.0
	}

	marginLeft, marginRight := 60.0, 30.0
	marginTop, marginBottom := 30.0, 50.0
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
	for y := math.Ceil(yMin); y <= math.Floor(yMax); y += 10 {
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
		label.TextSize = 9
		label.Move(fyne.NewPos(xToPixel(x)-8, yToPixel(0)+8))
		objects = append(objects, label)
	}
	for y := math.Ceil(yMin); y <= math.Floor(yMax); y += 10 {
		if y != 0 {
			label := canvas.NewText(fmt.Sprintf("%.0f", y), color.Black)
			label.TextSize = 9
			label.Move(fyne.NewPos(xToPixel(0)-30, yToPixel(y)-5))
			objects = append(objects, label)
		}
	}

	// Заголовок графика
	title := canvas.NewText(schemeName, color.Black)
	title.TextSize = 11
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(float32(width/2)-50, 5))
	objects = append(objects, title)

	// Аналитическое решение (черная линия)
	for i := 0; i < len(results)-1; i++ {
		var y1, y2 float64
		if derivOrder == 1 {
			y1 = results[i].analytical1
			y2 = results[i+1].analytical1
		} else {
			y1 = results[i].analytical2
			y2 = results[i+1].analytical2
		}

		if y1 >= yMin && y1 <= yMax && y2 >= yMin && y2 <= yMax {
			line := canvas.NewLine(color.Black)
			line.Position1 = fyne.NewPos(xToPixel(results[i].x), yToPixel(y1))
			line.Position2 = fyne.NewPos(xToPixel(results[i+1].x), yToPixel(y2))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}

	// Цвет для схемы
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Красный
		color.RGBA{0, 150, 0, 255},   // Зеленый
		color.RGBA{0, 0, 255, 255},   // Синий
		color.RGBA{255, 165, 0, 255}, // Оранжевый
		color.RGBA{128, 0, 128, 255}, // Фиолетовый
		color.RGBA{0, 200, 200, 255}, // Голубой
	}
	schemeColor := colors[schemeIndex%len(colors)]

	// Численное решение (цветные точки)
	for _, res := range results {
		var val float64
		if derivOrder == 1 {
			values := []float64{res.rightDiff, res.leftDiff, res.central,
				res.threePointForward, res.threePointBackward, res.fivePoint}
			val = values[schemeIndex]
		} else {
			values := []float64{res.secondDeriv3Point, res.secondDeriv4FWD,
				res.secondDeriv4BWD, res.secondDeriv5Point}
			val = values[schemeIndex]
		}

		if val != 0 && val >= yMin && val <= yMax {
			circle := canvas.NewCircle(schemeColor)
			circle.FillColor = schemeColor
			circle.Resize(fyne.NewSize(6, 6))
			circle.Move(fyne.NewPos(xToPixel(res.x)-3, yToPixel(val)-3))
			objects = append(objects, circle)
		}
	}

	return container.NewWithoutLayout(objects...)
}

func createResultsTable(results1, results2 []DerivativeResults, h1, h2 float64) string {
	var text string
	text += fmt.Sprintf("РЕЗУЛЬТАТЫ ЧИСЛЕННОГО ДИФФЕРЕНЦИРОВАНИЯ\n")
	text += fmt.Sprintf("Шаг h = %.2f\n\n", h1)
	text += fmt.Sprintf("%-8s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "f'(аналит)", "Правая", "Левая", "Центр.", "3-точ.впер", "5-точ.")
	text += fmt.Sprintf("%s\n", "----------------------------------------------------------------------------")

	for _, res := range results1 {
		text += fmt.Sprintf("%-8.2f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f\n",
			res.x, res.analytical1, res.rightDiff, res.leftDiff, res.central,
			res.threePointForward, res.fivePoint)
	}

	text += fmt.Sprintf("\n\nВТОРЫЕ ПРОИЗВОДНЫЕ:\n")
	text += fmt.Sprintf("%-8s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "f''(аналит)", "3-точ.", "4-точ.впер", "4-точ.наз", "5-точ.")
	text += fmt.Sprintf("%s\n", "----------------------------------------------------------------------------")

	for _, res := range results1 {
		text += fmt.Sprintf("%-8.2f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f\n",
			res.x, res.analytical2, res.secondDeriv3Point, res.secondDeriv4FWD,
			res.secondDeriv4BWD, res.secondDeriv5Point)
	}

	return text
}

func createErrorTable(results1, results2 []DerivativeResults, h1, h2 float64) string {
	var text string
	text += fmt.Sprintf("АНАЛИЗ ПОГРЕШНОСТЕЙ\n\n")
	text += fmt.Sprintf("ПЕРВАЯ ПРОИЗВОДНАЯ (h = %.2f):\n", h1)
	text += fmt.Sprintf("%-8s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "Правая", "Левая", "Центр.", "3-точ.впер", "3-точ.наз", "5-точ.")
	text += fmt.Sprintf("%s\n", "----------------------------------------------------------------------------")

	for _, res := range results1 {
		text += fmt.Sprintf("%-8.2f | %12.2e | %12.2e | %12.2e | %12.2e | %12.2e | %12.2e\n",
			res.x,
			math.Abs(res.rightDiff-res.analytical1),
			math.Abs(res.leftDiff-res.analytical1),
			math.Abs(res.central-res.analytical1),
			math.Abs(res.threePointForward-res.analytical1),
			math.Abs(res.threePointBackward-res.analytical1),
			math.Abs(res.fivePoint-res.analytical1))
	}

	text += fmt.Sprintf("\n\nВТОРАЯ ПРОИЗВОДНАЯ (h = %.2f):\n", h1)
	text += fmt.Sprintf("%-8s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "3-точ.", "4-точ.впер", "4-точ.наз", "5-точ.")
	text += fmt.Sprintf("%s\n", "----------------------------------------------------------------------------")

	for _, res := range results1 {
		text += fmt.Sprintf("%-8.2f | %12.2e | %12.2e | %12.2e | %12.2e\n",
			res.x,
			math.Abs(res.secondDeriv3Point-res.analytical2),
			math.Abs(res.secondDeriv4FWD-res.analytical2),
			math.Abs(res.secondDeriv4BWD-res.analytical2),
			math.Abs(res.secondDeriv5Point-res.analytical2))
	}

	return text
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Численное дифференцирование")
	myWindow.Resize(fyne.NewSize(1000, 800))

	a, b := -0.5, 3.5
	h1 := 0.2
	h2 := h1 / 2

	results1 := computeDerivatives(a, b, h1)
	results2 := computeDerivatives(a, b, h2)

	// Имена схем для первой производной
	scheme1Names := []string{
		"Правая разностная (O(h))",
		"Левая разностная (O(h))",
		"Центральная разностная (O(h²))",
		"Трехточечная смещенная вперед (O(h²))",
		"Трехточечная смещенная назад (O(h²))",
		"Пятиточечная центральная (O(h⁴))",
	}

	scheme1Descriptions := []string{
		"f'(x) ≈ [f(x+h) - f(x)] / h",
		"f'(x) ≈ [f(x) - f(x-h)] / h",
		"f'(x) ≈ [f(x+h) - f(x-h)] / (2h)",
		"f'(x) ≈ [-3f(x) + 4f(x+h) - f(x+2h)] / (2h)",
		"f'(x) ≈ [3f(x) - 4f(x-h) + f(x-2h)] / (2h)",
		"f'(x) ≈ [f(x-2h) - 8f(x-h) + 8f(x+h) - f(x+2h)] / (12h)",
	}

	// Имена схем для второй производной
	scheme2Names := []string{
		"Трехточечная центральная (O(h²))",
		"Четырехточечная смещенная вперед (O(h²))",
		"Четырехточечная смещенная назад (O(h²))",
		"Пятиточечная центральная (O(h⁴))",
	}

	scheme2Descriptions := []string{
		"f''(x) ≈ [f(x-h) - 2f(x) + f(x+h)] / h²",
		"f''(x) ≈ [2f(x) - 5f(x+h) + 4f(x+2h) - f(x+3h)] / h²",
		"f''(x) ≈ [2f(x) - 5f(x-h) + 4f(x-2h) - f(x-3h)] / h²",
		"f''(x) ≈ [-f(x-2h) + 16f(x-h) - 30f(x) + 16f(x+h) - f(x+2h)] / (12h²)",
	}

	// Вкладка с общим графиком первой производной (все 6 схем)
	combinedPlot1 := createCombinedPlot(results1, results2, 1, h1, h2, 1100, 600)
	plot1AllContent := container.NewVBox(
		widget.NewLabel("ПЕРВАЯ ПРОИЗВОДНАЯ: Все 6 схем численного дифференцирования"),
		widget.NewLabel("Черная линия — аналитическое решение"),
		widget.NewLabel("Цветные точки — различные численные схемы"),
		widget.NewSeparator(),
		combinedPlot1,
	)
	plot1AllContainer := container.NewVScroll(plot1AllContent)

	// Создание вкладок для каждой схемы первой производной
	var scheme1Tabs []*container.TabItem
	for i := 0; i < 6; i++ {
		compPlot := createComparisonPlot(results1, 1, i, scheme1Names[i], 1100, 600)
		content := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("ПЕРВАЯ ПРОИЗВОДНАЯ: %s", scheme1Names[i])),
			widget.NewLabel(fmt.Sprintf("Формула: %s", scheme1Descriptions[i])),
			widget.NewLabel("Черная линия — аналитическое решение, цветные точки — численная схема"),
			widget.NewSeparator(),
			compPlot,
		)
		tabName := fmt.Sprintf("f' схема %d", i+1)
		scheme1Tabs = append(scheme1Tabs, container.NewTabItem(tabName, container.NewVScroll(content)))
	}

	// Вкладка с общим графиком второй производной (все 4 схемы)
	combinedPlot2 := createCombinedPlot(results1, results2, 2, h1, h2, 1100, 600)
	plot2AllContent := container.NewVBox(
		widget.NewLabel("ВТОРАЯ ПРОИЗВОДНАЯ: Все 4 схемы численного дифференцирования"),
		widget.NewLabel("Черная линия — аналитическое решение"),
		widget.NewLabel("Цветные точки — различные численные схемы"),
		widget.NewSeparator(),
		combinedPlot2,
	)
	plot2AllContainer := container.NewVScroll(plot2AllContent)

	// Создание вкладок для каждой схемы второй производной
	var scheme2Tabs []*container.TabItem
	for i := 0; i < 4; i++ {
		compPlot := createComparisonPlot(results1, 2, i, scheme2Names[i], 1100, 600)
		content := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("ВТОРАЯ ПРОИЗВОДНАЯ: %s", scheme2Names[i])),
			widget.NewLabel(fmt.Sprintf("Формула: %s", scheme2Descriptions[i])),
			widget.NewLabel("Черная линия — аналитическое решение, цветные точки — численная схема"),
			widget.NewSeparator(),
			compPlot,
		)
		tabName := fmt.Sprintf("f'' схема %d", i+1)
		scheme2Tabs = append(scheme2Tabs, container.NewTabItem(tabName, container.NewVScroll(content)))
	}

	// Таблица результатов
	resultsText := createResultsTable(results1, results2, h1, h2)
	resultsContainer := container.NewVScroll(widget.NewLabel(resultsText))

	// Таблица погрешностей
	errorText := createErrorTable(results1, results2, h1, h2)
	errorContainer := container.NewVScroll(widget.NewLabel(errorText))

	// Информация
	infoText := `ЧИСЛЕННОЕ ДИФФЕРЕНЦИРОВАНИЕ

Функция: y = ln[(cos²x + 3)/(x+1)] · (x³ + 3cos2x)
Интервал: [-0.5, 3.5]
Начальный шаг: h = 0.2

СХЕМЫ ПЕРВОГО ПОРЯДКА:
1. Правая разностная (O(h))
2. Левая разностная (O(h))
3. Центральная разностная (O(h²))
4. Трехточечная смещенная вперед (O(h²))
5. Трехточечная смещенная назад (O(h²))
6. Пятиточечная центральная (O(h⁴))

СХЕМЫ ВТОРОГО ПОРЯДКА:
1. Трехточечная центральная (O(h²))
2. Четырехточечная смещенная вперед (O(h²))
3. Четырехточечная смещенная назад (O(h²))
4. Пятиточечная центральная (O(h⁴))`

	infoContainer := container.NewVScroll(widget.NewLabel(infoText))

	// Создание всех вкладок
	var allTabs []*container.TabItem

	// Вкладка с общим графиком первой производной
	allTabs = append(allTabs, container.NewTabItem("f' все схемы", plot1AllContainer))

	// Вкладки для каждой схемы первой производной
	allTabs = append(allTabs, scheme1Tabs...)

	// Вкладка с общим графиком второй производной
	allTabs = append(allTabs, container.NewTabItem("f'' все схемы", plot2AllContainer))

	// Вкладки для каждой схемы второй производной
	allTabs = append(allTabs, scheme2Tabs...)

	// Вкладки с таблицами и информацией
	allTabs = append(allTabs, container.NewTabItem("Результаты", resultsContainer))
	allTabs = append(allTabs, container.NewTabItem("Погрешности", errorContainer))
	allTabs = append(allTabs, container.NewTabItem("Информация", infoContainer))

	tabs := container.NewAppTabs(allTabs...)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
