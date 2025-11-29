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

// Метод Рунге-Ромберга для первой производной
func rungeRombergFirstDerivative(fn func(float64) float64, x, h float64, k float64, p float64) float64 {
	phiH := firstDerivCentral(fn, x, h)
	phiKH := firstDerivCentral(fn, x, k*h)
	return phiH + (phiH-phiKH)/(math.Pow(k, p)-1)
}

// Метод Рунге-Ромберга для второй производной
func rungeRombergSecondDerivative(fn func(float64) float64, x, h float64, k float64, p float64) float64 {
	phiH := secondDerivThreePoint(fn, x, h)
	phiKH := secondDerivThreePoint(fn, x, k*h)
	return phiH + (phiH-phiKH)/(math.Pow(k, p)-1)
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
	rungeRomberg1      float64
	rungeRomberg2      float64
}

func computeDerivatives(a, b, h float64) []DerivativeResults {
	var results []DerivativeResults

	for x := a; x <= b+1e-10; x += h {
		res := DerivativeResults{x: x}
		res.analytical1 = fPrimeAnalytical(x)
		res.analytical2 = fDoublePrimeAnalytical(x)

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

			// Метод Рунге-Ромберга (требует возможность использовать h и 2*h)
			// firstDerivCentral(x, h) требует x-h, x+h
			// firstDerivCentral(x, 2*h) требует x-2*h, x+2*h
			// Поэтому достаточно проверки x-2*h >= a && x+2*h <= b (уже проверено выше)
			res.rungeRomberg1 = rungeRombergFirstDerivative(f, x, h, 2, 2)
			res.rungeRomberg2 = rungeRombergSecondDerivative(f, x, h, 2, 2)
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

	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Красный
		color.RGBA{0, 150, 0, 255},   // Зеленый
		color.RGBA{0, 0, 255, 255},   // Синий
		color.RGBA{255, 165, 0, 255}, // Оранжевый
		color.RGBA{128, 0, 128, 255}, // Фиолетовый
		color.RGBA{0, 200, 200, 255}, // Голубой
	}

	// Сначала рисуем линии, соединяющие точки
	numMethods := 6
	if derivOrder == 2 {
		numMethods = 4
	}

	for methodIdx := 0; methodIdx < numMethods; methodIdx++ {
		for i := 0; i < len(results1)-1; i++ {
			var val1, val2 float64
			if derivOrder == 1 {
				values1 := []float64{results1[i].rightDiff, results1[i].leftDiff, results1[i].central,
					results1[i].threePointForward, results1[i].threePointBackward, results1[i].fivePoint}
				values2 := []float64{results1[i+1].rightDiff, results1[i+1].leftDiff, results1[i+1].central,
					results1[i+1].threePointForward, results1[i+1].threePointBackward, results1[i+1].fivePoint}
				val1 = values1[methodIdx]
				val2 = values2[methodIdx]
			} else {
				values1 := []float64{results1[i].secondDeriv3Point, results1[i].secondDeriv4FWD,
					results1[i].secondDeriv4BWD, results1[i].secondDeriv5Point}
				values2 := []float64{results1[i+1].secondDeriv3Point, results1[i+1].secondDeriv4FWD,
					results1[i+1].secondDeriv4BWD, results1[i+1].secondDeriv5Point}
				val1 = values1[methodIdx]
				val2 = values2[methodIdx]
			}

			if val1 != 0 && val2 != 0 && val1 >= yMin && val1 <= yMax && val2 >= yMin && val2 <= yMax {
				line := canvas.NewLine(colors[methodIdx%len(colors)])
				line.Position1 = fyne.NewPos(xToPixel(results1[i].x), yToPixel(val1))
				line.Position2 = fyne.NewPos(xToPixel(results1[i+1].x), yToPixel(val2))
				line.StrokeWidth = 2
				objects = append(objects, line)
			}
		}
	}

	// Затем рисуем точки поверх линий
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

	title := canvas.NewText(schemeName, color.Black)
	title.TextSize = 11
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(float32(width/2)-50, 5))
	objects = append(objects, title)

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

	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Красный
		color.RGBA{0, 150, 0, 255},   // Зеленый
		color.RGBA{0, 0, 255, 255},   // Синий
		color.RGBA{255, 165, 0, 255}, // Оранжевый
		color.RGBA{128, 0, 128, 255}, // Фиолетовый
		color.RGBA{0, 200, 200, 255}, // Голубой
	}
	schemeColor := colors[schemeIndex%len(colors)]

	// Сначала рисуем линии, соединяющие точки
	for i := 0; i < len(results)-1; i++ {
		var val1, val2 float64
		if derivOrder == 1 {
			values1 := []float64{results[i].rightDiff, results[i].leftDiff, results[i].central,
				results[i].threePointForward, results[i].threePointBackward, results[i].fivePoint}
			values2 := []float64{results[i+1].rightDiff, results[i+1].leftDiff, results[i+1].central,
				results[i+1].threePointForward, results[i+1].threePointBackward, results[i+1].fivePoint}
			val1 = values1[schemeIndex]
			val2 = values2[schemeIndex]
		} else {
			values1 := []float64{results[i].secondDeriv3Point, results[i].secondDeriv4FWD,
				results[i].secondDeriv4BWD, results[i].secondDeriv5Point}
			values2 := []float64{results[i+1].secondDeriv3Point, results[i+1].secondDeriv4FWD,
				results[i+1].secondDeriv4BWD, results[i+1].secondDeriv5Point}
			val1 = values1[schemeIndex]
			val2 = values2[schemeIndex]
		}

		if val1 != 0 && val2 != 0 && val1 >= yMin && val1 <= yMax && val2 >= yMin && val2 <= yMax {
			line := canvas.NewLine(schemeColor)
			line.Position1 = fyne.NewPos(xToPixel(results[i].x), yToPixel(val1))
			line.Position2 = fyne.NewPos(xToPixel(results[i+1].x), yToPixel(val2))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}

	// Затем рисуем точки поверх линий
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

func createPointwiseErrorPlot(results1, results2 []DerivativeResults, h1, h2 float64, derivOrder int, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	type PointError struct {
		x       float64
		errorH  float64
		errorH2 float64
		errorRR float64
	}

	var pointErrors []PointError

	for _, res1 := range results1 {
		var errH, errRR float64
		if derivOrder == 1 {
			if res1.central != 0 {
				errH = math.Abs(res1.central - res1.analytical1)
			}
			if res1.rungeRomberg1 != 0 {
				errRR = math.Abs(res1.rungeRomberg1 - res1.analytical1)
			}
		} else {
			if res1.secondDeriv3Point != 0 {
				errH = math.Abs(res1.secondDeriv3Point - res1.analytical2)
			}
			if res1.rungeRomberg2 != 0 {
				errRR = math.Abs(res1.rungeRomberg2 - res1.analytical2)
			}
		}

		pe := PointError{x: res1.x, errorH: errH, errorRR: errRR}
		pointErrors = append(pointErrors, pe)
	}

	for _, res2 := range results2 {
		var errH2 float64
		if derivOrder == 1 {
			if res2.central != 0 {
				errH2 = math.Abs(res2.central - res2.analytical1)
			}
		} else {
			if res2.secondDeriv3Point != 0 {
				errH2 = math.Abs(res2.secondDeriv3Point - res2.analytical2)
			}
		}

		for i := range pointErrors {
			if math.Abs(pointErrors[i].x-res2.x) < 1e-9 {
				pointErrors[i].errorH2 = errH2
				break
			}
		}
	}

	xMin, xMax := -0.5, 3.5
	yMin, yMax := math.MaxFloat64, 0.0

	for _, pe := range pointErrors {
		errors := []float64{pe.errorH, pe.errorH2, pe.errorRR}
		for _, e := range errors {
			if e > 0 && e < yMin {
				yMin = e
			}
			if e > yMax {
				yMax = e
			}
		}
	}

	yMin = yMin * 0.5
	yMax = yMax * 1.2

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

	for x := math.Ceil(xMin); x <= math.Floor(xMax); x++ {
		label := canvas.NewText(fmt.Sprintf("%.0f", x), color.Black)
		label.TextSize = 10
		label.Move(fyne.NewPos(xToPixel(x)-8, yToPixel(yMin)+10))
		objects = append(objects, label)
	}

	var title string
	if derivOrder == 1 {
		title = "Ошибки первой производной в каждой точке"
	} else {
		title = "Ошибки второй производной в каждой точке"
	}
	titleLabel := canvas.NewText(title, color.Black)
	titleLabel.TextSize = 14
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Move(fyne.NewPos(float32(width/2)-150, 10))
	objects = append(objects, titleLabel)

	yLabel := canvas.NewText("Абсолютная ошибка (log шкала)", color.Black)
	yLabel.TextSize = 11
	yLabel.Move(fyne.NewPos(10, float32(height/2)-50))
	objects = append(objects, yLabel)

	xLabel := canvas.NewText("x", color.Black)
	xLabel.TextSize = 11
	xLabel.Move(fyne.NewPos(float32(width/2), float32(height)-30))
	objects = append(objects, xLabel)

	colorH := color.RGBA{255, 100, 100, 255}  // красный для h
	colorH2 := color.RGBA{100, 100, 255, 255} // синий для h/2
	colorRR := color.RGBA{100, 200, 100, 255} // зеленый для RR

	for i := 0; i < len(pointErrors)-1; i++ {
		pe1 := pointErrors[i]
		pe2 := pointErrors[i+1]

		if pe1.errorH > 0 && pe2.errorH > 0 {
			line := canvas.NewLine(colorH)
			line.Position1 = fyne.NewPos(xToPixel(pe1.x), yToPixel(pe1.errorH))
			line.Position2 = fyne.NewPos(xToPixel(pe2.x), yToPixel(pe2.errorH))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}

		if pe1.errorH2 > 0 && pe2.errorH2 > 0 {
			line := canvas.NewLine(colorH2)
			line.Position1 = fyne.NewPos(xToPixel(pe1.x), yToPixel(pe1.errorH2))
			line.Position2 = fyne.NewPos(xToPixel(pe2.x), yToPixel(pe2.errorH2))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}

		if pe1.errorRR > 0 && pe2.errorRR > 0 {
			line := canvas.NewLine(colorRR)
			line.Position1 = fyne.NewPos(xToPixel(pe1.x), yToPixel(pe1.errorRR))
			line.Position2 = fyne.NewPos(xToPixel(pe2.x), yToPixel(pe2.errorRR))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}

	for _, pe := range pointErrors {
		if pe.errorH > 0 {
			circle := canvas.NewCircle(colorH)
			circle.FillColor = colorH
			circle.Resize(fyne.NewSize(5, 5))
			circle.Move(fyne.NewPos(xToPixel(pe.x)-2.5, yToPixel(pe.errorH)-2.5))
			objects = append(objects, circle)
		}

		if pe.errorH2 > 0 {
			circle := canvas.NewCircle(colorH2)
			circle.FillColor = colorH2
			circle.Resize(fyne.NewSize(5, 5))
			circle.Move(fyne.NewPos(xToPixel(pe.x)-2.5, yToPixel(pe.errorH2)-2.5))
			objects = append(objects, circle)
		}

		if pe.errorRR > 0 {
			circle := canvas.NewCircle(colorRR)
			circle.FillColor = colorRR
			circle.Resize(fyne.NewSize(5, 5))
			circle.Move(fyne.NewPos(xToPixel(pe.x)-2.5, yToPixel(pe.errorRR)-2.5))
			objects = append(objects, circle)
		}
	}

	legendX := float32(width) - 200
	legendY := float32(marginTop) + 20
	legendTitle := canvas.NewText("Легенда:", color.Black)
	legendTitle.TextSize = 10
	legendTitle.TextStyle = fyne.TextStyle{Bold: true}
	legendTitle.Move(fyne.NewPos(legendX, legendY))
	objects = append(objects, legendTitle)

	labels := []string{fmt.Sprintf("h=%.2f", h1), fmt.Sprintf("h/2=%.2f", h2), "Рунге-Ромберг"}
	colors := []color.Color{colorH, colorH2, colorRR}

	for i, label := range labels {
		square := canvas.NewRectangle(colors[i])
		square.Move(fyne.NewPos(legendX, legendY+20+float32(i)*20))
		square.Resize(fyne.NewSize(15, 15))
		objects = append(objects, square)

		text := canvas.NewText(label, color.Black)
		text.TextSize = 9
		text.Move(fyne.NewPos(legendX+20, legendY+20+float32(i)*20))
		objects = append(objects, text)
	}

	return container.NewWithoutLayout(objects...)
}

func createErrorComparisonPlot(results1, results2 []DerivativeResults, h1, h2 float64, width, height float64) *fyne.Container {
	plotCanvas := canvas.NewRectangle(color.White)
	plotCanvas.Resize(fyne.NewSize(float32(width), float32(height)))

	var objects []fyne.CanvasObject
	objects = append(objects, plotCanvas)

	type ErrorData struct {
		name      string
		errorH1_1 float64 // первая производная, h
		errorH2_1 float64 // первая производная, h/2
		errorRR_1 float64 // первая производная, Рунге-Ромберг
		errorH1_2 float64 // вторая производная, h
		errorH2_2 float64 // вторая производная, h/2
		errorRR_2 float64 // вторая производная, Рунге-Ромберг
	}

	methods := []ErrorData{
		{name: "Центр./3-точ."},
	}

	count1_1, count1_2 := 0, 0   // счетчики для h (первая и вторая производная)
	count2_1, count2_2 := 0, 0   // счетчики для h/2 (первая и вторая производная)
	countRR_1, countRR_2 := 0, 0 // счетчики для RR (первая и вторая производная)
	var sumH1_1, sumH2_1, sumRR_1, sumH1_2, sumH2_2, sumRR_2 float64

	for _, res1 := range results1 {
		if res1.central != 0 {
			sumH1_1 += math.Abs(res1.central - res1.analytical1)
			count1_1++
		}
		if res1.secondDeriv3Point != 0 {
			sumH1_2 += math.Abs(res1.secondDeriv3Point - res1.analytical2)
			count1_2++
		}
	}

	for _, res2 := range results2 {
		if res2.central != 0 {
			sumH2_1 += math.Abs(res2.central - res2.analytical1)
			count2_1++
		}
		if res2.secondDeriv3Point != 0 {
			sumH2_2 += math.Abs(res2.secondDeriv3Point - res2.analytical2)
			count2_2++
		}
	}

	for _, res1 := range results1 {
		if res1.rungeRomberg1 != 0 {
			sumRR_1 += math.Abs(res1.rungeRomberg1 - res1.analytical1)
			countRR_1++
		}
		if res1.rungeRomberg2 != 0 {
			sumRR_2 += math.Abs(res1.rungeRomberg2 - res1.analytical2)
			countRR_2++
		}
	}

	if count1_1 > 0 {
		methods[0].errorH1_1 = sumH1_1 / float64(count1_1)
	}
	if count1_2 > 0 {
		methods[0].errorH1_2 = sumH1_2 / float64(count1_2)
	}
	if count2_1 > 0 {
		methods[0].errorH2_1 = sumH2_1 / float64(count2_1)
	}
	if count2_2 > 0 {
		methods[0].errorH2_2 = sumH2_2 / float64(count2_2)
	}
	if countRR_1 > 0 {
		methods[0].errorRR_1 = sumRR_1 / float64(countRR_1)
	}
	if countRR_2 > 0 {
		methods[0].errorRR_2 = sumRR_2 / float64(countRR_2)
	}

	marginLeft, marginRight := 100.0, 50.0
	marginTop, marginBottom := 60.0, 100.0
	plotWidth := width - marginLeft - marginRight
	plotHeight := height - marginTop - marginBottom

	maxError := 0.0
	for _, m := range methods {
		errors := []float64{m.errorH1_1, m.errorH2_1, m.errorRR_1, m.errorH1_2, m.errorH2_2, m.errorRR_2}
		for _, e := range errors {
			if e > maxError {
				maxError = e
			}
		}
	}
	yMax := maxError * 1.2

	xToPixel := func(x float64) float32 {
		return float32(marginLeft + x*plotWidth)
	}
	yToPixel := func(y float64) float32 {
		return float32(marginTop + plotHeight - (y/yMax)*plotHeight)
	}

	gridColor := color.RGBA{220, 220, 220, 255}
	for i := 0; i <= 10; i++ {
		y := yMax * float64(i) / 10.0
		gridLine := canvas.NewLine(gridColor)
		gridLine.Position1 = fyne.NewPos(xToPixel(0), yToPixel(y))
		gridLine.Position2 = fyne.NewPos(xToPixel(1), yToPixel(y))
		gridLine.StrokeWidth = 0.5
		objects = append(objects, gridLine)

		label := canvas.NewText(fmt.Sprintf("%.2e", y), color.Black)
		label.TextSize = 9
		label.Move(fyne.NewPos(float32(marginLeft)-70, yToPixel(y)-5))
		objects = append(objects, label)
	}

	xAxis := canvas.NewLine(color.Black)
	xAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(0))
	xAxis.Position2 = fyne.NewPos(xToPixel(1), yToPixel(0))
	xAxis.StrokeWidth = 2
	objects = append(objects, xAxis)

	yAxis := canvas.NewLine(color.Black)
	yAxis.Position1 = fyne.NewPos(xToPixel(0), yToPixel(0))
	yAxis.Position2 = fyne.NewPos(xToPixel(0), yToPixel(yMax))
	yAxis.StrokeWidth = 2
	objects = append(objects, yAxis)

	title := canvas.NewText("Сравнение средних ошибок", color.Black)
	title.TextSize = 14
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(float32(width/2)-80, 10))
	objects = append(objects, title)

	yLabel := canvas.NewText("Средняя абсолютная ошибка", color.Black)
	yLabel.TextSize = 11
	yLabel.Move(fyne.NewPos(10, float32(height/2)-50))
	objects = append(objects, yLabel)

	barWidth := 0.08
	groupWidth := 0.3
	groups := []struct {
		name   string
		x      float64
		errors []float64
	}{
		{name: "Первая производная", x: 0.25, errors: []float64{methods[0].errorH1_1, methods[0].errorH2_1, methods[0].errorRR_1}},
		{name: "Вторая производная", x: 0.65, errors: []float64{methods[0].errorH1_2, methods[0].errorH2_2, methods[0].errorRR_2}},
	}

	colors := []color.Color{
		color.RGBA{255, 100, 100, 255}, // h - красный
		color.RGBA{100, 100, 255, 255}, // h/2 - синий
		color.RGBA{100, 200, 100, 255}, // RR - зеленый
	}

	labels := []string{fmt.Sprintf("h=%.2f", h1), fmt.Sprintf("h/2=%.2f", h2), "Рунге-Р."}

	for _, group := range groups {
		groupLabel := canvas.NewText(group.name, color.Black)
		groupLabel.TextSize = 10
		groupLabel.TextStyle = fyne.TextStyle{Bold: true}
		groupLabel.Move(fyne.NewPos(xToPixel(group.x-0.08), yToPixel(0)+15))
		objects = append(objects, groupLabel)

		// Столбцы для каждого метода
		for i, err := range group.errors {
			if err > 0 {
				xLeft := group.x - groupWidth/2 + float64(i)*groupWidth/3
				xRight := xLeft + barWidth

				// Прямоугольник
				rect := canvas.NewRectangle(colors[i])
				rect.Move(fyne.NewPos(xToPixel(xLeft), yToPixel(err)))
				rect.Resize(fyne.NewSize(xToPixel(xRight)-xToPixel(xLeft), yToPixel(0)-yToPixel(err)))
				objects = append(objects, rect)

				// Значение над столбцом
				valLabel := canvas.NewText(fmt.Sprintf("%.2e", err), color.Black)
				valLabel.TextSize = 8
				valLabel.Move(fyne.NewPos(xToPixel(xLeft)-10, yToPixel(err)-15))
				objects = append(objects, valLabel)
			}
		}
	}

	// Легенда
	legendX := float32(width) - 150
	legendY := float32(marginTop) + 20
	legendTitle := canvas.NewText("Легенда:", color.Black)
	legendTitle.TextSize = 10
	legendTitle.TextStyle = fyne.TextStyle{Bold: true}
	legendTitle.Move(fyne.NewPos(legendX, legendY))
	objects = append(objects, legendTitle)

	for i, label := range labels {
		// Цветной квадрат
		square := canvas.NewRectangle(colors[i])
		square.Move(fyne.NewPos(legendX, legendY+20+float32(i)*20))
		square.Resize(fyne.NewSize(15, 15))
		objects = append(objects, square)

		// Текст
		text := canvas.NewText(label, color.Black)
		text.TextSize = 9
		text.Move(fyne.NewPos(legendX+20, legendY+20+float32(i)*20))
		objects = append(objects, text)
	}

	return container.NewWithoutLayout(objects...)
}

