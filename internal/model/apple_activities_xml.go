package model

import "encoding/xml"

// HealthData — корневой элемент
type HealthData struct {
	XMLName  xml.Name  `xml:"HealthData"`
	Locale   string    `xml:"locale,attr"`
	Workouts []Workout `xml:"Workout"`
}

// Workout — элемент тренировки
type Workout struct {
	ActivityType  string              `xml:"workoutActivityType,attr"`
	Duration      float64             `xml:"duration,attr"`
	DurationUnit  string              `xml:"durationUnit,attr"`
	SourceName    string              `xml:"sourceName,attr"`
	SourceVersion string              `xml:"sourceVersion,attr"`
	CreationDate  string              `xml:"creationDate,attr"`
	StartDate     string              `xml:"startDate,attr"`
	EndDate       string              `xml:"endDate,attr"`
	Statistics    []WorkoutStatistics `xml:"WorkoutStatistics"`
}

// WorkoutStatistics — статистика внутри тренировки
type WorkoutStatistics struct {
	Type      string  `xml:"type,attr"`
	StartDate string  `xml:"startDate,attr"`
	EndDate   string  `xml:"endDate,attr"`
	Sum       float64 `xml:"sum,attr"`
	Unit      string  `xml:"unit,attr"`
}
