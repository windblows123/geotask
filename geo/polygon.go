package geo

import (
	"encoding/json"
	"github.com/kellydunn/golang-geo"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

const filename = "/geo/polygons.json"

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type PolygonChecker interface {
	Contains(point Point) bool // проверить, находится ли точка внутри полигона
	Allowed() bool             // разрешено ли входить в полигон
	RandomPoint() Point        // сгенерировать случайную точку внутри полигона
}

type Polygon struct {
	polygon *geo.Polygon
	allowed bool
}

func (p Polygon) Contains(point Point) bool {
	return p.polygon.Contains(geo.NewPoint(point.Lat, point.Lng))
}

func (p Polygon) Allowed() bool {
	return p.allowed
}

func (p Polygon) RandomPoint() Point {
	rand.Seed(time.Now().Unix())
	points := p.polygon.Points()
	var minLat, minLng, maxLat, maxLng, midLat, midLng, totalLat, totalLng, randomLat, randomLng float64
	minLat = points[0].Lat()
	minLng = points[0].Lng()
	for _, point := range points {
		if point.Lat() < minLat {
			minLat = point.Lat()
		}
		if point.Lng() < minLng {
			minLng = point.Lng()
		}
		if point.Lat() > maxLat {
			maxLat = point.Lat()
		}
		if point.Lng() > maxLng {
			maxLng = point.Lng()
		}
		totalLat += point.Lat()
		totalLng += point.Lng()
	}
	midLat = totalLat / float64(len(points))
	midLng = totalLng / float64(len(points))
	deltaLat := maxLat - minLat
	deltaLng := maxLng - minLng
	randomLat = midLat + math.Pow(-1, float64(rand.Intn(2)))*deltaLat
	randomLng = midLng + math.Pow(-1, float64(rand.Intn(2)))*deltaLng
	point := Point{
		Lat: randomLat,
		Lng: randomLng,
	}
	if p.Contains(point) {
		return point
	} else {
		return p.RandomPoint()
	}
}

func NewPolygon(geoPoints []Point, allowed bool) *Polygon {
	// используем библиотеку golang-geo для создания полигона
	points := make([]*geo.Point, 0, len(geoPoints))
	for _, point := range geoPoints {
		p := geo.NewPoint(point.Lat, point.Lng)
		points = append(points, p)
	}
	return &Polygon{
		polygon: geo.NewPolygon(points),
		allowed: allowed,
	}
}

func CheckPointIsAllowed(point Point, allowedZone PolygonChecker, disabledZones []PolygonChecker) bool {
	// проверить, находится ли точка в разрешенной зоне
	return allowedZone.Contains(point)
}

func GetRandomAllowedLocation(allowedZone PolygonChecker, disabledZones []PolygonChecker) Point {
	var point Point
	point = allowedZone.RandomPoint()

	return point
}

func UnmarshallPoints(filename string, zoneName string) []Point {
	var zones map[string][][]float64

	file, err := os.Open(filename)
	if err != nil {
		log.Println("Error opening file:", err)
		log.Fatal(err)
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&zones); err != nil {
		log.Println("Error decoding file:", err)
	}
	points, ok := zones[zoneName]
	parsedPoints := make([]Point, 0, len(points))
	if !ok {
		log.Fatal("No zone with name ", zoneName)
	}
	for _, point := range points {
		var p Point
		p.Lat = point[0]
		p.Lng = point[1]
		parsedPoints = append(parsedPoints, p)
	}
	return parsedPoints
}

func NewDisAllowedZone1() *Polygon {
	// добавить полигон с разрешенной зоной
	// полигоны лежат в /public/js/polygons.js
	points := UnmarshallPoints(filename, "noOrdersPolygon1")
	return NewPolygon(points, false)
}

func NewDisAllowedZone2() *Polygon {
	// добавить полигон с разрешенной зоной
	// полигоны лежат в /public/js/polygons.js
	points := UnmarshallPoints(filename, "noOrdersPolygon2")
	return NewPolygon(points, false)
}

func NewAllowedZone() *Polygon {
	// добавить полигон с разрешенной зоной
	// полигоны лежат в /public/js/polygons.js
	points := UnmarshallPoints(filename, "mainPolygon")
	return NewPolygon(points, true)
}