// Создает график для метода Рунге-Ромберга
func createRungeRombergPlot(results []DerivativeResults, derivOrder int, schemeName string, width, height float64) *fyne.Container {
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
	title.Move(fyne.NewPos(float32(width/2)-100, 5))
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

	// Цвет для Рунге-Ромберга (фиолетовый)
	rrColor := color.RGBA{128, 0, 128, 255}

	// Сначала рисуем линии, соединяющие точки Рунге-Ромберга
	for i := 0; i < len(results)-1; i++ {
		var val1, val2 float64
		if derivOrder == 1 {
			val1 = results[i].rungeRomberg1
			val2 = results[i+1].rungeRomberg1
		} else {
			val1 = results[i].rungeRomberg2
			val2 = results[i+1].rungeRomberg2
		}

		if val1 != 0 && val2 != 0 && val1 >= yMin && val1 <= yMax && val2 >= yMin && val2 <= yMax {
			line := canvas.NewLine(rrColor)
			line.Position1 = fyne.NewPos(xToPixel(results[i].x), yToPixel(val1))
			line.Position2 = fyne.NewPos(xToPixel(results[i+1].x), yToPixel(val2))
			line.StrokeWidth = 2
			objects = append(objects, line)
		}
	}

	// Затем рисуем точки Рунге-Ромберга поверх линий
	for _, res := range results {
		var val float64
		if derivOrder == 1 {
			val = res.rungeRomberg1
		} else {
			val = res.rungeRomberg2
		}

		if val != 0 && val >= yMin && val <= yMax {
			circle := canvas.NewCircle(rrColor)
			circle.FillColor = rrColor
			circle.Resize(fyne.NewSize(7, 7))
			circle.Move(fyne.NewPos(xToPixel(res.x)-3.5, yToPixel(val)-3.5))
			objects = append(objects, circle)
		}
	}

	return container.NewWithoutLayout(objects...)
}

func createResultsTable(results1, results2 []DerivativeResults, h1, h2 float64) string {
	var text string
	text += fmt.Sprintf("РЕЗУЛЬТАТЫ ЧИСЛЕННОГО ДИФФЕРЕНЦИРОВАНИЯ\n")
	text += fmt.Sprintf("Шаг h = %.2f\n\n", h1)
	text += fmt.Sprintf("%-8s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "f'(аналит)", "Правая", "Левая", "Центр.", "3-точ.впер", "5-точ.", "Рунге-Р.")
	text += fmt.Sprintf("%s\n", "---------------------------------------------------------------------------------------------------")

	for _, res := range results1 {
		text += fmt.Sprintf("%-8.2f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f\n",
			res.x, res.analytical1, res.rightDiff, res.leftDiff, res.central,
			res.threePointForward, res.fivePoint, res.rungeRomberg1)
	}

	text += fmt.Sprintf("\n\nВТОРЫЕ ПРОИЗВОДНЫЕ:\n")
	text += fmt.Sprintf("%-8s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "f''(аналит)", "3-точ.", "4-точ.впер", "4-точ.наз", "5-точ.", "Рунге-Р.")
	text += fmt.Sprintf("%s\n", "---------------------------------------------------------------------------------------------------")

	for _, res := range results1 {
		text += fmt.Sprintf("%-8.2f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f | %12.6f\n",
			res.x, res.analytical2, res.secondDeriv3Point, res.secondDeriv4FWD,
			res.secondDeriv4BWD, res.secondDeriv5Point, res.rungeRomberg2)
	}

	return text
}

func createErrorTable(results1, results2 []DerivativeResults, h1, h2 float64) string {
	var text string
	text += fmt.Sprintf("АНАЛИЗ ПОГРЕШНОСТЕЙ\n\n")
	text += fmt.Sprintf("ПЕРВАЯ ПРОИЗВОДНАЯ (h = %.2f):\n", h1)
	text += fmt.Sprintf("%-8s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "Правая", "Левая", "Центр.", "3-точ.впер", "3-точ.наз", "5-точ.", "Рунге-Р.")
	text += fmt.Sprintf("%s\n", "---------------------------------------------------------------------------------------------------")

	for _, res := range results1 {
		text += fmt.Sprintf("%-8.2f | %12.2e | %12.2e | %12.2e | %12.2e | %12.2e | %12.2e | %12.2e\n",
			res.x,
			math.Abs(res.rightDiff-res.analytical1),
			math.Abs(res.leftDiff-res.analytical1),
			math.Abs(res.central-res.analytical1),
			math.Abs(res.threePointForward-res.analytical1),
			math.Abs(res.threePointBackward-res.analytical1),
			math.Abs(res.fivePoint-res.analytical1),
			math.Abs(res.rungeRomberg1-res.analytical1))
	}

	text += fmt.Sprintf("\n\nВТОРАЯ ПРОИЗВОДНАЯ (h = %.2f):\n", h1)
	text += fmt.Sprintf("%-8s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"x", "3-точ.", "4-точ.впер", "4-точ.наз", "5-точ.", "Рунге-Р.")
	text += fmt.Sprintf("%s\n", "---------------------------------------------------------------------------------------------------")

	for _, res := range results1 {
		text += fmt.Sprintf("%-8.2f | %12.2e | %12.2e | %12.2e | %12.2e | %12.2e\n",
			res.x,
			math.Abs(res.secondDeriv3Point-res.analytical2),
			math.Abs(res.secondDeriv4FWD-res.analytical2),
			math.Abs(res.secondDeriv4BWD-res.analytical2),
			math.Abs(res.secondDeriv5Point-res.analytical2),
			math.Abs(res.rungeRomberg2-res.analytical2))
	}

	// Добавляем статистику средних ошибок
	text += fmt.Sprintf("\n\nСТАТИСТИКА СРЕДНИХ ОШИБОК:\n")
	text += fmt.Sprintf("---------------------------------------------------------------------------------------------------\n")

	// Вычисляем средние для первой производной
	var sumCentral1, sumRR1 float64
	countCentral1, countRR1 := 0, 0
	for _, res := range results1 {
		if res.central != 0 {
			sumCentral1 += math.Abs(res.central - res.analytical1)
			countCentral1++
		}
		if res.rungeRomberg1 != 0 {
			sumRR1 += math.Abs(res.rungeRomberg1 - res.analytical1)
			countRR1++
		}
	}

	// Вычисляем средние для второй производной
	var sumThreePoint2, sumRR2 float64
	countThreePoint2, countRR2 := 0, 0
	for _, res := range results1 {
		if res.secondDeriv3Point != 0 {
			sumThreePoint2 += math.Abs(res.secondDeriv3Point - res.analytical2)
			countThreePoint2++
		}
		if res.rungeRomberg2 != 0 {
			sumRR2 += math.Abs(res.rungeRomberg2 - res.analytical2)
			countRR2++
		}
	}

	text += fmt.Sprintf("\nПервая производная (центральная разность):\n")
	if countCentral1 > 0 {
		text += fmt.Sprintf("  Средняя ошибка: %12.6e\n", sumCentral1/float64(countCentral1))
	}
	if countRR1 > 0 {
		text += fmt.Sprintf("  Рунге-Ромберг:  %12.6e\n", sumRR1/float64(countRR1))
	}

	text += fmt.Sprintf("\nВторая производная (3-точечная):\n")
	if countThreePoint2 > 0 {
		text += fmt.Sprintf("  Средняя ошибка: %12.6e\n", sumThreePoint2/float64(countThreePoint2))
	}
	if countRR2 > 0 {
		text += fmt.Sprintf("  Рунге-Ромберг:  %12.6e\n", sumRR2/float64(countRR2))
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
7. Метод Рунге-Ромберга (уточнение центральной разностной)

СХЕМЫ ВТОРОГО ПОРЯДКА:
1. Трехточечная центральная (O(h²))
2. Четырехточечная смещенная вперед (O(h²))
3. Четырехточечная смещенная назад (O(h²))
4. Пятиточечная центральная (O(h⁴))
5. Метод Рунге-Ромберга (уточнение трехточечной)

МЕТОД РУНГЕ-РОМБЕРГА:
Уточнение результата по формуле:
φ_refined = φ_h + (φ_h - φ_kh) / (k^p - 1)

где:
  k = 2 (коэффициент увеличения шага)
  p = 2 (порядок точности базового метода)
  φ_h - значение с шагом h
  φ_kh - значение с шагом k*h

Для первой производной использует центральную разностную схему.
Для второй производной использует трехточечную схему.`

	infoContainer := container.NewVScroll(widget.NewLabel(infoText))

	// Вкладки для метода Рунге-Ромберга
	rrPlot1 := createRungeRombergPlot(results1, 1, "Метод Рунге-Ромберга (первая производная)", 1100, 600)
	rrContent1 := container.NewVBox(
		widget.NewLabel("ПЕРВАЯ ПРОИЗВОДНАЯ: Метод Рунге-Ромберга"),
		widget.NewLabel("Формула: φ_refined = φ_h + (φ_h - φ_2h) / (2² - 1)"),
		widget.NewLabel("Уточнение центральной разностной схемы O(h²)"),
		widget.NewLabel("Черная линия — аналитическое решение, фиолетовые точки — метод Рунге-Ромберга"),
		widget.NewSeparator(),
		rrPlot1,
	)
	rrTab1 := container.NewTabItem("f' Рунге-Ромберг", container.NewVScroll(rrContent1))

	rrPlot2 := createRungeRombergPlot(results1, 2, "Метод Рунге-Ромберга (вторая производная)", 1100, 600)
	rrContent2 := container.NewVBox(
		widget.NewLabel("ВТОРАЯ ПРОИЗВОДНАЯ: Метод Рунге-Ромберга"),
		widget.NewLabel("Формула: φ_refined = φ_h + (φ_h - φ_2h) / (2² - 1)"),
		widget.NewLabel("Уточнение трехточечной центральной схемы O(h²)"),
		widget.NewLabel("Черная линия — аналитическое решение, фиолетовые точки — метод Рунге-Ромберга"),
		widget.NewSeparator(),
		rrPlot2,
	)
	rrTab2 := container.NewTabItem("f'' Рунге-Ромберг", container.NewVScroll(rrContent2))

	// Вкладка сравнения средних ошибок
	errorCompPlot := createErrorComparisonPlot(results1, results2, h1, h2, 1100, 600)
	errorCompContent := container.NewVBox(
		widget.NewLabel("СРАВНЕНИЕ СРЕДНИХ ОШИБОК"),
		widget.NewLabel("График показывает среднюю абсолютную ошибку для каждого шага интегрирования"),
		widget.NewLabel("Красный - шаг h, Синий - шаг h/2, Зеленый - метод Рунге-Ромберга"),
		widget.NewSeparator(),
		errorCompPlot,
	)
	errorCompTab := container.NewTabItem("Сравнение средних ошибок", container.NewVScroll(errorCompContent))

	// Вкладка точечных ошибок для первой производной
	pointError1Plot := createPointwiseErrorPlot(results1, results2, h1, h2, 1, 1100, 600)
	pointError1Content := container.NewVBox(
		widget.NewLabel("ОШИБКИ ПЕРВОЙ ПРОИЗВОДНОЙ В КАЖДОЙ ТОЧКЕ"),
		widget.NewLabel("График показывает абсолютную ошибку |численное - аналитическое| в каждой точке x"),
		widget.NewLabel("Красный - шаг h, Синий - шаг h/2, Зеленый - метод Рунге-Ромберга"),
		widget.NewLabel("Ось Y в логарифмической шкале"),
		widget.NewSeparator(),
		pointError1Plot,
	)
	pointError1Tab := container.NewTabItem("Ошибки f' по точкам", container.NewVScroll(pointError1Content))

	// Вкладка точечных ошибок для второй производной
	pointError2Plot := createPointwiseErrorPlot(results1, results2, h1, h2, 2, 1100, 600)
	pointError2Content := container.NewVBox(
		widget.NewLabel("ОШИБКИ ВТОРОЙ ПРОИЗВОДНОЙ В КАЖДОЙ ТОЧКЕ"),
		widget.NewLabel("График показывает абсолютную ошибку |численное - аналитическое| в каждой точке x"),
		widget.NewLabel("Красный - шаг h, Синий - шаг h/2, Зеленый - метод Рунге-Ромберга"),
		widget.NewLabel("Ось Y в логарифмической шкале"),
		widget.NewSeparator(),
		pointError2Plot,
	)
	pointError2Tab := container.NewTabItem("Ошибки f'' по точкам", container.NewVScroll(pointError2Content))

	// Создание всех вкладок
	var allTabs []*container.TabItem

	// Вкладка с общим графиком первой производной
	allTabs = append(allTabs, container.NewTabItem("f' все схемы", plot1AllContainer))

	// Вкладки для каждой схемы первой производной
	allTabs = append(allTabs, scheme1Tabs...)

	// Вкладка Рунге-Ромберга для первой производной
	allTabs = append(allTabs, rrTab1)

	// Вкладка с общим графиком второй производной
	allTabs = append(allTabs, container.NewTabItem("f'' все схемы", plot2AllContainer))

	// Вкладки для каждой схемы второй производной
	allTabs = append(allTabs, scheme2Tabs...)

	// Вкладка Рунге-Ромберга для второй производной
	allTabs = append(allTabs, rrTab2)

	// Вкладки с таблицами и информацией
	allTabs = append(allTabs, container.NewTabItem("Результаты", resultsContainer))
	allTabs = append(allTabs, container.NewTabItem("Погрешности", errorContainer))
	allTabs = append(allTabs, errorCompTab)
	allTabs = append(allTabs, pointError1Tab)
	allTabs = append(allTabs, pointError2Tab)
	allTabs = append(allTabs, container.NewTabItem("Информация", infoContainer))

	tabs := container.NewAppTabs(allTabs...)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
